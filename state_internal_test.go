package cron

import "testing"

func TestState_String(t *testing.T) {
	var s state

	s = StateFound
	if s.String() != "found" {
		t.Errorf("expected 'found', got '%s'", s.String())
	}

	s = StateOnceExec
	if s.String() != "once-exec" {
		t.Errorf("expected 'once-exec', got '%s'", s.String())
	}

	s = StateNoMatches
	if s.String() != "no-matches" {
		t.Errorf("expected 'no-matches', got '%s'", s.String())
	}

	s = StateZeroTime
	if s.String() != "zero-time" {
		t.Errorf("expected 'zero-time', got '%s'", s.String())
	}

	// Code cannot be reached in the production code...
	s = -1
	if s.String() != "unknown" {
		t.Errorf("expected 'unknown', got '%s'", s.String())
	}
}
