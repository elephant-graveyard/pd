# pd

[![License](https://img.shields.io/github/license/homeport/pd.svg)](https://github.com/homeport/pd/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/homeport/pd)](https://goreportcard.com/report/github.com/homeport/pd)
[![Tests](https://github.com/homeport/pd/workflows/Tests/badge.svg)](https://github.com/homeport/pd/actions?query=workflow%3A%22Tests%22)
[![Codecov](https://img.shields.io/codecov/c/github/homeport/pd/main.svg)](https://codecov.io/gh/homeport/pd)
[![Go Reference](https://pkg.go.dev/badge/github.com/homeport/pd.svg)](https://pkg.go.dev/github.com/homeport/pd)
[![Release](https://img.shields.io/github/release/homeport/pd.svg)](https://github.com/homeport/pd/releases/latest)

/pe.de/ - command line tool to query PagerDuty for the current on-calls

## Setup

Install the `pd` binary via `brew` on macOS, or find the respective binary of the [latest release](https://github.com/homeport/pd/releases/latest).

```sh
brew install homeport/tap/pd
```

Create a `.pd.yml` file in your home directory that contains a PagerDuty authentication token. Go to your PagerDuty profile settings and under "User Settings" you will find a `Create API User Token` button.

![pd-yaml-example](.docs/images/pd-yaml.png?raw=true "Example of the PagerDuty config file")

Next, you'll need to configure different shifts in the `.pd.yml` file. This step can be skipped if you don't need to use the `current-shift` command. The file should now look somewhat like this:

```yaml
authtoken: bm9ub25vbm9ub25vbm8K
own-shift: Team Bar
shift-times:
    - end: 08:00
      name: Team Foo
      start: 00:00
    - end: 16:00
      name: Team Bar
      start: 08:00
    - end: 00:00
      name: Team Foobar
      start: 16:00
```

## Commands

### pd on-call

Displays all current on calls.

### pd current-shift

Displays which shift is currently on-call (if shifts are configured in the `.pd.yml` file).

### pd set-own-shift [shift-name]

Updates `own-shift` in `.pd.yml` file.

### pd list-alerts

Lists all alerts that happened in a specified timeframe.

Flag | Description
--- | ---
--id \<user-ID> | list all alerts for the user
--from \<time> | lists all alerts after the specified time
--to \<time> | lists all alerts until the specified time

`user-ID` is a 7 character long ID found on `PagerDuty`.

`time` has to be provided using the format `RFC3339` (`2006-01-02T15:04:05Z07:00`)

If `--from` and `--to` are both not used, all non-resolved issues for the user are displayed.
