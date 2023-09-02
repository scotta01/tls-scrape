# TLS Scrape Library and Tool
[![Go Reference](https://pkg.go.dev/badge/github.com/scotta01/tls-scrape.svg)](https://pkg.go.dev/github.com/scotta01/tls-scrape)
TLS Scrape is both a Go library and a CLI tool designed to scrape websites for TLS certificate details and log relevant metrics.

## Features

- **Library**:
  - Scrape domains for TLS details programmatically.
  - Check OCSP status of certificates.
  - Capture and retrieve scraping metrics.

- **CLI Tool**:
  - Scrape individual domains or lists from CSV files for TLS details.
  - Expose metrics related to scraping process for Prometheus monitoring.
  - Log all scraped details.
  - Dockerized for easy deployment.

## Getting Started

### Prerequisites

- Golang (tested with version `1.20.7-bullseye`)
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
- **outfile**: Output path if you wish to save the results as a JSON file.
- **concurrency**: Maximum number of concurrent TLS connections. Default is 10.

> [!NOTE]  
> Only provide either fqdn or (filepath and header). Both can't be provided together.

Example Usage:

```bash
tls-scrape --fqdn=www.google.com --outfile=./google.json
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
