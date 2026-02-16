// Implement a websocket server, that responds to client requests
//

package web

import (
	"log"
	"net/http"
	"strunck/awculle_datahub/globals"
	"strunck/awculle_datahub/initialize"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WSPathEcho(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Echoes incoming messages back to the client
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			break
		}
		err = conn.WriteMessage(messageType, message)
		if err != nil {
			break
		}
	}
}

func WSPathAlarms(gs *globals.GlobalState) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		log.Println("⚠️ Client connected to Alarms", r.Host)

		// Periodically send the contents of Konfig.json
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		alarmList := make([]initialize.Alarm, 0)
		var count int = 0
		for _, one := range gs.Alarms.Actual {
			count++
			if len(one.Alarm.Key) > 0 {
				alarmList = append(alarmList, one.Alarm)
				log.Printf("WSPathAlarms AlarmList: %v", alarmList)
			}
		}
		err = conn.WriteMessage(websocket.TextMessage, []byte("[]")) // send immediately once
		if err != nil {
			log.Println("⚠️ Initial Empty AlarmList could not be sent:", err)
			return
		}

		for range ticker.C {
			if len(alarmList) > 1 {
				if err = conn.WriteJSON(alarmList); err != nil {
					log.Println("⚠️ Error sending ticking AlarmList:", err)
					return
				}
			} else {
				// Write an empty array to keep the connection alive
				if err = conn.WriteMessage(websocket.TextMessage, []byte("[]")); err != nil {
					log.Println("⚠️ Error sending empty AlarmList:", err)
					return
				}
			}
		}
	}
}

func WSPathDATA(gs *globals.GlobalState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		log.Println("Client connected to DATA", r.Host)

		// Periodically send the contents of the HMI data
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			err = conn.WriteJSON(gs.Plc.Flat)
			if err != nil {
				log.Println("Error sending message:", err)
				return
			}
		}
	}
}

// WebSocket Path /STRINGS
func WSPathSTRINGS(gs *globals.GlobalState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		log.Printf("Client connected to STRINGS: %s\n", r.Host)

		// Periodically send the contents of Konfig.json
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for range ticker.C {

			err = conn.WriteJSON(gs.Cfg.Plc.SWSettings)
			if err != nil {
				log.Println("Error sending message:", err)
				return
			}
		}
	}
}

/*
String, that helps to set up Grafana Dashboard's
*/
func WSPathRaum(gs *globals.GlobalState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		log.Printf("Client connected to RAUM: %s\n", r.Host)

		// Periodically send the contents of Konfig.json
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for range ticker.C {

			err = conn.WriteMessage(websocket.TextMessage, []byte(gs.Cfg.Plc.GrafanRaum))
			if err != nil {
				log.Println("Error sending message:", err)
				return
			}
		}
	}
}

func WSPathDATAv2(gs *globals.GlobalState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		log.Println("Client connected to DATAv2", r.Host)

		// Periodically send the contents of the HMI data
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			err = conn.WriteJSON(gs.DataV2.Scope)
			if err != nil {
				log.Println("Error sending message:", err)
				return
			}
		}
	}
}
