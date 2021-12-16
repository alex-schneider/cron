// Copyright 2021 Alex Schneider. All rights reserved.

package cron

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

type combination struct {
	unit   string
	values []int
}

type field struct {
	expression   string
	combinations []*combination
}

type fields struct {
	seconds *field
	minutes *field
	hours   *field
	dom     *field
	month   *field
	dow     *field
	year    *field
	once    bool
}

var reFieldsMatcher = regexp.MustCompile(`\S+`)

/* ==================================================================================================== */

func getFields(expression string) (*fields, error) {
	expression = strings.TrimSpace(expression)

	e, err := expressionFromMacro(expression)
	if err != nil {
		return nil, err
	} else if e == "~" {
		return &fields{once: true}, nil
	} else if e != "" {
		expression = e
	}

	fieldsParts := reFieldsMatcher.FindAllString(expression, -1)
	fieldsCount := len(fieldsParts)

	if fieldsCount < 5 || fieldsCount > 7 {
		return nil, fmt.Errorf("invalid expression given '%s'", expression)
	} else if fieldsCount < 7 {
		if fieldsCount == 5 {
			fieldsParts = append([]string{"0"}, fieldsParts...)
		}

		fieldsParts = append(fieldsParts, "*")
	}

	return createFields(fieldsParts)
}

func createFields(fieldsParts []string) (*fields, error) {
	fields := &fields{}

	field, err := createField(fieldsParts[0], typeSeconds)
	if err != nil {
		return nil, err
	}
	fields.seconds = field

	field, err = createField(fieldsParts[1], typeMinutes)
	if err != nil {
		return nil, err
	}
	fields.minutes = field

	field, err = createField(fieldsParts[2], typeHours)
	if err != nil {
		return nil, err
	}
	fields.hours = field

	field, err = createField(fieldsParts[3], typeDoM)
	if err != nil {
		return nil, err
	}
	fields.dom = field

	field, err = createField(fieldsParts[4], typeMonth)
	if err != nil {
		return nil, err
	}
	fields.month = field

	field, err = createField(fieldsParts[5], typeDoW)
	if err != nil {
		return nil, err
	}
	fields.dow = field

	field, err = createField(fieldsParts[6], typeYear)
	if err != nil {
		return nil, err
	}
	fields.year = field

	if fields.dom.combinations[0].unit == "?" && fields.dow.combinations[0].unit == "?" {
		return nil, fmt.Errorf("the cronjob will never run; both DoM and DoW contain the special character '?'")
	}

	return fields, nil
}

func createField(expression string, ft fieldType) (*field, error) {
	field := &field{
		expression: expression,
	}

	for _, expr := range strings.Split(expression, ",") {
		if expr == "" || (strings.Contains(expression, "?") && expression != "?") {
			return nil, fmt.Errorf("invalid expression in field '%s' given: '%s'", ft, expression)
		}

		expr = strings.ToUpper(expr)

		if isFlexValue(expr) {
			values, unit, err := getFlexValues(expr, ft)
			if err != nil {
				return nil, err
			}

			field.combinations = append(field.combinations, &combination{values: values, unit: unit})
		} else {
			values, err := getFixValues(expr, ft)
			if err != nil {
				return nil, err
			}

			field.combinations = append(field.combinations, &combination{values: values})
		}
	}

	field.mergeCombinations()

	return field, nil
}

func (f *field) mergeCombinations() {
	if len(f.combinations) < 2 {
		return
	}

	var values []int
	var combinations []*combination

	for _, combi := range f.combinations {
		if combi.unit != "" {
			var exists bool

			for _, check := range combinations {
				if reflect.DeepEqual(combi, check) {
					exists = true
					break
				}
			}

			if !exists {
				combinations = append(combinations, combi)
			}
		} else {
			values = append(values, combi.values...)
		}
	}

	if len(values) > 0 {
		values = uniqueValues(values)

		sort.Ints(values)

		combinations = append(combinations, &combination{values: values})
	}

	f.combinations = combinations
}
