package cron

import "testing"

func TestTypes_String(t *testing.T) {
	var ft fieldType

	ft = typeSeconds
	if ft.String() != "seconds" {
		t.Errorf("expected 'second', got '%s'", ft.String())
	}

	ft = typeMinutes
	if ft.String() != "minutes" {
		t.Errorf("expected 'minute', got '%s'", ft.String())
	}

	ft = typeHours
	if ft.String() != "hours" {
		t.Errorf("expected 'minute', got '%s'", ft.String())
	}

	ft = typeDoM
	if ft.String() != "day-of-month" {
		t.Errorf("expected 'minute', got '%s'", ft.String())
	}

	ft = typeMonth
	if ft.String() != "month" {
		t.Errorf("expected 'minute', got '%s'", ft.String())
	}

	ft = typeDoW
	if ft.String() != "day-of-week" {
		t.Errorf("expected 'minute', got '%s'", ft.String())
	}

	ft = typeYear
	if ft.String() != "year" {
		t.Errorf("expected 'year', got '%s'", ft.String())
	}

	// Code cannot be reached in the production code...
	ft = -1
	if ft.String() != "unknown" {
		t.Errorf("expected 'unknown', got '%s'", ft.String())
	}
}
