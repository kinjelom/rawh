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

`$ rawh --help`
```text
rawh functions either as an HTTP server or as a client to diagnose requests and responses.

Usage:
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

Use "rawh [command] --help" for more information about a command.
```

#### Server Usage
`$ rawh server --help`
```text
Run as an HTTP server

Usage:
  rawh server [flags]

Flags:
  -h, --help       help for server
  -p, --port int   Specify the port the server will listen on (default 8080)

Global Flags:
  -v, --verbose   Enables verbose output for the operation (client and server modes).
  -V, --version   Displays the application version 
```

#### Client usage
`$ rawh client --help`
```text
Run as an HTTP client

Usage:
  rawh client <url> [flags]

Flags:
  -C, --canonical                   Specifies whether the 'canonical' client should be used; by default, the 'raw' client will be used.
  -d, --data string                 Data to be sent as the body of the request, typically with 'POST'.
      --generate-data-size string   Data size [B|KB|MB|GB] to be generated and sent as the body of the request, typically with 'POST'.
  -H, --header stringArray          Adds a header to the request, format 'key: value'.
  -h, --help                        help for client
      --http string                 Specifies the HTTP version to use (options: 1.0, 1.1, 2). (default "1.1")
  -k, --insecure                    Allow insecure server connections.
  -X, --method string               Specifies the HTTP method to use (e.g., 'GET', 'POST'). (default "GET")
      --normalize-headers           Normalize header names format.
      --tls string                  Specifies the TLS version to use (options: 1.0, 1.1, 1.2, 1.3). (default "1.2")

Global Flags:
  -v, --verbose   Enables verbose output for the operation (client and server modes).
  -V, --version   Displays the application version.
```

### Additional Options

The server allows for artificially extending the query execution time by sleeping for the specified [duration](https://pkg.go.dev/time#ParseDuration), an option is useful for testing timeouts:
- query parameter: `?rawh-sleep-duration=10m`
- HTTP header: `Rawh-Sleep-Duration: 10m`

You can use the special `Rawh-Echo` HTTP header to request that the `rawh` server returns the content of the header as HTTP response headers, with the values matching the requested header names (case-sensitive).
For example, the request header `Rawh-Echo: header-1 hEADERr-2` prompts the `rawh` server to include the following headers in its response:
- `header-1: header-1`
- `hEADERr-2: hEADERr-2`


## Example

### 1. Server: run
```bash
$ rawh server -v --port=8080
```
```text
TCP Server is running on :8080
````
### 2. Client: request
```bash
$ rawh client -v -X POST -H "hEaDeR: abc" -H "Rawh-Echo: test-1 tESt-2" --generate-data-size=10B http://localhost:8080/?rawh-sleep-duration=5s
```
```text
> POST /?rawh-sleep-duration=5s HTTP/1.1
> Host: localhost:8080
> hEaDeR: abc
> Rawh-Echo: test-1 tESt-2
> Content-Length: 10
> 
> 1234567890
````

### 3. Server: log
```text
# Read request: start
< POST /?rawh-sleep-duration=5s HTTP/1.1
< Host: localhost:8080
< hEaDeR: abc
< Rawh-Echo: test-1 tESt-2
< Content-Length: 10
< 
# Start of body reading
# body-size: 10
# body-hash: MD5:e807f1fcf82d132f9bb018ca6738a19f
# End of body reading
# Read request: done [0s]
# Going to sleep for 5s
# Woke up after 5.004s
> HTTP/1.1 200 OK
> Content-Type: text/plain
> test-1: test-1
> tESt-2: tESt-2
> 
> request-start-line: POST /?rawh-sleep-duration=5s HTTP/1.1
> request-header-lines:
> - Host: localhost:8080
> - hEaDeR: abc
> - Rawh-Echo: test-1 tESt-2
> - Content-Length: 10
> request-body-size: 10.00 B
> request-body-hash: MD5:e807f1fcf82d132f9bb018ca6738a19f
> request-read-duration: 0s
> request-sleep-duration: 5s
```  

### 4. Client: response
```text
< HTTP/1.1 200 OK
< Content-Type: text/plain
< test-1: test-1
< tESt-2: tESt-2
< 
request-start-line: POST /?rawh-sleep-duration=5s HTTP/1.1
request-header-lines:
- Host: localhost:8080
- hEaDeR: abc
- Rawh-Echo: test-1 tESt-2
- Content-Length: 10
request-body-size: 10.00 B
request-body-hash: MD5:e807f1fcf82d132f9bb018ca6738a19f
request-read-duration: 0s
request-sleep-duration: 5s
```
