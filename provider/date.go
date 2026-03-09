package provider

import (
	"fmt"
	"time"
)

const _dateInputFormat = "01/02/2006"

func formatDate(s string) (string, error) {
	t, err := time.Parse(_dateInputFormat, s)
	if err != nil {
		return "", fmt.Errorf("parse date %q: %w", s, err)
	}
	return t.Format(time.DateOnly), nil
}
