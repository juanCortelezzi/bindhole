package blacklist

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	ParserSimple    = "simple"
	ParserIpSkipper = "ip_skipper"
)

var (
	ErrLoadingConfig = errors.New("error loading config")
	ErrParsingConfig = errors.New("error parsing config")
)

type rawBlacklist struct {
	Url    string
	Reason string
	Source string
	Parser string
}

type ParsedBlacklist struct {
	Url    *url.URL
	Reason string
	Source string
	Parser string
}

type TomlBlacklists struct {
	Blacklist []rawBlacklist
}

func loadConfigFile(path string) ([]rawBlacklist, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w on path '%s': %w", ErrLoadingConfig, path, err)
	}

	var tomlBlacklists TomlBlacklists
	if err := toml.Unmarshal(file, &tomlBlacklists); err != nil {
		log.Fatal(err)
		return nil, fmt.Errorf("%w on path '%s': %w", ErrParsingConfig, path, err)
	}

	return tomlBlacklists.Blacklist, nil
}

func filterMapInvalidBlacklists(rawBlacklists []rawBlacklist) []ParsedBlacklist {

	var blacklists []ParsedBlacklist
	for _, rawBlacklist := range rawBlacklists {
		url, err := url.ParseRequestURI(rawBlacklist.Url)
		if err != nil {
			log.Printf("Error parsing url: %v, skipping.", err)
			continue
		}

		if rawBlacklist.Parser != ParserSimple && rawBlacklist.Parser != ParserIpSkipper {
			log.Printf("Parser '%s' not supported, skipping.", rawBlacklist.Parser)
			continue
		}

		blacklists = append(blacklists, ParsedBlacklist{
			Url:    url,
			Reason: rawBlacklist.Reason,
			Source: rawBlacklist.Source,
			Parser: rawBlacklist.Parser,
		})
	}

	return blacklists
}

func GetBlacklistConfigPath() (string, error) {
	configHome, found := os.LookupEnv("XDG_CONFIG_HOME")
	if !found {
		configHome = os.ExpandEnv("$HOME")
	}

	if configHome == "" {
		return "", fmt.Errorf("%w: could not find config home", ErrLoadingConfig)
	}

	if strings.HasSuffix(configHome, "/") {
		configHome = strings.TrimSuffix(configHome, "/")
	}

	configPath := fmt.Sprintf("%s/bindhole/blacklists.toml", configHome)

	return configPath, nil

}

func GetBlacklistsFromConfig(configPath string) ([]ParsedBlacklist, error) {
	rawBlacklists, err := loadConfigFile(configPath)
	if err != nil {
		return nil, err
	}

	blacklists := filterMapInvalidBlacklists(rawBlacklists)

	return blacklists, nil
}
