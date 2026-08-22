# WarBay — Simple E-Commerce API

> A clean-architecture Go backend for a small warung (shop) where buyers order **without accounts** (guest checkout), and the admin manages products & categories.

---

## 🎯 Project Goals

- **Portfolio-ready** Go backend showcasing clean architecture, proper error handling, concurrency-safe stock management, and file uploads.
- **Real-world patterns**: pessimistic locking (`SELECT FOR UPDATE`), transactional order creation with snapshot pricing, status state machine, JWT role-based auth.
- **Interview talking points**: every design decision has a "why" (see [Architecture Decisions](#-architecture-decisions)).

---

## ✨ Features

| Domain | Capabilities |
|--------|--------------|
| **Auth** | Register, Login (JWT), Role-based access (ADMIN / MEMBER) |
| **Categories** | Admin CRUD, public list/detail |
| **Products** | Admin CRUD + **image upload (multipart)**, public browse, slug-based lookup |
| **Orders** | **Guest checkout** (name, phone, COD/TRANSFER/QRIS), stock decrement with row lock, snapshot price in order_items, status machine (PENDING → CONFIRMED → READY → COMPLETED, cancellable anytime with auto-restock) |
| **Health** | `/health/live` (liveness), `/health/ready` (readiness: DB + Redis) |
| **Static Files** | Uploaded images served at `/public/product/<uuid>.jpg` |

---

## 🏗️ Architecture

```
cmd/api/
  main.go                 # Bootstrap, wiring
internal/
  config/                 # Viper config, Fiber setup, DB, Redis, JWT, Static
  controller/             # HTTP layer: bind, validate, response
  service/                # Business logic, transactions, validation
  repository/             # GORM data access (interfaces + impl)
  model/                  # Request/Response DTOs, converters
  entity/                 # GORM entities (Money, Order, Product, etc.)
  pkg/
    apperror/             # Structured error types + HTTP mapping
    filevalidation/       # MIME sniffing, size limit, safe naming
    slugutil/             # Slugify helper
    phonenumber/          # Indonesian phone → E.164
    slug/                 # (internal) slug utilities
  storage/                # FileStorage interface + LocalStorage impl
  route/                  # Route registration
  middleware/             # JWT auth, admin guard, recover
```

**Principles applied:**
- **Dependency inversion**: services depend on repository/storage interfaces.
- **Single responsibility**: each layer does one thing.
- **No logic in controllers** — only binding & response.
- **No DB calls in services** — only via repositories.
- **Errors as values**: `AppError` with code/message, wrapped for context.

---

## 🛠️ Tech Stack

| Layer | Library |
|-------|---------|
| HTTP | [Fiber v3](https://gofiber.io) |
| ORM | [GORM v2](https://gorm.io) (PostgreSQL) |
| Cache/Session | [Redis](https://redis.io) (go-redis/v9) |
| Auth | JWT (HS256) with custom claims |
| Validation | go-playground/validator v10 |
| Config | spf13/viper (YAML + env override) |
| Logging | logrus (JSON, structured) |
| UUID | google/uuid |
| Money | Custom `entity.Money` (int64 cents, JSON = number) |

---

## 📁 Project Structure

```
.
├── cmd/api/main.go
├── config.yaml              # (gitignored) local config
├── config.example.yaml      # template
├── go.mod / go.sum
├── internal/
│   ├── config/
│   ├── controller/
│   ├── service/
│   ├── repository/
│   ├── model/
│   ├── entity/
│   ├── pkg/
│   ├── storage/
│   └── route/
├── public/                  # (gitignored) uploaded images
│   └── product/
├── Dockerfile
├── docker-compose.yml
└── README.md
```

---

## ⚙️ Configuration

`config.yaml` (copy from `config.example.yaml`):

```yaml
app:
  name: "golang e commerce"
  host: "localhost"
  port: 8080

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: ""
  dbname: "simple_ecommerce"
  sslmode: "disable"

redis:
  host: "localhost"
  port: 6379
  password: ""
  dbname: 0

jwt:
  secret_key: "your-super-secret-key-change-in-production"

logger:
  level: 4          # logrus.DebugLevel = 4, Info = 5
  file_path: "logs/app.log"

storage:
  base_path: "./public"
  public_path: "/public"
```

> **Secrets**: never commit `config.yaml`. Use environment variables in production (Viper `AutomaticEnv()` is wired).

---

## 🚀 Quick Start

### Prerequisites
- Go 1.22+
- PostgreSQL 15+
- Redis 7+

### 1. Clone & Configure
```bash
git clone https://github.com/mrbayss/golang-simple-ecommerce.git
cd golang-simple-ecommerce
cp config.example.yaml config.yaml
# edit config.yaml with your DB/Redis credentials
```

### 2. Run Database (Docker)
```bash
docker compose up -d postgres redis
# or start your own Postgres/Redis
```

### 3. Run Migrations & Server
```bash
go run ./cmd/api
# or build first
go build -o bin/warbay ./cmd/api
./bin/warbay
```

Server starts at `http://localhost:8080`.

---

## 🐳 Docker (Production-style)

```bash
# Build image
docker build -t warbay:latest .

# Run with compose (app + postgres + redis)
docker compose up -d
```

`docker-compose.yml` includes:
- `app` (Go binary, waits for DB/Redis healthchecks)
- `postgres` (with init SQL for uuid-ossp)
- `redis`

---

## 📚 API Reference

Base URL: `http://localhost:8080/api/v1`

### Auth
| Method | Endpoint | Auth | Body | Description |
|--------|----------|------|------|-------------|
| POST | `/register` | — | `{email, password}` | Register (default role: MEMBER) |
| POST | `/login` | — | `{email, password}` | Login → returns `access_token` |

### Public (No Auth)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/categories` | List categories (paginated) |
| GET | `/categories/:id` | Category detail |
| GET | `/products` | List products (paginated, `?page=&limit=`) |
| GET | `/products/:id` | Product by ID |
| GET | `/products/slug/:slug` | Product by slug |
| POST | `/orders` | **Guest checkout** |
| GET | `/orders/:code` | Check order status by code |

### Admin (Requires `Authorization: Bearer <token>` + ADMIN role)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/admin/categories` | Create category |
| PUT | `/admin/categories/:id` | Update category |
| DELETE | `/admin/categories/:id` | Delete category |
| POST | `/admin/products` | Create product (multipart) |
| PUT | `/admin/products/:id` | Update product (multipart) |
| DELETE | `/admin/products/:id` | Soft delete product |
| DELETE | `/admin/images/:image_id` | Delete single product image |
| GET | `/admin/orders` | List all orders (paginated) |
| PUT | `/admin/orders/:id/status` | Update order status |

### Health
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health/live` | Liveness probe |
| GET | `/health/ready` | Readiness (DB + Redis) |
| GET | `/api/v1/health` | Detailed health |

---

## 📦 Request/Response Examples

### Register
```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@warbay.id","password":"rahasia123"}'
```

### Login → Get Token
```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@warbay.id","password":"rahasia123"}'
# Response: { "data": { "access_token": "eyJhbG...", "expired_at": "..." } }
```

### Admin: Create Product with Images
```bash
TOKEN="<your-access-token>"

curl -X POST http://localhost:8080/api/v1/admin/products \
  -H "Authorization: Bearer $TOKEN" \
  -F "name=Nasi Goreng Spesial" \
  -F "price=25000" \
  -F "stock=30" \
  -F "category_id=9a475b7a-92b3-4850-98bd-d53a95e3a307" \
  -F "images=@/path/to/photo1.jpg" \
  -F "images=@/path/to/photo2.jpg"
```
**Response (201):**
```json
{
  "success": true,
  "message": "product created successfully",
  "data": {
    "id": "uuid",
    "name": "Nasi Goreng Spesial",
    "slug": "nasi-goreng-spesial",
    "price": 25000,
    "stock": 30,
    "images": [
      { "id": "uuid", "image_url": "/public/product/abc123.jpg", "is_primary": true },
      { "id": "uuid", "image_url": "/public/product/def456.jpg", "is_primary": false }
    ]
  }
}
```

### Guest Checkout
```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_name": "Budi",
    "customer_phone": "0812-3456-789",
    "payment_method": "COD",
    "items": [
      { "product_id": "1bd451c8-14e0-4f6e-a27b-cf5705c2de22", "quantity": 2 },
      { "product_id": "ee336e31-8a80-41a5-ab22-4be660d8f9ac", "quantity": 1 }
    ]
  }'
```
**Response (201):**
```json
{
  "success": true,
  "message": "order created successfully",
  "data": {
    "order_code": "WB-20260822-0005",
    "status": "PENDING",
    "total_price": 35000,
    "items": [
      { "product_name": "Bakso Biasa", "product_price": 15000, "quantity": 2, "subtotal": 30000 },
      { "product_name": "Es Teh Manis", "product_price": 5000, "quantity": 1, "subtotal": 5000 }
    ]
  }
}
```

### Admin: Update Order Status
```bash
curl -X PUT http://localhost:8080/api/v1/admin/orders/<order-id>/status \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status": "CONFIRMED"}'
```
**Valid transitions:**
```
PENDING → CONFIRMED | CANCELLED
CONFIRMED → READY | CANCELLED
READY → COMPLETED | CANCELLED
COMPLETED → (final)
CANCELLED → (final, stock restored)
```

---

## 🔒 Security Highlights

| Area | Implementation |
|------|----------------|
| **Passwords** | bcrypt (cost 10) |
| **JWT** | HS256, 24h expiry, claims: `user_id`, `email`, `role`, `session_id` |
| **Auth Middleware** | Validates `Bearer <token>`, populates `ctx.Locals` |
| **Admin Guard** | Separate middleware, checks `role == "ADMIN"` |
| **File Upload** | Allowlist (jpg/jpeg/png/webp), MIME sniffing, 5MB limit, UUID names |
| **SQL Injection** | GORM parameterized queries |
| **Concurrency** | `SELECT FOR UPDATE` on product row during order creation |

---

## 🧪 Testing

```bash
# Unit tests (when added)
go test ./internal/service/...

# Manual e2e test script
./scripts/e2e_test.sh   # creates admin, category, product, order, verifies stock
```

---

## 📖 Architecture Decisions (Interview Talking Points)

| Decision | Why |
|----------|-----|
| **Guest checkout (no user account)** | Warung use-case: buyers walk in, order, pay, leave — no signup friction. |
| **Snapshot price in `order_items`** | Product price can change; order must reflect price *at time of purchase*. |
| **Pessimistic lock (`FOR UPDATE`)** | Prevents overselling under concurrent orders; row locked until transaction commits. |
| **Status state machine** | Prevents invalid transitions (e.g. COMPLETED → PENDING). Enforced in service, not DB. |
| **Money as int64 cents** | Avoids floating-point errors; JSON serializes as number (15000 = Rp15.000). |
| **Soft delete (GORM `DeletedAt`)** | Audit trail; orders reference products that may be "deleted". |
| **FileStorage interface** | Swappable: LocalStorage now, S3/R2 later without touching service layer. |
| **MIME sniffing not header** | Attackers can fake `Content-Type`; we read first 512 bytes via `http.DetectContentType`. |
| **Structured JSON logs** | `level`, `msg`, `time` — ready for Loki/ELK. GORM logger silenced for `record not found`. |
| **Error classification** | `AppError` (4xx/409) vs technical error (500). Client gets actionable message; internals stay hidden. |

---

## 📈 Future Improvements

- [ ] Unit tests (service layer: order status machine, stock logic)
- [ ] Dockerfile multi-stage build
- [ ] CI/CD (GitHub Actions: build, test, lint)
- [ ] Rate limiting on `/login`
- [ ] Request ID middleware + structured access logs
- [ ] Swagger/OpenAPI docs
- [ ] Webhook for payment callback (Midtrans/Xendit)
- [ ] Redis cache for product list

---

## 📄 License

MIT — free to use, modify, distribute.

---

## 👨‍💻 Author

**Muhammad Bayu Yusuf**  
Golang Backend Engineer  
- GitHub: [@mrbayss](https://github.com/mrbayss)  
- LinkedIn: [muhamad-bayu-yusuf](https://linkedin.com/in/muhamad-bayu-yusuf-a8613b214)  
- Email: bayu@warbay.id (fictional)

---

> Built with ☕ and Go — "The best code is the code never written."