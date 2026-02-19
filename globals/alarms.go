package globals

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

// Alarms holds a map of the current active alarms and the alarm history
type Alarms struct {
	AlCount int                    // Current number of active alarms
	Actual  map[string]ActiveAlarm // map of active alarms, key is alarm ID (f.e. "xAlarm.5", "AL_Rmm07_VerdaEis_ar.1")
	History []ActiveAlarm          // Last 100 cleared alarms. TODO add serialization
}

type ActiveAlarm struct {
	Start time.Time
	End   time.Time
	Alarm Alarm
	Text  string
	Send  bool
}

const historyLimit = 100 // Limit for alarm history entries

func (hal *Alarms) pushHistory(act ActiveAlarm) {
	if hal.History == nil {
		hal.History = make([]ActiveAlarm, 0, historyLimit)
	}

	if len(hal.History) > historyLimit {
		hal.History = hal.History[1:]
	}
	hal.History = append(hal.History, act)
}

func (al *Alarms) addOrPersistAlarm(alarmList []Alarm, gs *GlobalState) {
	for _, act := range alarmList {
		if al.Actual == nil {
			al.Actual = make(map[string]ActiveAlarm)
		}
		if _, exists := al.Actual[act.Id]; !exists {

			activeAlarm := ActiveAlarm{
				Start: time.Now(),
				Alarm: act,
				Send:  true,
				Text:  gs.GetAlarmTextFromSymbol(act),
			}

			// New active alarm
			al.Actual[act.Id] = activeAlarm

			gs.SendTelegramMessage(activeAlarm.Text)

			log.Printf("⚠️ New active alarm added: id: %s Symbol: %s - %s (Ort: %s)", act.Id, act.Key, act.Message, act.Ort)
		}
	}

	for id, a := range al.Actual {
		found := false
		for _, alarm := range alarmList {
			if alarm.Id == id {
				found = true
				break
			}
		}
		if !found {
			a.End = time.Now()
			al.pushHistory(a)

			log.Printf("✅ Alarm cleared: %s - %s (Ort: %s)", a.Alarm.Key, a.Alarm.Message, a.Alarm.Ort)
			delete(al.Actual, id)
		}
	}
}

// Extracs the 4. char from the symbol
func (gs *GlobalState) GetAlarmTextFromSymbol(al Alarm) string {
	// get 4. char from Key
	if len(al.Key) < 4 {
		log.Printf("⚠️ GetAlarmTextFromSymbol: Key too short: %s", al.Key)
		return al.Id
	}
	scope := string(al.Key[3])

	orte, ok := gs.Cfg.Plc.DimMap[scope]
	if !ok {
		return al.Message
	}
	idx, err := strconv.Atoi(al.Ort)
	if err != nil {
		log.Printf("⚠️ GetAlarmTextFromSymbol: Error converting Ort to int: %s", al.Ort)
		return al.Message
	}
	if idx < 1 || idx > len(orte.Scopers) {
		log.Printf("⚠️ GetAlarmTextFromSymbol: Ort index out of range: %d", idx)
		return al.Message
	}
	ortText := orte.Scopers[idx-1]

	return al.Message + " " + ortText

}

// ReadAlarms reads alarm values from OPC UA and populates cfg.Plc.AlarmList
// It get populated if an alarm is active.
// This list get rewritten on user request or periodically every 5 seconds
// so no need to delete a key, since everytime it gets regenerated
func (gs *GlobalState) ReadAlarms(bgctx context.Context) (err error) {

	// Prepare the Reader
	// check if we are below 100 items
	if len(gs.Cfg.Plc.AlarmDict) > 100 {
		err = fmt.Errorf("⚠️ config Plc.AlarmDict has too many alarms to read, @me split the reader and change the logic")
		return err
	}

	// Check if AlarmReader is properly configured
	if gs.Cfg.Plc.AlarmReader == nil {
		err = fmt.Errorf("⚠️ config Plc.AlarmReader is nil, so something bad happened")
		return err
	}

	// Read alarm values using the pre-configured AlarmReader
	response, err := gs.OpcClient.Read(bgctx, gs.Cfg.Plc.AlarmReader)
	if err != nil {
		return fmt.Errorf("⚠️ failed to read alarm values: %w", err)
	}

	// Process the response and populate AlarmList
	activeAlSet := make(map[string]Alarm)
	var oneAlarm Alarm

	for i, resp := range response.Results {
		if resp.StatusCode.IsBad() {
			log.Printf("⚠️ Error reading alarm node %s: %s", gs.Cfg.Plc.AlarmReader.NodesToRead[i].NodeID, resp.StatusCode.Error())
			continue
		}

		// Get the node ID as string for dictionary lookup
		nodeID := fmt.Sprintf("%s", gs.Cfg.Plc.AlarmReader.NodesToRead[i].NodeID)
		// we get a Array of booleans or a single boolean
		if bvals, ok := resp.Value.([]bool); ok {
			// Array of booleans
			for alarmIndex, alarmIsAktive := range bvals {
				if alarmIsAktive {
					oneAlarm, err = gs.setAlarm(nodeID, len(bvals), alarmIndex)

					if err != nil {
						return fmt.Errorf("⚠️ Error setting alarm for node %s: %w", nodeID, err)
					}
				}
			}
		} else if alarmIsAktive, ok := resp.Value.(bool); ok {
			// Single boolean
			if alarmIsAktive {
				oneAlarm, err = gs.setAlarm(nodeID, 0, 0)
				if err != nil {
					return fmt.Errorf("⚠️ Error setting alarm for node %s: %w", nodeID, err)
				}
			}
		}
		activeAlSet[oneAlarm.Id] = oneAlarm
	}

	// Update the alarm list in the configuration
	//gs.Cfg.Plc.AlarmList = make([]Alarm, 0, len(activeAlSet))
	for _, alarm := range activeAlSet {
		if len(alarm.Key) > 0 {
			gs.Cfg.Plc.AlarmList = append(gs.Cfg.Plc.AlarmList, alarm)
			log.Printf("func ReadAlarms append Alarm: %v", alarm)
		}
	}

	gs.Alarms.addOrPersistAlarm(gs.Cfg.Plc.AlarmList, gs)

	// log.Printf("DEBUG ReadAlarms Processed %d alarm nodes, found %d active alarms", len(response.Results), len(activeAlSet))

	return nil
}

// setAlarm creates an alarm object from an OPC UA node ID and configuration parameters.
// It extracts the alarm message from the alarm dictionary and constructs a properly
// formatted alarm with key, message, location, and unique identifier.
//
// Parameters:
//   - cfg: Configuration containing the alarm dictionary for message lookup
//   - nodeID: OPC UA node identifier (e.g., "ns=4;s=|var|EXOR ARM32-Linux QT5.Application.GTB.Wn[5].xAlarm")
//   - vlen: Length of alarm array. If > 0, indicates this is an array-based alarm with location info
//   - idx: Zero-based index in the alarm array, converted to 1-based location number
//
// Returns:
//   - alarm: Populated Alarm struct with Key, Message, Ort (location), and Id fields
//   - err: Error if the nodeID is not found in the alarm dictionary
//
// Examples:
//   - For "GTB.Wn[5].xAlarm" with vlen=20, idx=4: creates alarm with Ort="5", Id="xAlarm.5"
//   - For "Al.AL_KaelteDetektor" with vlen=0: creates alarm with Ort="0", Id="AL_KaelteDetektor"
func (gs *GlobalState) setAlarm(nodeID string, vlen int, idx int) (alarm Alarm, err error) {

	// example nodeid:
	// "ns=4;s=|var|EXOR ARM32-Linux QT5.Application.GTB.Wn[5].xAlarm" -> Plc.Dict[nodeID]
	// "ns=4;s=|var|EXOR ARM32-Linux QT5.Application.Al.AL_KaelteDetektor"
	// -> Plc.Dict[nodeID] (get message)

	var ok bool
	alarm.Message, ok = gs.Cfg.Plc.AlarmDict[nodeID]
	if !ok {
		err = fmt.Errorf("⚠️ could not find a message for %s", nodeID)
		return Alarm{}, err
	}

	s := strings.Split(nodeID, ".")
	if len(s) > 0 {
		alarm.Key = s[len(s)-1]
	}

	if vlen > 0 {
		// Create the alarm object
		alarm.Ort = fmt.Sprintf("%d", idx+1)
		alarm.Id = fmt.Sprintf("%s.%s", alarm.Key, alarm.Ort)

	} else {
		alarm.Ort = "0"
		alarm.Id = alarm.Key
	}

	//log.Printf("DEBUG setAlarm. Active alarm added: %#v", alarm)
	return alarm, nil

}

// StartPeriodicAlarmReader starts a goroutine that reads alarms every 5 seconds
func (gs *GlobalState) StartPeriodicAlarmReader(bgctx context.Context) {

	ticker := time.NewTicker(5 * time.Second)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:

				if err := gs.ReadAlarms(bgctx); err != nil {
					log.Printf("⚠️ Error reading alarms: %v", err)
					if gs.OpcClient != nil {
						// gs.OpcClient.Close(bgctx)
						log.Println("⚠️ Aborted session for Alarm Reader")
						// return
					} else {
						log.Fatalf("⚠️ No OPC UA client to close session. Exiting...")
						return
					}
				}

				return
			case <-bgctx.Done():
				log.Println("⚠️ Stopping periodic alarm reader")
				return
			}
		}
	}()

	log.Println("⚠️ Started periodic alarm reader (every 5 seconds)")
}

// ReadAlarmsOnDemand provides a way to manually trigger an alarm read
// This can be called from web handlers or other parts of the application
func (gs *GlobalState) ReadAlarmsOnDemand(bgctx context.Context) error {
	ctx, cancel := context.WithTimeout(bgctx, READMAXWAIT)
	defer cancel()

	return gs.ReadAlarms(ctx)
}

func (al Alarms) ActualToJson() []byte {
	res, err := json.MarshalIndent(al.Actual, "", "  ")
	if err != nil {
		log.Println("Error marshalling Alarms to JSON:", err)
		return nil
	}
	return res
}

func (al Alarms) HistoryToJson() []byte {
	res, err := json.MarshalIndent(al.History, "", "  ")
	if err != nil {
		log.Println("Error marshalling Alarm History to JSON:", err)
		return nil
	}
	return res
}
