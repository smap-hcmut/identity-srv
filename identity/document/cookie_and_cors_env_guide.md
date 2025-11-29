## Cookie & CORS Configuration by Environment

This document explains how to configure **CORS** and **HttpOnly cookies** for the Identity service in two main scenarios.

**Important**: In both scenarios below, the **backend API is always**:

- `https://smap-api.tantai.dev/identity/...`  (never running on localhost)

Only the **frontend changes**:

- Dev: `http://localhost:3000`
- Prod: `https://smap.tantai.dev`

The behavior is controlled by:

- `ENV` (environment mode) – affects **CORS**
- `COOKIE_*` variables – affect **cookie domain / security attributes**

---

## 1. Dev: Frontend Localhost → Backend smap-api.tantai.dev

**Goal**: Developer ở nhà, frontend chạy:

- Frontend: `http://localhost:3000`
- Backend API (K8s): `https://smap-api.tantai.dev/identity/...`

Browser should:

- Accept `smap_auth_token` cookie from `smap-api.tantai.dev`
- Send cookie trên các request tiếp theo từ localhost (XHR/fetch)

### 1.1. Environment Variables (Dev Cluster)

Trong `identity/k8s/configmap.yaml` cho môi trường dev:

```yaml
data:
  # Environment mode (enables localhost + private subnets CORS)
  ENV: "dev"

  # Cookie configuration for DEV
  # Backend host luôn là smap-api.tantai.dev
  COOKIE_DOMAIN: ".tantai.dev"
  COOKIE_SECURE: "true"

  # Frontend là localhost (khác site với smap-api.tantai.dev),
  # để cookie đi kèm trong XHR/fetch thì cần:
  COOKIE_SAMESITE: "None"

  COOKIE_NAME: "smap_auth_token"
  COOKIE_MAX_AGE: "7200"
  COOKIE_MAX_AGE_REMEMBER: "2592000"
```

**Vì sao như vậy:**

- `ENV=dev`:
  - CORS middleware cho phép origin `http://localhost:3000`
  - Set `Access-Control-Allow-Credentials: true` để browser cho gửi cookie
- `COOKIE_DOMAIN=.tantai.dev`:
  - Cookie hợp lệ cho host `smap-api.tantai.dev` (parent domain)
- `COOKIE_SAMESITE=None` + `COOKIE_SECURE=true`:
  - Bắt buộc khi **frontend localhost** gọi API trên **domain khác**
  - Nếu để `Lax`, cookie sẽ không đi trong XHR từ `localhost:3000`

### 1.2. Frontend Configuration (Dev)

**Axios example:**

```javascript
import axios from 'axios';

const api = axios.create({
  baseURL: 'https://smap-api.tantai.dev/identity',
  withCredentials: true, // REQUIRED: send/receive cookies
});

// Login
await api.post('/authentication/login', {
  email: 'user@example.com',
  password: 'password123',
});

// Get current user
const me = await api.get('/authentication/me');
```

**Fetch example:**

```javascript
await fetch('https://smap-api.tantai.dev/identity/authentication/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  credentials: 'include', // REQUIRED
  body: JSON.stringify({ email, password }),
});

const me = await fetch('https://smap-api.tantai.dev/identity/authentication/me', {
  credentials: 'include',
});
```

---

## 2. Production: smap.tantai.dev ↔ smap-api.tantai.dev

**Goal**: Frontend at `https://smap.tantai.dev` calls Identity API at:

- `https://smap-api.tantai.dev/identity/authentication/...`

Browser should:

- Accept `smap_auth_token` cookie from `smap-api.tantai.dev`
- Send cookie on subsequent API calls
- Enforce secure attributes (Secure, SameSite)

### 2.1. Domain Topology

Current domains:

- Frontend: `smap.tantai.dev`
- API: `smap-api.tantai.dev`

Common parent domain: `tantai.dev`

**Important**:

- `COOKIE_DOMAIN=.smap.tantai.dev` is **NOT** valid for `smap-api.tantai.dev`  
  (they are siblings: `smap` vs `smap-api`)
- Recommended cookie domain for sharing across both is: `COOKIE_DOMAIN=.tantai.dev`

### 2.2. Environment Variables for Production

In Kubernetes **ConfigMap** for production Identity API:

```yaml
data:
  # Environment Configuration
  ENV: "production"

  # Cookie Configuration (shared across smap.tantai.dev & smap-api.tantai.dev)
  COOKIE_DOMAIN: ".tantai.dev"
  COOKIE_SECURE: "true"
  COOKIE_SAMESITE: "Lax"
  COOKIE_MAX_AGE: "7200"
  COOKIE_MAX_AGE_REMEMBER: "2592000"
  COOKIE_NAME: "smap_auth_token"
```

**Behavior:**

- API host: `smap-api.tantai.dev`
- Cookie domain: `.tantai.dev` (parent) → browser accepts cookie
- Cookie is sent on requests to `https://smap-api.tantai.dev`
- Frontend at `https://smap.tantai.dev` can also read authentication state via API calls

### 2.3. Environment Variables for Remote Dev / Staging (Cluster)

For a dev/staging environment on the same cluster:

```yaml
data:
  ENV: "dev"  # or "staging"

  COOKIE_DOMAIN: ".tantai.dev"
  COOKIE_SECURE: "true"

  # When testing from http://localhost:3000 to https://smap-api.tantai.dev,
  # SameSite=Lax will NOT send cookies with XHR/fetch.
  # To allow fully cross-site dev testing, set:
  COOKIE_SAMESITE: "None"
```

**Notes:**

- `SameSite=Lax` is good default for production (protects against CSRF)
- For **local frontend (`localhost`) calling remote API (`smap-api.tantai.dev`)**:
  - This is cross-site
  - To send cookies on XHR/fetch, **browser requires**:
    - `SameSite=None`
    - `Secure=true`
  - So for dev/staging where this pattern is needed, set:

    ```yaml
    COOKIE_SAMESITE: "None"
    COOKIE_SECURE: "true"
    ```

---

## 3. Summary: How to Switch Between Dev & Prod

### 3.1. Dev (frontend localhost, backend smap-api.tantai.dev)

- **Backend**:
  - `ENV=dev`
  - `COOKIE_DOMAIN=.tantai.dev`
  - `COOKIE_SECURE=true`
  - `COOKIE_SAMESITE=None`  // cho phép cross-site XHR từ localhost
- **Frontend**:
  - Base URL: `https://smap-api.tantai.dev/identity`
  - `withCredentials: true` / `credentials: 'include'`

### 3.2. Staging (tương tự dev, có thể dùng domain riêng)

- **Backend (K8s ConfigMap)**:
  - `ENV=dev` (or `staging`)
  - `COOKIE_DOMAIN=.tantai.dev`
  - `COOKIE_SECURE=true`
  - `COOKIE_SAMESITE=None` (for cross-site XHR from localhost or other test domains)
- **Frontend**:
  - Base URL: `https://smap-api.tantai.dev/identity`
  - `withCredentials: true` / `credentials: 'include'`

### 3.3. Production (smap.tantai.dev ↔ smap-api.tantai.dev)

- **Backend (K8s ConfigMap)**:
  - `ENV=production`
  - `COOKIE_DOMAIN=.tantai.dev`
  - `COOKIE_SECURE=true`
  - `COOKIE_SAMESITE=Lax` (recommended)
- **Frontend**:
  - Base URL: `https://smap-api.tantai.dev/identity`
  - `withCredentials: true` / `credentials: 'include'`

---

## 4. Troubleshooting Checklist

If cookies are not set or sent:

- Check **Network → login response → Set-Cookie**:
  - Is `Domain` valid for the API host?
  - Is `SameSite` compatible with your frontend origin?
  - Is `Secure` set correctly (must be true when SameSite=None)?
- Check **Network → /authentication/me request**:
  - Does the request header include `Cookie: smap_auth_token=...`?
  - Are CORS headers present?
    - `Access-Control-Allow-Origin`
    - `Access-Control-Allow-Credentials: true`


