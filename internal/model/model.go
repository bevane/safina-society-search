package model

type SearchResponseVideos struct {
	Hits               []FormattedVideoHit `json:"hits"`
	EstimatedTotalHits int64               `json:"estimatedTotalHits,omitempty"`
	Offset             int64               `json:"offset,omitempty"`
	Limit              int64               `json:"limit,omitempty"`
	ProcessingTimeMs   int64               `json:"processingTimeMs"`
	Query              string              `json:"query"`
	FacetDistribution  interface{}         `json:"facetDistribution,omitempty"`
	TotalHits          int64               `json:"totalHits,omitempty"`
	HitsPerPage        int64               `json:"hitsPerPage,omitempty"`
	Page               int64               `json:"page,omitempty"`
	TotalPages         int64               `json:"totalPages,omitempty"`
	FacetStats         interface{}         `json:"facetStats,omitempty"`
	IndexUID           string              `json:"indexUid,omitempty"`
}

type FormattedVideoHit struct {
	VideoHit
	Formatted       VideoHit        `json:"_formatted"`
	MatchesPosition MatchesPosition `json:"_matchesPosition"`
}

type VideoHit struct {
	Id         string `json:"id"`
	Title      string `json:"title"`
	Transcript string `json:"transcript"`
}
type Result struct {
	Title        string
	Url          string
	ThumbnailUrl string
	Snippet      string
	MatchesCount int
}

type Results struct {
	Items []Result
}

type MatchesPosition struct {
	Title      []Position `json:"title"`
	Transcript []Position `json:"transcript"`
}

type Position struct {
	Start  int `json:"start"`
	Length int `json:"length"`
}
