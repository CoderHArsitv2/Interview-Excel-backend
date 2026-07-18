# Interview Excel — Backend

> Go-based REST API that powers the **Interview Excel** platform, connecting students with industry experts for mock interviews, session booking, and payments.

---

## Table of Contents

- [Tech Stack](#tech-stack)
- [Architecture Overview](#architecture-overview)
- [Project Structure](#project-structure)
- [Data Models & Relationships](#data-models--relationships)
- [API Routes](#api-routes)
- [Authentication & Authorization](#authentication--authorization)
- [Configuration Management](#configuration-management)
- [Database & Migrations](#database--migrations)
- [Redis](#redis)
- [Docker](#docker)
- [CI/CD Pipeline](#cicd-pipeline)
- [Environment Setup](#environment-setup)
- [Deployment Architecture](#deployment-architecture)
- [Interview Talking Points](#interview-talking-points)

---

## Tech Stack

| Layer           | Technology                                                                 |
| --------------- | -------------------------------------------------------------------------- |
| **Language**    | Go 1.23                                                                    |
| **Framework**   | [Gin](https://github.com/gin-gonic/gin) — high-performance HTTP router    |
| **ORM**         | [GORM](https://gorm.io/) — Go ORM with AutoMigrate                        |
| **Database**    | PostgreSQL (local dev) / [Neon](https://neon.tech) (staging & production)  |
| **Cache**       | Redis (optional, via [go-redis](https://github.com/go-redis/redis))        |
| **Auth**        | JWT (access + refresh tokens) with Google OAuth 2.0                        |
| **Payments**    | [Razorpay](https://razorpay.com/) payment gateway                         |
| **Video Calls** | Jitsi Meet (room link generation)                                          |
| **Container**   | Multi-stage Docker (Alpine builder → distroless runtime)                   |
| **CI/CD**       | GitHub Actions (CI → staging auto-deploy → manual prod deploy)             |
| **Staging**     | [Render](https://render.com/) (Web Service from Dockerfile)                |
| **Production**  | AWS App Runner + ECR                                                       |

---

## Architecture Overview

```
┌──────────────┐       ┌─────────────────────────────────────────────┐
│   Frontend   │       │              Backend (Go / Gin)             │
│  (Next.js)   │◄─────►│                                             │
└──────────────┘  CORS │  ┌─────────┐  ┌────────────┐  ┌──────────┐ │
                       │  │ Routes  │─►│ Controllers│─►│ Services │ │
                       │  └─────────┘  └────────────┘  └──────────┘ │
                       │       │              │              │       │
                       │  ┌─────────┐         │        ┌──────────┐ │
                       │  │Middleware│         │        │  Models  │ │
                       │  │  (JWT)  │         │        │  (GORM)  │ │
                       │  └─────────┘         │        └──────────┘ │
                       │                      │              │       │
                       │               ┌──────┴──────┐       │       │
                       │               │   Config    │       │       │
                       │               │  (Runtime)  │       │       │
                       │               └──────┬──────┘       │       │
                       └──────────────────────┼──────────────┼───────┘
                                              │              │
                              ┌───────────────┼──────────────┼──────┐
                              │               ▼              ▼      │
                              │  ┌──────────────┐   ┌────────────┐  │
                              │  │    Redis      │   │ PostgreSQL │  │
                              │  │  (optional)   │   │   / Neon   │  │
                              │  └──────────────┘   └────────────┘  │
                              └─────────────────────────────────────┘
```

**Request flow:** Client → Gin Router → Middleware (CORS, JWT Auth) → Controller → Service/Model → Database → Response

---

## Project Structure

```
interview-excel-backend/
├── main.go                  # Entry point — CLI (serve / migrate), router setup, health check
├── Dockerfile               # Multi-stage build (Alpine → distroless)
├── go.mod / go.sum          # Go module dependencies
│
├── config/
│   ├── runtime.go           # Environment variable loading (singleton pattern with sync.Once)
│   ├── config.go            # DB init (GORM), GORM AutoMigrate, Razorpay client, Google OAuth
│   ├── redis.go             # Redis client initialization (optional, TLS-aware)
│   └── logger.go            # Custom GORM logger using logrus
│
├── routes/
│   ├── auth.go              # Public auth routes (register, signin, Google OAuth, refresh)
│   ├── expert.go            # Protected expert routes (profile, slots, dashboard)
│   └── student.go           # Protected student routes (profile, experts, bookings, sessions)
│
├── controllers/
│   ├── auth.go              # Signup, signin, JWT token generation
│   ├── google.go            # Google OAuth callback handler
│   ├── logout.go            # Logout with token blacklisting
│   ├── user.go              # Get current user from JWT
│   ├── expert.go            # Expert profile CRUD, slot management, dashboard
│   ├── student.go           # Student profile CRUD, expert browsing
│   ├── booking.go           # Slot booking initiation
│   ├── booking_helpers.go   # Booking validation, Razorpay order creation
│   ├── payment.go           # Razorpay payment verification & confirmation
│   ├── object.go            # Shared response/request object builders
│   ├── jitsi.go             # Jitsi Meet room link generation
│   └── base.go              # Base controller utilities
│
├── models/
│   ├── user.go              # User model + UserRepo (CRUD, auth queries)
│   ├── expert.go            # Expert profile model
│   ├── student.go           # Student profile model
│   ├── session.go           # Interview session model
│   ├── availability.go      # AvailabilitySlot model (expert time slots)
│   ├── payment.go           # Payment model (Razorpay orders)
│   ├── wallet.go            # Wallet model (user balance)
│   ├── wallet_transaction.go# Wallet transaction history
│   ├── service.go           # Service/specialization model
│   ├── interface.go         # Repository interfaces (dependency inversion)
│   └── migration.go         # Models list for AutoMigrate
│
├── middleware/
│   └── auth.go              # JWT middleware — validates Bearer token, checks blacklist, sets user context
│
├── utils/
│   ├── token_utils.go       # JWT generation, refresh, blacklisting helpers
│   ├── slot_generator.go    # Weekly availability slot generation logic
│   ├── availability_calender.go  # Calendar-based availability utilities
│   └── helpers.go           # General utility functions
│
├── pkg/errors/
│   └── main.go              # Centralized error logging package
│
├── docs/
│   └── ci-cd.md             # CI/CD documentation
│
└── .github/workflows/
    ├── backend-ci.yml            # CI: format check, go vet, tests
    ├── backend-staging-deploy.yml # Auto-deploy to staging (Render) on main
    └── backend-prod-deploy.yml    # Manual deploy to production (AWS)
```

---

## Data Models & Relationships

```
┌──────────┐
│   User   │
│──────────│
│ id (PK)  │
│ user_uuid│ (unique)
│ full_name│
│ email    │ (unique)
│ password │ (nullable — Google OAuth users have none)
│ phone    │ (unique, nullable)
│ role     │ ("student" | "expert" | "admin")
│ picture  │
└────┬─────┘
     │ 1:1 (user_uuid → UserID)
     ├──────────────────────┐
     ▼                      ▼
┌──────────┐          ┌──────────┐
│  Student │          │  Expert  │
│──────────│          │──────────│
│ user_id  │ (FK)     │ user_id  │ (FK)
│ profile  │          │ profile  │
│ fields.. │          │ fields.. │
└──────────┘          └────┬─────┘
                           │ 1:N
                           ▼
                    ┌──────────────────┐
                    │ AvailabilitySlot │
                    │──────────────────│
                    │ expert_id        │
                    │ start_time       │
                    │ end_time         │
                    │ is_booked        │
                    └────────┬─────────┘
                             │ (when booked)
                             ▼
                    ┌──────────────────┐
                    │     Session      │
                    │──────────────────│       ┌─────────────┐
                    │ session_uuid     │       │   Payment   │
                    │ student_uuid     │──────►│─────────────│
                    │ expert_uuid      │       │ order_id    │
                    │ slot_id          │       │ payment_id  │
                    │ status           │       │ amount      │
                    └──────────────────┘       └─────────────┘

┌──────────┐       ┌─────────────────────┐
│  Wallet  │ 1:N   │  WalletTransaction  │
│──────────│──────►│─────────────────────│
│ user_uuid│       │ wallet_id           │
│ balance  │       │ type (credit/debit) │
└──────────┘       │ amount              │
                   │ reference_id        │
                   └─────────────────────┘
```

**Key design decisions:**
- **UUID-based references** — `user_uuid` is used for cross-model references instead of auto-increment IDs, making the system more secure and frontend-friendly.
- **Soft deletes** — `User` model uses GORM's `DeletedAt` for soft deletes.
- **Repository pattern** — Each model has a repository with an interface (`models/interface.go`), enabling dependency inversion and testability.
- **Nullable password** — Google OAuth users don't have a password, so `Password` is a `*string`.

---

## API Routes

### Public (No Auth Required)

| Method | Path                    | Description                        |
| ------ | ----------------------- | ---------------------------------- |
| POST   | `/auth/register`        | Email/password signup              |
| POST   | `/auth/signin`          | Email/password login               |
| POST   | `/auth/google/login`    | Google OAuth sign-in               |
| POST   | `/auth/user`            | Get user from token (body)         |
| GET    | `/auth/refresh`         | Refresh JWT session                |
| GET    | `/healthz`              | Health check (DB + Redis status)   |

### Expert Routes (JWT Protected)

| Method | Path                            | Description                        |
| ------ | ------------------------------- | ---------------------------------- |
| GET    | `/expert/profile`               | Get expert's own profile           |
| PUT    | `/expert/profile`               | Update expert profile              |
| POST   | `/expert/generate-slots`        | Generate weekly availability slots |
| GET    | `/expert/my-slots`              | Get available slots (expert view)  |
| GET    | `/expert/all-slots`             | Get all slots (including booked)   |
| DELETE | `/expert/availability/:slot_id` | Cancel a specific slot             |
| GET    | `/expert/dashboard`             | Expert dashboard metrics           |

### Student Routes (JWT Protected)

| Method | Path                              | Description                       |
| ------ | --------------------------------- | --------------------------------- |
| GET    | `/student/profile`                | Get student's own profile         |
| PUT    | `/student/profile`                | Update student profile            |
| GET    | `/student/experts`                | Browse all experts                |
| GET    | `/student/expert/:id/slots`       | View expert's available slots     |
| POST   | `/student/book-slot/:slot_id`     | Initiate booking + Razorpay order |
| POST   | `/student/confirm-booking`        | Confirm payment & create session  |
| GET    | `/student/sessions`               | List student's sessions           |

---

## Authentication & Authorization

### Flow

```
1. User signs up via /auth/register (or /auth/google/login for OAuth)
2. Server returns JWT access token + refresh token in cookies
3. Client sends "Authorization: Bearer <token>" on every request
4. AuthMiddleware parses JWT, checks expiry, checks blacklist (Redis)
5. Sets user_uuid and role in Gin context for downstream handlers
6. /auth/refresh issues new access token using refresh token
```

### JWT Structure (Claims)

```go
type Claims struct {
    UserID string `json:"user_uuid"`   // UUID of authenticated user
    Role   string `json:"role"`        // "student" | "expert" | "admin"
    jwt.RegisteredClaims               // exp, iat, etc.
}
```

### Token Blacklisting

On logout, the token is added to a blacklist. The `AuthMiddleware` checks this blacklist before accepting any token. If Redis is disabled, blacklisting still works via an in-memory fallback.

### Google OAuth 2.0

- Uses `golang.org/x/oauth2` for the OAuth dance
- Configured via `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL`
- On successful OAuth, creates or finds the user and issues a JWT

---

## Configuration Management

### How It Works

Configuration is managed via **environment variables** loaded through the `config/runtime.go` singleton:

```go
// Loaded once via sync.Once — thread-safe singleton
RuntimeConfig() Runtime
```

**Loading order:**
1. `.env` file loaded via `godotenv` (local development only)
2. OS environment variables override `.env` values
3. Built-in defaults for optional values (`PORT=8080`, `REDIS_ENABLED=false`, etc.)

### Key Environment Variables

| Variable                | Required | Description                                          |
| ----------------------- | -------- | ---------------------------------------------------- |
| `APP_ENV`               | No       | `development` / `staging` / `production`             |
| `PORT`                  | No       | Server port (default: `8080`)                        |
| `DATABASE_URL`          | **Yes**  | Postgres connection string                           |
| `JWT_SECRET`            | **Yes**  | Secret key for JWT signing                           |
| `GOOGLE_CLIENT_ID`      | **Yes**  | Google OAuth client ID                               |
| `GOOGLE_CLIENT_SECRET`  | **Yes**  | Google OAuth client secret                           |
| `GOOGLE_REDIRECT_URL`   | No       | OAuth callback URL                                   |
| `RAZORPAY_KEY`          | **Yes**  | Razorpay API key                                     |
| `RAZORPAY_SECRET`       | **Yes**  | Razorpay secret key                                  |
| `REDIS_ENABLED`         | No       | Enable Redis (`true`/`false`)                        |
| `REDIS_ADDR`            | If Redis | Redis server address                                 |
| `REDIS_PASSWORD`        | If Redis | Redis password                                       |
| `REDIS_USE_TLS`         | No       | Enable TLS for Redis (for Upstash)                   |
| `COOKIE_DOMAIN`         | No       | Domain for auth cookies                              |
| `COOKIE_SECURE`         | No       | Secure flag for cookies                              |
| `CORS_ALLOWED_ORIGINS`  | No       | Comma-separated allowed origins                      |

### Planned: YAML Config Files

We are migrating to environment-specific YAML config files (`config/staging.yaml`, `config/development.yaml`, `config/production.yaml`) for non-secret defaults. Environment variables will **always override** YAML values. See [Implementation Plan](#deployment-architecture) for details.

---

## Database & Migrations

### Connection

The app supports two ways to specify the database connection:

1. **`DATABASE_URL`** (preferred) — full Postgres connection string
2. **Individual vars** (`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_NAME`, `DB_PASSWORD`, `DB_SSLMODE`) — fallback

```go
// config/runtime.go — DatabaseDSN() picks DATABASE_URL if set, otherwise builds DSN from parts
```

### GORM AutoMigrate

Migrations run automatically via `go run . migrate`:

```go
// models/migration.go — all models registered for migration
var modelsForMigration = []interface{}{
    &User{}, &Expert{}, &AvailabilitySlot{}, &Payment{},
    &Student{}, &Session{}, &Wallet{}, &WalletTransaction{},
}
```

**Why AutoMigrate:**
- Creates tables, adds missing columns, creates indexes
- Does **not** delete columns or change types (safe for production)
- Runs as a separate CLI command (`migrate`) before starting the server

### Neon (Serverless Postgres)

For staging and production, we use **Neon** — a serverless Postgres service:
- **Connection strings** look like: `postgres://user:pass@ep-xxx.region.aws.neon.tech/dbname?sslmode=require`
- **`sslmode=require`** is mandatory for Neon
- Neon supports branching (useful for preview environments)
- Neon auto-scales to zero when idle (cost-effective for staging)

---

## Redis

Redis is **optional** and used for:
- **Token blacklisting** — storing invalidated JWTs on logout
- **Session caching** (planned)

```go
// config/redis.go — skips init if REDIS_ENABLED=false
if !runtimeConfig.RedisEnabled {
    RedisClient = nil
    return nil
}
```

For staging/production, we use **Upstash** (serverless Redis) with TLS enabled (`REDIS_USE_TLS=true`).

---

## Docker

### Multi-stage Build

```dockerfile
# Stage 1: Build (Alpine — small image with Go toolchain)
FROM golang:1.23.2-alpine AS builder
# ... compile to static binary with CGO_ENABLED=0

# Stage 2: Runtime (distroless — minimal, secure, no shell)
FROM gcr.io/distroless/base-debian12
# ... copy only the compiled binary
ENTRYPOINT ["/app/interviewexcel-backend"]
CMD ["serve"]
```

**Why this setup:**
- **Alpine builder** — Small base image for fast CI builds
- **`CGO_ENABLED=0`** — Static binary, no C dependencies needed at runtime
- **distroless** — Minimal attack surface (no shell, no package manager)
- Final image is ~15-20 MB vs ~800 MB for a full Go image

### Build & Run Locally

```bash
# Build
docker build -t ie-backend .

# Run
docker run -p 8080:8080 --env-file .env ie-backend
```

---

## CI/CD Pipeline

### Pipeline Overview

```
                  PR / push to any branch
                          │
                          ▼
              ┌──────────────────────┐
              │   Backend CI         │
              │ (backend-ci.yml)     │
              │                      │
              │ • gofmt check        │
              │ • go vet             │
              │ • go test ./...      │
              └──────────┬───────────┘
                         │ (on main, if CI passes)
                         ▼
              ┌──────────────────────┐
              │ Staging Deploy       │
              │ (backend-staging-    │
              │  deploy.yml)         │
              │                      │
              │ • Run migrations     │
              │   (against Neon)     │
              │ • Trigger Render     │
              │   deploy hook        │
              │ • Smoke test /healthz│
              └──────────────────────┘

              ┌──────────────────────┐
              │ Production Deploy    │  (manual trigger)
              │ (backend-prod-       │
              │  deploy.yml)         │
              │                      │
              │ • Run migrations     │
              │   (against Neon)     │
              │ • Build Docker image │
              │ • Push to AWS ECR    │
              │ • Trigger App Runner │
              │ • Smoke test /healthz│
              └──────────────────────┘
```

### Workflow Details

#### 1. `backend-ci.yml` — Continuous Integration

Runs on every PR and push to `main`:
- **Format check** — `gofmt -l .` ensures consistent code formatting
- **Static analysis** — `go vet ./...` catches common bugs
- **Tests** — `go test ./...` runs all unit tests

#### 2. `backend-staging-deploy.yml` — Staging Auto-Deploy

Triggered automatically when `Backend CI` succeeds on `main`:
- **Checkout** at the exact commit SHA that passed CI
- **Run migrations** against the staging Neon database (`STAGING_DATABASE_URL`)
- **Trigger Render** deploy hook via `curl POST` — Render pulls the latest code, builds from `Dockerfile`, and deploys
- **Smoke test** — polls `STAGING_HEALTHCHECK_URL` every 10 seconds for up to 4 minutes

#### 3. `backend-prod-deploy.yml` — Production Manual Deploy

Triggered via `workflow_dispatch` (manual button in GitHub):
- Runs production migrations against production Neon database
- Builds Docker image and pushes to **AWS ECR** with commit SHA tag + `production` tag
- Triggers **AWS App Runner** deployment
- Smoke tests production health endpoint

### GitHub Secrets Required

| Secret                          | Environment | Purpose                               |
| ------------------------------- | ----------- | ------------------------------------- |
| `STAGING_DATABASE_URL`          | Staging     | Neon connection string for staging    |
| `RENDER_STAGING_DEPLOY_HOOK_URL`| Staging     | Render deploy hook URL                |
| `STAGING_HEALTHCHECK_URL`       | Staging     | Staging health endpoint URL           |
| `PRODUCTION_DATABASE_URL`       | Production  | Neon connection string for production |
| `PRODUCTION_HEALTHCHECK_URL`    | Production  | Production health endpoint URL        |
| `AWS_ROLE_TO_ASSUME`            | Production  | AWS IAM role ARN (OIDC)               |
| `AWS_REGION`                    | Production  | AWS region for ECR/App Runner         |
| `ECR_REPOSITORY`                | Production  | ECR repository name                   |
| `APP_RUNNER_SERVICE_ARN`        | Production  | App Runner service ARN                |

---

## Environment Setup

### Local Development

```bash
# 1. Clone the repository
git clone https://github.com/CoderHArsit/interview-excel-backend.git
cd interview-excel-backend

# 2. Copy environment file
cp .env.example .env
# Fill in the values in .env

# 3. Start local Postgres (or use Neon dev branch)
# Make sure DATABASE_URL points to your DB

# 4. Run migrations
go run . migrate

# 5. Start the server
go run . serve
# Server starts on http://localhost:8080

# 6. Health check
curl http://localhost:8080/healthz
# {"status":"ok","db":"ok","redis":"disabled"}
```

### CLI Commands

| Command              | Description                                     |
| -------------------- | ----------------------------------------------- |
| `go run . serve`     | Start the API server                            |
| `go run . migrate`   | Run GORM AutoMigrate (create/update tables)     |

---

## Deployment Architecture

### Staging (Render + Neon)

```
GitHub (main branch)
    │
    ▼ (CI passes)
┌────────────────────────────────┐
│  GitHub Actions                │
│  1. Run migrations on Neon DB  │
│  2. POST to Render deploy hook │
└────────────────┬───────────────┘
                 │
                 ▼
┌────────────────────────────────┐     ┌──────────────────┐
│  Render Web Service            │────►│  Neon PostgreSQL  │
│  • Builds from Dockerfile      │     │  (staging branch) │
│  • Env vars set in dashboard   │     │  sslmode=require  │
│  • Auto-scales                 │     └──────────────────┘
│  • Free/starter tier friendly  │
└────────────────────────────────┘

Render env vars to set:
  APP_ENV=staging
  DATABASE_URL=postgres://...@neon.tech/...?sslmode=require
  JWT_SECRET=...
  GOOGLE_CLIENT_ID=...
  GOOGLE_CLIENT_SECRET=...
  GOOGLE_REDIRECT_URL=https://staging-api.../google_callback
  RAZORPAY_KEY=...
  RAZORPAY_SECRET=...
  CORS_ALLOWED_ORIGINS=https://staging.interviewexcel.com
  COOKIE_DOMAIN=.staging.interviewexcel.com
  COOKIE_SECURE=true
```

### Production (AWS App Runner + Neon)

```
GitHub (manual dispatch)
    │
    ▼
┌────────────────────────────────┐
│  GitHub Actions                │
│  1. Run migrations on Neon DB  │
│  2. Build & push to ECR       │
│  3. Trigger App Runner deploy  │
└────────────────┬───────────────┘
                 │
                 ▼
┌────────────────────────────────┐     ┌──────────────────┐
│  AWS App Runner                │────►│  Neon PostgreSQL  │
│  • Pulls from ECR              │     │  (production)     │
│  • Auto-scales                 │     └──────────────────┘
│  • HTTPS by default            │
└────────────────────────────────┘
```

### Rollback Strategy

| Environment | How to Rollback                                                             |
| ----------- | --------------------------------------------------------------------------- |
| **Staging** | Redeploy the previous successful build from Render dashboard                |
| **Prod**    | Retag/redeploy previous ECR image, rerun `start-deployment`                 |
| **DB**      | Neon restore from backup / point-in-time recovery before rerunning deploy   |

---

## Interview Talking Points

### Architecture Decisions — *"Why did you choose X?"*

1. **Why Go + Gin?**
   - Go's goroutine model handles concurrent requests efficiently without thread management
   - Gin is one of the fastest Go HTTP frameworks with zero-allocation router
   - Static typing catches bugs at compile time; deploying a single binary simplifies ops

2. **Why GORM instead of raw SQL?**
   - AutoMigrate makes schema evolution easy during rapid development
   - Interface-based repository pattern (`models/interface.go`) allows swapping implementations for testing
   - Trade-off: less control over query optimization, but acceptable for this scale

3. **Why Neon over a traditional managed Postgres (RDS)?**
   - **Serverless** — scales to zero when idle, perfect for staging (cost savings)
   - **Branching** — can create DB branches for feature testing
   - **Standard Postgres** — no driver changes needed, just a different connection string
   - **Free tier** — generous for early-stage projects

4. **Why Render for staging?**
   - Simple Docker-based deployments with deploy hooks for CI/CD integration
   - Free/starter tier suitable for staging workloads
   - Keeps staging infra separate from production (AWS), reducing blast radius

5. **Why multi-stage Docker with distroless?**
   - Build stage (Alpine) has Go toolchain; runtime stage (distroless) has only the binary
   - Final image is ~15-20 MB — fast pulls, small attack surface
   - No shell in distroless = no shell-based exploits

6. **Why separate `serve` and `migrate` CLI commands?**
   - Migrations run in CI before deploy (fail-fast if schema change breaks)
   - Server startup doesn't block on migrations
   - Can run migrations independently for debugging

### Design Patterns — *"How does X work?"*

7. **Singleton configuration loading**
   - `sync.Once` ensures `RuntimeConfig()` loads env vars exactly once, thread-safe
   - Avoids repeated file I/O and parsing on every request

8. **Repository pattern with interfaces**
   - Each model has a concrete repo and an interface in `models/interface.go`
   - Controllers depend on interfaces, not concrete types → testable with mocks
   - Separation of concerns: controllers handle HTTP, repos handle data access

9. **JWT with token blacklisting**
   - Stateless auth via JWT (no session store needed)
   - Blacklist in Redis handles logout (JWTs are otherwise irrevocable until expiry)
   - Middleware extracts `user_uuid` and `role` into Gin context for all downstream handlers

10. **Booking flow with Razorpay**
    - Two-phase commit: `InitiateBooking` creates Razorpay order → `ConfirmPayment` verifies signature
    - Prevents double-booking via slot status check before order creation
    - Payment verification happens server-side (Razorpay webhook signature)

### CI/CD & DevOps — *"How do you deploy?"*

11. **Three-workflow pipeline**
    - CI runs on every PR (format, vet, test) — fast feedback loop
    - Staging auto-deploys on merge to main — zero manual steps
    - Production requires manual trigger — intentional human gate for safety

12. **Migrations in CI, not at startup**
    - `go run . migrate` runs as a GitHub Actions step with the target DB's connection string
    - If migration fails, deploy is aborted — database and code stay in sync
    - No migration lock contention at server startup in multi-instance deployments

13. **Health check design**
    - `/healthz` checks DB connectivity (`sqlDB.Ping()`) and Redis connectivity
    - Returns `200 OK` if healthy, `503 Service Unavailable` if degraded
    - Used by Render/App Runner for readiness probes and by CI smoke tests

### Scaling Considerations — *"What would you do differently at scale?"*

14. **Connection pooling** — GORM uses `database/sql` under the hood; configure `SetMaxOpenConns`, `SetMaxIdleConns` for high traffic
15. **Read replicas** — Neon supports read replicas for read-heavy workloads
16. **Rate limiting** — Add Gin middleware for API rate limiting
17. **Structured logging** — Migrate from `log.Println` to structured JSON logs (logrus is already imported)
18. **Database migrations** — Move from AutoMigrate to versioned migrations (e.g., `golang-migrate/migrate`) for production safety

---

## License

Private — Interview Excel © 2025-2026
