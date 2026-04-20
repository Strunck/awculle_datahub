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
	gs.Cfg.Opcua.XmlFileName = globals.SYMBOLXML

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
