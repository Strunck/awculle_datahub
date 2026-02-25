package globals

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Telegram struct {
	Bot                *tgbotapi.BotAPI
	ChatID             int64
	Token              string
	Reachable          bool
	SkipStartupMessage bool
	err                error
}

func (gs *GlobalState) InitTelegramBot(SkipTelegramStartup bool) (err error) {
	gs.Telegram.Reachable = false
	gs.Telegram.SkipStartupMessage = SkipTelegramStartup

	gs.Telegram.Token = os.Getenv("TELGRAMM_API")
	cid := os.Getenv("TELGRAMM_CHATID")
	if len(cid) == 0 {
		return fmt.Errorf("📨 TELGRAMM_CHATID environment variable is not set, telegram bot will not work")
	}
	gs.Telegram.ChatID, err = strconv.ParseInt(cid, 10, 64)
	if err != nil {
		return fmt.Errorf("📨 Error parsing TELGRAMM_CHATID: %w", err)
	}

	gs.Telegram.Bot, err = tgbotapi.NewBotAPI(gs.Telegram.Token)
	if err != nil {
		return err
	}
	gs.Telegram.Bot.Debug = true
	log.Printf("📨 Authorized on account %s 📨", gs.Telegram.Bot.Self.UserName)

	log.Println("📨 Telegram Bot initialized 📨")

	if gs.Telegram.SkipStartupMessage {
		log.Println("📨 Skipping startup message (--skip-telegram-startup flag enabled)")
		gs.Telegram.Reachable = true
		return nil
	}

	// Test the preconfigured ChatID
	txt := fmt.Sprintf("📨 Middleware gestartet. Site: %s - SPS: %s", gs.Cfg.Plc.Base, gs.Cfg.Plc.IPAddress)
	msg := tgbotapi.NewMessage(gs.Telegram.ChatID, txt)
	_, err = gs.Telegram.Bot.Send(msg)
	if err != nil {
		return fmt.Errorf("📨 Error sending test message to preconfigured ChatID: %w", err)
	} else {
		gs.Telegram.Reachable = true
	}
	return nil
}

// TODO Routine to regualarly check if telegram is reachable
func (gs *GlobalState) IsTelegramReachable() {
	if gs.Telegram.Reachable { //early return if it is reachable
		return
	}

	// try to get the chatId, if successful, set reachable to true
	// TODO implement

}

/*
Details
The msg.payload must be an object that contains a complete set of telegram message properties, at a minimum these should contain:

content the message contents
type the type of message contents
chatId the chatId number or an array of chatIds
*/
func (gs *GlobalState) SendTelegramMessage(msg string) (ok bool) {
	// try to send a telegram message
	if !gs.Telegram.Reachable {
		log.Println("📨 Telegram Bot is not working!!!") //just log it for now, figure out later
		return false
	}

	if len(msg) == 0 {
		log.Println("📨 Empty Telegram message, not sending that!!!")
		return false
	}

	message := tgbotapi.NewMessage(gs.Telegram.ChatID, msg)
	_, err := gs.Telegram.Bot.Send(message)
	if err != nil {
		gs.Telegram.Reachable = false
		log.Println("📨 Error sending Telegram message:", err)
		return false
	}
	return true
}

func PostTelegramMsg(gs *GlobalState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		msg := make([]byte, r.ContentLength)
		r.Body.Read(msg)
		log.Printf("Received Telegram message: %s", string(msg))
		if len(msg) == 0 {
			http.Error(w, "empty message in body", http.StatusBadRequest)
			return
		}
		ok := gs.SendTelegramMessage(string(msg))
		w.Header().Set("Content-Type", "application/json")
		if ok {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"status":"error"}`))
		}
	}
}

func PathTelegramState(gs *GlobalState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		state := map[string]bool{"reachable": false}

		if gs.Telegram.Reachable {
			state["reachable"] = true
		}
		json.NewEncoder(w).Encode(state)
	}
}
