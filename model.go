package main

type Result struct {
	Title       string
	ThumbailUrl string
	Url         string
	Snippet     string
}

type Results struct {
	Items []Result
}
