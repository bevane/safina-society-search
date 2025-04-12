package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

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
		AttributesToCrop:      []string{"transcript"},
		CropLength:            40,
		AttributesToHighlight: []string{"title", "transcript"},
		HighlightPreTag:       "<mark>",
		HighlightPostTag:      "</mark>",
		ShowMatchesPosition:   true,
	})

	searchResponse := model.SearchResponseVideos{}
	err := json.Unmarshal(*resRaw, &searchResponse)
	if err != nil {
		fmt.Println(err)
	}
	results := model.Results{
		Items: make([]model.Result, len(searchResponse.Hits)),
	}
	for i, hit := range searchResponse.Hits {
		// look for the timestamp from the middle of the cropped transcript
		// as the crop is centered around the word
		// it is still possible to miss the timestamp since the search term
		// is centered in regards to number of words instead of number of
		// chars
		timestampSeconds, err := getTimestampSecondsAtPosition(hit.Formatted.Transcript, len(hit.Formatted.Transcript)/2)
		if err != nil {
			fmt.Println(err)
		}
		results.Items[i] = model.Result{
			Title: hit.Formatted.Title,
			// construct url linking to timestamp of the crop/snippet
			Url:          fmt.Sprintf("https://youtu.be/%s&t=%s", hit.Id, timestampSeconds),
			ThumbnailUrl: fmt.Sprintf("https://i.ytimg.com/vi/%s/hqdefault.jpg", hit.Id),
			Snippet:      hit.Formatted.Transcript,
			MatchesCount: len(hit.MatchesPosition.Transcript),
		}
	}
	views.Results(results).Render(r.Context(), w)

}

func getTimestampSecondsAtPosition(text string, position int) (string, error) {
	var timestampStr string
	r, _ := regexp.Compile("((?:\\d?\\d?:)?\\d?\\d?:\\d?\\d?)")
	for position >= 0 {
		// avoid looking for timestamp if position is not at a space
		// or beginning of transcript to prevent capturing partial
		// matches of timestamp
		if position != 0 && string(text[position]) != " " {
			position--
			continue
		}
		// keep looking if timestamp is not found within the next 8 chars
		if !r.MatchString(text[position : position+9]) {
			position--
			continue
		}
		timestampStr = r.FindString(text[position : position+9])
		break
	}
	if timestampStr == "" {
		return "", fmt.Errorf("No timestamp found")
	}
	if len(timestampStr) < 4 || len(timestampStr) > 7 {
		return "", fmt.Errorf("timestamp malformed: %v", timestampStr)
	}
	var timeStampStrPadded string
	switch len(timestampStr) {
	case 4:
		timeStampStrPadded = "00:0" + timestampStr
	case 5:
		timeStampStrPadded = "00:" + timestampStr
	case 7:
		timeStampStrPadded = "0" + timestampStr
	default:
	}
	timestamp, err := time.Parse(time.TimeOnly, timeStampStrPadded)
	if err != nil {
		return "", err
	}
	reference, _ := time.Parse(time.TimeOnly, "00:00:00")
	timestampSeconds := timestamp.Sub(reference).Seconds()
	// offset by 2 seconds to avoid cases where the timestamp will start
	// right at the mention of the search query making it likely for the
	// user to miss the search query when watching the video
	timestampSeconds = max(timestampSeconds-2, 0)
	return strconv.Itoa(int(timestampSeconds)), nil
}
