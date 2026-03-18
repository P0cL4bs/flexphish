![flexphish logo](/docs/banner.png)

# Flexphish

**Flexphish** is a flexible and modular phishing framework designed for **security professionals, red teams, and researchers** to simulate real-world phishing campaigns in controlled environments. It allows controlled testing of phishing scenarios by creating realistic login pages and capturing interactions for analysis, it provides a modern architecture with support for **custom templates, campaign management, and traffic monitor**, making it ideal for **penetration testing, awareness training, and development of phishing simulations**.

![flexphish dashboard](/docs/dashboard.png)

## Features

- Modular phishing campaign engine
- Wildcard domain campaign 
- YAML-based template phishing pages
- Campaign tracking and credential capture

## Releases & Installation Guide

This guide explains how to download, install, and run **Flexphish** using pre-built binaries or from source.


### Releases

You can download the latest stable release from GitHub:

https://github.com/P0cL4bs/flexphish/releases

Pre-built binaries are available for multiple platforms, including:

* Linux (amd64)
* Windows (amd64) *(future support)*

Each release includes compiled binaries and release notes describing changes, improvements, and fixes.


### Quick Start (Binary Installation)

Follow the steps below to quickly get Flexphish running on Linux.

### Download the Binary

```bash
wget https://github.com/P0cL4bs/flexphish/releases/download/v1.2.1/flexphish_linux_amd64_18e0104.zip
```

### Extract the Archive

```bash
unzip flexphish_linux_amd64_18e0104.zip
cd flexphish
```


### Make the Binary Executable

```bash
chmod +x flexphish
```


### Run Flexphish

```bash
./flexphish
```


### Flexphish CLI

```text
        ████           
   ██████████████      
  █████▓▓▓▓▓▓█████                                                         
  ███▓▓▓▓▓▓░░▓▓███     ██████ ▄▄    ▄▄▄▄▄ ▄▄ ▄▄ ▄▄▄▄  ▄▄ ▄▄ ▄▄  ▄▄▄▄ ▄▄ ▄▄
 ███▓▓▓▓▓▓▓██▓▓▓███    ██▄▄   ██    ██▄▄  ▀█▄█▀ ██▄█▀ ██▄██ ██ ███▄▄ ██▄██
████▓▓▓██▓▓██▓▓▓████   ██     ██▄▄▄ ██▄▄▄ ██ ██ ██    ██ ██ ██ ▄▄██▀ ██ ██
 ███▓▓▓██▓▓██▓▓▓███                                            version 1.2.1-dev
  ███▓▓▓████▓▓▓███     The ultimate Red Team toolkit for phishing operations.
  █████▓▓▓▓▓▓█████     
   ██████████████      [built for linux amd64]
        ████            by: @mh4x0f (PocL4bs Team  - 10 Years Anniversary
                       )

[+] Campaign server running on http://0.0.0.0:8001  
[+] API server starting on 0.0.0.0:8088  
[+] Dashboard running on http://0.0.0.0:8000  
```


### Creating a User

Flexphish allows user management directly from the command line.

To create a new user:

```bash
./flexphish -create-user \
-email admin@example.com \
-password StrongPassword
```


### Accessing the Application

After starting the server, you can access:

* Dashboard → http://localhost:8000
* API → http://localhost:8088
* Campaign Server → http://localhost:8001


### Development Build

If you prefer to build Flexphish from source:

### Requirements

- Go **1.24.0**
- Nginx (for production)
- pnpm (10.11.0)
---


```bash
git clone https://github.com/P0cL4bs/flexphish.git
cd flexphish

go mod tidy
go build -o flexphish

make frontend

./flexphish
```

## Documentation

Full documentation is available in the docs/ directory:

- [`docs/development.md`](/docs/development.md) - Local development setup

- [`docs/nginx-cloudflare-setup.md`](/docs/nginx-cloudflare-setup.md) - Production deployment

- [`docs/templates.md`](/docs/templates.md) - Template structure and behavior of phishing pages


## Templates Flows

**Flexphish** templates define the **structure and behavior of phishing pages** used in campaigns.

They are built using **YAML configuration files** combined with **HTML pages and static assets**, allowing the creation of highly realistic and customizable phishing flows.

### What a Template Controls

A template flow is responsible for:

- Capturing user data (credentials, form fields, tokens)  
- Input validation and rules enforcement  
- Step transitions and flow control
- Redirect behavior after completion  
- Dynamic and reusable variables  
- Client-side scripts and interactions  
- Multi-step authentication sequences (e.g. login → password → 2FA)  


### Execution Flow

Templates are executed sequentially, step-by-step:

```text
username → password → 2FA → redirect
```

Full documentation: [`docs/templates.md`](/docs/templates.md)


## Issues

If you encounter a bug, have a feature request, or need help, please open an issue on GitHub:

https://github.com/P0cL4bs/flexphish/issues


## License

This project is licensed under the **Apache License 2.0**

## Disclaimer

```bash
This tool is intended for educational purposes and authorized security testing only.
The author is not responsible for any misuse or damage caused by this software.
Users are responsible for complying with applicable laws and regulations.
```