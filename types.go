// Copyright 2021 Alex Schneider. All rights reserved.

package cron

const (
	typeSeconds fieldType = iota
	typeMinutes
	typeHours
	typeDoM
	typeMonth
	typeDoW
	typeYear
)

type fieldType int

/* ==================================================================================================== */

// String implements the <fmt.Stringer> interface.
func (ft fieldType) String() string {
	switch ft {
	case typeSeconds:
		return "seconds"
	case typeMinutes:
		return "minutes"
	case typeHours:
		return "hours"
	case typeDoM:
		return "day-of-month"
	case typeMonth:
		return "month"
	case typeDoW:
		return "day-of-week"
	case typeYear:
		return "year"
	}

	// Code cannot be reached in the production code...
	return "unknown"
}
