package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/bevane/safina-society-search/views"
)

func main() {
	port := 3000
	publicHandler := http.StripPrefix("/public", http.FileServer(http.Dir("./public")))
	http.Handle("/", templ.Handler(views.Index()))
	http.Handle("/public/", publicHandler)
	fmt.Printf("Server started on port %v\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}
