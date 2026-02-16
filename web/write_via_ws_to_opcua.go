package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strunck/awculle_datahub/globals"

	"github.com/awcullen/opcua/ua"
)

/*
Einschränkung:
Es wird immer nur ein Wert zur Zeit geschrieben.
Der zu schreibende Wert steht in "value".
Eine ggf. nötige Umrechnung soll vorher in Svelte stattfinden.

Ob es geklappt hat, wertet OPCUA aus. Dort steht im payload und topic, ob es geklappt hat und wenn nein, warum nicht.

Der Index Parameter ist nur bei Arraytypen anzugeben.

Beispiel für Arraytyp:
`{
    "value": 600,
    "channel": "PA_Rmm11_SollTemp_ari",
    "group": "PR",
    "index": 0,
    "datatype": "Int16"
}`

Beispiel für Scalaren Typ:
`{
    "value": 100,
    "channel": "PA_S01_MaxWarteKuehl_w",
    "group": "P",
    "datatype": "UInt16"
}`


`{
    "value": 600,
    "channel": "PA_Rmm11_SollTemp_ari",
    "group": "PR",
    "index": 0,
    "datatype": "Int16"
}`


{ "value": 600, "channel": "PA_Rmm11_SollTemp_ari", "group": "PR", "index": 0, "datatype": "Int16" }
{ "value": 150, "channel": "PA_Rmm11_SollTemp_ari", "group": "PR", "index": 0, "datatype": "Int16" }


Array index beginnt mit 0, OPC UA aber mit 1!!!

*/

type WriteReq struct {
	Value    any    `json:"value"`
	Channel  string `json:"channel"`
	Group    string `json:"group"`
	Index    int    `json:"index,omitempty"`
	DataType string `json:"datatype"`
}

func WSListenWrite(gs *globals.GlobalState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("websocket read error:", err)
				break
			}
			wr, err := NewWriteReq(message)
			if err != nil {
				log.Println("Invalid write request:", err)
				continue
			}

			wval, err := wr.valiadate(gs)
			if err != nil {
				log.Println("Validation error:", err)
				continue
			}

			req := ua.WriteRequest{}
			req.NodesToWrite = append(req.NodesToWrite, *wval)

			res, err := gs.OpcClient.Write(r.Context(), &req)
			if err != nil {
				log.Println("OPCUA write error:", err)
				continue
			}
			log.Printf("wrote value: %v to NodeId %s .OPCUA write status: %v", wval.Value.Value, wval.NodeID, res.Results[0])

		}
	}
}

func NewWriteReq(msg []byte) (*WriteReq, error) {
	var wr WriteReq
	err := json.Unmarshal(msg, &wr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse this '%s' as a write request: %w", msg, err)
	}
	return &wr, nil
}

func (wr WriteReq) valiadate(gs *globals.GlobalState) (wval *ua.WriteValue, err error) {

	// Check, if channel exists
	sym := wr.Group + "." + wr.Channel
	oc, ok := gs.Plc.Flat[sym]
	if !ok {
		return nil, fmt.Errorf("symbol %s does not exist", sym)
	}

	if oc.Length <= wr.Index {
		return nil, fmt.Errorf("index %d out of range for symbol %s with length %d", wr.Index, sym, oc.Length)
	}

	if !oc.Scalar {
		sym = sym + "[" + strconv.Itoa(wr.Index+1) + "]"
	}

	log.Printf("OPCUA Symbol to write: %+v", oc)

	nidstr := gs.Cfg.Opcua.Prefix + sym
	log.Printf("OPCUA NodeID to write: %s", nidstr)

	nid := ua.ParseNodeID(nidstr)

	// Create NodeID
	wval = &ua.WriteValue{
		NodeID:      nid,
		AttributeID: ua.AttributeIDValue,
	}

	// Convert value based on DataType
	convertedValue, err := wr.convertValue()
	if err != nil {
		return nil, fmt.Errorf("value conversion error: %w", err)
	}

	wval.Value.Value = convertedValue

	return wval, nil

}

// Add this method to convert values based on DataType
func (wr WriteReq) convertValue() (interface{}, error) {
	switch wr.DataType {
	case "Int16":
		if val, ok := wr.Value.(float64); ok {
			return int16(val), nil
		}
		return nil, fmt.Errorf("cannot convert %v to Int16", wr.Value)

	case "UInt16":
		if val, ok := wr.Value.(float64); ok {
			return uint16(val), nil
		}
		return nil, fmt.Errorf("cannot convert %v to UInt16", wr.Value)

	case "Int32":
		if val, ok := wr.Value.(float64); ok {
			return int32(val), nil
		}
		return nil, fmt.Errorf("cannot convert %v to Int32", wr.Value)

	case "UInt32":
		if val, ok := wr.Value.(float64); ok {
			return uint32(val), nil
		}
		return nil, fmt.Errorf("cannot convert %v to UInt32", wr.Value)

	case "Float":
		if val, ok := wr.Value.(float64); ok {
			return float32(val), nil
		}
		return nil, fmt.Errorf("cannot convert %v to Float", wr.Value)

	case "Double":
		if val, ok := wr.Value.(float64); ok {
			return val, nil
		}
		return nil, fmt.Errorf("cannot convert %v to Double", wr.Value)

	case "Boolean":
		if val, ok := wr.Value.(bool); ok {
			return val, nil
		}
		return nil, fmt.Errorf("cannot convert %v to Boolean", wr.Value)

	case "String":
		if val, ok := wr.Value.(string); ok {
			return val, nil
		}
		return nil, fmt.Errorf("cannot convert %v to String", wr.Value)

	default:
		return nil, fmt.Errorf("unsupported data type: %s", wr.DataType)
	}
}
