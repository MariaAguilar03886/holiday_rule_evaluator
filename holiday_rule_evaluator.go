package holiday_rule_evaluator

import (
	"errors"
	"fmt"
	"time"
)

// Common weekday constants, re-exported as this package's own integers so callers
// can pass a plain int (e.g. Monday) instead of importing time.
const (
	Sunday = int(time.Sunday)
	Monday = int(time.Monday)
	Tuesday = int(time.Tuesday)
	Wednesday = int(time.Wednesday)
	Thursday = int(time.Thursday)
	Friday = int(time.Friday)
	Saturday = int(time.Saturday)
)

var (
	ErrInvalidYear        = errors.New("year out of supported range")
	ErrInvalidMonth       = errors.New("month must be 1-12")
	ErrInvalidDay         = errors.New("day out of range for month")
	ErrInvalidWeekday     = errors.New("weekday must be 0-6 (Sunday-Saturday)")
	ErrInvalidOccurrence  = errors.New("occurrence must be 1-5")
	ErrNoSuchWeekdayOccur = errors.New("no such Nth weekday in this month")
)

// minYear/maxYear bound the supported range. They keep Easter's Gregorian
// algorithm inside the years where it is historically defined and let callers
// reason about the domain.
const (
	minYear = 1900
	maxYear = 9999
)

// Fixed returns the date of a fixed month/day holiday for the given year.
// It validates month and day explicitly rather than relying on time.Date's
// silent normalisation, which would turn July 32 into August 1 unnoticed.
func Fixed(year, month, day int) (time.Time, error) {
	if err := checkYear(year); err != nil {
		return time.Time{}, err
	}
	if month < 1 || month > 12 {
		return time.Time{}, fmt.Errorf("%w: got %d", ErrInvalidMonth, month)
	}
	maxDay := daysInMonth(year, month)
	if day < 1 || day > maxDay {
		return time.Time{}, fmt.Errorf("%w: got %d for month %d", ErrInvalidDay, day, month)
	}
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
}

// Easter computes the date of Western (Gregorian) Easter Sunday for the given
// year using the Anonymous Gregorian algorithm, which is valid for all years
// from 1583 onward. We cap the lower bound at 1900 because that is far enough
// in the past to cover practical calendar use while remaining inside the
// algorithm's domain.
func Easter(year int) (time.Time, error) {
	if err := checkYear(year); err != nil {
		return time.Time{}, err
	}
	a := year % 19
	b := year / 100
	c := year % 100
	d := b / 4
	e := b % 4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i := c / 4
	k := c % 4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451
	month := (h + l - 7*m + 114) / 31
	day := ((h + l - 7*m + 114) % 31) + 1
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
}

// NthWeekday returns the date of the Nth occurrence of weekday in the given
// month and year. occurrence must be 1-5. If the Nth weekday does not exist
// in that month (e.g. the 5th Tuesday of February), it returns an error
// instead of silently returning a date in the next month.
func NthWeekday(year, month, weekday, occurrence int) (time.Time, error) {
	if err := checkYear(year); err != nil {
		return time.Time{}, err
	}
	if month < 1 || month > 12 {
		return time.Time{}, fmt.Errorf("%w: got %d", ErrInvalidMonth, month)
	}
	if weekday < 0 || weekday > 6 {
		return time.Time{}, fmt.Errorf("%w: got %d", ErrInvalidWeekday, weekday)
	}
	if occurrence < 1 || occurrence > 5 {
		return time.Time{}, fmt.Errorf("%w: got %d", ErrInvalidOccurrence, occurrence)
	}

	firstOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	firstWeekday := int(firstOfMonth.Weekday())
	offset := (weekday - firstWeekday + 7) % 7
	day := offset + 1 + (occurrence-1)*7
	if day > daysInMonth(year, month) {
		return time.Time{}, fmt.Errorf("%w: %dth weekday %d in month %d", ErrNoSuchWeekdayOccur, occurrence, weekday, month)
	}
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
}

// LastWeekday returns the date of the last occurrence of weekday in the given
// month and year. Useful for holidays like Memorial Day (last Monday of May).
func LastWeekday(year, month, weekday int) (time.Time, error) {
	if err := checkYear(year); err != nil {
		return time.Time{}, err
	}
	if month < 1 || month > 12 {
		return time.Time{}, fmt.Errorf("%w: got %d", ErrInvalidMonth, month)
	}
	if weekday < 0 || weekday > 6 {
		return time.Time{}, fmt.Errorf("%w: got %d", ErrInvalidWeekday, weekday)
	}

	lastDay := daysInMonth(year, month)
	lastOfMonth := time.Date(year, time.Month(month), lastDay, 0, 0, 0, 0, time.UTC)
	lastWeekday := int(lastOfMonth.Weekday())
	offset := (lastWeekday - weekday + 7) % 7
	day := lastDay - offset
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
}

func checkYear(year int) error {
	if year < minYear || year > maxYear {
		return fmt.Errorf("%w: got %d, want %d-%d", ErrInvalidYear, year, minYear, maxYear)
	}
	return nil
}

// daysInMonth returns the number of days in the given month, accounting for
// leap years in February.
func daysInMonth(year, month int) int {
	switch month {
	case 1, 3, 5, 7, 8, 10, 12:
		return 31
	case 4, 6, 9, 11:
		return 30
	case 2:
		if isLeap(year) {
			return 29
		}
		return 28
	default:
		return 0
	}
}

// isLeap returns true if year is a Gregorian leap year. The standard rule is
// kept inline rather than calling time.Time.IsLeap because no such method
// exists in the standard library.
func isLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}
