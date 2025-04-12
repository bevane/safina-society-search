package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/a-h/templ"
	"github.com/bevane/safina-society-search/internal/views"
	"github.com/joho/godotenv"
	"github.com/meilisearch/meilisearch-go"
)

type Config struct {
	searchClient meilisearch.ServiceManager
	port         int
}

func main() {
	app := Config{}
	godotenv.Load(".env")
	app.port, _ = strconv.Atoi(os.Getenv("PORT"))

	searchClient, err := meilisearch.Connect(os.Getenv("MEILISEARCH_URL"), meilisearch.WithAPIKey(os.Getenv("MEILISEARCH_API_KEY")))
	if err != nil {
		fmt.Printf("Unable to connect to meilisearch: %s\n", err.Error())
	}
	app.searchClient = searchClient

	publicHandler := http.StripPrefix("/public", http.FileServer(http.Dir("./public")))
	http.Handle("/", templ.Handler(views.Index()))
	http.Handle("/public/", publicHandler)
	http.HandleFunc("POST /search", app.handlerSearch)
	fmt.Printf("Server started on port %v\n", app.port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", app.port), nil))
}
