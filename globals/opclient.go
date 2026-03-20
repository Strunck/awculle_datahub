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

func (gs *GlobalState) PlcConfigDone() {
	// Print the parsed setup configuration for debugging
	output, _ := json.MarshalIndent(&gs.Cfg, "", "  ")
	os.WriteFile("result.json", output, 0644)

	gs.Stats.InitDone = true
	gs.Stats.Reportdatum = "Noch nicht implementiert"
	gs.Stats.Uv4Version = gs.Cfg.Opcua.XmlFileName

}

func (gs *GlobalState) PlcConfigRead(ctx context.Context) (err error) {

	if err := gs.Cfg.ReadKonfigViaOpcUa(READMAXWAIT, gs.OpcClient, ctx); err != nil {
		return fmt.Errorf("error reading configuration via OPC UA: %w", err)
	}

	if err := gs.Cfg.SetupKonfigGroup(gs.OpcClient, ctx); err != nil {
		return fmt.Errorf("error setting up Konfig group: %w", err)
	}

	if err := gs.Cfg.InitSWSettings(); err != nil {
		return fmt.Errorf("error initializing SW settings: %w", err)
	}

	return nil
}

func (gs *GlobalState) Dial_OPCUA(ctx context.Context) (err error) {

	// Set the session name
	if len(gs.SessionName) == 0 {
		gs.SessionName = "uv4_" + fmt.Sprintf("%d", time.Now().Unix())
		log.Println("Renew Init OPC UA Session Name to", gs.SessionName)
	}

	gs.OpcClient, err = client.Dial(
		ctx,
		gs.Cfg.Opcua.Url,
		client.WithSessionName(gs.SessionName),
		client.WithInsecureSkipVerify(),
		client.WithSessionTimeout(1000*30),
		client.WithConnectTimeout(1000*5),
	)

	if err != nil {
		return fmt.Errorf("'Dial_OPCUA' failed: %w", err)
	}

	return nil
}

func (gs *GlobalState) GetOpcPrefix(ctx context.Context) (err error) {
	//	ns=0;i=2261  can be either
	// 	-> EXOR ARM32-Linux QT5
	//  -> EXOR ARM64-Linux QT5
	nid := ua.NewNodeIDNumeric(0, 2261)
	req := &ua.ReadRequest{
		NodesToRead: []ua.ReadValueID{
			{
				NodeID:      nid,
				AttributeID: ua.AttributeIDValue,
			},
		},
	}

	res, err := gs.OpcClient.Read(ctx, req)

	if err != nil {
		//gs.OpcClient.Abort(ctx)
		return fmt.Errorf("error reading NumericNodeId %s: %w", nid.String(), err)
	}

	gs.Cfg.Opcua.Prefix = fmt.Sprintf("ns=4;s=|var|%s.Application.", res.Results[0].Value)

	return nil
}

func (gs *GlobalState) CloseOpcClient(ctx context.Context) (err error) {

	if gs.OpcClient == nil {
		return fmt.Errorf("no OPC UA client to close")
	}

	err = gs.OpcClient.Close(ctx)
	if err != nil {
		gs.OpcClient.Abort(ctx)
		return fmt.Errorf("error closing OPC Client: %w", err)
	} else {
		log.Println("Successfully closed OpcClient-Session ", gs.SessionName)

	}

	return nil
}

func (gs *GlobalState) Connected() (err error) {
	gs.Stats.Attempts = 0
	return nil
}

func (gs *GlobalState) Retry_Wait(ctx context.Context) (err error) {

	timeoutInt := min(((gs.Stats.Attempts) * TIMEOUT_OPCUA), RETRY_WAIT_SECONDS_MAX)
	timeout := time.Duration(timeoutInt) * time.Second

	if timeoutInt >= CRASH_EXIT_WAIT_SECONDS {
		return fmt.Errorf("retried too often, waited for %d seconds", timeoutInt)
	}

	// wait for TIMEOUT_OPCUA sec before switching to Retry
	log.Printf("Waiting for %d seconds before retrying connection...\n", timeoutInt)
	time.Sleep(timeout)

	return nil
}

func (gs *GlobalState) Retry() (err error) {

	err = gs.retryJustTheNetwork()

	if err != nil {
		return fmt.Errorf("Retry() failed: %w", err)
	}

	// Network connection established
	return nil
}

func (gs *GlobalState) retryJustTheNetwork() error {

	// Try a Network connection re-establishment, Dial via TCP
	conn, err := net.DialTimeout("tcp", gs.Cfg.Plc.IPAddress+":80", TIMEOUT_PING)
	if err != nil {
		return fmt.Errorf("🔱 Network Dial TCP to PLC %s failed: %w\n", gs.Cfg.Plc.IPAddress, err)
	}
	defer conn.Close()

	log.Printf("🔱 Network Dial TCP to PLC %s succeeded\n", gs.Cfg.Plc.IPAddress)

	return nil
}
