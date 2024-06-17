package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/juancortelezzi/bindhole/blacklist"
)

type Args struct {
	config string
	output string
}

func NewArgs() (*Args, error) {
	args := Args{}

	flagSet := flag.NewFlagSet("config", flag.ExitOnError)
	flagSet.StringVar(&args.config, "config", "", "Path to the configuration file")
	flagSet.StringVar(&args.config, "c", "", "Path to the configuration file")
	flagSet.StringVar(&args.output, "output", "./bindhole.zone", "Path to the output file")
	flagSet.StringVar(&args.output, "o", "./bindhole.zone", "Path to the output file")
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	if args.config == "" {
		configPath, err := blacklist.GetBlacklistConfigPath()
		if err != nil {
			return nil, fmt.Errorf("could not get default config path: %v", err)
		}
		args.config = configPath
	}

	return &args, nil
}

func main() {
	args, err := NewArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing arguments: %v\n", err)
		return
	}

	blacklists, err := blacklist.GetBlacklistsFromConfig(args.config)
	if err != nil {
		log.Fatalf("Could not get blacklists from config: %v", err)
	}

	zoneFile, err := blacklist.NewRPZFile(args.output)
	if err != nil {
		log.Fatalf("Could not create zone file: %v", err)
	}

	defer zoneFile.Close()

	for _, list := range blacklists {
		handleBlacklist(list, zoneFile)
	}
}

func handleBlacklist(list blacklist.ParsedBlacklist, zoneFile *blacklist.RPZFile) {
	resp, err := http.Get(list.Url.String())
	if err != nil {
		log.Panic(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error fetching %s: %s", list.Url.String(), resp.Status)
		return
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

			log.Fatalf("error reading host: %v", err)
		}

		host := string(hostBuffer[:hostBufferPtr])
		if err := zoneFile.BlacklistHost(host); err != nil {
			log.Fatalf("Could not write host to zone file: %v", err)
		}
	}
}
