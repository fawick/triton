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

Action             | Command
-------------------|-----------------------------------------------------------------------------
List all Droplets  | `triton list droplets`
List all SSH Keys  | `triton list keys `
List all Images    | `triton list images`
Create a Droplet   | `triton droplet create <name> <image> [-region=<region slug>] [--keys=keyid1,keyid2]`
Delete a Droplet   | `triton droplet delete <name>`
Transfer an Image  | `triton transfer <image name> <region-slug>`


Planned actions    | Command
-------------------|-----------------------------------------------------------------------------
Power on a Droplet | `triton droplet poweron <name>`
Shutdown a Droplet | `triton droplet shutdown <name>`
Reboot a Droplet   | `triton droplet reboot <name>`
Create an Image    | `triton image snapshot <droplet name> <image name>`
