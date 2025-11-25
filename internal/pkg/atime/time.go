package atime

import (
	"fmt"
	"time"
)

const (
	defaultTimeZone = TimezoneJakarta
)

var addedNanoSecond = Now().Nanosecond()

func CurrentTime() *time.Time {
	currentTime := Now()
	return &currentTime
}

func YesterdayTime() *time.Time {
	yesterday := Now().AddDate(0, 0, -1)
	return &yesterday
}

func Now() time.Time {
	loc := GetLocation()
	return time.Now().In(loc)
}

func GetLocation() *time.Location {
	loc, err := time.LoadLocation(defaultTimeZone)
	if err != nil {
		return Now().Location()
	}
	return loc
}

func NowZeroTime() (now time.Time, err error) {
	now, err = time.ParseInLocation(DateFormatYYYYMMDD, Now().Format(DateFormatYYYYMMDD), Now().Location())
	return
}

func NowWithTime() (now time.Time, err error) {
	now, err = time.ParseInLocation(DateFormatYYYYMMDDWithTime, Now().Format(DateFormatYYYYMMDDWithTime), Now().Location())
	return
}

func NowWithNanoTime() (now time.Time, err error) {
	now, err = time.ParseInLocation(time.RFC3339Nano, Now().Format(time.RFC3339Nano), Now().Location())
	return
}

func ParseStringToDatetime(layout, dateString string) (time.Time, error) {
	date, err := time.ParseInLocation(layout, dateString, Now().Location())
	return date, err
}

func FormatDatetimeToString(date time.Time, formatLayout string) string {
	return date.Format(formatLayout)
}

func ParseStringDateToDateWithTimeNow(layout, date string) (time.Time, error) {
	timeParam, err := ParseStringToDatetime(layout, date)
	if err != nil {
		return time.Time{}, err
	}
	datetime := time.Date(timeParam.Year(), timeParam.Month(), timeParam.Day(), Now().Hour(), Now().Minute(), Now().Second(), addedNanoSecond, Now().Location())

	return datetime, nil
}

// GetTotalDiffDayBetweenTwoDate ...
func GetTotalDiffDayBetweenTwoDate(dateFrom, dateTo time.Time) float64 {
	t1 := time.Date(dateFrom.Year(), dateFrom.Month(), dateFrom.Day(), 0, 0, 0, 0, dateFrom.Location())
	t2 := time.Date(dateTo.Year(), dateTo.Month(), dateTo.Day(), 0, 0, 0, 0, dateTo.Location())
	diff := (t2.Sub(t1).Hours() / 24) / 1

	return diff
}

func PrevMonth(t time.Time) (time.Time, time.Time) {
	t = ToZeroTime(t)
	t = BeginningOfMonth(t).AddDate(0, 0, -1)
	start := BeginningOfMonth(t)
	end := EndOfMonth(t)
	return start, end
}

func BeginningOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func EndOfMonth(t time.Time) time.Time {
	return BeginningOfMonth(t).AddDate(0, 1, 0).Add(-time.Second)
}

func FirstOfNextMonth(t time.Time) time.Time {
	return time.Date(
		t.Year(),
		t.Month()+1,
		1,
		0, 0, 0, 0,
		t.Location(),
	)
}

func StartDateEndDate(t1, t2 time.Time) (time.Time, time.Time) {
	t1 = time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, t1.Location())
	t2 = time.Date(t2.Year(), t2.Month(), t2.Day(), 23, 59, 59, 999999999, t1.Location())
	return t1, t2
}

func DateEqualToday(date time.Time) bool {
	loc := Now().Location()
	return date.In(loc).Format(DateFormatDDMMMYYYY) == Now().Format(DateFormatDDMMMYYYY)
}

func ToZeroTime(date time.Time) time.Time {
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	return date
}

func GenerateNextDate(start, end time.Time) ([]time.Time, error) {
	if start.After(end) {
		return nil, fmt.Errorf("end date must be greater than start date")
	}

	var dates []time.Time
	totalDiff := GetTotalDiffDayBetweenTwoDate(start, end)
	for i := 0; i <= int(totalDiff); i++ {
		dates = append(dates, start.AddDate(0, 0, i))
	}
	return dates, nil
}

func GenerateStartAndEndDateByMonthYear(year, month string) (startDate, endDate string) {

	dateStr := fmt.Sprintf("%s-%s-01", year, month)

	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		panic(err)
	}

	startDate = t.Format(DateFormatYYYYMMDD)
	endDate = t.AddDate(0, 1, -1).Format(DateFormatYYYYMMDD)

	return
}
