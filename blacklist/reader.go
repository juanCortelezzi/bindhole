package blacklist

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
)

const (
	simpleBlacklistComment = "#"
)

var (
	ErrParserNotImplemented = errors.New("parser not implemented")
	ErrBufferTooSmall       = fmt.Errorf("buffer is too small to contain the host: %w", io.ErrShortBuffer)
	localhosts              = []string{"127.0.0.1", "255.255.255.255", "::1", "localhost", "0.0.0.0"}
)

type SimpleParsedReader struct {
	innerReader *bufio.Reader
}

var _ io.Reader = &SimpleParsedReader{}

func NewSimpleParsedReader(reader io.Reader) *SimpleParsedReader {
	return &SimpleParsedReader{
		innerReader: bufio.NewReader(reader),
	}
}

func (r *SimpleParsedReader) Read(p []byte) (int, error) {
	for {
		line, err := r.innerReader.ReadBytes('\n')
		// unkwnown error.
		if err != nil && !errors.Is(err, io.EOF) {
			return 0, err
		}

		// reached end of file and there is no data to read.
		if err != nil && errors.Is(err, io.EOF) && len(line) == 0 {
			return 0, err
		}

		line = bytes.TrimSpace(line)

		if len(line) == 0 {
			continue
		}

		if bytes.HasPrefix(line, []byte(simpleBlacklistComment)) {
			continue
		}

		domain := normalizeDomain(string(line))
		if domain == "" {
			log.Printf("Invalid domain: %s", line)
			continue
		}

		n := copy(p, domain)
		if n < len(domain) {
			return 0, ErrBufferTooSmall
		}

		return n, nil
	}
}

type IPSkipperReader struct {
	innerReader *bufio.Reader
}

var _ io.Reader = &IPSkipperReader{}

func NewIPSkipperParsedReader(reader io.Reader) *IPSkipperReader {
	return &IPSkipperReader{
		innerReader: bufio.NewReader(reader),
	}
}

func (r *IPSkipperReader) Read(p []byte) (int, error) {
readloop:
	for {
		line, err := r.innerReader.ReadBytes('\n')
		// unkwnown error.
		if err != nil && !errors.Is(err, io.EOF) {
			return 0, err
		}

		// reached end of file and there is no data to read.
		if err != nil && errors.Is(err, io.EOF) && len(line) == 0 {
			return 0, err
		}

		line = bytes.TrimSpace(line)

		prefix := []byte("0.0.0.0 ")
		if !bytes.HasPrefix(line, prefix) {
			continue
		}

		line = bytes.TrimPrefix(line, prefix)

		if len(line) == 0 {
			continue
		}

		line, _, _ = bytes.Cut(line, []byte(" "))

		for _, localhost := range localhosts {
			if bytes.Equal(line, []byte(localhost)) {
				continue readloop
			}
		}

		domain := normalizeDomain(string(line))
		if domain == "" {
			log.Printf("Invalid domain: %s", line)
			continue
		}

		n := copy(p, domain)
		if n < len(domain) {
			return 0, ErrBufferTooSmall
		}

		return n, nil
	}
}

func NewParsedReader(parserName string, reader io.Reader) (io.Reader, error) {
	switch parserName {
	case ParserSimple:
		return NewSimpleParsedReader(reader), nil
	case ParserIpSkipper:
		return NewIPSkipperParsedReader(reader), nil
	default:
		return nil, fmt.Errorf("parser '%s' not implemented", parserName)
	}
}
