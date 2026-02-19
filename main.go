package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strunck/uv4-middleware/globals"
	"strunck/uv4-middleware/web"
	"time"

	"github.com/go-co-op/gocron/v2"
)

// will be set during build, see Makefile
// GOFLAGS="-ldflags=-X=main.Version=${VERSION}"
var Version = "development"

func main() {

	log.Printf("Starting datahub in a Container Version: %s ...", Version)

	gs := globals.GlobalState{
		Loop:        globals.INIT_STATE,
		SessionName: "uv4_" + fmt.Sprintf("%d", time.Now().Unix()),
	}
	gs.Cfg.SettingsFile = globals.SETTINGSJSON
	gs.Cfg.Opcua.XmlFileName = globals.SYMBOLXML

	gs.Stats.InitDone = false
	gs.Stats.Version = Version
	if Version == "development" {
		gs.Stats.Developer = true
	} else {
		gs.Stats.Developer = false
	}

	gs.SetState(globals.SETUP_ENV)
	if err := gs.Cfg.Environment(gs.Stats.Developer); err != nil {
		gs.SetState(globals.SHUTDOWN)
		log.Fatalf("Error initializing environment: %v\nExiting...", err)
	}

	if err := gs.InitTelegramBot(); err != nil {
		log.Printf("📨 Error initializing Telegram Bot: %v\nContinuing without Telegram...", err)
	}

	ctx := context.Background()

	gs.SetState(globals.REQ_CONNECTION)
	if err := gs.InitOpcClient(ctx); err != nil {
		gs.SetState(globals.DISCONNECTED)
		log.Fatalf("Error requesting connection: %v\nExiting...", err)
	}

	gs.SetState(globals.SETUP_FILES)
	if err := gs.SetupFiles(); err != nil {
		log.Fatalf("🪑 Error reading from setup files: %v\nExiting...", err)
	}

	gs.SetState(globals.PLC_CONFIG_READ)
	if err := gs.Cfg.ReadKonfigViaOpcUa(globals.READMAXWAIT, gs.OpcClient, ctx); err != nil {
		gs.SetState(globals.RETRY)
		log.Printf("Error reading configuration via OPC UA: %s\n", err.Error())
	}

	if err := gs.Cfg.SetupKonfigGroup(gs.OpcClient, ctx); err != nil {
		log.Fatalf("Error setting up Konfig group: %s\n", err.Error())
	}

	if err := gs.Cfg.InitSWSettings(); err != nil {
		log.Fatalf("Error initializing SW settings: %s\nExiting...", err.Error())
	}
	gs.SetState(globals.PLC_CONFIG_ALARMS)
	if err := gs.Cfg.SetupAlarmMessages(gs.OpcClient, globals.READMAXWAIT, ctx); err != nil {
		log.Printf("Error setting up alarm messages: %s\nExiting...", err.Error())
	}

	gs.SetState(globals.PLC_CONFIG_DONE)
	// Print the parsed setup configuration for debugging
	output, _ := json.MarshalIndent(&gs.Cfg, "", "  ")
	os.WriteFile("result.json", output, 0644)

	gs.SetState(globals.SETUP_DONE)
	gs.Stats.InitDone = true
	gs.Stats.Reportdatum = "Noch nicht implementiert"
	gs.Stats.Uv4Version = gs.Cfg.Opcua.XmlFileName

	//*****************************************************************
	//***************************   MAIN LOOP   ***********************
	//*****************************************************************

	// Channel für Fehler
	errCh := make(chan error, 1)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	go RunHttp(errCh, &gs)

	runNats(4222, &gs)

	// Initialize glocal.Plc.Flat für Websocker middleware/data/v2
	gs.InitFlatHMIvalues()

	// Prepare the OPC UA Monitoring, monitor.go
	// We need the Opc UA Client connection, see global.InitOpcClient in file opclient.go
	gs.SetState(globals.REQ_MONITORING)
	gs.InitMonitorOpcUa(ctx)
	log.Println("Monitoring started")
	go gs.RecieveOpcUaMessage(ctx)
	gs.SetState(globals.MONITORING)

	log.Printf("Start Reading Alarms every %v ..", globals.INTERVAL)
	go gs.StartPeriodicAlarmReader(ctx)

	// Scheduled Tasks Rare

	// FTP Client
	if err := gs.FetchReportFiles(); err != nil {
		log.Printf("Error fetching report files: %v\n", err)
	}
	if err := gs.DailyReportGatherer(); err != nil {
		log.Printf("Error gathering daily reports: %v\n", err)
	}

	s, err := gocron.NewScheduler()
	defer func() { _ = s.Shutdown() }()
	if err != nil {
		log.Fatalf("Error creating scheduler: %v\nExiting...", err)
	}

	s.NewJob(
		gocron.DailyJob(
			1,
			gocron.NewAtTimes(
				gocron.NewAtTime(2, 2, 2),
			),
		),
		gocron.NewTask(
			TaskReportDaily(&gs),
		),
	)

	s.Start()

	select {
	case err := <-errCh:
		if err != nil {
			log.Printf("main.go error: %v\n", err)
		}
		log.Println("where am i?")
	case <-signalCh: // Wait for Ctrl+C
		log.Println("Received interrupt signal, shutting down...")
		gs.SetState(globals.SHUTDOWN)

		// Clean up and close the OPC UA client
		if err := gs.CloseOpcClient(ctx); err != nil {
			log.Fatalf("Error closing connection: %v\nExiting...", err)
		}
		gs.SetState(globals.DISCONNECTED)

		log.Println("Exiting gracefully....")

		os.Exit(0)
	}
}

func RunHttp(errCh chan error, gs *globals.GlobalState) {

	httpPort := fmt.Sprintf(":%d", globals.HttpPort)
	log.Printf("Starting Webserver port %s\n", httpPort)

	srv, err := web.ServeHttp(gs)
	if err != nil {
		errCh <- fmt.Errorf("error starting HTTP server: %w", err)
		return
	}
	errCh <- http.ListenAndServe(httpPort, srv)
}

func runNats(natsSrvPort int, gs *globals.GlobalState) {

	// ---------------------- NATS ---------------------
	// Initialilize NATS server, Options can be set in nats.go
	nats_server := globals.StartNatServer(natsSrvPort)
	// Start the NATS server in a separate goroutine
	go nats_server.Start()

	// Wait for the server to be ready
	if !nats_server.ReadyForConnections(4 * time.Second) {
		log.Fatalf("NATS server did not start in time")
	}
	log.Printf("NATS server is running on 127.0.0.1:%d\n", natsSrvPort)

	ncerr := gs.ConnectToNats(nats_server)
	if ncerr != nil {
		log.Fatalf("main init we got Error connecting to NATS server: %v", ncerr)
	}
}

func TaskReportDaily(gs *globals.GlobalState) func() {
	return func() {
		log.Println("📃 Entering Cron Job...")

		gs.FetchReportFiles()
		gs.DailyReportGatherer()

		log.Println("📃 Cron Job finished...")

	}
}
