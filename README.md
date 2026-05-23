# KBTUspace

KBTUspace is a university community app for KBTU students, organizers, and admins.

The project contains:

- `kbtuspace-backend`: Go/Gin API, PostgreSQL, Redis cache, migrations.
- `kbtuspace-frontend`: React/Vite frontend.

## Features

- Student registration and login with JWT auth.
- KBTU email domain validation (`@kbtu.kz`).
- Faculty and global feeds.
- Organizer/admin post creation.
- Organizer/admin event creation.
- Student event registration and cancellation.
- Global content moderation by admins.
- Reports workflow for posts/events.
- Admin user management.

## Requirements

- Docker Desktop for the backend stack.
- Node.js 20+ and npm for the frontend.

## Start With Docker Compose

From the repository root:

```powershell
Copy-Item .env.example .env
docker compose up --build
```

This starts PostgreSQL, Redis, backend API, and frontend.

URLs:

- Frontend: `http://localhost:5173`
- API health check: `http://localhost:8080/ping`
- Swagger: `http://localhost:8080/swagger/index.html`
- API base URL: `http://localhost:8080/api/v1`

Before the first run, edit `.env` and set real local values. At minimum, change:

- `DB_PASSWORD`
- `JWT_SECRET`
- `DEFAULT_ADMIN_PASSWORD`

The first admin is created only when `DEFAULT_ADMIN_PASSWORD` is set. The password is never logged.

There are three local env locations by design:

- `.env` in the repository root is used by Docker Compose.
- `kbtuspace-backend/.env` is only for running the Go backend directly from `kbtuspace-backend`.
- `kbtuspace-frontend/.env` is only for running Vite directly from `kbtuspace-frontend`.

Any real `.env` file is local-only and must not be committed. Use `.env.example` files as templates.

## Start Frontend Without Docker

```powershell
cd kbtuspace-frontend
npm.cmd install
npm.cmd run dev
```

Open:

```text
http://localhost:5173
```

The frontend API URL is configured with:

```env
VITE_API_URL=http://localhost:8080/api/v1
```

See `kbtuspace-frontend/.env.example`.

## Backend Environment

Important backend environment variables:

- `PORT`
- `DB_HOST`
- `DB_PORT`
- `DB_USER`
- `DB_PASSWORD`
- `DB_NAME`
- `DB_SSLMODE`
- `REDIS_URL`
- `JWT_SECRET`
- `ENVIRONMENT`
- `CORS_ALLOWED_ORIGINS`
- `DEFAULT_ADMIN_EMAIL`
- `DEFAULT_ADMIN_PASSWORD`

Use `.env.example` as a template. Do not commit `.env`; it is ignored by git.

`CORS_ALLOWED_ORIGINS` is a comma-separated list, for example:

```env
CORS_ALLOWED_ORIGINS=http://localhost:5173,https://your-domain.kz
```

## Checks

Backend:

```powershell
cd kbtuspace-backend
go test ./...
```

Frontend:

```powershell
cd kbtuspace-frontend
npm.cmd run lint
npm.cmd run build
```

On Windows, use `npm.cmd` if PowerShell blocks `npm.ps1`.

End-to-end tests:

```powershell
docker compose up --build
cd kbtuspace-frontend
npm.cmd install
npx.cmd playwright install
$env:E2E_ADMIN_PASSWORD = "<same value as DEFAULT_ADMIN_PASSWORD>"
npm.cmd run test:e2e
```

Optional e2e variables:

- `E2E_BASE_URL` defaults to a Vite dev server started by Playwright.
- `E2E_API_URL` defaults to `http://127.0.0.1:8080/api/v1`.
- `E2E_ADMIN_EMAIL` defaults to `admin@kbtu.kz`.

## Notes

Reports record moderation decisions. Rejecting a report keeps the content, while closing a report deletes the reported post or event.
