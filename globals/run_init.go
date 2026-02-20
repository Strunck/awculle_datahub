package globals

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func (gs *GlobalState) Run_Init(ctx context.Context) (err error) {

	gs.SetState(SETUP_ENV)
	if err := gs.Cfg.Environment(gs.Stats.Developer); err != nil {
		gs.SetState(SHUTDOWN)
		return fmt.Errorf("string Version from GOFLAGS, probably: %w", err)
	}

	if err := gs.InitTelegramBot(); err != nil {
		gs.Telegram.err = err
		log.Printf("📨 Error initializing Telegram Bot: %v\nContinuing without Telegram...", err)
	}

	gs.SetState(REQ_CONNECTION)
	if err := gs.InitOpcClient(ctx); err != nil {
		gs.SetState(DISCONNECTED)
		return fmt.Errorf("error initializing OPC UA client: %w", err)
	}

	gs.SetState(SETUP_FILES)
	if err := gs.SetupFiles(); err != nil {
		return fmt.Errorf("🪑 Error reading from setup files: %v\nExiting...", err)
	}

	gs.SetState(PLC_CONFIG_READ)
	if err := gs.Cfg.ReadKonfigViaOpcUa(READMAXWAIT, gs.OpcClient, ctx); err != nil {
		gs.SetState(RETRY)
		return fmt.Errorf("error reading configuration via OPC UA: %w", err)
	}

	if err := gs.Cfg.SetupKonfigGroup(gs.OpcClient, ctx); err != nil {
		return fmt.Errorf("error setting up Konfig group: %w", err)
	}

	if err := gs.Cfg.InitSWSettings(); err != nil {
		return fmt.Errorf("error initializing SW settings: %w", err)
	}
	gs.SetState(PLC_CONFIG_ALARMS)
	if err := gs.Cfg.SetupAlarmMessages(gs.OpcClient, READMAXWAIT, ctx); err != nil {
		return fmt.Errorf("error setting up alarm messages: %w", err)
	}

	gs.SetState(PLC_CONFIG_DONE)
	// Print the parsed setup configuration for debugging
	output, _ := json.MarshalIndent(&gs.Cfg, "", "  ")
	os.WriteFile("result.json", output, 0644)

	gs.SetState(SETUP_DONE)
	gs.Stats.InitDone = true
	gs.Stats.Reportdatum = "Noch nicht implementiert"
	gs.Stats.Uv4Version = gs.Cfg.Opcua.XmlFileName

	return nil
}
