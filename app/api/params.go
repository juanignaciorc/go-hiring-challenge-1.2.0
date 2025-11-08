package api

import (
	"strconv"
	"strings"
)

// Defaults and limits for pagination.
const (
	DefaultOffset = 0
	DefaultLimit  = 10
	MinLimit      = 1
	MaxLimit      = 100
)

// Centralized error messages for query parameter parsing.
const (
	errOffsetMustBeInt = "offset must be an integer"
	errLimitMustBeInt  = "limit must be an integer"
	errPriceMustBeNum  = "price_lt must be numeric"
	errPriceGteZero    = "price_lt must be greater than or equal to 0"
)

// ParseOffset parses the "offset" query parameter.
// - Empty input returns the DefaultOffset.
// - Non-integer input returns ok=false and a user-facing error message.
// - Negative values are clamped to 0.
func ParseOffset(raw string) (int, bool, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return DefaultOffset, true, ""
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false, errOffsetMustBeInt
	}
	if v < 0 {
		v = 0
	}
	return v, true, ""
}

// ParseLimit parses the "limit" query parameter.
// - Empty input returns the DefaultLimit.
// - Non-integer input returns ok=false and a user-facing error message.
// - The result is clamped into the inclusive range [MinLimit, MaxLimit].
func ParseLimit(raw string) (int, bool, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return DefaultLimit, true, ""
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false, errLimitMustBeInt
	}
	if v < MinLimit {
		v = MinLimit
	} else if v > MaxLimit {
		v = MaxLimit
	}
	return v, true, ""
}

// ParsePriceLT parses the "price_lt" query parameter.
// - Empty input returns nil to indicate "no filter".
// - Non-numeric input returns ok=false and a user-facing error message.
// - Values must be >= 0; otherwise ok=false is returned.
// - On success, returns a pointer to the parsed float64 to distinguish from "not provided".
func ParsePriceLT(raw string) (*float64, bool, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, true, ""
	}
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil, false, errPriceMustBeNum
	}
	if f < 0 {
		return nil, false, errPriceGteZero
	}
	return &f, true, ""
}

// Normalize trims surrounding spaces and lowercases the input to build case-insensitive filters.
func Normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
