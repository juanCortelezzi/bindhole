package main

import (
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/juancortelezzi/bindhole/blacklist"
)

func handleBlacklist(list blacklist.ParsedBlacklist, zoneFile *blacklist.RPZFile) {
	resp, err := http.Get(list.Url.String())
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Error fetching %s: %s", list.Url.String(), resp.Status)
	}

	reader, err := blacklist.NewParsedReader(list.Parser, resp.Body)
	if err != nil {
		log.Printf("Could not wrap reader: %v", err)
		return
	}

	hostBuffer := make([]byte, 512)
	for {
		hostBufferPtr, err := reader.Read(hostBuffer)

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			log.Panicf("error reading host: %v", err)
		}

		host := string(hostBuffer[:hostBufferPtr])
		if err := zoneFile.BlacklistHost(host); err != nil {
			log.Fatalf("Could not write host to zone file: %v", err)
		}
	}
}

func main() {
	blacklists, err := blacklist.GetBlacklistsFromConfig()
	if err != nil {
		log.Fatalf("Could not get blacklists from config: %v", err)
	}

	zoneFile, err := blacklist.NewRPZFile("bindhole.zone")
	if err != nil {
		log.Fatalf("Could not create zone file: %v", err)
	}

	defer zoneFile.Close()

	for _, list := range blacklists {
		handleBlacklist(list, zoneFile)
	}
}
