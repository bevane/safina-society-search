package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bevane/safina-society-search/internal/model"
	"github.com/bevane/safina-society-search/internal/views"
	"github.com/meilisearch/meilisearch-go"
)

func (cfg *Config) handlerSearch(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("search")
	if len(query) <= 1 {
		views.Results(model.Results{}).Render(r.Context(), w)
		return
	}
	resRaw, _ := cfg.searchClient.Index("videos").SearchRaw(query, &meilisearch.SearchRequest{
		AttributesToCrop: []string{"transcript"},
		CropLength:       40,
	})

	searchResponse := model.SearchResponseVideos{}
	json.Unmarshal(*resRaw, &searchResponse)
	results := model.Results{
		Items: make([]model.Result, len(searchResponse.Hits)),
	}
	for i, hit := range searchResponse.Hits {
		results.Items[i] = model.Result{
			Title:        truncateString(hit.Title, 60),
			Url:          fmt.Sprintf("https://youtu.be/%s", hit.Id),
			ThumbnailUrl: hit.ThumbnailUrl,
			Snippet:      hit.Formatted.Transcript,
		}
	}
	views.Results(results).Render(r.Context(), w)

}

func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen < 3 {
		maxLen = 3
	}
	return string(runes[0:maxLen-3]) + "..."
}
