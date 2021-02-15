package util

import (
	"strings"
	"time"

	"github.com/EducationEKT/xserver/x_utils/x_time"
)

func Now() int64 {
	return time.Now().UnixNano() / 1e6
}

func TodayStart() int64 {
	date := time.Now()
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	return date.UnixNano() / 1e6
}

func Date(days int) time.Time {
	date := time.Now()
	return time.Date(date.Year(), date.Month(), date.Day()+days, 0, 0, 0, 0, date.Location())
}

func DayStart(date time.Time) int64 {
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	return date.UnixNano() / 1e6
}

func Datestr() string {
	return time.Now().Format(x_time.DEFAULT_DATE)
}

func GetYestdayDateStr() string {
	return time.Now().AddDate(0, 0, -1).Format(x_time.DEFAULT_DATE)
}

func Get2dayBefore() string {
	return time.Now().AddDate(0, 0, -2).Format(x_time.DEFAULT_DATE)
}

//格式化http传输的字符串中间有%20
func HttpDatetimeStrFormat(datetime string) string{
	return strings.Replace(datetime,"%20"," ",1)
}