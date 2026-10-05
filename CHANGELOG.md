# Changelog

All notable changes to this project are documented in this file. The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project follows [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- Functions `to_ascii` and `to_unicode`: conversion of domain names between Unicode and the ASCII form (Punycode) after UTS #46, nontransitional processing, label by label: the IDNA rules apply to internationalized labels, ASCII labels are checked as DNS labels (`_`, `--` inside, also next to right-to-left labels); a leftmost `*` wildcard, `@` and a trailing dot are kept; the result is checked as a DNS name; errors name the label position and the reason, never the name
