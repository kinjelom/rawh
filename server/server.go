package server

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net"
	"rawh/common"
	"strings"
	"time"
)

type Server struct {
	port    int
	verbose bool
}

func NewServer(port int, verbose bool) (s *Server) {
	return &Server{port: port, verbose: verbose}
}

type RequestData struct {
	startLine     string
	method        string
	requestURI    string
	httpVersion   string
	headers       *common.HttpHeaders
	bodySize      int
	bodyHash      string
	readDuration  time.Duration
	sleepDuration time.Duration
	error         error
	contentLength int
}

func NewRequestData(normalizeHeaders bool) *RequestData {
	return &RequestData{
		headers:       common.NewHttpHeaders(normalizeHeaders),
		sleepDuration: time.Duration(0),
		contentLength: 0,
	}
}

func (r *RequestData) setStartLine(line string) {
	r.startLine = line
	parts := strings.Split(r.startLine, " ")
	if len(parts) == 3 {
		r.method, r.requestURI, r.httpVersion = parts[0], parts[1], parts[2]
		r.sleepDuration = common.ExtractSleepDurationFromQuery(r.requestURI)
	}
}

func writeln(w io.Writer, message string, verbose bool) {
	if verbose {
		log.Printf("> %s\n", message)
	}
	_, err := fmt.Fprintln(w, message)
	if err != nil {
		log.Printf("print error: %v", err)
	}
}

func (s *Server) PrintPlainTextResponse(w io.Writer, reqData *RequestData, verbose bool) {
	writeln(w, "HTTP/1.1 200 OK", verbose)
	writeln(w, "Content-Type: text/plain", verbose)
	for key, values := range reqData.headers.ResponseHeadersData {
		for _, value := range values {
			writeln(w, key+": "+value, verbose)
		}
	}
	writeln(w, "", verbose)
	writeln(w, "request-start-line: "+reqData.startLine, verbose)
	writeln(w, "request-header-lines:", verbose)
	for key, values := range reqData.headers.HeadersData {
		for _, value := range values {
			writeln(w, "- "+key+": "+value, verbose)
		}
	}
	writeln(w, "request-body-size: "+common.PrittyByteSize(reqData.bodySize), verbose)
	writeln(w, "request-body-hash: "+reqData.bodyHash, verbose)
	writeln(w, "request-read-duration: "+reqData.readDuration.String(), verbose)
	writeln(w, "request-sleep-duration: "+reqData.sleepDuration.String(), verbose)
}

func (s *Server) ReadRequestData(reader *bufio.Reader) *RequestData {
	log.Printf("Read request: start\n")
	readStart := time.Now().UnixMilli()
	reqData := NewRequestData(false)
	if requestLine, err := reader.ReadString('\n'); err == nil {
		reqData.setStartLine(strings.TrimSpace(requestLine))
		log.Printf("< start-line: %s\n", reqData.startLine)
		// header
		for {
			line, err := reader.ReadString('\n')
			if err != nil || line == "\r\n" {
				break
			}
			line = strings.TrimSpace(line)
			log.Printf("< header-line: %s\n", line)
			_ = reqData.headers.AddLine(line)
		}
		// get data found in headers
		if reqData.headers.SleepDuration > 0 {
			reqData.sleepDuration = reqData.headers.SleepDuration
		}
		if reqData.headers.ContentLength > 0 {
			reqData.contentLength = reqData.headers.ContentLength
		}
		// body
		if reqData.method != "GET" && reqData.method != "HEAD" && reqData.contentLength > 0 {
			log.Printf("< *** start of body reading ***\n")
			body := make([]byte, reqData.contentLength)
			_, err := io.ReadFull(reader, body)
			if err == nil {
				reqData.bodySize = len(body)
				hashAlg := md5.New()
				hashAlg.Write(body)
				reqData.bodyHash = fmt.Sprintf("MD5:%x", hashAlg.Sum(nil))
				log.Printf("< body-size: %d\n", reqData.bodySize)
				log.Printf("< body-hash: %s\n", reqData.bodyHash)
			}
			log.Printf("< *** end of body reading ***\n")
		} else {
			log.Printf("< *** body reading skipped ***\n")
		}
	}
	reqData.readDuration = time.Duration(time.Now().UnixMilli()-readStart) * time.Millisecond
	log.Printf("Read request: done [%s]", reqData.readDuration.String())
	return reqData
}

func (s *Server) handleTcpConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Printf("Error closing connection: %v", err)
		}
	}(conn)
	reqData := s.ReadRequestData(bufio.NewReader(conn))
	if reqData.sleepDuration.Milliseconds() > 0 {
		readStart := time.Now().UnixMilli()
		log.Printf("Going to sleep for %s", reqData.sleepDuration.String())
		time.Sleep(reqData.sleepDuration)
		actualSleepDuration := time.Duration(time.Now().UnixMilli()-readStart) * time.Millisecond
		log.Printf("Woke up after %s", actualSleepDuration.String())
	}
	s.PrintPlainTextResponse(conn, reqData, s.verbose)
}

func (s *Server) Serve() error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("error setting up TCP server: %v\n", err)
	}
	defer func(ln net.Listener) {
		err := ln.Close()
		if err != nil {
			fmt.Printf("Error closing TCP listener: %v\n", err)
		}
	}(ln)

	fmt.Printf("TCP Server is running on :%d\n", s.port)
	for {
		conn, err := ln.Accept()
		if err != nil {
			return fmt.Errorf("error accepting connection: %v\n", err)
		}
		go s.handleTcpConnection(conn)
	}
}
