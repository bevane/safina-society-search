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

	// if the user clears the search input, show the quickstart section again
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
	// only conduct search if user enters more than 2 chars
	// generally a user would not actually want to search words like "a" or "is"
	// there are cases where a user might want to actually search for 2 char words
	// for example 'AI', if the user wraps it in quotes it will work and also
	// give the results they want instead of returning results with words like m[ai]n
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

	// prevents the user from getting an invalid page by editing the url
	// the max number of pages must coincide with the maxTotalHits (configured in meilisearch instance)
	// divided by HitsPerPage (defined in search request sent to meiliesearch)
	if err != nil || pageNumber < 1 || pageNumber > 5 {
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
		// crop to show a snippet for each search result
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
		if err != nil {
			slog.Error("unable to get timestamp in from transcript", slog.Any("error", err))
		}
		// remove timestamps and anything that are not subtitles from
		// the snippet
		cleanedSnippet := cleanSnippet(hit.Formatted.Transcript)
		results.Items[i] = model.Result{
			Title: hit.Formatted.Title,
			// construct url linking to timestamp of the crop/snippet
			Url:          fmt.Sprintf("https://youtu.be/%s&t=%s", hit.Id, timestampSeconds),
			ThumbnailUrl: fmt.Sprintf("https://i.ytimg.com/vi/%s/hqdefault.jpg", hit.Id),
			Snippet:      cleanedSnippet,
			// number of occurences of search term in the video
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
	// regex to match srt timing "00:20:30,50 -->" and capture the time only
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
	// timestamp is an absolute time, so subtract reference from it to get
	// a delta time which can be converted into seconds
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
		// write characters that were not skipped into new string
		sb.WriteRune(char)
	}
	return sb.String()
}
