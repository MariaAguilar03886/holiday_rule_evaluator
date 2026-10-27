# holiday_rule_evaluator

Computes the date of fixed and floating holidays — including Easter via the Anonymous Gregorian algorithm and the Nth weekday of a month — for a given year.

```go
package main

import (
    "fmt"
    "github.com/yourname/holiday_rule_evaluator"
)

func main() {
    easter, err := holiday_rule_evaluator.Easter(2024)
    if err != nil {
        panic(err)
    }
    fmt.Println(easter) // 2024-03-31

    mlk, err := holiday_rule_evaluator.NthWeekday(2024, 1, holiday_rule_evaluator.Monday, 3)
    if err != nil {
        panic(err)
    }
    fmt.Println(mlk) // 2024-01-15 — 3rd Monday of January

    fixed := holiday_rule_evaluator.Fixed(2024, 7, 4)
    fmt.Println(fixed) // 2024-07-04
}
```

## Why

Every calendar app re-implements Easter or the Nth-weekday-of-month calculation, usually slightly wrong. `Fixed` validates month/day and returns a `time.Time` so the caller does not have to remember how `time.Date` normalises out-of-range values. `NthWeekday` returns a clear error when the Nth weekday does not exist in the month — for example the 5th Tuesday of February 2025 — rather than silently rolling into March. `Easter` supports the Western/Gregorian computus only; Orthodox Easter uses the Julian calendar plus a correction and is out of scope.

## Edge

`NthWeekday` with `occurrence = 5` fails for most month/weekday combinations. If you need the last weekday of a month, use `LastWeekday` instead.

```go
last, err := holiday_rule_evaluator.LastWeekday(2024, 2, holiday_rule_evaluator.Sunday) // 2024-02-25
```
