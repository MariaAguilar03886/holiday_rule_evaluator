package holiday_rule_evaluator

import (
	"errors"
	"testing"
	"time"
)

// Easter reference values sourced from public astronomical references.
var easterCases = []struct {
	year int
	want time.Time
}{
	{2023, date(2023, 4, 9)},
	{2024, date(2024, 3, 31)},
	{2025, date(2025, 4, 20)},
	{2026, date(2026, 4, 5)},
	{2000, date(2000, 4, 23)}, // Y2K boundary
	{1900, date(1900, 4, 15)}, // earliest supported year
	{9999, date(9999, 3, 28)}, // latest supported year
}

func TestEaster(t *testing.T) {
	for _, c := range easterCases {
		got, err := Easter(c.year)
		if err != nil {
			t.Errorf("Easter(%d): unexpected error: %v", c.year, err)
			continue
		}
		if !got.Equal(c.want) {
			t.Errorf("Easter(%d) = %s, want %s", c.year, got, c.want)
		}
	}
}

func TestEasterOutOfRange(t *testing.T) {
	for _, y := range []int{1899, 10000, -1} {
		if _, err := Easter(y); !errors.Is(err, ErrInvalidYear) {
			t.Errorf("Easter(%d): expected ErrInvalidYear, got %v", y, err)
		}
	}
}

func TestFixedValid(t *testing.T) {
	got, err := Fixed(2024, 7, 4)
	if err != nil {
		t.Fatalf("Fixed(2024, 7, 4): unexpected error: %v", err)
	}
	if !got.Equal(date(2024, 7, 4)) {
		t.Errorf("Fixed(2024, 7, 4) = %s, want 2024-07-04", got)
	}
}

func TestFixedInvalidMonth(t *testing.T) {
	for _, m := range []int{0, 13, -1} {
		if _, err := Fixed(2024, m, 15); !errors.Is(err, ErrInvalidMonth) {
			t.Errorf("Fixed(_, %d, _): expected ErrInvalidMonth, got %v", m, err)
		}
	}
}

func TestFixedInvalidDay(t *testing.T) {
	cases := []struct {
		year, month, day int
	}{
		{2024, 2, 30}, // Feb 30 never exists
		{2023, 2, 29}, // 2023 is not a leap year
		{2024, 2, 0},  // day 0
		{2024, 2, -1}, // negative day
		{2024, 13, 1}, // also catches bad month path indirectly
	}
	for _, c := range cases {
		_, err := Fixed(c.year, c.month, c.day)
		if err == nil {
			t.Errorf("Fixed(%d, %d, %d): expected error, got nil", c.year, c.month, c.day)
		}
	}
}

func TestFixedLeapDay(t *testing.T) {
	got, err := Fixed(2024, 2, 29)
	if err != nil {
		t.Fatalf("Fixed(2024, 2, 29): unexpected error: %v", err)
	}
	if !got.Equal(date(2024, 2, 29)) {
		t.Errorf("Fixed(2024, 2, 29) = %s, want 2024-02-29", got)
	}
}

func TestNthWeekdayValid(t *testing.T) {
	cases := []struct {
		name                          string
		year, month, weekday, nth     int
		want                          time.Time
	}{
		{"MLK Day 2024", 2024, 1, Monday, 3, date(2024, 1, 15)},
		{"Labor Day 2024", 2024, 9, Monday, 1, date(2024, 9, 2)},
		{"Thanksgiving 2024", 2024, 11, Thursday, 4, date(2024, 11, 28)},
		{"1st Sunday Jan 2024", 2024, 1, Sunday, 1, date(2024, 1, 7)},
		{"5th Wednesday Oct 2024", 2024, 10, Wednesday, 5, date(2024, 10, 30)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := NthWeekday(c.year, c.month, c.weekday, c.nth)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equal(c.want) {
				t.Errorf("NthWeekday(%d, %d, %d, %d) = %s, want %s",
					c.year, c.month, c.weekday, c.nth, got, c.want)
			}
		})
	}
}

func TestNthWeekdayNonexistent(t *testing.T) {
	// February 2024 has 29 days. The 5th Thursday would be Feb 29 if Feb 1
	// were a Thursday, but Feb 1 2024 is a Thursday, so Feb 29 is a Thursday.
	// We pick a combination that genuinely does not exist: 5th Wednesday of Feb 2024.
	_, err := NthWeekday(2024, 2, Wednesday, 5)
	if !errors.Is(err, ErrNoSuchWeekdayOccur) {
		t.Errorf("NthWeekday(2024, 2, Wednesday, 5): expected ErrNoSuchWeekdayOccur, got %v", err)
	}
}

func TestNthWeekdayInvalidInputs(t *testing.T) {
	cases := []struct {
		year, month, weekday, nth int
		wantErr                   error
	}{
		{1899, 1, Monday, 1, ErrInvalidYear},
		{10000, 1, Monday, 1, ErrInvalidYear},
		{2024, 0, Monday, 1, ErrInvalidMonth},
		{2024, 13, Monday, 1, ErrInvalidMonth},
		{2024, 1, 7, 1, ErrInvalidWeekday},
		{2024, 1, -1, 1, ErrInvalidWeekday},
		{2024, 1, Monday, 0, ErrInvalidOccurrence},
		{2024, 1, Monday, 6, ErrInvalidOccurrence},
	}
	for _, c := range cases {
		_, err := NthWeekday(c.year, c.month, c.weekday, c.nth)
		if !errors.Is(err, c.wantErr) {
			t.Errorf("NthWeekday(%d, %d, %d, %d): expected %v, got %v",
				c.year, c.month, c.weekday, c.nth, c.wantErr, err)
		}
	}
}

func TestLastWeekday(t *testing.T) {
	cases := []struct {
		name                   string
		year, month, weekday   int
		want                   time.Time
	}{
		{"Memorial Day 2024", 2024, 5, Monday, date(2024, 5, 27)},
		{"Last Sunday Feb 2024", 2024, 2, Sunday, date(2024, 2, 25)},
		{"Last Friday Dec 2024", 2024, 12, Friday, date(2024, 12, 27)},
		{"Last day is the weekday", 2024, 10, Thursday, date(2024, 10, 31)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := LastWeekday(c.year, c.month, c.weekday)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equal(c.want) {
				t.Errorf("LastWeekday(%d, %d, %d) = %s, want %s",
					c.year, c.month, c.weekday, got, c.want)
			}
		})
	}
}

func TestLastWeekdayInvalidInputs(t *testing.T) {
	cases := []struct {
		year, month, weekday int
		wantErr              error
	}{
		{1899, 1, Monday, ErrInvalidYear},
		{2024, 0, Monday, ErrInvalidMonth},
		{2024, 1, 7, ErrInvalidWeekday},
		{2024, 1, -1, ErrInvalidWeekday},
	}
	for _, c := range cases {
		_, err := LastWeekday(c.year, c.month, c.weekday)
		if !errors.Is(err, c.wantErr) {
			t.Errorf("LastWeekday(%d, %d, %d): expected %v, got %v",
				c.year, c.month, c.weekday, c.wantErr, err)
		}
	}
}

func TestIsLeap(t *testing.T) {
	cases := []struct {
		year int
		want bool
	}{
		{2000, true},
		{1900, false},
		{2024, true},
		{2023, false},
		{2100, false},
	}
	for _, c := range cases {
		if got := isLeap(c.year); got != c.want {
			t.Errorf("isLeap(%d) = %v, want %v", c.year, got, c.want)
		}
	}
}

func TestDaysInMonth(t *testing.T) {
	cases := []struct {
		year, month, want int
	}{
		{2024, 1, 31},
		{2024, 2, 29},
		{2023, 2, 28},
		{2024, 4, 30},
		{2000, 2, 29},
		{1900, 2, 28},
	}
	for _, c := range cases {
		if got := daysInMonth(c.year, c.month); got != c.want {
			t.Errorf("daysInMonth(%d, %d) = %d, want %d", c.year, c.month, got, c.want)
		}
	}
}

func date(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
