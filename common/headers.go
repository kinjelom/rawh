package common

import (
	"fmt"
	"log"
	"net/textproto"
	"strconv"
	"strings"
	"time"
)

const SleepDurationHeaderName = "rawh-Sleep-Duration"
const ContentLengthHeaderName = "Content-Length"
const ReturnEchoHeaderNamePrefix = "rawh-Echo-"

type HttpHeaders struct {
	normalizeHeaders    bool
	HeadersData         map[string][]string
	ResponseHeadersData map[string][]string
	Host                string
	SleepDuration       time.Duration
	ContentLength       int
}

func NewHttpHeaders(normalizeHeaders bool) *HttpHeaders {
	return &HttpHeaders{
		normalizeHeaders:    normalizeHeaders,
		HeadersData:         make(map[string][]string),
		ResponseHeadersData: make(map[string][]string),
		SleepDuration:       0,
		ContentLength:       0,
	}
}

func (h *HttpHeaders) Add(key, value string) {
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	if h.normalizeHeaders {
		key = textproto.CanonicalMIMEHeaderKey(key)
	}
	h.HeadersData[key] = append(h.HeadersData[key], value)
	lk := strings.ToLower(key)
	if lk == "host" && value != "" {
		h.Host = value
	}
	if strings.HasPrefix(lk, strings.ToLower(ReturnEchoHeaderNamePrefix)) {
		h.ResponseHeadersData[key] = append(h.ResponseHeadersData[key], value)
	}

	if strings.ToLower(SleepDurationHeaderName) == lk {
		duration, err := time.ParseDuration(value)
		if err == nil {
			h.SleepDuration = duration
		} else {
			log.Println(fmt.Errorf("wrong header '%s' value '%s': %v", SleepDurationHeaderName, value, err))
		}
	}

	if strings.ToLower(ContentLengthHeaderName) == lk {
		cl, err := strconv.Atoi(value)
		if err == nil {
			h.ContentLength = cl
		} else {
			log.Println(fmt.Errorf("wrong header '%s' value '%s': %v", ContentLengthHeaderName, value, err))
		}
	}
}
func (h *HttpHeaders) AddLine(headerLine string) error {
	parts := strings.SplitN(headerLine, ":", 2)
	if len(parts) >= 2 {
		h.Add(parts[0], parts[1])
		return nil
	} else {
		return fmt.Errorf("invalid header line: %s", headerLine)
	}
}
