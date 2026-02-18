# OpenID Connect Identity Lab

## Overview

This project demonstrates OpenID Connect (OIDC) authentication using an identity provider to securely authenticate users and issue JSON Web Tokens (JWTs). The lab simulates enterprise Single Sign-On (SSO) and modern identity-based authentication workflows.

This implementation demonstrates how applications delegate authentication to a trusted identity provider and validate cryptographically signed tokens to authorize access.

---

## Architecture

Authentication flow:

```text
User
  │
  │  Requests access to protected application
  ▼
Application
  │
  │  Redirects to Identity Provider
  ▼
Identity Provider (OIDC)
  │
  │  Authenticates user and issues JWT
  ▼
Application
  │
  │  Validates JWT signature and claims
  ▼
Access Granted
```

---

## Features

• OpenID Connect authentication  
• Identity provider configuration  
• JWT token issuance  
• JWT signature validation  
• Protected application endpoints  
• Secure delegated authentication  

---

## Technologies Used

OpenID Connect (OIDC)  
OAuth 2.0  
JSON Web Tokens (JWT)  
Keycloak (or Authentik)  
TLS / HTTPS  

---

## Security Benefits

• Eliminates need for application-managed passwords  
• Centralized identity management  
• Cryptographically signed authentication tokens  
• Secure Single Sign-On (SSO)  
• Reduced credential exposure risk  

---

## Token Example

Example JWT structure:

```text
Header.Payload.Signature
```

Decoded payload example:

```json
{
  "sub": "user123",
  "iss": "https://identity-provider",
  "aud": "application",
  "exp": 1710000000,
  "iat": 1709996400
}
```

---

## Learning Objectives

This project demonstrates understanding of:

• OpenID Connect authentication flow  
• OAuth 2.0 authorization concepts  
• JWT structure and validation  
• Identity provider integration  
• Secure delegated authentication  

---

## Repository Structure

```text
oidc-identity-lab/
├── identity-provider/
├── application/
├── tokens/
├── LICENSE
└── README.md
```

---

## Future Improvements

• Multi-factor authentication integration  
• Role-based access control (RBAC)  
• Token refresh implementation  
• Integration with enterprise identity providers (Azure AD, Okta)  

---

## Author

Stephanie Shinn  
IAM Consultant | PKI | Identity Security | Quantum-Safe Cryptography  

---

## License

MIT License
