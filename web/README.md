# Bitrok Web Dashboard

Next.js 16 dashboard for account authentication, immutable username claims, CLI token issuance, and relay activity views.

## Required environment

```dotenv
DATABASE_URL=postgresql://user:password@host:5432/bitrok
BETTER_AUTH_SECRET=<at-least-32-random-bytes>
BETTER_AUTH_URL=http://localhost:3000
NEXT_PUBLIC_APP_URL=http://localhost:3000
GITHUB_CLIENT_ID=<github-oauth-client-id>
GITHUB_CLIENT_SECRET=<github-oauth-client-secret>
BITROK_JWT_SECRET=<same-secret-as-the-go-relay>
BITROK_SERVER_URL=http://localhost:8080
```

Optional variables are documented in the repository `.env.example`. Use the same `BITROK_REDIS_URL` on the web and relay deployments when Redis-backed opaque CLI tokens are enabled.

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
