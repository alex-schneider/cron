// Copyright 2021 Alex Schneider. All rights reserved.

package cron

import (
	"context"
	"fmt"
	"os/exec"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// Schedule represents the <cron.Schedule> object.
type Schedule struct {
	args        []string
	command     string
	deaktivated bool
	fields      *fields
}

/* ==================================================================================================== */

// Bind binds a command to the scheduler.
func (s *Schedule) Bind(command string) *Schedule {
	parts := reFieldsMatcher.FindAllString(command, -1)

	s.command = parts[0]
	if len(parts) > 1 {
		s.args = parts[1:]
	}

	return s
}

// Run starts the execution of a job.
func (s *Schedule) Run(ctx context.Context, doneCh <-chan struct{}, logger logrus.FieldLogger) {
	go s.run(ctx, doneCh, logger)
}

/* ==================================================================================================== */

// deactivate deactivates the execution of a job with the `@reboot` expression.
func (s *Schedule) deactivate() {
	s.deaktivated = true
}

// next calculates the <time.Time> for the next execution of the <cron.Schedule>,
// that is greater than the given reference time.
//
// A zero time value is returned if one of the following special cases is accured:
//  * the given reference time is zero.
//  * the expression is defined as `@reboot`.
//  * no new time can be found for the next execution.
// The second return value notifies the caller about the accured case.
func (s *Schedule) next(referenceTime time.Time) (time.Time, state) {
	if referenceTime.IsZero() {
		return time.Time{}, StateZeroTime
	} else if s.fields.once {
		return time.Time{}, StateOnceExec
	}

	referenceTime = referenceTime.Add(time.Duration(1) * time.Second)

	return s.fromNextBestYear(referenceTime)
}

func (s *Schedule) run(ctx context.Context, doneCh <-chan struct{}, logger logrus.FieldLogger) {
	var ticker *time.Ticker

	defer func() {
		if rec := recover(); rec != nil {
			logger.WithField("panic", string(debug.Stack())).Panic(rec)
		}

		if ticker != nil {
			ticker.Stop()
		}
	}()

	if s.fields.once && !s.deaktivated {
		s.deactivate()

		go func() {
			err := exec.CommandContext(ctx, s.command, s.args...).Run()
			if err != nil {
				if e, ok := err.(*exec.ExitError); ok {
					logger.Errorf("Cronjob finishes with error: %v", e)
				}
			}
		}()

		logger.Debug("Cronjob finishes (once)")

		return
	}

	now := time.Now()
	next, state := s.next(now)
	if state == StateFound {
		ticker = time.NewTicker(next.Sub(now))

		for {
			select {
			case <-doneCh:
				ticker.Stop()

				logger.Debug("Context done")
				return
			case now := <-ticker.C:
				ticker.Stop()

				logger.Debug("Cronjob runs")

				next, state = s.next(now)
				if state != StateFound {
					logger.Debugf("Cronjob finishes with state '%s'", state)

					return
				}

				ticker = time.NewTicker(next.Sub(now))

				go func() {
					err := exec.CommandContext(ctx, s.command, s.args...).Run()
					if err != nil {
						if e, ok := err.(*exec.ExitError); ok {
							logger.Errorf("Cronjob runs with error: %v", e)
						}
					}
				}()
			}
		}
	}
}

/* ==================================================================================================== */

func (s *Schedule) fromNextBestYear(referenceTime time.Time) (time.Time, state) {
	values := s.fields.year.combinations[0].values

	i := sort.SearchInts(values, referenceTime.Year())
	if i == len(values) {
		return time.Time{}, StateNoMatches
	}

	if values[i] != referenceTime.Year() {
		return s.fromNextBestYear(time.Date(
			values[i],
			time.Month(s.fields.month.combinations[0].values[0]),
			1,
			s.fields.hours.combinations[0].values[0],
			s.fields.minutes.combinations[0].values[0],
			s.fields.seconds.combinations[0].values[0],
			0,
			referenceTime.Location(),
		))
	}

	return s.fromNextBestMonth(referenceTime)
}

func (s *Schedule) fromNextBestMonth(referenceTime time.Time) (time.Time, state) {
	values := s.fields.month.combinations[0].values

	i := sort.SearchInts(values, int(referenceTime.Month()))
	if i == len(values) {
		return s.fromNextBestYear(time.Date(
			referenceTime.AddDate(1, 0, 0).Year(),
			time.Month(s.fields.month.combinations[0].values[0]),
			1,
			s.fields.hours.combinations[0].values[0],
			s.fields.minutes.combinations[0].values[0],
			s.fields.seconds.combinations[0].values[0],
			0,
			referenceTime.Location(),
		))
	} else if values[i] != int(referenceTime.Month()) {
		referenceTime = referenceTime.AddDate(0, values[i]-int(referenceTime.Month()), 0)

		return s.fromNextBestYear(time.Date(
			referenceTime.Year(),
			time.Month(referenceTime.Month()),
			1,
			s.fields.hours.combinations[0].values[0],
			s.fields.minutes.combinations[0].values[0],
			s.fields.seconds.combinations[0].values[0],
			0,
			referenceTime.Location(),
		))
	}

	return s.fromNextBestDay(referenceTime)
}

func (s *Schedule) fromNextBestDay(referenceTime time.Time) (time.Time, state) {
	values := s.getDaysValues(referenceTime)

	i := sort.SearchInts(values, referenceTime.Day())
	if i == len(values) {
		referenceTime = referenceTime.AddDate(0, 1, 0)

		return s.fromNextBestYear(time.Date(
			referenceTime.Year(),
			time.Month(referenceTime.Month()),
			1,
			s.fields.hours.combinations[0].values[0],
			s.fields.minutes.combinations[0].values[0],
			s.fields.seconds.combinations[0].values[0],
			0,
			referenceTime.Location(),
		))
	} else if values[i] != referenceTime.Day() {
		referenceTime = referenceTime.AddDate(0, 0, values[i]-referenceTime.Day())

		return time.Date(
			referenceTime.Year(),
			time.Month(referenceTime.Month()),
			referenceTime.Day(),
			s.fields.hours.combinations[0].values[0],
			s.fields.minutes.combinations[0].values[0],
			s.fields.seconds.combinations[0].values[0],
			0,
			referenceTime.Location(),
		), StateFound
	}

	return s.fromNextBestHour(referenceTime)
}

func (s *Schedule) fromNextBestHour(referenceTime time.Time) (time.Time, state) {
	values := s.fields.hours.combinations[0].values

	i := sort.SearchInts(values, referenceTime.Hour())
	if i == len(values) {
		referenceTime = referenceTime.AddDate(0, 0, 1)

		return s.fromNextBestYear(time.Date(
			referenceTime.Year(),
			time.Month(referenceTime.Month()),
			referenceTime.Day(),
			s.fields.hours.combinations[0].values[0],
			s.fields.minutes.combinations[0].values[0],
			s.fields.seconds.combinations[0].values[0],
			0,
			referenceTime.Location(),
		))
	} else if values[i] != referenceTime.Hour() {
		referenceTime = referenceTime.Add(time.Duration(values[i]-referenceTime.Hour()) * time.Hour)

		return time.Date(
			referenceTime.Year(),
			time.Month(referenceTime.Month()),
			referenceTime.Day(),
			referenceTime.Hour(),
			s.fields.minutes.combinations[0].values[0],
			s.fields.seconds.combinations[0].values[0],
			0,
			referenceTime.Location(),
		), StateFound
	}

	return s.fromNextBestMinute(referenceTime)
}

func (s *Schedule) fromNextBestMinute(referenceTime time.Time) (time.Time, state) {
	values := s.fields.minutes.combinations[0].values

	i := sort.SearchInts(values, referenceTime.Minute())
	if i == len(values) {
		referenceTime = referenceTime.Add(time.Duration(1) * time.Hour)

		return s.fromNextBestYear(time.Date(
			referenceTime.Year(),
			time.Month(referenceTime.Month()),
			referenceTime.Day(),
			referenceTime.Hour(),
			s.fields.minutes.combinations[0].values[0],
			s.fields.seconds.combinations[0].values[0],
			0,
			referenceTime.Location(),
		))
	} else if values[i] != referenceTime.Minute() {
		referenceTime = referenceTime.Add(time.Duration(values[i]-referenceTime.Minute()) * time.Minute)

		return time.Date(
			referenceTime.Year(),
			time.Month(referenceTime.Month()),
			referenceTime.Day(),
			referenceTime.Hour(),
			referenceTime.Minute(),
			s.fields.seconds.combinations[0].values[0],
			0,
			referenceTime.Location(),
		), StateFound
	}

	return s.fromNextBestSecond(referenceTime)
}

func (s *Schedule) fromNextBestSecond(referenceTime time.Time) (time.Time, state) {
	values := s.fields.seconds.combinations[0].values

	i := sort.SearchInts(values, referenceTime.Second())
	if i == len(values) {
		referenceTime = referenceTime.Add(time.Duration(1) * time.Minute)

		return s.fromNextBestYear(time.Date(
			referenceTime.Year(),
			time.Month(referenceTime.Month()),
			referenceTime.Day(),
			referenceTime.Hour(),
			referenceTime.Minute(),
			s.fields.seconds.combinations[0].values[0],
			0,
			referenceTime.Location(),
		))
	} else if values[i] != referenceTime.Second() {
		referenceTime = referenceTime.Add(time.Duration(values[i]-referenceTime.Second()) * time.Second)
	}

	return referenceTime, StateFound
}

/* ==================================================================================================== */

func (s *Schedule) getDaysValues(referenceTime time.Time) []int {
	min := referenceTime.AddDate(0, 0, -1*referenceTime.Day()+1)
	max := min.AddDate(0, 1, -1)

	domValues := s.getDaysValuesFromDoM(min, max)
	dowValues := s.getDaysValuesFromDoW(min, max)

	return uniqueValues(append(domValues, dowValues...))
}

func (s *Schedule) getDaysValuesFromDoM(min, max time.Time) []int {
	var values []int

	for _, combi := range s.fields.dom.combinations {
		if combi.unit == "?" {
			break
		}

		switch combi.unit {
		case "L": // Last day of month
			values = append(values, max.Day())
		case "LW": // Last weekday (MON-FRI) of month
			values = append(values, getDaysValuesFromDoMLastWeekday(max))
		case "W": // `15W` (nearest weekday (MON-FRI) of the month to the 15.)
			values = append(values, getDaysValuesFromDoMWeekday(combi.values, min, max)...)
		default:
			for _, d := range combi.values {
				if d <= max.Day() {
					values = append(values, d)
				}
			}
		}
	}

	return values
}

func getDaysValuesFromDoMLastWeekday(max time.Time) int {
	switch int(max.Weekday()) {
	case 0:
		return max.Day() - 2
	case 6:
		return max.Day() - 1
	default:
		return max.Day()
	}
}

func getDaysValuesFromDoMWeekday(pool []int, min, max time.Time) []int {
	var values []int

	refDay := pool[0]
	curr := min.AddDate(0, 0, refDay-1)
	wd := int(curr.Weekday())

	if wd >= 1 && wd <= 5 {
		values = append(values, curr.Day())
	} else if wd == 0 {
		if curr.Day() == max.Day() {
			values = append(values, curr.AddDate(0, 0, -2).Day())
		} else {
			values = append(values, curr.AddDate(0, 0, 1).Day())
		}
	} else if wd == 6 {
		if curr.Day() == min.Day() {
			values = append(values, curr.AddDate(0, 0, 2).Day())
		} else {
			values = append(values, curr.AddDate(0, 0, -1).Day())
		}
	}

	return values
}

/* ==================================================================================================== */

func (s *Schedule) getDaysValuesFromDoW(min, max time.Time) []int {
	var values []int

	for _, combi := range s.fields.dow.combinations {
		if combi.unit == "?" {
			break
		}

		switch combi.unit {
		case "L": // `5L` (last FRI of the month)
			values = append(values, getDaysValuesFromDoWLast(combi.values, min, max)...)
		case "#": // `5#3` (nth (1-5, here 3) DoW (0-7, here FRI) of the month)
			values = append(values, getDaysValuesFromDoWHash(combi.values, min, max)...)
		default:
			values = append(values, getDaysValuesFromDoWDefault(combi.values, min, max)...)
		}
	}

	return values
}

func getDaysValuesFromDoWLast(pool []int, min, max time.Time) []int {
	var values []int

	refDay := pool[0]

	for curr := max.AddDate(0, 0, 0); curr.Day() > min.Day(); curr = curr.AddDate(0, 0, -1) {
		if int(curr.Weekday()) == refDay {
			values = append(values, curr.Day())

			break
		}
	}

	return values
}

func getDaysValuesFromDoWHash(pool []int, min, max time.Time) []int {
	var values []int
	var seen int

	day := pool[0]
	nth := pool[1]

	for curr := min.AddDate(0, 0, 0); curr.Day() <= max.Day(); curr = curr.AddDate(0, 0, 1) {
		if day == int(curr.Weekday()) {
			seen++

			if seen == nth {
				values = append(values, curr.Day())

				break
			}
		}

		if curr.Day() == max.Day() {
			break
		}
	}

	return values
}

func getDaysValuesFromDoWDefault(pool []int, min, max time.Time) []int {
	var values []int

	for _, refDay := range pool {
		for curr := min.AddDate(0, 0, 0); curr.Day() <= max.Day(); curr = curr.AddDate(0, 0, 1) {
			if int(curr.Weekday()) == refDay {
				values = append(values, curr.Day())
			}

			if curr.Day() == max.Day() {
				break
			}
		}
	}

	return values
}

/* ==================================================================================================== */

func expressionFromMacro(macro string) (string, error) {
	if !strings.HasPrefix(macro, "@") {
		return "", nil
	}

	switch macro {
	case "@yearly", "@annually":
		return "0 0 0 1 1 * *", nil
	case "@monthly":
		return "0 0 0 1 * * *", nil
	case "@weekly":
		return "0 0 0 * * 0 *", nil
	case "@daily", "@midnight":
		return "0 0 0 * * * *", nil
	case "@hourly":
		return "0 0 * * * * *", nil
	case "@minutely", "@every_minute":
		return "0 * * * * * *", nil
	case "@secondly", "@every_second":
		return "* * * * * * *", nil
	case "@reboot":
		return "~", nil
	}

	return "", fmt.Errorf("unsupported macro given '%s'", macro)
}
