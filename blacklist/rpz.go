package blacklist

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

const rpzHeader = `$TTL 2w
@ IN SOA localhost. root.localhost. (
       2   ; serial
       2w  ; refresh
       2w  ; retry
       2w  ; expiry
       2w) ; minimum
    IN NS localhost.

`

var (
	ErrWritingToFile = fmt.Errorf("could not write to file")
)

type RPZFile struct {
	file         *os.File
	writer       *bufio.Writer
	blacklistSet map[string]struct{}
}

func NewRPZFile(path string) (*RPZFile, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	writer := bufio.NewWriter(file)

	if _, err := writer.WriteString(rpzHeader); err != nil {
		return nil, err
	}

	return &RPZFile{
		file:         file,
		writer:       writer,
		blacklistSet: make(map[string]struct{}),
	}, nil
}

func (r *RPZFile) BlacklistHost(host string) error {

	if _, ok := r.blacklistSet[host]; ok {
		log.Printf("host '%s' already in the list, skipping.", host)
		return nil
	}

	r.blacklistSet[host] = struct{}{}
	baseHostLine := fmt.Sprintf("%s CNAME .\n", host)
	if _, err := r.writer.WriteString(baseHostLine); err != nil {
		return fmt.Errorf("Error writing host '%s': %w: %w", host, ErrWritingToFile, err)
	}

	if host[0] == '*' {
		return nil
	}

	widlcardHostLine := fmt.Sprintf("*.%s CNAME .\n", host)
	if _, err := r.writer.WriteString(widlcardHostLine); err != nil {
		return fmt.Errorf("Error writing host '%s': %w: %w", host, ErrWritingToFile, err)
	}
	return nil
}

func (r *RPZFile) Close() error {
	if err := r.writer.Flush(); err != nil {
		return fmt.Errorf("could not flush file: %w", err)
	}

	if err := r.file.Close(); err != nil {
		return fmt.Errorf("could not close file: %w", err)
	}

	return nil
}
