package main

import (
	"errors"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestCustomFormatter_Format(t *testing.T) {
	fixedZone := time.FixedZone("UTC-8", -8*60*60)
	fixedTime := time.Date(2024, 1, 2, 3, 4, 5, 0, fixedZone)

	cases := []struct {
		name            string
		enableTimestamp bool
		level           logrus.Level
		message         string
		data            logrus.Fields
		expected        string
	}{
		{
			name:            "info without timestamp",
			enableTimestamp: false,
			level:           logrus.InfoLevel,
			message:         "hello",
			data:            logrus.Fields{"b": 1, "a": "value"},
			expected:        "hello a=value b=1\n",
		},
		{
			name:            "error adds emoji",
			enableTimestamp: false,
			level:           logrus.ErrorLevel,
			message:         "failure",
			data:            nil,
			expected:        "❌ failure\n",
		},
		{
			name:            "timestamp and sorted fields",
			enableTimestamp: true,
			level:           logrus.WarnLevel,
			message:         "warn",
			data: logrus.Fields{
				"err": errors.New("boom"),
				"b":   2,
				"a":   "first",
			},
			expected: "2024-01-02T03:04:05-08:00 warn a=first b=2 err=boom\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			formatter := &customFormatter{EnableTimestamp: testCase.enableTimestamp}
			logger := logrus.New()
			entry := logrus.NewEntry(logger)
			entry.Time = fixedTime
			entry.Level = testCase.level
			entry.Message = testCase.message
			entry.Data = testCase.data

			formatted, err := formatter.Format(entry)
			if err != nil {
				t.Fatalf("Format returned error: %v", err)
			}

			if string(formatted) != testCase.expected {
				t.Fatalf("unexpected output:\nexpected: %q\nactual:   %q", testCase.expected, string(formatted))
			}
		})
	}
}

