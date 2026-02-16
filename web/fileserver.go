package web

import (
	"log"
	"net/http"
	"os"
)

func PathAssets() http.HandlerFunc {

	workdir, _ := os.Getwd()
	log.Println("Serving static files from ", workdir)

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Serving %v", r.URL.Path)
		http.StripPrefix("/middleware/assets/", http.FileServer(http.Dir("./assets")))
	}
}
