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

// Initialize OPC UA Montitorung. State REQ_MONITORING in main.go
func (gs *GlobalState) InitMonitorOpcUa(pctx context.Context) (err error) {

	// prepare create subscription request
	subreq := &ua.CreateSubscriptionRequest{
		RequestedPublishingInterval: 1000.0,
		RequestedMaxKeepAliveCount:  30,
		RequestedLifetimeCount:      30 * 3,
		PublishingEnabled:           true,
	}

	// send request to server. receive response or error
	subres, err := gs.OpcUa.Client.CreateSubscription(pctx, subreq)
	if err != nil {
		err = fmt.Errorf("🔱 Error creating subscription: %w", err)
		return err
	}

	if subres.ServiceResult.IsGood() {
		gs.OpcUa.SupscriptId = subres.SubscriptionID
		log.Printf("🔱 OPC UA Subscription created with ID: %d\n", subres.SubscriptionID)
	} else {
		return fmt.Errorf("🔱 OPC UA Subscription creation failed with status: %v", subres.ServiceResult)
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
	gs.OpcUa.ItemIds = gs.OpcUa.ItemIds[:0]
	totalCreated := 0
	for start := 0; start < len(mimi); start += OPCUA_MONITOR_BATCH {
		end := start + OPCUA_MONITOR_BATCH
		if end > len(mimi) {
			end = len(mimi)
		}

		monReq := &ua.CreateMonitoredItemsRequest{
			SubscriptionID:     subres.SubscriptionID,
			TimestampsToReturn: ua.TimestampsToReturnBoth,
			ItemsToCreate:      mimi[start:end],
		}

		// send request to server. receive response or error
		monRes, err := gs.OpcUa.Client.CreateMonitoredItems(pctx, monReq)
		if err != nil {
			return fmt.Errorf("🔱 error OpcClient.CreateMonitoredItems batch %d-%d: %w", start, end-1, err)
		}

		if !monRes.ServiceResult.IsGood() {
			return fmt.Errorf("🔱 OPC UA monitored item creation failed for batch %d-%d with status: %v", start, end-1, monRes.ServiceResult)
		}

		for idx, res := range monRes.Results {
			if !res.StatusCode.IsGood() {
				midi := mimi[start+idx].ItemToMonitor.NodeID
				fmt.Printf("🔱 OPC UA monitored item %v creation failed for client handle %d with status: %v\n", midi, mimi[start+idx].RequestedParameters.ClientHandle, res.StatusCode)
				// return fmt.Errorf("🔱 OPC UA monitored item %v creation failed for client handle %d with status: %v", midi, mimi[start+idx].RequestedParameters.ClientHandle, res.StatusCode)
			}
			gs.OpcUa.ItemIds = append(gs.OpcUa.ItemIds, res.MonitoredItemID)
		}

		totalCreated += len(monRes.Results)
	}

	log.Printf("🔱 %d Monitored Items created in batches of %d", totalCreated, OPCUA_MONITOR_BATCH)

	return nil
}

// RecieveOpcUaMessage listens to DataChangeNotifications (DCN) from the OPC UA server.
// When a new DCN is recieved, the Method storeAndPublish is called
func (gs *GlobalState) RecieveOpcUaMessage(ctx context.Context, errCh chan<- error) {

	init := true

	// Send initial publish request
	req := &ua.PublishRequest{
		RequestHeader:                ua.RequestHeader{TimeoutHint: 60000},
		SubscriptionAcknowledgements: []ua.SubscriptionAcknowledgement{},
	}

	go func() {
		// Loop to receive notifications
		for {
			// Send publish request
			res, err := gs.OpcUa.Client.Publish(ctx, req)
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
	}()

	if init {
		log.Println("RecieveOpcUaMessage started")
		init = false
	} else {
		log.Println("RecieveOpcUaMessage quit routine")
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
		//log.Printf("%s: Gauge.Set %s = %f ", p.NodeID, p.Metric, fval)
		p.Gauge.Set(fval)
	}

	// Publish to NATS
	err = gs.Nats.Client.Publish(p.Topic, []byte(sval))

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
	mh.Ok = mh.AgeMs < READWAIT_INC_MAX.Milliseconds()
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
