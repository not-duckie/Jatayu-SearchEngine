package main

import (
	"net/http"
)

func main() {

	http.HandleFunc("/", Homepage)
	http.Handle("/static/", http.FileServer(http.Dir("./templates")))
	http.HandleFunc("/crawler", Crawler)
	http.HandleFunc("/search", Search)
	http.HandleFunc("/autocomplete", Autocomplete)
	http.HandleFunc("/images", Image)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
