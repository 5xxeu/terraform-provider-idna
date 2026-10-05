// Package convert converts domain names between their Unicode form and the ASCII
// form used in DNS (Punycode, "xn--" labels), following UTS #46 (nontransitional
// processing, as browsers do) for internationalized labels, with what DNS zones need
// beyond host names: "_" labels (_acme-challenge, _sip._tcp), a "*" wildcard as the
// leftmost label, "@" for the zone apex and a trailing dot for a fully qualified name.
//
// Labels are processed one by one: the IDNA rules (mapping, contexts, bidi, hyphen
// positions, Punycode) apply to internationalized labels only, ASCII labels are checked
// as DNS labels. So "_acme-challenge" next to a right-to-left label and "ab--cd" are
// accepted, as DNS allows them.
//
// Errors name the label by its position and the reason, never the name or a label:
// the name may be a sensitive value, and Terraform shows a non-sensitive argument in its
// diagnostics anyway.
package convert

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/idna"
)

// UTS #46 lookup processing for one internationalized label
var profile = idna.New(
	idna.MapForLookup(),
	idna.BidiRule(),
	idna.Transitional(false),
)

var (
	// A label with an "xn--" prefix holds Punycode: letters, digits and hyphens only
	aLabel = regexp.MustCompile(`^xn--[a-z0-9-]+$`)
	// Other ASCII labels: letters, digits, "_" and "-", not starting or ending with "-"
	dnsLabel = regexp.MustCompile(`^[a-z0-9_]([a-z0-9_-]*[a-z0-9_])?$`)
	// The full stop and the characters UTS #46 maps to it
	dots = strings.NewReplacer("。", ".", "．", ".", "｡", ".")
)

var errEmpty = errors.New("the name is empty")

// ToASCII returns the ASCII form of a domain name: "münchen.de" is "xn--mnchen-3ya.de".
// ASCII labels are only lower-cased. "@" and a trailing dot are kept.
func ToASCII(name string) (string, error) {
	return convert(name, false)
}

// ToUnicode returns the Unicode form of a domain name: "xn--mnchen-3ya.de" is
// "münchen.de". The name is checked like in ToASCII, so an invalid "xn--" label fails
// instead of passing through.
func ToUnicode(name string) (string, error) {
	return convert(name, true)
}

func convert(name string, unicode bool) (string, error) {
	if name == "@" {
		return name, nil
	}
	body, root := strings.CutSuffix(dots.Replace(name), ".")
	if body == "" {
		return "", errEmpty
	}
	labels := strings.Split(body, ".")
	ascii := make([]string, len(labels))
	out := make([]string, len(labels))
	for i, label := range labels {
		a, u, err := convertLabel(label, i)
		if err != nil {
			return "", fmt.Errorf("label %d of %d: %w", i+1, len(labels), err)
		}
		ascii[i], out[i] = a, u
	}
	if n := len(strings.Join(ascii, ".")); n > 253 {
		return "", fmt.Errorf("the name is %d characters long in ASCII, more than the 253 a domain name can have", n)
	}
	if !unicode {
		out = ascii
	}
	result := strings.Join(out, ".")
	if root {
		result += "."
	}
	return result, nil
}

// convertLabel returns the ASCII and the Unicode form of one label
func convertLabel(label string, index int) (ascii, unicode string, err error) {
	switch {
	case label == "":
		return "", "", errors.New("the label is empty")
	case label == "*":
		if index != 0 {
			return "", "", errors.New("a wildcard \"*\" can only be the leftmost label")
		}
		return label, label, nil
	case !isASCII(label) || strings.HasPrefix(strings.ToLower(label), "xn--"):
		// An internationalized label: the IDNA rules apply
		ascii, err = profile.ToASCII(label)
		if err != nil {
			return "", "", errors.New("the label is not a valid internationalized label (UTS #46: mapping, IDNA 2008 contexts, bidi rule, hyphens, Punycode)")
		}
		unicode, err = profile.ToUnicode(ascii)
		if err != nil {
			return "", "", errors.New("the label is not a valid internationalized label (UTS #46: mapping, IDNA 2008 contexts, bidi rule, hyphens, Punycode)")
		}
		if len(ascii) > 63 {
			return "", "", fmt.Errorf("the label is %d characters long in ASCII, more than the 63 a label can have", len(ascii))
		}
		// Mapping can give an ASCII label ("ﬁ" is "fi"), which is then a DNS label
		if !aLabel.MatchString(ascii) && !dnsLabel.MatchString(ascii) {
			return "", "", errors.New("the label has characters that are not allowed in a DNS name")
		}
		return ascii, unicode, nil
	default:
		ascii = strings.ToLower(label)
		if len(ascii) > 63 {
			return "", "", fmt.Errorf("the label is %d characters long, more than the 63 a label can have", len(ascii))
		}
		if !dnsLabel.MatchString(ascii) {
			return "", "", errors.New("the label has characters that are not allowed in a DNS name: letters, digits, \"_\" and \"-\" only, not starting or ending with \"-\"")
		}
		return ascii, ascii, nil
	}
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}
