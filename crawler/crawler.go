package crawler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type MetaData struct {
	Rank int `json:"rank"`

	TypeDoc     string `json:"typedoc"`
	Url         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Favicon     string `json:"favicon"`
}

func checkImage(contenttype string) bool {
	log.Println(contenttype)
	for _, i := range []string{"octet-stream", "jpg", "jpeg", "png", "bmp", "svg", "ico", "JPG", "JPEG", "PNG", "BMP", "SVG", "ICO", "srt", "SRT"} {
		if ok := strings.HasSuffix(contenttype, i); ok {
			return true
		}
	}
	return false
}

func fetchMeta(page string, meta *MetaData) (bool, error) {
	if !(strings.Contains(page, "http") || strings.Contains(page, "https")) {
		page = "https:" + page
	}
	rank := 1

	log.Println("crawling ", page)

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(page)
	if err != nil {
		log.Println("Something went wrong file feteching ", page)
		return false, err
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("page not found")
	}

	var title, description, favicon string

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	//fetching title
	doc.Find("title").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		title = s.Text()
		return false
	})

	//feteching description
	doc.Find("meta").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		if val, ok := s.Attr("name"); ok {
			if val == "description" || val == "twitter:description" {
				description = s.AttrOr("content", "")
				return false
			}
		}
		return true
	})

	//fetching favicon
	doc.Find("link").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		if val, ok := s.Attr("rel"); ok {
			if url, ok := s.Attr("href"); ok && val == "shortcut icon" {
				favicon = url
				return false
			}
		}
		return true
	})

	if string(title) != "" {
		rank = rank + 3
	}
	if string(description) != "" {
		rank = rank + 2
	}

	meta.Url = page
	meta.Title = string(title)
	meta.Description = string(description)
	meta.Favicon = favicon
	meta.Rank = rank

	if checkImage(resp.Header["Content-Type"][0]) {
		return true, nil
	} else {
		return false, nil
	}
}

func fetchUrl(url string) (map[string]bool, error) {

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)

	if err != nil {
		log.Println("Something went wrong while fetching ", url)
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to read response of ", url)
		return nil, err
	}
	defer resp.Body.Close()

	findUrl := regexp.MustCompile("(http|https|)[:]*//[/a-zA-Z0-9_.-]+")
	tmp := findUrl.FindAll(data, -1)
	urlList := make(map[string]bool)
	for _, i := range tmp {
		urlList[string(i)] = false
	}
	return urlList, nil
}

func sendToElastic(meta *MetaData, image bool) error {
	var url string
	if image {
		url = "http://127.0.0.1:9200/searchengine_images/_doc"
	} else {
		url = "http://127.0.0.1:9200/searchengine/_doc"
	}
	data, err := json.Marshal(meta)
	if err != nil {
		log.Println("unable to json encode the payload ", meta)
	}

	log.Println(string(data))

	client := http.Client{}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		log.Println(err)
		return err
	}

	// set the request header Content-Type for json
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Basic ZWxhc3RpYzpXM2lDMVlyMnlGdjRkVFF1Qlp3Mw==")
	_, err = client.Do(req)
	if err != nil {
		log.Println("Error sending request to elastic search")
		return err
	}
	return nil
}

func InitiateCrawler(url string) error {
	meta := &MetaData{}
	log.Println("Starting crawler for url := ", url)

	//urlList := append(fetchUrl(url), []byte(url))
	urlList, err := fetchUrl(url)
	if err != nil {
		return err
	}
	urlList[url] = false
	for page := range urlList {
		img, err := fetchMeta(page, meta)

		if err != nil {
			log.Printf("skiping %v.\nreason %v", page, err)
		}
		if err := sendToElastic(meta, img); err != nil {
			return err
		}
		//log.Println(meta)
	}
	return nil
}
