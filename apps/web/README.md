# Travel Bot — Web

Frontend React SPA com autenticação Firebase e dashboard de filtros de viagem.

## Stack

- React 18 + TypeScript + Vite 5
- Tailwind CSS 3
- Firebase Auth SDK (`signInWithEmailAndPassword`)
- Axios (API client com Bearer token)
- Lucide React (ícones)
- Vitest + React Testing Library (testes)

## Pré-requisitos

- **bun** (necessário — npm não suportado)

## Variáveis de ambiente

Copie `.env.example` para `.env` e preencha:

```
VITE_FIREBASE_API_KEY=xxx
VITE_FIREBASE_AUTH_DOMAIN=xxx
VITE_FIREBASE_PROJECT_ID=xxx
```

## Scripts

| Comando | Descrição |
|---------|-----------|
| `bun run dev` | Dev server (Vite, porta 5173, proxy `/api` → `:8080`) |
| `bun run build` | `tsc -b && vite build` |
| `bun run preview` | Preview do build |
| `bun run test` | Vitest run |
| `bun run test:watch` | Vitest watch |

## Testes

```
src/
  components/
    FilterCard.test.tsx      — 8 testes (render, badges, callbacks)
    ProtectedRoute.test.tsx  — 3 testes (loading, redirect, auth)
    Spinner.test.tsx         — 3 testes (default/custom size, acessibilidade)
  pages/
    Login.test.tsx           — 2 testes (render, botão submit)
  services/
    api.test.ts              — 2 testes (instância axios, interceptors)
```

## Login

Usa `signInWithEmailAndPassword` do Firebase Auth SDK. O token JWT é armazenado em `localStorage` e enviado como `Authorization: Bearer` em toda requisição ao backend via interceptor Axios.

## Rotas

| Path | Componente | Descrição |
|------|-----------|-----------|
| `/login` | Login | Tela de login |
| `/dashboard` | Dashboard (protegida) | CRUD de filtros |
| `*` | Redirect → `/dashboard` | Fallback |
