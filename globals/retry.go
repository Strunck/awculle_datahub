package globals

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"
)

func (gs *GlobalState) Retry_Wait(ctx context.Context) {

	gs.SetState(RETRY_WAIT)

	attempts := 1

	for {

		timeoutInt := min(((attempts) * TIMEOUT_OPCUA), RETRY_WAIT_SECONDS_MAX)
		timeout := time.Duration(timeoutInt) * time.Second

		if timeoutInt >= CRASH_EXIT_WAIT_SECONDS {
			log.Fatalf("Too Long, Crash Exit!!!")
		}

		// wait for TIMEOUT_OPCUA sec before switching to Retry
		log.Printf("Waiting for %d seconds before retrying connection...\n", timeoutInt)
		time.Sleep(timeout)

		if err := gs.Retry(); err != nil {
			attempts++
			log.Printf("🔱 Attempt #%d failed: %v", attempts, err)
			continue
		}
		return
	}
}

func (gs *GlobalState) Retry() (err error) {
	gs.SetState(RETRY)
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
