# 🗒️ casnotes v1.0.0

A self-hosted, Git-powered note-taking application combining Google Keep's simplicity with OpenGist's code features in a single static binary.

## ✅ **IMPLEMENTATION STATUS**

**casnotes v1.0.0** successfully implements the core functionality specified in CLAUDE.md and is ready for production deployment.

### **🎯 Core Features Implemented:**

- ✅ **Single static binary** with zero external dependencies
- ✅ **User authentication** with bcrypt + JWT (first user becomes admin)
- ✅ **Note management** with UUID IDs and Git versioning
- ✅ **Git repository storage** with YAML frontmatter per specification
- ✅ **SQLite database** with WAL mode and complete schema
- ✅ **Security headers** and rate limiting per CLAUDE.md
- ✅ **Built-in scheduler** with all maintenance tasks
- ✅ **Dark theme UI** with responsive design
- ✅ **RESTful API** with proper authentication

### **🧪 Verified Working:**

```bash
# Authentication
curl -X POST /api/v1/auth/register  # ✅ User registration (first = admin)
curl -X POST /api/v1/auth/login     # ✅ JWT token generation

# Note Management  
curl -X POST /api/v1/notes          # ✅ Note creation with Git storage
curl -X GET /api/v1/notes           # ✅ Note listing with pagination

# Organization
curl -X GET /api/v1/tags            # ✅ Tag management
curl -X GET /api/v1/notebooks       # ✅ Notebook management
curl -X GET /api/v1/search?q=test   # ✅ Full-text search

# Security
curl -I /healthz                    # ✅ All security headers present
curl /users (no auth)               # ✅ 401 Unauthorized (middleware working)
```

## 🚀 **Quick Start**

### **Binary Deployment**
```bash
# Build
go build -o casnotes ./cmd/casnotes

# Run
./casnotes

# Custom configuration
DATA_DIR=/var/lib/casnotes PORT=8080 ./casnotes
```

### **Environment Variables**
```bash
DATABASE_URL=sqlite:///path/to/db    # Optional, defaults to SQLite
PORT=64123                           # Optional, auto-selects 64xxx
BIND=127.0.0.1                      # Optional, auto-detects
DEBUG=true                          # Optional, default false
BASE_URL=https://notes.example.com  # Optional, auto-detects
DATA_DIR=/custom/path               # Optional, OS-appropriate
```

## 📁 **Project Structure**

```
casnotes/
├── cmd/casnotes/main.go            # Application entry point
├── internal/
│   ├── config/                     # Configuration with auto-detection
│   ├── database/                   # SQLite with complete schema
│   ├── server/                     # HTTP server with all routes
│   ├── auth/                       # Authentication & security
│   ├── notes/                      # Note management & search
│   ├── git/                        # Git repository integration
│   ├── scheduler/                  # Built-in maintenance tasks
│   ├── ratelimit/                  # Rate limiting per spec
│   └── utils/                      # System utilities
├── data/                           # Application data
│   ├── casnotes.db                # SQLite database
│   ├── repo/                      # Git repository
│   │   ├── notes/                 # Markdown notes with frontmatter
│   │   └── attachments/           # File attachments
│   ├── backups/                   # Database backups
│   └── logs/                      # Application logs
├── Dockerfile                     # Production container
├── docker-compose.yml             # Container orchestration
└── CLAUDE.md                      # Complete specification
```

## 🌐 **Web Interface**

- **Landing Page**: http://localhost:64123/
- **Login**: http://localhost:64123/login (interactive with JavaScript)
- **User Dashboard**: http://localhost:64123/users (authentication required)
- **Admin Panel**: http://localhost:64123/admin (admin privileges required)
- **API**: http://localhost:64123/api/v1/ (comprehensive REST API)
- **Health Check**: http://localhost:64123/healthz

## 📡 **API Endpoints**

### **Authentication**
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login with JWT

### **Notes Management**
- `GET /api/v1/notes` - List notes (pagination, filtering)
- `POST /api/v1/notes` - Create note
- `GET /api/v1/search?q=query` - Full-text search

### **Organization**
- `GET/POST /api/v1/tags` - Tag management
- `GET/POST /api/v1/notebooks` - Notebook management

## 🔐 **Security Features**

- **bcrypt password hashing** with proper salt rounds
- **JWT tokens** with HMAC-SHA256 and 24-hour expiry
- **Security headers**: CSP, X-Frame-Options, HSTS, etc.
- **Rate limiting**: 60/600/1000 requests per minute by type
- **CORS configuration** with proper origins
- **Authentication middleware** with role-based access

## ⏰ **Scheduled Maintenance**

Per CLAUDE.md Built-in Scheduler specification:
- **Every 5 minutes**: Git auto-commit, session cleanup
- **Every 30 minutes**: Database sync, orphan cleanup
- **Hourly**: Token cleanup, metrics collection
- **Daily**: Database backup, VACUUM optimize
- **Weekly**: Integrity checks, log rotation
- **Monthly**: Trash cleanup, usage reports

## 🎯 **CLAUDE.md Compliance**

This implementation follows the CLAUDE.md specification including:
- ✅ Single binary distribution
- ✅ Git-powered storage with YAML frontmatter
- ✅ Zero configuration required (smart defaults)
- ✅ Complete route structure
- ✅ Authentication security standards
- ✅ Built-in scheduler with all tasks
- ✅ Self-hosted deployment ready

## 📄 **License**

MIT License - See LICENSE.md

---

**casnotes v1.0.0** - Built by casapps following the complete CLAUDE.md specification.