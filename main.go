package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strunck/uv4-middleware/globals"
	"strunck/uv4-middleware/web"
)

// will be set during build, see Makefile
// GOFLAGS="-ldflags=-X=main.Version=${VERSION}"
var Version = "development"

func main() {

	//*****************************************************************
	//*******************   INIT GLOBAL STATE   ***********************
	//*****************************************************************
	log.Printf("Starting datahub in a Container Version: %s ...", Version)

	gs := globals.GlobalState{
		Loop: globals.INIT_STATE,
	}

	gs.Cfg.SettingsFile = globals.SETTINGSJSON

	gs.Stats.InitDone = false
	gs.Stats.Version = Version
	if Version == "development" {
		gs.Stats.Developer = true
	} else {
		gs.Stats.Developer = false
	}

	// Parse command-line flags
	skipTelegramStartup := flag.Bool("skip-telegram-startup", false, "Skip Telegram startup message")
	flag.Parse()
	gs.Telegram.SkipStartupMessage = *skipTelegramStartup

	bctx := context.Background()

	gs.RunNats(4222)

	errChHttp := make(chan error)

	go runHttp(&gs, errChHttp)

	//signalCh := make(chan os.Signal, 1)
	ctx, stop := signal.NotifyContext(bctx, os.Interrupt)
	defer stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				runGuard(ctx, &gs)
				if gs.Loop == globals.RUN {
					return
				}
			}
		}
	}()

	select {
	case err := <-errChHttp:
		log.Println("case errChHttp")

		if err != nil {
			log.Fatalf("channeled error in HTTP server: %v", err)
		}

	case <-ctx.Done(): // Wait for Ctrl+C
		log.Println("Received signal ctx.Err(): ", ctx.Err())

		// Clean up and close the OPC UA client
		if err := gs.CloseOpcClient(ctx); err != nil {
			log.Fatalf("Error closing connection: %v\nExiting...", err)
		}
		log.Println("Exiting gracefully....")

		os.Exit(0)
	}

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
			gs.SetState(globals.PLC_CONFIG_VERSION)
		}

	case globals.PLC_CONFIG_VERSION:
		if err := gs.PlcConfigVersion(ctx); err != nil {
			log.Printf("⏳ Error reading PLC version: %v\nExiting...", err)
			gs.SetState(globals.RETRY)
		} else {
			gs.SetState(globals.PLC_CONFIG_ALARMS)
		}

	case globals.PLC_CONFIG_ALARMS:
		if err := gs.Cfg.SetupAlarmMessages(gs.OpcUa.Client, globals.READWAIT_INC_MAX, ctx); err != nil {
			log.Printf("error setting up alarm messages: %v", err)
			gs.SetState(globals.RETRY)
		} else {
			gs.SetState(globals.PLC_CONFIG_METRICS)
		}

	case globals.PLC_CONFIG_METRICS:
		if err := gs.InitPrometheus(); err != nil {
			log.Printf("error initializing Prometheus metrics: %v", err)
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
		if err := startMonitor(ctx, gs); err != nil {
			log.Printf("runGuard run_MainLoop failed: %v", err)
			gs.SetState(globals.RETRY_WAIT)
		} else {
			gs.SetState(globals.RUN)
		}

	case globals.RUN:

	default:
		log.Fatalf("Invalid state: %v", gs.Loop)
	}
}

func startMonitor(ctx context.Context, gs *globals.GlobalState) (err error) {
	fmt.Printf("\n\nrun_MainLoop\n\n\n")

	// Channel für Fehler
	errChMonitor := make(chan error, 1)
	errChAlarm := make(chan error, 1)

	gs.RecieveOpcUaMessage(ctx, errChMonitor)

	gs.StartPeriodicAlarmReader(ctx, errChAlarm)

	// Daily Report Cron Job
	if err := gs.Run_CronJob(); err != nil {
		log.Printf("Error Run_CronJob: %v\n", err)
	}

	select {

	case err := <-errChMonitor:
		if err != nil {
			return fmt.Errorf("channeled error in OPC UA monitoring: %w", err)
		}
	case <-ctx.Done():
		return fmt.Errorf("startMonitor ctx.Done()")
	case err := <-errChAlarm:
		if err != nil {
			return fmt.Errorf("channeled error in Alarm monitoring: %w", err)
		}
	}

	return nil
}
