package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strunck/uv4-middleware/globals"
	"time"
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

	// Parse command-line flags
	skipTelegramStartup := flag.Bool("skip-telegram-startup", false, "Skip Telegram startup message")
	flag.Parse()

	ctx := context.Background()

	if err := gs.InitTelegramBot(*skipTelegramStartup); err != nil {
		log.Printf("📨 Error initializing Telegram Bot: %v\nContinuing without Telegram...", err)
	}

	//*****************************************************************
	//***************************   MAIN LOOP   ***********************
	//*****************************************************************

	for {
		runGuard(ctx, &gs)
	}

}
