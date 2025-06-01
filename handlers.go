package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
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
	page := params.Get("page")
	pageNumber, err := strconv.Atoi(page)
	isHTMX := r.Header.Get("Hx-Request") != ""
	slog.Info(fmt.Sprintf("GET /search: isHTMX: %v, params: %v", isHTMX, params))

	if len(query) == 0 {
		if isHTMX {
			quickStartComponent := views.QuickStart()
			err := quickStartComponent.Render(r.Context(), w)
			if err != nil {
				slog.Error("unable to render error component", slog.Any("error", err))
			}
		} else {
			err := views.Index(query, nil).Render(r.Context(), w)
			if err != nil {
				slog.Error("unable to render full html for empty query", slog.Any("error", err))
			}
		}
		return
	}
	if len(query) <= 2 {
		errComponent := views.InsufficientInput()
		if isHTMX {
			err := errComponent.Render(r.Context(), w)
			if err != nil {
				slog.Error("unable to render error component", slog.Any("error", err))
			}
		} else {
			err := views.Index(query, errComponent).Render(r.Context(), w)
			if err != nil {
				slog.Error("unable to render full html with error", slog.Any("error", err))
			}
		}
		return
	}

	if err != nil || pageNumber < 1 || pageNumber > 3 {
		errComponent := views.BadRequestPageNumber()
		if isHTMX {
			err = errComponent.Render(r.Context(), w)
			if err != nil {
				slog.Error("unable to render error component", slog.Any("error", err))
			}
		} else {
			err = views.Index(query, errComponent).Render(r.Context(), w)
			if err != nil {
				slog.Error("unable to render full html with error", slog.Any("error", err))
			}
		}
		return
	}

	results, totalPages, err := getSearchResults(query, pageNumber, cfg.searchClient)
	if err != nil {
		errComponent := views.InternalError()
		if isHTMX {
			err = errComponent.Render(r.Context(), w)
			if err != nil {
				slog.Error("unable to render error component", slog.Any("error", err))
			}
		} else {
			err = views.Index(query, errComponent).Render(r.Context(), w)
			if err != nil {
				slog.Error("unable to render full html with error", slog.Any("error", err))
			}
		}
		return
	}

	resultsComponent := views.Results(results, totalPages, pageNumber, query)
	if isHTMX {
		err = resultsComponent.Render(r.Context(), w)
		if err != nil {
			slog.Error("unable to render result component", slog.Any("error", err))
		}
	} else {
		err = views.Index(query, resultsComponent).Render(r.Context(), w)
		if err != nil {
			slog.Error("unable to full html with results", slog.Any("error", err))
		}
	}
}

func getSearchResults(query string, page int, searchClient meilisearch.ServiceManager) (model.Results, int, error) {
	resRaw, err := searchClient.Index("videos").SearchRaw(query, &meilisearch.SearchRequest{
		AttributesToCrop:      []string{"transcript"},
		CropLength:            70,
		AttributesToHighlight: []string{"title", "transcript"},
		HighlightPreTag:       "<mark>",
		HighlightPostTag:      "</mark>",
		ShowMatchesPosition:   true,
		Page:                  int64(page),
		HitsPerPage:           10,
	})
	if err != nil {
		slog.Error("unable to get search results from meilisearch", slog.Any("error", err))
		return model.Results{}, 0, err
	}

	searchResponse := model.SearchResponseVideos{}
	err = json.Unmarshal(*resRaw, &searchResponse)
	if err != nil {
		slog.Error("unable to unmarshal search results from meilisearch", slog.Any("error", err))
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
		if err != nil {
			slog.Error("unable to get timestamp in from transcript", slog.Any("error", err))
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
	return results, int(searchResponse.TotalPages), nil
}

func getTimestampSeconds(text string) (string, error) {
	if text == "" {
		return "", errors.New("error getting timestamp: text is empty")
	}
	var timestampStr string
	r, _ := regexp.Compile(`(\d{2}:\d{2}:\d{2}),\d{3} -->`)
	if !r.MatchString(text) {
		return "", errors.New("no timestamp found")
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
