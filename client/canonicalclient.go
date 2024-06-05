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

type CanonicalClient struct {
	verbose     bool
	httpVersion common.HttpVersion
	httpClient  *http.Client
	tlsConfig   *tls.Config
}

func NewCanonicalClient(tlsVersionName string, insecure bool, httpVersionName string, verbose bool) (Client, error) {
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
	return &CanonicalClient{
		verbose:     verbose,
		tlsConfig:   tlsConfig,
		httpVersion: httpVersion,
		httpClient:  &http.Client{Transport: transport},
	}, nil
}

func (c *CanonicalClient) DoRequest(method string, url string, customHeaders common.MultiString, data string) error {
	req, err := http.NewRequest(method, url, strings.NewReader(data))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	req.Proto, req.ProtoMajor, req.ProtoMinor = c.httpVersion.Proto, c.httpVersion.Major, c.httpVersion.Minor
	var reqHeaders = common.NewHttpHeaders(true)
	for _, headerLine := range customHeaders {
		key, val, err := common.SplitHeaderLine(headerLine)
		if err == nil {
			req.Header.Add(key, val)
		}
	}
	req.Header.Add(common.ContentLengthHeaderName, fmt.Sprint(len(data)))
	req.Host = reqHeaders.Host
	c.verboseRequest(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer common.SafeClose(resp.Body)
	c.verboseResponse(resp)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %v", err)
	}
	fmt.Println(string(body))
	return nil
}

func (c *CanonicalClient) verboseResponse(resp *http.Response) {
	if c.verbose {
		fmt.Printf("< HTTP/%d.%d %s\n", resp.ProtoMajor, resp.ProtoMinor, resp.Status)
		if len(resp.Header) > 0 {
			for k, va := range resp.Header {
				for _, v := range va {
					fmt.Printf("< %s: %s\n", k, v)
				}
			}
		}
		fmt.Printf("<\n")
	}
}

func (c *CanonicalClient) verboseRequest(req *http.Request) {
	if c.verbose {
		fmt.Printf("> %s %s %s\n", req.Method, req.URL.RequestURI(), req.Proto)
		fmt.Printf("> req.Host: %s\n", req.Host)
		if len(req.Header) > 0 {
			for k, va := range req.Header {
				for _, v := range va {
					fmt.Printf("> %s: %s\n", k, v)
				}
			}
		}
		fmt.Printf(">\n")
	}
}
