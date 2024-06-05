package common

import (
	"fmt"
	"log"
	"net/textproto"
	"strconv"
	"strings"
	"time"
)

const SleepDurationHeaderName = "Rawh-Sleep-Duration"
const ContentLengthHeaderName = "Content-Length"
const EchoHeaderName = "Rawh-Echo"

type HttpHeaders struct {
	normalizeHeaders bool
	HeadersData      map[string][]string
	EchoHeadersData  map[string][]string
	Host             string
	SleepDuration    time.Duration
	ContentLength    int
}

func NewHttpHeaders(normalizeHeaders bool) *HttpHeaders {
	return &HttpHeaders{
		normalizeHeaders: normalizeHeaders,
		HeadersData:      make(map[string][]string),
		EchoHeadersData:  make(map[string][]string),
		SleepDuration:    0,
		ContentLength:    0,
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
	if lk == strings.ToLower(EchoHeaderName) {
		parts := strings.Split(value, " ")
		for _, echoHeader := range parts {
			h.EchoHeadersData[echoHeader] = append(h.EchoHeadersData[echoHeader], echoHeader)
		}
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
	key, val, err := SplitHeaderLine(headerLine)
	if err == nil {
		h.Add(key, val)
	}
	return err
}

func (h *HttpHeaders) AddLines(headers []string) error {
	for _, headerLine := range headers {
		err := h.AddLine(headerLine)
		if err != nil {
			return err
		}
	}
	return nil
}

func SplitHeaderLine(headerLine string) (key string, value string, err error) {
	parts := strings.SplitN(headerLine, ":", 2)
	if len(parts) >= 2 {
		return parts[0], parts[1], nil
	} else {
		return "", "", fmt.Errorf("invalid header line: %s", headerLine)
	}
}
