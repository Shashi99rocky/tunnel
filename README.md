# Go Tunnel (Self-hosted Ngrok Alternative)

A lightweight self-hosted tunneling solution written in Go. Expose your local web applications to the internet using your own VPS with wildcard HTTPS certificates from Let's Encrypt.

---

## Features
- 🌍 Public HTTPS endpoint for local services
- 🔐 Token-based authentication for security
- 🪄 Automatic TLS with Let's Encrypt (supports `*.<domain>`)
- 🔁 Auto-routing requests to correct client subdomain
- ⚡ Lightweight, no external dependencies
- 🛠 Systemd service support for server auto-start

---

## Requirements
- **VPS with a public IP**
- **Domain name** pointing to VPS (`A` record and `*.subdomain` record)
- **Go 1.20+** installed on both VPS and local machine

---

## Installation
### 1. Install Go
```bash
sudo apt update
sudo apt install golang -y
```

### 2. Build
```bash
go build -o tunnel tunnel.go
```

---

## Usage

### 🔹 Server (VPS)
Start server with your domain:
```bash
./tunnel server example.com
```

This will:
- Start an HTTPS server on port `443`
- Start a tunnel server on port `9000`
- Issue & auto-renew TLS certs for `*.example.com`

### 🔹 Client (Local Machine)
Run client to expose a local port:
```bash
./tunnel client your-vps-ip:9000 demo 3000 supersecrettoken123
```
- `your-vps-ip:9000` → Address of VPS tunnel server
- `demo` → Subdomain (`demo.example.com`)
- `3000` → Local port of your app (`http://localhost:3000`)
- `supersecrettoken123` → Token required for authentication

Now access your local app at:
```
https://demo.example.com
```

---

## Token Authentication
Edit `tunnel.go` to configure subdomains and tokens:
```go
var tokens = map[string]string{
    "demo": "supersecrettoken123",
    "test": "othertoken456",
}
```
Only clients with valid tokens can connect.

---

## Systemd Service (Server)
### Create Service File
```bash
sudo nano /etc/systemd/system/tunnel.service
```

Paste:
```ini
[Unit]
Description=Go Tunnel Server
After=network.target

[Service]
User=root
WorkingDirectory=/root/tunnel
ExecStart=/root/tunnel/tunnel server example.com
Restart=always
RestartSec=5
Environment=GODEBUG=tls13=1

[Install]
WantedBy=multi-user.target
```

### Enable & Start
```bash
sudo systemctl daemon-reexec
sudo systemctl daemon-reload
sudo systemctl enable tunnel
sudo systemctl start tunnel
```

### Logs
```bash
sudo journalctl -u tunnel -f
```

---

## Roadmap
- [ ] Client systemd unit for auto-start
- [ ] Web dashboard for managing tunnels
- [ ] Multi-user support with database
- [ ] Traffic monitoring & rate limiting

---

## License
MIT License
