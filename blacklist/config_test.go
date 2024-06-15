package blacklist

import (
	"testing"
)

func TestFilterMapInvalidBlacklists(t *testing.T) {
	rawBlacklists := []rawBlacklist{
		{
			Url:    "https://v.firebog.net/hosts/Easylist.txt",
			Reason: "testReason1",
			Source: "testSource1",
			Parser: ParserSimple,
		},
		{
			Url:    "https://adaway.org/hosts.txt",
			Reason: "testReason2",
			Source: "testSource2",
			Parser: ParserIpSkipper,
		},
	}

	parsedBlacklists := filterMapInvalidBlacklists(rawBlacklists)

	if len(parsedBlacklists) != len(rawBlacklists) {
		t.Fatalf("expected %d blacklists, got %d", len(rawBlacklists), len(parsedBlacklists))
	}

	for i, parsedBlacklist := range parsedBlacklists {
		if parsedBlacklist.Url.String() != rawBlacklists[i].Url {
			t.Fatalf("expected url %s, got %s", rawBlacklists[i].Url, parsedBlacklist.Url.String())
		}
	}
}

func TestFilterMapInvalidBlacklistsSkipInvalidParser(t *testing.T) {
	rawBlacklists := []rawBlacklist{{
		Url:    "https://v.firebog.net/hosts/Easylist.txt",
		Reason: "testReason1",
		Source: "testSource1",
		Parser: "invalidParser",
	}}

	blacklists := filterMapInvalidBlacklists(rawBlacklists)

	if len(blacklists) != 0 {
		t.Fatalf("expected 0 blacklists, got %d", len(blacklists))
	}
}

func TestFilterMapInvalidBlacklistsSkipInvalidUrl(t *testing.T) {
	rawBlacklists := []rawBlacklist{{
		Url:    "invalidUrl",
		Reason: "testReason1",
		Source: "testSource1",
		Parser: ParserSimple,
	}}

	blacklists := filterMapInvalidBlacklists(rawBlacklists)

	if len(blacklists) != 0 {
		t.Fatalf("expected 0 blacklists, got %d", len(blacklists))
	}
}
