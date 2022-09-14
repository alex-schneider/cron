// Copyright 2021 Alex Schneider. All rights reserved.

package cron

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var dowMap = map[string]int{
	"SUN": 0, "MON": 1, "TUE": 2, "WED": 3, "THU": 4, "FRI": 5, "SAT": 6,
}

var monthMap = map[string]int{
	"JAN": 1, "FEB": 2, "MAR": 3, "APR": 4, "MAY": 5, "JUN": 6,
	"JUL": 7, "AUG": 8, "SEP": 9, "OCT": 10, "NOV": 11, "DEC": 12,
}

var tableValues = []int{
	0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
	10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
	20, 21, 22, 23, 24, 25, 26, 27, 28, 29,
	30, 31, 32, 33, 34, 35, 36, 37, 38, 39,
	40, 41, 42, 43, 44, 45, 46, 47, 48, 49,
	50, 51, 52, 53, 54, 55, 56, 57, 58, 59,
}

var yearValues = []int{
	1970, 1971, 1972, 1973, 1974, 1975, 1976, 1977, 1978, 1979,
	1980, 1981, 1982, 1983, 1984, 1985, 1986, 1987, 1988, 1989,
	1990, 1991, 1992, 1993, 1994, 1995, 1996, 1997, 1998, 1999,
	2000, 2001, 2002, 2003, 2004, 2005, 2006, 2007, 2008, 2009,
	2010, 2011, 2012, 2013, 2014, 2015, 2016, 2017, 2018, 2019,
	2020, 2021, 2022, 2023, 2024, 2025, 2026, 2027, 2028, 2029,
	2030, 2031, 2032, 2033, 2034, 2035, 2036, 2037, 2038, 2039,
	2040, 2041, 2042, 2043, 2044, 2045, 2046, 2047, 2048, 2049,
	2050, 2051, 2052, 2053, 2054, 2055, 2056, 2057, 2058, 2059,
	2060, 2061, 2062, 2063, 2064, 2065, 2066, 2067, 2068, 2069,
	2070, 2071, 2072, 2073, 2074, 2075, 2076, 2077, 2078, 2079,
	2080, 2081, 2082, 2083, 2084, 2085, 2086, 2087, 2088, 2089,
	2090, 2091, 2092, 2093, 2094, 2095, 2096, 2097, 2098, 2099,
}

var (
	startupTime                = time.Now()
	dowList                    = "SUN|MON|TUE|WED|THU|FRI|SAT|SUN"
	monthList                  = "JAN|FEB|MAR|APR|MAY|JUN|JUL|AUG|SEP|OCT|NOV|DEC"
	listRegex                  = `(\d+|` + dowList + `|` + monthList + `)`
	reWeekdayDoM               = regexp.MustCompile(`^(0?[1-9]|[12][0-9]|3[01])W$`)
	reLastDoWInMonth           = regexp.MustCompile(`^([0-7]|` + dowList + `)L$`)
	reDoWInSpecificWeek        = regexp.MustCompile(`^([0-7]|` + dowList + `)#([1-5])$`)
	reSingleValue              = regexp.MustCompile(`^` + listRegex + `$`)
	reRangeValue               = regexp.MustCompile(`^` + listRegex + `-` + listRegex + `$`)
	reWildcardIntervalValue    = regexp.MustCompile(`^\*/(\d+)$`)
	reSingleValueIntervalValue = regexp.MustCompile(`^` + listRegex + `/(\d+)$`)
	reRangeIntervalValue       = regexp.MustCompile(`^` + listRegex + `-` + listRegex + `/(\d+)$`)
)

/* ==================================================================================================== */

func getFlexValues(expr string, ft fieldType) ([]int, string, error) {
	if err := getFlexValuesError(expr, ft); err != nil {
		return nil, "", err
	}

	// `R`
	if expr == "R" {
		switch ft {
		case typeSeconds:
			fallthrough
		case typeMinutes:
			return []int{random.Intn(60)}, "", nil // Random 0-59
		case typeHours:
			return []int{random.Intn(24)}, "", nil // Random 0-23
		case typeDoM:
			return []int{tableValues[1:29][random.Intn(28)]}, "", nil // Random 1-28
		case typeMonth:
			return []int{tableValues[1:13][random.Intn(12)]}, "", nil // Random 1-12
		case typeDoW:
			return []int{random.Intn(6)}, "", nil // Random 0-6
		case typeYear:
			return []int{yearValues[random.Intn(130)]}, "", nil // Random 1970-2099
		}
	}

	// `L`
	if expr == "L" {
		if ft == typeDoW {
			return []int{6}, "", nil // Last day of week (SAT)
		}

		return nil, expr, nil // Last day of month
	}

	// `LW`
	if expr == "LW" {
		return nil, expr, nil // Last weekday (MON-FRI) of month
	}

	// `?`
	if expr == "?" {
		return nil, expr, nil // Empty value
	}

	// `15W` (nearest weekday (MON-FRI) of the month to the 15.)
	if matches := reWeekdayDoM.FindStringSubmatch(expr); len(matches) == 2 {
		v, _ := strconv.Atoi(matches[1])

		return []int{v}, "W", nil
	}

	// `5L` (last FRI of the month)
	if matches := reLastDoWInMonth.FindStringSubmatch(expr); len(matches) == 2 {
		if ft != typeDoW {
			return nil, "", fmt.Errorf("the '{x}L' is only allowed in the DoW field")
		}

		_, _, _, v, _ := toNumVal(matches[1])

		return []int{v}, "L", nil
	}

	// `5#3` (nth (1-5, here 3) DoW (0-7, here FRI) of the month)
	if matches := reDoWInSpecificWeek.FindStringSubmatch(expr); len(matches) == 3 {
		_, _, _, v1, _ := toNumVal(matches[1])

		v2, _ := strconv.Atoi(matches[2])

		return []int{v1, v2}, "#", nil
	}

	// Code cannot be reached in the production code...
	return nil, "", fmt.Errorf("unsupported expression value given: '%s'", expr)
}

func getFlexValuesError(expr string, ft fieldType) error {
	if expr == "R" {
		return nil
	}

	if ft != typeDoM && ft != typeDoW {
		return fmt.Errorf("the special characters 'L', 'W', '?' and '#' are only allowed in the DoM and DoW fields")
	} else if strings.Contains(expr, "W") && ft != typeDoM {
		return fmt.Errorf("the special character 'W' is only allowed in the DoM field")
	} else if strings.Contains(expr, "#") && ft != typeDoW {
		return fmt.Errorf("the special character '#' is only allowed in the DoW field")
	}

	return nil
}

func getFixValues(expr string, ft fieldType) ([]int, error) {
	// `*`
	if expr == "*" {
		return getWildcardValues(ft)
	}

	// `.`
	if expr == "." {
		return getCurrentTimeValues(ft)
	}

	// `3`, `NOV`, `FRI`, `2020`
	if match := reSingleValue.FindString(expr); match != "" {
		return getSingleValue(match, ft)
	}

	// `3-5`, `APR-JUL`, `2020-2035`, `MON-5` (eq. `MON-FRI`)
	if matches := reRangeValue.FindStringSubmatch(expr); len(matches) == 3 {
		return getRangeValues(matches[1], matches[2], ft)
	}

	// `*/5`
	if matches := reWildcardIntervalValue.FindStringSubmatch(expr); len(matches) == 2 {
		return getWildcardIntervalValues(matches[1], ft)
	}

	// `10/5`, `MON/2`, `JAN/4`
	if matches := reSingleValueIntervalValue.FindStringSubmatch(expr); len(matches) == 3 {
		return getSingleValueIntervalValues(matches[1], matches[2], ft)
	}

	// `10-30/5`, `APR-JUL/2`, `2020-2035/5`, `MON-6/2` (eq. `MON-SAT/2`)
	if matches := reRangeIntervalValue.FindStringSubmatch(expr); len(matches) == 4 {
		return getRangeIntervalValues(matches[1], matches[2], matches[3], ft)
	}

	return errValue(expr, ft)
}

func errValue(value string, ft fieldType) ([]int, error) {
	return nil, fmt.Errorf("invalid value in field '%s' given: '%s'", ft, value)
}

/* ==================================================================================================== */

func getWildcardValues(ft fieldType) ([]int, error) {
	var values []int

	switch ft {
	case typeSeconds, typeMinutes:
		values = make([]int, 60)
		copy(values, tableValues)
	case typeHours:
		values = make([]int, 24)
		copy(values, tableValues[:24])
	case typeDoM:
		values = make([]int, 31)
		copy(values, tableValues[1:32])
	case typeMonth:
		values = make([]int, 12)
		copy(values, tableValues[1:13])
	case typeDoW:
		values = make([]int, 7)
		copy(values, tableValues[:7])
	case typeYear:
		values = make([]int, 130)
		copy(values, yearValues)
	default:
		return nil, fmt.Errorf("unsupported fieldType given: '%s'", ft)
	}

	return values, nil
}

func getCurrentTimeValues(ft fieldType) ([]int, error) {
	switch ft {
	case typeSeconds:
		return []int{startupTime.Second()}, nil
	case typeMinutes:
		return []int{startupTime.Minute()}, nil
	case typeHours:
		return []int{startupTime.Hour()}, nil
	case typeDoM:
		return []int{startupTime.Day()}, nil
	case typeMonth:
		return []int{int(startupTime.Month())}, nil
	case typeDoW:
		return []int{int(startupTime.Weekday())}, nil
	case typeYear:
		return []int{startupTime.Year()}, nil
	}

	// Code cannot be reached in the production code...
	return nil, fmt.Errorf("unsupported fieldType given: '%s'", ft)
}

func getSingleValue(value string, ft fieldType) ([]int, error) {
	isDoW, isMonth, isNum, numVal, err := toNumVal(value)
	if err != nil {
		return nil, err
	}

	switch ft {
	case typeSeconds, typeMinutes, typeHours:
		return getSingleValueFromTime(value, ft, isNum, numVal)
	case typeDoM, typeDoW:
		return getSingleValueFromDayOf(value, ft, isDoW, isNum, numVal)
	case typeMonth, typeYear:
		return getSingleValueFromDate(value, ft, isMonth, isNum, numVal)
	}

	// Code cannot be reached in the production code...
	return nil, fmt.Errorf("unsupported fieldType given: '%s'", ft)
}

func getSingleValueFromTime(
	value string, ft fieldType, isNum bool, numVal int,
) ([]int, error) {
	switch ft {
	case typeSeconds, typeMinutes:
		if !isNum || numVal < 0 || numVal > 59 {
			return errValue(value, ft)
		}

		return []int{numVal}, nil
	case typeHours:
		if !isNum || numVal < 0 || numVal > 23 {
			return errValue(value, ft)
		}

		return []int{numVal}, nil
	}

	// Code cannot be reached in the production code...
	return nil, fmt.Errorf("unsupported fieldType given: '%s'", ft)
}

func getSingleValueFromDayOf(
	value string, ft fieldType, isDoW, isNum bool, numVal int,
) ([]int, error) {
	switch ft {
	case typeDoM:
		if !isNum || numVal < 1 || numVal > 31 {
			return errValue(value, ft)
		}

		return []int{numVal}, nil
	case typeDoW:
		if !isNum && !isDoW || numVal < 0 || numVal > 7 {
			return errValue(value, ft)
		} else if numVal == 7 {
			return []int{0}, nil
		}

		return []int{numVal}, nil
	}

	// Code cannot be reached in the production code...
	return nil, fmt.Errorf("unsupported fieldType given: '%s'", ft)
}

func getSingleValueFromDate(
	value string, ft fieldType, isMonth, isNum bool, numVal int,
) ([]int, error) {
	switch ft {
	case typeMonth:
		if !isNum && !isMonth || numVal < 1 || numVal > 12 {
			return errValue(value, ft)
		}

		return []int{numVal}, nil
	case typeYear:
		if !isNum || numVal < 1970 || numVal > 2099 {
			return errValue(value, ft)
		}

		return []int{numVal}, nil
	}

	// Code cannot be reached in the production code...
	return nil, fmt.Errorf("unsupported fieldType given: '%s'", ft)
}

func getRangeValues(v1, v2 string, ft fieldType) ([]int, error) {
	numVal1, err := getSingleValue(v1, ft)
	if err != nil {
		return nil, err
	}

	numVal2, err := getSingleValue(v2, ft)
	if err != nil {
		return nil, err
	}

	if numVal1[0] > numVal2[0] && ft == typeYear {
		return errValue(v1+"-"+v2, ft)
	}

	var values []int

	if numVal1[0] > numVal2[0] {
		min, max := getMinMax(ft)

		for i := min; i <= numVal2[0]; i++ {
			values = append(values, i)
		}
		for i := numVal1[0]; i <= max; i++ {
			values = append(values, i)
		}
	} else {
		for i := numVal1[0]; i <= numVal2[0]; i++ {
			values = append(values, i)
		}
	}

	return values, nil
}

func getWildcardIntervalValues(interval string, ft fieldType) ([]int, error) {
	step, err := strconv.Atoi(interval)
	if err != nil {
		return nil, err
	}

	var values []int
	min, max := getMinMax(ft)

	for i := min; i <= max; i += step {
		values = append(values, i)
	}

	return values, nil
}

func getSingleValueIntervalValues(value, interval string, ft fieldType) ([]int, error) {
	step, err := strconv.Atoi(interval)
	if err != nil {
		return nil, err
	}

	nums, err := getSingleValue(value, ft)
	if err != nil {
		return nil, err
	}

	var values []int
	_, max := getMinMax(ft)

	for i := nums[0]; i <= max; i += step {
		values = append(values, i)
	}

	return values, nil
}

func getRangeIntervalValues(v1, v2, interval string, ft fieldType) ([]int, error) {
	step, err := strconv.Atoi(interval)
	if err != nil {
		return nil, err
	}

	numVal1, err := getSingleValue(v1, ft)
	if err != nil {
		return nil, err
	}

	numVal2, err := getSingleValue(v2, ft)
	if err != nil {
		return nil, err
	}

	if numVal1[0] > numVal2[0] && ft == typeYear {
		return errValue(v1+"-"+v2, ft)
	}

	var values []int

	if numVal1[0] > numVal2[0] {
		list, _ := getWildcardValues(ft)

		var i int
		var running bool

		for _, v := range append(list, list...) {
			if v == numVal1[0] {
				running = true
			}

			if !running {
				continue
			}

			if i%step == 0 {
				values = append(values, v)
			}

			i++

			if v < numVal1[0] && v == numVal2[0] {
				break
			}
		}

		sort.Ints(values)
	} else {
		for i := numVal1[0]; i <= numVal2[0]; i += step {
			values = append(values, i)
		}
	}

	return values, nil
}

/* ==================================================================================================== */

func getMinMax(ft fieldType) (int, int) {
	var min, max int

	switch ft {
	case typeSeconds, typeMinutes:
		min = 0
		max = 59
	case typeHours:
		min = 0
		max = 23
	case typeDoM:
		min = 1
		max = 31
	case typeMonth:
		min = 1
		max = 12
	case typeDoW:
		min = 0
		max = 6
	case typeYear:
		min = 1970
		max = 2099
	}

	// Code cannot be reached in the production code...
	return min, max
}

func isFlexValue(expr string) bool {
	var flex bool

	flex = flex || expr == "R"
	flex = flex || expr == "L"
	flex = flex || expr == "LW"
	flex = flex || expr == "?"
	flex = flex || reWeekdayDoM.MatchString(expr)
	flex = flex || reLastDoWInMonth.MatchString(expr)
	flex = flex || reDoWInSpecificWeek.MatchString(expr)

	return flex
}

func toNumVal(value string) (bool, bool, bool, int, error) {
	var isDoW, isMonth, isNum bool
	var numVal int

	if n, ok := dowMap[value]; ok {
		isDoW = true
		numVal = n
	} else if n, ok := monthMap[value]; ok {
		isMonth = true
		numVal = n
	} else {
		n, err := strconv.Atoi(value)
		if err != nil {
			return false, false, false, 0, err
		}

		isNum = true
		numVal = n
	}

	return isDoW, isMonth, isNum, numVal, nil
}

func uniqueValues(values []int) []int {
	var unique []int

	for _, i := range values {
		var exists bool

		for _, j := range unique {
			if i == j {
				exists = true
				break
			}
		}

		if !exists {
			unique = append(unique, i)
		}
	}

	sort.Ints(unique)

	return unique
}
