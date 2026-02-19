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
