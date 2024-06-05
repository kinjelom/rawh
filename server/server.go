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
		bodyHash:      "empty",
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
func (s *Server) logVerbose(line string) {
	line = strings.TrimSpace(line)
	if s.verbose {
		log.Printf("# %s\n", line)
	}
}
func (s *Server) reqVerbose(line string) {
	line = strings.TrimSpace(line)
	if s.verbose {
		log.Printf("< %s\n", line)
	}
}

func (s *Server) respPrintln(w io.Writer, line string) {
	line = strings.TrimSpace(line)
	if s.verbose {
		log.Printf("> %s\n", line)
	}
	_, err := fmt.Fprint(w, line+"\r\n")
	if err != nil {
		log.Printf("print error: %v", err)
	}
}

func (s *Server) PrintPlainTextResponse(w io.Writer, reqData *RequestData) {
	s.respPrintln(w, "HTTP/1.1 200 OK")
	s.respPrintln(w, "Content-Type: text/plain")
	for key, values := range reqData.headers.EchoHeadersData {
		for _, value := range values {
			s.respPrintln(w, key+": "+value)
		}
	}
	s.respPrintln(w, "")
	s.respPrintln(w, "request-start-line: "+reqData.startLine)
	s.respPrintln(w, "request-header-lines:")
	for key, values := range reqData.headers.HeadersData {
		for _, value := range values {
			s.respPrintln(w, "- "+key+": "+value)
		}
	}
	s.respPrintln(w, "request-body-size: "+common.PrittyByteSize(reqData.bodySize))
	s.respPrintln(w, "request-body-hash: "+reqData.bodyHash)
	s.respPrintln(w, "request-read-duration: "+reqData.readDuration.String())
	s.respPrintln(w, "request-sleep-duration: "+reqData.sleepDuration.String())
}

func (s *Server) ReadRequestData(reader *bufio.Reader) *RequestData {
	s.logVerbose("Read request: start")
	readStart := time.Now().UnixMilli()
	reqData := NewRequestData(false)
	if requestLine, err := reader.ReadString('\n'); err == nil {
		reqData.setStartLine(strings.TrimSpace(requestLine))
		s.reqVerbose(reqData.startLine)
		// header
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				log.Printf("read error: %v", err)
				break
			}
			s.reqVerbose(line)
			line = strings.TrimSpace(line)
			if line == "" {
				break // end of header
			}
			err = reqData.headers.AddLine(line)
			if err != nil {
				log.Printf("read header line '%s' error: %v", line, err)
			}
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
			s.logVerbose("Start of body reading")
			body := make([]byte, reqData.contentLength)
			_, err := io.ReadFull(reader, body)
			if err == nil {
				reqData.bodySize = len(body)
				hashAlg := md5.New()
				hashAlg.Write(body)
				reqData.bodyHash = fmt.Sprintf("MD5:%x", hashAlg.Sum(nil))
				s.logVerbose(fmt.Sprintf("body-size: %d", reqData.bodySize))
				s.logVerbose(fmt.Sprintf("body-hash: %s", reqData.bodyHash))
			}
			s.logVerbose("End of body reading")
		} else {
			s.logVerbose("Body reading skipped")
		}
	}
	reqData.readDuration = time.Duration(time.Now().UnixMilli()-readStart) * time.Millisecond
	s.logVerbose(fmt.Sprintf("Read request: done [%s]", reqData.readDuration.String()))
	return reqData
}

func (s *Server) handleTcpConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}(conn)
	reqData := s.ReadRequestData(bufio.NewReader(conn))
	if reqData.sleepDuration.Milliseconds() > 0 {
		readStart := time.Now().UnixMilli()
		s.logVerbose(fmt.Sprintf("Going to sleep for %s", reqData.sleepDuration.String()))
		time.Sleep(reqData.sleepDuration)
		actualSleepDuration := time.Duration(time.Now().UnixMilli()-readStart) * time.Millisecond
		s.logVerbose(fmt.Sprintf("Woke up after %s", actualSleepDuration.String()))
	}
	s.PrintPlainTextResponse(conn, reqData)
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

	log.Printf("TCP Server is running on :%d\n", s.port)
	for {
		conn, err := ln.Accept()
		if err != nil {
			return fmt.Errorf("error accepting connection: %v\n", err)
		}
		go s.handleTcpConnection(conn)
	}
}
