package cron

import (
	"fmt"
	"reflect"
	"testing"
)

func TestValues_getFlexValues(t *testing.T) {
	type testCase struct {
		expr string
		ft   fieldType
		expV []int
		expU string
		err  string
	}

	for _, tc := range []testCase{
		{"?", typeSeconds, nil, "", `the special characters 'L', 'W', '?' and '#' are only allowed in the DoM and DoW fields`},
		{"LW", typeDoW, nil, "", `the special character 'W' is only allowed in the DoM field`},
		{"5#3", typeDoM, nil, "", `the special character '#' is only allowed in the DoW field`},

		{"L", typeDoM, nil, "L", ``},
		{"L", typeDoW, []int{6}, "", ``},
		{"LW", typeDoM, nil, "LW", ``},
		{"?", typeDoM, nil, "?", ``},
		{"15W", typeDoM, []int{15}, "W", ``},
		{"01W", typeDoM, []int{1}, "W", ``},
		{"1W", typeDoM, []int{1}, "W", ``},
		{"5L", typeDoM, nil, "", `the '{x}L' is only allowed in the DoW field`},
		{"5L", typeDoW, []int{5}, "L", ``},
		{"FRI#3", typeDoW, []int{5, 3}, "#", ``},
		{"XXX", typeDoM, nil, "", `unsupported expression value given: 'XXX'`},
	} {
		got, unit, err := getFlexValues(tc.expr, tc.ft)
		if !reflect.DeepEqual(tc.expV, got) {
			t.Errorf("'%s': expected '%#v', got '%#v'", tc.expr, tc.expV, got)
		}

		if tc.expU != unit {
			t.Errorf("'%s': expected '%#v', got '%#v'", tc.expr, tc.expU, unit)
		}

		if tc.err != "" || err != nil {
			if eerr := fmt.Sprintf("%s", err); tc.err != eerr {
				t.Errorf("'%s': expected '%s', got '%s'", tc.expr, tc.err, eerr)
			}
		}
	}
}

func TestValues_getFixValues(t *testing.T) {
	type testCase struct {
		expr string
		ft   fieldType
		exp  []int
		err  string
	}

	for _, tc := range []testCase{
		{"*", typeSeconds, tableValues, ``},
		{".", typeSeconds, []int{startupTime.Second()}, ``},
		{"NOV", typeMonth, []int{11}, ``},
		{"MON-3", typeDoW, []int{1, 2, 3}, ``},
		{"*/6", typeHours, []int{0, 6, 12, 18}, ``},
		{"MON/2", typeDoW, []int{1, 3, 5}, ``},
		{"10-20/5", typeMinutes, []int{10, 15, 20}, ``},
		{"XXX", typeYear, nil, `invalid value in field 'year' given: 'XXX'`},
	} {
		got, err := getFixValues(tc.expr, tc.ft)
		if !reflect.DeepEqual(tc.exp, got) {
			t.Errorf("'%s': expected '%#v', got '%#v'", tc.expr, tc.exp, got)
		}

		if tc.err != "" || err != nil {
			if eerr := fmt.Sprintf("%s", err); tc.err != eerr {
				t.Errorf("'%s': expected '%s', got '%s'", tc.expr, tc.err, eerr)
			}
		}
	}
}

/* ==================================================================================================== */

func TestValues_getWildcardValues(t *testing.T) {
	type testCase struct {
		ft  fieldType
		exp []int
		err string
	}

	for _, tc := range []testCase{
		{typeSeconds, tableValues[:], ``},
		{typeMinutes, tableValues[:], ``},
		{typeHours, tableValues[:24], ``},
		{typeDoM, tableValues[1:32], ``},
		{typeMonth, tableValues[1:13], ``},
		{typeDoW, tableValues[:7], ``},
		{typeYear, yearValues[:], ``},
		{-1, nil, `unsupported fieldType given: 'unknown'`},
	} {
		if got, err := getWildcardValues(tc.ft); !reflect.DeepEqual(tc.exp, got) {
			t.Errorf("'%s': expected '%#v', got '%#v'", tc.ft, tc.exp, got)
		} else if tc.err != "" || err != nil {
			if eerr := fmt.Sprintf("%s", err); tc.err != eerr {
				t.Errorf("'%s': expected '%s', got '%s'", tc.ft, tc.err, eerr)
			}
		}
	}
}

func TestValues_getCurrentTimeValues(t *testing.T) {
	type testCase struct {
		ft  fieldType
		exp []int
		err string
	}

	for _, tc := range []testCase{
		{typeSeconds, []int{startupTime.Second()}, ``},
		{typeMinutes, []int{startupTime.Minute()}, ``},
		{typeHours, []int{startupTime.Hour()}, ``},
		{typeDoM, []int{startupTime.Day()}, ``},
		{typeMonth, []int{int(startupTime.Month())}, ``},
		{typeDoW, []int{int(startupTime.Weekday())}, ``},
		{typeYear, []int{startupTime.Year()}, ``},
		{-1, nil, `unsupported fieldType given: 'unknown'`},
	} {
		if got, err := getCurrentTimeValues(tc.ft); !reflect.DeepEqual(tc.exp, got) {
			t.Errorf("'%s': expected '%#v', got '%#v'", tc.ft, tc.exp, got)
		} else if err != nil {
			if eerr := fmt.Sprintf("%s", err); tc.err != eerr {
				t.Errorf("'%s': expected '%s', got '%s'", tc.ft, tc.err, eerr)
			}
		}
	}
}

func TestValues_getSingleValue(t *testing.T) {
	type testCase struct {
		val string
		ft  fieldType
		exp []int
		err string
	}

	for _, tc := range []testCase{
		// Global
		{"XXX", typeSeconds, nil, `strconv.Atoi: parsing "XXX": invalid syntax`},
		{"5", -1, nil, `unsupported fieldType given: 'unknown'`},

		// Seconds
		{"MON", typeSeconds, nil, `invalid value in field 'seconds' given: 'MON'`},
		{"-1", typeSeconds, nil, `invalid value in field 'seconds' given: '-1'`},
		{"60", typeSeconds, nil, `invalid value in field 'seconds' given: '60'`},
		{"30", typeSeconds, []int{30}, ``},

		// Minutes
		{"MON", typeMinutes, nil, `invalid value in field 'minutes' given: 'MON'`},
		{"-1", typeMinutes, nil, `invalid value in field 'minutes' given: '-1'`},
		{"60", typeMinutes, nil, `invalid value in field 'minutes' given: '60'`},
		{"30", typeMinutes, []int{30}, ``},

		// Hours
		{"MON", typeHours, nil, `invalid value in field 'hours' given: 'MON'`},
		{"-1", typeHours, nil, `invalid value in field 'hours' given: '-1'`},
		{"24", typeHours, nil, `invalid value in field 'hours' given: '24'`},
		{"12", typeHours, []int{12}, ``},

		// DoM
		{"MON", typeDoM, nil, `invalid value in field 'day-of-month' given: 'MON'`},
		{"0", typeDoM, nil, `invalid value in field 'day-of-month' given: '0'`},
		{"32", typeDoM, nil, `invalid value in field 'day-of-month' given: '32'`},
		{"15", typeDoM, []int{15}, ``},

		// Month
		{"MON", typeMonth, nil, `invalid value in field 'month' given: 'MON'`},
		{"0", typeMonth, nil, `invalid value in field 'month' given: '0'`},
		{"13", typeMonth, nil, `invalid value in field 'month' given: '13'`},
		{"11", typeMonth, []int{11}, ``},

		// DoW
		{"JAN", typeDoW, nil, `invalid value in field 'day-of-week' given: 'JAN'`},
		{"-1", typeDoW, nil, `invalid value in field 'day-of-week' given: '-1'`},
		{"8", typeDoW, nil, `invalid value in field 'day-of-week' given: '8'`},
		{"5", typeDoW, []int{5}, ``},
		{"7", typeDoW, []int{0}, ``},

		// Year
		{"MON", typeYear, nil, `invalid value in field 'year' given: 'MON'`},
		{"1969", typeYear, nil, `invalid value in field 'year' given: '1969'`},
		{"2100", typeYear, nil, `invalid value in field 'year' given: '2100'`},
		{"2020", typeYear, []int{2020}, ``},
	} {
		got, err := getSingleValue(tc.val, tc.ft)
		if !reflect.DeepEqual(tc.exp, got) {
			t.Errorf("'%s': expected '%#v', got '%#v'", tc.val, tc.exp, got)
		}

		if tc.err != "" || err != nil {
			if eerr := fmt.Sprintf("%s", err); tc.err != eerr {
				t.Errorf("'%s': expected '%s', got '%s'", tc.val, tc.err, eerr)
			}
		}
	}
}

func TestValues_getSingleValueFromTime(t *testing.T) {
	_, err := getSingleValueFromTime("XXX", typeMonth, true, 0)

	exp := `unsupported fieldType given: 'month'`
	if eerr := fmt.Sprintf("%s", err); eerr != exp {
		t.Errorf("expected '%s', got '%s'", exp, err)
	}
}

func TestValues_getSingleValueFromDayOf(t *testing.T) {
	_, err := getSingleValueFromDayOf("XXX", typeMonth, true, true, 0)

	exp := `unsupported fieldType given: 'month'`
	if eerr := fmt.Sprintf("%s", err); eerr != exp {
		t.Errorf("expected '%s', got '%s'", exp, err)
	}
}

func TestValues_getSingleValueFromDate(t *testing.T) {
	_, err := getSingleValueFromDate("XXX", typeHours, true, true, 0)

	exp := `unsupported fieldType given: 'hours'`
	if eerr := fmt.Sprintf("%s", err); eerr != exp {
		t.Errorf("expected '%s', got '%s'", exp, err)
	}
}

func TestValues_getRangeValues(t *testing.T) {
	type testCase struct {
		v1  string
		v2  string
		ft  fieldType
		exp []int
		err string
	}

	for _, tc := range []testCase{
		// Global
		{"XXX", "XXX", typeSeconds, nil, `strconv.Atoi: parsing "XXX": invalid syntax`},
		{"5", "XXX", typeSeconds, nil, `strconv.Atoi: parsing "XXX": invalid syntax`},
		{"5", "6", -1, nil, `unsupported fieldType given: 'unknown'`},

		// Seconds
		{"30", "35", typeSeconds, []int{30, 31, 32, 33, 34, 35}, ``},
		{"58", "2", typeSeconds, []int{0, 1, 2, 58, 59}, ``},
		{"5", "5", typeSeconds, []int{5}, ``},

		// Minutes
		{"30", "35", typeMinutes, []int{30, 31, 32, 33, 34, 35}, ``},
		{"58", "2", typeMinutes, []int{0, 1, 2, 58, 59}, ``},
		{"5", "5", typeMinutes, []int{5}, ``},

		// Hours
		{"12", "15", typeHours, []int{12, 13, 14, 15}, ``},
		{"22", "2", typeHours, []int{0, 1, 2, 22, 23}, ``},
		{"5", "5", typeHours, []int{5}, ``},

		// DoM
		{"12", "15", typeDoM, []int{12, 13, 14, 15}, ``},
		{"30", "2", typeDoM, []int{1, 2, 30, 31}, ``},
		{"5", "5", typeDoM, []int{5}, ``},

		// Month
		{"3", "6", typeMonth, []int{3, 4, 5, 6}, ``},
		{"MAR", "6", typeMonth, []int{3, 4, 5, 6}, ``},
		{"3", "JUN", typeMonth, []int{3, 4, 5, 6}, ``},
		{"11", "2", typeMonth, []int{1, 2, 11, 12}, ``},
		{"5", "5", typeMonth, []int{5}, ``},

		// DoW
		{"3", "6", typeDoW, []int{3, 4, 5, 6}, ``},
		{"WED", "6", typeDoW, []int{3, 4, 5, 6}, ``},
		{"3", "SAT", typeDoW, []int{3, 4, 5, 6}, ``},
		{"7", "2", typeDoW, []int{0, 1, 2}, ``},
		{"5", "5", typeDoW, []int{5}, ``},

		// Year
		{"2020", "2025", typeYear, []int{2020, 2021, 2022, 2023, 2024, 2025}, ``},
		{"2020", "2010", typeYear, nil, `invalid value in field 'year' given: '2020-2010'`},
		{"2020", "2020", typeYear, []int{2020}, ``},
	} {
		got, err := getRangeValues(tc.v1, tc.v2, tc.ft)
		if !reflect.DeepEqual(tc.exp, got) {
			t.Errorf("'%s/%s': expected '%#v', got '%#v'", tc.v1, tc.v2, tc.exp, got)
		}

		if tc.err != "" || err != nil {
			if eerr := fmt.Sprintf("%s", err); tc.err != eerr {
				t.Errorf("'%s/%s': expected '%s', got '%s'", tc.v1, tc.v2, tc.err, eerr)
			}
		}
	}
}

func TestValues_getWildcardIntervalValues(t *testing.T) {
	type testCase struct {
		val string
		ft  fieldType
		exp []int
		err string
	}

	for _, tc := range []testCase{
		// Global
		{"XXX", typeSeconds, nil, `strconv.Atoi: parsing "XXX": invalid syntax`},
		{"5", -1, []int{0}, ``},
		{"100", typeSeconds, []int{0}, ``},

		// Seconds
		{"30", typeSeconds, []int{0, 30}, ``},

		// Minutes
		{"30", typeMinutes, []int{0, 30}, ``},

		// Hours
		{"6", typeHours, []int{0, 6, 12, 18}, ``},

		// DoM
		{"10", typeDoM, []int{1, 11, 21, 31}, ``},

		// Month
		{"3", typeMonth, []int{1, 4, 7, 10}, ``},

		// DoW
		{"2", typeDoW, []int{0, 2, 4, 6}, ``},

		// Year
		{"20", typeYear, []int{1970, 1990, 2010, 2030, 2050, 2070, 2090}, ``},
	} {
		got, err := getWildcardIntervalValues(tc.val, tc.ft)
		if !reflect.DeepEqual(tc.exp, got) {
			t.Errorf("'%s': expected '%#v', got '%#v'", tc.val, tc.exp, got)
		}

		if tc.err != "" || err != nil {
			if eerr := fmt.Sprintf("%s", err); tc.err != eerr {
				t.Errorf("'%s': expected '%s', got '%s'", tc.val, tc.err, eerr)
			}
		}
	}
}

func TestValues_getSingleValueIntervalValues(t *testing.T) {
	type testCase struct {
		val  string
		step string
		ft   fieldType
		exp  []int
		err  string
	}

	for _, tc := range []testCase{
		// Global
		{"XXX", "XXX", typeSeconds, nil, `strconv.Atoi: parsing "XXX": invalid syntax`},
		{"5", "XXX", typeSeconds, nil, `strconv.Atoi: parsing "XXX": invalid syntax`},
		{"XXX", "5", typeSeconds, nil, `strconv.Atoi: parsing "XXX": invalid syntax`},
		{"5", "5", -1, nil, `unsupported fieldType given: 'unknown'`},

		// Seconds
		{"45", "5", typeSeconds, []int{45, 50, 55}, ``},
		{"45", "60", typeSeconds, []int{45}, ``},

		// Minutes
		{"45", "5", typeMinutes, []int{45, 50, 55}, ``},
		{"45", "60", typeMinutes, []int{45}, ``},

		// Hours
		{"6", "6", typeHours, []int{6, 12, 18}, ``},
		{"6", "20", typeHours, []int{6}, ``},

		// DoM
		{"10", "10", typeDoM, []int{10, 20, 30}, ``},
		{"10", "30", typeDoM, []int{10}, ``},

		// Month
		{"3", "3", typeMonth, []int{3, 6, 9, 12}, ``},
		{"3", "30", typeMonth, []int{3}, ``},

		// DoW
		{"TUE", "3", typeDoW, []int{2, 5}, ``},
		{"TUE", "8", typeDoW, []int{2}, ``},

		// Year
		{"2020", "20", typeYear, []int{2020, 2040, 2060, 2080}, ``},
		{"2020", "100", typeYear, []int{2020}, ``},
	} {
		got, err := getSingleValueIntervalValues(tc.val, tc.step, tc.ft)
		if !reflect.DeepEqual(tc.exp, got) {
			t.Errorf("'%s/%s': expected '%#v', got '%#v'", tc.val, tc.step, tc.exp, got)
		}

		if tc.err != "" || err != nil {
			if eerr := fmt.Sprintf("%s", err); tc.err != eerr {
				t.Errorf("'%s/%s': expected '%s', got '%s'", tc.val, tc.step, tc.err, eerr)
			}
		}
	}
}

func TestValues_getRangeIntervalValues(t *testing.T) {
	type testCase struct {
		v1   string
		v2   string
		step string
		ft   fieldType
		exp  []int
		err  string
	}

	for _, tc := range []testCase{
		// Global
		{"XXX", "XXX", "XXX", typeSeconds, nil, `strconv.Atoi: parsing "XXX": invalid syntax`},
		{"5", "XXX", "XXX", typeSeconds, nil, `strconv.Atoi: parsing "XXX": invalid syntax`},
		{"XXX", "10", "XXX", typeSeconds, nil, `strconv.Atoi: parsing "XXX": invalid syntax`},
		{"XXX", "XXX", "5", typeSeconds, nil, `strconv.Atoi: parsing "XXX": invalid syntax`},
		{"5", "10", "XXX", typeSeconds, nil, `strconv.Atoi: parsing "XXX": invalid syntax`},
		{"5", "XXX", "5", typeSeconds, nil, `strconv.Atoi: parsing "XXX": invalid syntax`},
		{"XXX", "10", "5", typeSeconds, nil, `strconv.Atoi: parsing "XXX": invalid syntax`},
		{"5", "10", "5", -1, nil, `unsupported fieldType given: 'unknown'`},

		// Seconds
		{"30", "40", "5", typeSeconds, []int{30, 35, 40}, ``},
		{"58", "2", "2", typeSeconds, []int{0, 2, 58}, ``},
		{"5", "5", "5", typeSeconds, []int{5}, ``},

		// Minutes
		{"30", "40", "5", typeMinutes, []int{30, 35, 40}, ``},
		{"58", "2", "2", typeMinutes, []int{0, 2, 58}, ``},
		{"5", "5", "5", typeMinutes, []int{5}, ``},

		// Hours
		{"12", "18", "2", typeHours, []int{12, 14, 16, 18}, ``},
		{"22", "2", "2", typeHours, []int{0, 2, 22}, ``},
		{"5", "5", "2", typeHours, []int{5}, ``},

		// DoM
		{"12", "18", "2", typeDoM, []int{12, 14, 16, 18}, ``},
		{"30", "2", "2", typeDoM, []int{1, 30}, ``},
		{"5", "5", "2", typeDoM, []int{5}, ``},

		// Month
		{"3", "9", "3", typeMonth, []int{3, 6, 9}, ``},
		{"9", "3", "3", typeMonth, []int{3, 9, 12}, ``},
		{"5", "5", "3", typeMonth, []int{5}, ``},

		// DoW
		{"1", "5", "2", typeDoW, []int{1, 3, 5}, ``},
		{"7", "4", "2", typeDoW, []int{0, 2, 4}, ``},
		{"5", "5", "2", typeDoW, []int{5}, ``},

		// Year
		{"2020", "2030", "5", typeYear, []int{2020, 2025, 2030}, ``},
		{"2020", "2010", "5", typeYear, nil, `invalid value in field 'year' given: '2020-2010'`},
		{"2020", "2020", "5", typeYear, []int{2020}, ``},
	} {
		got, err := getRangeIntervalValues(tc.v1, tc.v2, tc.step, tc.ft)
		if !reflect.DeepEqual(tc.exp, got) {
			t.Errorf("'%s-%s/%s': expected '%#v', got '%#v'", tc.v1, tc.v2, tc.step, tc.exp, got)
		}

		if tc.err != "" || err != nil {
			if eerr := fmt.Sprintf("%s", err); tc.err != eerr {
				t.Errorf("'%s-%s/%s': expected '%s', got '%s'", tc.v1, tc.v2, tc.step, tc.err, eerr)
			}
		}
	}
}

/* ==================================================================================================== */

func TestValues_getMinMax(t *testing.T) {
	type testCase struct {
		ft     fieldType
		expMin int
		expMax int
	}

	for _, tc := range []testCase{
		{typeSeconds, 0, 59},
		{typeMinutes, 0, 59},
		{typeHours, 0, 23},
		{typeDoM, 1, 31},
		{typeMonth, 1, 12},
		{typeDoW, 0, 6},
		{typeYear, 1970, 2099},
		{-1, 0, 0}, // Code cannot be reached in the production code...
	} {
		if min, max := getMinMax(tc.ft); min != tc.expMin || max != tc.expMax {
			t.Errorf("expected min/max '%d/%d', got '%d/%d'", tc.expMin, tc.expMax, min, max)
		}
	}
}

func TestValues_isFlexValue(t *testing.T) {
	type testCase struct {
		expr string
		exp  bool
	}

	for _, tc := range []testCase{
		{"L", true},
		{"LW", true},
		{"?", true},
		{"15W", true},
		{"5L", true},
		{"5#3", true},
		{"*", false},
		{".", false},
		{"5", false},
		{"5-10", false},
	} {
		if is := isFlexValue(tc.expr); tc.exp != is {
			t.Errorf("'%s': expected '%#v', got '%#v'", tc.expr, tc.exp, is)
		}
	}
}

func TestValues_toNumVal(t *testing.T) {
	type testCase struct {
		test    string
		isDoW   bool
		isMonth bool
		isNum   bool
		numVal  int
		err     error
	}

	for _, tc := range []testCase{
		{"FRI", true, false, false, 5, nil},
		{"AUG", false, true, false, 8, nil},
		{"5", false, false, true, 5, nil},
		{"XXX", false, false, false, 0, fmt.Errorf("strconv.Atoi: parsing \"XXX\": invalid syntax")},
	} {
		isDoW, isMonth, isNum, numVal, err := toNumVal(tc.test)
		if isDoW != tc.isDoW {
			t.Fatalf("expected '%#v', got '%#v'", tc.isDoW, isDoW)
		}
		if isMonth != tc.isMonth {
			t.Fatalf("expected '%#v', got '%#v'", tc.isMonth, isMonth)
		}
		if isNum != tc.isNum {
			t.Fatalf("expected '%#v', got '%#v'", tc.isNum, isNum)
		}
		if numVal != tc.numVal {
			t.Fatalf("expected '%#v', got '%#v'", tc.numVal, numVal)
		}
		if fmt.Sprintf("%s", err) != fmt.Sprintf("%s", tc.err) {
			t.Fatalf("expected '%#v', got '%#v'", fmt.Sprintf("%s", tc.err), fmt.Sprintf("%s", err))
		}
	}
}

func TestValues_uniqueValues_Empty(t *testing.T) {
	var values []int

	unique := uniqueValues(values)
	if !reflect.DeepEqual(values, unique) {
		t.Errorf("expected '%#v', got '%#v'", values, unique)
	}
}

func TestValues_uniqueValues(t *testing.T) {
	values := []int{3, 2, 1, 5, 5, 5}
	unique := uniqueValues(values)
	exp := []int{1, 2, 3, 5}

	if len(exp) != len(unique) {
		t.Fatalf("expected '%#v', got '%#v'", exp, unique)
	}

	// Test sorting
	for idx := range unique {
		if exp[idx] != unique[idx] {
			t.Errorf("expected '%d', got '%d'", exp[idx], unique[idx])
		}
	}
}
