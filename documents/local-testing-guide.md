# Hướng Dẫn Test Local với HTTP Cookies

## Tổng Quan Vấn Đề

Khi sử dụng JWT thông thường (gửi qua Authorization header), việc test local rất đơn giản - chỉ cần copy token và thêm vào header. Tuy nhiên, với **HttpOnly Cookies**, có một số thách thức:

### 🔴 Vấn Đề 1: CORS Configuration cho Localhost
**Vấn đề**: Browser sẽ block cookies từ cross-origin requests nếu CORS không được config đúng.

**Giải pháp**: Service đã được config sẵn để hỗ trợ localhost trong development mode.

### 🔴 Vấn Đề 2: Cookie Domain Configuration
**Vấn đề**: Cookies chỉ được gửi khi domain khớp với cookie domain setting.

**Giải pháp**: Phải config `cookie.domain` phù hợp với môi trường test.

### 🔴 Vấn đề 3: Không Thể Lấy Token Để Test Thủ Công
**Vấn đề**: HttpOnly cookies không thể đọc được từ JavaScript, khó debug và test.

**Giải pháp**: Sử dụng Browser DevTools hoặc test client đã được chuẩn bị sẵn.

---

## Cấu Hình Cho Local Testing

### 1. Cấu Hình Cookie Settings

Mở file `config/auth-config.yaml` và điều chỉnh phần cookie:

```yaml
# Cookie Configuration
cookie:
  domain: localhost              # ✅ Quan trọng: Dùng "localhost" cho local testing
  secure: false                  # ✅ Phải là false cho HTTP (localhost)
  samesite: Lax                  # ✅ Lax cho phép cookies trong redirects
  max_age: 28800                 # 8 hours
  max_age_remember: 604800       # 7 days
  name: smap_auth_token
```

**⚠️ LƯU Ý QUAN TRỌNG**:
- `domain: localhost` - KHÔNG dùng `.localhost` (dấu chấm sẽ gây lỗi)
- `secure: false` - Bắt buộc cho HTTP (localhost không có SSL)
- `samesite: Lax` - Cho phép cookies được gửi trong OAuth redirects

### 2. Cấu Hình CORS

Service đã tự động hỗ trợ localhost khi `environment.name != "production"`:

```yaml
# Environment Configuration
environment:
  name: development  # ✅ Không dùng "production" khi test local
```

**Cách hoạt động của CORS middleware** (file `internal/middleware/cors.go`):

```go
// Development/Staging mode tự động cho phép:
// 1. Production domains (https://smap.tantai.dev)
// 2. Localhost với bất kỳ port nào (http://localhost:3000, http://localhost:8080)
// 3. Private subnets (172.16.21.0/24, 172.16.19.0/24, 192.168.1.0/24)

if environment != "production" {
    config.AllowOriginFunc = func(origin string) bool {
        // Allow localhost (any port)
        if isLocalhostOrigin(origin) {
            return true
        }
        // Allow private subnets
        if isPrivateOrigin(origin) {
            return true
        }
        return false
    }
}
```

### 3. Cấu Hình OAuth Redirect URI

Trong Google Cloud Console, thêm redirect URI cho localhost:

```
http://localhost:8080/authentication/callback
```

Trong `config/auth-config.yaml`:

```yaml
oauth2:
  provider: google
  client_id: YOUR_CLIENT_ID.apps.googleusercontent.com
  client_secret: YOUR_CLIENT_SECRET
  redirect_uri: http://localhost:8080/authentication/callback  # ✅ HTTP cho localhost
```

---

## Phương Pháp Test

### Phương Pháp 1: Sử Dụng Test Client (Khuyến Nghị) ⭐

Service đã có sẵn một test client HTML tại `cmd/test-client/`.

#### Bước 1: Start Auth Service
```bash
# Terminal 1: Start dependencies
docker-compose up -d postgres redis kafka

# Terminal 2: Start auth service
make run-api
# hoặc
go run cmd/api/main.go
```

#### Bước 2: Start Test Client
```bash
# Terminal 3: Start test client
go run cmd/test-client/main.go
```

Test client sẽ chạy tại: `http://localhost:3000`

#### Bước 3: Test Flow

1. **Mở browser**: `http://localhost:3000`
2. **Click "Login with Google"**: 
   - Redirect đến `http://localhost:8080/authentication/login`
   - Redirect đến Google OAuth
   - Sau khi login, redirect về `http://localhost:3000`
   - Cookie `smap_auth_token` được set tự động
3. **Click "Get My Info"**:
   - Gửi request đến `/authentication/me` với `credentials: 'include'`
   - Cookie được gửi tự động
   - Nhận thông tin user
4. **Click "Logout"**:
   - Cookie bị expire
   - Session bị revoke

**✅ Ưu điểm**:
- Tự động xử lý cookies
- Giống production flow
- Dễ debug với Browser DevTools

---

### Phương Pháp 2: Sử Dụng Browser DevTools

#### Bước 1: Login qua Browser

Mở browser và truy cập:
```
http://localhost:8080/authentication/login
```

Sau khi login thành công, bạn sẽ được redirect về dashboard.

#### Bước 2: Kiểm Tra Cookie

Mở **DevTools** → **Application** → **Cookies** → `http://localhost:8080`

Bạn sẽ thấy cookie:
```
Name: smap_auth_token
Value: eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
Domain: localhost
Path: /
HttpOnly: ✓
Secure: (empty - vì đang dùng HTTP)
SameSite: Lax
```

#### Bước 3: Test API với Browser Console

Mở **DevTools** → **Console** và chạy:

```javascript
// Test GET /authentication/me
fetch('http://localhost:8080/authentication/me', {
  method: 'GET',
  credentials: 'include'  // ✅ Quan trọng: Gửi cookies
})
.then(r => r.json())
.then(data => console.log(data));

// Test POST /authentication/logout
fetch('http://localhost:8080/authentication/logout', {
  method: 'POST',
  credentials: 'include'
})
.then(r => r.json())
.then(data => console.log(data));
```

**⚠️ LƯU Ý**: Phải thêm `credentials: 'include'` để browser gửi cookies!

---

### Phương Pháp 3: Sử Dụng Postman (Chi Tiết) 🔧

Postman có thể test HttpOnly cookies nhưng cần setup đúng cách. Có 2 phương pháp:

#### Phương Pháp 3A: Sử Dụng Postman Interceptor (Khuyến Nghị)

**Bước 1: Cài Đặt Postman Interceptor**

1. Cài extension "Postman Interceptor" cho Chrome/Edge
2. Trong Postman Desktop App:
   - Click icon "Capture requests" (satellite icon) ở bottom-right
   - Enable "Capture Cookies"
   - Chọn domain: `localhost`

**Bước 2: Login Qua Browser**

1. Mở Chrome/Edge (browser có Interceptor)
2. Truy cập: `http://localhost:8080/authentication/login`
3. Login với Google
4. Sau khi login thành công, cookie được lưu trong browser

**Bước 3: Test API Trong Postman**

Postman Interceptor sẽ tự động sync cookies từ browser.

1. **Request: Get User Info**
   ```
   GET http://localhost:8080/authentication/me
   ```
   - Tab "Cookies": Bạn sẽ thấy `smap_auth_token` được sync từ browser
   - Click "Send"
   - Response: User information

2. **Request: Logout**
   ```
   POST http://localhost:8080/authentication/logout
   ```
   - Cookie sẽ bị expire
   - Kiểm tra lại tab "Cookies" - cookie đã mất

**✅ Ưu điểm**:
- Tự động sync cookies từ browser
- Không cần copy/paste thủ công
- Giống production flow

**❌ Nhược điểm**:
- Cần cài extension
- Chỉ hoạt động với Chrome/Edge

---

#### Phương Pháp 3B: Manual Cookie Management (Không Cần Extension)

**Bước 1: Setup Postman**

1. Mở Postman Settings (⚙️)
2. General tab:
   - ✅ Enable "Automatically follow redirects" 
   - ✅ Enable "Send cookies"
   - ✅ Enable "Capture cookies"

**Bước 2: Login Qua Browser và Extract Cookie**

1. Mở browser: `http://localhost:8080/authentication/login`
2. Login với Google
3. Mở DevTools → Application → Cookies → `http://localhost:8080`
4. Copy giá trị của cookie `smap_auth_token`

**Bước 3: Add Cookie Vào Postman**

1. Trong Postman, click "Cookies" (dưới "Send" button)
2. Chọn domain: `localhost`
3. Click "Add Cookie"
4. Nhập cookie string theo format:

```
smap_auth_token=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...; Path=/; Domain=localhost; HttpOnly; SameSite=Lax
```

Hoặc dùng form:
- Name: `smap_auth_token`
- Value: `eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...` (paste token)
- Domain: `localhost`
- Path: `/`
- HttpOnly: ✅
- Secure: ❌ (vì đang dùng HTTP)

**Bước 4: Test API**

1. **Request: Get User Info**
   ```
   GET http://localhost:8080/authentication/me
   ```
   - Tab "Headers": Không cần thêm gì
   - Tab "Cookies": Cookie đã được add ở bước 3
   - Click "Send"
   - Response 200: User information

2. **Request: Get JWKS**
   ```
   GET http://localhost:8080/authentication/.well-known/jwks.json
   ```
   - Không cần cookie (public endpoint)
   - Response: Public keys

3. **Request: Logout**
   ```
   POST http://localhost:8080/authentication/logout
   ```
   - Cookie tự động được gửi
   - Response 200: Success
   - Cookie bị expire (check lại tab "Cookies")

**Bước 5: Verify Cookie Expired**

Sau khi logout, thử request lại:
```
GET http://localhost:8080/authentication/me
```
- Response 401: Unauthorized (cookie đã expire)

---

#### Phương Pháp 3C: Sử Dụng Postman Collection với Pre-request Script

Tạo một collection với script tự động quản lý cookies.

**Bước 1: Tạo Environment**

1. Tạo environment mới: "Local Auth"
2. Add variables:
   - `base_url`: `http://localhost:8080`
   - `auth_token`: (để trống, sẽ được set tự động)

**Bước 2: Tạo Request "Manual Set Token"**

Dùng để set token sau khi login qua browser:

```
POST {{base_url}}/authentication/me
```

Pre-request Script:
```javascript
// Paste token vào đây sau khi login qua browser
const token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...";

// Set vào environment
pm.environment.set("auth_token", token);

// Set cookie
pm.cookies.set({
    url: pm.environment.get("base_url"),
    name: "smap_auth_token",
    value: token,
    path: "/",
    domain: "localhost",
    httpOnly: true,
    sameSite: "Lax"
});
```

**Bước 3: Tạo Request "Get Me"**

```
GET {{base_url}}/authentication/me
```

Pre-request Script:
```javascript
// Đảm bảo cookie được set
const token = pm.environment.get("auth_token");
if (token) {
    pm.cookies.set({
        url: pm.environment.get("base_url"),
        name: "smap_auth_token",
        value: token,
        path: "/",
        domain: "localhost",
        httpOnly: true,
        sameSite: "Lax"
    });
}
```

Tests Script:
```javascript
// Verify response
pm.test("Status code is 200", function () {
    pm.response.to.have.status(200);
});

pm.test("Response has user data", function () {
    const jsonData = pm.response.json();
    pm.expect(jsonData.data).to.have.property('id');
    pm.expect(jsonData.data).to.have.property('email');
});
```

**Bước 4: Tạo Request "Logout"**

```
POST {{base_url}}/authentication/logout
```

Tests Script:
```javascript
pm.test("Logout successful", function () {
    pm.response.to.have.status(200);
});

// Clear token from environment
pm.environment.unset("auth_token");
```

**Bước 5: Sử Dụng Collection**

1. Login qua browser → Copy token
2. Run "Manual Set Token" request → Paste token vào script
3. Run "Get Me" request → Verify user info
4. Run "Logout" request → Cookie expired

---

#### Troubleshooting Postman

**Issue 1: Cookie Không Được Gửi**

Kiểm tra:
1. Tab "Cookies" → Verify cookie tồn tại cho domain `localhost`
2. Settings → "Send cookies" phải được enable
3. Cookie domain phải khớp với request URL

Debug:
```javascript
// Pre-request Script để debug
pm.cookies.jar().getAll(pm.request.url.toString(), (error, cookies) => {
    console.log("Cookies for this request:", cookies);
});
```

**Issue 2: Cookie Bị Reject**

Nguyên nhân:
- Domain không khớp (dùng `127.0.0.1` thay vì `localhost`)
- Secure flag = true nhưng dùng HTTP
- SameSite = Strict

Giải pháp:
```javascript
// Đảm bảo cookie settings đúng
pm.cookies.set({
    url: "http://localhost:8080",  // Phải dùng localhost, không dùng 127.0.0.1
    name: "smap_auth_token",
    value: token,
    path: "/",
    domain: "localhost",           // Không có dấu chấm
    httpOnly: true,
    secure: false,                 // false cho HTTP
    sameSite: "Lax"               // Lax hoặc None
});
```

**Issue 3: Postman Không Thể Login Trực Tiếp**

Postman không thể xử lý OAuth flow (redirect đến Google) một cách tự động.

Giải pháp:
1. **Dùng Postman Interceptor** (sync cookies từ browser)
2. **Login qua browser** → Copy cookie thủ công
3. **Dùng test client** (`cmd/test-client/`) thay vì Postman

---

#### So Sánh Các Phương Pháp Postman

| Phương Pháp | Ưu Điểm | Nhược Điểm | Khuyến Nghị |
|-------------|---------|------------|-------------|
| **Interceptor** | Tự động sync, dễ dùng | Cần extension | ⭐⭐⭐⭐⭐ |
| **Manual Cookie** | Không cần extension | Phải copy/paste | ⭐⭐⭐ |
| **Pre-request Script** | Automation, reusable | Phức tạp setup | ⭐⭐⭐⭐ |

---

#### Postman Collection Mẫu

Tạo file `postman_collection.json`:

```json
{
  "info": {
    "name": "SMAP Auth Service",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Authentication",
      "item": [
        {
          "name": "Get Me",
          "request": {
            "method": "GET",
            "header": [],
            "url": {
              "raw": "{{base_url}}/authentication/me",
              "host": ["{{base_url}}"],
              "path": ["authentication", "me"]
            }
          }
        },
        {
          "name": "Logout",
          "request": {
            "method": "POST",
            "header": [],
            "url": {
              "raw": "{{base_url}}/authentication/logout",
              "host": ["{{base_url}}"],
              "path": ["authentication", "logout"]
            }
          }
        },
        {
          "name": "Get JWKS",
          "request": {
            "method": "GET",
            "header": [],
            "url": {
              "raw": "{{base_url}}/authentication/.well-known/jwks.json",
              "host": ["{{base_url}}"],
              "path": ["authentication", ".well-known", "jwks.json"]
            }
          }
        }
      ]
    },
    {
      "name": "Internal API",
      "item": [
        {
          "name": "Validate Token",
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              },
              {
                "key": "Authorization",
                "value": "{{internal_key}}"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"token\": \"{{auth_token}}\"\n}"
            },
            "url": {
              "raw": "{{base_url}}/internal/validate",
              "host": ["{{base_url}}"],
              "path": ["internal", "validate"]
            }
          }
        }
      ]
    }
  ]
}
```

Import vào Postman và sử dụng với environment:
```json
{
  "name": "Local",
  "values": [
    {
      "key": "base_url",
      "value": "http://localhost:8080",
      "enabled": true
    },
    {
      "key": "auth_token",
      "value": "",
      "enabled": true
    },
    {
      "key": "internal_key",
      "value": "project-service-key",
      "enabled": true
    }
  ]
}
```

---

**⚠️ Lưu Ý Quan Trọng Khi Dùng Postman**:

1. **OAuth Flow**: Postman KHÔNG thể tự động xử lý OAuth redirect. Bạn phải:
   - Dùng Postman Interceptor + Browser
   - Hoặc login qua browser → copy cookie

2. **Cookie Domain**: Phải dùng `localhost` trong cả:
   - Request URL: `http://localhost:8080`
   - Cookie domain: `localhost`
   - KHÔNG dùng `127.0.0.1`

3. **HttpOnly Cookies**: Postman có thể set/send HttpOnly cookies (khác với browser JavaScript)

4. **Testing Recommendation**: 
   - Development: Dùng test client (`cmd/test-client/`)
   - API Testing: Dùng Postman với Interceptor
   - Automation: Dùng Postman Collection với scripts

---

### Phương Pháp 4: Sử Dụng curl (Advanced)

#### Bước 1: Login và Lưu Cookie

```bash
# Lưu cookies vào file
curl -c cookies.txt -L \
  'http://localhost:8080/authentication/login' \
  -H 'User-Agent: Mozilla/5.0'
```

**⚠️ Vấn đề**: curl không thể xử lý OAuth redirect tự động. Bạn cần:
1. Copy authorization URL từ response
2. Mở browser, login
3. Copy cookie từ browser

#### Bước 2: Extract Cookie từ Browser

Sau khi login qua browser, mở DevTools và copy cookie value:

```bash
# Tạo file cookies.txt thủ công
echo "localhost	FALSE	/	FALSE	0	smap_auth_token	eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..." > cookies.txt
```

#### Bước 3: Test với Cookie

```bash
# Get user info
curl -b cookies.txt \
  http://localhost:8080/authentication/me

# Logout
curl -b cookies.txt -X POST \
  http://localhost:8080/authentication/logout
```

**✅ Ưu điểm**: Tốt cho automation scripts
**❌ Nhược điểm**: Phức tạp, không xử lý OAuth tốt

---

## Debug Common Issues

### Issue 1: Cookie Không Được Set

**Triệu chứng**: Sau khi login, không thấy cookie trong DevTools.

**Nguyên nhân & Giải pháp**:

1. **Cookie domain không khớp**
   ```yaml
   # ❌ SAI
   cookie:
     domain: .localhost  # Dấu chấm gây lỗi
   
   # ✅ ĐÚNG
   cookie:
     domain: localhost
   ```

2. **Secure flag = true với HTTP**
   ```yaml
   # ❌ SAI (cho localhost HTTP)
   cookie:
     secure: true
   
   # ✅ ĐÚNG
   cookie:
     secure: false
   ```

3. **SameSite = Strict**
   ```yaml
   # ❌ SAI (block OAuth redirects)
   cookie:
     samesite: Strict
   
   # ✅ ĐÚNG
   cookie:
     samesite: Lax
   ```

### Issue 2: CORS Error

**Triệu chứng**: 
```
Access to fetch at 'http://localhost:8080/authentication/me' from origin 
'http://localhost:3000' has been blocked by CORS policy
```

**Nguyên nhân & Giải pháp**:

1. **Environment = production**
   ```yaml
   # ❌ SAI
   environment:
     name: production
   
   # ✅ ĐÚNG
   environment:
     name: development
   ```

2. **Thiếu credentials: 'include'**
   ```javascript
   // ❌ SAI
   fetch('http://localhost:8080/authentication/me')
   
   // ✅ ĐÚNG
   fetch('http://localhost:8080/authentication/me', {
     credentials: 'include'
   })
   ```

### Issue 3: Cookie Được Set Nhưng Không Được Gửi

**Triệu chứng**: Cookie hiển thị trong DevTools nhưng request không có cookie.

**Nguyên nhân & Giải pháp**:

1. **Request từ origin khác**
   - Cookie domain: `localhost`
   - Request từ: `127.0.0.1` ❌
   - Giải pháp: Dùng `localhost` thống nhất

2. **Path không khớp**
   - Cookie path: `/authentication`
   - Request: `/api/me` ❌
   - Giải pháp: Set cookie path = `/`

3. **Cookie đã expire**
   - Check "Expires / Max-Age" trong DevTools
   - Giải pháp: Login lại

### Issue 4: 401 Unauthorized Sau Khi Login

**Triệu chứng**: Login thành công nhưng `/me` trả về 401.

**Debug steps**:

1. **Kiểm tra cookie có được set không**
   ```javascript
   // DevTools Console
   document.cookie
   ```

2. **Kiểm tra cookie có được gửi không**
   - DevTools → Network → Request → Headers → Cookie

3. **Kiểm tra JWT token có valid không**
   ```bash
   # Copy token từ cookie và decode tại jwt.io
   ```

4. **Kiểm tra blacklist**
   ```bash
   # Connect to Redis
   docker exec -it redis redis-cli
   
   # Check if token is blacklisted
   GET blacklist:YOUR_JTI
   ```

---

## So Sánh: JWT Header vs HttpOnly Cookie

### JWT qua Authorization Header

```javascript
// Client code
const token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...";

fetch('http://localhost:8080/api/users', {
  headers: {
    'Authorization': `Bearer ${token}`
  }
});
```

**✅ Ưu điểm**:
- Dễ test (copy/paste token)
- Dễ debug (thấy token trong request)
- Không cần CORS credentials

**❌ Nhược điểm**:
- Dễ bị XSS attack (JavaScript có thể đọc token)
- Phải tự quản lý token storage
- Token có thể bị leak qua console.log

### JWT qua HttpOnly Cookie

```javascript
// Client code
fetch('http://localhost:8080/authentication/me', {
  credentials: 'include'  // Browser tự động gửi cookie
});
```

**✅ Ưu điểm**:
- Bảo mật cao (JavaScript không đọc được)
- Tự động gửi cookie (không cần code)
- Chống XSS attack

**❌ Nhược điểm**:
- Khó test hơn (không thấy token)
- Cần config CORS đúng
- Cần config cookie domain/secure/samesite

---

## Best Practices

### 1. Development Environment

```yaml
# config/auth-config.yaml
environment:
  name: development

cookie:
  domain: localhost
  secure: false
  samesite: Lax

oauth2:
  redirect_uri: http://localhost:8080/authentication/callback
```

### 2. Staging Environment

```yaml
environment:
  name: staging

cookie:
  domain: .staging.smap.com
  secure: true
  samesite: Lax

oauth2:
  redirect_uri: https://auth-staging.smap.com/authentication/callback
```

### 3. Production Environment

```yaml
environment:
  name: production

cookie:
  domain: .smap.com
  secure: true
  samesite: Strict  # Hoặc Lax nếu cần cross-site

oauth2:
  redirect_uri: https://auth.smap.com/authentication/callback
```

---

## Testing Checklist

### ✅ Pre-Test Setup
- [ ] Config `cookie.domain = localhost`
- [ ] Config `cookie.secure = false`
- [ ] Config `environment.name = development`
- [ ] Start PostgreSQL, Redis, Kafka
- [ ] Start Auth Service
- [ ] Add `http://localhost:8080/authentication/callback` to Google OAuth

### ✅ Login Flow Test
- [ ] Access `http://localhost:8080/authentication/login`
- [ ] Redirect to Google OAuth
- [ ] Login with Google account
- [ ] Redirect back to localhost
- [ ] Cookie `smap_auth_token` được set
- [ ] Cookie có HttpOnly flag
- [ ] Cookie có SameSite=Lax

### ✅ Authenticated Request Test
- [ ] Request `/authentication/me` với `credentials: 'include'`
- [ ] Response 200 với user info
- [ ] Cookie được gửi trong request header

### ✅ Logout Test
- [ ] Request `/authentication/logout`
- [ ] Cookie bị expire (Max-Age=-1)
- [ ] Request `/authentication/me` trả về 401

---

## Tóm Tắt

### Vấn Đề Chính Khi Test Local với HttpOnly Cookies:

1. **CORS Configuration**: Phải cho phép localhost origin và enable credentials
2. **Cookie Domain**: Phải dùng `localhost` (không có dấu chấm)
3. **Cookie Secure**: Phải là `false` cho HTTP localhost
4. **Cookie SameSite**: Phải là `Lax` để cho phép OAuth redirects
5. **Không Thể Lấy Token**: Phải dùng Browser DevTools hoặc test client

### Giải Pháp:

1. **Sử dụng Test Client** (`cmd/test-client/`) - Khuyến nghị nhất
2. **Sử dụng Browser DevTools** - Tốt cho debug
3. **Config đúng** `auth-config.yaml` cho development
4. **Luôn dùng** `credentials: 'include'` trong fetch requests

### Quick Start:

```bash
# 1. Config
cp config/auth-config.example.yaml config/auth-config.yaml
# Edit: cookie.domain=localhost, cookie.secure=false, environment.name=development

# 2. Start services
docker-compose up -d
make run-api

# 3. Start test client
go run cmd/test-client/main.go

# 4. Open browser
open http://localhost:3000
```

---

## Tài Liệu Tham Khảo

- [MDN: HTTP Cookies](https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies)
- [MDN: CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)
- [OWASP: HttpOnly Cookie](https://owasp.org/www-community/HttpOnly)
- [SameSite Cookie Explained](https://web.dev/samesite-cookies-explained/)

---

**Cập nhật lần cuối**: 14/02/2026
