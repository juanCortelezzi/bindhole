package blacklist

import (
	"errors"
	"net"
	"regexp"
	"strings"
)

var (
	ErrInvalidDomain = errors.New("invalid domain")

	domainRegexp = regexp.MustCompile(`^(?i)[a-z0-9-]+(\.[a-z0-9-]+)+\.?$`)
)

// isValidDomain returns true if the domain is valid.
//
// It uses a simple regular expression to check the domain validity.
func isValidDomain(domain string) bool {
	return domainRegexp.MatchString(domain)
}

// validateDomainByResolvingIt queries DNS for the given domain name,
// and returns nil if the the name resolves, or an error.
func validateDomainByResolvingIt(domain string) error {
	if !isValidDomain(domain) {
		return ErrInvalidDomain
	}
	addr, err := net.LookupHost(domain)
	if err != nil {
		return err
	}
	if len(addr) == 0 {
		return ErrInvalidDomain
	}
	return nil
}

// normalizeDomain returns a normalized domain.
// It returns an empty string if the domain is not valid.
func normalizeDomain(domain string) string {
	// Trim whitespace.
	domain = strings.TrimSpace(domain)
	// Check validity.
	if !isValidDomain(domain) {
		return ""
	}
	// Remove trailing dot.
	domain = strings.TrimRight(domain, ".")
	// Convert to lower case.
	domain = strings.ToLower(domain)
	return domain
}
