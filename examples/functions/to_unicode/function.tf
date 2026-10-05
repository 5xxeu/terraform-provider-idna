# "münchen.de"
output "zone" {
  value = provider::idna::to_unicode("xn--mnchen-3ya.de")
}
