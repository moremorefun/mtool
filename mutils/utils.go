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
