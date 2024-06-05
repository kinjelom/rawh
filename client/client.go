package client

import (
	"crypto/tls"
	"fmt"
	"golang.org/x/net/http2"
	"io"
	"net/http"
	"rawh/common"
	"strings"
)

type Client struct {
	verbose          bool
	normalizeHeaders bool
	tlsVersion       uint16
	httpVersion      common.HttpVersion
	httpClient       *http.Client
}

func NewClient(normalizeHeaders bool, tlsVersionName string, insecure bool, httpVersionName string, verbose bool) (*Client, error) {
	tlsVer, err := common.ParseTlsVersionName(tlsVersionName)
	if err != nil {
		return &Client{}, err
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
	return &Client{
		verbose:          verbose,
		normalizeHeaders: normalizeHeaders,
		tlsVersion:       tlsVer,
		httpVersion:      httpVersion,
		httpClient:       &http.Client{Transport: transport},
	}, nil
}

func (c *Client) DoRequest(method string, url string, customHeaders common.MultiString, data string) error {
	req, err := http.NewRequest(method, url, strings.NewReader(data))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	req.Proto, req.ProtoMajor, req.ProtoMinor = c.httpVersion.Proto, c.httpVersion.Major, c.httpVersion.Minor
	var reqHeaders = common.NewHttpHeaders(c.normalizeHeaders)
	for _, headerLine := range customHeaders {
		_ = reqHeaders.AddLine(headerLine)
	}
	reqHeaders.Add(common.ContentLengthHeaderName, fmt.Sprint(len(data)))
	req.Header = reqHeaders.HeadersData
	req.Host = reqHeaders.Host
	c.verboseRequest(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}(resp.Body)
	c.verboseResponse(resp)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %v", err)
	}

	fmt.Println("\n" + string(body))
	return nil
}

func (c *Client) verboseResponse(resp *http.Response) {
	if c.verbose {
		fmt.Printf("< = RESPONSE:\n")
		fmt.Printf("< HTTP/%d.%d %s\n", resp.ProtoMajor, resp.ProtoMinor, resp.Status)
		if len(resp.Header) > 0 {
			fmt.Printf("< == RESPONSE HEADERS:\n")
			for k, va := range resp.Header {
				for _, v := range va {
					fmt.Printf("< - %s: %s\n", k, v)
				}
			}
		}
	}
}

func (c *Client) verboseRequest(req *http.Request) {
	if c.verbose {
		fmt.Printf("\n> = REQUEST:\n")
		fmt.Printf("> %s %s %s\n", req.Method, req.URL.RequestURI(), req.Proto)
		fmt.Printf("> req.Host: %s\n", req.Host)
		if len(req.Header) > 0 {
			fmt.Printf("> == REQUEST HEADERS:\n")
			for k, va := range req.Header {
				for _, v := range va {
					fmt.Printf("> - %s: %s\n", k, v)
				}
			}
		}
	}
}
