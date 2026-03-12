package globals

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
)

// 🪑 reads settings.json and Codesys xml with symbolconfiguration
// pases the files.
// If this fails, the files are corrupred on we crash with a fatal error.
func (gs *GlobalState) SetupFiles() (err error) {
	gs.SetState(SETUP_FILES)

	settingsFileName := gs.Cfg.KoDataPath + "/" + gs.Cfg.SettingsFile
	symbolConfigFileName := gs.Cfg.KoDataPath + "/" + gs.Cfg.Opcua.XmlFileName

	//settings.json
	fs, err := os.ReadFile(settingsFileName)

	if err != nil {
		return fmt.Errorf("🪑 failed to read file %s: %w", settingsFileName, err)
	}

	if err := json.Unmarshal(fs, &gs.Cfg); err != nil {
		return fmt.Errorf("🪑 failed to unmarshal file %s: %w", settingsFileName, err)
	}

	// Load Codesys Symbolkonfiguration, see sysmbolconfiguration.go
	if err := gs.Cfg.Opcua.Symbole.parse(symbolConfigFileName); err != nil {
		return fmt.Errorf("🪑 Konnte Symbolkonfiguration nicht parsen %w", err)
	}

	return nil
}

func (gs *GlobalState) SetupTopics() (err error) {
	gs.SetState(INIT_TOPICS)
	// Parse XML-Nodes, that are relevant for the HMI
	for _, grp_node := range gs.Cfg.Opcua.Symbole.NodeList.Node.Node {
		//log.Printf("🪑 DEBUG Gruppe: %s\t%s\n", grp_node.Name, grp_node.Comment)

		// Statistik über die Anzahl der Knoten in der Gruppe
		grp := Group{
			Id:    grp_node.Name,
			Count: 0,
		}

		for _, leaf_node := range grp_node.Node {

			//TODO Da für BA_Rmm09_Luftfeu_arw keine Farbe hinterlegt ist, stürzt das Frontend ab. Siehe Frontend/lib/utils.ts
			if leaf_node.Name == "BA_Rmm09_Luftfeu_arw" {
				continue
			}

			comment := parseComment(leaf_node.Comment)
			comment.CDName = leaf_node.Name
			comment.Group = grp_node.Name

			// if c.Group is in config.Nats.GroupIds.
			// if there is comment that does not start with "#short=" -> ignore this node
			if contains(gs.Cfg.Nats.GroupIds, comment.Group) && len(comment.Short) > 0 {

				so := fmt.Sprintf("%s%s.%s", gs.Cfg.Opcua.Prefix, comment.Group, comment.CDName)
				natstopic_lastpart := fmt.Sprintf("%s.%s", comment.Group, comment.Short)

				top := OPCtopic{
					Natstopic: natstopic_lastpart, // BD.Temp
					Short:     comment.Short,
					Group:     comment.Group,
					Opcnode:   so,
					Scale:     comment.Scale,
					Scope:     getScope(comment.CDName),
					Length:    0,
					Type:      leaf_node.Type,
					Name:      leaf_node.Name,
					Symbol:    leaf_node.Name,
					Vartype:   getVartype(leaf_node.Type),
					IsArray:   strings.HasPrefix(leaf_node.Type, "T_ARRAY_"),
					Konfig:    comment.Konfig,
					Metric:    comment.Metric,
					Descr:     comment.Descr,
				}

				gs.Cfg.Nats.Topics = append(gs.Cfg.Nats.Topics, top)
				grp.Count++ // Statistik for Debug, How many Nodes in this group
			}
		}
		gs.Cfg.Plc.Stats.Group = append(gs.Cfg.Plc.Stats.Group, grp)
	}

	// Make a map of metrics name to setting
	if err := gs.Cfg.makeMetricMap(); err != nil {
		return fmt.Errorf("🪑 error creating metric map: %w", err)
	}

	return nil

}

func getVartype(s string) (r string) {
	s1 := strings.Split(s, "_")
	if len(s1) == 0 {
		return "🪑 scheisse"
	}
	return s1[len(s1)-1]
}

var (
	specialScopes = map[string]string{
		"Max_CO2_24h_arw":         "R",
		"Min_CO2_24h_arw":         "R",
		"Max_O2_24h_arw":          "R",
		"Min_O2_24h_arw":          "R",
		"Max_KontTemp_24h_ari":    "R",
		"Min_KontTemp_24h_ari":    "R",
		"Max_Temp_24h_ari":        "R",
		"Min_Temp_24h_ari":        "R",
		"gari_NrRaum":             "R",
		"gari_RaumVisible":        "R",
		"garw_Jahr":               "R",
		"garw_Monat":              "R",
		"garw_Tag":                "R",
		"gi_RaumTemp_24h_AVG":     "R",
		"gi_KontTemp_24h_AVG":     "R",
		"gw_E_Menge":              "R",
		"gw_Sollentfeuchtung":     "R",
		"Gar_KP_Verf":             "A",
		"Gar_Offset_Verd":         "A",
		"Gar_Offset_Verf":         "A",
		"Gar_TN_Verd":             "A",
		"Gar_TN_Verf":             "A",
		"Gar_TV_Verd":             "A",
		"Gar_TV_Verf":             "A",
		"Gar_AggVisible":          "A",
		"Gar_Bar_HDOben1":         "A",
		"Gar_Bar_HDOben2":         "A",
		"Gar_Bar_HDUnten1":        "A",
		"Gar_Bar_HDUnten2":        "A",
		"gari_Bar_NDOben1":        "A",
		"gari_Bar_NDOben2":        "A",
		"gari_Bar_NDUnten1":       "A",
		"gari_Bar_NDUnten2":       "A",
		"gari_Bar_VerdampfMax":    "A",
		"gari_Bar_VerdampfMin":    "A",
		"gari_KondTempSoll_arw":   "A",
		"gari_NrAgg":              "A",
		"gari_SollTVerfMax":       "A",
		"gari_SollTVerfMin_arw":   "A",
		"gari_Temp_HDOben1":       "A",
		"gari_Temp_HDOben2":       "A",
		"gari_Temp_HDUnten1":      "A",
		"gari_Temp_HDUnten2":      "A",
		"gari_Temp_NDOben1":       "A",
		"gari_Temp_NDOben2":       "A",
		"gari_Temp_NDUnten1":      "A",
		"gari_Temp_NDUnten2":      "A",
		"gari_Temp_PuDoSchwelle":  "A",
		"gari_Temp_RegelSaug":     "A",
		"gari_Temp_SammlVordruck": "A",
		"gari_Temp_SollEffHD":     "A",
		"gari_Temp_SollEffND":     "A",
		"gari_Temp_WarnUber":      "A",
		"Garx_NDUberbruck":        "A",
		"garx_pdAktiv":            "A",
		"gx_Auto_CO2_1":           "M",
		"gx_Auto_CO2_2":           "M",
		"gx_Auto_CO2_3":           "M",
		"gx_Auto_O2_1":            "M",
		"gx_Auto_O2_2":            "M",
		"gx_Auto_O2_3":            "M",
	}
)

func getScope(n string) (r string) {

	// check if n is in specialScopes
	if r, ok := specialScopes[n]; ok {
		return r
	}

	if len(n) < 4 {
		return "S"
	}
	// extract the forth chatacter from n
	r = string(n[3])

	if string(n[4]) == "m" || string(n[4]) == "n" {
		return r
	}
	return "S"
}

func parseComment(comment string) Comment {
	parts := strings.Split(comment, "#")
	var c Comment

	for _, part := range parts {

		p := strings.Split(strings.TrimSpace(part), "=")
		if len(p) > 1 {
			p1 := strings.TrimSpace(p[1])

			switch p[0] {
			case "metric":
				c.Metric = p1
			case "descr":
				c.Descr = strings.Trim(p1, "\"")
			case "short":
				c.Short = strings.Trim(p1, "\"")
			case "konfig":
				c.Konfig = strings.Trim(p1, "\"")
			case "lb0":
				c.Lb0 = strings.Trim(p1, "\"")
			case "lb1":
				c.Lb1 = strings.Trim(p1, "\"")
			case "lb2":
				c.Lb2 = strings.Trim(p1, "\"")
			case "lb3":
				c.Lb3 = strings.Trim(p1, "\"")
			case "lb4":
				c.Lb4 = strings.Trim(p1, "\"")
			case "lb5":
				c.Lb5 = strings.Trim(p1, "\"")
			case "lb6":
				c.Lb6 = strings.Trim(p1, "\"")
			case "scale":
				n := p1
				scale, err := strconv.ParseInt(n, 10, 8)
				if err != nil {
					log.Println("Error parsing scale:", err)
				}
				c.Scale = int8(scale)
			}
		}
	}
	return c
}

// Hilfsfunktion, um zu prüfen, ob ein Wert in einer Liste enthalten ist
func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}

// MakeMetricMap read .Metrics.Default and alters cfg to include a map of metrics name to setting
// needed to have a help string for the Prometheus metric
func (cfg *Config) makeMetricMap() (err error) {
	var r = make(map[string]MetricSetting)

	if len(cfg.Metrics.Defaults) == 0 {
		return fmt.Errorf("no default metrics found in settings.json")
	}

	// read .Metrics.Default
	for _, m := range cfg.Metrics.Defaults {
		// fmt.Println("DEBUG Metric Init:", m.Name)
		// append a key value pair to the map Metrics.result
		r[m.Name] = MetricSetting{
			Help:   m.Help,
			Labels: labelsArray(m.Labels),
		}
	}
	cfg.Metrics.Results = r

	return nil
}

func labelsArray(labels string) []string {
	splitted := strings.Split(labels, ",")
	for i := range len(splitted) {
		splitted[i] = strings.TrimSpace(splitted[i])
	}
	return splitted
}
