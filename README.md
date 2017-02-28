# It's 2017!

This tool is deprecated. Use github.com/digitalocean/doctl instead.

# triton

According to [Wikipedia](http://en.wikipedia.org/wiki/Triton_%28mythology%29),
the mythological Greek god Triton was the messenger of the sea. `triton` is the
messenger for the [DigitalOcean](https://www.digitalocean.com) [API
v2](https://developers.digitalocean.com/#images).

`triton` is a command-line tool written in [Go](http://www.golang.org). It was
started to handle my personal use cases, not to have the full API covered from
the start on. However, it is easy enough to extend it with all the API parts
that one might find missing. I am happy to receive your Pull Requests if you
want to contribute to `triton`.

## Installation

	go get github.com/fawick/triton

## Usage

Set the authentication token via environment variable `$DIGITALOCEAN_API_TOKEN`
or via global option `--token`.

Action              | Command
--------------------|-----------------------------------------------------------------------------
List all Droplets   | `triton list droplets`
List all SSH Keys   | `triton list keys `
List all Images     | `triton list images`
Create a Droplet    | `triton droplet create <name> <image> [-region=<region slug>] [--keys=keyid1,keyid2]`
Delete a Droplet    | `triton droplet delete <name>`
Power on a Droplet  | `triton droplet poweron <name>`
Power off a Droplet | `triton droplet poweroff <name>`
Shutdown a Droplet  | `triton droplet shutdown <name>`
Reboot a Droplet    | `triton droplet reboot <name>`
Power off and on    | `triton droplet powercycle <name>`
Password reset      | `triton droplet passwordreset <name>`
Enable IPv6         | `triton droplet ipv6 <name>`
Disable backups     | `triton droplet disablebackups <name>`
Private networking  | `triton droplet privatenetworking <name>`
Transfer an Image   | `triton transfer <image name> <region-slug>`


Planned actions    | Command
-------------------|-----------------------------------------------------------------------------
Create an Image    | `triton image snapshot <droplet name> <image name>`

