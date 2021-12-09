package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch"
	"github.com/olivere/elastic/v7"
)

type Result struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Rank        int    `json:"rank"`
	Url         string `json:"url"`
}

var es *elastic.Client

func init() {
	var err error
	es, err = elastic.NewClient(elastic.SetURL("http://127.0.0.1:8081"))
	if err != nil {
		panic("Elasitc Search Server Is Down!!!!!!\n " + err.Error())
	}
	log.Println("es initialized")
}

func GetSuggestions(query string) ([]byte, error) {
	data := `{
		  "_source": ["educative"],
		  "size": 0,
		  "min_score": 0.5,
		  "query": {
		    "bool": {
		      "must": [
		        {
		          "match_phrase_prefix": {
		            "title": {
		              "query": "%s"
		            }
		          }
		        }
		      ],
		      "filter": [],
		      "should": [],
		      "must_not": []
		    }
		  },
		  "aggs": {
		    "auto_complete": {
		      "terms": {
		        "field": "title.keyword",
		        "order": {
		          "_count": "desc"
		        },
		        "size": 5
		      }
		    }
		  }
		}`
	payload := fmt.Sprintf(data, html.EscapeString(query))

	es, _ := elasticsearch.NewDefaultClient()

	res, err := es.Search(
		es.Search.WithBody(strings.NewReader(payload)),
	)
	if err != nil {
		log.Printf("Error getting the response: %s", err)
	}
	buf, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	//umarshling this hell demon

	var tmp interface{}

	json.Unmarshal(buf, &tmp)

	aggs := (tmp.(map[string]interface{}))["aggregations"]
	buckets := (((aggs.(map[string]interface{}))["auto_complete"]).(map[string]interface{}))["buckets"]

	result, err := json.Marshal(buckets)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return result, nil

}

func ElasticSearch(result *Query) {
	c := context.Background()

	esQuery := elastic.NewMultiMatchQuery(result.Search, "title", "Description", "url").
		Fuzziness("2").
		MinimumShouldMatch("2")

	searchResult, err := es.Search().
		Index("educative").
		Query(esQuery).
		Sort("_score", false).
		Sort("rank", false).
		From(0).Size(10).Do(c)

	if err != nil {
		log.Println(err)
	}
	if searchResult.Hits.TotalHits.Value > 0 {
		result.Number = searchResult.Hits.TotalHits.Value
		result.Time = float64(searchResult.TookInMillis) / 1000

		// Iterate through results
		for _, hit := range searchResult.Hits.Hits {
			// hit.Index contains the name of the index

			// Deserialize hit.Source into a Tweet (could also be just a map[string]interface{}).
			var t Result
			err := json.Unmarshal(hit.Source, &t)
			if err != nil {
				// Deserialization failed
				log.Println("deserilisation failed")
			}
			//t.Title = html.UnescapeString(t.Title)
			//t.Description = html.UnescapeString(t.Description)

			result.Results = append(result.Results, t)
		}
	} else {
		// No hits
		fmt.Print("not hits")
	}
}
