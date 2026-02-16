package globals

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"strunck/awculle_datahub/initialize"
	"time"
)

type DataV2 struct {
	Scope map[string]map[string]map[string]DataLeaf
}

type DataLeaf struct {
	Origin string    `json:"origin"`
	T      int64     `json:"t"`
	V      DataValue `json:"v"`
}

type DataValue struct {
	IsBool      bool
	BoolValue   bool   `json:"-"`
	StringValue string `json:"-"`
}

func (d DataV2) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Scope)
}

func (d *DataV2) ToJson() []byte {

	res, err := json.Marshal(d)
	if err != nil {
		log.Println("Error marshalling DataV2 to JSON:", err)
	}
	return res
}

func (v DataValue) MarshalJSON() ([]byte, error) {
	if v.IsBool {
		return json.Marshal(v.BoolValue)
	}
	return json.Marshal(v.StringValue)
}

// Setvalue and timestamp of DataV2 structure, create maps if needed
func (d *DataV2) SetLeafValue(v initialize.VariantToMessage, sval string) (err error) {

	leaf := DataLeaf{
		Origin: v.FlatRef,
		T:      time.Now().UnixMilli(),
	}
	isBinary := v.Type == "BOOL"
	if isBinary {
		leaf.V.IsBool = true
		if sval == "true" {
			leaf.V.BoolValue = true
		} else {
			leaf.V.BoolValue = false
		}
	} else {
		leaf.V.StringValue = sval
	}

	// Initialize root map if nil
	if d.Scope == nil {
		d.Scope = make(map[string]map[string]map[string]DataLeaf)
	}

	// does Scoper exist?
	if _, ok := d.Scope[v.Scoper]; !ok {

		d.Scope[v.Scoper] = make(map[string]map[string]DataLeaf)
	}
	// does Symbol exist?
	if _, ok := d.Scope[v.Scoper][v.Group]; !ok {

		d.Scope[v.Scoper][v.Group] = make(map[string]DataLeaf)
	}

	d.Scope[v.Scoper][v.Group][v.Short] = leaf

	return nil
}

// Gets a symbol from DataV2 (e.g. scope=R01 group=BA skey=Temp)
func (d *DataV2) GetSymbol(scope string, group string, short string) (*DataLeaf, error) {
	scopeMap, ok := d.Scope[scope]
	if !ok {
		return nil, fmt.Errorf("scope not found: %s", scope)
	}
	groupMap, ok := scopeMap[group]
	if !ok {
		return nil, fmt.Errorf("group not found: %s", group)
	}
	leaf, ok := groupMap[short]
	if !ok {
		return nil, fmt.Errorf("short not found: %s", short)
	}
	return &leaf, nil
}

// One may also pass a full key like "R01-BA-Temp"
func (d *DataV2) GetSymbolKey(skey string) (*DataLeaf, error) {
	parts := strings.Split(skey, "-")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid skey format: %s", skey)
	}
	if len(parts[0]) == 0 {
		return nil, fmt.Errorf("empty scope key: %s", skey)
	}
	if len(parts[1]) == 0 {
		return nil, fmt.Errorf("empty group key: %s", skey)
	}
	if len(parts[2]) == 0 {
		return nil, fmt.Errorf("empty short key: %s", skey)
	}
	scope := parts[0]
	group := parts[1]
	short := parts[2]
	return d.GetSymbol(scope, group, short)
}

func (d *DataV2) GetScopeAll(scope string) []byte {
	scopeMap, ok := d.Scope[scope]
	if !ok {
		return []byte("{}")
	}
	res, err := json.Marshal(scopeMap)
	if err != nil {
		log.Println("Error marshalling scope to JSON:", err)
	}
	return res
}

func (d *DataV2) GetGroupAll(scope string, group string) []byte {
	scopeMap, ok := d.Scope[scope]
	if !ok {
		return []byte("{}")
	}
	groupMap, ok := scopeMap[group]
	if !ok {
		return []byte("{}")
	}
	res, err := json.Marshal(groupMap)
	if err != nil {
		log.Println("Error marshalling group to JSON:", err)
	}
	return res
}

func (d *DataV2) GetShort(scope string, group string, short string) []byte {
	scopeMap, ok := d.Scope[scope]
	if !ok {
		return []byte("{}")
	}
	groupMap, ok := scopeMap[group]
	if !ok {
		return []byte("{}")
	}
	leaf, ok := groupMap[short]
	if !ok {
		return []byte("{}")
	}
	res, err := json.Marshal(leaf)
	if err != nil {
		log.Println("Error marshalling leaf to JSON:", err)
	}
	return res
}
