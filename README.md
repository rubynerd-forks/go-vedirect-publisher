# go-vedirect-publisher
A utility for publishing [VE.Direct](https://www.victronenergy.com/live/vedirect_protocol:faq) frames over MQTT. With support for embedded devices (Teltonika RUT)

This program is for reading data from a Victron device using the VE.Direct protocol.
It's built for the Teltonika RUT955 Router, however likely can be compiled for any device which Go can be compiled for. The RUT955 has a MIPS processor which runs RutOS, a fork of OpenWrt.

The two devices are connected together using the [VE.Direct to USB interface](https://www.victronenergy.com/accessories/ve-direct-to-usb-interface)

go-vedirect-publisher has currently been tested with Mosquitto MQTT Server and AWS IoT Core

### Example payload

```
{"CS":"0","ERR":"0","FW":"150","H19":"325","H20":"0","H21":"0","H22":"0","H23":"0","HSDS":"213","I":"-230","IL":"200","LOAD":"ON","MPPT":"0","OR":"0x00000001","PID":"0xA053","PPV":"0","SER#":"XXXXXXXXXXX","V":"13210","VPV":"10","timestamp":"1605565193"}
```

## Features
- Send payload to an MQTT Server
- Save payload to file
- Multi-device support via repeated `-dev` flags
- Custom MQTT fields via `--extras` parameter
- Version information included in MQTT messages

## Installation

### Download Pre-built Binaries

Download the latest release for your platform from the [Releases page](https://github.com/rubynerd-forks/go-vedirect-publisher/releases).

**Supported platforms:**
- Linux (x86_64, ARM64, ARMv6, ARMv7, MIPS)
- macOS (x86_64, ARM64)

**Example:**
```bash
# Download Linux AMD64 version
curl -LO https://github.com/rubynerd-forks/go-vedirect-publisher/releases/latest/download/go-vedirect-publisher_v0.1.0_Linux_x86_64.tar.gz

# Verify checksum
curl -LO https://github.com/rubynerd-forks/go-vedirect-publisher/releases/latest/download/checksums.txt
sha256sum -c checksums.txt --ignore-missing

# Extract
tar xzf go-vedirect-publisher_v0.1.0_Linux_x86_64.tar.gz

# Run
./go-vedirect-publisher -v
```

### Build from Source

```bash
git clone https://github.com/rubynerd-forks/go-vedirect-publisher
cd go-vedirect-publisher
go build
```

**Requirements:**
- Go 1.25 or later

## Usage
```
Usage of ./bin/go-vedirect:
  -dev string
		full path to serial device node (default "/dev/ttyUSB0")
  -mqtt.server string
		MQTT Server address (default "tcp://localhost:1883")
  -mqtt.tls_cert string
		MQTT TLS Private Cert
  -mqtt.tls_key string
		MQTT TLS Private Key
  -mqtt.tls_rootca string
		MQTT TLS Root CA
  -mqtt.topic string
		The MQTT Topic to publish messages to
  -out-file string
		File to write json data to
  -v	Print Version
  -verbose
		Verbose Output
```

## Releases

This project uses semantic versioning and automated releases via GitHub Actions.

### Creating a Release

Maintainers can create a release by tagging a commit:

```bash
git tag -a v0.2.0 -m "Release v0.2.0: Description of changes"
git push origin v0.2.0
```

The GitHub Actions workflow will automatically:
1. Run tests and security scans
2. Build binaries for all platforms
3. Generate SBOM (Software Bill of Materials)
4. Create GitHub release with artifacts
5. Generate changelog

### Security

All releases include:
- **Checksums** (SHA256) for all binaries
- **SBOM** in SPDX and CycloneDX formats
- **Vulnerability scanning** results
- **Reproducible builds** (verifiable via ldflags)

### Version Information

Each binary includes version information accessible via `-v` flag:

```bash
./go-vedirect-publisher -v
# Output:
# v0.1.0
# commit: abc123def456
# built at: 2026-02-01T22:30:15Z
```

Version information is also included in all MQTT messages:
- `publisher_version`: Semantic version (e.g., "v0.1.0")
- `publisher_commit`: Git commit hash
- `publisher_build_date`: Build timestamp

## Roadmap
- Tests
- Configurable sample rate, vedirect updates every second
- Send payload to HTTP endpoint
- Rewrite in better language for embedded systems (C++)
