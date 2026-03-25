package globals

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/awcullen/opcua/client"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
)

type LoopState uint8

type GlobalState struct {
	OpcClient      *client.Client
	Loop           LoopState
	SessionName    string
	Cfg            Config
	NatsClient     *nats.Conn
	NatsServer     *server.Server
	Plc            PlcState
	DataV2         DataV2
	Stats          Stats `json:"Stats"`
	Report         ReportData
	Telegram       Telegram
	Alarms         Alarms
	MetricRegistry *prometheus.Registry
}

type ReportData struct {
	mu    sync.Mutex
	Daily map[string]DailyReportExtension
	All   [][]ReportCsvLine
}

type Stats struct {
	InitDone    bool      `json:"init_done"`
	Reportdatum string    `json:"Reportdatum"`
	Uv4Version  string    `json:"uv4_version"`
	Version     string    `json:"middleware_version"`
	Developer   bool      `json:"developer_mode"`
	NatsHealth  MsgHealth `json:"nats_health"`
	OpcHealth   MsgHealth `json:"opc_health"`
	Attempts    int       `json:"attempts"`
}

func (stats *Stats) ToJson() []byte {
	stats.NatsHealth.calcAge()
	stats.OpcHealth.calcAge()
	res, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		log.Println("Error marshalling Stats to JSON:", err)
	}
	return res
}

type PlcState struct {
	Flat map[string]HMIVariableTyped
}

func (ps *PlcState) ToJson() []byte {
	res, err := json.MarshalIndent(ps.Flat, "", "  ")
	if err != nil {
		log.Println("Error marshalling PlcState to JSON:", err)
	}
	return res
}

func (gs *GlobalState) SetState(state LoopState) {
	gs.Loop = state
	if DEBUG {
		gs.PrintState()
	}
}

func (gs *GlobalState) PrintState() {
	switch gs.Loop {
	case INIT_STATE:
		log.Println("🥅 INIT_STATE 🥅")
	case SETUP_ENV:
		log.Println("🪬 SETUP_ENV 🪬")
	case SETUP_FILES:
		log.Println("🪑 SETUP_FILES 🪑")
	case INIT_TELEGRAM:
		log.Println("🤖 INIT_TELEGRAM 🤖")
	case SETUP_DONE:
		log.Println("🥅 SETUP_DONE 🥅")

	case DIAL_OPCUA_SRV:
		log.Println("👋 DIAL_OPCUA_SRV 👋")
	case CONNECTED:
		log.Println("✅ CONNECTED ✅")
	case DISCONNECTED:
		log.Println("💢 DISCONNECTED 💢")

	case INIT_TOPICS:
		log.Println("🥅 INIT_TOPICS 🥅")
	case PLC_CONFIG_READ:
		log.Println("🪑 PLC_CONFIG_READ 🪑")
	case PLC_CONFIG_ALARMS:
		log.Println("🪑 PLC_CONFIG_ALARMS 🪑")
	case PLC_CONFIG_DONE:
		log.Println("🪑 PLC_CONFIG_DONE 🪑")

	case REQ_MONITORING:
		log.Println("👁️‍🗨️ REQ_MONITORING 👁️‍🗨️")

	case MONITORING:
		log.Println("👀 MONITORING 👀")

	case RUN:
		log.Println("🏃 RUN 🏃")

	case SHUTDOWN:
		log.Println("💀 SHUTDOWN 💀")

	case RETRY:
		log.Println("🔄 RETRY 🔄")

	case RETRY_WAIT:
		log.Println("⏳ RETRY_WAIT ⏳")

	default:
		log.Println("❓ UNKNOWN ❓")
	}
}

func (gs *GlobalState) GetDataV2Instance() *DataV2 {
	return &gs.DataV2
}
