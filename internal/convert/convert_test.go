package convert

import (
	"strings"
	"testing"
)

func TestToASCII(t *testing.T) {
	cases := []struct{ in, want string }{
		{"münchen.de", "xn--mnchen-3ya.de"},
		{"MÜNCHEN.DE", "xn--mnchen-3ya.de"},
		{"faß.de", "xn--fa-hia.de"}, // nontransitional: ß is kept, not mapped to ss
		{"日本.jp", "xn--wgv71a.jp"},
		{"привет.рф", "xn--b1agh1afp.xn--p1ai"},
		{"שלום.com", "xn--9dbne9b.com"},
		{"_acme-challenge.bücher.de", "_acme-challenge.xn--bcher-kva.de"},
		{"_sip._tcp.münchen.de", "_sip._tcp.xn--mnchen-3ya.de"},
		{"*.münchen.de", "*.xn--mnchen-3ya.de"},
		{"münchen.de.", "xn--mnchen-3ya.de."},
		{"www", "www"},
		{"Example.COM", "example.com"},
		{"xn--mnchen-3ya.de", "xn--mnchen-3ya.de"},
		{"@", "@"},
		{"ﬁ.de", "fi.de"}, // UTS #46 mapping
	}
	for _, c := range cases {
		got, err := ToASCII(c.in)
		if err != nil || got != c.want {
			t.Errorf("ToASCII(%q) = %q, %v; want %q", c.in, got, err, c.want)
		}
	}
}

func TestToUnicode(t *testing.T) {
	cases := []struct{ in, want string }{
		{"xn--mnchen-3ya.de", "münchen.de"},
		{"XN--MNCHEN-3YA.DE", "münchen.de"},
		{"xn--b1agh1afp.xn--p1ai", "привет.рф"},
		{"_acme-challenge.xn--bcher-kva.de", "_acme-challenge.bücher.de"},
		{"*.xn--mnchen-3ya.de.", "*.münchen.de."},
		{"münchen.de", "münchen.de"},
		{"example.com", "example.com"},
		{"@", "@"},
	}
	for _, c := range cases {
		got, err := ToUnicode(c.in)
		if err != nil || got != c.want {
			t.Errorf("ToUnicode(%q) = %q, %v; want %q", c.in, got, err, c.want)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	for _, name := range []string{"münchen.de", "日本.jp", "привет.рф", "_dmarc.bücher.de", "*.faß.de."} {
		ascii, err := ToASCII(name)
		if err != nil {
			t.Fatalf("ToASCII(%q): %v", name, err)
		}
		back, err := ToUnicode(ascii)
		if err != nil || back != name {
			t.Errorf("ToUnicode(ToASCII(%q)) = %q, %v", name, back, err)
		}
	}
}

func TestErrors(t *testing.T) {
	cases := []struct{ in, contains string }{
		{"", "is not a domain name"},
		{".", "is not a domain name"},
		{"a b.de", "not allowed in a DNS name"},
		{"a..de", "empty label"},
		{".de", "empty label"},
		{"-bad.de", "not a valid internationalized domain name"}, // hyphen rules of IDNA
		{"bücher_x.de", "not allowed in an internationalized label"},
		{"www.*.de", "wildcard that is not the leftmost label"},
		{"a‍b.de", "not a valid internationalized domain name"},    // ZWJ outside its context
		{"abcש.com", "not a valid internationalized domain name"},  // bidi rule
		{"xn--zz.de", "not a valid internationalized domain name"}, // invalid Punycode
		{strings.Repeat("a", 64) + ".de", "more than the 63"},
		{strings.Repeat("ü", 60) + ".de", "more than the 63"}, // short in Unicode, long in ASCII
		{strings.Repeat("a.", 127) + "de", "more than the 253"},
	}
	for _, c := range cases {
		got, err := ToASCII(c.in)
		if err == nil || !strings.Contains(err.Error(), c.contains) {
			t.Errorf("ToASCII(%q) = %q, %v; want an error with %q", c.in, got, err, c.contains)
		}
	}
	if got, err := ToUnicode("xn--zz.de"); err == nil {
		t.Errorf("ToUnicode(%q) = %q, want an error", "xn--zz.de", got)
	}
}
