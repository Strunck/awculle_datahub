package globals

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/awcullen/opcua/ua"
)

// Initialize OPC UA Montitorung, by
func (gs *GlobalState) InitMonitorOpcUa(pctx context.Context) (err error) {

	// prepare create subscription request
	subreq := &ua.CreateSubscriptionRequest{
		RequestedPublishingInterval: 1000.0,
		RequestedMaxKeepAliveCount:  30,
		RequestedLifetimeCount:      30 * 3,
		PublishingEnabled:           true,
	}

	// send request to server. receive response or error
	subres, err := gs.OpcClient.CreateSubscription(pctx, subreq)
	if err != nil {
		err = fmt.Errorf("🔱 Error creating subscription: %w", err)
		return err
	}

	// List of monitored items, that have been configured in func
	mimi := make([]ua.MonitoredItemCreateRequest, gs.Cfg.Plc.Stats.TotalSubcriptions)
	for i, vtm := range gs.Cfg.Plc.Publish {

		mimi[i] = ua.MonitoredItemCreateRequest{
			ItemToMonitor: ua.ReadValueID{
				NodeID:      ua.ParseNodeID(vtm.NodeID),
				AttributeID: ua.AttributeIDValue,
			},
			MonitoringMode: ua.MonitoringModeReporting,
			// specify a unique ClientHandle. The ClientHandle is returned in the PublishResponse
			RequestedParameters: ua.MonitoringParameters{
				ClientHandle:     i,
				QueueSize:        1,
				DiscardOldest:    true,
				SamplingInterval: 1000.0,
			},
		}
	}

	// prepare create monitored items request
	monReq := &ua.CreateMonitoredItemsRequest{
		SubscriptionID:     subres.SubscriptionID,
		TimestampsToReturn: ua.TimestampsToReturnBoth,
		ItemsToCreate:      mimi,
	}

	// send request to server. receive response or error
	_, err = gs.OpcClient.CreateMonitoredItems(pctx, monReq)
	if err != nil {
		err = fmt.Errorf("🔱 error OpcClient.CreateMonitoredItems: %w", err)
		return err
	}
	return nil
}

// RecieveOpcUaMessage listens to DataChangeNotifications (DCN) from the OPC UA server.
// When a new DCN is recieved, the Method storeAndPublish is called
func (gs *GlobalState) RecieveOpcUaMessage(ctx context.Context, errCh chan<- error) {
	//log.Println("RecieveOpcUaMessage started")

	// Send initial publish request
	req := &ua.PublishRequest{
		RequestHeader:                ua.RequestHeader{TimeoutHint: 60000},
		SubscriptionAcknowledgements: []ua.SubscriptionAcknowledgement{},
	}

	// Loop to receive notifications
	for {
		// Send publish request
		res, err := gs.OpcClient.Publish(ctx, req)
		if err != nil {
			log.Printf("🔱 Error publishing request: %s\n", err.Error())
			errCh <- fmt.Errorf("🔱 Error publishing request: %w", err)
			return
		}

		// Process notification data
		for _, data := range res.NotificationMessage.NotificationData {
			switch body := data.(type) {
			case ua.DataChangeNotification:
				// Process data change notification
				for _, item := range body.MonitoredItems {

					if err := gs.storeAndPublish(item); err != nil {
						log.Printf("Error storing and publishing item: %s\n", err.Error())
						errCh <- fmt.Errorf("🔱 Error storing and publishing item: %w", err)
						return
					}
				}
			default:
				// Other notification types can be handled here
				log.Printf("Received unhandled notification type: %T\n", body)
			}
		}

		// Prepare next publish request with acknowledgement
		req = &ua.PublishRequest{
			RequestHeader: ua.RequestHeader{TimeoutHint: 60000},
			SubscriptionAcknowledgements: []ua.SubscriptionAcknowledgement{
				{
					SequenceNumber: res.NotificationMessage.SequenceNumber,
					SubscriptionID: res.SubscriptionID,
				},
			},
		}
	}
}

// Store Item in the global structure and Publish to NATS
func (gs *GlobalState) storeAndPublish(item ua.MonitoredItemNotification) (err error) {

	// get Handler
	p, ok := gs.Cfg.Plc.Publish[item.ClientHandle]
	if !ok {
		return fmt.Errorf("🔱 ClientHandle %d not found in Publish map", item.ClientHandle)
	}
	gs.Stats.OpcHealth.Update()

	// Convert Variant to string in US format, e.g. 12.34
	sval, err := p.ToString(item)
	if err != nil {
		return fmt.Errorf("🔱 Error converting value to string: %w", err)
	}

	// Publish to Prometheus, see file publish.go
	//gs.publishMetric(item, &p)
	if len(p.Metric) > 0 {
		//log.Printf("Published to Prometheus: %s = %s\n%s", p.Topic, sval, p.Gauge.Desc())
		var fval float64

		// Boolean special case
		if p.Type == "BOOL" {
			if sval == "true" {
				fval = 1
			} else {
				fval = 0
			}
		} else {

			fval, err = strconv.ParseFloat(sval, 64)
			if err != nil {
				return fmt.Errorf("🔱 NodeID %s could not convert value to float: %w", p.NodeID, err)
			}
		}
		p.Gauge.Set(fval)
	}

	// Publish to NATS
	err = gs.NatsClient.Publish(p.Topic, []byte(sval))

	if err != nil {
		log.Printf("failed to publish message: %s\nretrying...", err.Error())
		return fmt.Errorf("🔱 Error publishing to NATS: %w", err)
	}
	gs.Stats.NatsHealth.Update()

	// Store to flat map
	fr, ok := gs.Plc.Flat[p.FlatRef]
	if !ok {
		return fmt.Errorf("🔱 FlatRef %s not found in Flat map\nClientHandleNo: %d\nPublishConfig: %+v\n------------------------", p.FlatRef, item.ClientHandle, p)
	}
	fr.SetStringValue(sval, p.Ort-1)

	if err := gs.DataV2.SetLeafValue(p, sval); err != nil {
		return fmt.Errorf("🔱 Error setting DataV2 leaf value: %w", err)
	}

	return nil
}

type MsgHealth struct {
	AgeMs   int64
	Ok      bool
	LastMsg time.Time
}

func (mh *MsgHealth) calcAge() {
	now := time.Now()
	//log.Println("Now", now.Unix(), "Last", mh.LastMsg.Unix())
	mh.AgeMs = now.UnixMilli() - mh.LastMsg.UnixMilli()
	mh.Ok = mh.AgeMs < READMAXWAIT.Milliseconds()
}

func (mh *MsgHealth) Update() {
	//log.Println("Update last Message", mh)
	mh.LastMsg = time.Now()
}

func (mh *MsgHealth) ToJson() []byte {
	mh.calcAge()
	res, err := json.MarshalIndent(mh, "", "  ")
	if err != nil {
		log.Println("Error marshalling MsgHealth to JSON:", err)
	}
	return res
}
