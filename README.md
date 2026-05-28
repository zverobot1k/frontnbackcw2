# Backend21 E-commerce Project

Full-stack e-commerce demo with Go + React.

## Implemented requirements

- JWT authentication + RBAC with roles: `customer` and `admin`
- Product catalog with search and filtering
- Shopping cart with `localStorage` persistence + server synchronization
- Checkout flow with Stripe integration for payments
- Inventory management (stock decrement on successful checkout)
- User order history
- Admin panel for product management

## Tech stack

- Backend: Go, net/http, GORM, PostgreSQL, Redis, JWT
- Frontend: React + Vite
- Payments: Stripe Payment Intents API
- Containers: Docker Compose

## Project structure

- `cmd/main.go` - app bootstrap
- `internal/models` - domain models (`User`, `Product`, `CartItem`, `Order`, `OrderItem`)
- `internal/service` - business logic
- `internal/repository` - DB access layer
- `internal/transport` - HTTP handlers and DTOs
- `frontend` - React app

## Roles and access

- `customer`
  - browse product catalog
  - manage cart
  - checkout and view personal order history
- `admin`
  - all customer permissions
  - create/update/delete products
  - view all orders
  - manage users via admin user endpoints

Note: In a fresh database, the first registered user is promoted to `admin` automatically.

## Environment variables

Backend:

- `APP_PORT` (default: `8080`)
- `DB_HOST` (default: `localhost`)
- `DB_PORT` (default: `5432`)
- `DB_USER` (default: `postgres`)
- `DB_PASSWORD` (default: `postgres`)
- `DB_NAME` (default: `webdb`)
- `DB_SSLMODE` (default: `disable`)
- `JWT_SECRET` (default: `dev_secret`)
- `REDIS_ADDR` (default: `localhost:6379`)
- `STRIPE_SECRET_KEY` (optional, empty by default)
- `STRIPE_WEBHOOK_SECRET` (optional, required for signed webhook processing)

Frontend:

- `VITE_PROXY_TARGET` (default in compose: `http://api:8080`)

## Run with Docker (recommended)

From project root:

```bash
docker compose up --build
```

Services:

- API: `http://localhost:8080`
- Frontend: `http://localhost:5173`
- Swagger: `http://localhost:8080/swagger/`

## Run locally (without Docker)

### 1) Start PostgreSQL and Redis

Make sure PostgreSQL and Redis are running locally.

Example PostgreSQL values expected by default:

- host: `localhost`
- port: `5432`
- db: `webdb`
- user: `postgres`
- password: `postgres`

### 2) Start backend

```bash
go mod tidy
go run ./cmd
```

### 3) Start frontend

In another terminal:

```bash
cd frontend
npm install
npm run dev
```

Open `http://localhost:5173`.

## Stripe notes

- If `STRIPE_SECRET_KEY` is set, checkout creates a Stripe PaymentIntent and returns `stripe_client_secret`.
- If `STRIPE_SECRET_KEY` is empty, checkout still creates the order in local system, but without Stripe payment intent.
- Webhook endpoint: `POST /api/webhooks/stripe`.
- Webhook listens for `payment_intent.succeeded` and moves order status to `paid` by matching `payment_ref`.

Use Stripe test keys for development.

## How to test Stripe webhook locally

1. Add Stripe keys and restart API:

```bash
export STRIPE_SECRET_KEY=sk_test_xxx
export STRIPE_WEBHOOK_SECRET=whsec_xxx
go run ./cmd
```

2. In another terminal, start Stripe CLI forwarding to local webhook:

```bash
stripe listen --forward-to localhost:8080/api/webhooks/stripe
```

Copy the printed signing secret (`whsec_...`) and set it to `STRIPE_WEBHOOK_SECRET`.

3. Create a test order from app/API via checkout endpoint (`POST /api/orders/checkout`) so order gets `payment_ref` equal to Stripe PaymentIntent ID.

4. Confirm the exact PaymentIntent from checkout response (`payment_reference`) so webhook hits the related order:

```bash
stripe payment_intents confirm pi_xxx --payment-method pm_card_visa
```

This sends `payment_intent.succeeded` for the same `pi_xxx` value saved as order `payment_ref`.

5. (Optional) Trigger generic test event from Stripe CLI:

```bash
stripe trigger payment_intent.succeeded
```

6. Verify result:

- check API response for your user orders: `GET /api/orders`
- confirm order status changed from `created` to `paid`

Tip: generic `stripe trigger` may produce a new test PaymentIntent unrelated to your local order, so deterministic validation should use `stripe payment_intents confirm pi_xxx ...`.

## Main API endpoints

Auth:

- `POST /api/auth/register`
- `POST /api/auth/login`
- `POST /api/auth/refresh`
- `GET /api/auth/me`

Products:

- `GET /api/products?search=&category=&min_price=&max_price=&in_stock=true`
- `GET /api/products/{id}`
- `POST /api/products` (admin)
- `PUT /api/products/{id}` (admin)
- `DELETE /api/products/{id}` (admin)

Cart and orders:

- `GET /api/cart`
- `PUT /api/cart`
- `POST /api/orders/checkout`
- `GET /api/orders`
- `GET /api/admin/orders` (admin)
- `POST /api/webhooks/stripe`

Users (admin):

- `GET /api/users`
- `GET /api/users/{id}`
- `PUT /api/users/{id}`
- `DELETE /api/users/{id}`

## Quick manual test flow

1. Open frontend and register first account (becomes admin).
2. Create several products in Admin panel.
3. Logout and register/login as customer account.
4. Filter/search products in catalog.
5. Add products to cart and confirm cart sync.
6. Checkout order.
7. Verify stock was decremented.
8. Open order history.

## Build checks

Validated successfully:

- `go build ./...`
- `npm run build` (in `frontend`)
