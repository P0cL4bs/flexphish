# Templates

FlexPhish templates define the **structure and behavior of phishing pages** used in campaigns.

Templates are defined using **YAML configuration files** combined with **HTML pages and static assets**.
A template describes **how a phishing flow works**, including:

* the pages shown to the user
* the data captured
* validation rules
* redirects
* dynamic variables
* client-side scripts
* behavioral hooks

Each template can simulate **multi-step authentication flows** such as:

* login → password
* login → password → 2FA
* email → password → OTP

---

# Template Structure

A typical template contains:

```
templates/
    assets/
    └── okta_login/
        ├── username.html
        ├── password.html
        └── static/
            ├── style.css
            ├── script.js
            └── logo.png
    ├── okta-login.yaml
```

Where:

| File            | Description                      |
| --------------- | -------------------------------- |
| `okta-login.yaml` | Defines the phishing flow        |
| `.html` files   | Pages displayed to the user      |
| `static/`       | CSS, JS, images and other assets |

---

# Template YAML Example

Example template definition:

```yaml
info:
  name: Okta Login
  author: mh4x0f
  description: A phishing template mimicking the example login page.
  category: corporate
  system: false
  tags:
    - okta
    - mobile
    - web

template_dir: okta_login

steps:
  - id: username
    title: Okta Username Page
    path: /login
    method: POST
    template_file: username.html
    next: password
    capture:
      enabled: true
      fields:
        - name: username
          required: true
          validate_regex: ^\S+$
          error_message: This field cannot be left blank
    simulate_delay_ms: 300

  - id: password
    title: Okta Password Page
    path: /verify
    method: POST
    template_file: password.html
    redirect_url: https://www.okta.com
    capture:
      enabled: true
      fields:
        - name: password
          required: true
          validate_regex: ^.{6,}$
          error_message: Unable to sign in

hooks:
  onLoad:
    - /hooks/fingerprint.js
    - /hooks/behavior.js
```

---

# Template Metadata (`info`)

The `info` section describes the template.

```yaml
info:
  name: Okta Login
  author: mh4x0f
  description: A phishing template mimicking the example login page.
  category: corporate
  system: false
  tags:
    - okta
    - mobile
    - web
```

| Field         | Description                     |
| ------------- | ------------------------------- |
| `name`        | Template display name           |
| `author`      | Template creator                |
| `description` | Template description            |
| `category`    | Template category               |
| `system`      | Reserved for internal templates |
| `tags`        | Searchable tags                 |

---

# Steps

The `steps` section defines the **phishing flow**.

Each step represents **one page in the authentication process**.

Example:

```yaml
steps:
  - id: username
    path: /login
    template_file: username.html
```

Steps are executed sequentially using:

```
next: password
```

Or finalized using:

```
redirect_url: https://example.com
```

---

# Step Fields

| Field           | Description     |
| --------------- | --------------- |
| `id`            | Step identifier |
| `title`         | Step title      |
| `path`          | URL path        |
| `method`        | HTTP method     |
| `template_file` | HTML file       |
| `next`          | Next step       |
| `redirect_url`  | Final redirect  |

---

# Capturing Data

Steps can capture user input using `capture`.

Example:

```yaml
capture:
  enabled: true
  fields:
    - name: username
      required: true
      validate_regex: ^\S+$
      error_message: This field cannot be left blank
```

Field options:

| Field            | Description         |
| ---------------- | ------------------- |
| `name`           | Form field name     |
| `required`       | Required field      |
| `validate_regex` | Regex validation    |
| `error_message`  | Error shown to user |

---

# Multi-Step Data Access

Captured values are stored in the **session state**.

They can be accessed from other steps using:

```
{{ .Vars.FIELD_NAME }}
```

Example:

```html
<div class="email">
    <span class="font-bold">{{ .Vars.username }}</span>
</div>
```

This allows pages like **password pages** to display the previously captured email.

---

# Step Variables

Steps can define custom variables accessible in HTML templates.

Example:

```yaml
vars:
  logo_url: /static/logo.png
  page_title: ""
  target_email: user@example.com
```

Access them in HTML:

```html
<img src="{{ .Vars.logo_url }}">
<h1>{{ .Vars.page_title }}</h1>
```

---

# Global Variables

Templates may also define variables accessible to **all steps**.

Example:

```yaml
global_vars:
  company_name: ExampleCorp
  support_email: support@example.com
```

Use them inside templates:

```html
<footer>
Contact {{ .Vars.support_email }}
</footer>
```

---

# Hooks

Hooks allow injecting **JavaScript files** into templates automatically.

Example:

```yaml
hooks:
  onLoad:
    - /hooks/fingerprint.js
    - /hooks/behavior.js
```

Inside HTML:

```html
{{ if .Hooks }}
{{ range .Hooks }}
<script src="{{ . }}"></script>
{{ end }}
{{ end }}
```

Hooks can be used for:

* browser fingerprinting
* behavior tracking
* anti-bot detection
* analytics

---

# HTML Templates

Each step loads an HTML file defined by:

```
template_file: login.html
```

Example login page:

```html
<form id="login-form" method="POST" action="/login">
    <input type="text" name="username" class="form-control" placeholder="Email">
    <button type="submit">Login</button>
</form>
```

---

# Static Files

Templates can include static assets such as:

* CSS
* JavaScript
* Images
* Fonts

Example usage:

```html
<link rel="stylesheet" href="/static/style.css">
<script src="/static/script.js"></script>
<img src="/static/logo.png">
```

These files are served automatically by the template engine.

---

# JavaScript Form Handler

Every template should include a **form submission handler** to interact with the FlexPhish backend.

Example script:

```html
<script>
const form = document.getElementById('login-form');
const errorDiv = document.getElementById('error-message');

form.addEventListener('submit', async (e) => {
    e.preventDefault();

    const formData = new URLSearchParams();
    for (const [key, value] of new FormData(form)) {
        formData.append(key, value);
    }

    const response = await fetch(form.action, {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: formData.toString(),
    });

    const data = await response.json();

    if (data.redirect) {
        window.location.href = data.redirect;
        return;
    }

    if (!data.success) {
        errorDiv.textContent = data.error || "Login failed.";
        errorDiv.classList.remove('d-none');
    }
});
</script>
```

This script:

1. Prevents normal form submission
2. Sends data using `fetch`
3. Receives JSON responses
4. Handles validation errors
5. Redirects users when needed

---

# Example Multi-Step Template

### Step 1 – Login

```html
<form id="login-form" method="POST" action="/login">

<input type="text" name="username" placeholder="Email" required>

<button type="submit">Next</button>

</form>
```

---

### Step 2 – Password

```html
<div class="email-display">
{{ .Vars.username }}
</div>

<form id="password-form" method="POST" action="/verify">

<input type="password" name="password" placeholder="Password">

<button type="submit">Sign in</button>

</form>
```

---

### Step 3 – 2FA Verification

```html
<form id="verify-form" method="POST" action="/verify">

<input type="text" name="sms_code" placeholder="Enter verification code">

<button type="submit">Verify</button>

</form>
```

---

# Template Execution Flow

Example authentication flow:

```
username step
   ↓
password step
   ↓
2FA step
   ↓
redirect to legitimate site
```

Each step:

1. renders an HTML page
2. captures user input
3. validates fields
4. stores session data
5. proceeds to the next step

---

# Summary

FlexPhish templates provide a **powerful and flexible phishing framework** capable of simulating real authentication flows.

Features include:

* multi-step login flows
* field validation
* dynamic variables
* reusable templates
* JavaScript hooks
* static assets
* session data sharing

Templates allow creating **highly realistic phishing campaigns** with minimal configuration.
