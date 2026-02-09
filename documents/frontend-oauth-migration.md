# Frontend OAuth Migration Guide

## Table of Contents

1. [Overview](#overview)
2. [Migration Summary](#migration-summary)
3. [Login Page Migration](#login-page-migration)
4. [Axios Configuration](#axios-configuration)
5. [Authentication State Management](#authentication-state-management)
6. [Error Handling](#error-handling)
7. [Logout Implementation](#logout-implementation)
8. [Role-Based UI Rendering](#role-based-ui-rendering)
9. [Testing Checklist](#testing-checklist)
10. [Troubleshooting](#troubleshooting)

---

## Overview

This guide provides step-by-step instructions for migrating the frontend from email/password authentication to Google OAuth2 SSO.

### What's Changing

**Before (Email/Password)**:
- User enters email and password
- Token stored in localStorage
- Manual token management in axios

**After (OAuth2)**:
- User clicks "Login with Google"
- Token stored in HttpOnly cookie (automatic)
- No manual token management needed

### Key Benefits

- **Better Security**: HttpOnly cookies prevent XSS attacks
- **Better UX**: Single Sign-On with Google
- **Simpler Code**: No token management in JavaScript
- **Centralized Auth**: Managed by Identity Service

---

## Migration Summary

### Changes Required

1. ✅ **Remove**: Email/password login form
2. ✅ **Add**: "Login with Google" button
3. ✅ **Remove**: localStorage token management
4. ✅ **Update**: Axios configuration (withCredentials: true)
5. ✅ **Update**: Authentication state management
6. ✅ **Add**: Error interceptors for 401/403
7. ✅ **Update**: Logout flow
8. ✅ **Add**: Role-based UI rendering

### No Changes Required

- ✅ Axios base configuration (already using cookies)
- ✅ CORS configuration (already configured)
- ✅ API endpoint calls (no changes needed)

---

## Login Page Migration

### Step 1: Remove Old Login Form

**Before** (`src/pages/Login.jsx`):

```jsx
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';

function Login() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      const response = await axios.post('/api/auth/login', {
        email,
        password,
      });
      
      // Store token in localStorage
      localStorage.setItem('token', response.data.token);
      
      // Redirect to dashboard
      navigate('/dashboard');
    } catch (err) {
      setError(err.response?.data?.message || 'Login failed');
    }
  };

  return (
    <div className="login-page">
      <form onSubmit={handleSubmit}>
        <input
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="Email"
        />
        <input
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="Password"
        />
        <button type="submit">Login</button>
        {error && <div className="error">{error}</div>}
      </form>
    </div>
  );
}
```

### Step 2: Add OAuth Login Button

**After** (`src/pages/Login.jsx`):

```jsx
import React from 'react';

function Login() {
  const handleGoogleLogin = () => {
    // Redirect to Identity Service OAuth endpoint
    // The service will handle OAuth flow and redirect back
    window.location.href = `${import.meta.env.VITE_API_URL}/authentication/login?redirect=${encodeURIComponent(window.location.origin + '/dashboard')}`;
  };

  return (
    <div className="login-page">
      <div className="login-container">
        <h1>Welcome to SMAP</h1>
        <p>Sign in with your Google Workspace account</p>
        
        <button 
          onClick={handleGoogleLogin}
          className="google-login-button"
        >
          <img src="/google-icon.svg" alt="Google" />
          Sign in with Google
        </button>
      </div>
    </div>
  );
}

export default Login;
```

**Styling** (`src/pages/Login.css`):

```css
.login-page {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-container {
  background: white;
  padding: 3rem;
  border-radius: 12px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.1);
  text-align: center;
  max-width: 400px;
}

.login-container h1 {
  margin-bottom: 0.5rem;
  color: #333;
}

.login-container p {
  color: #666;
  margin-bottom: 2rem;
}

.google-login-button {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  width: 100%;
  padding: 12px 24px;
  background: white;
  border: 2px solid #ddd;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 500;
  color: #333;
  cursor: pointer;
  transition: all 0.2s;
}

.google-login-button:hover {
  background: #f8f9fa;
  border-color: #4285f4;
  box-shadow: 0 2px 8px rgba(66, 133, 244, 0.2);
}

.google-login-button img {
  width: 24px;
  height: 24px;
}
```

### Step 3: Handle OAuth Callback (Optional)

If you want to show a loading state during OAuth:

```jsx
// src/pages/OAuthCallback.jsx
import React, { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

function OAuthCallback() {
  const navigate = useNavigate();

  useEffect(() => {
    // OAuth callback is handled by Identity Service
    // Cookie is already set, just redirect to dashboard
    navigate('/dashboard');
  }, [navigate]);

  return (
    <div className="oauth-callback">
      <div className="spinner"></div>
      <p>Completing sign in...</p>
    </div>
  );
}

export default OAuthCallback;
```

---

## Axios Configuration

### Step 1: Verify Axios Configuration

**Check** (`src/api/axios.js`):

```javascript
import axios from 'axios';

// Create axios instance
const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
  withCredentials: true, // ✅ CRITICAL: Must be true for cookies
  headers: {
    'Content-Type': 'application/json',
  },
});

export default api;
```

**Important**: `withCredentials: true` is REQUIRED for HttpOnly cookies to work.

### Step 2: Remove Token Management

**Before** (with localStorage):

```javascript
// ❌ Remove this code
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});
```

**After** (with cookies):

```javascript
// ✅ No token management needed!
// Cookies are sent automatically by the browser
```

### Step 3: Verify CORS Configuration

**Backend CORS** (already configured in Identity Service):

```go
// internal/middleware/cors.go
corsConfig := middleware.DefaultCORSConfig(environment)
// Allows credentials (cookies) from configured origins
```

**Frontend** (no changes needed):

```javascript
// Axios already configured with withCredentials: true
// CORS headers handled by backend
```

---

## Authentication State Management

### Step 1: Create Auth Context

**Create** (`src/contexts/AuthContext.jsx`):

```jsx
import React, { createContext, useContext, useState, useEffect } from 'react';
import api from '../api/axios';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  // Fetch current user on mount
  useEffect(() => {
    fetchCurrentUser();
  }, []);

  const fetchCurrentUser = async () => {
    try {
      setLoading(true);
      const response = await api.get('/authentication/me');
      setUser(response.data);
      setError(null);
    } catch (err) {
      if (err.response?.status === 401) {
        // Not authenticated
        setUser(null);
      } else {
        setError(err.response?.data?.message || 'Failed to fetch user');
      }
    } finally {
      setLoading(false);
    }
  };

  const logout = async () => {
    try {
      await api.post('/authentication/logout');
      setUser(null);
      window.location.href = '/login';
    } catch (err) {
      console.error('Logout failed:', err);
      // Force logout even if API call fails
      setUser(null);
      window.location.href = '/login';
    }
  };

  const value = {
    user,
    loading,
    error,
    isAuthenticated: !!user,
    logout,
    refreshUser: fetchCurrentUser,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}
```

### Step 2: Wrap App with AuthProvider

**Update** (`src/main.jsx` or `src/App.jsx`):

```jsx
import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import { AuthProvider } from './contexts/AuthContext';
import App from './App';

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <BrowserRouter>
      <AuthProvider>
        <App />
      </AuthProvider>
    </BrowserRouter>
  </React.StrictMode>
);
```

### Step 3: Use Auth Context in Components

**Example** (`src/components/Header.jsx`):

```jsx
import React from 'react';
import { useAuth } from '../contexts/AuthContext';

function Header() {
  const { user, isAuthenticated, logout } = useAuth();

  if (!isAuthenticated) {
    return null;
  }

  return (
    <header className="app-header">
      <div className="header-left">
        <h1>SMAP</h1>
      </div>
      
      <div className="header-right">
        <div className="user-info">
          <img 
            src={user.avatar_url} 
            alt={user.name}
            className="user-avatar"
          />
          <div className="user-details">
            <span className="user-name">{user.name}</span>
            <span className="user-role">{user.role}</span>
          </div>
        </div>
        
        <button onClick={logout} className="logout-button">
          Logout
        </button>
      </div>
    </header>
  );
}

export default Header;
```

### Step 4: Protected Routes

**Create** (`src/components/ProtectedRoute.jsx`):

```jsx
import React from 'react';
import { Navigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

function ProtectedRoute({ children, requiredRole }) {
  const { user, loading, isAuthenticated } = useAuth();

  if (loading) {
    return (
      <div className="loading-screen">
        <div className="spinner"></div>
        <p>Loading...</p>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  // Check role if required
  if (requiredRole) {
    const roleHierarchy = { ADMIN: 3, ANALYST: 2, VIEWER: 1 };
    const userRoleLevel = roleHierarchy[user.role] || 0;
    const requiredRoleLevel = roleHierarchy[requiredRole] || 0;

    if (userRoleLevel < requiredRoleLevel) {
      return (
        <div className="permission-denied">
          <h1>403 - Permission Denied</h1>
          <p>You don't have permission to access this page.</p>
          <p>Required role: {requiredRole}</p>
          <p>Your role: {user.role}</p>
        </div>
      );
    }
  }

  return children;
}

export default ProtectedRoute;
```

**Use in Routes** (`src/App.jsx`):

```jsx
import { Routes, Route } from 'react-router-dom';
import ProtectedRoute from './components/ProtectedRoute';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import Projects from './pages/Projects';
import AdminPanel from './pages/AdminPanel';

function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      
      <Route
        path="/dashboard"
        element={
          <ProtectedRoute>
            <Dashboard />
          </ProtectedRoute>
        }
      />
      
      <Route
        path="/projects"
        element={
          <ProtectedRoute requiredRole="ANALYST">
            <Projects />
          </ProtectedRoute>
        }
      />
      
      <Route
        path="/admin"
        element={
          <ProtectedRoute requiredRole="ADMIN">
            <AdminPanel />
          </ProtectedRoute>
        }
      />
    </Routes>
  );
}

export default App;
```

---

## Error Handling

### Step 1: Add Axios Interceptors

**Update** (`src/api/axios.js`):

```javascript
import axios from 'axios';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Response interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response) {
      const { status, data } = error.response;

      // Handle 401 Unauthorized - redirect to login
      if (status === 401) {
        console.error('Unauthorized - redirecting to login');
        window.location.href = '/login';
        return Promise.reject(error);
      }

      // Handle 403 Forbidden - show permission denied
      if (status === 403) {
        console.error('Permission denied:', data);
        // You can show a toast notification here
        return Promise.reject(error);
      }

      // Handle 429 Too Many Requests - rate limited
      if (status === 429) {
        console.error('Rate limited - too many requests');
        // Show rate limit message
        return Promise.reject(error);
      }
    }

    return Promise.reject(error);
  }
);

export default api;
```

### Step 2: Display Error Messages

**Create** (`src/components/ErrorBoundary.jsx`):

```jsx
import React from 'react';

class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error) {
    return { hasError: true, error };
  }

  componentDidCatch(error, errorInfo) {
    console.error('Error caught by boundary:', error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="error-page">
          <h1>Something went wrong</h1>
          <p>{this.state.error?.message}</p>
          <button onClick={() => window.location.reload()}>
            Reload Page
          </button>
        </div>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;
```

### Step 3: Toast Notifications for Errors

**Install** toast library:

```bash
npm install react-hot-toast
```

**Setup** (`src/main.jsx`):

```jsx
import { Toaster } from 'react-hot-toast';

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <BrowserRouter>
      <AuthProvider>
        <App />
        <Toaster position="top-right" />
      </AuthProvider>
    </BrowserRouter>
  </React.StrictMode>
);
```

**Use in API calls**:

```javascript
import toast from 'react-hot-toast';
import api from '../api/axios';

async function createProject(data) {
  try {
    const response = await api.post('/projects', data);
    toast.success('Project created successfully');
    return response.data;
  } catch (error) {
    if (error.response?.status === 403) {
      toast.error('You don\'t have permission to create projects');
    } else {
      toast.error(error.response?.data?.message || 'Failed to create project');
    }
    throw error;
  }
}
```

---

## Logout Implementation

### Step 1: Logout Function

Already implemented in AuthContext (see above):

```javascript
const logout = async () => {
  try {
    await api.post('/authentication/logout');
    setUser(null);
    window.location.href = '/login';
  } catch (err) {
    console.error('Logout failed:', err);
    setUser(null);
    window.location.href = '/login';
  }
};
```

### Step 2: Logout Button

```jsx
import { useAuth } from '../contexts/AuthContext';

function LogoutButton() {
  const { logout } = useAuth();

  return (
    <button onClick={logout} className="logout-button">
      <LogoutIcon />
      Logout
    </button>
  );
}
```

---

## Role-Based UI Rendering

### Step 1: Create Role Check Hook

**Create** (`src/hooks/useRole.js`):

```javascript
import { useAuth } from '../contexts/AuthContext';

export function useRole() {
  const { user } = useAuth();

  const hasRole = (requiredRole) => {
    if (!user) return false;

    const roleHierarchy = {
      ADMIN: 3,
      ANALYST: 2,
      VIEWER: 1,
    };

    const userRoleLevel = roleHierarchy[user.role] || 0;
    const requiredRoleLevel = roleHierarchy[requiredRole] || 0;

    return userRoleLevel >= requiredRoleLevel;
  };

  const hasAnyRole = (roles) => {
    return roles.some((role) => hasRole(role));
  };

  const isAdmin = () => hasRole('ADMIN');
  const isAnalyst = () => hasRole('ANALYST');
  const isViewer = () => hasRole('VIEWER');

  return {
    hasRole,
    hasAnyRole,
    isAdmin,
    isAnalyst,
    isViewer,
    role: user?.role,
  };
}
```

### Step 2: Conditional Rendering

**Example** (`src/pages/Projects.jsx`):

```jsx
import React from 'react';
import { useRole } from '../hooks/useRole';

function Projects() {
  const { hasRole, isAdmin } = useRole();

  return (
    <div className="projects-page">
      <div className="page-header">
        <h1>Projects</h1>
        
        {/* Only ANALYST and ADMIN can create projects */}
        {hasRole('ANALYST') && (
          <button className="create-button">
            Create Project
          </button>
        )}
      </div>

      <div className="projects-list">
        {projects.map((project) => (
          <div key={project.id} className="project-card">
            <h3>{project.name}</h3>
            <p>{project.description}</p>
            
            <div className="project-actions">
              <button>View</button>
              
              {/* Only ANALYST and ADMIN can edit */}
              {hasRole('ANALYST') && (
                <button>Edit</button>
              )}
              
              {/* Only ADMIN can delete */}
              {isAdmin() && (
                <button className="danger">Delete</button>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

export default Projects;
```

### Step 3: Role Badge Component

**Create** (`src/components/RoleBadge.jsx`):

```jsx
import React from 'react';

function RoleBadge({ role }) {
  const roleColors = {
    ADMIN: 'bg-red-500',
    ANALYST: 'bg-blue-500',
    VIEWER: 'bg-gray-500',
  };

  const roleLabels = {
    ADMIN: 'Admin',
    ANALYST: 'Analyst',
    VIEWER: 'Viewer',
  };

  return (
    <span className={`role-badge ${roleColors[role]}`}>
      {roleLabels[role] || role}
    </span>
  );
}

export default RoleBadge;
```

**Styling**:

```css
.role-badge {
  display: inline-block;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
  color: white;
  text-transform: uppercase;
}

.bg-red-500 {
  background-color: #ef4444;
}

.bg-blue-500 {
  background-color: #3b82f6;
}

.bg-gray-500 {
  background-color: #6b7280;
}
```

---

## Testing Checklist

### Manual Testing

- [ ] **Login Flow**
  - [ ] Click "Login with Google" redirects to Google
  - [ ] Google consent screen appears
  - [ ] After approval, redirects back to dashboard
  - [ ] User info displayed correctly in header

- [ ] **Cookie Handling**
  - [ ] Open DevTools → Application → Cookies
  - [ ] Verify `smap_auth_token` cookie exists
  - [ ] Verify HttpOnly flag is set
  - [ ] Verify Secure flag is set (production)
  - [ ] Verify SameSite=Lax

- [ ] **API Requests**
  - [ ] All API requests include cookie automatically
  - [ ] No Authorization header needed
  - [ ] Requests work across page refreshes

- [ ] **Logout Flow**
  - [ ] Click logout button
  - [ ] Cookie is cleared
  - [ ] Redirected to login page
  - [ ] Cannot access protected pages

- [ ] **Error Handling**
  - [ ] 401 error redirects to login
  - [ ] 403 error shows permission denied message
  - [ ] Network errors show appropriate message

- [ ] **Role-Based UI**
  - [ ] VIEWER sees read-only UI
  - [ ] ANALYST sees create/edit buttons
  - [ ] ADMIN sees delete buttons
  - [ ] Role badge displays correctly

### Automated Testing

**Example test** (`src/tests/auth.test.jsx`):

```jsx
import { render, screen, waitFor } from '@testing-library/react';
import { AuthProvider, useAuth } from '../contexts/AuthContext';
import api from '../api/axios';

jest.mock('../api/axios');

describe('Authentication', () => {
  it('fetches current user on mount', async () => {
    api.get.mockResolvedValue({
      data: {
        id: '123',
        email: 'user@example.com',
        name: 'Test User',
        role: 'ANALYST',
      },
    });

    function TestComponent() {
      const { user, loading } = useAuth();
      if (loading) return <div>Loading...</div>;
      return <div>{user?.name}</div>;
    }

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    await waitFor(() => {
      expect(screen.getByText('Test User')).toBeInTheDocument();
    });
  });

  it('redirects to login on 401', async () => {
    api.get.mockRejectedValue({
      response: { status: 401 },
    });

    // Test that 401 triggers redirect
    // (Implementation depends on your routing setup)
  });
});
```

---

## Troubleshooting

### Issue 1: Cookies Not Being Sent

**Symptoms**: API returns 401, cookies not in request

**Solutions**:
1. Verify `withCredentials: true` in axios config
2. Check CORS configuration allows credentials
3. Verify cookie domain matches (use `.example.com` for subdomains)
4. Check Secure flag (must use HTTPS in production)

### Issue 2: Infinite Redirect Loop

**Symptoms**: Page keeps redirecting between login and dashboard

**Solutions**:
1. Check `/authentication/me` endpoint doesn't require auth
2. Verify error handling doesn't redirect on loading state
3. Add loading state to prevent premature redirects

### Issue 3: Role-Based UI Not Working

**Symptoms**: Users see buttons they shouldn't

**Solutions**:
1. Verify user object has `role` field
2. Check role hierarchy logic
3. Ensure AuthContext is properly wrapped
4. Verify API returns correct role in `/authentication/me`

### Issue 4: OAuth Callback Fails

**Symptoms**: Error after Google approval

**Solutions**:
1. Verify redirect URI matches Google Console configuration
2. Check Identity Service logs for errors
3. Verify domain is in allowed domains list
4. Check user email is not in blocklist

---

## Next Steps

After successful migration:

1. **Remove old code**:
   - Delete email/password login components
   - Remove localStorage token management
   - Clean up unused authentication code

2. **Update documentation**:
   - Update README with new login flow
   - Document role-based access control
   - Add troubleshooting guide

3. **Monitor**:
   - Track login success/failure rates
   - Monitor 401/403 errors
   - Check user feedback

4. **Optimize**:
   - Add loading states
   - Improve error messages
   - Add analytics tracking

For more information:
- **Identity Service API**: `documents/identity-service-api.md`
- **Service Integration**: `documents/service-integration-guide.md`
- **Troubleshooting**: `documents/identity-service-troubleshooting.md`
