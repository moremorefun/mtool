package mutils

import (
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

// IsStringInSlice 字符串是否在数组中
func IsStringInSlice(arr []string, str string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}
	return false
}

// IsIntInSlice 数字是否在数组中
func IsIntInSlice(arr []int64, str int64) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}
	return false
}

// IsInSlice 数字是否在数组中
func IsInSlice(arr []interface{}, iv interface{}) bool {
	for _, v := range arr {
		if v == iv {
			return true
		}
	}
	return false
}

// GetUUIDStr 获取唯一字符串
func GetUUIDStr() string {
	u1 := uuid.NewV4()
	return strings.Replace(u1.String(), "-", "", -1)
}

// GetLocalDayStart 获取当天开始时间
func GetLocalDayStart() (time.Time, error) {
	cstSh := time.FixedZone("CST", 8*3600)
	now := time.Now().In(cstSh)
	nowStr := now.Format("2006-01-02")
	localDayStart, err := time.ParseInLocation("2006-01-02", nowStr, cstSh)
	if err != nil {
		return time.Time{}, err
	}
	return localDayStart, nil
}

// GetLocalWeekStart 获取周天开始时间
func GetLocalWeekStart() (time.Time, error) {
	// 计算起始时间
	cstSh := time.FixedZone("CST", 8*3600)
	now := time.Now().In(cstSh)
	nowStr := now.Format("2006-01-02")
	localDayStart, err := time.ParseInLocation("2006-01-02", nowStr, cstSh)
	if err != nil {
		return time.Time{}, err
	}
	weekDay := localDayStart.Weekday()
	if weekDay == 0 {
		weekDay = 6
	} else {
		weekDay -= 1
	}
	dailyStart := localDayStart.AddDate(0, 0, -int(weekDay))
	return dailyStart, nil
}

// GetLocalDayStr 获取当天日期字符串
func GetLocalDayStr() string {
	cstSh := time.FixedZone("CST", 8*3600)
	now := time.Now().In(cstSh)
	nowStr := now.Format("2006-01-02")
	return nowStr
}

// GetLocalDayStrByStartHour 获取当天日期字符串
func GetLocalDayStrByStartHour(hour int) (string, error) {
	todayStr := GetLocalDayStr()
	cstSh := time.FixedZone("CST", 8*3600)
	now := time.Now().In(cstSh)

	dataStart, err := GetLocalDayStart()
	if err != nil {
		return "", err
	}
	if now.Hour() < hour {
		// 计入上一天
		todayStr = dataStart.AddDate(0, 0, -1).Format("2006-01-02")
	}
	return todayStr, nil
}

// GetLocalWeekStrByStartHour 获取当周日期字符串
func GetLocalWeekStrByStartHour(hour int) (string, error) {
	cstSh := time.FixedZone("CST", 8*3600)
	now := time.Now().In(cstSh)
	nowStr := now.Format("2006-01-02")

	weekStartTime, err := GetLocalWeekStart()
	if err != nil {
		return "", err
	}
	weekStr := weekStartTime.Format("2006-01-02")
	if now.Hour() < hour && nowStr == weekStr {
		// 计入上一周
		weekStartTime = weekStartTime.AddDate(0, 0, -7)
	}
	return weekStartTime.Format("2006-01-02"), nil
}

// GetLocalDayByStartHour 获取当天日期开始时间
func GetLocalDayByStartHour(hour int) (time.Time, error) {
	dayStart, err := GetLocalDayStart()
	if err != nil {
		return time.Time{}, err
	}
	dayStart = dayStart.Add(time.Hour * time.Duration(hour))
	cstSh := time.FixedZone("CST", 8*3600)
	now := time.Now().In(cstSh)

	if now.Hour() < hour {
		// 计入上一天
		return dayStart.AddDate(0, 0, -1), nil

	}
	return dayStart, nil
}

// GetLocalWeekByStartHour 获取当周日期开始时间
func GetLocalWeekByStartHour(hour int) (time.Time, error) {
	cstSh := time.FixedZone("CST", 8*3600)
	now := time.Now().In(cstSh)
	nowStr := now.Format("2006-01-02")

	weekStartTime, err := GetLocalWeekStart()
	if err != nil {
		return time.Time{}, err
	}
	weekStr := weekStartTime.Format("2006-01-02")
	if now.Hour() < hour && nowStr == weekStr {
		// 计入上一周
		weekStartTime = weekStartTime.AddDate(0, 0, -7)
	}
	return weekStartTime.Add(time.Hour * time.Duration(hour)), nil
}

// GetUnixLocalWeekStart 获取周天开始时间
func GetUnixLocalWeekStart(unixAt int64) (time.Time, error) {
	// 计算起始时间
	cstSh := time.FixedZone("CST", 8*3600)
	now := time.Unix(unixAt, 0).In(cstSh)
	nowStr := now.Format("2006-01-02")
	localDayStart, err := time.ParseInLocation("2006-01-02", nowStr, cstSh)
	if err != nil {
		return time.Time{}, err
	}
	weekDay := localDayStart.Weekday()
	if weekDay == 0 {
		weekDay = 6
	} else {
		weekDay -= 1
	}
	dailyStart := localDayStart.AddDate(0, 0, -int(weekDay))
	return dailyStart, nil
}

// GetUnixOfLocalWeekStrByStartHour 获取时间点所在的当周日期字符串
func GetUnixOfLocalWeekStrByStartHour(unixAt int64, hour int) (string, error) {
	cstSh := time.FixedZone("CST", 8*3600)
	now := time.Unix(unixAt, 0).In(cstSh)
	nowStr := now.Format("2006-01-02")

	weekStartTime, err := GetUnixLocalWeekStart(unixAt)
	if err != nil {
		return "", err
	}
	weekStr := weekStartTime.Format("2006-01-02")
	if now.Hour() < hour && nowStr == weekStr {
		// 计入上一周
		weekStartTime = weekStartTime.AddDate(0, 0, -7)
	}
	return weekStartTime.Format("2006-01-02"), nil
}
