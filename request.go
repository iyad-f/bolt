package bolt

import (
	"bufio"
	"errors"
	"io"
	"net/url"
	"strconv"
	"strings"
)

// Request represents a parsed HTTP request.
type Request struct {
	Method        string
	RequestURI    string
	URL           *url.URL
	Proto         string
	ProtoMajor    int
	ProtoMinor    int
	Header        Header
	Body          io.ReadCloser
	ContentLength int64
	Host          string
	RemoteAddr    string
}

// ReadRequest parses an HTTP/1.x request from the given buffered reader.
func ReadRequest(br *bufio.Reader) (*Request, error) {
	method, uri, proto, err := parseRequestLine(br)
	if err != nil {
		return nil, err
	}

	protoMajor, protoMinor, err := parseProto(proto)
	if err != nil {
		return nil, err
	}

	parsedURL, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	header, err := parseHeaders(br)
	if err != nil {
		return nil, err
	}

	contentLength, err := parseContentLength(header)
	if err != nil {
		return nil, err
	}

	var body io.ReadCloser
	if contentLength > 0 {
		body = io.NopCloser(io.LimitReader(br, contentLength))
	} else {
		body = io.NopCloser(strings.NewReader(""))
	}

	return &Request{
		Method:        method,
		RequestURI:    uri,
		URL:           parsedURL,
		Proto:         proto,
		ProtoMajor:    protoMajor,
		ProtoMinor:    protoMinor,
		Header:        header,
		Body:          body,
		ContentLength: contentLength,
		Host:          header.Get("Host"),
	}, nil
}

func parseRequestLine(br *bufio.Reader) (method, uri, proto string, err error) {
	line, err := br.ReadString('\n')
	if err != nil {
		return "", "", "", err
	}
	line = strings.TrimRight(line, "\r\n")

	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return "", "", "", errors.New("malformed HTTP request line")
	}

	return parts[0], parts[1], parts[2], nil
}

func parseProto(proto string) (major, minor int, err error) {
	if !strings.HasPrefix(proto, "HTTP/") {
		return 0, 0, errors.New("malformed HTTP protocol")
	}

	version := proto[len("HTTP/"):]
	parts := strings.Split(version, ".")
	if len(parts) != 2 {
		return 0, 0, errors.New("malformed HTTP version")
	}

	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, errors.New("malformed HTTP major version")
	}

	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, errors.New("malformed HTTP minor version")
	}

	return major, minor, nil
}

func parseHeaders(br *bufio.Reader) (Header, error) {
	header := Header{}

	for {
		line, err := br.ReadString('\n')
		if err != nil {
			return header, err
		}

		line = strings.TrimRight(line, "\r\n")

		if line == "" {
			return header, nil
		}

		key, value, ok := strings.Cut(line, ":")
		if !ok {
			return header, errors.New("malformed HTTP header")
		}
		header.Add(key, strings.TrimLeft(value, " \t"))
	}
}

func parseContentLength(header Header) (int64, error) {
	cl := header.Get("Content-Length")
	if cl == "" {
		return -1, nil
	}
	n, err := strconv.ParseUint(cl, 10, 63)
	if err != nil {
		return 0, errors.New("bad Content-Length")
	}
	return int64(n), nil
}
