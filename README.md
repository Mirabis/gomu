[![Release](https://img.shields.io/github/release/mirabis/gomu.svg)](https://github.com/mirabis/gomu/releases/latest)
[![Actions Status](https://github.com/mirabis/gomu/workflows/Release/badge.svg)](https://github.com/mirabis/gomu/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/Mirabis/gomu)](https://goreportcard.com/report/github.com/Mirabis/gomu)

# gomu 

Small, fast, simple tool for performing reverse Office365 lookups.

You feed **GOMU** (Get OnMicrosoft Users) e-mail aliases, it returns onMicrosoft ID's

This can be a useful way of finding usernames belonging to a company using Azure Directories from their e-mail addresses.

## Installation

```sh
go get github.com/mirabis/gomu
```
or download a binary from the Releases page.

## Usage
The most basic usage is to simply pipe a list of e-mail addresses into the tool, for example:

```sh
mirabis~$ cat company.emails | gomu 
peter.paul@contoso.com, peter.paul@contoso.mail.onmicrosoft.com
john.williams@contoso.com, john.williams@contoso.mail.onmicrosoft.com
andrea.keeper@contoso.com, andrea.keeper@contoso.mail.onmicrosoft.com
darryl.blue@contoso.com, darryl.blue@contoso.mail.onmicrosoft.com
...
```

### Parameters

```sh
mirabis~$ gomu -h

Usage:
  gomu [OPTIONS]

  	██████╗  ██████╗ ███╗   ███╗██╗   ██╗
	██╔════╝ ██╔═══██╗████╗ ████║██║   ██║
	██║  ███╗██║   ██║██╔████╔██║██║   ██║
	██║   ██║██║   ██║██║╚██╔╝██║██║   ██║
	╚██████╔╝╚██████╔╝██║ ╚═╝ ██║╚██████╔╝
	╚═════╝  ╚═════╝ ╚═╝     ╚═╝ ╚═════╝ 

Application Options: (/* windows, -* Unix)
  /t, /threads:     How many threads should be used (default: 20)
  /i, /input:       Input file containing line seperated e-mail addresses,
                    otherwise defaults to STDIN
  /d, /domain:      Autodiscover domain to use
                    (default: outlook.office365.com)
  /u, /user-agent:  User specified User agent to override default
                    (default: Microsoft Office/16.0 (Windows NT 10.0; Microsoft
                    Outlook 16.0.12026; Pro))
  /v, /verbose      Turns on verbose logging
      /insecure     Switches all HTTPS calls to HTTP

Help Options:
  /?, /h, /help     Show this help message
```


## Credits
- [hakluke](https://twitter.com/hakluke) my inspiration to start transitioning from Python/.NET to golang
- [s0md3v](https://github.com/s0md3v) repository formatting and tool inspiration
- [raikia](https://github.com/Raikia/UhOh365) initial idea

### Contribution & License
You can contribute in following ways:

- Report bugs
- Give suggestions to make it better (I'm new to golang)
- Fix issues & submit a pull request

Do you want to have a conversation in private? Hit me up on my [twitter](https://twitter.com/iMirabis/), inbox is open :)

**gomu** is licensed under [GPL v3.0 license](https://www.gnu.org/licenses/gpl-3.0.en.html)