# OIDC Identity Lab — End-to-End Tutorial

This tutorial walks through the OpenID Connect (OIDC) lab implemented in this repository using **Keycloak** as the identity provider and a **Go web application** as the client.

The goal of this lab is to demonstrate how a real application can delegate authentication to an identity provider, complete the Authorization Code Flow with PKCE, verify identity information, create an authenticated session, and perform logout through the provider.

---

# Overview

OpenID Connect (OIDC) is an identity layer built on top of OAuth 2.0.

In this lab, OIDC is used to:

- authenticate a user with Keycloak
- return an authorization code to the application
- exchange that code for tokens
- verify the returned ID token
- retrieve user identity information from the UserInfo endpoint
- create a local authenticated session
- log the user out through Keycloak

This mirrors a common enterprise identity pattern used in modern IAM systems.

---

# Architecture

```text
User
  │
  │  Visits application
  ▼
Go Web Application
  │
  │  Redirects user to Keycloak
  ▼
Keycloak
  │
  │  Authenticates user
  │  Returns authorization code
  ▼
Go Web Application
  │
  │  Exchanges code for tokens
  │  Verifies ID token
  │  Retrieves UserInfo claims
  │  Creates session
  ▼
Protected Pages
```

---

# Prerequisites

Required software:

- Go
- Docker
- Web browser

Optional:

- curl
- Git

---

# Project Directory

```text
oidc-identity-lab/
├── application/
│   ├── main.go
│   ├── go.mod
│   ├── go.sum
│   └── .env
├── docs/
│   └── tutorial.md
├── LICENSE
└── README.md
```

---

# Step 1 — Start Keycloak

Run Keycloak locally with Docker:

```bash
docker run -p 8080:8080 \
  -e KEYCLOAK_ADMIN=admin \
  -e KEYCLOAK_ADMIN_PASSWORD=admin \
  quay.io/keycloak/keycloak:latest start-dev
```

Then open:

```text
http://localhost:8080/admin/
```

Log in with:

```text
admin / admin
```

---

# Step 2 — Create the Realm

In the Keycloak admin console:

1. Click the realm dropdown in the upper-left
2. Click **Create realm**
3. Enter: `oidc-lab`
4. Save

Make sure you are working inside the `oidc-lab` realm, not `master`.

---

# Step 3 — Create the OIDC Client

Inside the `oidc-lab` realm:

1. Click **Clients**
2. Click **Create client**
3. Set **Client type** to **OpenID Connect**
4. Set **Client ID** to `oidc-lab-app`
5. Continue

Configure the client with these settings:

- **Client authentication**: On
- **Authorization**: Off
- **Standard flow**: On
- **Implicit flow**: Off
- **Direct access grants**: Off

Then configure:

- **Valid redirect URIs**: `http://localhost:3000/callback`
- **Valid post logout redirect URIs**: `http://localhost:3000/`
- **Web origins**: `http://localhost:3000`

Save the client.

Then open the **Credentials** tab and copy the client secret.

---

# Step 4 — Create a Test User

Inside the `oidc-lab` realm:

1. Click **Users**
2. Click **Create new user**
3. Enter:
   - Username: `testuser`
   - First name: `Test`
   - Last name: `User`
   - Email: `testuser@example.com`
4. Make sure **Enabled** is on
5. Save

Then:

1. Open the new user
2. Click **Credentials**
3. Set a password
4. Turn **Temporary** off
5. Save

This user will be used to log into the Go application.

---

# Step 5 — Configure the Application

Create a `.env` file in the `application` directory with:

```env
OIDC_CLIENT_SECRET=your_client_secret_here
SESSION_SECRET=your_session_secret_here
KEYCLOAK_LOGOUT_REDIRECT=http://localhost:3000/
```

Replace the values with your actual client secret and session secret.

---

# Step 6 — Run the Application

From the `application` directory:

```bash
go mod tidy
go run main.go
```

The app should start at:

```text
http://localhost:3000
```

---

# Step 7 — Test the Login Flow

Open:

```text
http://localhost:3000/login
```

Then:

1. Log in with `testuser`
2. Approve the login flow
3. Return to the application
4. Confirm the profile page loads

At this point, the application should:
- exchange the authorization code for tokens
- verify the ID token
- call the UserInfo endpoint
- store authenticated session data

---

# Step 8 — View Claims

After login, open the profile and claims pages.

The project shows:

- ID token claims
- UserInfo claims

This helps demonstrate how identity attributes are returned and consumed by an application.

Examples include:

- name
- username
- email

---

# Step 9 — Test Logout

After logging in, click:

```text
Logout
```

A successful logout should:

1. clear the application session
2. call the Keycloak logout endpoint
3. redirect back to:

```text
http://localhost:3000/
```

Then test logging in again to confirm the logout worked.

---

# Key Security Concepts Demonstrated

This lab demonstrates:

- delegated authentication
- centralized identity management
- Authorization Code Flow
- PKCE protection for the code exchange
- verification of identity claims
- separation of authentication from application logic
- federated logout behavior

---

# Example OIDC Flow Summary

```text
1. User opens /login
2. Application redirects to Keycloak
3. User authenticates
4. Keycloak redirects back with authorization code
5. Application exchanges code for tokens
6. Application verifies ID token
7. Application retrieves UserInfo
8. Application creates authenticated session
9. User accesses protected pages
10. User logs out through Keycloak
```

---

# Result

By completing this tutorial, you demonstrate practical understanding of:

- OpenID Connect
- OAuth 2.0 Authorization Code Flow
- PKCE
- token and claim handling
- Keycloak client configuration
- session-based authenticated web applications

These are directly relevant to enterprise IAM, SSO, and identity engineering work.

---

# Next Steps

Possible future enhancements:

- route protection middleware
- role-based access control
- group claim mapping
- refresh token support
- improved UI styling
- additional identity provider integrations

---

# License

MIT License