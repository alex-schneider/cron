// Copyright 2021 Alex Schneider. All rights reserved.

package cron

const (
	// StateFound is returned if a new time could be found for the next execution.
	StateFound state = iota
	// StateOnceExec is returned if the expression is defined as `@reboot`.
	StateOnceExec
	// StateNoMatches is returned if a new time could not be found for the next execution.
	StateNoMatches
	// StateZeroTime is returned if the given reference time is zero.
	StateZeroTime
)

// State represents an accured case while the execution of the <cron.Next> method.
type state int

// String implements the <fmt.Stringer> interface.
func (s state) String() string {
	switch s {
	case StateFound:
		return "found"
	case StateOnceExec:
		return "once-exec"
	case StateNoMatches:
		return "no-matches"
	case StateZeroTime:
		return "zero-time"
	}

	// Code cannot be reached in the production code...
	return "unknown"
}
