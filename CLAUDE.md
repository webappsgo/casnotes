# casnotes - Complete Technical Specification v1.0.0

## Project Overview
**Name:** casnotes  
**Organization:** casapps  
**License:** MIT (LICENSE.md)  
**Description:** A self-hosted, Git-powered note-taking application combining Google Keep's simplicity with OpenGist's code features in a single static binary.

## Core Architecture

### Technology Stack
- **Language:** Go (for cross-compilation and single binary)
- **Storage:** Git repository (using go-git)
- **Database:** SQLite (default), PostgreSQL, MariaDB supported
- **Search:** SQLite FTS5 for full-text search
- **Frontend:** Embedded static files (no external dependencies)
- **No config files:** All settings in database

### Binary Distribution
- Single static binary for all platforms
- Embedded web UI, templates, and assets
- Zero external dependencies
- Smart auto-detection and self-configuration

### Platform Targets
- linux/amd64
- linux/arm64
- darwin/amd64
- darwin/arm64
- windows/amd64
- windows/arm64

### Binary Naming Scheme
- Host binary: `casnotes`
- Cross-platform: `casnotes-{os}-{arch}`
- Windows: `casnotes-windows-{arch}.exe`

## Build System

### Makefile Targets
```makefile
make build    # Build all platforms + host binary
make test     # Run all tests with coverage
make docker   # Build and push to ghcr.io
make release  # Create GitHub release
make clean    # Remove build artifacts
make run      # Run locally for development
make install  # Install host binary to /usr/local/bin
```

### CI/CD Configuration
**Jenkins Server:** ci.casjay.cc  
**Build Agents:** arm64, amd64  
**Docker Registry:** ghcr.io/casapps/casnotes  
**Image Tags:** latest, v1.0.0, edge, sha-xxxxxx

## Installation & Deployment

### Binary Behavior
```bash
casnotes              # Default: auto-detect, escalate if possible, start server
casnotes --help       # Show help
casnotes --version    # Show version
casnotes --debug      # Enable debug logging
```

### Privilege Escalation Strategy
**Linux (assume headless/CI/CD):**
- Silent check: `sudo -n true`
- If available: re-execute elevated
- If not: user mode in ~/.local/
- No GUI prompts, fully unattended

**macOS (GUI environment):**
- Check elevation status
- Request admin via native dialog
- Install system-wide if approved

**Windows (GUI environment):**
- Check Administrator status
- Trigger UAC prompt
- Install as service if approved

### Directory Layout

**System Mode (elevated):**
```
Linux:
/usr/local/bin/casnotes     # Binary
/var/lib/casnotes/          # Data
/var/log/casnotes/          # Logs
/etc/casnotes/              # Minimal (certs only)

macOS:
/usr/local/bin/casnotes
/Library/Application Support/casnotes/
/Library/Logs/casnotes/

Windows:
C:\Program Files\casnotes\
C:\ProgramData\casnotes\
```

**User Mode (fallback):**
```
Linux (XDG):
~/.local/bin/casnotes
~/.local/share/casnotes/
~/.config/casnotes/

macOS:
~/Library/Application Support/casnotes/

Windows:
%LOCALAPPDATA%\casnotes\
```

### Service Installation
- Binary creates its own service definitions
- System user with UID/GID < 999 (avoiding conflicts)
- Auto-start on boot
- No external scripts needed

### Container Detection
- Check for /.dockerenv
- Detect init: tini, dumb-init, s6-overlay
- Check PID 1 process
- Kubernetes service accounts
- Container-specific environment variables

## Environment Variables

**Limited set (everything else via WebUI):**
```bash
DATABASE_URL=postgresql://user:pass@host/db  # Optional, defaults to SQLite
PORT=64123                                    # Optional, auto-selects 64xxx
BIND=127.0.0.1                               # Optional, auto-detects
DEBUG=true                                    # Optional, default false
BASE_URL=https://notes.example.com          # Optional, auto-detects
DATA_DIR=/custom/path                        # Optional, OS-appropriate
```

**Smart defaults based on detection:**
- Reverse proxy detected → bind to 127.0.0.1
- No proxy → bind to 0.0.0.0
- Container → assume reverse proxy exists

## Features

### Note System

**Note Types:**
- Standard (Markdown/CommonMark)
- Code snippets (syntax highlighting)
- Checklists (interactive)
- Canvas (drawings)
- Encrypted notes

**Visibility Levels:**
- **Private** (default) - Owner only
- **Unlisted** - Accessible via UUID link
- **Public** - Discoverable and searchable

**Organization:**
- Notebooks/folders (unlimited nesting)
- Tags (color-coded, max 20 per note)
- Smart collections (saved searches)
- Archive & trash (30-day auto-delete)
- Pinning important notes

### Storage Structure
```
data/
├── casnotes.db           # SQLite with all settings
├── repo/                 # Git repository
│   ├── .git/
│   ├── notes/
│   │   ├── 2024-01-15-uuid.md
│   │   └── ...
│   └── attachments/
└── backups/
```

**Note Format:**
```markdown
---
id: uuid-v4
title: Note Title
created: 2024-01-15T10:30:00Z
modified: 2024-01-15T14:22:00Z
tags: [personal, todo]
color: yellow
pinned: false
archived: false
visibility: private
type: note
---

# Content here
```

### User Interface

**Views:**
- Grid (Google Keep style cards)
- List (condensed with snippets)
- Timeline (chronological)
- Code (OpenGist style)
- Editor (split markdown/preview)

**Themes:**
- **Dark** (default, Dracula-based)
- **Light** (GitHub-inspired)
- Auto (system preference)

**Editor Features:**
- Live preview
- Syntax highlighting (100+ languages)
- Markdown toolbar
- Image paste/drag-drop
- Auto-save (30 seconds)
- Templates

## User Management

### Registration & Authentication
- **Default:** Open registration
- **With SMTP:** Email validation required
- **Without SMTP:** Direct activation (with warning)
- First user becomes admin
- Must create 'administrator' account
- Modes: open, invite, closed, approval

### User Preferences (Persistent)
```sql
user_preferences:
- sort_order (modified_desc default)
- items_per_page (50 default)
- default_view (grid default)
- theme (dark default)
- editor_mode (split default)
- timezone (America/New_York default)
- time_format (24h default)
- date_format (MM/DD/YYYY default)
- week_starts (monday default)
```

## API System

### Authentication
**All standard headers supported:**
```
Authorization: Bearer cn_token
X-API-Key: cn_token
X-Auth-Token: cn_token
X-Access-Token: cn_token
Authorization: Basic base64(token:)
Api-Token: cn_token
Auth-Token: cn_token
Token: cn_token
?token=cn_token (fallback)
?api_key=cn_token
?access_token=cn_token
```

**Token Management:**
- User-friendly names
- Shown once (must copy)
- Default expiry: Never
- Stored as hash
- Format: `cn_xxxxxxxxxxxxx`

### Route Structure

**Public Routes:**
```
/                           # Index/landing
/login, /register          # Auth pages
/discover                  # Public notes feed
/notes/public/*           # Public notes
/notes/shared/*           # Unlisted notes
/privacy, /terms          # Legal pages
/security.txt            # Security contact
/.well-known/*           # Standard URIs
/healthz                 # Health check
/robots.txt              # Search engine rules
/sitemap.xml             # Sitemap
/api/v1/                # Public API
```

**User Routes:**
```
/users                    # Dashboard
/users/notes             # Notes management
/users/settings          # Settings
/users/trash             # Trash/archive
/users/settings/tokens   # API token management
/api/v1/users/*          # User API
```

**Admin Routes:**
```
/admin                    # Dashboard
/admin/users             # User management
/admin/server            # Server settings
/admin/smtp              # Email config
/admin/legal             # Legal pages
/admin/compliance        # Compliance settings
/admin/scheduler         # Scheduled tasks
/admin/tokens            # All API tokens
/api/v1/admin/*          # Admin API
```

**Support Routes:**
```
/support                  # Support hub
/support/docs            # Documentation
/support/faq             # FAQ
/support/shortcuts       # Keyboard shortcuts
/support/contact         # Contact form
/api/v1/support/docs     # Docs API
```

## SMTP & Notifications

### SMTP Configuration
**Provider Presets:** CUSTOM (default), GMAIL, YAHOO, OUTLOOK

**Smart UI:**
- Security dropdown: "587 (STARTTLS)", "465 (TLS)", "25 (NONE)"
- Port auto-fills but remains editable
- Security method persists if port changed
- Test connection with diagnostics

**Configuration Fields:**
- SMTP_HOST, SMTP_PORT
- SMTP_USER, SMTP_PASSWORD (optional)
- SMTP_FROM_NAME, SMTP_FROM_ADDRESS
- ADMIN_EMAIL

### Notification Events
- Note shared/accessed
- Storage quota warnings (80%)
- Backup completion/failure
- Certificate expiry warnings
- System updates available
- Git sync conflicts
- New device login
- Admin action audit
- Emergency alerts
- Certificate renewal alerts
- Bug report notifications

## Database Strategy

### Multi-Database Support
**SQLite (default):**
- Single file: data/casnotes.db
- WAL mode for concurrency
- Auto-backup before migrations

**PostgreSQL/MariaDB:**
- Connection via DATABASE_URL
- SQLite becomes cache/fallback
- Auto-sync when reconnected
- Queue writes during outage

### Migration System
```sql
migrations:
- Version tracking
- Checksum validation
- Automatic rollback on failure
- Progress tracking
- Error recovery
- Full transaction support
```

**Self-Healing:**
- Integrity checks
- Index rebuilding
- Orphan cleanup
- Auto-vacuum
- Corruption recovery

## Security

### Authentication Security
- bcrypt password hashing
- TOTP 2FA (RFC 6238)
- WebAuthn support
- Session management
- API tokens

### Rate Limiting
- Public: 60 req/min per IP
- Authenticated: 600 req/min per user
- API tokens: 1000 req/min per token
- Login: 5 failures = 15 min lockout
- Search: 30 searches/min

### Security Headers
- Content-Security-Policy
- X-Frame-Options: DENY
- X-Content-Type-Options: nosniff
- Strict-Transport-Security
- Referrer-Policy
- CORS default: * (configurable)

### Certificate Management
- Auto-scan `/etc/letsencrypt/live/` for existing certs
- Built-in ACME client:
  - DNS-01 (all providers + RFC2136)
  - HTTP-01
  - TLS-ALPN-01
- Certificate renewal notifications

## Compliance System

### Regulation Toggles (all disabled by default)

**GDPR (EU):**
- Right to erasure (30-day delay)
- Data portability
- Consent tracking
- Breach notifications (72hr)
- Audit trail
- Privacy by default

**HIPAA (US Healthcare):**
- Encryption mandatory
- 15-min timeout
- 6-year audit retention
- Access controls
- Unique user identification
- 90-day password rotation
- BAA required

**COPPA (US Children):**
- Age verification (13+)
- Parental consent system
- No behavioral tracking
- No third-party sharing
- Limited data collection
- Enhanced moderation

**SOX (US Public Companies):**
- Immutable audit logs (7yr)
- Change control tracking
- Segregation of duties
- Tamper-evident logging
- Executive sign-off tracking

**FERPA (US Education):**
- Educational record protection
- Parent/student access rights
- Consent tracking
- Directory information controls
- School official access only

**PCI DSS (Payment Cards):**
- No card data storage
- Strong cryptography
- 90-day password changes
- 2FA mandatory
- 15-min timeout
- Daily log reviews

**CCPA (California):**
- "Do Not Sell" option
- Deletion within 45 days
- Opt-out mechanisms
- Consumer request logging
- Data inventory

**PIPEDA (Canada):**
- Consent requirements
- Limited collection principle
- 30-day response to requests
- Breach notifications
- Privacy officer designation

**Note:** When compliance is enabled, it overrides all normal settings

## Legal Pages

**All stored in database, editable via admin:**

**security.txt:**
```
Contact: mailto:{{admin_email}}
Expires: {{one_year_from_now}}
Canonical: {{base_url}}/.well-known/security.txt
```

**robots.txt:**
```
User-agent: *
Disallow: /admin
Disallow: /api/
Disallow: /users/
Allow: /notes/public/
Allow: /discover
Sitemap: {{base_url}}/sitemap.xml
```

**Privacy Policy & Terms:**
- Markdown editor
- Version history
- Template variables
- Required acceptance tracking

## Built-in Scheduler

**Scheduled Tasks:**
```
Every 5 minutes:
- Git auto-commit
- Search index refresh
- Session cleanup

Every 30 minutes:
- Git push to remote
- Database sync
- Orphan check

Hourly:
- Token cleanup
- Email retry
- Metrics collection

Daily (3 AM):
- Database backup
- VACUUM optimize
- Temp cleanup

Weekly (Sunday 3 AM):
- Log rotation
- Integrity check
- Certificate check

Monthly (1st day):
- Access log rotation
- Trash auto-delete (30+ days)
- Usage reports
```

## Resource Limits & Defaults

### Storage Limits
- Max note: 10MB
- Max attachment: 25MB
- Max per note: 10 attachments
- User quota: 5GB
- Image optimization: >2048px

### Performance
- DB connections: CPU×2
- Worker threads: CPU×4
- Cache: 100MB or 10% RAM
- HTTP timeout: 30s
- Shutdown grace: 30s

### Sessions
- Idle timeout: 7 days
- Absolute: 30 days
- Remember me: 90 days
- Concurrent: 10 per user

### Backups
- Schedule: Daily 3 AM
- Retention: 30 daily, 12 weekly, 24 monthly
- Format: tar.gz with optional AES-256
- Verification: Automatic

### Logs
- server.log: 10MB max, weekly rotation, no archive
- error.log: 10MB max, weekly rotation, no archive
- access.log: No limit, monthly rotation, no archive
- Audit: Database, 90 days (unless compliance)

### Git Sync
- Auto-commit: Every 5 minutes if changes
- Commit message: "Auto-save: {timestamp} - {change_summary}"
- Branch: main for stable, drafts/{username} for unsaved
- Conflict resolution: Last-write-wins with backup
- Push frequency: Every 30 minutes
- History limit: Keep last 1000 commits

## Import/Export

### Import Support
- Google Keep takeout
- Joplin JEX
- Evernote ENEX
- Standard Notes
- Plain markdown
- OpenGist repos
- Max import size: 100MB

### Export Formats
- ZIP with markdown+JSON
- PDF (single/bulk)
- HTML static site
- CSV data export

## Standards Compliance

### Web Standards
- REST API (RFC 7231)
- HTTP/2 with HTTP/1.1 fallback
- OpenAPI 3.0 documentation
- WebSocket for real-time
- Standard HTTP status codes
- Content negotiation

### File Standards
- CommonMark Markdown
- Git repository format
- JSON (RFC 8259)
- CSV (RFC 4180)
- YAML 1.2

### Security Standards
- OAuth 2.0/OIDC
- JWT (RFC 7519)
- TOTP (RFC 6238)
- WebAuthn
- SMTP (RFC 5321)
- MIME (RFC 2045)

### Other Standards
- ISO 8601 timestamps
- UTF-8 encoding everywhere
- WCAG 2.1 Level AA
- Prometheus metrics
- Syslog (RFC 5424)
- XDG Base Directory
- Filesystem Hierarchy Standard

## Self-Healing Capabilities

**Automatic Recovery:**
- Database corruption repair
- Service restart on crash
- Storage cleanup on low space
- Network reconnection with backoff
- Git repository repair
- Session cleanup
- Performance optimization
- Certificate renewal
- Configuration validation
- Port conflict resolution
- Memory leak detection
- DNS cache poisoning fix
- Proxy misconfiguration bypass

**Emergency Mode:**
- Read-only fallback
- Data preservation priority
- Admin notifications via all channels
- Diagnostic collection
- Auto-recovery schedule

## Boolean Value Support

**Accepted TRUE values:**
true, yes, on, enable, enabled, active, 1, t, y

**Accepted FALSE values:**
false, no, off, disable, disabled, inactive, 0, f, n

(All case-insensitive, applied everywhere)

## Default Configuration Summary

### System Defaults
- Port: Auto-select in 64000-64999 range
- Bind: Auto-detect (127.0.0.1 with proxy, 0.0.0.0 without)
- Theme: Dark (Dracula)
- Timezone: America/New_York
- Time format: 24h
- Week starts: Monday
- Date format: MM/DD/YYYY
- CORS: * (allow all)

### Security Defaults
- Registration: Open
- Email validation: Required if SMTP configured
- Password minimum: 8 characters
- Session timeout: 7 days idle
- API tokens: Never expire
- Rate limiting: Enabled
- 2FA: Optional

### Performance Defaults
- Items per page: 50
- Search results: 25
- Cache size: 100MB or 10% RAM
- Worker threads: CPU cores × 4
- Database connections: CPU cores × 2

### All Compliance: Disabled by default

## Success Criteria

- Single binary runs on all platforms
- Zero configuration required
- All features in v1.0.0
- Smart defaults for everything
- Self-healing and resilient
- Standards-compliant
- Enterprise-ready security
- Regulation compliance options
- Excellent user experience
- Unattended operation
- Complete WebUI administration

---

**Version:** 1.0.0  
**Status:** Complete Specification  
**License:** MIT  
**Organization:** casapps  
**Registry:** ghcr.io/casapps/casnotes

