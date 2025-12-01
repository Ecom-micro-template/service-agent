# Agent Service

Agent/reseller management microservice for Niaga Platform. Handles agent registration, commission tracking, and payout management.

## Features

- **Agent Management:** CRUD operations for agents/resellers
- **Commission Tracking:** Automatic commission calculation from orders
- **Payout System:** Period-based payout tracking with approval workflow
- **Statistics API:** Agent performance metrics and earnings

## Setup

###Prerequisites

- Go 1.21+
- PostgreSQL 12+

### Installation

1. Copy `.env.example` to `.env`:
```bash
cp .env.example .env
```

2. Configure database in `.env`

3. Install dependencies:
```bash
go mod download
```

4. Run:
```bash
go run cmd/server/main.go
```

## API Endpoints

**Agents:**
- `POST /api/v1/agents` - Create agent
- `GET /api/v1/agents` - List agents
- `GET /api/v1/agents/:id` - Get agent
- `PUT /api/v1/agents/:id` - Update agent
- `DELETE /api/v1/agents/:id` - Delete agent (soft)
- `GET /api/v1/agents/:id/stats` - Agent statistics

**Commissions:**
- `POST /api/v1/commissions` - Create commission
- `GET /api/v1/agents/:id/commissions` - Agent commissions
- `GET /api/v1/commissions/pending` - Pending commissions
- `PUT /api/v1/commissions/:id/approve` - Approve commission

**Payouts:**
- `POST /api/v1/payouts` - Create payout
- `GET /api/v1/agents/:id/payouts` - Agent payouts
- `GET /api/v1/payouts/:id` - Get payout
- `PUT /api/v1/payouts/:id/mark-paid` - Mark as paid

**Health:**
- `GET /health` - Health check
- `GET /ready` - Readiness check

## License

Copyright © 2024 Niaga Platform
