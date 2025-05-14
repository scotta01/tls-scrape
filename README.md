[![Go Reference](https://pkg.go.dev/badge/github.com/scotta01/tls-scrape.svg)](https://pkg.go.dev/github.com/scotta01/tls-scrape)

# TLS Scrape Library and Tool
TLS Scrape is both a Go library and a CLI tool designed to scrape websites for TLS certificate details and log relevant metrics.

## Features

- **Library**:
  - Scrape domains for TLS details programmatically.
  - Scan IP addresses and subnets for TLS details.
  - Check OCSP status of certificates.
  - Perform reverse DNS lookups.
  - Check if hostnames are in certificates or SANs.
  - Validate certificates and get detailed information about invalid certificates (expired, self-signed, etc.).
  - Capture and retrieve scraping metrics.

- **CLI Tool**:
  - Scrape individual domains or lists from CSV files for TLS details.
  - Scan individual IP addresses or entire subnets.
  - Expose metrics related to scraping process for Prometheus monitoring.
  - Log all scraped details including reverse DNS information.
  - Dockerized for easy deployment.

## Getting Started

### Prerequisites

- Golang (tested with version `1.24.1-bullseye`)
- Docker (if you want to build and run the Docker image)

### Building the Project

You can build and run the CLI tool using the provided Makefile:

```bash
# To build the project
make build

# To run the project
make run

# To build the Docker image
make docker-build
```

Ensure you have the required environment variables or flags set (like fqdn or filepath) when running.

## Library Usage
To use the TLS Scrape library in your Go project:

1. Import the library:
```go
import "github.com/scotta01/tls-scrape"
```
2. Utilize the scraping functions and data structures as needed in your application.
   [![Go Reference](https://pkg.go.dev/badge/github.com/scotta01/tls-scrape.svg)](https://pkg.go.dev/github.com/scotta01/tls-scrape)


## CLI Tool Configuration
You can configure the TLS Scrape tool using flags or environment variables:

- **fqdn**: Fully Qualified Domain Name. Use this if you're scraping a single domain.
- **filepath**: Path to a CSV file containing a list of websites to scrape.
- **header**: The column header in the CSV to look for. Default is url.
- **outdir**: Output directory if you wish to save the results as JSON files.
- **concurrency**: Maximum number of concurrent TLS connections. Default is 10.
- **prettyjson**: Pretty print the JSON output. Default is false.
- **bundle**: Bundle all output into a single JSON file. Default is false.
- **ip**: IP address to scan for TLS details.
- **subnet**: Subnet in CIDR notation (e.g., 192.168.1.0/24) to scan for TLS details.
- **port**: Port to connect to for TLS scanning. Default is 443.

> [!NOTE]  
> Only provide one of: fqdn, filepath, ip, or subnet. These options are mutually exclusive.
>
> The tool will gracefully skip any IPs or domains that don't connect (e.g., unreachable or not running a TLS service) and continue scanning the rest. A message will be displayed for each skipped IP or domain.

Example Usage:

```bash
# Scan a single domain
tls-scrape --fqdn=www.google.com --outdir=./output

# Scan a list of domains from a CSV file
tls-scrape --filepath=./websites.csv --header=url --outdir=./output

# Scan a single IP address
tls-scrape --ip=192.168.1.1 --port=443 --outdir=./output

# Scan an entire subnet
tls-scrape --subnet=192.168.1.0/24 --port=443 --outdir=./output

# Bundle all output into a single file
tls-scrape --subnet=192.168.1.0/24 --port=443 --outdir=./output --bundle --prettyjson
```

## Docker
After building the Docker image using the Makefile, you can run it with:

```bash
docker run -e YOUR_ENV_VARIABLES scotta01/tls-scrape
```

> [!NOTE]
> Replace YOUR_ENV_VARIABLES with the necessary environment variables.

## Contributing

If you'd like to contribute to TLS Scrape, please fork the repository and use a feature branch. Pull requests are warmly welcome.

## Licensing

The code in this project is licensed under the MIT license. Please see the [LICENSE file](LICENSE) for more information.
