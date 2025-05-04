# glcmd

## 🎯 About

**Version**: 0.1.1
**Date**: 2025-05-04

`glcmd` is a command-line tool designed to retrieve and display blood glucose information from the **LibreView API** using a **LibreLinkUp follower account**. It allows users to quickly and easily monitor their glucose levels directly in the terminal, without the need for proprietary apps.

This tool is ideal for people who want to have more control and flexibility over their glucose data, providing a simple, open-source alternative for tracking and displaying their measurements.

### 🌟 Key Features

- **Retrieve current glucose readings** from the LibreView API using a **follower account**.
- **Display glucose levels** in the terminal in a human-readable format (mmol/L).
- **Open-source**: freely available to use, modify, and contribute to.
- **Planned improvements**:
  - JSON output for better integration.
  - ASCII graph to visualize glucose trends.
  - Continuous monitoring mode to track glucose levels over time and store data in a file or database.

### 💡 Why `glcmd`?

Managing diabetes requires constant tracking and monitoring of glucose levels. `glcmd` was created to offer users a lightweight, no-frills tool to access their glucose data without being tied to a proprietary platform or app. It aims to give people more flexibility, transparency, and control over their health data in a simple command-line interface.

## 📦 Prerequisites

- **Go** 1.24.1

> 🚨 This project has only been tested on **Linux** for now.
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

## 📄 License

This project is licensed under the [MIT](LICENSE) license.

## ⚠️ Disclaimer

This tool is provided for informational and personal use only.
It is not a certified medical device and should not be used to make health-related decisions.
Use it at your own risk.
