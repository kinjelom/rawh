package common

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

var bUnits = map[string]float64{
	"B":  1,
	"KB": 1024,
	"MB": 1024 * 1024,
	"GB": 1024 * 1024 * 1024,
}

// PrittyByteSize converts byte size to a human-readable string.
func PrittyByteSize(bytes int) string {
	if bytes <= 0 {
		return strconv.Itoa(bytes) + " B"
	}
	i64Bytes := int64(bytes)
	var result float64
	var unit string

	for u, factor := range bUnits {
		if float64(i64Bytes) >= factor {
			result = float64(i64Bytes) / factor
			unit = u
			break // it will break at the highest unit it finds
		}
	}
	return fmt.Sprintf("%.2f %s", result, unit)
}

// ParsePrittyByteSize parses a formatted size string back to the number of bytes.
func ParsePrittyByteSize(s string) (int, error) {
	s = strings.TrimSpace(s)
	var numberStr, unit string
	// Split the string into numeric and unit parts
	for i, char := range s {
		if unicode.IsLetter(char) {
			numberStr = s[:i]
			unit = s[i:]
			break
		}
	}
	// If no unit is found, assume it is in bytes
	if unit == "" {
		numberStr = s
		unit = "B"
	}
	number, err := strconv.ParseFloat(numberStr, 64)
	if err != nil {
		return 0, err
	}
	// Convert string unit to upper to match keys in the map
	unit = strings.ToUpper(unit)
	if factor, ok := bUnits[unit]; ok {
		return int(number * factor), nil
	}
	return 0, fmt.Errorf("unknown unit %s", unit)
}
