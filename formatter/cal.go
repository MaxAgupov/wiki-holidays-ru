package formatter

import (
	"strconv"
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

func GetDateString(day *time.Time) string {
	_, month, dayNum := day.Date()
	return strconv.Itoa(dayNum) + " " + monthsGenitive[month-1]
}

