# Deploying

The two sides deploy independently and in different ways. Merging to `main` ships the web app and does **not** ship the API.

## API (Railway)

Service `fearless-adaptation` in project `ntx-production`, built from `apps/api/Dockerfile` with the repo root as context. The service is not linked to GitHub: every deploy is an upload from a checkout.

```sh
railway login                       # once per machine
railway link                        # once per checkout: project ntx-production
railway up --service fearless-adaptation --detach
```

Migrations run on start (`AutoMigrate`), so a schema change needs nothing beyond the deploy. Variables set on the service: `CORS_ORIGINS`, `NTX_DB_PATH` (also the Dockerfile default), `PORT`. `NEPSE_TLS_VERIFY=true` turns certificate checks on for go-nepse; unset, they are off.

### Railway configuration

Service and volume settings are infrastructure as code in `.railway/railway.ts` (IaC partial `fearless-adaptation`). Edit the file, then:

```sh
railway config plan      # preview what Railway would change
railway config apply     # apply it
```

**Apply before pushing a commit that changes `.railway/`.** The file is the record of what is live; a pushed change nobody applied drifts from production and the next `plan` surprises whoever runs it.

### Database

Production SQLite lives on the Railway volume `fearless-adaptation-volume`, mounted at `/data` (`NTX_DB_PATH=/data/market.db`). The worker runs Monday to Friday only: prices, market day, companies and fundamentals at 15:05 NPT, ownership and dividends at 10:00. Nothing fills an empty database. After a deploy onto an empty volume, run the backfill inside the service, where the volume is mounted:

```sh
railway ssh --service fearless-adaptation -- /bin/ntx backfill
```

Flags: `-companies -fundamentals -prices -ownership -corporate-actions`; no flag runs all five. See `docs/backfill.md` for what each pulls.

## Web (Cloudflare Workers)

Worker `ntx` (custom domain ntx.anishshrestha.com, attached in the Cloudflare dashboard, not in `wrangler.jsonc`), built with `@sveltejs/adapter-cloudflare` from `apps/web/wrangler.jsonc`. Cloudflare Workers Builds deploys automatically from `main`; the build image reads the `packageManager` field in `apps/web/package.json` for the pnpm version, so keep that field pinned.

To ship by hand, from `apps/web`:

```sh
wrangler login            # once per machine
pnpm build && pnpm exec wrangler deploy
```

`API_URL` is a Worker var in `apps/web/wrangler.jsonc`, read through `$app/env/private`. Changing the Railway domain means changing it there and redeploying the web.

## After a merge that touches both sides

Deploy the API first, then let Workers Builds ship the web. On 2026-11-04 the web went live calling RPCs the old API binary did not have.
