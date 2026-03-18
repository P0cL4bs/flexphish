# Local Deployment (Development Mode)

This guide explains how to run **FlexPhish locally** using `/etc/hosts`.
This setup is ideal for **development, testing templates, and creating campaigns locally** without requiring a public domain.

With this configuration you will be able to generate campaigns such as:

```
microsoft.localhost:8001
google.localhost:8001
login.localhost:8001
```

---

# Architecture (Local)

```bash
Browser
   │
   ▼
localhost
   │
   ├── localhost:8000        → Dashboard
   │
   ├── localhost:8088        → API
   │
   └── *.localhost:8001      → Campaign Server
```

---

# 1. Configure `/etc/hosts`

FlexPhish uses **subdomain campaigns**, so we configure the base domain as `localhost`.

Edit the hosts file:

```bash
sudo nano /etc/hosts
```

Example configuration:

```bash
# /etc/hosts
127.0.0.1   localhost
```

Since `localhost` resolves to `127.0.0.1`, subdomains like:

```
microsoft.localhost
google.localhost
login.localhost
```

will also resolve locally in most browsers.

---

# 2. FlexPhish Configuration

FlexPhish configuration is stored in:

```
configs/app.yaml
```

Example configuration:

```yaml
campaign:
  base_domain: localhost:8001
  subdomain_mode: true

security:
  jwt_secret: "GENERATED AUTO"
  test_mode_token: "GENERATED AUTO"

server:
  host: 0.0.0.0
  api_port: 8088
  dashboard_port: 8000
  campaign_port: 8001

session:
  cookie_name: fp_session
  cookie_domain: ""
  cookie_secure: false
  cookie_http_only: true
  ttl: 24h

template_dir: templates
template_assets_dir: templates/assets
```

---

# Configuration Explanation

### Campaign

```yaml
campaign:
  base_domain: localhost:8001
  subdomain_mode: true
```

This allows FlexPhish to dynamically generate campaign URLs such as:

```
http://microsoft.localhost:8001
http://google.localhost:8001
```

---

### Server

```yaml
server:
  host: 0.0.0.0
  api_port: 8088
  dashboard_port: 8000
  campaign_port: 8001
```

Ports used by FlexPhish:

| Service         | Port |
| --------------- | ---- |
| Dashboard       | 8000 |
| API             | 8088 |
| Campaign Server | 8001 |

---

### Sessions

```yaml
session:
  cookie_name: fp_session
  cookie_domain: ""
  cookie_secure: false
  cookie_http_only: true
  ttl: 24h
```

For **local development**, `cookie_secure` is disabled because HTTPS is not used.

---

### Templates

```yaml
template_dir: templates
template_assets_dir: templates/assets
```

These directories store:

* phishing templates
* static assets (CSS, JS, images)

---

# 3. Start FlexPhish

Run the application:

```bash
./flexphish
```

Or specify configuration manually:

```bash
./flexphish -config configs/app.yaml
```

---

# 4. Access the Services

Once running, you can access:

### Dashboard

```
http://localhost:8000
```

### API

```
http://localhost:8088
```

### Campaign Example

```
http://microsoft.localhost:8001
```

---

# 5. Creating a User

FlexPhish allows user management directly from the **command line**.

To create a new user:

```bash
./flexphish -create-user \
-email admin@example.com \
-password StrongPassword 
```

Example:

```bash
./flexphish -create-user \
-email admin@localhost \
-password admin123 
```

Available roles:

```
admin
user
```

---

# 6. Deleting a User

To delete a user:

```bash
./flexphish -delete-user -email admin@localhost
```

---

# 7. CLI Options

FlexPhish provides several command-line options:

```
-api-port int
    API port (default 8088)

-campaign-port int
    Campaign port (default 8001)

-config string
    Config file path (default "configs/app.yaml")

-create-user
    Create a new user

-dashboard
    Start the dashboard server (default true)

-dashboard-port int
    Dashboard port (default 8000)

-db string
    Database path (default "flexphish.db")

-delete-user
    Delete a user

-dev
    Enable development mode (default true)

-email string
    User email

-host string
    Server host

-password string
    User password

-role string
    User role (admin/user)
```

---

# Example: Start Server with Custom Ports

```bash
./flexphish \
-dashboard-port 9000 \
-api-port 9090 \
-campaign-port 9001
```

---

# Example Campaign URL

With the default configuration, FlexPhish will generate campaign links like:

```
http://microsoft.localhost:8001
http://google.localhost:8001
http://login.localhost:8001
```

These subdomains simulate **real phishing campaign infrastructure** while running entirely on your local machine.

---

# Recommended Workflow

1. Configure `/etc/hosts`
2. Configure `configs/app.yaml`
3. Start FlexPhish
4. Create an admin user
5. Access the dashboard
6. Create templates and campaigns

---

# Next Steps

For production deployments see:

```
docs/nginx-cloudflare-setup.md
```
