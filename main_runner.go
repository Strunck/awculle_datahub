package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strunck/uv4-middleware/globals"
	"strunck/uv4-middleware/web"
)

// runGuard is a state machine. Rules: Every state will call a function, that will return an error or nil if successful.
// To keep the best overview, all states must be set in this swich case. Do not set the states in the functions, that get called here!
// the default case will crash the programm.
func runGuard(ctx context.Context, gs *globals.GlobalState) {

	switch gs.Loop {

	case globals.INIT_STATE:
		gs.SetState(globals.SETUP_ENV)

	case globals.SETUP_ENV:
		if err := gs.Cfg.Environment(gs.Stats.Developer); err != nil {
			gs.SetState(globals.SHUTDOWN)
			log.Fatalf("string Version from GOFLAGS, probably: %v", err)
		}
		gs.SetState(globals.SETUP_FILES)

	case globals.SETUP_FILES:
		if err := gs.SetupFiles(); err != nil {
			log.Fatalf("🪑 Error reading from setup files: %v\nExiting...", err)
		}
		gs.SetState(globals.INIT_TELEGRAM)

	case globals.INIT_TELEGRAM:
		if err := gs.InitTelegramBot(); err != nil {
			log.Printf("🤖 Error initializing Telegram bot: %v\nContinuing without Telegram...", err)
		}
		gs.SetState(globals.DIAL_OPCUA_SRV)

	case globals.INIT_TOPICS:
		if err := gs.InitTopics(); err != nil {
			log.Printf("Error initializing topics: %v\nExiting...", err)
			gs.SetState(globals.RETRY_WAIT)
		} else {
			gs.SetState(globals.PLC_CONFIG_READ)
		}

	case globals.PLC_CONFIG_READ:
		if err := gs.PlcConfigRead(ctx); err != nil {
			log.Printf("⏳ Error reading PLC configuration: %v\nExiting...", err)
			gs.SetState(globals.RETRY)
		} else {
			gs.SetState(globals.PLC_CONFIG_ALARMS)
		}

	case globals.PLC_CONFIG_ALARMS:
		if err := gs.Cfg.SetupAlarmMessages(gs.OpcClient, globals.READWAIT_INC_MAX, ctx); err != nil {
			log.Printf("error setting up alarm messages: %v", err)
			gs.SetState(globals.RETRY)
		} else {
			gs.SetState(globals.PLC_CONFIG_DONE)
		}

	case globals.PLC_CONFIG_DONE:
		gs.PlcConfigDone()
		gs.SetState(globals.REQ_MONITORING)

	case globals.DIAL_OPCUA_SRV:
		if err := gs.Dial_OPCUA(ctx); err != nil {
			log.Printf("runGuard Dial_OPCUA failed: %v", err)
			gs.SetState(globals.RETRY_WAIT)
			break
		}
		if err := gs.GetOpcPrefix(ctx); err != nil {
			log.Printf("runGuard GetOpcPrefix failed: %v", err)
			gs.SetState(globals.RETRY_WAIT)
			break
		}
		if err := gs.GetConnectionStatus(ctx); err != nil {
			log.Printf("runGuard GetConnectionStatus failed: %v", err)
			gs.SetState(globals.RETRY_WAIT)
			break
		}

		gs.SetState(globals.CONNECTED)

	case globals.RETRY:
		gs.Stats.Attempts++
		if err := gs.Retry(); err != nil {
			log.Printf("runGuard Retry failed: %v", err)
			gs.SetState(globals.RETRY_WAIT)
		} else {
			gs.SetState(globals.DIAL_OPCUA_SRV)
		}

	case globals.RETRY_WAIT:
		if err := gs.Retry_Wait(ctx); err != nil {
			log.Fatalf("runGuard Crash exit: %v", err)
		}
		gs.SetState(globals.RETRY)

	case globals.CONNECTED:
		gs.Connected()
		gs.SetState(globals.INIT_TOPICS)

	case globals.REQ_MONITORING:

		// Initialize glocal.Plc.Flat für Websocker middleware/data/v2
		gs.InitFlatHMIvalues()
		// Prepare the OPC UA Monitoring, monitor.go
		// We need the Opc UA Client connection, see global.InitOpcClient in file opclient.go
		if err := gs.InitMonitorOpcUa(ctx); err != nil {
			log.Printf("runGuard error InitMonitorOpcUa: %v", err)
			gs.SetState(globals.RETRY)
		} else {
			gs.SetState(globals.MONITORING)
		}

	case globals.MONITORING:
		err := run_MainLoop(ctx, gs)
		if err != nil {
			log.Printf("runGuard run_MainLoop failed: %v", err)
			gs.SetState(globals.RETRY_WAIT)
		}

	default:
		log.Fatalf("Invalid state: %v", gs.Loop)
	}
}

func run_MainLoop(ctx context.Context, gs *globals.GlobalState) (err error) {

	// Channel für Fehler
	errChHttp := make(chan error, 1)
	errChMonitor := make(chan error, 1)
	errChAlarm := make(chan error, 1)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	go runHttp(gs, errChHttp)

	gs.RunNats(4222)

	go gs.RecieveOpcUaMessage(ctx, errChMonitor)

	go gs.StartPeriodicAlarmReader(ctx, errChAlarm)

	// Daily Report Cron Job
	if err := gs.Run_CronJob(); err != nil {
		log.Printf("Error Run_CronJob: %v\n", err)
	}

	select {
	case err := <-errChHttp:
		if err != nil {
			return fmt.Errorf("channeled error in HTTP server: %w", err)
		}

	case err := <-errChMonitor:
		if err != nil {
			return fmt.Errorf("channeled error in OPC UA monitoring: %w", err)
		}
	case err := <-errChAlarm:
		if err != nil {
			return fmt.Errorf("channeled error in Alarm monitoring: %w", err)
		}
	case <-signalCh: // Wait for Ctrl+C
		log.Println("Received interrupt signal, shutting down...")

		// Clean up and close the OPC UA client
		if err := gs.CloseOpcClient(ctx); err != nil {
			log.Fatalf("Error closing connection: %v\nExiting...", err)
		}
		log.Println("Exiting gracefully....")

		os.Exit(0)
	}

	return nil
}

func runHttp(gs *globals.GlobalState, errCh chan error) {

	httpPort := fmt.Sprintf(":%d", globals.HttpPort)
	log.Printf("Starting Webserver port %s\n", httpPort)

	srv, err := web.ServeHttp(gs)
	if err != nil {
		errCh <- fmt.Errorf("error starting HTTP server: %w", err)
		return
	}
	errCh <- http.ListenAndServe(httpPort, srv)
}
