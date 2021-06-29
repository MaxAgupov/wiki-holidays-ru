package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
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
	response, err := http.Get(wikiRequest)
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

type TypedDayHolidays struct {
	Month  time.Month
	Day    int
	Report wiki.Report
}

type Job struct {
	Month time.Month
	Day   int
	resp  chan *TypedDayHolidays
}

type MonthHolidays map[int]*DayHolidays
type Holidays map[time.Month]MonthHolidays

func loader(job chan *Job, wg *sync.WaitGroup) {

	for j := range job {
		date := strconv.Itoa(j.Day) + " " + monthsGenetive[j.Month-1]

		resp := getWikiReport(date)
		if resp == "" {
			log.Print(date)

			return
		}
		report, err := wiki.Parse(resp)
		if err != nil {
			log.Print("Error:", err)
			wg.Done()
			return
		}

		d := TypedDayHolidays{j.Month, j.Day, report}
		j.resp <- &d
		wg.Done()
	}
}

func main() {
	log.Println("Load Data from Wiki")
	var done = make(chan bool)

	var jobsNum = 20
	var jobs = make(chan *Job, jobsNum)

	var reports = Holidays{}
	var days = make(chan *TypedDayHolidays)
	var wg sync.WaitGroup

	for j := 0; j < jobsNum; j++ {
		go loader(jobs, &wg)
	}

	go func() {
		for d := range days {
			h := DayHolidays{d.Month.String(), strconv.Itoa(d.Day), d.Report}
			reports[d.Month][d.Day] = &h
		}
		done <- true
	}()

	for m := time.January; m <= time.December; m++ {
		month := MonthHolidays{}
		reports[m] = month
		for day := 1; day <= monthDays[m-1]; day++ {
			wg.Add(1)
			jobs <- &Job{m, day, days}
		}
	}

	wg.Wait()
	close(days)
	log.Println("Wait last results")
	<-done
	tmpFile, err := os.OpenFile("holidays.v1.17.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)

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
