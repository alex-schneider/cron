// Copyright 2021 Alex Schneider. All rights reserved.

package cron

// Parse parses the given expression spec and returns a new <cron.Schedule> instance.
func Parse(expression string) (*Schedule, error) {
	fields, err := getFields(expression)
	if err != nil {
		return nil, err
	}

	return &Schedule{fields: fields}, nil
}
