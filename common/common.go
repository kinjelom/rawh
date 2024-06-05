package common

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"
	"time"
)

const SleepDurationQueryParamName = "rawh-sleep-duration"

var tlsVersions = map[string]uint16{
	"1.0": tls.VersionTLS10,
	"1.1": tls.VersionTLS11,
	"1.2": tls.VersionTLS12,
	"1.3": tls.VersionTLS13,
}

func ParseTlsVersionName(versionName string) (uint16, error) {
	if version, ok := tlsVersions[versionName]; ok {
		return version, nil
	}
	return 0, fmt.Errorf("unsupported TLS version: %s", versionName)
}

type HttpVersion struct {
	Proto string
	Major int
	Minor int
}

var httpVersions = map[string]HttpVersion{
	"1.0": {"HTTP/1.0", 1, 0},
	"1.1": {"HTTP/1.1", 1, 1},
	"2.0": {"HTTP/2.0", 2, 0},
	"2":   {"HTTP/2.0", 2, 0},
}

func ParseHttpVersionName(versionName string) (HttpVersion, error) {
	if version, ok := httpVersions[versionName]; ok {
		return HttpVersion{version.Proto, version.Major, version.Minor}, nil
	}
	return HttpVersion{fmt.Sprintf("HTTP/%s", versionName), 0, 0}, fmt.Errorf("unsupported HTTP version: %s", versionName)
}

func ExtractSleepDurationFromQuery(requestURI string) time.Duration {
	u, err := url.Parse(requestURI)
	if err != nil {
		log.Printf("Error parsing URL: %v", err)
		return 0
	}
	values, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		log.Printf("Error parsing query: %v", err)
		return 0
	}
	sleepDurationStr := values.Get(SleepDurationQueryParamName)
	if sleepDurationStr == "" {
		return 0
	}
	sleepDuration, err := time.ParseDuration(sleepDurationStr)
	if err != nil {
		log.Printf("Invalid %s format: %v", SleepDurationQueryParamName, err)
		return 0
	}
	return sleepDuration
}

type MultiString []string

func (m *MultiString) String() string {
	return strings.Join(*m, ", ")
}
func (m *MultiString) Set(value string) error {
	*m = append(*m, value)
	return nil
}

func GenerateSampleDataString(length int) string {
	pattern := "1234567890"
	patternLength := len(pattern)
	if length <= 0 {
		return ""
	}
	repeatCount := length / patternLength
	if length%patternLength != 0 {
		repeatCount++
	}
	repeatedString := strings.Repeat(pattern, repeatCount)
	return repeatedString[:length]
}

func SafeClose(closable io.ReadCloser) {
	err := closable.Close()
	if err != nil {
		log.Println("Error closing: ", err)
	}
}
