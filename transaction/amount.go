package transaction

import (
	"fmt"
	"strconv"
	"strings"
)

// Amount represents a monetary amount stored as integer cents
// to avoid floating-point precision issues.
type Amount struct {
	cents int64
}

// ParseAmount parses a decimal string like "42.50" or "-13.99" into an Amount.
func ParseAmount(s string) (Amount, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Amount{}, nil
	}

	negative := false
	if s[0] == '-' {
		negative = true
		s = s[1:]
	}

	var dollars, cents int64
	if i := strings.IndexByte(s, '.'); i >= 0 {
		if i > 0 {
			d, err := strconv.ParseInt(s[:i], 10, 64)
			if err != nil {
				return Amount{}, fmt.Errorf("parse amount dollars %q: %w", s, err)
			}
			dollars = d
		}
		frac := s[i+1:]
		switch len(frac) {
		case 0:
			// trailing dot, e.g. "42."
		case 1:
			c, err := strconv.ParseInt(frac, 10, 64)
			if err != nil {
				return Amount{}, fmt.Errorf("parse amount cents %q: %w", s, err)
			}
			cents = c * 10
		case 2:
			c, err := strconv.ParseInt(frac, 10, 64)
			if err != nil {
				return Amount{}, fmt.Errorf("parse amount cents %q: %w", s, err)
			}
			cents = c
		default:
			return Amount{}, fmt.Errorf("parse amount %q: more than 2 decimal places", s)
		}
	} else {
		d, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return Amount{}, fmt.Errorf("parse amount %q: %w", s, err)
		}
		dollars = d
	}

	total := dollars*100 + cents
	if negative {
		total = -total
	}
	return Amount{cents: total}, nil
}

// Negate returns the amount with the opposite sign.
func (a Amount) Negate() Amount {
	return Amount{cents: -a.cents}
}

// String formats the amount as a decimal string, e.g. "42.50" or "-13.99".
func (a Amount) String() string {
	if a.cents < 0 {
		return "-" + Amount{cents: -a.cents}.String()
	}
	return fmt.Sprintf("%d.%02d", a.cents/100, a.cents%100)
}

// MarshalJSON outputs the amount as a JSON number, e.g. 42.50.
func (a Amount) MarshalJSON() ([]byte, error) {
	return []byte(a.String()), nil
}

// UnmarshalJSON parses a JSON number into an Amount.
func (a *Amount) UnmarshalJSON(data []byte) error {
	parsed, err := ParseAmount(string(data))
	if err != nil {
		return err
	}
	*a = parsed
	return nil
}
