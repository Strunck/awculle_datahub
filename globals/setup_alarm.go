package globals

import (
	"context"
	"fmt"
	"strings"

	"time"

	"github.com/awcullen/opcua/client"
	"github.com/awcullen/opcua/ua"
)

// Filter alarm messages from .Nats.Topics
// Create a map of shorts -> description to assotiate the descr
// to the Alarm, if one should occur
func (cfg *Config) SetupAlarmMessages(plcCon *client.Client, readmaxwait time.Duration, bgctx context.Context) (err error) {
	normAlarmMap := make(map[string]string)
	var normalAlarmList []ua.ReadValueID
	for _, v := range cfg.Nats.Topics {
		if v.Name == "AL_AlarmAktiv" || v.Name == "AL_AlarmMerk" {
			continue
		}
		if v.Group == "Al" {
			normAlarmMap[v.Opcnode] = v.Descr

			nid := ua.ParseNodeID(v.Opcnode)
			normalAlarmList = append(normalAlarmList, ua.ReadValueID{
				NodeID:      nid,
				AttributeID: ua.AttributeIDValue,
			})
		}
	}

	// Get the descr field from the Wn Node and if they are active
	// array of node id

	firstRequest := make([]ua.ReadValueID, 100)
	wnAlarmReadList := make([]ua.ReadValueID, 20)
	//cfg.Plc.AlarmReader = make([]*ua.ReadRequest, 3)

	var wn_id, vnn_id string

	for i := 1; i <= 20; i++ {
		//
		wn_id = fmt.Sprintf("%sGTB.Wn[%d].sName", cfg.Opcua.Prefix, i)
		wn_nodeid := ua.ParseNodeID(wn_id)

		firstRequest[i-1] = ua.ReadValueID{
			NodeID:      wn_nodeid,
			AttributeID: ua.AttributeIDValue,
		}

		wn_alarm_nid := fmt.Sprintf("%sGTB.Wn[%d].xAlarm", cfg.Opcua.Prefix, i)
		wn_alarm_nodeid := ua.ParseNodeID(wn_alarm_nid)

		wnAlarmReadList[i-1] = ua.ReadValueID{
			NodeID:      wn_alarm_nodeid,
			AttributeID: ua.AttributeIDValue,
		}
	}

	for i := 21; i <= 100; i++ {
		vnn_id = fmt.Sprintf("%sGTB.Vnn[%d].sBezeichner", cfg.Opcua.Prefix, i-20)
		vnn_nodeid := ua.ParseNodeID(vnn_id)
		firstRequest[i-1] = ua.ReadValueID{
			NodeID:      vnn_nodeid,
			AttributeID: ua.AttributeIDValue,
		}

	}

	secondRequest := make([]ua.ReadValueID, 80)
	for i := 1; i <= 80; i++ {
		vnn_id = fmt.Sprintf("%sGTB.Vnn[%d].sBezeichner", cfg.Opcua.Prefix, i)
		vnn_nodeid := ua.ParseNodeID(vnn_id)

		secondRequest[i-1] = ua.ReadValueID{
			NodeID:      vnn_nodeid,
			AttributeID: ua.AttributeIDValue,
		}
	}

	//  Read the Nodes
	ctx, cancel := context.WithTimeout(bgctx, readmaxwait)
	defer cancel()

	firstResponse, err := plcCon.Read(ctx, &ua.ReadRequest{MaxAge: 2000, NodesToRead: firstRequest})
	if err != nil {
		return fmt.Errorf("setting up alarms from firstResponse: %v", err)
	}

	fr := firstResponse.Header()
	if fr.ServiceResult.IsBad() {
		return fmt.Errorf("setting up alarms from firstResponse: header servcice result 'is bad'")
	}

	var wn_key, wn_message string

	for i, resp := range firstResponse.Results {

		if i < 20 {

			wn_key = fmt.Sprintf("%s", firstRequest[i].NodeID)
			// replace sName with xAlarm
			wn_key = strings.Replace(wn_key, "sName", "xAlarm", 1)
			wn_message = resp.Value.(string)
			// log.Printf("%s %s", wn_key, wn_message)
			normAlarmMap[wn_key] = wn_message
		} else {
			vnn_msg := resp.Value.(string)
			cfg.Plc.SWSettings.Vnn = append(cfg.Plc.SWSettings.Vnn, vnn_msg)
			// log.Printf("%s", vnn_msg)
		}
	}

	wnAlarmReadList = append(wnAlarmReadList, normalAlarmList...)

	cfg.Plc.AlarmReader = &ua.ReadRequest{
		MaxAge:      2000,
		NodesToRead: wnAlarmReadList,
	}
	cfg.Plc.AlarmDict = normAlarmMap
	return nil
}
