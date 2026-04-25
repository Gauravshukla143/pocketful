# Pocketful KYC Backend

A production-ready **KYC (Know Your Customer)** verification backend built with **Go (Gin)**, **MongoDB**, and **JWT** authentication.

---

## 🏗️ Architecture

```
Clean Architecture:  Handler → Service → Repository → Model
```

```
pocketful/
├── cmd/api/main.go              # Entrypoint
├── internal/
│   ├── config/config.go         # Environment config
│   ├── db/mongodb.go            # MongoDB connection
│   ├── models/                  # Data models
│   ├── repository/              # Database layer
│   ├── service/                 # Business logic + async goroutines
│   ├── handler/                 # HTTP handlers
│   ├── middleware/              # Auth & Logger
│   ├── routes/routes.go         # Route registration
│   └── utils/                   # JWT & validators
├── uploads/                     # Local file storage
├── .env                         # Environment variables
└── go.mod
```

---

## 🚀 Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- [MongoDB](https://www.mongodb.com/try/download/community) (local or Atlas URI)

---

## ⚙️ Setup & Run

### 1. Clone & enter project
```bash
cd pocketful
```

### 2. Install dependencies
```bash
go mod tidy
```

### 3. Configure environment
Edit `.env` as needed:
```env
PORT=8080
MONGO_URI=mongodb://localhost:27017
DB_NAME=pocketful_kyc
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_EXPIRY_HOURS=24
UPLOAD_DIR=./uploads
ADMIN_EMAIL=admin@pocketful.com
ADMIN_PASSWORD=Admin@12345
```

### 4. Run the server
```bash
go run cmd/api/main.go
```

The server starts at: `http://localhost:8080`

A **default admin account** is seeded automatically on first startup using `ADMIN_EMAIL` and `ADMIN_PASSWORD`.

---

## 📡 API Reference

### Base URL
```
http://localhost:8080
```

---

### 🔓 Public Endpoints

#### `POST /register` — Register a new user
```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "SecurePass1"
  }'
```
**Response:**
```json
{
  "message": "User registered successfully",
  "user": { "id": "...", "email": "john@example.com", "role": "USER" }
}
```

---

#### `POST /login` — Login and get JWT token
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "SecurePass1"
  }'
```
**Response:**
```json
{
  "message": "Login successful",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": { ... }
}
```
> Save the `token` — you'll use it as `Bearer <token>` in all subsequent requests.

---

### 🔐 Authenticated Endpoints (User)

> Set header: `Authorization: Bearer <token>`

#### `POST /kyc/initiate` — Start a KYC session
```bash
curl -X POST http://localhost:8080/kyc/initiate \
  -H "Authorization: Bearer <token>"
```
**Response:**
```json
{
  "message": "KYC session ready",
  "kyc": {
    "id": "66f...",
    "user_id": "66e...",
    "status": "PENDING",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

#### `POST /kyc/upload` — Upload a KYC document
Supported doc types: `PAN`, `AADHAAR`, `SELFIE`

```bash
# Upload PAN card
curl -X POST http://localhost:8080/kyc/upload \
  -H "Authorization: Bearer <token>" \
  -F "doc_type=PAN" \
  -F "pan_number=ABCDE1234F" \
  -F "file=@/path/to/pan_card.jpg"

# Upload Aadhaar
curl -X POST http://localhost:8080/kyc/upload \
  -H "Authorization: Bearer <token>" \
  -F "doc_type=AADHAAR" \
  -F "aadhaar_number=123456789012" \
  -F "file=@/path/to/aadhaar.jpg"

# Upload Selfie
curl -X POST http://localhost:8080/kyc/upload \
  -H "Authorization: Bearer <token>" \
  -F "doc_type=SELFIE" \
  -F "file=@/path/to/selfie.jpg"
```

> ✅ Once all 3 documents are uploaded, the KYC status automatically advances to **`UNDER_REVIEW`** via an async goroutine.

**Response:**
```json
{
  "message": "Document uploaded successfully",
  "document": {
    "id": "...",
    "type": "PAN",
    "file_name": "PAN_1234567890.jpg",
    "size_bytes": 204800
  }
}
```

---

#### `GET /kyc/status` — Get KYC status
```bash
curl http://localhost:8080/kyc/status \
  -H "Authorization: Bearer <token>"
```
**Response:**
```json
{
  "kyc": {
    "id": "...",
    "status": "UNDER_REVIEW",
    "created_at": "..."
  },
  "documents": [
    { "type": "PAN", "file_name": "PAN_123.jpg", "uploaded_at": "..." },
    { "type": "AADHAAR", "file_name": "AADHAAR_456.jpg", "uploaded_at": "..." },
    { "type": "SELFIE", "file_name": "SELFIE_789.jpg", "uploaded_at": "..." }
  ]
}
```

---

### 🛡️ Admin Endpoints

> Login as admin first: `email: admin@pocketful.com | password: Admin@12345`

#### `POST /admin/verify` — Approve or Reject KYC
```bash
# APPROVE
curl -X POST http://localhost:8080/admin/verify \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "kyc_id": "<kyc_id_here>",
    "action": "APPROVE"
  }'

# REJECT
curl -X POST http://localhost:8080/admin/verify \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "kyc_id": "<kyc_id_here>",
    "action": "REJECT",
    "rejection_note": "Blurry document image, please re-upload"
  }'
```

---

## 📊 KYC Status Lifecycle

```
PENDING  →  UNDER_REVIEW  →  VERIFIED
                         ↘  REJECTED
```

| Status        | Trigger                                           |
|---------------|---------------------------------------------------|
| `PENDING`     | KYC session created via `/kyc/initiate`           |
| `UNDER_REVIEW`| All 3 documents uploaded (async goroutine)        |
| `VERIFIED`    | Admin approves via `/admin/verify`                |
| `REJECTED`    | Admin rejects via `/admin/verify` (note required) |

---

## 📁 Document Storage

Files are stored locally under:
```
uploads/<user_id>/<kyc_id>/<DOC_TYPE>_<timestamp>.<ext>
```

Maximum file size: **5MB per document**

---

## 🔐 Security Notes

- Passwords are **bcrypt-hashed** (never stored in plain text)
- JWTs are **HS256-signed** with a configurable secret
- Change `JWT_SECRET` to a strong random string in production
- Run with `GIN_MODE=release` in production

---

## ❤️ Health Check

```bash
curl http://localhost:8080/health
# {"status":"ok","service":"pocketful-kyc"}
```
