package cron

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"
)

// 2022-12-31 23:59:59 is a Saturday.
var testScheduleTime = time.Date(2022, 12, 31, 23, 59, 59, 0, startupTime.Location())

/* ==================================================================================================== */

func TestSchedule_Next(t *testing.T) {
	type testCase struct {
		expr    string
		state   state
		refTime time.Time
		expTime time.Time
		err     string
	}

	for _, tc := range []testCase{
		{"1 1 1 1 1", StateZeroTime, time.Time{}, time.Time{}, ``},
		{"@reboot", StateOnceExec, startupTime, time.Time{}, ``},
		{
			"0 0 0 29 2 ? *",
			StateFound,
			testScheduleTime,
			time.Date(2024, 2, 29, 0, 0, 0, 0, startupTime.Location()),
			``,
		},
	} {
		s, err := createTestScheduler(tc.expr)
		if err != nil {
			t.Errorf("'%s': unexpected error: %#v", tc.expr, err)
		}

		tm, state := s.next(tc.refTime)
		if state != tc.state {
			t.Errorf("'%s': expected '%d', got '%d'", tc.expr, tc.state, state)
		}

		if tm.String() != tc.expTime.String() {
			t.Errorf("'%s': expected '%s', got '%s'", tc.expr, tc.expTime.String(), tm.String())
		}
	}
}

func TestSchedule_Run(t *testing.T) {
	s, err := createTestScheduler("* * * * * * 1970")
	if err != nil {
		t.Errorf("'Unexpected error: %#v", err)
	}

	s.ctx = context.TODO()
	s.jobCh = make(chan *Job)

	nowFn := func() time.Time {
		return time.Date(1970, 12, 31, 23, 59, 58, 0, startupTime.Location())
	}

	go s.run(nowFn)

	var i int

testLoop:
	for {
		select {
		case job := <-s.jobCh:
			if i == 0 && job.State != int(StateNoMatches) {
				t.Errorf("Unexpected state: %#v", job.State)
			}

			i++

			if i >= 1 {
				break testLoop
			}
		case <-s.ctx.Done():
			t.Errorf("Unexpected 'ctx.Done()'")
		}
	}
}

/* ==================================================================================================== */

func TestSchedule_fromNextBestYear(t *testing.T) {
	type testCase struct {
		expr    string
		state   state
		refTime time.Time
		expTime time.Time
	}

	for _, tc := range []testCase{
		{"1 1 1 1 1 ? 2020-2022", StateNoMatches, testScheduleTime, time.Time{}},
		{
			"1 1 1 1 1 ? 2023-2024",
			StateFound,
			testScheduleTime,
			time.Date(2023, 1, 1, 1, 1, 1, 0, startupTime.Location()),
		},
	} {
		s, err := createTestScheduler(tc.expr)
		if err != nil {
			t.Errorf("'%s': unexpected error: %#v", tc.expr, err)
		}

		tm, state := s.fromNextBestYear(tc.refTime)
		if state != tc.state {
			t.Errorf("'%s': expected '%d', got '%d'", tc.expr, tc.state, state)
		}

		if tm.String() != tc.expTime.String() {
			t.Errorf("'%s': expected '%s', got '%s'", tc.expr, tc.expTime.String(), tm.String())
		}
	}
}

func TestSchedule_fromNextBestMonth(t *testing.T) {
	type testCase struct {
		expr    string
		state   state
		refTime time.Time
		expTime time.Time
	}

	for _, tc := range []testCase{
		{"1 1 1 1 1 ? 2020-2022", StateNoMatches, testScheduleTime, time.Time{}},
		{
			"1 1 1 1 1 ? 2023-2024",
			StateFound,
			testScheduleTime,
			time.Date(2023, 1, 1, 1, 1, 1, 0, startupTime.Location()),
		},
		{
			"1 1 1 1 12 ? 2022-2024",
			StateFound,
			testScheduleTime,
			time.Date(2023, 12, 1, 1, 1, 1, 0, startupTime.Location()),
		},
	} {
		s, err := createTestScheduler(tc.expr)
		if err != nil {
			t.Errorf("'%s': unexpected error: %#v", tc.expr, err)
		}

		tm, state := s.fromNextBestMonth(tc.refTime)
		if state != tc.state {
			t.Errorf("'%s': expected '%d', got '%d'", tc.expr, tc.state, state)
		}

		if tm.String() != tc.expTime.String() {
			t.Errorf("'%s': expected '%s', got '%s'", tc.expr, tc.expTime.String(), tm.String())
		}
	}
}

func TestSchedule_fromNextBestDay(t *testing.T) {
	type testCase struct {
		expr    string
		state   state
		refTime time.Time
		expTime time.Time
	}

	for _, tc := range []testCase{
		{"1 1 1 1 1 ? 2020-2022", StateNoMatches, testScheduleTime, time.Time{}},
		{
			"1 1 1 1 1 ? 2023-2024",
			StateFound,
			testScheduleTime,
			time.Date(2023, 1, 1, 1, 1, 1, 0, startupTime.Location()),
		},
		{
			"1 1 1 31 1 ? 2022-2024",
			StateFound,
			testScheduleTime,
			time.Date(2023, 1, 31, 1, 1, 1, 0, startupTime.Location()),
		},
		{
			"1 1 1 31 1 ? 2022-2024",
			StateFound,
			testScheduleTime,
			time.Date(2023, 1, 31, 1, 1, 1, 0, startupTime.Location()),
		},
		{
			"1 1 1 30,31 1,12 ? 2022-2024",
			StateFound,
			time.Date(2022, 12, 28, 23, 59, 59, 0, startupTime.Location()),
			time.Date(2022, 12, 30, 1, 1, 1, 0, startupTime.Location()),
		},
	} {
		s, err := createTestScheduler(tc.expr)
		if err != nil {
			t.Errorf("'%s': unexpected error: %#v", tc.expr, err)
		}

		tm, state := s.fromNextBestDay(tc.refTime)
		if state != tc.state {
			t.Errorf("'%s': expected '%d', got '%d'", tc.expr, tc.state, state)
		}

		if tm.String() != tc.expTime.String() {
			t.Errorf("'%s': expected '%s', got '%s'", tc.expr, tc.expTime.String(), tm.String())
		}
	}
}

func TestSchedule_fromNextBestHour(t *testing.T) {
	type testCase struct {
		expr    string
		state   state
		refTime time.Time
		expTime time.Time
	}

	for _, tc := range []testCase{
		{"1 1 1 1 1 ? 2020-2022", StateNoMatches, testScheduleTime, time.Time{}},
		{
			"1 1 1 1 1 ? 2023-2024",
			StateFound,
			testScheduleTime,
			time.Date(2023, 1, 1, 1, 1, 1, 0, startupTime.Location()),
		},
		{
			"1 1 3 1,2,31 12 ? 2022",
			StateFound,
			time.Date(2022, 12, 15, 2, 1, 1, 0, startupTime.Location()),
			time.Date(2022, 12, 15, 3, 1, 1, 0, startupTime.Location()),
		},
	} {
		s, err := createTestScheduler(tc.expr)
		if err != nil {
			t.Errorf("'%s': unexpected error: %#v", tc.expr, err)
		}

		tm, state := s.fromNextBestHour(tc.refTime)
		if state != tc.state {
			t.Errorf("'%s': expected '%d', got '%d'", tc.expr, tc.state, state)
		}

		if tm.String() != tc.expTime.String() {
			t.Errorf("'%s': expected '%s', got '%s'", tc.expr, tc.expTime.String(), tm.String())
		}
	}
}

func TestSchedule_fromNextBestMinute(t *testing.T) {
	type testCase struct {
		expr    string
		state   state
		refTime time.Time
		expTime time.Time
	}

	for _, tc := range []testCase{
		{"1 1 1 1 1 ? 2020-2022", StateNoMatches, testScheduleTime, time.Time{}},
		{
			"1 1 1 1 1 ? 2023-2024",
			StateFound,
			testScheduleTime,
			time.Date(2023, 1, 1, 1, 1, 1, 0, startupTime.Location()),
		},
		{
			"1 5,6 1 15 12 ? 2022",
			StateFound,
			time.Date(2022, 12, 15, 1, 2, 1, 0, startupTime.Location()),
			time.Date(2022, 12, 15, 1, 5, 1, 0, startupTime.Location()),
		},
	} {
		s, err := createTestScheduler(tc.expr)
		if err != nil {
			t.Errorf("'%s': unexpected error: %#v", tc.expr, err)
		}

		tm, state := s.fromNextBestMinute(tc.refTime)
		if state != tc.state {
			t.Errorf("'%s': expected '%d', got '%d'", tc.expr, tc.state, state)
		}

		if tm.String() != tc.expTime.String() {
			t.Errorf("'%s': expected '%s', got '%s'", tc.expr, tc.expTime.String(), tm.String())
		}
	}
}

func TestSchedule_fromNextBestSecond(t *testing.T) {
	type testCase struct {
		expr    string
		state   state
		refTime time.Time
		expTime time.Time
	}

	for _, tc := range []testCase{
		{"1 1 1 1 1 ? 2020-2022", StateNoMatches, testScheduleTime, time.Time{}},
		{
			"1 1 1 1 1 ? 2023-2024",
			StateFound,
			testScheduleTime,
			time.Date(2023, 1, 1, 1, 1, 1, 0, startupTime.Location()),
		},
		{
			"50 1 2 15 12 ? 2022",
			StateFound,
			time.Date(2022, 12, 15, 1, 2, 1, 0, startupTime.Location()),
			time.Date(2022, 12, 15, 1, 2, 50, 0, startupTime.Location()),
		},
	} {
		s, err := createTestScheduler(tc.expr)
		if err != nil {
			t.Errorf("'%s': unexpected error: %#v", tc.expr, err)
		}

		tm, state := s.fromNextBestSecond(tc.refTime)
		if state != tc.state {
			t.Errorf("'%s': expected '%d', got '%d'", tc.expr, tc.state, state)
		}

		if tm.String() != tc.expTime.String() {
			t.Errorf("'%s': expected '%s', got '%s'", tc.expr, tc.expTime.String(), tm.String())
		}
	}
}

/* ==================================================================================================== */

func TestSchedule_getDaysValuesFromDoM(t *testing.T) {
	type testCase struct {
		expr string
		t    time.Time
		exp  []int
	}

	for _, tc := range []testCase{
		{"1 1 1 L 1 ? 2022", testScheduleTime, []int{31}},
		{"1 1 1 LW 1 ? 2022", testScheduleTime, []int{30}},
		{"1 1 1 LW 1 ? 2022", time.Date(2021, 10, 1, 0, 0, 0, 0, startupTime.Location()), []int{29}},
		{"1 1 1 LW 1 ? 2022", time.Date(2021, 12, 1, 0, 0, 0, 0, startupTime.Location()), []int{31}},
		{"1 1 1 18W 1 ? 2022", testScheduleTime, []int{19}},
		{"1 1 1 17W 1 ? 2022", testScheduleTime, []int{16}},
		{"1 1 1 15W 1 ? 2022", testScheduleTime, []int{15}},
		{"1 1 1 31W 1 ? 2022", time.Date(2021, 10, 1, 0, 0, 0, 0, startupTime.Location()), []int{29}},
		{"1 1 1 1W 1 ? 2022", time.Date(2022, 1, 1, 0, 0, 0, 0, startupTime.Location()), []int{3}},
		{"1 1 1 10,20,25 1 ? 2022", testScheduleTime, []int{10, 20, 25}},
		{"1 1 1 ? 1 0 2022", testScheduleTime, []int{4, 11, 18, 25}}, // From DoW
	} {
		s, err := createTestScheduler(tc.expr)
		if err != nil {
			t.Errorf("'%s': unexpected error: %#v", tc.expr, err)
		}

		got := s.getDaysValues(tc.t)
		if !reflect.DeepEqual(tc.exp, got) {
			t.Errorf("'%s': expected '%#v', got '%#v'", tc.expr, tc.exp, got)
		}
	}
}

func TestSchedule_getDaysValuesFromDoW(t *testing.T) {
	type testCase struct {
		expr string
		t    time.Time
		exp  []int
	}

	for _, tc := range []testCase{
		{"1 1 1 ? 1 L 2022", testScheduleTime, []int{3, 10, 17, 24, 31}},
		{"1 1 1 ? 1 5L 2022", testScheduleTime, []int{30}},
		{"1 1 1 ? 1 5#4 2022", testScheduleTime, []int{23}},
		{"1 1 1 ? 1 6#5 2022", time.Date(2021, 12, 15, 1, 2, 1, 0, startupTime.Location()), nil},
		{"1 1 1 1 1 ? 2022", testScheduleTime, []int{1}}, // From DoM
	} {
		s, err := createTestScheduler(tc.expr)
		if err != nil {
			t.Errorf("'%s': unexpected error: %#v", tc.expr, err)
		}

		got := s.getDaysValues(tc.t)
		if !reflect.DeepEqual(tc.exp, got) {
			t.Errorf("'%s': expected '%#v', got '%#v'", tc.expr, tc.exp, got)
		}
	}
}

/* ==================================================================================================== */

func TestSchedule_expressionFromMacro(t *testing.T) {
	type testCase struct {
		macro string
		exp   string
		err   string
	}

	for _, tc := range []testCase{
		{"test", "", ""},
		{"@test", "", "unsupported macro given '@test'"},
		{"@yearly", "0 0 0 1 1 * *", ""},
		{"@annually", "0 0 0 1 1 * *", ""},
		{"@monthly", "0 0 0 1 * * *", ""},
		{"@weekly", "0 0 0 * * 0 *", ""},
		{"@daily", "0 0 0 * * * *", ""},
		{"@midnight", "0 0 0 * * * *", ""},
		{"@hourly", "0 0 * * * * *", ""},
		{"@minutely", "0 * * * * * *", ""},
		{"@every_minute", "0 * * * * * *", ""},
		{"@secondly", "* * * * * * *", ""},
		{"@every_second", "* * * * * * *", ""},
		{"@reboot", "~", ""},
	} {
		expression, err := expressionFromMacro(tc.macro)
		if tc.exp != expression {
			t.Errorf("'%s': expected '%s', got '%s'", tc.macro, tc.exp, expression)
		}

		if tc.err != "" || err != nil {
			if eerr := fmt.Sprintf("%s", err); tc.err != eerr {
				t.Errorf("'%s': expected '%s', got '%s'", tc.macro, tc.err, eerr)
			}
		}
	}
}

/* ==================================================================================================== */

func createTestScheduler(expression string) (*schedule, error) {
	fields, err := getFields(expression)
	if err != nil {
		return nil, err
	}

	s := &schedule{
		fields: fields,
	}

	return s, nil
}
