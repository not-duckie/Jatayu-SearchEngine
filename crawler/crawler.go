package crawler

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type MetaData struct {
	Rank        int    `json:"rank"`
	Url         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func fetchMeta(page string, meta *MetaData) {
	if !(strings.Contains(page, "http") || strings.Contains(page, "https")) {
		page = "https:" + page
	}
	rank := 1

	log.Println("crawling ", page)
	resp, err := http.Get(page)
	if err != nil {
		log.Println("Something went wrong file feteching ", page)
		return
	}
	if resp.StatusCode == http.StatusNotFound {
		return
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error while reading the body")
		return
	}
	defer resp.Body.Close()

	titleRegex := regexp.MustCompile(`<title[^<>]*>[^<>]+</title>`)
	descrpRegex := regexp.MustCompile(`<meta[^><]+name=("|')+(d|D)escription"* content="?[^"><]+"?`)

	clean1Desc := regexp.MustCompile(`<meta[^><]+name="?(d|D)escription"? (c|C)ontent="`)
	clean2Desc := regexp.MustCompile(`".*`)

	clean1Title := regexp.MustCompile(`<(t|T)itle[^<>]*>`)
	clean2Title := regexp.MustCompile(`</(t|T)itle>`)

	title := titleRegex.Find(data)
	description := descrpRegex.Find(data)

	description = []byte(clean1Desc.ReplaceAllString(string(description), ""))
	description = []byte(clean2Desc.ReplaceAllString(string(description), ""))

	title = []byte(clean1Title.ReplaceAllString(string(title), ""))
	title = []byte(clean2Title.ReplaceAllString(string(title), ""))

	//log.Printf("%s", title)

	if string(title) != "" {
		rank = rank + 3
	}
	if string(description) != "" {
		rank = rank + 2
	}
	//log.Printf("%s", description)

	meta.Url = page
	meta.Title = string(title)
	meta.Description = string(description)
	meta.Rank = rank

	if string(title) == "" {
		return
	}
}

func fetchUrl(url string) map[string]bool {
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Something went wrong while fetching ", url)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to read response of ", url)
	}
	defer resp.Body.Close()

	findUrl := regexp.MustCompile("(http|https|)[:]*//[/a-zA-Z0-9_.-]+")
	tmp := findUrl.FindAll(data, -1)
	urlList := make(map[string]bool)
	for _, i := range tmp {
		urlList[string(i)] = false
	}
	return urlList
}

func sendToElastic(meta *MetaData) {
	url := "http://127.0.0.1:9200/educative/_doc"
	data, err := json.Marshal(meta)
	if err != nil {
		log.Println("unable to json encode the payload ", meta)
	}

	log.Println(string(data))

	client := http.Client{}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		panic(err)
	}

	// set the request header Content-Type for json
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	_, err = client.Do(req)
	if err != nil {
		panic("Error sending request to elastic search")
	}

}

func InitiateCrawler(url string) {
	meta := &MetaData{}
	log.Println("Starting crawler for url := ", url)

	//urlList := append(fetchUrl(url), []byte(url))
	urlList := fetchUrl(url)
	urlList[url] = false
	for page := range urlList {
		fetchMeta(page, meta)
		sendToElastic(meta)
		//log.Println(meta)
	}
}
