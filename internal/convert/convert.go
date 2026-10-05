// Package convert converts domain names between their Unicode form and the ASCII
// form used in DNS (Punycode, "xn--" labels), following UTS #46 (nontransitional
// processing, as browsers do) with the extensions DNS zones need: "_" in labels
// (_acme-challenge, _sip._tcp), a "*" wildcard as the leftmost label, "@" for the
// zone apex and a trailing dot for a fully qualified name.
package convert

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/idna"
)

// The UTS #46 lookup profile without the STD3 rules, so "_" and "*" pass through;
// the result is checked by validate instead.
var profile = idna.New(
	idna.MapForLookup(),
	idna.BidiRule(),
	idna.Transitional(false),
	idna.StrictDomainName(false),
)

var (
	// A label with an "xn--" prefix holds Punycode: letters, digits and hyphens only
	aLabel = regexp.MustCompile(`^xn--[a-z0-9-]+$`)
	// Other labels: letters, digits, "_" and "-", not starting or ending with "-"
	ldhLabel = regexp.MustCompile(`^[a-z0-9_]([a-z0-9_-]*[a-z0-9_])?$`)
)

// ToASCII returns the ASCII form of a domain name: "münchen.de" is "xn--mnchen-3ya.de".
// ASCII labels are only lower-cased. "@" and a trailing dot are kept.
func ToASCII(name string) (string, error) {
	if name == "@" {
		return name, nil
	}
	body, root := strings.CutSuffix(name, ".")
	if body == "" {
		return "", fmt.Errorf("%q is not a domain name", name)
	}
	out, err := profile.ToASCII(body)
	if err != nil {
		return "", fmt.Errorf("%q is not a valid internationalized domain name: %w", name, err)
	}
	if err := validate(name, out); err != nil {
		return "", err
	}
	if root {
		out += "."
	}
	return out, nil
}

// ToUnicode returns the Unicode form of a domain name: "xn--mnchen-3ya.de" is
// "münchen.de". The name is checked like in ToASCII first, so an invalid "xn--" label
// fails instead of passing through.
func ToUnicode(name string) (string, error) {
	ascii, err := ToASCII(name)
	if err != nil {
		return "", err
	}
	if ascii == "@" {
		return ascii, nil
	}
	body, root := strings.CutSuffix(ascii, ".")
	out, err := profile.ToUnicode(body)
	if err != nil {
		return "", fmt.Errorf("%q is not a valid internationalized domain name: %w", name, err)
	}
	if root {
		out += "."
	}
	return out, nil
}

// validate checks the ASCII form against the rules of DNS names in zones: labels of
// 1 to 63 characters, at most 253 characters in total, "*" only as the whole leftmost
// label.
func validate(name, ascii string) error {
	if len(ascii) > 253 {
		return fmt.Errorf("%q is %d characters long in ASCII (%q), more than the 253 a domain name can have", name, len(ascii), ascii)
	}
	for i, label := range strings.Split(ascii, ".") {
		switch {
		case label == "":
			return fmt.Errorf("%q has an empty label", name)
		case len(label) > 63:
			return fmt.Errorf("%q has a label of %d characters in ASCII (%q), more than the 63 a label can have", name, len(label), label)
		case label == "*":
			if i != 0 {
				return fmt.Errorf("%q has a wildcard that is not the leftmost label", name)
			}
		case strings.HasPrefix(label, "xn--"):
			if !aLabel.MatchString(label) {
				return fmt.Errorf("%q has a label with characters that are not allowed in an internationalized label (%q)", name, label)
			}
		case !ldhLabel.MatchString(label):
			return fmt.Errorf("%q has a label with characters that are not allowed in a DNS name (%q): letters, digits, \"_\" and \"-\" only, not starting or ending with \"-\"", name, label)
		}
	}
	return nil
}
