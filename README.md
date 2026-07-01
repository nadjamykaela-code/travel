# Travel Bot — Busca de Passagens (GCP Always Free)

Sistema automático de busca de passagens aéreas rodando 100% no **GCP Free Tier**. Usuários criam filtros de busca pela web, e um worker agendado consulta a Skyscanner API, notificando por e-mail (SendGrid) e push (FCM) quando encontra voos dentro dos critérios.

## Arquitetura

```
                 ┌──────────────┐
                 │  Web (React) │
                 │  nginx:80    │
                 └──────┬───────┘
                        │ /api/
                 ┌──────▼───────┐     ┌──────────────┐
                 │  API (Gin)   │────▶│  Firestore   │
                 │  :8080       │     │  (NoSQL)     │
                 └──────┬───────┘     └──────▲────────┘
                        │                    │
                 ┌──────▼───────┐     ┌──────┴────────┐
                 │  Worker      │────▶│  Skyscanner   │
                 │  Cloud Run   │     │  API          │
                 │  Scheduler   │     │  Live Pricing │
                 │  15 min cron │     │  + Indicative │
                 └──────┬───────┘     │  (failover)   │
                        │             └───────────────┘
                   ┌────┴────┐
                   ▼         ▼
             ┌──────────┐ ┌────────┐
             │ SendGrid │ │  FCM   │
             │ (E-mail) │ │ (Push) │
             └──────────┘ └────────┘
```

## Estrutura do Projeto

```
├── apps/
│   ├── api/              # API HTTP (Gin, Firestore, Firebase Auth)
│   │   └── cmd/main.go   # Entrypoint (+ CORS middleware)
│   ├── web/              # Frontend React + Vite + Tailwind (bun)
│   │   └── src/          # Login (Firebase Auth) + Dashboard (CRUD)
│   └── worker/           # Worker agendado (HTTP + ticker)
├── pkg/
│   ├── clients/          # Skyscanner API client
│   ├── filters/          # Motor de filtragem (7 critérios)
│   ├── models/           # Filter + Itinerary (Validate, StorageEstimate)
│   └── notifications/    # SendGrid + FCM interfaces
├── internal/
│   ├── config/           # Config (env vars, Validate)
│   └── firestore/        # Client struct (não global, DI)
├── infra/
│   ├── docker/           # Dockerfiles + nginx.conf + compose
│   ├── terraform/        # Cloud Run + Scheduler + IAM
│   └── firebase/         # Firestore security rules com validação
└── .github/workflows/    # CI (test + lint + build) + Deploy (WIF)
```

## Stack

| Camada       | Tecnologia                                              |
|-------------|--------------------------------------------------------|
| Backend      | Go 1.25+ (Gin, Firestore SDK)                          |
| Frontend     | React 18 + TypeScript + Vite 5 + Tailwind CSS 3        |
| Worker       | Go (HTTP server — Cloud Scheduler trigger, Live Pricing + Indicative failover) |
| Database     | Firestore (Native Mode)                                |
| APIs externas| Skyscanner, SendGrid, FCM                              |
| Auth         | Firebase Auth SDK (browser) + Bearer token (API)       |
| Infra        | Cloud Run + Cloud Scheduler + Terraform                |
| Testes       | Go (`testing -race`), Vitest + React Testing Library   |

## Rotas da API

| Método | Rota                    | Serviço | Auth    | Descrição                         |
|-------------------------|---------|---------|-----------------------------------|
| GET    `/health`        | API     | ❌      | Health check                      |
| GET    `/metrics`       | API     | ❌      | Prometheus metrics                |
| POST   `/api/filters`  | API     | ✅       | Criar filtro                      |
| GET    `/api/filters`   | API     | ✅       | Listar filtros do usuário         |
| PUT    `/api/filters/:id`| API   | ✅       | Atualizar filtro                  |
| DELETE `/api/filters/:id`| API  | ✅       | Soft delete                       |
| GET    `/api/auth/verify`| API   | ✅       | Verificar token Firebase          |
| GET    `/health`        | Worker  | ❌      | Health check                      |
| POST   `/run`           | Worker  | ❌      | Executa ciclo de busca (Scheduler)|
| GET    `/metrics`       | Worker  | ❌      | Prometheus metrics                |

## Como rodar localmente

### Docker (recomendado)

```bash
cd infra/docker
cp .env.example .env      # preencha as chaves
docker compose up --build
```

### Manual

```bash
# API
go run ./apps/api/cmd/main.go

# Worker
go run ./apps/worker/cmd/main.go

# Frontend (requer bun)
cd apps/web
cp .env.example .env
bun install
bun run dev
```

## Testes

```bash
# Go — todos os pacotes
go test -race ./...

# End-to-end (requer docker-compose rodando)
E2E_TEST=1 go test -race -count=1 ./test/e2e/

# Frontend (bun)
cd apps/web && bun run test
```

## Deploy

```bash
cd infra/terraform
terraform init
terraform workspace select dev
terraform apply
```

Cloud Run (api + worker) com Cloud Scheduler disparando o worker a cada 15 min — tudo dentro do GCP Always Free.

## CI/CD — GitHub Actions

### Secrets & Variables

The deploy workflow (`deploy.yml`) uses the following **GitHub Actions secrets** and **variables**:

| Name | Type | Description |
|------|------|-------------|
| `GCP_PROJECT_ID` | variable | GCP project ID (e.g. `myka-travel`) |
| `WIF_PROVIDER` | variable | Workload Identity Federation provider resource name |
| `DEPLOY_SA` | variable | Deploy service account email for OIDC auth |
| `SKYSCANNER_API_KEY` | secret | Skyscanner API key for flight search |
| `SENDGRID_API_KEY` | secret | SendGrid API key for email notifications |
| `FCM_CREDENTIALS` | secret | Firebase Admin SDK service account JSON (base64) |

### Workload Identity Federation Setup

The deploy job authenticates to GCP via WIF (no static service account keys in CI):

```yaml
- id: auth
  uses: google-github-actions/auth@v2
  with:
    workload_identity_provider: ${{ vars.WIF_PROVIDER }}
    service_account: ${{ vars.DEPLOY_SA }}
```

Requirements:
1. Create a Workload Identity Pool and Provider in GCP (IAM & Admin)
2. Grant the deploy SA the roles `run.admin`, `storage.admin`, `iam.serviceAccountUser`
3. Set `WIF_PROVIDER` and `DEPLOY_SA` as **GitHub variables** (not secrets)

### Local environment variables

Copy `.env.example` in `infra/docker/` and fill in:

```bash
cp infra/docker/.env.example infra/docker/.env
```

Required vars:
- `SKYSCANNER_API_KEY` — from [Skyscanner Developer Portal](https://developers.skyscanner.net)
- `SENDGRID_API_KEY` — from [SendGrid](https://sendgrid.com)
- `VITE_FIREBASE_API_KEY`, `VITE_FIREBASE_AUTH_DOMAIN`, `VITE_FIREBASE_PROJECT_ID` — from [Firebase Console](https://console.firebase.google.com)

## GCP Always Free — Limites

| Recurso        | Limite grátis                          |
|---------------|----------------------------------------|
| Cloud Run     | 2M requisições/mês, 360K vCPU-min/dia |
| Cloud Scheduler| 3 jobs gratuitos                       |
| Firestore     | 1GB armazenamento, 50K leituras/dia   |

## Licença

MIT
