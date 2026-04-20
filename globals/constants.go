package globals

import "time"

const DEBUG = true

const (
	INIT_STATE         LoopState = iota // Initialize global variables
	SETUP_ENV                           // Setup environment variables
	SETUP_FILES                         // Setup configuration files, setup.json
	SETUP_DONE                          // Setup from files and env is done
	INIT_TELEGRAM                       // Initialize the Telegram bot, but do not fail
	INIT_TOPICS                         // Initialize the OPC UA client topics
	DIAL_OPCUA_SRV                      // Dials OPC UA server (PLC)
	CONNECTED                           // Connected to the OPC UA server
	PLC_PREFIX                          // Read PLC prefix from the OPC UA server
	PLC_CONFIG_READ                     // Read PLC configuration
	PLC_CONFIG_ALARMS                   // Read PLC alarm configuration
	PLC_CONFIG_METRICS                  // Registers Prometheus Metrics
	PLC_CONFIG_DONE                     // PLC configuration is read
	REQ_MONITORING                      // Request monitoring of nodes
	MONITORING                          // Monitoring nodes
	RUN                                 // Run the main loop
	DISCONNECTED                        // Disconnected from the OPC UA server
	RETRY                               // Retry connection to the OPC UA server
	RETRY_WAIT                          // Wait before retrying connection
	SHUTDOWN                            // Shutdown the OPC UA client
)

const (
	READ_ALARM_INTERVAL    = 5000 * time.Millisecond
	LAG                    = 1 * time.Millisecond
	RETRYWAIT_INCREMENT    = 4               // seconds to increment for each retry attempt
	NETDIAL_TIMEOUT        = 2 * time.Second // timeout for net.Dial TCP Port 4840
	READWAIT_INC_MAX       = 5 * time.Second
	RETRY_WAIT_SECONDS_MAX = 12                                // max seconds, between retry attempt
	RETRYWAIT_ATTEMPTS_MAX = 15                                // max 15 attempt are made. Otherwise it will crash exit and restart automatically via podman restart policy.
	OPCUA_SESSION_TIMEOUT  = RETRYWAIT_ATTEMPTS_MAX * 3 * 1000 // WithSessionTimeout sets the number of milliseconds that a session may be unused before being closed by the server. (default: 2 min)
	OPCUA_CONNECT_TIMEOUT  = 1000 * 10                         // sets the number of milliseconds to wait for a connection response. (default:5000)
	SYMBOLXML              = "ULOVisor_V0.14.20_AAA.04.Device.Application.xml"
	SETTINGSJSON           = "settings.json"
	HttpPort               = 1881
)
