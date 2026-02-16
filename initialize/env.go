package initialize

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Environment overwrites settings.json with .env 🪬
// returns errors if essential environment variables are not set
func (cfg *Config) Environment(devMode bool) (err error) {

	// Show all environment variables, similar to `printenv`
	envVars := os.Environ()
	if devMode {
		log.Printf("🪬 Environment variables:")
		for _, envVar := range envVars {
			log.Printf("🪬 - %s", envVar)
		}
	}
	var ok bool
	if cfg.KoDataPath, ok = os.LookupEnv("KO_DATA_PATH"); !ok {
		return fmt.Errorf("🪬 KO_DATA_PATH environment variable is not set")
	}

	// List current directory files for debugging
	files, err := os.ReadDir(cfg.KoDataPath)
	if err != nil {
		return fmt.Errorf("🪬 reading current directory: %w", err)
	}
	log.Printf("🪬 List of files for directory: %s", cfg.KoDataPath)
	for _, file := range files {
		log.Printf("🪬 - %s (%#v)", file.Name(), file.Type())
	}

	// Fallback for local dev mode. Container mode must use option --env-file=.../.env
	if err := godotenv.Load(cfg.KoDataPath + "/.env"); err != nil {
		log.Printf("🪬 loading .env file: %v", err)
	}

	// Access environment variables
	name := os.Getenv("NAME")
	pcName := os.Getenv("PC_NAME")
	uvHost := os.Getenv("UVHOST")
	if len(uvHost) == 0 {
		return fmt.Errorf("🪬 UVHOST environment variable is not set")
	}
	cfg.Plc.Base = uvHost

	spsFtpPass := os.Getenv("SPS_FTP_PASS")
	spsPass := os.Getenv("SPS_PASS")
	spsIpAddress := os.Getenv("SPS_IPADRESS")
	if len(spsIpAddress) == 0 {
		return fmt.Errorf("🪬 SPS_IPADRESS environment variable is not set")
	}
	cfg.Plc.IPAddress = spsIpAddress

	spsOpcUaPort := os.Getenv("SPS_OPC_PORT")
	if len(spsOpcUaPort) == 0 {
		spsOpcUaPort = "4840" // default port
	}
	cfg.Plc.OpcUaPort = spsOpcUaPort

	spsHostName := os.Getenv("SPS_HOSTNAME")
	spsOpcEndpoint := os.Getenv("SPS_OPC_ENDPOINT")
	if len(spsOpcEndpoint) == 0 {
		if len(cfg.Opcua.Url) == 0 {
			return fmt.Errorf("🪬 neither SPS_OPC_ENDPOINT nor Opcua.Url in settings.json is set")
		}
	}
	cfg.Opcua.Url = spsOpcEndpoint

	if devMode {
		// Use the environment variables as needed
		log.Printf("🪬 Loaded environment variables:\n")
		log.Printf("🪬 NAME: %s\n", name)
		log.Printf("🪬 PC_NAME: %s\n", pcName)
		log.Printf("🪬 UVHOST: %s\n", uvHost)
		log.Printf("🪬 SPS_FTP_PASS: %s\n", spsFtpPass)
		log.Printf("🪬 SPS_PASS: %s\n", spsPass)
		log.Printf("🪬 SPS_IPADRESS: %s\n", spsIpAddress)
		log.Printf("🪬 SPS_HOSTNAME: %s\n", spsHostName)
		log.Printf("🪬 SPS_OPC_ENDPOINT: %s\n", spsOpcEndpoint)
	}

	return nil
}
