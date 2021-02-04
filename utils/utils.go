package utils

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
