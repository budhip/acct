package atime

// DateLayout
const (
	DateFormatYYYYMMDD                     = "2006-01-02"
	DateFormatDDMMYYYYWithSlash            = "02/01/2006"
	DateFormatYYYYMM                       = "2006-01"
	DateFormatYYYYMMWithUnderscore         = "2006_01"
	DateFormatYYYY                         = "2006"
	DateFormatMM                           = "01"
	DateFormatYYYYMMDDWithoutDash          = "20060102"
	DateFormatDDMMMYYYY                    = "02-Jan-2006"
	DateFormatYYYYMMDDWithTime             = "2006-01-02 15:04:05"
	DateFormatYYYYMMDDWithTimeWithoutColon = "2006-01-02_150405"
	DateFormatYYYYMMDDWithTimeWithoutDash  = "20060102150405"
	DateFormatDDMMMMYYYYWithTime           = "02-January-2006/15:04:05"
	DateFormatDDMMMMYYYYWithSpace          = "02 January 2006"
	DateFormatDDMMMYYYYWithSpace           = "02 Jan 2006"
	TimeFormatHHMM                         = "15:04"
	DateFormatHHMMSS                       = "15:04:05"
	DateFormatDDMMMYYYYTimeWithSpace       = "02 Jan 2006 15:04:05"
	DateFormatRFC3339NanoWithTimeZone      = "2006-01-02T15:04:05.999999-07:00"
)

// HOUR FORMAT
const (
	HourFormat000000 = "00:00:00"
	HourFormat235959 = "23:59:59"
)

// TIMEZONE
const (
	TimezoneJakarta = "Asia/Jakarta"
)

// MAP TIMEZONE
var (
	MapTimezone = map[string]int{
		TimezoneJakarta: 7,
	}
)
