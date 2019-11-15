package wiki

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

var monthsGenitive = [...]string{
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

var weekDays = [...]string{
	"воскресенье",
	"понедельник",
	"вторник",
	"среда",
	"четверг",
	"пятница",
	"суббота",
}

const holidaysHeader = "Праздники и памятные дни"
const intHolidaysSubheader = "Международные"
const locHolidaysSubheader = "Национальные"
const rlgHolidaysSubheader = "Религиозные"
const profHolidaysSubheader = "Профессиональные"
const nameDaysSubheader = "Именины"
const regHolidaysSubheader = "Региональные"

const MoscowLocation = "Europe/Moscow"

var reportCache = ReportCache{}

type Report struct {
	Stats        string
	Common       []string
	HolidaysInt  []string
	HolidaysLoc  []string
	HolidaysProf []string
	HolidaysRlg  ReligiousHolidays
	NameDays     []string
	Omens        []string
	//sections     map[string][]*Section
}

type ReligiousHolidayDescr struct {
	Descriptions []string
	GroupAbbr    string
}

type ReligiousHolidays struct {
	Holidays []*ReligiousHolidayDescr
}

func (holidays *ReligiousHolidays) Empty() bool {
	empty := true
	for _, item := range holidays.Holidays {
		if len(item.Descriptions) > 0 {
			empty = false
		}
	}
	return empty
}

func (holidays *ReligiousHolidays) AppendString(formatted *string) {
	if len(holidays.Holidays) > 0 {
		for _, items := range holidays.Holidays {
			for _, line := range items.Descriptions {
				*formatted += "- " + line
				if items.GroupAbbr != "" {
					*formatted += " (" + items.GroupAbbr + ")"
				}
				*formatted += "\n"
			}
		}
	}
}

//type Section struct {
//	header  string
//	content []string
//}

func (report *Report) String() string {
	formattedStr := ""
	if report.Stats != "" {
		formattedStr += report.Stats + "\n"
	}

	if len(report.HolidaysInt) > 0 || len(report.HolidaysLoc) > 0 || len(report.HolidaysProf) > 0 || !report.HolidaysRlg.Empty() {
		formattedStr += "*" + holidaysHeader + "*\n"
		if len(report.HolidaysInt) > 0 {
			formattedStr += "\n_" + intHolidaysSubheader + "_\n"
			for _, line := range report.HolidaysInt {
				formattedStr += "- " + line + "\n"
			}
		}
		if len(report.HolidaysLoc) > 0 {
			formattedStr += "\n_" + locHolidaysSubheader + "_\n"
			for _, line := range report.HolidaysLoc {
				formattedStr += "- " + line + "\n"
			}
		}
		if len(report.HolidaysProf) > 0 {
			formattedStr += "\n_" + profHolidaysSubheader + "_\n"
			for _, line := range report.HolidaysProf {
				formattedStr += "- " + line + "\n"
			}
		}
		if !report.HolidaysRlg.Empty() {
			formattedStr += "\n_" + rlgHolidaysSubheader + "_\n"
			report.HolidaysRlg.AppendString(&formattedStr)
		}
	}

	if len(report.NameDays) > 0 {
		formattedStr += "\n_" + nameDaysSubheader + "_"
		append := false
		for _, line := range report.NameDays {
			if strings.Contains(line, ":") {
				formattedStr += "\n- " + line
				append = false
			} else {
				if append {
					formattedStr += ", " + line
				} else {
					formattedStr += "\n- " + line
					append = true
				}
			}
		}
		formattedStr += "\n"
	}

	if l := len(report.Omens); l > 0 {
		formattedStr += "\n*" + "Приметы" + "*\n\n"
		for i, line := range report.Omens {
			if i > 0 && i < 5 {
				formattedStr += line + "\n"
			} else if i == 0 {
				formattedStr += "_" + line + "_\n"
			} else {
				break
			}
		}
	}
	return formattedStr
}

func (report *Report) SetCalendarInfo(day *time.Time) {
	report.Stats = GenerateCalendarStats(day)
}

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
	data := getDateString(reportDay)
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

func GetTodaysReport() string {
	location, _ := time.LoadLocation(MoscowLocation)
	log.Print(location)
	now := time.Now().In(location)
	report := reportCache.getCachedReport(&now)
	return report.String()
}

func getDateString(day *time.Time) string {
	_, month, dayNum := day.Date()
	return strconv.Itoa(dayNum) + " " + monthsGenitive[month-1]
}

func getFullDateString(day *time.Time) string {
	year, month, dayNum := day.Date()
	weekDay := strings.Title(getWeekDateString(day))
	return "*" + weekDay + ", " + strconv.Itoa(dayNum) + " " + monthsGenitive[month-1] + " " + strconv.Itoa(year) + " года" + "*"
}

func getWeekDateString(day *time.Time) string {
	weekday := int(day.Weekday())
	return weekDays[weekday]
}

func GetDayNoun(day int) string {
	rest := day % 10
	if (day > 10) && (day < 20) {
		// для второго десятка - всегда третья форма
		return "дней"
	} else if rest == 1 {
		return "день"
	} else if rest > 1 && rest < 5 {
		return "дня"
	} else {
		return "дней"
	}
}

func GenerateCalendarStats(reportDay *time.Time) string {
	firstLine := getFullDateString(reportDay)

	year := time.Date(reportDay.Year(), time.December, 31, 0, 0, 0, 0, time.UTC)
	infoDay := reportDay.YearDay()
	full_days := year.YearDay()

	rest := full_days - infoDay
	secondLine := ""
	if rest > 0 {
		secondLine = strconv.Itoa(infoDay) + "-й день года. До конца года " + strconv.Itoa(rest) + " " + GetDayNoun(rest)
	} else {
		secondLine = "Завтра уже Новый Год!"
	}

	return firstLine + "\n" + secondLine + "\n"
}

type ReportCache struct {
	sync.Mutex
	year   int
	month  time.Month
	day    int
	report *Report
}

func (cache *ReportCache) getCachedReport(date *time.Time) Report {
	cache.Lock()
	defer cache.Unlock()
	year, month, day := date.Date()
	if year == cache.year && month == cache.month && day == cache.day {
		return *cache.report
	}
	fullReport := getWikiReport(date)
	report, err := Parse(fullReport)
	if err != nil {
		log.Print("Error:", err)
		return Report{}
	}
	report.SetCalendarInfo(date)
	cache.report = &report
	cache.year = year
	cache.month = month
	cache.day = day

	return *cache.report
}
