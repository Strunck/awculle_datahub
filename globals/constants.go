package globals

import "time"

const DEBUG = true

const (
	INIT_STATE        LoopState = iota // Initialize global variables
	SETUP_ENV                          // Setup environment variables
	SETUP_FILES                        // Setup configuration files, setup.json
	SETUP_DONE                         // Setup from files and env is done
	INIT_TELEGRAM                      // Initialize the Telegram bot, but do not fail
	INIT_TOPICS                        // Initialize the OPC UA client topics
	DIAL_OPCUA_SRV                     // Dials OPC UA server (PLC)
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
	INTERVAL                = 200 * time.Millisecond
	LAG                     = 1 * time.Millisecond
	TIMEOUT_OPCUA           = 4 // seconds to increment for each retry attempt
	TIMEOUT_PING            = 2 * time.Second
	READMAXWAIT             = 5 * time.Second
	RETRY_WAIT_SECONDS_MAX  = 12
	CRASH_EXIT_WAIT_SECONDS = 60 // 3 h, should be enough time to fix the issue and restart the container, otherwise it will crash exit and restart automatically via podman restart policy
)

const (
	MAXREAD = 100
)

const (
	SYMBOLXML    = "ULOVisor_V0.14.20_AAA.04.Device.Application.xml"
	SETTINGSJSON = "settings.json"
	HttpPort     = 1881
)
