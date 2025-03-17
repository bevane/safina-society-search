package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := 3000
	appHandler := http.FileServer(http.Dir("./app"))
	assetsHandler := http.StripPrefix("/assets", http.FileServer(http.Dir("./assets")))
	http.Handle("/", appHandler)
	http.Handle("/assets/", assetsHandler)
	fmt.Println("Server started")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}
