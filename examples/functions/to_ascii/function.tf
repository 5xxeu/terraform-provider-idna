# "xn--mnchen-3ya.de"
output "zone" {
  value = provider::idna::to_ascii("münchen.de")
}

# Names in a zone keep "_" labels, a leftmost "*", "@" and a trailing dot:
# "_acme-challenge.xn--bcher-kva", "*.xn--bcher-kva", "@"
output "names" {
  value = [for name in ["_acme-challenge.bücher", "*.bücher", "@"] : provider::idna::to_ascii(name)]
}

variable "zone_id" {
  type = string
}

# With the easy-dns module (github.com/i386dev/terraform-cloudflare-easy-dns): the zone
# name and the record names in Punycode, as Cloudflare expects them
module "dns" {
  source = "git::https://github.com/i386dev/terraform-cloudflare-easy-dns.git?ref=v2.12.0"

  zone_id   = var.zone_id
  zone_name = provider::idna::to_ascii("münchen.de")
  records = {
    (provider::idna::to_ascii("bücher")) = { A = [{ content = "192.0.2.1" }] }
  }
}
