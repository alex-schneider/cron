package cron

import (
	"fmt"
	"reflect"
	"testing"
)

func TestFields_getFields(t *testing.T) {
	expression := "@test"
	f, err := getFields(expression)
	if eerr := fmt.Sprintf("%s", err); eerr != `unsupported macro given '@test'` {
		t.Errorf("'%s': expected '%s', got '%s'", expression, `unsupported macro given '@test'`, eerr)
	}
	if f != nil {
		t.Errorf("'%s': expected '%#v', got '%#v'", expression, nil, f)
	}

	expression = "@reboot"
	f, err = getFields(expression)
	if err != nil {
		t.Errorf("'%s': unexpected error: '%#v'", expression, err)
	}
	if f == nil {
		t.Fatalf("'%s': expected 'fields', got '%#v'", expression, f)
	}
	if !f.once {
		t.Errorf("'%s': expected once flag 'true', got 'false'", expression)
	}

	expression = "@yearly"
	f, err = getFields(expression)
	if err != nil {
		t.Errorf("'%s': unexpected error: '%#v'", expression, err)
	}
	if f == nil {
		t.Errorf("'%s': expected 'fields', got '%#v'", expression, f)
	}

	expression = "* * * * *"
	f, err = getFields(expression)
	if err != nil {
		t.Errorf("'%s': unexpected error: '%#v'", expression, err)
	}
	if f == nil {
		t.Fatalf("'%s': expected 'fields', got '%#v'", expression, f)
	}
	if !reflect.DeepEqual(f.seconds.combinations[0].values, []int{0}) {
		t.Errorf("'%s': expected '%#v', got '%#v'", expression, []int{0}, f.seconds.combinations[0].values)
	}
	if !reflect.DeepEqual(f.year.combinations[0].values, yearValues) {
		t.Errorf("'%s': expected '%#v', got '%#v'", expression, yearValues, f.seconds.combinations[0].values)
	}
}

func TestFields_getFields_InvalidExpression(t *testing.T) {
	expression := "* * *"
	f, err := getFields(expression)
	if eerr := fmt.Sprintf("%s", err); eerr != `invalid expression given '* * *'` {
		t.Errorf("'%s': expected '%s', got '%s'", expression, `invalid expression given '* * *'`, eerr)
	}
	if f != nil {
		t.Errorf("'%s': expected '%#v', got '%#v'", expression, nil, f)
	}

	expression = "* * * * * * * *"
	f, err = getFields(expression)
	if eerr := fmt.Sprintf("%s", err); eerr != `invalid expression given '* * * * * * * *'` {
		t.Errorf("'%s': expected '%s', got '%s'", expression, `invalid expression given '* * * * * * * *'`, eerr)
	}
	if f != nil {
		t.Errorf("'%s': expected '%#v', got '%#v'", expression, nil, f)
	}
}

func TestFields_createFields(t *testing.T) {
	type testCase struct {
		parts  []string
		fields bool
		err    string
	}

	for _, tc := range []testCase{
		{[]string{"X"}, false, `invalid value in field 'seconds' given: 'X'`},
		{[]string{"1", "X"}, false, `invalid value in field 'minutes' given: 'X'`},
		{[]string{"1", "1", "X"}, false, `invalid value in field 'hours' given: 'X'`},
		{[]string{"1", "1", "1", "X"}, false, `invalid value in field 'day-of-month' given: 'X'`},
		{[]string{"1", "1", "1", "1", "X"}, false, `invalid value in field 'month' given: 'X'`},
		{[]string{"1", "1", "1", "1", "1", "X"}, false, `invalid value in field 'day-of-week' given: 'X'`},
		{[]string{"1", "1", "1", "1", "1", "1", "X"}, false, `invalid value in field 'year' given: 'X'`},
		{[]string{"1", "1", "1", "?", "1", "?", "*"}, false, `the cronjob will never run; both DoM and DoW contain the special character '?'`},
		{[]string{"1", "1", "1", "1", "1", "1", "*"}, true, ``},
	} {
		fields, err := createFields(tc.parts)
		if tc.fields && fields == nil {
			t.Errorf("'%#v': expected 'fields', got '%#v'", tc.parts, fields)
		} else if !tc.fields && fields != nil {
			t.Errorf("'%#v': expected '%#v', got '%#v'", tc.parts, nil, fields)
		}

		if tc.err != "" || err != nil {
			if eerr := fmt.Sprintf("%s", err); tc.err != eerr {
				t.Errorf("'%#v': expected '%s', got '%s'", tc.parts, tc.err, eerr)
			}
		}
	}
}

func TestFields_createField(t *testing.T) {
	type testCase struct {
		expr  string
		ft    fieldType
		field bool
		err   string
	}

	for _, tc := range []testCase{
		{",", typeMinutes, false, `invalid expression in field 'minutes' given: ','`},
		{"?,1", typeDoM, false, `invalid expression in field 'day-of-month' given: '?,1'`},
		{"LW", typeDoW, false, `the special character 'W' is only allowed in the DoM field`},
		{"32", typeDoM, false, `invalid value in field 'day-of-month' given: '32'`},
		{"LW", typeDoM, true, ``},
		{"1", typeDoM, true, ``},
	} {
		field, err := createField(tc.expr, tc.ft)
		if tc.field && field == nil {
			t.Errorf("'%#v': expected 'fields', got '%#v'", tc.expr, field)
		} else if !tc.field && field != nil {
			t.Errorf("'%#v': expected '%#v', got '%#v'", tc.expr, nil, field)
		}

		if tc.err != "" || err != nil {
			if eerr := fmt.Sprintf("%s", err); tc.err != eerr {
				t.Errorf("'%#v': expected '%s', got '%s'", tc.expr, tc.err, eerr)
			}
		}
	}
}

func TestFields_mergeCombinations(t *testing.T) {
	field, err := createField("1", typeDoM)
	if err != nil {
		t.Fatalf("%#v", err)
	}
	if len(field.combinations) != 1 {
		t.Fatalf("expected 2 combinations, got %d", len(field.combinations))
	}
	if !reflect.DeepEqual(field.combinations[0].values, []int{1}) {
		t.Errorf("expected '%#v', got '%#v'", []int{1}, field.combinations[0].values)
	}

	field, err = createField("L,1,L,5", typeDoM)
	if err != nil {
		t.Fatalf("%#v", err)
	}
	if len(field.combinations) != 2 {
		t.Fatalf("expected 2 combinations, got %d", len(field.combinations))
	}
	if field.combinations[0].values != nil {
		t.Errorf("expected '%#v', got '%#v'", nil, field.combinations[0].values)
	}
	if field.combinations[0].unit != "L" {
		t.Errorf("expected '%#v', got '%#v'", "L", field.combinations[0].unit)
	}
	if !reflect.DeepEqual(field.combinations[1].values, []int{1, 5}) {
		t.Errorf("expected '%#v', got '%#v'", []int{1, 5}, field.combinations[1].values)
	}
	if field.combinations[1].unit != "" {
		t.Errorf("expected '%#v', got '%#v'", "", field.combinations[1].unit)
	}
}
