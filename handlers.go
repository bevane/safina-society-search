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
	params := r.URL.Query()
	query := params.Get("q")

	if len(query) <= 1 {
		views.InsufficientInput().Render(r.Context(), w)
		return
	}
	resRaw, err := cfg.searchClient.Index("videos").SearchRaw(query, &meilisearch.SearchRequest{
		AttributesToCrop:      []string{"transcript"},
		CropLength:            200,
		AttributesToHighlight: []string{"title", "transcript"},
		HighlightPreTag:       "<mark>",
		HighlightPostTag:      "</mark>",
		ShowMatchesPosition:   true,
	})
	if err != nil {
		views.InternalError().Render(r.Context(), w)
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
		// remove timestamps and anything that are not subtitles from
		// the snippet
		cleanedSnippet := cleanSnippet(hit.Formatted.Transcript)
		// sometimes timestamps crowd the snippet and only a few words
		// remain after cleaning it due to a bug in the transcription
		// process which makes timestamps word by word instead of
		// making timestamps for long sentences
		// To mitigate this problem, we intially get a very large snippet
		// and then truncate it after removing the timestamps so that
		// more words will remain in the snippet
		cleanedAndTruncatedSnippet := truncateSnippetAroundCenter(cleanedSnippet, 40)
		if err != nil {
			fmt.Println(err)
		}
		results.Items[i] = model.Result{
			Title: hit.Formatted.Title,
			// construct url linking to timestamp of the crop/snippet
			Url:          fmt.Sprintf("https://youtu.be/%s&t=%s", hit.Id, timestampSeconds),
			ThumbnailUrl: fmt.Sprintf("https://i.ytimg.com/vi/%s/hqdefault.jpg", hit.Id),
			Snippet:      cleanedAndTruncatedSnippet,
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
	if !r.MatchString(text) {
		return "", fmt.Errorf("No timestamp found")
	}
	timestampStr = r.FindStringSubmatch(text)[1]
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
		if i > 0 && char == '>' && text[i-1] == '-' {
			continue
		}
		// skip the second '-' in '-->'
		if i > 0 && char == '-' && text[i-1] == '-' {
			continue
		}
		// skip the first '-' in '-->'
		if i > 0 && char == '-' && text[i-1] == ' ' {
			continue
		}
		if unicode.IsNumber(char) || char == ':' || char == ',' {
			continue
		}
		sb.WriteRune(char)
	}
	return sb.String()
}

func truncateSnippetAroundCenter(text string, cropLength int) string {
	words := strings.Fields(text)
	if len(words) <= cropLength {
		return text
	}
	mid := len(words) / 2
	start := mid - cropLength/2
	end := mid + cropLength/2
	truncatedWords := words[start:end]
	return strings.Join(truncatedWords, " ")
}
