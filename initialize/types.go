package initialize

import (
	"encoding/json"
	"log"

	"github.com/awcullen/opcua/ua"
	"github.com/prometheus/client_golang/prometheus"
)

type Config struct {
	Nats         NATSConfig    `json:"Nats"`
	Plc          PlcConfig     `json:"Plc"`
	Opcua        OpcUaConfig   `json:"Opcua"`
	Metrics      MetricsConfig `json:"Metrics"`
	KoDataPath   string        `json:"KoDataPath"`   // path to load settings.json and symbolxml
	SettingsFile string        `json:"SettingsFile"` // filename of the settings.json file
}

// Nats -----------------------------------------
type NATSConfig struct {
	Port     int        `json:"port"`
	Url      string     `json:"Url"`
	GroupIds []string   `json:"GroupIds"`
	Topics   []OPCtopic `json:"Topics"`
}

type OPCtopic struct {
	Natstopic string `json:"natstopic"`
	Short     string `json:"short"`
	Group     string `json:"group"`
	Opcnode   string `json:"opcnode"`
	Scale     int8   `json:"scale"`
	Length    uint8  `json:"length"`
	Scope     string `json:"scope"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	Vartype   string `json:"vartype"`
	IsArray   bool   `json:"isarray"`
	Konfig    string `json:"konfig"`
	Metric    string `json:"metric"`
	Symbol    string `json:"symbol"`
	Descr     string `json:"descr"`
}

// .Plc ---------------------------------------------------------------------------------------------------------------------------------
type PlcConfig struct {
	Base        string                      `json:"Base"`        // Important User Setting to indicate the customer site (base)
	IPAddress   string                      `json:"IPAddress"`   // IP Address of the PLC
	OpcUaPort   string                      `json:"OpcPort"`     // Port of the OPCUA server on the PLC, usually 4840
	DimMap      map[string]DimScopeArray    `json:"DimMap"`      // DimMap is used to map the dimension of the array to the scope of the plc program loop
	Stats       Stats                       `json:"Stats"`       // How many node id get processed, helps to debug
	Konfig      map[string]any              `json:"Konfig"`      // Konfig is used to configure the plc program loop
	Tmp         []string                    `json:"Tmp"`         // Can I delete this?
	SWSettings  SWSettings                  `json:"SWSettings"`  // SWSettings is needed for the frontend to start
	AlarmDict   map[string]string           `json:"AlarmDict"`   // Prefill from Nats.Topics and Wn to setup an alarm message part
	AlarmList   []Alarm                     `json:"Alarms"`      // Current active alarms
	AlarmReader *ua.ReadRequest             `json:"AlarmReader"` // OpcUa Read Request
	Publish     map[uint32]VariantToMessage `json:"Publish"`     // Maps the opcua message id to setting that are needed to process the message
	GrafanRaum  string                      `json:"Raum"`        // Raumnummern to be processed by Grafana
}

// VariantToMessage helps to convert the opcua.Node.value to a human readable format
type VariantToMessage struct {
	NodeID  string           // NodeID is the opcua node id of the subscribed item
	FlatRef string           // FlatKey is the key to store the value in the flat map
	Topic   string           // Topic is the NATS subject
	Scale   int8             // Scale is used to convert the value to a human readable format
	Type    string           // Type is the type of the value and is used togeher with the Scale to convert the value
	Metric  string           // Will expose a prometheus metric
	Ort     int              // Ort is the array index +1, used to label a metric
	IsArray bool             // Message is part of an array, the index is Ort-1
	Scope   string           // Scope tell us, of which plc program loop the variable belongs to
	Scoper  string           // Scoper is from DimMap, and needed for DATAv2, and Nat Topic
	Group   string           // Group is the plc variable group, e.g. PA, PX, QX
	Short   string           // Short is the short name of the plc variable, e.g. "DruckPCO", "KaelteDetektor", "SollTemp"
	Symbol  string           // Symbol is the symbol name of the plc variable, e.g. "PA_DruckPCO", "PX_An_KaelteDetektor", "PA_Rm_SollTemp"
	Gauge   prometheus.Gauge // optional prometheus gauge metric configuration for this message
}

type DimScopeArray struct {
	Length  uint8
	Scopers []string
}

// Alarm will be parsed from the OPCua Group Alarm
type Alarm struct {
	Key     string `json:"key"` // on frontend key is used look up a markdown file, with a detailed Alarm message
	Message string `json:"message"`
	Ort     string `json:"ort"` //
	Id      string `json:"id"`  // e.g AL_Wn.5, AL_KaelteDetektor.1
}

// OPCUA stuff from the Codesys xml file -------------------------------------------------------------------------------------------------

type OpcUaConfig struct {
	Prefix      string              `json:"Prefix"`
	XmlFileName string              `json:"XmlFileName"`
	Url         string              `json:"Url"`
	Symbole     SymbolConfiguration `json:"SymbolConfiguration"`
}

type Comment struct {
	Group  string
	CDName string
	Metric string
	Descr  string
	Short  string
	Scale  int8
	Konfig string
	Lb0    string
	Lb1    string
	Lb2    string
	Lb3    string
	Lb4    string
	Lb5    string
	Lb6    string
	Lb7    string
	Lb8    string
	Lb9    string
	Lb10   string
}

// Metrics ---------------------------------------------------------------------------------------------------------------------------------
// for prometheus endpoint /metrics
type MetricsConfig struct {
	Defaults []MetricsPreset          `json:"Defaults"`
	Results  map[string]MetricSetting `json:"results"`
}

type MetricsPreset struct {
	Id     string `json:"id"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	Help   string `json:"help"`
	Labels string `json:"labels"`
	Mtype  string `json:"mtype"`
}

type MetricSetting struct {
	Help   string   `json:"help"`
	Labels []string `json:"labels"`
}

func (cfg *Config) ToJson() []byte {
	res, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		log.Println("Error marshalling to JSON:", err)
	}
	return res
}

type Stats struct {
	TotalSubcriptions uint32
	Group             []Group
}
type Group struct {
	Id    string
	Count int
}

// NodeHelper stores informations to convert OPC UA nodes to go types
type NodeHelper struct {
	NodeId  ua.NodeID  // Parsed NodeId to be used for Reading
	Name    string     // SymbolName e.g. Gb_Anzahl_Verf
	Vals    ua.Variant // Vals hold the processed values, mostly a variant of simple types or an array of simple types
	Key     string     // Target key of the resulting structure, where ExtensionObject, are convertet to Arrays
	Index   int        // -1 if its underling node is not an array
	Type    string     // Used in switch case to determin how to handle the node
	Comment string     // Just the text, if the node a a comment tag in the xml
}

type SWSettings struct {
	Raumnummer  []string
	Raumanzahl  int
	Kunde       string
	Version     string
	Reportdatum string
	Header      string `json:"header"`
	Base        string `json:"base"`
	A           int
	C           int
	D           int
	K           int
	M           int
	R           int
	V           int
	J           int
	Profil      string
	Archive     string
	Vnn         []string `json:"vnn"`
}

/*
const report = global.get("Report");
const base = global.get("base");
const dim = global.get("dimension");
const profile = global.get("uv4_compiler");
const archive = global.get("uv4_version");

const settings = {

	    "Raumnummer": report?.Raumnummer || ["R?"],
	    "Raumanzahl": report?.Raumanzahl || 1,
	    "Kunde": report?.Kunde || "NOBODY",
	    "Version": report?.Version || "NOVERS",
	    "Reportdatum": report?.Reportdatum ||  "01.01.1000",
	    "header": report?.header || "Tagesreport 01.01.1000 NOHEAD",
	    base,
	    ...dim,
	    "Profil": profile,
	    "Archive": archive
	}
*/
