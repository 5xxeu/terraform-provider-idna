package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// The functions through Terraform itself (provider-defined functions need 1.8)
var factories = map[string]func() (tfprotov6.ProviderServer, error){
	"idna": providerserver.NewProtocol6WithError(New("test")()),
}

func TestFunctions(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_8_0)},
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{
			{
				Config: `
output "ascii" { value = provider::idna::to_ascii("münchen.de") }
output "prefixed" { value = provider::idna::to_ascii("_acme-challenge.Bücher.de.") }
output "unicode" { value = provider::idna::to_unicode("xn--b1agh1afp.xn--p1ai") }
output "apex" { value = provider::idna::to_ascii("@") }
output "zone_records" {
  value = { for name in ["www", "bücher"] : provider::idna::to_ascii(name) => true }
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("ascii", knownvalue.StringExact("xn--mnchen-3ya.de")),
					statecheck.ExpectKnownOutputValue("prefixed", knownvalue.StringExact("_acme-challenge.xn--bcher-kva.de.")),
					statecheck.ExpectKnownOutputValue("unicode", knownvalue.StringExact("привет.рф")),
					statecheck.ExpectKnownOutputValue("apex", knownvalue.StringExact("@")),
					statecheck.ExpectKnownOutputValue("zone_records", knownvalue.MapExact(map[string]knownvalue.Check{
						"www":           knownvalue.Bool(true),
						"xn--bcher-kva": knownvalue.Bool(true),
					})),
				},
			},
		},
	})
}

func TestFunctionErrors(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_8_0)},
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{
			{
				Config:      `output "x" { value = provider::idna::to_ascii("a b.de") }`,
				ExpectError: regexp.MustCompile(`not\s+allowed\s+in\s+a\s+DNS\s+name`),
			},
			{
				Config:      `output "x" { value = provider::idna::to_unicode("xn--zz.de") }`,
				ExpectError: regexp.MustCompile(`not\s+a\s+valid\s+internationalized\s+label`),
			},
		},
	})
}
