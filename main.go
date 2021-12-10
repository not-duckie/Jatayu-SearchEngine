package main

import (
	"net/http"
)

func main() {

	//	url := "https://www.ndtv.com/"

	//	crawler.InitiateCrawler(url)

	http.HandleFunc("/", Homepage)
	http.Handle("/static/", http.FileServer(http.Dir("./templates")))
	http.HandleFunc("/crawler", Crawler)
	http.HandleFunc("/search", Search)
	http.HandleFunc("/autocomplete", Autocomplete)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
