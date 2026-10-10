package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"

	"github.com/5xxeu/terraform-provider-idna/internal/convert"
)

// stringFunction is a function of one string that returns a string or an error about
// its argument.
type stringFunction struct {
	name, summary, description string
	convert                    func(string) (string, error)
}

func (f *stringFunction) Metadata(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = f.name
}

func (f *stringFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:             f.summary,
		MarkdownDescription: f.description,
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:                "name",
				MarkdownDescription: "A domain name or a name in a zone: `münchen.de`, `www.bücher`, `_acme-challenge.münchen.de`, `*.münchen.de`, `münchen.de.` or `@`.",
			},
		},
		Return: function.StringReturn{},
	}
}

func (f *stringFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var name string
	resp.Error = function.ConcatFuncErrors(req.Arguments.Get(ctx, &name))
	if resp.Error != nil {
		return
	}
	out, err := f.convert(name)
	if err != nil {
		resp.Error = function.NewArgumentFuncError(0, err.Error())
		return
	}
	resp.Error = function.ConcatFuncErrors(resp.Result.Set(ctx, out))
}

func newToASCII() function.Function {
	return &stringFunction{
		name:    "to_ascii",
		summary: "Converts a domain name to its ASCII form (Punycode)",
		description: "Returns the ASCII form of a domain name, as DNS and the Cloudflare API expect it: " +
			"`münchen.de` is `xn--mnchen-3ya.de`. Follows UTS #46 (nontransitional processing, as browsers do): " +
			"the name is mapped (lower case, compatibility characters) and checked (IDNA 2008 and bidi rules). " +
			"ASCII labels pass unchanged except for lower case; `_` labels, a leftmost `*` wildcard, `@` and a " +
			"trailing dot are kept. Fails on names that are not valid, with the reason.",
		convert: convert.ToASCII,
	}
}

func newToUnicode() function.Function {
	return &stringFunction{
		name:    "to_unicode",
		summary: "Converts a domain name to its Unicode form",
		description: "Returns the Unicode form of a domain name: `xn--mnchen-3ya.de` is `münchen.de`. " +
			"The name is checked like in `to_ascii` first, so an invalid `xn--` label fails instead of passing through.",
		convert: convert.ToUnicode,
	}
}
