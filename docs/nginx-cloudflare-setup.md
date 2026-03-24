# Nginx + Cloudflare Setup with Wildcard SSL


![flexphish dashboard](/docs/nginx-cloudflare-setup.png)

This guide explains how to deploy **FlexPhish** behind **Nginx** with **Cloudflare DNS** and configure a **wildcard SSL certificate** using **Let's Encrypt (Certbot)**.
Using a wildcard certificate allows FlexPhish to dynamically generate campaign domains such as:

* `microsoft.redacted.com`
* `google.redacted.com`
* `netflix.redacted.com`

without needing to modify Nginx or generate new certificates.

---

The architecture will look like this:

```bash
User
  │
  ▼
HTTPS
  │
  ▼
Cloudflare
  │
  ▼
HTTPS
  │
  ▼
Nginx
  │
  ├── Dashboard :8000
  ├── API :8088
  └── Campaign :8001
```

Cloudflare provides **DNS, SSL, and protection**, while **Nginx acts as the reverse proxy** forwarding requests to the FlexPhish services.

---

# Requirements

Before starting, ensure the following:

- A **Linux server** (Ubuntu recommended)
- A **domain name**
- **Cloudflare account**
- **Nginx installed**
- **FlexPhish running on the server**

Example internal ports used by FlexPhish:

| Service | Port |
|-------|------|
| API | 8088 |
| Dashboard | 8000 |
| Campaign Server | 8001 |

---

# Install Nginx

Update the system and install Nginx:

```bash
sudo apt update
sudo apt install nginx -y
```

Start and enable the service:

```bash
sudo systemctl enable nginx
sudo systemctl start nginx
```

Check the status:

```bash
sudo systemctl status nginx
```

# Architecture Overview

```bash
Internet
   │
   ▼
Cloudflare
   │
   ▼
Nginx
   │
   ├── redacted.com      → Dashboard :8000
   │
   ├── api.redacted.com  → API :8088
   │
   └── *.redacted.com   → Campaign :8001
```

## Cloudflare DNS Setup

This section explains how to configure **Cloudflare DNS** so that traffic is routed correctly to your server and then forwarded by **Nginx** to the appropriate FlexPhish service.

### Architecture Overview

```bash
Internet
   │
   ▼
Cloudflare
   │
   ▼
Nginx
   │
   ├── redacted.com      → Dashboard :8000
   │
   ├── api.redacted.com  → API :8088
   │
   └── *.redacted.com    → Campaign :8001
```

In this architecture:

* **Cloudflare** handles DNS and traffic routing.
* **Nginx** acts as a reverse proxy.
* **FlexPhish services** run on internal ports.

---

# 1. Add Domain to Cloudflare

1. Log in to the **Cloudflare Dashboard**.
2. Click **Add a Site**.
3. Enter your domain:

```
redacted.com
```

4. Select a plan (Free plan is sufficient).
5. Continue to DNS configuration.

---

# 2. Configure DNS Records

Create the following DNS records inside the **DNS** section of the Cloudflare dashboard.

| Type | Name | Content   | Proxy   |
| ---- | ---- | --------- | ------- |
| A    | @    | SERVER_IP | Proxied |
| A    | api  | SERVER_IP | Proxied |
| A    | *    | SERVER_IP | Proxied |

Example:

| Type | Name             | Value   |
| ---- | ---------------- | ------- |
| A    | redacted.com     | 1.2.3.4 |
| A    | api.redacted.com | 1.2.3.4 |
| A    | *.redacted.com   | 1.2.3.4 |

Make sure the **orange cloud (Proxy Enabled)** is active.

This ensures traffic flows through **Cloudflare** before reaching your server.

---

# 3. Update Domain Nameservers

Cloudflare will provide **two nameservers**, for example:

```
anna.ns.cloudflare.com
mark.ns.cloudflare.com
```

Go to your **domain registrar** and replace the existing nameservers with the ones provided by Cloudflare.

DNS propagation may take several minutes to a few hours.

---

# 4. Verify DNS Configuration

After propagation, verify that the DNS records are active.

You can test using:

```bash
dig redacted.com
dig api.redacted.com
dig test.redacted.com
```

All of them should resolve to your **server IP**.

---

# 5. Configure SSL Mode

Inside the Cloudflare dashboard:

1. Go to **SSL/TLS**
2. Set encryption mode to:

```
Full (Strict)
```

This ensures encrypted communication between:

* User → Cloudflare
* Cloudflare → Nginx

---

# Result

After completing the configuration, the following domains will route correctly:

```
https://redacted.com            → Dashboard
https://api.redacted.com        → API
https://login.redacted.com      → Campaign
https://bank.redacted.com       → Campaign
https://anything.redacted.com   → Campaign
```

All traffic will follow this path:

```
Internet → Cloudflare → Nginx → FlexPhish services
```

The wildcard DNS record allows FlexPhish campaigns to dynamically generate **unlimited subdomains** without modifying DNS.


# Why Use DNS Challenge?

When your domain is managed by **Cloudflare**, there are two ways to generate Let's Encrypt certificates:

| Method         | Use Case                           |
| -------------- | ---------------------------------- |
| HTTP Challenge | Simple domains                     |
| DNS Challenge  | Required for wildcard certificates |

Since we want to generate:

```
*.redacted.com
```

we must use the **DNS Challenge method**.

---

# 1. Install Certbot

On Ubuntu / Debian:

```bash
sudo apt update
sudo apt install certbot python3-certbot-nginx python3-certbot-dns-cloudflare
```

---

# 2. Create Cloudflare API Token

In the **Cloudflare Dashboard**:

1. Go to **My Profile**
2. Open **API Tokens**
3. Click **Create Token**

Use the template:

```
Edit Zone DNS
```

Permissions:

```
Zone
DNS
Edit
```

Zone:

```
redacted.com
```

After creating the token, **copy it**.

---

# 3. Create Cloudflare Credentials File

Create a secure directory for secrets:

```bash
sudo mkdir -p /root/.secrets
```

Create the credentials file:

```bash
sudo nano /root/.secrets/cloudflare.ini
```

Content:

```
dns_cloudflare_api_token = YOUR_API_TOKEN
```

Set correct permissions (required):

```bash
sudo chmod 600 /root/.secrets/cloudflare.ini
```

---

# 4. Generate Wildcard SSL Certificate

Run Certbot using the Cloudflare DNS plugin:

```bash
sudo certbot certonly \
--dns-cloudflare \
--dns-cloudflare-credentials /root/.secrets/cloudflare.ini \
-d redacted.com \
-d "*.redacted.com"
```

This will generate the certificates:

```
/etc/letsencrypt/live/redacted.com/fullchain.pem
/etc/letsencrypt/live/redacted.com/privkey.pem
```

---

# 5. Configure Nginx

Edit your configuration file:

```
/etc/nginx/sites-available/flexphish
```

---

# HTTP → HTTPS Redirect

```nginx
server {
    listen 80;
    server_name redacted.com api.redacted.com *.redacted.com;

    return 301 https://$host$request_uri;
}
```

---

# Dashboard Server

```nginx
server {

    listen 443 ssl;
    server_name redacted.com;

    ssl_certificate /etc/letsencrypt/live/redacted.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/redacted.com/privkey.pem;

    location / {

        proxy_pass http://127.0.0.1:8000;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;

    }
}
```

---

# API Server

```nginx
server {

    listen 443 ssl;
    server_name api.redacted.com;

    ssl_certificate /etc/letsencrypt/live/redacted.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/redacted.com/privkey.pem;

    location / {

        proxy_pass http://127.0.0.1:8088;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;

    }
}
```

---

# Campaign Wildcard Server

```nginx
server {

    listen 443 ssl;
    server_name *.redacted.com;

    ssl_certificate /etc/letsencrypt/live/redacted.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/redacted.com/privkey.pem;

    location / {

        proxy_pass http://127.0.0.1:8001;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;

    }
}
```

---

# 6. Test Nginx Configuration

```bash
sudo nginx -t
```

---

# 7. Reload Nginx

```bash
sudo systemctl reload nginx
```

---

# 8. Automatic Certificate Renewal

Certbot automatically installs a cron job for renewal.

You can test the renewal process:

```bash
sudo certbot renew --dry-run
```

---

# Final Result

Your server will automatically support:

```
https://redacted.com
https://api.redacted.com
https://login.redacted.com
https://bank.redacted.com
https://anything.redacted.com
```

All domains will use a **valid Let's Encrypt SSL certificate**.

---

# FlexPhish Campaign Advantage

Using a wildcard certificate allows FlexPhish campaigns to generate unlimited subdomains such as:

```
microsoft.redacted.com
google.redacted.com
netflix.redacted.com
paypal.redacted.com
```

without:

* modifying Nginx
* generating new certificates
* restarting the server

This makes the campaign infrastructure **scalable and fully automated**.
