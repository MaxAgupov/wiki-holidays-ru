package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
	"wikiholidays/wiki"
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

	//log.Print(wikiRequest)
	response, err := http.Get(wikiRequest);
	if err != nil {
		log.Print("Wikipedia is not respond", err)
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

var monthsGenetive = [...]string{
	"января",
	"февраля",
	"марта",
	"апреля",
	"мая",
	"июня",
	"июля",
	"августа",
	"сентября",
	"октября",
	"ноября",
	"декабря",
}
var monthDays = [...]int{
	31,
	29,
	31,
	30,
	31,
	30,
	31,
	31,
	30,
	31,
	30,
	31,
}

type DayHolidays struct {
	Month  string      `json:"month"`
	Day    string      `json:"day"`
	Report wiki.Report `json:"report"`
}

type MonthHolidays map[int]*DayHolidays
type Holidays map[time.Month]*MonthHolidays

func main() {
	var reports = Holidays{}

	for m := time.January; m <= time.January; m++ {
		month := MonthHolidays{}
		reports[m] = &month
		for day := 1; day <= monthDays[m-1]; day++ {
			date := strconv.Itoa(day) + " " + monthsGenetive[m-1]
			log.Print(date)
			resp := getWikiReport(date)
			if resp == "" {
				log.Print(date)
				break
			}

			report, err := wiki.Parse(resp)
			if err != nil {
				log.Print("Error:", err)
				return
			}
			location, _ := time.LoadLocation(wiki.MoscowLocation)
			log.Print(location)
			now := time.Now().In(location)

			dStatInfo := time.Date(now.Year(), m, day, 0, 0, 0, 0, time.UTC)
			report.SetCalendarInfo(&dStatInfo)

			d := DayHolidays{m.String(), strconv.Itoa(day), report}
			month[day] = &d
		}
	}
	tmpFile, err := os.OpenFile("holidays.v1.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)

	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := tmpFile.Close(); err != nil {
			log.Print(err)
		}
	}()
	repJ, err := json.MarshalIndent(reports, "", " ")

	if err != nil {
		log.Fatal(err)
	}

	tmpFile.Write(repJ)

	log.Printf("Len: %d, filename=%s", len(reports), tmpFile.Name())
}
