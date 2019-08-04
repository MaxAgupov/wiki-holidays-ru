package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type Response struct {
	Batchcomplete string `json:"batchcomplete"`
	Query         Query  `json:"query"`
}

type Query struct {
	Pages map[string]Pages `json:"pages"`
}

type Pages struct {
	Title   string `json:"title"`
	Extract string `json:"extract"`
	PageId  uint64 `json:"pageid"`
	NS      uint64 `json:"ns"`
}

func getWikiReport(day string) string {
	wikiRequest := "https://ru.wikipedia.org/w/api.php?" +
		"action=query&format=json&&prop=extracts&exlimit=1&explaintext&titles=" + url.QueryEscape(day)

	log.Print(wikiRequest)
	response, err := http.Get(wikiRequest);
	if  err != nil {
		log.Print("Wikipedia is not respond",  err)
		return ""
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			log.Print(err)
		}
	}()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Print(err)
		return ""
	}
	var wr Response
	if err := json.Unmarshal(contents, &wr); err != nil {
		log.Print("Error", err)
		return ""
	}
	if l := len(wr.Query.Pages); l == 0 || l > 1 {
		log.Print("There must be only one page - ", l)
		return ""
	}
	var content string
	for _, v := range wr.Query.Pages {
		content = v.Extract
	}
	return content
}

var calendar = map[string]int{
	"января":   31,
	"февраля":  29,
	"марта":    31,
	"апреля":   30,
	"мая":      31,
	"июня":     30,
	"июля":     31,
	"августа":  31,
	"сентября": 30,
	"октября":  31,
	"ноября":   30,
	"декабря":  31,
}

func main() {

	for month, lastDate := range calendar {
		log.Print(month, lastDate)
	}
	resp := getWikiReport("1 декабря")

	log.Print(string(resp))
}
