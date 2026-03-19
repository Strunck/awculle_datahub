States 

	INIT_STATE        // Initialize global variables
	SETUP_ENV         // Setup environment variables
	SETUP_FILES       // Setup configuration files, settings.json and parse Symbolconfiguration
    INIT_TELEGRAM     // Init Telegram Bot, but do not fail.
	INIT_TOPICS       // Initialize the OPC UA client
	DIAL_OPCUA_SRV    // Dial OPCUA Server (PLC)
	CONNECTED         // Connected to the OPC UA server
	PLC_PREFIX        // Read PLC prefix from the OPC UA server
	PLC_CONFIG_READ   // Read PLC configuration
    PLC_CONFIG_ALARMS // Setup alarm messages
	PLC_CONFIG_DONE   // PLC configuration is read
	SETUP_DONE        // Setup from files and env is done
	REQ_MONITORING    // Request monitoring of nodes
	MONITORING        // Monitoring nodes
	DISCONNECTED      // Disconnected from the OPC UA server
	RETRY             // Retry connection to the OPC UA server
    RETRY_WAIT        // Waits, for connection with exponential backup
	SHUTDOWN          // Shutdown the OPC UA client



```mermaid
stateDiagram-v2

    state CONNECT_OK <<choice>>
    
    [*] --> INIT_STATE    

    INIT_STATE --> SETUP_ENV

    SETUP_ENV --> SETUP_FILES
    
    SETUP_FILES --> INIT_TELEGRAM

    INIT_TELEGRAM --> DIAL_OPCUA_SRV

    DIAL_OPCUA_SRV --> CONNECT_OK

    CONNECT_OK --> CONNECTED

    CONNECTED --> INIT_TOPICS
    
    INIT_TOPICS -->  PLC_CONFIG_READ

    PLC_CONFIG_READ --> PLC_CONFIG_ALARMS
    
    PLC_CONFIG_ALARMS --> PLC_CONFIG_DONE

    PLC_CONFIG_READ --> RETRY : timeout

    PLC_CONFIG_DONE --> REQ_MONITORING         

    REQ_MONITORING --> MONITORING

    REQ_MONITORING --> RETRY: general timeout

    MONITORING --> RETRY: opcua timeout

    RETRY --> DIAL_OPCUA_SRV
    
    RETRY_WAIT --> CRASH

    CONNECT_OK --> RETRY_WAIT : network err
    
    RETRY_WAIT --> RETRY
    
    CRASH --> [*]: Crash Exit 
    
    
 ```
