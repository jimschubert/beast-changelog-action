package main

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/sirupsen/logrus"
)

type customFormatter struct {
	// Set to true to enable timestamps
	EnableTimestamp bool
}

func (f *customFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	if f.EnableTimestamp {
		b.WriteString(entry.Time.Format("2006-01-02T15:04:05-07:00"))
		b.WriteString(" ")
	}

	// Add red X for error and above
	if entry.Level <= logrus.ErrorLevel {
		b.WriteString("❌ ")
	}
	b.WriteString(entry.Message)

	if len(entry.Data) > 0 {
		keys := make([]string, 0, len(entry.Data))
		for k := range entry.Data {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			b.WriteString(" ")
			b.WriteString(k)
			b.WriteString("=")
			b.WriteString(formatValue(entry.Data[k]))
		}
	}

	b.WriteString("\n")
	return b.Bytes(), nil
}

func formatValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		// No quotes for strings
		return v
	case error:
		return v.Error()
	default:
		return fmt.Sprintf("%v", v)
	}
}
