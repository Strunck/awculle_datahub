package globals

import (
	"encoding/json"
	"fmt"
	"log"
)

// HMIData represents the complete HMI data structure
type HMIData map[string]HMIVariableTyped

// Alternative with more specific value types
type HMIVariableTyped struct {
	Group    string    `json:"group"`
	Name     string    `json:"name"`
	Scalar   bool      `json:"scalar"`
	Length   int       `json:"length"`
	Type     string    `json:"type"`
	IsNumber bool      `json:"isNumber"`
	Scale    int       `json:"scale"`
	Age      int       `json:"age"`
	Values   HMIValues `json:"values"`
}

// HMIValues handles the different value types
type HMIValues struct {
	marshalType string
	BoolValue   bool              `json:"-"`
	StringValue string            `json:"-"`
	ArrayBool   map[string]bool   `json:"-"`
	ArrayString map[string]string `json:"-"`
}

func (h *HMIVariableTyped) SetStringValue(sval string, pos int) (err error) {

	if h.Scalar {
		if h.Type == "BOOL" {
			if sval == "true" {
				h.Values.BoolValue = true
			} else {
				h.Values.BoolValue = false
			}
		} else {
			h.Values.StringValue = sval
		}
	} else {
		p := fmt.Sprintf("%d", pos)

		if h.Type == "BOOL" {
			_, ok := h.Values.ArrayBool[p]
			if !ok {
				return fmt.Errorf("position %s not found in ArrayBool map for %s.%s", p, h.Group, h.Name)
			}
			if sval == "true" {
				h.Values.ArrayBool[p] = true
			} else {
				h.Values.ArrayBool[p] = false
			}
		} else {
			_, ok := h.Values.ArrayString[p]
			if !ok {
				return fmt.Errorf("position %s not found in ArrayString map for %s.%s", p, h.Group, h.Name)
			}
			h.Values.ArrayString[p] = sval
		}
	}

	return nil

}

// InitFlatHMIvalues initializes a map, that will be transered via Websocket, to be used in the HMI
// I need some blank values in the Init Phase, that will be filed later
func (gs *GlobalState) InitFlatHMIvalues() {
	hmiDataMap := make(map[string]HMIVariableTyped)

	for _, t := range gs.Cfg.Nats.Topics {
		// Create a new HMIVariableTyped
		hmiVar := HMIVariableTyped{
			Group:    t.Group,
			Name:     t.Name,
			Scalar:   !t.IsArray,
			Length:   int(t.Length),
			Type:     t.Vartype,
			IsNumber: isNumberType(t.Vartype),
			Scale:    int(t.Scale),
			Age:      0,
		}
		key := fmt.Sprintf("%s.%s", t.Group, t.Symbol)
		if len(key) == 0 {
			log.Printf("Error: key is empty for topic: %+v\n", t)
		}

		if t.IsArray {
			if t.Vartype == "BOOL" {
				hmiVar.Values.marshalType = "boolarray"
				hmiVar.Values.ArrayBool = make(map[string]bool)
			} else {
				hmiVar.Values.marshalType = "stringarray"
				hmiVar.Values.ArrayString = make(map[string]string)
			}

			for i := 0; i < int(t.Length); i++ {
				valkey := fmt.Sprintf("%d", i)
				if t.Vartype == "BOOL" {
					hmiVar.Values.ArrayBool[valkey] = false
				} else {
					hmiVar.Values.ArrayString[valkey] = "---"
				}
			}
		} else {
			if t.Vartype == "BOOL" {
				hmiVar.Values.marshalType = "bool"
				hmiVar.Values.BoolValue = false
			} else {
				hmiVar.Values.marshalType = "string"
				hmiVar.Values.StringValue = "---"
			}
		}

		hmiDataMap[key] = hmiVar

	}
	gs.Plc.Flat = hmiDataMap

}

func isNumberType(vartype string) bool {
	switch vartype {
	case "REAL", "LREAL", "INT", "WORD", "DWORD", "BYTE":
		return true
	default:
		return false
	}
}

// UnmarshalJSON custom unmarshaling for HMIValues
func (v *HMIValues) UnmarshalJSON(data []byte) error {

	switch v.marshalType {
	case "bool":
		var boolVal bool
		if err := json.Unmarshal(data, &boolVal); err != nil {
			return err
		}
		v.BoolValue = boolVal
		return nil
	case "string":
		var stringVal string
		if err := json.Unmarshal(data, &stringVal); err != nil {
			return err
		}
		v.StringValue = stringVal
		return nil
	case "boolarray":
		var mapVal map[string]bool
		if err := json.Unmarshal(data, &mapVal); err != nil {
			return err
		}
		v.ArrayBool = mapVal
		return nil
	case "stringarray":
		var mapVal map[string]string
		if err := json.Unmarshal(data, &mapVal); err != nil {
			return err
		}
		v.ArrayString = mapVal
		return nil
	}

	return fmt.Errorf("unable to unmarshal HMIValues")
}

// MarshalJSON custom marshaling for HMIValues
func (v HMIValues) MarshalJSON() ([]byte, error) {
	switch v.marshalType {
	case "bool":
		return json.Marshal(v.BoolValue)
	case "string":
		return json.Marshal(v.StringValue)
	case "boolarray":
		return json.Marshal(v.ArrayBool)
	case "stringarray":
		return json.Marshal(v.ArrayString)
	}

	return []byte("null"), nil
}

func (hv HMIVariableTyped) ToJson() []byte {
	res, err := json.MarshalIndent(hv, "", "  ")
	if err != nil {
		log.Println("Error marshalling HMIVariableTyped to JSON:", err)
	}
	return res
}

func (hd HMIData) ToJson() []byte {
	res, err := json.MarshalIndent(hd, "", "  ")
	if err != nil {
		log.Println("Error marshalling HMIData to JSON:", err)
	}
	return res
}
