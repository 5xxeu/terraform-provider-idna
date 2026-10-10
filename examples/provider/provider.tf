terraform {
  required_version = ">= 1.8.0"
  required_providers {
    idna = {
      source  = "5xxeu/idna"
      version = "~> 0.1"
    }
  }
}

# The provider has no configuration; its functions are called as provider::idna::<name>
output "zone_name" {
  value = provider::idna::to_ascii("münchen.de") # "xn--mnchen-3ya.de"
}
