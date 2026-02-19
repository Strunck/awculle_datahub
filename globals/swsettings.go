package globals

import (
	"fmt"
	"strings"
)

func (cfg *Config) InitSWSettings() (err error) {

	var s *SWSettings = &SWSettings{}

	var f any
	var ok bool

	// Make Array from string
	f, ok = cfg.Plc.Konfig["gs_Raumnummer"]
	if !ok {
		return fmt.Errorf("key 'gs_Raumnummer' not found in Plc.Konfig")
	}
	rn := strings.Split(f.(string), ",")
	s.Raumnummer = rn

	// Raumstring for Grafana	// Raumanzahl
	el, ok := cfg.Plc.DimMap["R"]
	if !ok {
		return fmt.Errorf("key 'R' not found in Plc.DimMap")
	}
	s.Raumanzahl = int(el.Length)
	s.R = s.Raumanzahl

	// Error checks
	if s.Raumanzahl > len(rn) {
		return fmt.Errorf("string Raumnummer has %d entries, but Raumanzahl is %d", len(rn), s.Raumanzahl)
	}
	if s.Raumanzahl == 0 {
		return fmt.Errorf("initSWSettings Raumanzahl is 0, please check your configuration")
	}

	// Raumstring for Grafana
	var rs []string
	for i := 0; i < s.Raumanzahl; i++ {
		rs = append(rs, fmt.Sprintf("%s : %d", rn[i], i))
	}
	cfg.Plc.GrafanRaum = strings.Join(rs, ", ")

	f, ok = cfg.Plc.Konfig["Gs_Kunde"]
	if !ok {
		return fmt.Errorf("key 'Gs_Kunde' not found in Plc.Konfig")
	}
	s.Kunde = f.(string)

	s.Version = cfg.Plc.Konfig["Gs_Version"].(string)
	f, ok = cfg.Plc.Konfig["Gs_Version"]
	if !ok {
		return fmt.Errorf("key 'Gs_Version' not found in Plc.Konfig")
	}
	s.Version = f.(string)

	s.Base = cfg.Plc.Base

	s.A = int(cfg.Plc.DimMap["A"].Length)
	s.C = int(cfg.Plc.DimMap["C"].Length)
	s.D = int(cfg.Plc.DimMap["D"].Length)
	s.K = int(cfg.Plc.DimMap["K"].Length)
	s.M = int(cfg.Plc.DimMap["M"].Length)

	s.V = int(cfg.Plc.DimMap["V"].Length)

	f, ok = cfg.Plc.Konfig["Gb_AnzahlEMeter"]
	if !ok {
		return fmt.Errorf("key 'Gb_AnzahlEMeter' not found in Plc.Konfig")
	}
	s.J = f.(int)

	s.Profil = cfg.Opcua.Symbole.Header.Version

	s.Archive = cfg.Opcua.XmlFileName

	cfg.Plc.SWSettings = *s

	return nil
}
