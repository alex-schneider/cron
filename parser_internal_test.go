package cron

import (
	"fmt"
	"testing"
)

func TestParser_Parse(t *testing.T) {
	s, err := Parse("X")
	if eerr := fmt.Sprintf("%s", err); eerr != `invalid expression given 'X'` {
		t.Errorf("expected '%s', got '%s'", `invalid expression given 'X'`, eerr)
	}
	if s != nil {
		t.Errorf("expected '%#v', got '%#v'", nil, s)
	}

	s, err = Parse("* * * * * * *")
	if err != nil {
		t.Errorf("%#v", err)
	}
	if s == nil {
		t.Errorf("expected 'schedule', got '%#v'", nil)
	}
}
