# pd

[![License](https://img.shields.io/github/license/homeport/pd.svg)](https://github.com/homeport/pd/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/homeport/pd)](https://goreportcard.com/report/github.com/homeport/pd)
[![Build Status](https://travis-ci.org/homeport/pd.svg?branch=main)](https://travis-ci.org/homeport/pd)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/homeport/pd)](https://pkg.go.dev/github.com/homeport/pd)
[![Release](https://img.shields.io/github/release/homeport/pd.svg)](https://github.com/homeport/pd/releases/latest)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/homeport/pd)

/pe.de/ - command line tool to query PagerDuty for the current on-calls

## Setup

Install the `pd` binary via `brew` on macOS, or find the respective binary of the [latest release](https://github.com/homeport/pd/releases/latest).

```sh
brew install homeport/tap/pd
```

Create a `.pd.yml` file in your home directory that contains a PagerDuty authentication token. Go to your PagerDuty profile settings and under "User Settings" you will find a `Create API User Token` button.

![pd-yaml-example](.docs/images/pd-yaml.png?raw=true "Example of the PagerDuty config file")
