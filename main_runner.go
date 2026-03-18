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
	"time"
)

func run_MainLoop(ctx context.Context, gs *globals.GlobalState) (err error) {

	// Channel für Fehler
	errCh := make(chan error, 1)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	go runHttp(gs, errCh)

	runNats(gs, 4222)

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

	// Daily Report Cron Job
	if err := gs.Run_CronJob(); err != nil {
		log.Printf("Error Run_CronJob: %v\n", err)
	}

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

func runNats(gs *globals.GlobalState, natsSrvPort int) {

	// ---------------------- NATS ---------------------
	// Initialilize NATS server, Options can be set in nats.go
	nats_server := globals.StartNatServer(natsSrvPort)
	// Start the NATS server in a separate goroutine
	nats_server.Start()

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

func runGuard(ctx context.Context, gs *globals.GlobalState) {
	gs.PrintState()
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
		gs.SetState(globals.DIAL_OPCUA_SRV)

	case globals.DIAL_OPCUA_SRV:
		if err := gs.DialAndConfigure(ctx); err != nil {
			log.Printf("Error in DialAndConfigure: %v\nEntering Retry Wait Loop...", err)
			gs.Retry_Wait(ctx)
		}
		gs.SetState(globals.CONNECTED)
		log.Println("Successfully connected to OPC UA server and configured.")

	case globals.RETRY:
		err := gs.Retry()
		if err != nil {
			log.Printf("Error in Retry: %v\nEntering Retry Wait Loop...", err)
			gs.Retry_Wait(ctx)
		}
	case globals.RETRY_WAIT:
		gs.Retry_Wait(ctx)

	case globals.CONNECTED:
		err := run_MainLoop(ctx, gs)
		if err != nil {
			log.Printf("Error in MainLoop: %v\nEntering Retry Wait Loop...", err)
			gs.Retry_Wait(ctx)
		}
	}
}
