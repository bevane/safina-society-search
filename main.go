package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/bevane/safina-society-search/internal/views"
	"github.com/meilisearch/meilisearch-go"
)

type Config struct {
	searchClient meilisearch.ServiceManager
	port         int
}

func main() {
	app := Config{}
	app.port = 3000
	searchClient := meilisearch.New("http://localhost:7700", meilisearch.WithAPIKey("aSampleMasterKey"))
	app.searchClient = searchClient

	publicHandler := http.StripPrefix("/public", http.FileServer(http.Dir("./public")))
	http.Handle("/", templ.Handler(views.Index()))
	http.Handle("/public/", publicHandler)
	http.HandleFunc("POST /search", app.handlerSearch)
	fmt.Printf("Server started on port %v\n", app.port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", app.port), nil))
}
