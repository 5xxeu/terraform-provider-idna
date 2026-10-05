# Terraform Provider: idna

Converts internationalized domain names (IDN) between Unicode and the ASCII form used in DNS, so a configuration can say `münchen.de` where an API expects `xn--mnchen-3ya.de`:

```hcl
terraform {
  required_version = ">= 1.8.0"
  required_providers {
    idna = {
      source  = "i386dev/idna"
      version = "~> 0.1"
    }
  }
}

locals {
  zone = provider::idna::to_ascii("münchen.de")       # "xn--mnchen-3ya.de"
  name = provider::idna::to_unicode("xn--bcher-kva")  # "bücher"
}
```

The provider has only [provider-defined functions](https://developer.hashicorp.com/terraform/plugin/framework/functions) (Terraform 1.8 or later, OpenTofu 1.7 or later): no configuration, resources or data sources, and no network access.

## Functions

| Function | Result |
|----------|--------|
| `to_ascii(name)` | The ASCII form: `münchen.de` is `xn--mnchen-3ya.de`, `MÜNCHEN.DE` too |
| `to_unicode(name)` | The Unicode form: `xn--mnchen-3ya.de` is `münchen.de` |

Both take a domain name or a name in a zone: `münchen.de`, `www.bücher`, `_acme-challenge.münchen.de`, `*.münchen.de`, `münchen.de.` (fully qualified) or `@` (the zone apex). See [docs/functions](docs/functions).

## Rules

- The conversion follows [UTS #46](https://www.unicode.org/reports/tr46/) with nontransitional processing, like browsers: names are mapped (lower case, compatibility characters such as `ﬁ` to `fi`) and checked (IDNA 2008 contexts, bidi rules). `ß` stays `ß` (`faß.de` is `xn--fa-hia.de`), as registries handle it today.
- What zones need beyond host names is kept: labels with `_` (`_dmarc`, `_sip._tcp`), a `*` wildcard as the leftmost label, `@`, a trailing dot. ASCII labels are only lower-cased.
- The result is checked as a DNS name: labels of 1 to 63 characters and at most 253 in total in ASCII (a short Unicode label can be long in Punycode), Punycode labels of letters, digits and `-`, other labels of letters, digits, `_` and `-`.
- A name that is not valid fails at `plan` with the reason, for example `"a b.de" has a label with characters that are not allowed in a DNS name`.
- UTS #46 accepts some symbols that IDNA 2008 registries refuse, such as emoji (`😀.de` is `xn--e28h.de`), as browsers do. Whether a registry accepts a name is up to the registry.

The conversion uses [golang.org/x/net/idna](https://pkg.go.dev/golang.org/x/net/idna).

## With easy-dns

[easy-dns](https://github.com/i386dev/terraform-cloudflare-easy-dns) expects internationalized names in Punycode, as the Cloudflare API does. With this provider, the configuration can keep them readable:

```hcl
module "dns" {
  source = "git::https://github.com/i386dev/terraform-cloudflare-easy-dns.git?ref=v2.12.0"

  zone_id   = var.zone_id
  zone_name = provider::idna::to_ascii("münchen.de")
  records = {
    (provider::idna::to_ascii("bücher")) = { A = [{ content = "192.0.2.1" }] }
  }
}
```

Record keys in the state then contain the ASCII form (`xn--bcher-kva A 192.0.2.1`), the same as if the names were written in Punycode.

## Development

```sh
go test ./...        # unit tests and the functions through Terraform (terraform in PATH, 1.8 or later)
go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.25.0 generate --provider-name idna
```

To try a local build, point Terraform at it with `dev_overrides` in the [CLI configuration](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers).

## License

[MIT](LICENSE)
