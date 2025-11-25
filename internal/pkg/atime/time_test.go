package atime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestATime(t *testing.T) {
	now := Now()
	t.Log(now)

	currentTime := CurrentTime()
	t.Log(currentTime)

	yesterdayTime := YesterdayTime()
	t.Log(yesterdayTime)

	getLocation := GetLocation()
	t.Log(getLocation)

	nowZeroTime, err := NowZeroTime()
	assert.NoError(t, err)
	t.Log(nowZeroTime)

	nowWithTime, err := NowWithTime()
	assert.NoError(t, err)
	t.Log(nowWithTime)

	nowWithNanoTime, err := NowWithNanoTime()
	assert.NoError(t, err)
	t.Log(nowWithNanoTime)

	parseStringToDatetime, err := ParseStringToDatetime(DateFormatYYYYMMDD, DateFormatYYYYMMDD)
	assert.NoError(t, err)
	t.Log(parseStringToDatetime)

	parseStringToDatetime, err = ParseStringToDatetime(DateFormatYYYYMMDDWithTime, DateFormatYYYYMMDD)
	assert.Error(t, err)
	t.Log(parseStringToDatetime)

	formatDatetimeToString := FormatDatetimeToString(now, DateFormatYYYYMMDD)
	t.Log(formatDatetimeToString)

	parseStringDateToDateWithTimeNow, err := ParseStringDateToDateWithTimeNow(DateFormatYYYYMMDD, DateFormatYYYYMMDD)
	assert.NoError(t, err)
	t.Log(parseStringDateToDateWithTimeNow)

	parseStringDateToDateWithTimeNow, err = ParseStringDateToDateWithTimeNow(DateFormatYYYYMMDDWithTime, DateFormatYYYYMMDD)
	assert.Error(t, err)
	t.Log(parseStringDateToDateWithTimeNow)

	getTotalDiffDayBetweenTwoDate := GetTotalDiffDayBetweenTwoDate(now, now.AddDate(0, 0, 7))
	assert.Equal(t, getTotalDiffDayBetweenTwoDate, float64(7))
	t.Log(formatDatetimeToString)

	start, end := PrevMonth(now)
	t.Log(start, end)

	start = BeginningOfMonth(now)
	t.Log(start)

	end = EndOfMonth(now)
	t.Log(end)

	start, end = StartDateEndDate(start, end)
	t.Log(start, end)

	isEqual := DateEqualToday(Now())
	assert.Equal(t, isEqual, true)
	t.Log(isEqual)

	zeroDatetime := ToZeroTime(Now())
	t.Log(zeroDatetime)

	dates, _ := GenerateNextDate(Now().AddDate(0, 0, -10), Now())
	t.Log(dates)
}
