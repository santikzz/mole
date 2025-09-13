# Mole

> **Version:** 0.0.1-beta  
> **Status:** Development - Active Beta

A free, open-source tunneling solution that provides secure HTTP/HTTPS tunnels from localhost to custom domains. Mole serves as a self-hosted alternative to ngrok with full control over your infrastructure and domains.

---

## Features

- **Custom Domain Support** - Use your own domains and subdomains
- **Automatic SSL** - Let's Encrypt integration with auto-renewal
- **Docker Ready** - One-command deployment with Docker Compose
- **Verbose Logging** - Comprehensive request/response logging to `mole.log`
- **Self-Hosted** - Complete control over your tunneling infrastructure
- **Lightweight** - Written in Go for optimal performance
- **Cross-Platform** - Runs on Linux, macOS, and Windows

## Project Status

Mole is currently in active development. While functional, it should be considered beta software. We welcome contributions and feedback from the community to help improve the project.

## Installation

### Prerequisites

- Docker and Docker Compose
- A server with a public IP address
- Domain name with DNS access

### Docker Deployment (Recommended)

Clone the repository:

```bash
git clone https://github.com/santikzz/mole.git
cd mole
```

Copy and configure environment variables:

```bash
cp .env.example .env
```

Edit `.env` with your settings:

```bash
MOLE_DOMAIN=yourdomain.com
MOLE_EMAIL=your@email.com
MOLE_USE_HTTPS=true
```

Start the server:

```bash
docker-compose up -d
```

### Building from Source

For development or custom builds:

```bash
make all
```

This creates binaries in the `bin/` directory:
- `mole-server` - The tunnel server
- `mole` - The client application

## Quick Start

### 1. Server Setup

**Docker (Recommended)**:

```bash
# Clone and configure
git clone https://github.com/santikzz/mole.git
cd mole
cp .env.example .env

# Edit .env with your domain and email
echo "MOLE_DOMAIN=yourdomain.com" > .env
echo "MOLE_EMAIL=your@email.com" >> .env
echo "MOLE_USE_HTTPS=true" >> .env

# Start server with automatic SSL
docker-compose up -d
```

**Manual Build**:

Create a configuration file:

```json
{
    "port": 3000,
    "domain": "example.com",
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

### Environment Variables

Configure the server using `.env` file:

| Variable | Description | Default |
|----------|-------------|----------|
| `MOLE_PORT` | Server listening port | `3000` |
| `MOLE_DOMAIN` | Base domain for tunnels | Required |
| `MOLE_EMAIL` | Email for Let's Encrypt | Required for HTTPS |
| `MOLE_USE_HTTPS` | Enable HTTPS with auto SSL | `false` |

**Example `.env`:**

```bash
MOLE_DOMAIN=tunnel.yourdomain.com
MOLE_EMAIL=admin@yourdomain.com
MOLE_USE_HTTPS=true
MOLE_PORT=3000
```

### Legacy Configuration

For manual builds, use `config.json` or command-line flags:

```bash
./bin/mole-server -port 8080 -domain mydomain.com -https
```

### SSL Certificate Management

#### Automatic SSL (Docker - Recommended)

When using Docker with `MOLE_USE_HTTPS=true`, SSL certificates are automatically:
- **Generated** using Let's Encrypt on first startup
- **Renewed** automatically via cron job
- **Managed** internally with persistent volumes

No manual certificate setup required!

#### Manual SSL Setup

For manual deployments, install certbot:

```bash
sudo apt install certbot
```

Obtain certificates:

```bash
sudo certbot certonly --manual --preferred-challenges dns -d example.com -d *.example.com
```

Certificates are automatically detected at `/etc/letsencrypt/live/example.com/`.

## Logging and Monitoring

### Verbose Logging

The server provides comprehensive logging for debugging and monitoring:

- **Request Logging**: All HTTP requests with method, path, client IP, and user agent
- **Response Tracking**: Request duration and completion status
- **Error Logging**: Detailed error messages for failed requests
- **WebSocket Events**: Tunnel connection and disconnection events

### Log File Access

**Docker Deployment**:
Logs are automatically saved to `mole.log` and persisted in Docker volumes:

```bash
# View live logs
docker-compose logs -f mole-server

# Access log file directly
docker exec -it mole_mole-server_1 tail -f /var/log/mole.log
```

**Manual Deployment**:
Logs are output to stdout/stderr and can be redirected:

```bash
./bin/mole-server > mole.log 2>&1
```

### Log Format Example

```
[REQUEST] GET /tunnel from 172.17.0.1:45678 - User-Agent: Mozilla/5.0...
[RESPONSE] GET /tunnel completed in 1.2ms
[ERROR] Invalid method GET for /response endpoint
[RESPONSE] Handling response for request ID: abc123
```

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