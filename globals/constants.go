package globals

import "time"

const DEBUG = true

const (
	INIT_STATE        LoopState = iota // Initialize global variables
	SETUP_FILES                        // Setup configuration files, setup.json
	SETUP_ENV                          // Setup environment variables
	SETUP_DONE                         // Setup from files and env is done
	INIT_CLIENT                        // Initialize the OPC UA client
	REQ_CONNECTION                     // Request connection to the OPC UA server
	CONNECTED                          // Connected to the OPC UA server
	PLC_PREFIX                         // Read PLC prefix from the OPC UA server
	PLC_CONFIG_READ                    // Read PLC configuration
	PLC_CONFIG_ALARMS                  // Read PLC alarm configuration
	PLC_CONFIG_DONE                    // PLC configuration is read
	REQ_MONITORING                     // Request monitoring of nodes
	MONITORING                         // Monitoring nodes
	DISCONNECTED                       // Disconnected from the OPC UA server
	RETRY                              // Retry connection to the OPC UA server
	RETRY_WAIT                         // Wait before retrying connection
	SHUTDOWN                           // Shutdown the OPC UA client
)

const (
	INTERVAL      = 200 * time.Millisecond
	LAG           = 1 * time.Millisecond
	TIMEOUT_OPCUA = 4 * time.Second
	TIMEOUT_PING  = 2 * time.Second
	READMAXWAIT   = 5 * time.Second
)

const (
	MAXREAD = 100
)

const (
	SYMBOLXML    = "ULOVisor_V0.14.20_AAA.04.Device.Application.xml"
	SETTINGSJSON = "settings.json"
	HttpPort     = 1881
)
