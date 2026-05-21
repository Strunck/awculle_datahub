package globals

import (
	"fmt"
	"strings"

	"github.com/awcullen/opcua/ua"
)

func (v *VariantToMessage) ToString(msg ua.MonitoredItemNotification) (s string, err error) {
	// if v.IsArray {
	// 	return "", fmt.Errorf("ToString: arrays not supported")
	// }

	switch v.Type {
	case "BOOL":
		if msg.Value.Value.(bool) {
			s = "true"
		} else {
			s = "false"
		}
	case "BYTE":
		s = scaleInt(fmt.Sprintf("%d", msg.Value.Value.(uint8)), v.Scale)
	case "INT":
		s = scaleInt(fmt.Sprintf("%d", msg.Value.Value.(int16)), v.Scale)
	case "EMT":
		s = EnumEMTtoStr(msg.Value)
	case "WORD":
		s = scaleInt(fmt.Sprintf("%d", msg.Value.Value.(uint16)), v.Scale)
	case "DWORD":
		s = scaleInt(fmt.Sprintf("%d", msg.Value.Value.(uint32)), v.Scale)
	case "REAL":
		s = fmt.Sprintf("%f", msg.Value.Value.(float32))
	case "LREAL":
		s = fmt.Sprintf("%f", msg.Value.Value.(float64))
	case "TIME":
		s = fmt.Sprintf("%v", msg.Value.Value)
	case "PVMOD":
		s = EnumPVMODtoStr(msg.Value)

	default:
		return "", fmt.Errorf("ToString for node %s has unknown type %s", v.NodeID, v.Type)
	}

	return s, nil
}

func scaleInt(val string, scale int8) (scaledVal string) {

	if scale == 0 {
		return val
	}

	minus := ""
	betrag, hasMinus := strings.CutPrefix(val, "-")
	if hasMinus {
		minus = "-"
	}

	if len(betrag) <= int(scale) {
		x := int(scale) - len(betrag)
		scaledVal = minus + "0." + strings.Repeat("0", x) + betrag
	} else {
		x := len(betrag) - int(scale)
		scaledVal = minus + betrag[0:x] + "." + betrag[x:]
	}

	return scaledVal
}

func EnumEMTtoStr(val ua.DataValue) (s string) {
	switch val.Value.(int32) {
	case 0:
		s = "Sonderfall"
	case 1:
		s = "Hausanschluss"
	case 2:
		s = "Schaltschrank"
	case 3:
		s = "KaelteErzeuger"
	case 4:
		s = "KaelteVerteilung"
	case 5:
		s = "ULO_Technik"
	case 6:
		s = "Scrubber"
	case 7:
		s = "Stickstofferz"
	case 8:
		s = "Solepumpe"
	case 9:
		s = "CEE_Dose"
	case 10:
		s = "Verdampferlueft"
	case 11:
		s = "KaelteAnlage"
	case 255:
		s = "Undefiniert"
	default:
		s = fmt.Sprintf("Unknown EMT value: %d", val.Value.(int32))
	}
	return s
}

func EnumPVMODtoStr(val ua.DataValue) (s string) {
	switch val.Value.(int32) {
	case 0:
		s = "AUS"
	case 1:
		s = "AUTO"
	case 2:
		s = "HAND"
	case 3:
		s = "PV"
	default:
		s = fmt.Sprintf("Unknown PVMOD value: %d", val.Value.(int32))
	}
	return s
}
