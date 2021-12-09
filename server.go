package main

import (
	"html"
	"log"
	"net/http"
	"text/template"
)

type Query struct {
	Search  string
	Results []Result
	Number  int64
	Time    float64
}

func Homepage(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Println(err)
	}

	tmpl.Execute(w, nil)

}

func Search(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/search.html")
	result := &Query{}

	if err != nil {
		log.Println(err)
	}

	if value, ok := r.URL.Query()["q"]; ok {
		result.Search = html.EscapeString(value[0])
		ElasticSearch(result)
		//log.Println(result)
		if len(result.Results) == 0 {
			tmp, _ := template.ParseFiles("templates/no_results.html")
			tmp.Execute(w, result)
			return
		}

		tmpl.Execute(w, result)

	} else {
		tmpl.Execute(w, result)
	}

}

func Autocomplete(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	if value, ok := r.URL.Query()["q"]; ok {
		log.Println("got value", value[0])
		result, err := GetSuggestions(value[0])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("{Something went wrong}"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(result)
	}
}

func Crawler(w http.ResponseWriter, r *http.Request) {
	log.Println("HIT autocomplete")
}
