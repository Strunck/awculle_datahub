package globals

import (
	"context"
	"log"
	"net"
)

func (gs *GlobalState) Retry(ctx context.Context) {

	gs.SetState(RETRY)

	if err := gs.InitOpcClient(ctx); err != nil {
		log.Printf("Error requesting connection: %v\nExiting...", err)
	}

	gs.SetState(REQ_MONITORING)

	gs.InitMonitorOpcUa(ctx)

	go gs.RecieveOpcUaMessage(ctx)

	go gs.StartPeriodicAlarmReader(ctx)

}

func (gs *GlobalState) RetryJustTheNetwork(pctx context.Context) error {

	// Try a Network connection re-establishment, Dial via TCP
	conn, err := net.DialTimeout("tcp", gs.Cfg.Plc.IPAddress+":80", TIMEOUT_PING)
	if err != nil {
		log.Printf("🔱 Network Dial TCP to PLC %s failed: %s\n", gs.Cfg.Plc.IPAddress, err.Error())
		return err
	}
	conn.Close()

	log.Printf("🔱 Network Dial TCP to PLC %s succeeded\n", gs.Cfg.Plc.IPAddress)

	return nil
}

func (gs *GlobalState) PeriodicRetry(pctx context.Context) error {

	return nil
}
