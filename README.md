# 👥 Service Agent - Desa Murni Batik

Perkhidmatan portal ejen untuk platform **Desa Murni Batik**.

## 🚀 Ciri-ciri

- 👤 **Agent Portal** - Dashboard ejen
- 🛒 **Orders** - Buat pesanan untuk pelanggan
- 👥 **Customers** - Urus pelanggan sendiri
- 💰 **Commissions** - Kiraan komisen automatik
- 📊 **Performance** - Laporan prestasi

## 🛠️ Tech Stack

- Go 1.21+
- Gin Framework
- GORM
- PostgreSQL
- JWT Auth

## 📦 Setup

```bash
go mod download
go run cmd/server/main.go
```

Server: http://localhost:8080

## 🔗 Agent Portal Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/agent/profile` | Agent profile |
| GET | `/api/v1/agent/dashboard` | Dashboard stats |
| GET | `/api/v1/agent/orders` | Agent's orders |
| POST | `/api/v1/agent/orders` | Create order |
| GET | `/api/v1/agent/customers` | Agent's customers |
| GET | `/api/v1/agent/commissions` | Commissions |

---

**© 2024 Desa Murni Batik** | [KilangDesaMurniBatik](https://github.com/KilangDesaMurniBatik)
