package globals

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/awcullen/opcua/client"
	"github.com/awcullen/opcua/ua"
)

// PlcConfigDone prints the konfig to result.json
func (gs *GlobalState) PlcConfigDone() {
	// Print the parsed setup configuration for debugging
	output, _ := json.MarshalIndent(&gs.Cfg, "", "  ")
	os.WriteFile("result.json", output, 0644)

	gs.Stats.InitDone = true
	gs.Stats.Reportdatum = "Noch nicht implementiert"
	gs.Stats.Uv4Version = gs.Cfg.Plc.SymbolXML

}

// PlcConfigRead calls 3 functions to read the PLC congfigureation. See konfig.go
func (gs *GlobalState) PlcConfigRead(ctx context.Context) (err error) {

	if err := gs.Cfg.readKonfigViaOpcUa(READWAIT_INC_MAX, gs.OpcUa.Client, ctx); err != nil {
		return fmt.Errorf("error reading configuration via OPC UA: %w", err)
	}

	if err := gs.Cfg.setupKonfigGroup(gs.OpcUa.Client, ctx); err != nil {
		return fmt.Errorf("error setting up Konfig group: %w", err)
	}

	if err := gs.Cfg.initSWSettings(); err != nil {
		return fmt.Errorf("error initializing SW settings: %w", err)
	}

	return nil
}

// PlcConfigVersion reads the PLC version information.
func (gs *GlobalState) PlcConfigVersion(ctx context.Context) (err error) {
	name := fmt.Sprintf("|vprop|%s.Application.SoftwareRevision", gs.Cfg.Opcua.Prefix)
	nid := ua.NewNodeIDString(4, name)

	req := &ua.ReadRequest{
		NodesToRead: []ua.ReadValueID{
			{
				NodeID:      nid,
				AttributeID: ua.AttributeIDValue,
			},
		},
	}

	if gs.OpcUa.Client == nil {
		return fmt.Errorf("OPC UA client is not initialized")
	}

	res, err := gs.OpcUa.Client.Read(ctx, req)
	if err != nil {
		return fmt.Errorf("error reading SoftwareRevision: %w", err)
	}

	gs.Stats.CodesysVersion = fmt.Sprintf("%v", res.Results[0].Value)

	return nil
}

// Dial_OPCUA establishes a connection to the OPCUA server of the PLC.
func (gs *GlobalState) Dial_OPCUA(ctx context.Context) (err error) {

	var isNewSession bool

	// Set the session name
	if len(gs.OpcUa.SessionName) == 0 {
		gs.OpcUa.SessionName = "uv4_" + fmt.Sprintf("%d", time.Now().Unix())
		log.Println("Renew Init OPC UA Session Name to", gs.OpcUa.SessionName)

		isNewSession = true

		gs.OpcUa.Client, err = client.Dial(
			ctx,
			gs.Cfg.Opcua.Url,
			client.WithSessionName(gs.OpcUa.SessionName),
			client.WithInsecureSkipVerify(),
			client.WithSessionTimeout(OPCUA_SESSION_TIMEOUT),
			client.WithConnectTimeout(OPCUA_CONNECT_TIMEOUT.Microseconds()),
		)

		if err != nil {
			return fmt.Errorf("'Dial_OPCUA' failed: %w", err)
		}

	} else {
		sn := gs.OpcUa.SessionName
		if sn == "" {
			sn = "-no-session-name-"
		}
		sid := gs.OpcUa.Client.SessionID()

		log.Println("Reusing existing session name", sn, sid)
		isNewSession = false
	}

	if !isNewSession {

		ctxTimeout, cancel := context.WithTimeout(ctx, OPCUA_CONNECT_TIMEOUT)
		defer cancel()

		if err = gs.CloseOpcClient(ctxTimeout); err != nil {
			return fmt.Errorf("error closing existing OPC UA client: %w", err)
		}

		gs.OpcUa.SessionName = ""
		return fmt.Errorf("closing....")
	}
	return nil
}

func (gs *GlobalState) GetOpcPrefix(ctx context.Context) (err error) {
	//	ns=0;i=2261  can be either
	// 	-> EXOR ARM32-Linux QT5
	//  -> EXOR ARM64-Linux QT5
	nid := ua.NewNodeIDNumeric(0, 2261)
	nid_CodesysVersion := ua.NewNodeIDNumeric(0, 2264)

	req := &ua.ReadRequest{
		NodesToRead: []ua.ReadValueID{
			{
				NodeID:      nid,
				AttributeID: ua.AttributeIDValue,
			},
			{
				NodeID:      nid_CodesysVersion,
				AttributeID: ua.AttributeIDValue,
			},
		},
	}

	if gs.OpcUa.Client == nil {
		return fmt.Errorf("OPC UA client is not initialized")
	}

	res, err := gs.OpcUa.Client.Read(ctx, req)

	if err != nil {
		//gs.OpcUa.Client.Abort(ctx)
		return fmt.Errorf("error reading NumericNodeId %s: %w", nid.String(), err)
	}

	gs.Cfg.Opcua.Prefix = fmt.Sprintf("ns=4;s=|var|%s.Application.", res.Results[0].Value)
	gs.Stats.CodesysVersion = fmt.Sprintf("%v", res.Results[1].Value)

	return nil
}

func (gs *GlobalState) CloseOpcClient(ctx context.Context) (err error) {

	if gs.OpcUa.Client == nil {
		return fmt.Errorf("no OPC UA client to close")
	}

	err = gs.OpcUa.Client.Close(ctx)
	if err != nil {
		gs.OpcUa.Client.Abort(ctx)
		return fmt.Errorf("error closing OPC Client: %w", err)
	} else {
		log.Println("Successfully closed OpcClient-Session ", gs.OpcUa.SessionName)

	}

	return nil
}

// Connected resetts the attempt counter, which is read in the Retry_Wait function. Also Cleanup OPC Stuff
func (gs *GlobalState) Connected() (err error) {
	gs.Stats.Attempts = 0
	return nil
}

// Retry_Wait implements the waiting algorithm. For now it will increase the waiting time, until a maximum time is reached. At least it crashes the program.
func (gs *GlobalState) Retry_Wait(ctx context.Context) (err error) {

	timeoutInt := min(((gs.Stats.Attempts + 1) * RETRYWAIT_INCREMENT), RETRY_WAIT_SECONDS_MAX)
	timeout := time.Duration(timeoutInt) * time.Second

	if gs.Stats.Attempts >= RETRYWAIT_ATTEMPTS_MAX {
		return fmt.Errorf("⏳ retried too often, waited for %d seconds", timeoutInt)
	}

	// wait for TIMEOUT_OPCUA sec before switching to Retry
	log.Printf("⏳ Waiting for %d seconds before retrying connection...\n", timeoutInt)
	time.Sleep(timeout)

	return nil
}

// Retry tries to establish a TCP connenction to the OPCUA Port of the PLC.
func (gs *GlobalState) Retry() (err error) {

	// Try a Network connection re-establishment, Dial via TCP
	conn, err := net.DialTimeout("tcp", gs.Cfg.Plc.IPAddress+":"+gs.Cfg.Plc.OpcUaPort, NETDIAL_TIMEOUT)
	if err != nil {
		return fmt.Errorf("🔱 Network Dial attempt %d to PLC %s failed: %w\n", gs.Stats.Attempts, gs.Cfg.Plc.IPAddress, err)
	}
	defer conn.Close()

	log.Printf("🔱 Network Dial TCP to PLC %s succeeded\n", gs.Cfg.Plc.IPAddress)

	// Network connection established
	return nil
}

// ReEstablishSession tries to re-establish the OPCUA session.
func (gs *GlobalState) ReEstablishSession(ctx context.Context) (err error) {

	//session := gs.OpcClient.TransferSubscriptions()

	return nil
}

// monitors the connection status.
func (gs *GlobalState) GetConnectionStatus(ctx context.Context) error {

	if gs.OpcUa.Client == nil {
		return fmt.Errorf("🔱 OPC UA client is not initialized")
	}

	// Attempt a simple read to verify connection
	req := &ua.ReadRequest{
		NodesToRead: []ua.ReadValueID{
			{
				NodeID:      ua.NewNodeIDNumeric(0, 2259), // Server_ConnectionStatus
				AttributeID: ua.AttributeIDValue,
			},
		},
	}

	res, err := gs.OpcUa.Client.Read(ctx, req)
	if err != nil {
		return fmt.Errorf("🔱 OPC UA connection lost: %w", err)
	} else {
		rtext := "🔱 OPC UA connection status is "
		// get the test of the connection status
		if res.ServiceResult.IsGood() {
			rtext += "good"
		} else {
			rtext += "bad"
			return fmt.Errorf("%s: %v", rtext, res.Results[0].Value)
		}
		status := res.Results[0].Value
		log.Printf("%s: %v", rtext, status)
	}
	return nil
}
