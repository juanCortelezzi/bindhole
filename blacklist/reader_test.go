package blacklist_test

import (
	"errors"
	"io"
	"log"
	"strings"
	"testing"

	"github.com/juancortelezzi/bindhole/blacklist"
)

const simpleBlocklist = `# Easylist, parsed and mirrored by https://firebog.net
# Updated 10JUN24 from https://easylist-downloads.adblockplus.org/easylist.txt

# This is sourced from an "adblock" style list which is flat-out NOT designed to work with DNS sinkholes
# There WILL be mistakes with how this is parsed, due to how host names are extracted and exceptions handled
# Please bring any parsing issues up at https://github.com/WaLLy3K/wally3k.github.io/issues prior to raising a request upstream

# If your issue IS STILL PRESENT when using uBlock/ABP/etc, you should request a correction at https://github.com/easylist/easylist#list-issues

0008d6ba2e.com
0024ad98dd.com
0083334e84.com`

const ipSkipperBlocklist = `ff02::1 ip6-allnodes
ff02::2 ip6-allrouters
ff02::3 ip6-allhosts
0.0.0.0 0.0.0.0

# Custom host records are listed here.

#=====================================
# Title: Hosts contributed by Steven Black
# http://stevenblack.com

0.0.0.0 ck.getcookiestxt.com
0.0.0.0 eu1.clevertap-prod.com
0.0.0.0 wizhumpgyros.com
0.0.0.0 wizhumpgyros2.com # This is a very stupid comment`

func TestSimpleParsedReader(t *testing.T) {
	stringReader := strings.NewReader(simpleBlocklist)
	reader := blacklist.NewSimpleParsedReader(stringReader)
	buffer := make([]byte, 512)
	parsedHosts := make([]string, 0, 3)
	expectedHosts := []string{"0008d6ba2e.com", "0024ad98dd.com", "0083334e84.com"}
	for {
		bufferPtr, err := reader.Read(buffer)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			t.Fatalf("error reading host: %v", err)
		}

		host := string(buffer[:bufferPtr])
		parsedHosts = append(parsedHosts, host)
	}

	if len(parsedHosts) != len(expectedHosts) {
		t.Fatalf("expected %d hosts, got %d", len(expectedHosts), len(parsedHosts))
	}

	for i, parsedHost := range parsedHosts {
		if parsedHost != expectedHosts[i] {
			t.Fatalf("expected host %s, got %s", expectedHosts[i], parsedHost)
		}
	}

}

func TestIPSkipperParsedReader(t *testing.T) {
	stringReader := strings.NewReader(ipSkipperBlocklist)
	reader := blacklist.NewIPSkipperParsedReader(stringReader)
	buffer := make([]byte, 512)
	parsedHosts := make([]string, 0, 3)
	expectedHosts := []string{"ck.getcookiestxt.com", "eu1.clevertap-prod.com", "wizhumpgyros.com", "wizhumpgyros2.com"}
	for {
		bufferPtr, err := reader.Read(buffer)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			t.Fatalf("error reading host: %v", err)
		}

		host := string(buffer[:bufferPtr])
		parsedHosts = append(parsedHosts, host)
	}

	if len(parsedHosts) != len(expectedHosts) {
		t.Fatalf("expected %d hosts, got %d", len(expectedHosts), len(parsedHosts))
	}

	for i, parsedHost := range parsedHosts {
		if parsedHost != expectedHosts[i] {
			t.Fatalf("expected host %s, got %s", expectedHosts[i], parsedHost)
		}
	}
}
