package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

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
	err := godotenv.Load(".env")
	if err != nil {
		slog.Info("No .env file available. Ensure the required env variables are set")
	}
	app.port, _ = strconv.Atoi(os.Getenv("PORT"))

	searchClient, err := meilisearch.Connect(os.Getenv("MEILISEARCH_URL"), meilisearch.WithAPIKey(os.Getenv("MEILISEARCH_API_KEY")))
	if err != nil {
		slog.Error("unable to connect to meilisearch", slog.Any("error", err))
		os.Exit(1)

	}
	app.searchClient = searchClient

	publicHandler := http.StripPrefix("/public", http.FileServer(http.Dir("./public")))
	http.Handle("/", templ.Handler(views.Index("", nil)))
	http.Handle("/public/", publicHandler)
	http.HandleFunc("GET /search", app.handlerSearch)
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", app.port),
		ReadHeaderTimeout: 3 * time.Second,
	}
	slog.Info(fmt.Sprintf("Server started on port %v\n", app.port))
	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
