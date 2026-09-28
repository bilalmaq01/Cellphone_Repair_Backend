# Cellphone Repair

A repair shop application with a Go API and a React interface. It supports employee sign-in, repair intake with front and back phone photos, repair status updates, customer status checks with SMS verification, and signed repair authorization records.

## Requirements

- Go 1.26 or later
- Node.js 22 or later and npm
- PostgreSQL

Supabase can provide the PostgreSQL database and private storage for signatures and repair photos. Twilio is optional; without Twilio credentials, the app uses a development SMS logger.

Intake photos are resized in the browser to at most 1280 pixels on the longest edge and encoded as JPEG at 72% quality before upload. Only the resized images are stored.

## Setup

1. Copy `.env.example` to `.env` and fill in the required values. Keep `.env` private.
2. Install frontend dependencies and build the interface:

   ```sh
   cd web
   npm ci
   npm run build
   cd ..
   ```

3. Apply database migrations and start the server:

   ```sh
   go run ./cmd/migrate
   go run ./cmd/server
   ```

The server listens on port `8080` by default and serves the frontend from `web/dist`. `GET /health` provides a health check. Create an initial administrator with:

```sh
go run ./cmd/seedadmin -email admin@example.com -password 'change-this-password'
```

Before using the intake flow with customers, review the sample repair authorization terms in `web/src/pages/Intake.jsx` and update them to match your shop's policy.

## Configuration

The server reads its configuration from environment variables (a local `.env` file can be loaded by your development environment). `DATABASE_URL` and `JWT_SECRET` are required. Supabase Storage and Twilio settings are optional; see `.env.example` for the full list.

## Deployment

The Dockerfile builds the frontend and Go server into a single container that listens on port `8080`. See [DEPLOY.md](DEPLOY.md) for the Amazon ECS Express Mode deployment steps.
