# OpenID Connect Identity Lab

## Overview

This project demonstrates a working OpenID Connect (OIDC) login flow using **Keycloak** as the identity provider and a **Go web application** as the relying party.

The lab shows how an application can delegate authentication to an identity provider, securely handle the authorization code flow, validate the returned ID token, retrieve UserInfo claims, maintain authenticated session state, and perform logout through the provider.

This project was built as a practical IAM learning lab to better understand modern authentication patterns used in enterprise identity environments.

---

## What This Project Demonstrates

- OpenID Connect login with Keycloak
- OAuth 2.0 Authorization Code Flow
- PKCE (Proof Key for Code Exchange)
- State parameter validation
- ID token verification
- UserInfo retrieval
- Session-based authenticated user state
- Protected application pages
- Logout using the Keycloak end-session endpoint
- Display of ID token and UserInfo claims

---

## Authentication Flow

User
  │
  │  Visits application
  ▼
Go Web Application
  │
  │  Redirects user to Keycloak login
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
  │  Creates authenticated session
  ▼
Protected Application Pages

---

## Features

- Login through Keycloak
- Authorization Code Flow with PKCE
- Session-based access to protected routes
- Profile page showing authenticated user information
- Raw claims page for inspection and debugging
- Logout flow integrated with Keycloak
- Local development setup using `localhost`

---

## Technologies Used

- Go
- Keycloak
- OpenID Connect (OIDC)
- OAuth 2.0
- PKCE
- JSON Web Tokens (JWT)
- Gorilla Sessions

---

## Security Concepts Demonstrated

This lab demonstrates several important IAM and application security concepts:

- Delegated authentication
- Centralized identity management
- Avoiding application-managed passwords
- Authorization Code Flow instead of Implicit Flow
- PKCE protection for the authorization code exchange
- Verification of cryptographically signed ID tokens
- Use of standards-based identity protocols
- Secure logout through the identity provider

---

## Project Structure

oidc-identity-lab/
├── application/
│   ├── main.go
│   ├── go.mod
│   ├── go.sum
│   └── .env
├── LICENSE
└── README.md

---

## Example Pages

The application currently includes:

- `/` — Home page
- `/login` — Starts OIDC login flow
- `/profile` — Displays user profile information from claims
- `/claims` — Displays raw ID token and UserInfo claims
- `/logout` — Logs the user out of the application and Keycloak

---

## Local Configuration

This lab uses a local Keycloak instance and a local Go application.

### Keycloak Realm and Client

Example values used in this project:

- Realm: `oidc-lab`
- Client ID: `oidc-lab-app`
- Redirect URI: `http://localhost:3000/callback`
- Post logout redirect URI: `http://localhost:3000/`

### Example `.env`

OIDC_CLIENT_SECRET=your_client_secret_here
SESSION_SECRET=your_session_secret_here
KEYCLOAK_LOGOUT_REDIRECT=http://localhost:3000/

---

## Running the Application

From the `application` directory:

go mod tidy
go run main.go

The app runs at:

`http://localhost:3000`

Make sure Keycloak is running locally on:

`http://localhost:8080`

---

## Test Flow

A simple local validation flow:

1. Start Keycloak
2. Start the Go application
3. Open `http://localhost:3000/login`
4. Authenticate with the test user in Keycloak
5. Confirm the profile page loads
6. View the claims page
7. Click logout
8. Confirm the session is cleared and Keycloak logout succeeds

---

## Learning Objectives

This project demonstrates practical understanding of:

- OpenID Connect authentication flows
- OAuth 2.0 authorization code handling
- PKCE and state validation
- JWT verification and claims usage
- Identity provider integration
- Session handling in web applications
- Logout behavior in federated authentication systems

---

## Future Improvements

Potential future enhancements include:

- Route protection middleware
- Role-based access control (RBAC)
- Group and role claim mapping
- Refresh token handling
- Improved UI styling
- Support for additional identity providers such as Okta or Microsoft Entra ID
- Containerized lab setup with Docker Compose

---

## Author

**Stephanie Shinn**  
IAM Consultant | PKI | Identity Security | Quantum-Safe Cryptography

---

## License

MIT License