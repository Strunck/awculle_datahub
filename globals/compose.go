package globals

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/awcullen/opcua/ua"
)

func makeNodeHelperArray(count int, attr string, typ string, name string, prefix string) (res []NodeHelper) {

	for i := 1; i <= count; i++ {
		h := NodeHelper{
			Name:  fmt.Sprintf("%s[%d].%s", name, i, attr),
			Type:  typ,
			Key:   fmt.Sprintf("%s_%s[%d]", name, attr, i),
			Index: i - 1,
		}
		h.NodeId = ua.ParseNodeID(prefix + "." + h.Name)
		if h.NodeId == nil {
			log.Printf("🪑 invalid node id of %s", h.Name)
		}
		res = append(res, h)
	}

	return res
}

func ComposeExtensionObject(prefix string, n Node) (res []NodeHelper, err error) {

	switch n.Type {

	case "T_ARRAY__1__24__OF_TYPE_EMeterKonfig":
		res = append(res, makeNodeHelperArray(24, "xVorhanden", "T_BOOL", "EMeter", prefix)...)
		res = append(res, makeNodeHelperArray(24, "rPulswertigkeit", "T_REAL", "EMeter", prefix)...)
		res = append(res, makeNodeHelperArray(24, "EMeterType", "T_EMT", "EMeter", prefix)...)
		res = append(res, makeNodeHelperArray(24, "bOrt", "T_BYTE", "EMeter", prefix)...)
		res = append(res, makeNodeHelperArray(24, "sShort", "T_STRING", "EMeter", prefix)...)
		res = append(res, makeNodeHelperArray(24, "sDescr", "T_STRING", "EMeter", prefix)...)

	case "T_ARRAY__1__40__OF_TYPE_Rmm":
		res = nil
		err = errors.New(n.Name + " T_ARRAY__1__40__OF_TYPE_Rmm not implemented")

	case "T_ARRAY__1__40__OF_PVMOD":
		res = nil
		err = errors.New(n.Name + " T_ARRAY__1__40__OF_TYPE_PVMOD not implemented")

	case "T_ARRAY__1__20__OF_TYPE_Warnung":
		res = nil
		err = errors.New(n.Name + " T_ARRAY__1__20__OF_TYPE_Warnung not implemented")

	case "T_ARRAY__1__100__OF_TYPE_Verbraucher":
		res = nil
		err = errors.New(n.Name + " T_ARRAY__1__100__OF_TYPE_Verbraucher not implemented")

	case "T_ARRAY__1__14__OF_struct_Messwerte":
		res = nil
		err = errors.New(n.Name + " T_ARRAY__1__14__OF_struct_Messwerte not implemented")

	case "T_ARRAY__1__160__OF_TYPE_Luefterkonfig":
		res = nil
		err = errors.New("T_ARRAY__1__160__OF_TYPE_Luefterkonfig not implemented")

	case "T_ARRAY__1__2__OF_Struct_ArrOfPoints":
		res = nil
		err = errors.New(n.Name + " T_ARRAY__1__2__OF_Struct_ArrOfPoints not implemented")

	case "T_ARRAY__1__6__OF_STRUCT_VerdichterVisu":
		res = append(res, makeNodeHelperArray(6, "b_Verdichter1", "T_BYTE", "Gar_AggZuVerd", prefix)...)
		res = append(res, makeNodeHelperArray(6, "b_Verdichter2", "T_BYTE", "Gar_AggZuVerd", prefix)...)
		res = append(res, makeNodeHelperArray(6, "b_Verdichter3", "T_BYTE", "Gar_AggZuVerd", prefix)...)
		res = append(res, makeNodeHelperArray(6, "b_Verdichter4", "T_BYTE", "Gar_AggZuVerd", prefix)...)
		res = append(res, makeNodeHelperArray(6, "b_Verdichter5", "T_BYTE", "Gar_AggZuVerd", prefix)...)
		res = append(res, makeNodeHelperArray(6, "b_Verdichter6", "T_BYTE", "Gar_AggZuVerd", prefix)...)
		res = append(res, makeNodeHelperArray(6, "b_Verdichter7", "T_BYTE", "Gar_AggZuVerd", prefix)...)
		res = append(res, makeNodeHelperArray(6, "b_Verdichter8", "T_BYTE", "Gar_AggZuVerd", prefix)...)
		res = append(res, makeNodeHelperArray(6, "b_Verdichter9", "T_BYTE", "Gar_AggZuVerd", prefix)...)
		res = append(res, makeNodeHelperArray(6, "b_Verdichter10", "T_BYTE", "Gar_AggZuVerd", prefix)...)

	case "T_ARRAY__1__6__OF_STRUCT_VerflussigerVisu":
		res = nil
		err = errors.New(n.Name + " T_ARRAY__1__6__OF_STRUCT_VerflussigerVisu not implemented")

	default:
		lf := NodeHelper{
			Name:    n.Name,
			Type:    n.Type,
			Comment: n.Comment,
			Key:     n.Name,
		}
		lf.NodeId = ua.ParseNodeID(prefix + "." + n.Name)
		if lf.NodeId == nil {
			log.Printf("invalid node id of %s: %v", lf.Name, err)
		} else {
			res = append(res, lf)
		}
	}
	if len(res) == 0 {
		//err = fmt.Errorf("no nodes found for %s", n.Name)
		return nil, err
	} else {
		return res, nil
	}

}

// Aggregate Plc.Konfig so I have nice Arrays
func (cfg *Config) ComposeArrayFromObjects() (err error) {
	type tempArray struct {
		Oldkey string
		Index  int
		Value  any
	}

	m := make(map[string][]tempArray)
	// Iterate over .Plc.Konfig
	for k, v := range cfg.Plc.Konfig {
		// if key end with a ] it is part of an array
		if strings.HasSuffix(k, "]") {
			// all letters before [ are a new key in the map m
			sp := strings.Split(k, "[")

			if len(sp) == 2 {

				// the number between [] is the index +1
				index, err := strconv.Atoi(strings.TrimSuffix(sp[1], "]"))
				if err != nil {
					return fmt.Errorf("invalid index, String to Int for %v array index in key %s: %v", sp[1], k, err)
				}
				if index <= 0 {
					return fmt.Errorf("invalid index should be bigger than 0, but we have %d in key %s", index, k)
				}

				m[sp[0]] = append(m[sp[0]], tempArray{
					Oldkey: k,
					Index:  index - 1,
					Value:  v,
				})

				// delete a key
				delete(cfg.Plc.Konfig, k)

			} else {
				return fmt.Errorf("invalid array key %s", k)
			}

		}

	}

	// add the agregated arrays back to the config
	for k, liste := range m {
		arr := make([]any, len(liste))
		for _, e := range liste {
			arr[e.Index] = e.Value
		}
		cfg.Plc.Konfig[k] = arr
	}

	// Change Keys to a better Name
	keylist := map[string]string{
		"EMeter_EMeterType":      "EMeterType",
		"EMeter_bOrt":            "EMeterOrt",
		"EMeter_rPulswertigkeit": "EMeterPulswertigkeit",
		"EMeter_sDescr":          "EMeterDescr",
		"EMeter_sShort":          "EMeterShort",
		"EMeter_xVorhanden":      "EMeterVorhanden",
	}

	for k, v := range keylist {
		cfg.Plc.Konfig[v] = cfg.Plc.Konfig[k]
		delete(cfg.Plc.Konfig, k)
	}

	// Add Gb_AnzahlEMeter which counts the true values in EMeterVorhanden
	count := 0
	emeter, ok := cfg.Plc.Konfig["EMeterVorhanden"]
	emeterType, okType := emeter.([]any)

	if !okType {
		log.Printf("🪑 EMeterVorhanden is not an array, but %T", emeter)

	} else if len(emeterType) != 24 {
		log.Printf("🪑 EMeterVorhanden has %d elements, but expected 24", len(emeterType))

	} else if !ok {
		log.Printf("🪑 EMeterVorhanden not found in config")

	} else {
		for _, e := range emeter.([]any) {
			elemType, elemTok := e.(bool)
			if !elemTok {
				log.Printf("🪑 EMeterVorhanden element is not a bool, but %T", e)
			} else if elemType {
				count++
			}

		}
	}
	cfg.Plc.Konfig["Gb_AnzahlEMeter"] = count

	return nil
}
