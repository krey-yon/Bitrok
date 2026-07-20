# Bitrok Web Dashboard

Next.js 16 dashboard for account authentication, immutable username claims, CLI token issuance, and relay activity views.

## Environment

Copy [`web/.env.example`](.env.example) to `.env.local` for local development. It contains the required PostgreSQL, Better Auth, GitHub OAuth, relay JWT, and dashboard URL settings.

Use the same `BITROK_JWT_SECRET` on the web and relay deployments. Configure the same `BITROK_REDIS_URL` on both deployments when Redis-backed opaque CLI tokens or distributed rate limiting are enabled.

## Commands

```bash
npm ci
npm run db:generate
npm run db:deploy
npm run dev
```

Before release:

```bash
npm audit --omit=dev
npm run lint
npx tsc --noEmit
npm run build
```

Production deploys must run `npm run db:deploy` before serving a build that depends on a new Prisma migration. Do not use `prisma migrate dev` in production.
