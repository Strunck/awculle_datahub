States 

	INIT_STATE        / Initialize global variables
	SETUP_FILES       // Setup configuration files, setup.json
	SETUP_ENV         // Setup environment variables
	SETUP_DONE        // Setup from files and env is done
	INIT_CLIENT       // Initialize the OPC UA client
	REQ_CONNECTION    // Request connection to the OPC UA server
	CONNECTED         // Connected to the OPC UA server
	PLC_PREFIX        // Read PLC prefix from the OPC UA server
	PLC_CONFIG_READ   // Read PLC configuration
    PLC_CONFIG_ALARMS 
	PLC_CONFIG_DONE   // PLC configuration is read
	REQ_MONITORING    // Request monitoring of nodes
	MONITORING        // Monitoring nodes
	DISCONNECTED      // Disconnected from the OPC UA server
	RETRY             // Retry connection to the OPC UA server
    RETRY_WAIT         
	SHUTDOWN          // Shutdown the OPC UA client



```mermaid
stateDiagram-v2
    state ERROR <<join>>
    state FATAL <<join>>
    state CONNECT_OK <<choice>>
    
    [*] --> INIT_STATE    

    INIT_STATE --> SETUP_ENV

    SETUP_ENV --> SETUP_FILES

    SETUP_ENV --> FATAL

    FATAL --> REQ_CONNECTION

    SETUP_FILES --> FATAL

    FATAL --> CRASH
    
    CRASH --> [*]: Exit
    
    REQ_CONNECTION --> CONNECT_OK

    CONNECT_OK --> CONNECTED

    CONNECTED --> INIT_CLIENT
    
    INIT_CLIENT -->  PLC_CONFIG_READ

    PLC_CONFIG_READ --> PLC_CONFIG_DONE

    PLC_CONFIG_READ --> ERROR : timeout

    PLC_CONFIG_DONE --> SETUP_DONE 
    
    SETUP_DONE --> REQ_MONITORING

    REQ_MONITORING --> MONITORING

    REQ_MONITORING --> ERROR: general timeout

    MONITORING --> ERROR: opcua timeout

    ERROR --> RETRY
    
    RETRY --> FATAL

    CONNECT_OK --> DISCONNECTED 

    DISCONNECTED --> RETRY_WAIT
    
    RETRY_WAIT --> RETRY

    
 ```
