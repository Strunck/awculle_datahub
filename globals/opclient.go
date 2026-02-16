package globals

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/awcullen/opcua/client"
	"github.com/awcullen/opcua/ua"
)

func (gs *GlobalState) InitOpcClient(ctx context.Context) (err error) {

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
		return fmt.Errorf("'InitOpcClient' failed to create client: %w", err)
	}

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
		gs.OpcClient.Abort(ctx)
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
