package gofunc

import (
	"strconv"
	"time"
)

var (
	timeZone, _ = time.LoadLocation("Asia/Shanghai")
	DATE_LAYOUT = "2006-01-02 15:04:05"
	weekDay     = map[time.Weekday]int{
		time.Sunday:    0,
		time.Monday:    1,
		time.Tuesday:   2,
		time.Wednesday: 3,
		time.Thursday:  4,
		time.Friday:    5,
		time.Saturday:  6,
	}
)

func SetTimeZone(tz string) {

	var err error
	timeZone, err = time.LoadLocation(tz)
	if err != nil {

		panic(err.Error())
	}
}

func TimeNow() time.Time {

	return time.Now().In(timeZone)
}

func TimeUnix(sec int64, nsec int64) time.Time {

	return time.Unix(sec, nsec).In(timeZone)
}

func TimeNowUnix() int64 {

	return time.Now().In(timeZone).Unix()
}

// Date()
// Date("2006-01-02 15:04:05")
// Date("2006-01-02 15:04:05", 1431662400)
func Date(params ...interface{}) (s string) {

	var layout = DATE_LAYOUT
	if len(params) == 1 {

		switch params[0].(type) {

		case string:

			layout = params[0].(string)

		default:

			return
		}

		s = TimeNow().Format(layout)
	} else if len(params) == 2 {

		switch params[0].(type) {

		case string:

			layout = params[0].(string)

		default:

			return
		}

		var mtime int64

		switch params[1].(type) {

		case int64:

			mtime = params[1].(int64)

		default:

			return
		}

		s = TimeUnix(mtime, 0).Format(layout)
	} else {

		s = TimeNow().Format(layout)
	}

	return
}

// StrToTime("2015-05-15 12:00:00")
// StrToTime("2015-05-15", "2006-01-02")
func StrToTime(date string, params ...string) (int64, error) {

	layout := DATE_LAYOUT
	if len(params) == 1 {

		layout = params[0]
	}

	t, err := time.ParseInLocation(layout, date, timeZone)
	if err != nil {

		return 0, err
	}

	return t.Unix(), nil
}

func YmdHis(t ...int64) (ymd int, his int) {

	if len(t) == 0 {

		d := time.Now().In(timeZone).Format("20060102150405")
		ymd, _ = strconv.Atoi(SubStr(d, 0, 8))
		his, _ = strconv.Atoi(SubStr(d, 8))
	} else {

		d := time.Unix(t[0], 0).In(timeZone).Format("20060102150405")
		ymd, _ = strconv.Atoi(SubStr(d, 0, 8))
		his, _ = strconv.Atoi(SubStr(d, 8))
	}

	return
}

func WeekDay() int {

	wd := time.Now().In(timeZone).Weekday()

	return weekDay[wd]
}

func YearWeek(t ...int64) int {

	var year, week int

	if len(t) == 0 {

		year, week = time.Now().In(timeZone).ISOWeek()
	} else {

		year, week = time.Unix(t[0], 0).In(timeZone).ISOWeek()
	}
	return year*100 + week
}

func GetYMD(t ...int64) (year int, mon int, day int) {

	if len(t) == 0 {

		d := time.Now().In(timeZone).Format("20060102")
		year, _ = strconv.Atoi(SubStr(d, 0, 4))
		mon, _ = strconv.Atoi(SubStr(d, 4, 2))
		day, _ = strconv.Atoi(SubStr(d, 6))
	} else {

		d := time.Unix(t[0], 0).In(timeZone).Format("20060102")
		year, _ = strconv.Atoi(SubStr(d, 0, 4))
		mon, _ = strconv.Atoi(SubStr(d, 4, 2))
		day, _ = strconv.Atoi(SubStr(d, 6))
	}

	return
}
