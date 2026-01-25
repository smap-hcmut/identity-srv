# Frontend Migration Guide: HttpOnly Cookie Authentication

> **Complete guide for migrating frontend applications from LocalStorage JWT to HttpOnly cookie authentication**

---

## Table of Contents

1. [Overview](#overview)
2. [Breaking Changes](#breaking-changes)
3. [Migration Steps](#migration-steps)
4. [Framework-Specific Guides](#framework-specific-guides)
5. [Testing Your Migration](#testing-your-migration)
6. [Common Issues](#common-issues)
7. [Mobile App Considerations](#mobile-app-considerations)

---

## Overview

### What Changed?

The SMAP Identity Service has migrated from LocalStorage-based JWT tokens to HttpOnly cookies for enhanced security.

**Before** (LocalStorage):
```javascript
// Login
const { data } = await api.post('/login', credentials);
localStorage.setItem('token', data.token);  // Vulnerable to XSS

// Authenticated requests
api.defaults.headers.common['Authorization'] = `Bearer ${localStorage.getItem('token')}`;
```

**After** (HttpOnly Cookie):
```javascript
// Login
await api.post('/login', credentials);
// Cookie set automatically by browser (XSS-proof)

// Authenticated requests
// Cookie sent automatically (no manual management)
```

### Why This Change?

- **XSS Protection**: HttpOnly cookies cannot be accessed by JavaScript
- **Simplified Code**: No manual token storage/retrieval
- **Better Security**: Secure flag ensures HTTPS-only transmission
- **CSRF Protection**: SameSite=Lax provides built-in protection

---

## Breaking Changes

### 1. Login Response No Longer Contains Token

**Before**:
```json
{
  "code": 200,
  "data": {
    "user": { "id": "...", "email": "..." },
    "token": "eyJhbGc..."
  }
}
```

**After**:
```json
{
  "code": 200,
  "data": {
    "user": { "id": "...", "email": "..." }
  }
}
```
**Plus** `Set-Cookie` header: `smap_auth_token=eyJhbGc...; HttpOnly; Secure; SameSite=Lax`

### 2. You Must Set `withCredentials: true`

All HTTP requests to the API **must** include credentials:

```javascript
// Axios
const api = axios.create({
  baseURL: 'https://smap-api.tantai.dev',
  withCredentials: true
});

// Fetch
fetch('https://smap-api.tantai.dev/endpoint', {
  credentials: 'include'
});
```

### 3. New Logout Endpoint Required

**Before**:
```javascript
// Just clear localStorage
localStorage.removeItem('token');
```

**After**:
```javascript
// Call logout endpoint to expire cookie
await api.post('/authentication/logout');
```

### 4. New `/me` Endpoint for User Info

**Before**:
```javascript
// Decode JWT client-side
const token = localStorage.getItem('token');
const decoded = jwtDecode(token);
```

**After**:
```javascript
// Call /me endpoint
const { data } = await api.get('/authentication/me');
const user = data.data;
```

---

## Migration Steps

### Step 1: Update HTTP Client Configuration

#### Axios

```javascript
// src/api/client.js
import axios from 'axios';

const api = axios.create({
  baseURL: process.env.REACT_APP_API_URL || 'https://smap-api.tantai.dev/identity',
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
});

export default api;
```

#### Fetch

```javascript
// src/api/client.js
const baseURL = process.env.REACT_APP_API_URL || 'https://smap-api.tantai.dev/identity';

export const fetchAPI = async (endpoint, options = {}) => {
  const response = await fetch(`${baseURL}${endpoint}`, {
    ...options,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
  });

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }

  return response.json();
};
```

### Step 2: Remove Token Storage Logic

**Remove** all localStorage/sessionStorage operations:

```javascript
// DELETE THESE
localStorage.setItem('token', token);
localStorage.getItem('token');
localStorage.removeItem('token');
sessionStorage.setItem('token', token);
```

### Step 3: Remove Authorization Header Injection

**Remove** manual Authorization header setting:

```javascript
// DELETE THIS
api.defaults.headers.common['Authorization'] = `Bearer ${token}`;

// DELETE THIS
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});
```

### Step 4: Update Login Logic

**Before**:
```javascript
const login = async (email, password) => {
  const response = await api.post('/authentication/login', { email, password });
  const { token, user } = response.data.data;

  // Remove these
  localStorage.setItem('token', token);
  api.defaults.headers.common['Authorization'] = `Bearer ${token}`;

  return user;
};
```

**After**:
```javascript
const login = async (email, password, remember = false) => {
  const response = await api.post('/authentication/login', {
    email,
    password,
    remember  // Optional: extends cookie lifetime to 30 days
  });

  const { user } = response.data.data;
  // Cookie is set automatically by browser

  return user;
};
```

### Step 5: Update Logout Logic

**Before**:
```javascript
const logout = () => {
  localStorage.removeItem('token');  // Remove
  delete api.defaults.headers.common['Authorization'];  // Remove
};
```

**After**:
```javascript
const logout = async () => {
  try {
    await api.post('/authentication/logout');
    // Cookie is expired by server
  } catch (error) {
    console.error('Logout failed:', error);
  }
};
```

### Step 6: Update Authentication State Management

**Before**:
```javascript
const checkAuth = () => {
  const token = localStorage.getItem('token');
  if (!token) return false;

  try {
    const decoded = jwtDecode(token);
    return decoded.exp > Date.now() / 1000;
  } catch {
    return false;
  }
};
```

**After**:
```javascript
const checkAuth = async () => {
  try {
    const response = await api.get('/authentication/me');
    return response.data.data;  // Returns user object
  } catch (error) {
    return null;  // Not authenticated
  }
};
```

### Step 7: Update Protected Route Logic

**React Router Example**:

**Before**:
```javascript
const PrivateRoute = ({ children }) => {
  const token = localStorage.getItem('token');  // Remove

  if (!token) {
    return <Navigate to="/login" />;
  }

  return children;
};
```

**After**:
```javascript
const PrivateRoute = ({ children }) => {
  const [isAuthenticated, setIsAuthenticated] = useState(null);

  useEffect(() => {
    api.get('/authentication/me')
      .then(() => setIsAuthenticated(true))
      .catch(() => setIsAuthenticated(false));
  }, []);

  if (isAuthenticated === null) {
    return <div>Loading...</div>;
  }

  return isAuthenticated ? children : <Navigate to="/login" />;
};
```

---

## Framework-Specific Guides

### React + Axios

#### Complete Auth Context

```javascript
// src/contexts/AuthContext.jsx
import React, { createContext, useState, useContext, useEffect } from 'react';
import api from '../api/client';

const AuthContext = createContext(null);

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  // Check authentication status on mount
  useEffect(() => {
    checkAuth();
  }, []);

  const checkAuth = async () => {
    try {
      const response = await api.get('/authentication/me');
      setUser(response.data.data);
    } catch (error) {
      setUser(null);
    } finally {
      setLoading(false);
    }
  };

  const login = async (email, password, remember = false) => {
    const response = await api.post('/authentication/login', {
      email,
      password,
      remember
    });
    setUser(response.data.data.user);
    return response.data.data.user;
  };

  const logout = async () => {
    try {
      await api.post('/authentication/logout');
    } finally {
      setUser(null);
    }
  };

  const value = {
    user,
    loading,
    login,
    logout,
    checkAuth,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
};
```

#### Usage in Components

```javascript
// src/pages/LoginPage.jsx
import { useAuth } from '../contexts/AuthContext';

const LoginPage = () => {
  const { login } = useAuth();
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      await login(email, password, remember);
      navigate('/dashboard');
    } catch (error) {
      console.error('Login failed:', error);
    }
  };

  return (/* login form */);
};
```

### Vue 3 + Axios

#### Auth Store (Pinia)

```javascript
// src/stores/auth.js
import { defineStore } from 'pinia';
import api from '../api/client';

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null,
    loading: false,
  }),

  getters: {
    isAuthenticated: (state) => !!state.user,
  },

  actions: {
    async checkAuth() {
      try {
        const response = await api.get('/authentication/me');
        this.user = response.data.data;
      } catch (error) {
        this.user = null;
      }
    },

    async login(email, password, remember = false) {
      this.loading = true;
      try {
        const response = await api.post('/authentication/login', {
          email,
          password,
          remember
        });
        this.user = response.data.data.user;
      } finally {
        this.loading = false;
      }
    },

    async logout() {
      try {
        await api.post('/authentication/logout');
      } finally {
        this.user = null;
      }
    },
  },
});
```

### Angular + HttpClient

#### Auth Service

```typescript
// src/app/services/auth.service.ts
import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject, Observable } from 'rxjs';
import { tap } from 'rxjs/operators';

interface User {
  id: string;
  email: string;
  full_name?: string;
  role: string;
}

@Injectable({ providedIn: 'root' })
export class AuthService {
  private userSubject = new BehaviorSubject<User | null>(null);
  public user$ = this.userSubject.asObservable();

  private baseUrl = 'https://smap-api.tantai.dev/identity';

  constructor(private http: HttpClient) {
    this.checkAuth();
  }

  checkAuth(): void {
    this.http.get<{ data: User }>(`${this.baseUrl}/authentication/me`, {
      withCredentials: true
    }).subscribe({
      next: (response) => this.userSubject.next(response.data),
      error: () => this.userSubject.next(null),
    });
  }

  login(email: string, password: string, remember = false): Observable<any> {
    return this.http.post(`${this.baseUrl}/authentication/login`, {
      email,
      password,
      remember
    }, {
      withCredentials: true
    }).pipe(
      tap((response: any) => this.userSubject.next(response.data.user))
    );
  }

  logout(): Observable<any> {
    return this.http.post(`${this.baseUrl}/authentication/logout`, {}, {
      withCredentials: true
    }).pipe(
      tap(() => this.userSubject.next(null))
    );
  }
}
```

#### HTTP Interceptor (Global withCredentials)

```typescript
// src/app/interceptors/credentials.interceptor.ts
import { Injectable } from '@angular/core';
import { HttpInterceptor, HttpRequest, HttpHandler } from '@angular/common/http';

@Injectable()
export class CredentialsInterceptor implements HttpInterceptor {
  intercept(req: HttpRequest<any>, next: HttpHandler) {
    const clonedRequest = req.clone({
      withCredentials: true
    });
    return next.handle(clonedRequest);
  }
}
```

### Next.js (App Router)

#### Server Actions

```typescript
// src/app/actions/auth.ts
'use server';

import { cookies } from 'next/headers';

export async function login(email: string, password: string, remember = false) {
  const response = await fetch('https://smap-api.tantai.dev/identity/authentication/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password, remember }),
    credentials: 'include',
  });

  if (!response.ok) {
    throw new Error('Login failed');
  }

  const data = await response.json();
  return data.data.user;
}

export async function getCurrentUser() {
  const response = await fetch('https://smap-api.tantai.dev/identity/authentication/me', {
    credentials: 'include',
    cache: 'no-store',
  });

  if (!response.ok) {
    return null;
  }

  const data = await response.json();
  return data.data;
}
```

---

## Testing Your Migration

### Checklist

- [ ] Login successfully sets cookie
- [ ] Cookie is visible in DevTools (Application → Cookies)
- [ ] Cookie attributes are correct (HttpOnly, Secure, SameSite)
- [ ] Authenticated requests work without manual headers
- [ ] Logout expires cookie
- [ ] `/me` endpoint returns user info
- [ ] Protected routes redirect when not authenticated
- [ ] No CORS errors in console
- [ ] No token stored in localStorage/sessionStorage

### Browser DevTools Verification

1. Login and open DevTools (F12)
2. Go to Application → Cookies → https://smap-api.tantai.dev
3. Verify `smap_auth_token` cookie exists with:
   - HttpOnly: ✓
   - Secure: ✓
   - SameSite: Lax
   - Expires: ~2 hours from now (or 30 days with Remember Me)

4. Go to Network tab
5. Make an authenticated request
6. Check Request Headers:
   - Should see: `Cookie: smap_auth_token=<JWT>`
7. Should NOT see: `Authorization: Bearer ...`

---

## Common Issues

### Issue 1: CORS Error "Credentials not supported if the CORS header 'Access-Control-Allow-Origin' is '*'"

**Cause**: Backend using wildcard origin with credentials enabled.

**Solution**: Backend must specify exact origins:
```javascript
// Backend CORS config
AllowedOrigins: ["https://app.smap.com", "http://localhost:3000"]
AllowCredentials: true
```

### Issue 2: Cookie Not Being Sent

**Cause**: Missing `withCredentials: true` or `credentials: 'include'`.

**Solution**:
```javascript
// Axios
const api = axios.create({
  withCredentials: true
});

// Fetch
fetch(url, {
  credentials: 'include'
});
```

### Issue 3: Login Succeeds But No Cookie in DevTools

**Cause**: Domain mismatch or Secure flag on HTTP.

**Debug**:
- Check if frontend and backend are on same parent domain
- If testing locally, use `http://localhost` (not `http://127.0.0.1`)
- Or set `COOKIE_SECURE=false` for local development

### Issue 4: 401 Unauthorized on All Requests

**Cause**: Cookie domain mismatch or not being sent.

**Debug**:
```javascript
// Check if cookie is being sent
// Network tab → Request → Headers → Cookie
// Should see: smap_auth_token=<JWT>
```

### Issue 5: User State Not Persisting on Page Refresh

**Cause**: Not checking auth on app mount.

**Solution**:
```javascript
useEffect(() => {
  checkAuth();  // Call /me endpoint on mount
}, []);
```

---

## Mobile App Considerations

### React Native with Axios

```javascript
import axios from 'axios';
import CookieManager from '@react-native-cookies/cookies';

const api = axios.create({
  baseURL: 'https://smap-api.tantai.dev/identity',
  withCredentials: true,
});

// Login
const login = async (email, password) => {
  const response = await api.post('/authentication/login', { email, password });
  // Cookies handled automatically by @react-native-cookies/cookies
  return response.data.data.user;
};
```

### Flutter with HTTP Package

```dart
import 'package:http/http.dart' as http;
import 'package:cookie_jar/cookie_jar.dart';
import 'package:dio/dio.dart';
import 'package:dio_cookie_manager/dio_cookie_manager.dart';

final dio = Dio();
final cookieJar = CookieJar();
dio.interceptors.add(CookieManager(cookieJar));

Future<void> login(String email, String password) async {
  final response = await dio.post(
    'https://smap-api.tantai.dev/identity/authentication/login',
    data: {'email': email, 'password': password},
  );
  // Cookies managed by dio_cookie_manager
}
```

---

## Migration Timeline

### Phase 1: Update Frontend (Week 1-2)
- Update HTTP client configuration
- Remove token storage logic
- Update login/logout logic
- Update auth state management
- Test in development environment

### Phase 2: Deploy to Staging (Week 2)
- Deploy updated frontend
- Test with real backend
- Verify CORS configuration
- Test all authentication flows

### Phase 3: Production Deployment (Week 3)
- Deploy to production
- Monitor error logs
- Collect user feedback
- Have rollback plan ready

### Phase 4: Cleanup (Week 4+)
- Remove old token-based code
- Update documentation
- Remove localStorage fallback code

---

## Summary

### Required Changes

1. Set `withCredentials: true` in HTTP client
2. Remove localStorage token management
3. Remove Authorization header injection
4. Call `/authentication/logout` endpoint
5. Use `/authentication/me` for user info
6. Update authentication state management

### Benefits After Migration

- Enhanced XSS protection
- Simplified authentication code
- Automatic cookie management
- Better security compliance
- CSRF protection via SameSite

---

## Support

If you encounter issues during migration:

1. Check [TESTING.md](./TESTING.md) for testing procedures
2. Review [Common Issues](#common-issues) section
3. Verify CORS configuration on backend
4. Check browser DevTools for cookie presence

**Remember**: The Authorization header fallback is maintained during migration period, so existing clients continue working while you update.
