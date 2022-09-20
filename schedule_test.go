package cron_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/alex-schneider/cron"
)

func TestSchedule_NewJobCh_Error(t *testing.T) {
	s, err := cron.NewJobCh(context.TODO(), "X")
	if eerr := fmt.Sprintf("%s", err); eerr != `invalid expression given 'X'` {
		t.Errorf("expected '%s', got '%s'", `invalid expression given 'X'`, eerr)
	}
	if s != nil {
		t.Errorf("expected '%#v', got '%#v'", nil, s)
	}
}

func TestSchedule_NewJobCh(t *testing.T) {
	ctx, cancelFn := context.WithCancel(context.TODO())
	defer cancelFn()

	ch, err := cron.NewJobCh(ctx, "* * * * * * *")
	if err != nil {
		t.Errorf("%#v", err)
	}
	if ch == nil {
		t.Error("expected 'jobCh', got 'NIL'")
	}

	var i int

testLoop:
	for {
		select {
		case job := <-ch:
			if job.State == int(cron.StateFound) {
				i++ // Trigger some job...
			}

			if i == 2 {
				break testLoop
			}
		case <-ctx.Done():
			t.Errorf("Unexpected 'ctx.Done()'")
		}
	}

	if i != 2 {
		t.Errorf("expected '2', got '%d'", i)
	}
}

func TestSchedule_NewJobCh_Once(t *testing.T) {
	ctx, cancelFn := context.WithCancel(context.TODO())
	defer cancelFn()

	ch, err := cron.NewJobCh(ctx, "@reboot")
	if err != nil {
		t.Errorf("%#v", err)
	}
	if ch == nil {
		t.Error("expected 'jobCh', got 'NIL'")
	}

testLoop:
	for {
		select {
		case job := <-ch:
			if job.State != int(cron.StateOnceExec) {
				t.Errorf("Unexpected state '%#v'", job.State)
			}

			break testLoop
		case <-ctx.Done():
			t.Errorf("Unexpected 'ctx.Done()'")
		}
	}
}
