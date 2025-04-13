package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

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
	resRaw, err := cfg.searchClient.Index("videos").SearchRaw(query, &meilisearch.SearchRequest{
		AttributesToCrop:      []string{"transcript"},
		CropLength:            40,
		AttributesToHighlight: []string{"title", "transcript"},
		HighlightPreTag:       "<mark>",
		HighlightPostTag:      "</mark>",
		ShowMatchesPosition:   true,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	searchResponse := model.SearchResponseVideos{}
	err = json.Unmarshal(*resRaw, &searchResponse)
	if err != nil {
		fmt.Println(err)
	}
	results := model.Results{
		Items: make([]model.Result, len(searchResponse.Hits)),
	}
	for i, hit := range searchResponse.Hits {
		// will get the left most timestamp in the snippet
		timestampSeconds, err := getTimestampSeconds(hit.Formatted.Transcript)
		cleanedSnippet := cleanSnippet(hit.Formatted.Transcript)
		if err != nil {
			fmt.Println(err)
		}
		results.Items[i] = model.Result{
			Title: hit.Formatted.Title,
			// construct url linking to timestamp of the crop/snippet
			Url:          fmt.Sprintf("https://youtu.be/%s&t=%s", hit.Id, timestampSeconds),
			ThumbnailUrl: fmt.Sprintf("https://i.ytimg.com/vi/%s/hqdefault.jpg", hit.Id),
			Snippet:      cleanedSnippet,
			MatchesCount: len(hit.MatchesPosition.Transcript),
		}
	}
	views.Results(results).Render(r.Context(), w)

}

func getTimestampSeconds(text string) (string, error) {
	if text == "" {
		return "", fmt.Errorf("Error getting timestamp: text is empty")
	}
	var timestampStr string
	r, _ := regexp.Compile(`(\d{2}:\d{2}:\d{2}),\d{3} -->`)
	timestampStr = r.FindStringSubmatch(text)[1]
	if timestampStr == "" {
		return "", fmt.Errorf("No timestamp found")
	}
	timestamp, err := time.Parse(time.TimeOnly, timestampStr)
	if err != nil {
		return "", err
	}
	reference, _ := time.Parse(time.TimeOnly, "00:00:00")
	timestampSeconds := timestamp.Sub(reference).Seconds()
	return strconv.Itoa(int(timestampSeconds)), nil
}

func cleanSnippet(text string) string {
	var sb strings.Builder
	for i, char := range text {
		// skip the '>' in '-->'
		if i > 0 && text[i-1] == '-' {
			continue
		}
		// skip the char only if it is not a letter and space
		// and if it is not '<' '/' '>' so that the <mark> </mark> html
		// tags are preserved
		if !unicode.IsLetter(char) && char != ' ' && char != '<' && char != '>' && char != '/' {
			continue
		}
		sb.WriteRune(char)
	}
	return sb.String()
}
