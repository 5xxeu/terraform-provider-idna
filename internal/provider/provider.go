// Package provider is a function-only provider: no configuration, resources or data
// sources, only the functions to_ascii and to_unicode.
package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ provider.Provider              = &idnaProvider{}
	_ provider.ProviderWithFunctions = &idnaProvider{}
)

type idnaProvider struct {
	version string
}

// New returns the provider factory for the given version.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &idnaProvider{version: version}
	}
}

func (p *idnaProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "idna"
	resp.Version = p.version
}

func (p *idnaProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Converts internationalized domain names between Unicode and the ASCII form used in DNS (Punycode). No configuration.",
	}
}

func (p *idnaProvider) Configure(context.Context, provider.ConfigureRequest, *provider.ConfigureResponse) {
}

func (p *idnaProvider) Resources(context.Context) []func() resource.Resource { return nil }

func (p *idnaProvider) DataSources(context.Context) []func() datasource.DataSource { return nil }

func (p *idnaProvider) Functions(context.Context) []func() function.Function {
	return []func() function.Function{newToASCII, newToUnicode}
}
