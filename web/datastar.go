package web

// Wrapper for datastar SSE sdk
import (
	"fmt"
	"log"
	"net/http"
	"strunck/uv4-middleware/globals"
	"strunck/uv4-middleware/view"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
	"github.com/starfederation/datastar-go/datastar"
)

func DataStarHandler(gs *globals.GlobalState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		skey := chi.URLParam(r, "symbol")
		symbolVal, err := gs.DataV2.GetSymbolKey(skey)
		if err != nil {
			http.Error(w, "Symbol not found: "+skey, http.StatusNotFound)
			return
		}

		v := fmt.Sprintf("%#v", symbolVal)

		sse := datastar.NewSSE(w, r)

		log.Println(skey, v)

		if err := sse.PatchElements(`<code id=` + skey + `>` + symbolVal.V.StringValue + `</code>`); err != nil {
			http.Error(w, "Error initializing DataStar SSE", http.StatusInternalServerError)
			return
		}
	}
}

const LatestMessageCount = 30

// DataStarLatestMsgHandler serves the latest NATS message for debugging
func DataStarLatestMsgHandler(gs *globals.GlobalState) http.HandlerFunc {
	log.Println("LatestMesgHandler")
	return func(w http.ResponseWriter, r *http.Request) {
		sse := datastar.NewSSE(w, r)

		msgBufferChannel := make(chan *nats.Msg, LatestMessageCount)

		nsub, err := gs.ChanLatestNatsMessage(msgBufferChannel)
		if err != nil {
			http.Error(w, "error subscribing to NATS messages", http.StatusInternalServerError)
			return
		}
		defer nsub.Unsubscribe()

		// Get latest NATS messages, until r.Context() is done
		for {
			select {
			case <-r.Context().Done():
				log.Println("DataStarLatestMsgHandler: r.Context() done, client disconnected")
				return

			case msg := <-msgBufferChannel:
				// Start with the message that triggered this case
				strBuffer := []view.LatestNatsMessage{
					{
						Subject: msg.Subject,
						Data:    string(msg.Data),
					},
				}

				// Non-blocking drain of any additional messages
			drainLoop:
				for len(strBuffer) < LatestMessageCount {
					select {
					case additionalMsg := <-msgBufferChannel:
						strBuffer = append(strBuffer, view.LatestNatsMessage{
							Subject: additionalMsg.Subject,
							Data:    string(additionalMsg.Data),
						})
					default:
						// No more messages ready, break out
						break drainLoop
					}
				}

				if err := sse.PatchElementTempl(view.LatestNatsMessages(strBuffer)); err != nil {
					http.Error(w, "Error initializing DataStar SSE", http.StatusInternalServerError)
					return
				}
			}
		}
	}
}

/*
DataFlow
	Browser SSE Request (GET /ds/{symbol})
		-> subscribe to nats topic
			-> nats message recieved
				-> Trigger SSE update to clients
*/
