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
		// DNS labels next to a right-to-left label: the bidi rule applies to the
		// internationalized label only
		{"_acme-challenge.שלום.com", "_acme-challenge.xn--9dbne9b.com"},
		{"*.שלום.com", "*.xn--9dbne9b.com"},
		{"_acme-challenge.xn--9dbne9b.com", "_acme-challenge.xn--9dbne9b.com"},
		// "--" in positions 3 and 4 is allowed in ASCII labels
		{"ab--cd.example", "ab--cd.example"},
		{"r3--sn-abc.bücher.de", "r3--sn-abc.xn--bcher-kva.de"},
		// The dots that UTS #46 maps to "."
		{"münchen。de", "xn--mnchen-3ya.de"},
		{"münchen．de", "xn--mnchen-3ya.de"},
		// An internationalized label that starts with a digit: fine without a
		// right-to-left label in the name
		{"1é.com", "xn--1-bga.com"},
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
		{"_acme-challenge.xn--9dbne9b.com", "_acme-challenge.שלום.com"},
		{"*.xn--9dbne9b.com", "*.שלום.com"},
		{"ab--cd.example", "ab--cd.example"},
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
		{"", "the name is empty"},
		{".", "the name is empty"},
		{"a b.de", "not allowed in a DNS name"},
		{"a..de", "label 2 of 3: the label is empty"},
		{".de", "label 1 of 2: the label is empty"},
		{"-bad.de", "not allowed in a DNS name"},
		{"bad-.de", "not allowed in a DNS name"},
		{"bücher_x.de", "not a valid internationalized label"},
		{"www.*.de", "wildcard"},
		{"a\u200db.de", "not a valid internationalized label"}, // ZWJ outside its context
		{"abcש.com", "not a valid internationalized label"},    // bidi rule within the label
		{"xn--zz.de", "not a valid internationalized label"},   // invalid Punycode
		{"ab--cü.de", "not a valid internationalized label"},   // hyphens in 3 and 4 of an internationalized label
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

// With a right-to-left label in the name, every internationalized label must satisfy
// the bidi rule (RFC 5893), also in another label; ASCII labels stay DNS labels
func TestBidiName(t *testing.T) {
	for _, name := range []string{"1é.שלום", "שלום.1é", "xn--1-bga.xn--9dbne9b"} {
		for fn, f := range map[string]func(string) (string, error){"ToASCII": ToASCII, "ToUnicode": ToUnicode} {
			if got, err := f(name); err == nil || !strings.Contains(err.Error(), "bidi rule of a name") {
				t.Errorf("%s(%q) = %q, %v; want a bidi rule error", fn, name, got, err)
			}
		}
	}
	for _, name := range []string{"_acme-challenge.שלום.com", "1abc.שלום.com", "www.שלום.com", "é.שלום"} {
		if _, err := ToASCII(name); err != nil {
			t.Errorf("ToASCII(%q): %v", name, err)
		}
	}
}

// The name may be sensitive: errors never contain it or a part of it
func TestErrorsDoNotContainTheName(t *testing.T) {
	for _, name := range []string{"secret origin.example", "secret-ü_x.example", "xn--zz.secret.example", "abcש.secret.example"} {
		_, err := ToASCII(name)
		if err == nil {
			t.Fatalf("ToASCII(%q) did not fail", name)
		}
		for _, part := range []string{"secret", "origin", "xn--zz", "abc"} {
			if strings.Contains(err.Error(), part) {
				t.Errorf("ToASCII(%q): error %q contains %q", name, err, part)
			}
		}
	}
}
