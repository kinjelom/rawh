# Raw HTTP `rawh`

`rawh` acts either as an HTTP server or as a client to diagnose requests and responses.

Typically, HTTP clients and servers (including libraries) normalize headers, which means we can't diagnose what we are actually sending and receiving.

This application is a simple client or server that:
- does not normalize the format of HTTP headers 
- describes requests and responses to better understand what is happening

## HTTP Sepcyfications

- **Hypertext Transfer Protocol (HTTP/1.1): Message Syntax and Routing**
  > [3.2. Header Fields](https://datatracker.ietf.org/doc/html/rfc7230#section-3.2)  
  > Each header field consists of a case-insensitive field name followed
  > by a colon (":"), optional leading whitespace, the field value, and
  > optional trailing whitespace.

- **Hypertext Transfer Protocol Version 2 (HTTP/2)**
  > [8.1.2.  HTTP Header Fields](https://datatracker.ietf.org/doc/html/rfc7540#section-8.1.2)
  > HTTP header fields carry information as a series of key-value pairs.
  > For a listing of registered HTTP headers, see the "Message Header
  > Field" registry maintained at <https://www.iana.org/assignments/message-headers>.
  > Just as in HTTP/1.x, header field names are strings of ASCII
  > characters that are compared in a case-insensitive fashion.  However,
  > header field names MUST be converted to lowercase prior to their
  > encoding in HTTP/2.  A request or response containing uppercase
  > header field names MUST be treated as malformed (Section 8.1.2.6).

- **Haproxy case adjustment of HTTP/1 headers sent to bogus clients**
  > [option h1-case-adjust-bogus-client](https://www.haproxy.com/documentation/haproxy-configuration-manual/2-8r1/#4.2-option%20h1-case-adjust-bogus-client)
  > There is no standard case for header names because, as stated in RFC7230,
  > they are case-insensitive. So applications must handle them in a case-
  > insensitive manner. But some bogus applications violate the standards and
  > erroneously rely on the cases most commonly used by browsers. This problem
  > becomes critical with HTTP/2 because all header names must be exchanged in
  > lower case, and HAProxy follows the same convention. All header names are
  > sent in lower case to clients and servers, regardless of the HTTP version.
  > 
  > When HAProxy receives an HTTP/1 response, its header names are converted to
  > lower case and manipulated and sent this way to the clients. If a client is
  > known to violate the HTTP standards and to fail to process a response coming
  > from HAProxy, it is possible to transform the lower case header names to a
  > different format when the response is formatted and sent to the client, by
  > enabling this option and specifying the list of headers to be reformatted
  > using the global directives "h1-case-adjust" or "h1-case-adjust-file". This
  > must only be a temporary workaround for the time it takes the client to be
  > fixed, because clients which require such workarounds might be vulnerable to
  > content smuggling attacks and must absolutely be fixed.

## Usage

### Command line

```text
rawh [flags]
rawh [command]

Available Commands:
  client      Run as an HTTP client
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  server      Run as an HTTP server

Flags:
  -h, --help      help for rawh
  -v, --verbose   Enables verbose output for the operation (client and server modes).
  -V, --version   Displays the application version.
```

#### Server Usage
```text
Flags:
  -h, --help       help for server
  -p, --port int   Specify the port the server will listen on (default 8080)
``` 

#### Client usage
```text
Flags:
  -d, --data string              Data to be sent as the body of the request, typically with 'POST'.
      --generate-data-size int   Data size [bytes] to be generated and sent as the body of the request, typically with 'POST'.
  -H, --header stringArray       Adds a header to the request, format 'key: value'.
  -h, --help                     help for client
      --http string              Specifies the HTTP version to use (options: 1.0, 1.1, 2). (default "1.1")
  -X, --method string            Specifies the HTTP method to use (e.g., 'GET', 'POST'). (default "GET")
      --normalize-headers        Normalize header names format.
      --tls string               Specifies the TLS version to use (options: 1.0, 1.1, 1.2, 1.3). (default "1.2")
  -k, --insecure                 Allow insecure server connections.
      --url string               Specifies the URL for the client request (default "http://localhost:8080")
```

### Additional Options

The server allows for artificially extending the query execution time by sleeping for the specified [duration](https://pkg.go.dev/time#ParseDuration), an option is useful for testing timeouts:
- query parameter: `?rawh-sleep-duration=10m`
- HTTP header: `rawh-Sleep-Duration: 10m`

HTTP headers that start with the prefix `rawh-Echo-` will be included in the HTTP response headers. 

## Example

### 1. Server: run
```bash
rawh --verbose server --port=8080
```
```text
TCP Server is running on :8080
````
### 2. Client: request
```bash
rawh --verbose client -X POST -H "x-hEaDeR: abc" -H "x-header: abc" --data="abc"  --url="http://localhost:8080/?rawh-sleep-duration=5s"
```
```text
> = REQUEST:
> POST /?rawh-sleep-duration=5s HTTP/1.1
> req.Host:
> == REQUEST HEADERS:
> - x-hEaDeR: abc
> - x-header: abc
> - Content-Length: 3
````

### 3. Server: log
```text
2024/06/03 11:22:45 Read request: start
2024/06/03 11:22:45 < start-line: POST /?rawh-sleep-duration=5s HTTP/1.1
2024/06/03 11:22:45 < header-line: Host: localhost:8080
2024/06/03 11:22:45 < header-line: User-Agent: Go-http-client/1.1
2024/06/03 11:22:45 < header-line: Content-Length: 3
2024/06/03 11:22:45 < header-line: x-hEaDeR: abc
2024/06/03 11:22:45 < header-line: x-header: abc
2024/06/03 11:22:45 < header-line: Accept-Encoding: gzip
2024/06/03 11:22:45 < *** start of body reading ***
2024/06/03 11:22:45 < body-size: 3
2024/06/03 11:22:45 < body-hash: MD5:900150983cd24fb0d6963f7d28e17f72
2024/06/03 11:22:45 < *** end of body reading ***
2024/06/03 11:22:45 Read request: done [0s]
2024/06/03 11:22:45 Going to sleep for 5s
2024/06/03 11:22:50 Woke up after 5.004s
2024/06/03 11:22:50 > HTTP/1.1 200 OK
2024/06/03 11:22:50 > Content-Type: text/plain
2024/06/03 11:22:50 > 
2024/06/03 11:22:50 > request-start-line: POST /?rawh-sleep-duration=5s HTTP/1.1
2024/06/03 11:22:50 > request-header-lines:
2024/06/03 11:22:50 > - Host: localhost:8080
2024/06/03 11:22:50 > - User-Agent: Go-http-client/1.1
2024/06/03 11:22:50 > - Content-Length: 3
2024/06/03 11:22:50 > - x-hEaDeR: abc
2024/06/03 11:22:50 > - x-header: abc
2024/06/03 11:22:50 > - Accept-Encoding: gzip
2024/06/03 11:22:50 > request-body-size: 3.00 B
2024/06/03 11:22:50 > request-body-hash: MD5:900150983cd24fb0d6963f7d28e17f72
2024/06/03 11:22:50 > request-read-duration: 0s
2024/06/03 11:22:50 > request-sleep-duration: 5s
```  

### 4. Client: response
```text
< = RESPONSE:
< HTTP/1.1 200 OK
< == RESPONSE HEADERS:
< - Content-Type: text/plain

request-start-line: POST /?rawh-sleep-duration=5s HTTP/1.1
request-header-lines:
  - Host: localhost:8080
  - User-Agent: Go-http-client/1.1
  - Content-Length: 3
  - x-hEaDeR: abc
  - x-header: abc
  - Accept-Encoding: gzip
request-body-size: 3.00 B
request-body-hash: MD5:900150983cd24fb0d6963f7d28e17f72
request-read-duration: 1ms
request-sleep-duration: 5s
```
