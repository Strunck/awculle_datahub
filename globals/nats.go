package globals

import (
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

func (gs *GlobalState) RunNats(natsSrvPort int) {

	if gs.NatsServer != nil {
		return
	}

	// ---------------------- NATS ---------------------
	// Initialilize NATS server, Options can be set in nats.go
	gs.NatsServer = startNatServer(natsSrvPort)
	// Start the NATS server in a separate goroutine
	gs.NatsServer.Start()

	// Wait for the server to be ready
	if !gs.NatsServer.ReadyForConnections(4 * time.Second) {
		log.Fatalf("NATS server did not start in time")
	}
	log.Printf("NATS server is running on 127.0.0.1:%d\n", natsSrvPort)

	ncerr := gs.ConnectToNats(gs.NatsServer)
	if ncerr != nil {
		log.Fatalf("main init we got Error connecting to NATS server: %v", ncerr)
	}
}

// StartNatServer starts an internal Nats Server
func startNatServer(port int) *server.Server {

	// Create a new NATS server configuration
	opts := &server.Options{
		Host:     "0.0.0.0",
		Port:     port,
		HTTPPort: 8222,
	}

	// Create a new NATS server instance
	ns, err := server.NewServer(opts)
	if err != nil {
		log.Fatalf("Error creating NATS server: %v", err)
	}

	log.Printf("NATS server created at %d", port)

	return ns

}

func (gs *GlobalState) ConnectToNats(natssrv *server.Server) (err error) {
	// Connect to NATS
	gs.NatsClient, err = nats.Connect(natssrv.ClientURL())
	if err != nil {
		return fmt.Errorf("error connecting to NATS server at %s: %v", natssrv.ClientURL(), err)
	}

	return nil
}

func (gs *GlobalState) ChanSubscribeWrapper(subj string, ch chan *nats.Msg) (*nats.Subscription, error) {

	// todo: check if subj makes sense. will it return messages

	return gs.NatsClient.ChanSubscribe(subj, ch)
}

func (gs *GlobalState) ChanLatestNatsMessage(ch chan *nats.Msg) (*nats.Subscription, error) {
	mysub := gs.Cfg.Plc.Base + ".>"
	log.Println("get channel latest nats message for subject:", mysub)
	return gs.ChanSubscribeWrapper(mysub, ch)
}
