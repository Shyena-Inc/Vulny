# Vulny

**Vulny** is a multi-tool web vulnerability scanner written in Go. It automates the process of running multiple open-source security tools to identify vulnerabilities in web applications and servers.

## Features

- Runs a suite of popular security tools (Nmap, WPScan, JoomScan, Droopescan, SSLScan, Amass, Nikto, Dalfox, WAFW00F, and more)
- Aggregates and summarizes vulnerability findings
- Generates detailed vulnerability and debug reports
- Supports skipping specific tools
- Update checker and self-updater
- Colorful, user-friendly CLI output

## Installation

### Prerequisites

- Go 1.21 or higher
- External tools: `host`, `nmap`, `wpscan`, `joomscan`, `droopescan`, `sslscan`, `amass`, `nikto`, `dalfox`, `wafw00f`
- On Debian/Ubuntu, you can install dependencies using:
  ```sh
  sudo apt install dnsutils nmap ruby-dev joomscan python3-pip sslscan amass nikto
  sudo gem install wpscan
  pip3 install droopescan
  # Install dalfox and wafw00f as needed
  ```

## Usage

```
              _
 /\   /\_   _| |_ __  _   _
 \ \ / / | | | | '_ \| | | |
  \ V /| |_| | | | | | |_| |
   \_/  \__,_|_|_| |_|\__, |
                      |___/

```

```sh
vulny [options]
```

### Options

- `-target <url>`: Target URL to scan (e.g., example.com)
- `-skip <tools>`: Comma-separated list of tools to skip (e.g., host,nmap)
- `-update`: Check for updates and update the tool
- `-no-spinner`: Disable loading spinner during scan
- `-help`: Show help message

### Examples

```sh
vulny -target example.com
vulny -target example.com -skip host,nmap
vulny -update
vulny -help
```

## Output

- Vulnerability and debug reports are generated in the current directory after each scan.
- Temporary files are cleaned up automatically.

## Updating

To update Vulny to the latest version:

```sh
vulny -update
```

## License

[MIT](https://github.com/Shyena-Inc/Vulny/blob/main/LICENSE)

## Author

[aryanstha4859](https://github.com/aryanstha4859)
