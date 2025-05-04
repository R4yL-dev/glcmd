# glcmd

**Version**: 0.1
**Date**: 2025-05-04

`glcmd` is a small command-line tool that queries the LibreView API to retrieve blood glucose information from a LibreLinkUp (follower) account.

It displays the latest glucose value directly in your terminal.

## ✨ Features

- Retrieve the current glucose measurement via the LibreLinkUp API.
- Display it in the terminal.

## 📦 Prerequisites

- **Go** 1.24.1

> 📌 This project has only been tested on **Linux** for now.
> `make install` places the binary in `/usr/local/bin`. If this folder does not exist on macOS, simply compile it with `make` and move the binary to a folder included in your `PATH`.

## ⚙️ Setup

Before using `glcmd`, you need to set two environment variables: `GL_EMAIL` and `GL_PASSWORD`.

These credentials must belong to a **follower account** — meaning an associated device account (not your primary patient account from the Libre 3 app).
The follower account must be added as an associated device in the Libre 3 application.
Using direct patient account credentials will not work.

## 🚀 Install & Usage

```bash
export GL_EMAIL='<email>'
export GL_PASSWORD='<password>'
git clone https://github.com/R4yL-dev/glcmd.git
cd glcmd
make
./bin/glcmd
🩸 7.7(mmol/L) 🡒
```

## 📌 TODO

- **US format**: add a flag to display glucose in US format.
- **ToJSON()**: add an option to generate JSON output on stdout instead of classic text display.
- **ASCIIGraph**: add the ability to display an ASCII graph of recent measurements in the terminal.
- **Watcher**: implement a continuous monitoring mode, polling the API at a defined interval and logging the measurements to a file or database.

## 📄 License

This project is licensed under the [MIT](LICENSE) license.

## ⚠️ Disclaimer

This tool is provided for informational and personal use only.
It is not a certified medical device and should not be used to make health-related decisions.
Use it at your own risk.
