package blacklist

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

const flushRecordCount = 10

var (
	ErrWritingToFile = fmt.Errorf("could not write to file")
)

type RPZFile struct {
	outputFile   *os.File
	writer       *bufio.Writer
	blacklistSet map[string]struct{}
	recordCount  int
}

func NewRPZFile(outputPath, server string, port int) (*RPZFile, error) {
	outputFile, err := os.OpenFile(outputPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("could not open output file with path '%s' because: %w", outputPath, err)
	}

	writer := bufio.NewWriter(outputFile)

	initialConfig := fmt.Sprintf("server %s %d\nttl 600\nzone bindholerpz\n", server, port)
	if _, err := writer.WriteString(initialConfig); err != nil {
		return nil, fmt.Errorf("could not write server initial config to file: %w", err)
	}

	return &RPZFile{
		outputFile:   outputFile,
		writer:       writer,
		blacklistSet: make(map[string]struct{}),
		recordCount:  0,
	}, nil
}

func (r *RPZFile) BlacklistHost(host string) error {
	if _, ok := r.blacklistSet[host]; ok {
		log.Printf("Host '%s' already in the list, skipping.", host)
		return nil
	}

	r.blacklistSet[host] = struct{}{}

	baseHostLine := fmt.Sprintf("update add %s.bindholerpz CNAME .\n", host)

	if r.recordCount%flushRecordCount == 0 && r.recordCount != 0 {
		baseHostLine += "send\n"
	}

	if _, err := r.writer.WriteString(baseHostLine); err != nil {
		return fmt.Errorf("Error writing host '%s': %w: %w", host, ErrWritingToFile, err)
	}

	// widlcardHostLine := fmt.Sprintf("update add *.%s.bindholerpz 600 CNAME .\n", host)
	// if _, err := r.writer.WriteString(widlcardHostLine); err != nil {
	// 	return fmt.Errorf("Error writing host '%s': %w: %w", host, ErrWritingToFile, err)
	// }

	r.recordCount += 1
	return nil
}

func (r *RPZFile) Close() error {
	if r.recordCount%flushRecordCount == 0 && r.recordCount != 0 {
		if err := r.writer.Flush(); err != nil {
			return fmt.Errorf("could not write last send while flushing file file: %w", err)
		}
	}

	if err := r.writer.Flush(); err != nil {
		return fmt.Errorf("could not flush file: %w", err)
	}

	if err := r.outputFile.Close(); err != nil {
		return fmt.Errorf("could not close file: %w", err)
	}

	return nil
}
