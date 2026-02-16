package initialize

import (
	"context"
	"errors"
	"fmt"
	"log"

	"time"

	"github.com/awcullen/opcua/client"
	"github.com/awcullen/opcua/ua"

	"github.com/prometheus/client_golang/prometheus"
)

// ReadKonfigViaOpcUa reads site specific hardware configuration.
// First connect to the opcua server from cfg.OpcUa.Url
// Second read the amount of units
func (cfg *Config) ReadKonfigViaOpcUa(timeout time.Duration, c *client.Client, bgctx context.Context) (err error) {

	ctx, cancel := context.WithTimeout(bgctx, timeout)
	defer cancel()

	dimNodes := map[string]string{
		"A": "Konfig.Gb_Anzahl_Aggregate",
		"C": "Konfig.Gb_Anzahl_Verd",
		"D": "Konfig.Gb_Anzahl_Scrubberkreis",
		"K": "Konfig.Gb_Anzahl_Verf",
		"M": "Konfig.Gb_Anzahl_Messkreis",
		"R": "Konfig.Gb_Anzahl_Raeume",
		"V": "Konfig.Gb_AnzVentilatoren",
	}
	dim := map[string]DimScopeArray{}

	for dk, nodeIDpart := range dimNodes {

		nodeID := cfg.Opcua.Prefix + nodeIDpart

		id := ua.ParseNodeID(nodeID)

		if id == nil {
			log.Fatalf("invalid node id: %v", err)
		}

		req := &ua.ReadRequest{
			MaxAge: 2000,
			NodesToRead: []ua.ReadValueID{
				{
					NodeID:      id,
					AttributeID: ua.AttributeIDValue,
				},
			},
			TimestampsToReturn: ua.TimestampsToReturnBoth,
		}

		resp, err := c.Read(ctx, req)
		if err != nil {
			log.Fatalf("Read failed: %s", err)
		}

		if !resp.Results[0].StatusCode.IsGood() {
			log.Fatalf("Status not OK for %s: %v", nodeID, resp.Results[0].StatusCode.Error())
		}

		value := resp.Results[0].Value.(byte)

		log.Printf("%d:\t%s", value, nodeIDpart)
		scopearray := make([]string, value)
		for i := 0; i < int(value); i++ {
			switch dk {
			case "A", "D", "M":
				scopearray[i] = fmt.Sprintf("%s%d", dk, i+1)
			case "C", "K", "R", "J", "E":
				scopearray[i] = fmt.Sprintf("%s%02d", dk, i+1)
			case "V":
				scopearray[i] = fmt.Sprintf("%s%03d", dk, i+1)
			}
		}

		dim[dk] = DimScopeArray{
			Length:  uint8(value),
			Scopers: scopearray,
		}

	}
	dim["S"] = DimScopeArray{
		Length:  1,
		Scopers: []string{"SYS"},
	}
	cfg.Plc.DimMap = dim

	err = cfg.transferDimensons()
	if err != nil {
		log.Fatalf("failed to fill topic map: %s", err.Error())
	}

	err = cfg.makeVariantToMessage()
	if err != nil {
		log.Fatalf("failed to prepare the list of node id's for the Opcua Subcribtions: %s", err.Error())
	}

	return err
}

// transferDimensons uses .Nats.Topics[] and .Plc.DimMap to set the amount (length) of the configured plc scopes(loops)
func (cfg *Config) transferDimensons() (err error) {

	i := 0
	for _, t := range cfg.Nats.Topics {

		// get the amount of the configured plc scopes(loops)
		loop, ok := cfg.Plc.DimMap[t.Scope]
		if !ok {
			log.Printf("'%s' with scope %s is not Config.Plc.DimMap", t.Short, t.Scope)
		}
		cfg.Nats.Topics[i].Length = loop.Length
		i++
	}
	if i == 0 {
		return errors.New("no topics found in NATS Config")
	}

	return nil
}

// makeVariantToMessage is a method to save a list of node id's to .Opcua.ArrayNodes[],
// which will then be used as parameter
// for the Subcribe method of monitor.NodeMonitor
func (cfg *Config) makeVariantToMessage() (err error) {

	mmap := make(map[uint32]VariantToMessage) // Maps the NodeID from the Subscribed message to a NATS message and a metric

	var total uint32 = 0
	for _, t := range cfg.Nats.Topics {
		scope := cfg.Plc.DimMap[t.Scope]

		if t.IsArray {
			for i := 0; i < int(scope.Length); i++ {

				mmap[total] = VariantToMessage{
					NodeID:  fmt.Sprintf("%s[%d]", t.Opcnode, i+1),
					Topic:   fmt.Sprintf("%s.%s.%s.%s", cfg.Plc.Base, scope.Scopers[i], t.Group, t.Short),
					FlatRef: fmt.Sprintf("%s.%s", t.Group, t.Name),
					Scale:   t.Scale,
					Type:    t.Vartype,
					Metric:  t.Metric,
					Ort:     i + 1,
					IsArray: true,
					Scope:   t.Scope,
					Scoper:  scope.Scopers[i],
					Group:   t.Symbol[0:2],
					Short:   t.Short,
					Symbol:  t.Symbol,
					Gauge: prometheus.NewGauge(prometheus.GaugeOpts{
						Name: t.Metric,
						Help: cfg.Metrics.Results[t.Metric].Help,
						ConstLabels: prometheus.Labels{
							"base":   cfg.Plc.Base,
							"descr":  t.Descr,
							"symbol": t.Symbol,
							"short":  t.Short,
							"ort":    fmt.Sprintf("%d", i+1),
						},
					}),
				}
				total++
			}
		} else {

			mmap[total] = VariantToMessage{
				NodeID:  t.Opcnode,
				Topic:   fmt.Sprintf("%s.SYS.%s.%s", cfg.Plc.Base, t.Group, t.Short),
				FlatRef: fmt.Sprintf("%s.%s", t.Group, t.Name),
				Scale:   t.Scale,
				Type:    t.Vartype,
				Metric:  t.Metric,
				Ort:     0,
				IsArray: false,
				Scope:   t.Scope,
				Scoper:  "S",
				Group:   t.Symbol[0:2],
				Short:   t.Short,
				Symbol:  t.Symbol,
				Gauge: prometheus.NewGauge(prometheus.GaugeOpts{
					Name: t.Metric,
					Help: cfg.Metrics.Results[t.Metric].Help,
					ConstLabels: prometheus.Labels{
						"base":   cfg.Plc.Base,
						"descr":  t.Descr,
						"symbol": t.Symbol,
						"short":  t.Short,
					},
				}),
			}
			total++
		}
	}

	cfg.Plc.Publish = mmap

	cfg.Plc.Stats.TotalSubcriptions = total

	return nil
}

// SetupKonfigGroup read all NodeID, that are leafs of the NodeId with the name Konfig
func (cfg *Config) SetupKonfigGroup(client *client.Client, bgctx context.Context) (err error) {
	log.Println("------------  SetupKonfigGroup  -------------------")

	var nodeIdList []NodeHelper

	prefix := cfg.Opcua.Prefix + "Konfig"
	// Get all the leaf nodes from .cfg.Opcua.Symbols.NodeList.Nodes

	for _, grp_node := range cfg.Opcua.Symbole.NodeList.Node.Node {
		if grp_node.Name == "Konfig" {
			for _, leaf_node := range grp_node.Node {
				n := Node{
					Name:    leaf_node.Name,
					Type:    leaf_node.Type,
					Comment: leaf_node.Comment,
				}

				lf, err := ComposeExtensionObject(prefix, n)
				if err != nil {
					log.Printf("ComposeExtensionObject Error: %v\n", err)
				} else {
					//	log.Printf("Append: %v\n%v\n", leaf_node.Name, lf)
					nodeIdList = append(nodeIdList, lf...)
				}
			}
		}
	}

	log.Println("NodeList is created. Found", len(nodeIdList))

	var opcReadReqList []*ua.ReadRequest

	const MAXREAD = 100

	j := 0

	//var HundredNodeIds [MAXREAD]*ua.ReadValueID
	HundredNodeIds := make([]ua.ReadValueID, MAXREAD)

	nodeListErrCnt := 0
	for i, n := range nodeIdList {

		if (i > 0 && i%MAXREAD == 0) || (i == len(nodeIdList)-1) {
			opcReadReqList = append(opcReadReqList, &ua.ReadRequest{
				MaxAge:             2000,
				TimestampsToReturn: ua.TimestampsToReturnBoth,
				NodesToRead:        HundredNodeIds[:j],
			})
			//log.Printf("DEBUG Append NodeIds: %d\n", len(HundredNodeIds))
			j = 0
			HundredNodeIds = make([]ua.ReadValueID, MAXREAD) // Reset the slice for the next 100 NodeIds
		}

		if len(n.Name) > 0 {
			HundredNodeIds[j] = ua.ReadValueID{
				NodeID:      n.NodeId,
				AttributeID: ua.AttributeIDValue,
			}
			j++

		} else {
			log.Printf("WARN keine NodeID %s\t%v\n", n.Name, n)
			nodeListErrCnt++
		}

	}

	// Just for Debugging
	if nodeListErrCnt > 0 {
		log.Printf("WARNING %d NodeId could not be created\n", nodeListErrCnt)
	}

	// New Context for Reading OPC UA
	ctx, cancel := context.WithCancel(bgctx)
	defer cancel()

	m := make(map[string]any)

	readlistCount := len(opcReadReqList)
	idx := 0
	for i, readlist := range opcReadReqList {
		log.Println("Read Sequence:", i+1, "of", readlistCount, "Start")

		readed, err := client.Read(ctx, readlist)

		if err != nil {
			log.Println("WARNING Did not read all Konfig Nodes", err)

		} else {
			log.Println("Read Sequence", i+1, "SUCCESS")

			for k, rr := range readed.Results {
				idx = k + (i * MAXREAD)
				if idx < len(nodeIdList) {
					n := nodeIdList[idx]
					//n.Vals = rr.Value
					//log.Printf("--------\n%3d %s\t%s\n%+v\n%#v\n--------\n", idx, n.Name, n.Type, readlist.NodesToRead[k].NodeID, rr.Value)
					m[n.Key] = fmt.Sprintf("%v", rr.Value)
					m[n.Key] = ConvertSymbol(rr.Value, n.Type)
				}
			}
		}

		// wait for 500 ms
		time.Sleep(200 * time.Millisecond)
	}

	cfg.Plc.Konfig = m

	// Aggregate Plc.Konfig
	err = cfg.ComposeArrayFromObjects()
	if err != nil {
		return fmt.Errorf("failed to compose arrays from Plc.Konfig: %w", err)
		// log.Fatalf("failed to compose arrays from Plc.Konfig: %s", err.Error())
	}
	log.Println("SetupKonfigGroup: Selfconfiguration from PLC successfully finished.")

	return err
}
