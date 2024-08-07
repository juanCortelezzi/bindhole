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
	configPath     string
	outputFilePath string
	server         string
	port           int
}

func NewArgs() (*Args, error) {
	args := Args{}

	flagSet := flag.NewFlagSet("programConfiguration", flag.ExitOnError)

	flagSet.StringVar(&args.configPath, "config", "", "Path to the configuration file")
	flagSet.StringVar(&args.configPath, "c", "", "Path to the configuration file")

	flagSet.StringVar(&args.outputFilePath, "output", "./blockme.list", "Path to the output file")
	flagSet.StringVar(&args.outputFilePath, "o", "./blockme.list", "Path to the output file")

	flagSet.StringVar(&args.server, "server", "127.0.0.1", "DNS server to use")
	flagSet.StringVar(&args.server, "s", "127.0.0.1", "DNS server to use")

	flagSet.IntVar(&args.port, "port", 53, "DNS server port to use")
	flagSet.IntVar(&args.port, "p", 53, "DNS server port to use")

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	if args.configPath == "" {
		configPath, err := blacklist.GetBlacklistConfigPath()
		if err != nil {
			return nil, fmt.Errorf("could not get default config path: %v", err)
		}
		args.configPath = configPath
	}

	return &args, nil
}

func main() {
	args, err := NewArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing arguments: %v\n", err)
		return
	}

	log.Printf("Reading config file: '%s'\n", args.configPath)

	blacklists, err := blacklist.GetBlacklistsFromConfig(args.configPath)
	if err != nil {
		log.Fatalf("Could not get blacklists from config: %v", err)
	}

	log.Printf("Opening output file: '%s'\n", args.outputFilePath)

	zoneFile, err := blacklist.NewRPZFile(
		args.outputFilePath,
		args.server,
		args.port,
	)
	if err != nil {
		log.Fatalf("Could not create list file: %v", err)
	}

	defer zoneFile.Close()

	log.Printf("Writing configuration for '%s:%d'\n", args.server, args.port)

	for _, list := range blacklists {
		log.Printf("Gathering data from '%s' using '%s' parser\n", list.Url.String(), list.Parser)
		handleBlacklist(list, zoneFile)
	}
	log.Printf("Finished writing blacklists to '%s'\n", args.outputFilePath)
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
