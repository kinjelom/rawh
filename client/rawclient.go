package client

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"golang.org/x/net/http2"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"rawh/common"
	"strings"
)

type RawClient struct {
	verbose          bool
	normalizeHeaders bool
	httpVersion      common.HttpVersion
	httpClient       *http.Client
	tlsConfig        *tls.Config
}

func NewRawClient(normalizeHeaders bool, tlsVersionName string, insecure bool, httpVersionName string, verbose bool) (Client, error) {
	tlsVer, err := common.ParseTlsVersionName(tlsVersionName)
	if err != nil {
		return nil, err
	}
	var tlsConfig = &tls.Config{
		MinVersion:         tlsVer,
		InsecureSkipVerify: insecure,
	}
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	httpVersion, err := common.ParseHttpVersionName(httpVersionName)
	if err != nil {
		return nil, fmt.Errorf("error parsing HTTP version: %v", err)
	}
	if httpVersion.Major == 2 {
		err := http2.ConfigureTransport(transport)
		if err != nil {
			return nil, fmt.Errorf("error configure HTTP 2 transport: %v", err)
		}
	}
	return &RawClient{
		verbose:          verbose,
		normalizeHeaders: normalizeHeaders,
		tlsConfig:        tlsConfig,
		httpVersion:      httpVersion,
		httpClient:       &http.Client{Transport: transport},
	}, nil
}

func (c *RawClient) reqPrintln(w io.Writer, line string) {
	line = strings.TrimSpace(line)
	if c.verbose {
		log.Printf("> %s\n", line)
	}
	_, err := fmt.Fprint(w, line+"\r\n")
	if err != nil {
		log.Printf("printing request line '%s' error: %v", line, err)
	}
}
func (c *RawClient) respVerbose(line string) {
	if c.verbose {
		line = strings.TrimSpace(line)
		log.Printf("< %s\n", line)
	}
}

func (c *RawClient) DoRequest(method string, urlString string, customHeaders common.MultiString, data string) error {
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return fmt.Errorf("error parsing URL: %v", err)
	}

	var conn net.Conn
	if parsedURL.Scheme == "https" {
		conn, err = tls.Dial("tcp", parsedURL.Host, c.tlsConfig)
		if err != nil {
			return fmt.Errorf("error establishing secure connection: %v", err)
		}
	} else {
		conn, err = net.Dial("tcp", parsedURL.Host)
		if err != nil {
			return fmt.Errorf("error establishing connection: %v", err)
		}
	}
	defer common.SafeClose(conn)

	// request:
	reqHeaders := common.NewHttpHeaders(c.normalizeHeaders)
	reqHeaders.Host = parsedURL.Host
	err = reqHeaders.AddLines(customHeaders)
	if err != nil {
		return fmt.Errorf("error adding custom headers: %v", err)
	}
	c.reqPrintln(conn, fmt.Sprintf("%s %s %s", method, parsedURL.RequestURI(), c.httpVersion.Proto))
	c.reqPrintln(conn, fmt.Sprintf("%s: %s", "Host", reqHeaders.Host))
	for key, values := range reqHeaders.HeadersData {
		if strings.ToLower(key) != "host" {
			for _, value := range values {
				c.reqPrintln(conn, fmt.Sprintf("%s: %s", key, value))
			}
		}
	}
	c.reqPrintln(conn, fmt.Sprintf("%s: %s", common.ContentLengthHeaderName, fmt.Sprint(len(data))))
	c.reqPrintln(conn, "")
	if data != "" {
		c.reqPrintln(conn, data)
	}

	// response:
	responseReader := bufio.NewReader(conn)
	responseStatus, err := responseReader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("error reading response status: %v", err)
	}
	c.respVerbose(responseStatus)
	// Reading headers directly without normalization
	for {
		line, err := responseReader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading headers: %v", err)
		}
		line = strings.TrimSpace(line)
		c.respVerbose(line)
		if line == "" {
			break // header section end
		}
	}
	body, err := io.ReadAll(responseReader)
	if err != nil {
		return fmt.Errorf("error reading response: %v", err)
	}
	fmt.Println(string(body))
	return nil
}
