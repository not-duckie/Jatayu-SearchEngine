package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch"
	"github.com/elastic/go-elasticsearch/esapi"
)

type Result struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Rank        int    `json:"rank"`
	Url         string `json:"url"`
	Favicon     string `json:"favicon"`
	TypeDoc     string `json:"typedoc"`
}

var es *elasticsearch.Client

func init() {
	var err error

	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://elastic:W3iC1Yr2yFv4dTQuBZw3@localhost:9200",
		},
	}

	es, err = elasticsearch.NewClient(cfg)

	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}
}

func did_you_mean(phrase string, image bool) (string, error) {

	var data, payload string

	var res *esapi.Response
	var err error

	if image {

		data = `{
				"size":0,
				"suggest": 
				{ "text":"%s", 
				"searchengine": 
				{ "phrase": { "field": "url" }}
				},
				"sort": {
					"_score": "desc"
				}
		}`

		payload = fmt.Sprintf(data, phrase)

		res, err = es.Search(
			es.Search.WithBody(strings.NewReader(payload)),
			es.Search.WithIndex("searchengine_images"),
		)

	} else {

		data = `{
				"size":0,
				"suggest": 
				{ "text":"%s", 
				"searchengine": 
				{ "phrase": { "field": "title" }}
				},
				"sort": {
					"_score": "desc"
				}
		}`

		payload = fmt.Sprintf(data, phrase)

		res, err = es.Search(
			es.Search.WithBody(strings.NewReader(payload)),
			es.Search.WithIndex("searchengine"),
		)
	}

	if err != nil {
		log.Println("something went wrong while suggestion", err)
		return "", err
	}

	buf, err := io.ReadAll(res.Body)

	if err != nil {
		log.Println("error while reading response", err)
		return "", err
	}
	defer res.Body.Close()

	var tmp interface{}
	json.Unmarshal(buf, &tmp)

	suggest := tmp.(map[string]interface{})["suggest"]

	results := suggest.(map[string]interface{})["searchengine"]

	options := results.([]interface{})[0].(map[string]interface{})["options"]

	if len(options.([]interface{})) == 0 {
		return "", fmt.Errorf("no suggestions")
	}

	output := options.([]interface{})[0].(map[string]interface{})["text"].(string)

	return output, nil
}

func GetSuggestions(query string) ([]byte, error) {
	data := `{
		  "_source": ["searchengine"],
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

	//es, _ := elasticsearch.NewDefaultClient()

	res, err := es.Search(
		es.Search.WithIndex("searchengine"),
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

func ImageSearch(result *Query, pagenum int) error {

	data := `{
			"from":"%v",
			"size":"10",
			"sort" : {
				"_score": "desc",
				"rank": "desc"
			},
			"query": {
				"multi_match" : {
				"query":    "%s", 
				"fields": [ "url" ] 
				}
			},
			"collapse": {
    			"field": "url.keyword"
  			},
			"aggs": {
			"total": {
			"cardinality": {
					"field": "url.keyword"
					}
				}
			}
		}
		`

	from := pagenum*10 - 10
	payload := fmt.Sprintf(data, from, html.EscapeString(result.Search))
	resp, err := es.Search(
		es.Search.WithIndex("searchengine_images"),
		es.Search.WithBody(strings.NewReader(payload)),
	)
	if err != nil {
		log.Println("error while searching", err)
		return err
	}

	buf, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Println("error while reading response", err)
		return err
	}
	defer resp.Body.Close()

	var tmp interface{}

	json.Unmarshal(buf, &tmp)

	result.Time = tmp.(map[string]interface{})["took"].(float64) / 1000

	hits := tmp.(map[string]interface{})["hits"]

	aggr := tmp.(map[string]interface{})["aggregations"]

	total := aggr.(map[string]interface{})["total"]

	result.Number = int64(total.(map[string]interface{})["value"].(float64))

	count := result.Number
	if result.Number == 0 {
		result.Pages = append(result.Pages, 1)
		result.Suggestion, err = did_you_mean(result.Search, true)

		log.Println(result.Suggestion)

		if err != nil {
			return err
		}
		return fmt.Errorf("no result found")
	}

	for i := int64(0); i < count; i = i + 10 {
		result.Pages = append(result.Pages, 1+i/10)
	}

	results := hits.(map[string]interface{})["hits"]

	res := Result{}

	for _, i := range results.([]interface{}) {
		source := (i.(map[string]interface{})["_source"])

		res.Url = source.(map[string]interface{})["url"].(string)

		result.Results = append(result.Results, res)

		//log.Println(result.Results)
	}
	return nil
}

func ElasticSearch(result *Query, pagenum int) error {

	data := `{
			"from":"%v",
			"size":"10",
			"sort" : {
				"_score": "desc",
				"rank": "desc"
			},
			"query": {
				"multi_match" : {
				"query":    "%s", 
				"fields": [ "title", "description","url" ] 
				}
			},
			"collapse": {
    			"field": "url.keyword"
  			},
			"aggs": {
				"total": {
				"cardinality": {
						"field": "url.keyword"
					}
				}
			}
		}
		`

	from := pagenum*10 - 10
	payload := fmt.Sprintf(data, from, html.EscapeString(result.Search))
	resp, err := es.Search(
		es.Search.WithIndex("searchengine"),
		es.Search.WithBody(strings.NewReader(payload)),
	)
	if err != nil {
		log.Println("error while searching", err)
		return err
	}

	buf, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Println("error while reading response", err)
		return err
	}

	defer resp.Body.Close()

	var tmp interface{}

	json.Unmarshal(buf, &tmp)

	result.Time = tmp.(map[string]interface{})["took"].(float64) / 1000

	hits := tmp.(map[string]interface{})["hits"]

	aggr := tmp.(map[string]interface{})["aggregations"]
	total := aggr.(map[string]interface{})["total"]

	result.Number = int64(total.(map[string]interface{})["value"].(float64))

	count := result.Number
	if result.Number == 0 {
		result.Pages = append(result.Pages, 1)
		result.Suggestion, err = did_you_mean(result.Search, false)

		log.Println(result.Suggestion)

		if err != nil {
			return err
		}
		return fmt.Errorf("no result found")
	}

	for i := int64(0); i < count; i = i + 10 {
		result.Pages = append(result.Pages, 1+i/10)
	}

	results := hits.(map[string]interface{})["hits"]

	res := Result{}

	for _, i := range results.([]interface{}) {
		source := (i.(map[string]interface{})["_source"])

		//wtf do i do with rank ?
		res.Rank = int(source.(map[string]interface{})["rank"].(float64))

		res.Title = source.(map[string]interface{})["title"].(string)
		res.Url = source.(map[string]interface{})["url"].(string)
		res.Description = source.(map[string]interface{})["description"].(string)
		res.Favicon = source.(map[string]interface{})["favicon"].(string)
		res.TypeDoc = source.(map[string]interface{})["typedoc"].(string)

		result.Results = append(result.Results, res)
	}
	return nil
}
