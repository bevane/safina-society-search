package main

import (
	"net/http"

	"github.com/meilisearch/meilisearch-go"
)

func (cfg *Config) handlerSearch(w http.ResponseWriter, r *http.Request) {
	res, _ := cfg.searchClient.Index("videos").Search("man", &meilisearch.SearchRequest{})
	json, _ := res.MarshalJSON()
	w.Write(json)
}
