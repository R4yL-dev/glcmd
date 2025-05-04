# glcmd

**Version**: 0.1<br>
**Date**: 04.05.25

`glcmd` is a small command-line tool that interoperates with the libreview API to retrieve blood glucose information from a LibreLinkUp account.

## Prerequisites

- go 1.24.1

> I haven't tested on anything other than Linux. `make install` puts the binary in `/usr/local/bin`, so I don't know if the folder is present on Mac. If not, you can `make` and paste the binary there into `bin/` wherever you like.

## Setup

To enable `glcmd` to connect to your LibreLinkUp account, you need to add 2 environment variables: `GL_EMAIL` and `GL_PASSWORD`.

For `glcmd` to work, you need to give it the credentials of a follower account. That is, not your main account (on the libre3 application, for example), but a follower account that must be added as an associate device in the libre3 application. Using the credentials of a patient account directly does not work.

## Install

```bash
> export GL_EMAIL='<email>'
> export GL_PASSWORD='<password>'
> git clone git@github.com:R4yL-dev/glcmd.git
> cd glcmd
> make
> ./bin/glcmd
🩸 7.7(mmo/L) 🡒
```

## TODO

### ToJSON()

Add the possibility of generating json on standard output rather than the classic glucose display.

### ASCIIGraph

Adds the possibility of displaying the measurement graph in the terminal as an ascii graph.

### Watcher

`glcmd` simply connects to the libreview api to retrieve the current measurement. I'd like to add a watcher in the future. It should be able to run in infinity and retrieve measurements every defined period of time. This would make it possible, for example, to create a database or text file to track measurements over time.
