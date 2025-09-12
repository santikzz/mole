# Mole

> **Version:** 0.0.1-beta  
> **Status:** Development - Active Beta

A free, open-source tunneling solution that provides secure HTTP/HTTPS tunnels from localhost to custom domains. Mole serves as a self-hosted alternative to ngrok with full control over your infrastructure and domains.

---

## Features

- **Custom Domain Support** - Use your own domains and subdomains
- **HTTP/HTTPS Support** - Full SSL/TLS encryption with Let's Encrypt integration
- **Self-Hosted** - Complete control over your tunneling infrastructure
- **Lightweight** - Written in Go for optimal performance
- **Cross-Platform** - Runs on Linux, macOS, and Windows

## Project Status

Mole is currently in active development. While functional, it should be considered beta software. We welcome contributions and feedback from the community to help improve the project.

## Installation

### Prerequisites

- Go 1.19 or higher
- A server with a public IP address
- Domain name with DNS access

### Building from Source

Clone the repository and build the binaries:

```bash
git clone https://github.com/your-username/mole.git
cd mole
make all
```

This creates two binaries in the `bin/` directory:
- `mole-server` - The tunnel server
- `mole` - The client application

## Quick Start

### 1. Server Setup

Create a configuration file on your server:

```json
{
    "port": 3000,
    "domain": "example.com",
    "cert_file": "/etc/letsencrypt/live/example.com/fullchain.pem",
    "key_file": "/etc/letsencrypt/live/example.com/privkey.pem",
    "use_https": true
}
```

Start the server:

```bash
./bin/mole-server
```

### 2. Client Usage

Expose a local service running on port 8000:

```bash
./bin/mole http 8000
```

Use a custom subdomain:

```bash
./bin/mole http 8000 -d myapp
```

This makes your local service available at `myapp.example.com`.

## Configuration

### Server Configuration

The server can be configured via `config.json` or command-line flags:

| Option | Flag | Description |
|--------|------|-------------|
| `port` | `-port` | Server listening port |
| `domain` | `-domain` | Base domain for tunnels |
| `use_https` | `-https` | Enable HTTPS support |
| `cert_file` | `-cert` | Path to SSL certificate |
| `key_file` | `-key` | Path to SSL private key |

**Example with command-line flags:**

```bash
./bin/mole-server -port 8080 -domain mydomain.com -https -cert /path/cert.pem -key /path/key.pem
```

### SSL Certificate Setup

#### Using Let's Encrypt (Recommended)

Install certbot:

```bash
sudo apt install certbot
```

Obtain certificates for your domain and wildcard subdomain:

```bash
sudo certbot certonly --manual --preferred-challenges dns -d example.com -d *.example.com
```

Certificates will be stored in `/etc/letsencrypt/live/example.com/`. Update your server configuration with the paths to `fullchain.pem` and `privkey.pem`.

## DNS Configuration

Configure your domain's DNS to point to your server. Choose one of the following approaches:

### Wildcard DNS (Recommended)

Create a wildcard A record that routes all subdomains to your server:

```
Type: A
Name: *
Value: YOUR_SERVER_IP
TTL: 300
```

### Individual Subdomains

Create separate A records for each subdomain:

```
Type: A
Name: api
Value: YOUR_SERVER_IP
TTL: 300

Type: A
Name: web
Value: YOUR_SERVER_IP
TTL: 300
```

## Contributing

We welcome contributions! Please feel free to submit issues, feature requests, and pull requests.

### Development Setup

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International License. You are free to use, modify, and distribute this software for non-commercial purposes. See the [LICENSE](LICENSE) file for details.

## Support

- **Issues**: Report bugs or request features via [GitHub Issues](https://github.com/your-username/mole/issues)
- **Documentation**: Check the project wiki for detailed guides
- **Community**: Join discussions in the project's GitHub Discussions

---

**Disclaimer**: This software is provided as-is under active development. Use in production environments at your own discretion.