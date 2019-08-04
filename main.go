package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
	"wikiholidays/formatter"
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

func getWikiReport(reportDay *time.Time) string {
	wikiRequest := "https://ru.wikipedia.org/w/api.php?action=query&format=json&&prop=extracts&exlimit=1&explaintext"
	data := formatter.GetDateString(reportDay)
	wikiRequest += "&titles=" + url.QueryEscape(data)

	log.Print(wikiRequest)
	if response, err := http.Get(wikiRequest); err != nil {
		log.Print("Wikipedia is not respond")
	} else {
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
	return ""
}

func main() {

}
