package main

type searchResponseVideos struct {
	Hits               []videoHit  `json:"hits"`
	EstimatedTotalHits int64       `json:"estimatedTotalHits,omitempty"`
	Offset             int64       `json:"offset,omitempty"`
	Limit              int64       `json:"limit,omitempty"`
	ProcessingTimeMs   int64       `json:"processingTimeMs"`
	Query              string      `json:"query"`
	FacetDistribution  interface{} `json:"facetDistribution,omitempty"`
	TotalHits          int64       `json:"totalHits,omitempty"`
	HitsPerPage        int64       `json:"hitsPerPage,omitempty"`
	Page               int64       `json:"page,omitempty"`
	TotalPages         int64       `json:"totalPages,omitempty"`
	FacetStats         interface{} `json:"facetStats,omitempty"`
	IndexUID           string      `json:"indexUid,omitempty"`
}

type videoHit struct {
	Id           string `json:"id"`
	ThumbnailUrl string `json:"thumbnail"`
	Title        string `json:"title"`
	Transcript   string `json:"transcript"`
}
type Result struct {
	Title        string
	Url          string
	ThumbnailUrl string
	Snippet      string
}

type Results struct {
	Items []Result
}
