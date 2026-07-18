# Backend CI/CD

## What Each Workflow Does
- `backend-ci.yml`: your safety net. It checks formatting, static analysis, and tests before deployment.
- `backend-staging-deploy.yml`: automatically promotes `main` to staging after CI passes, runs DB migration, triggers Render, then smoke-tests `/healthz`.
- `backend-prod-deploy.yml`: manual production release. It runs the production migration, pushes a container image to ECR, triggers App Runner, then smoke-tests production.

## Runtime Contract
- `go run . serve`: starts the API server.
- `go run . migrate`: runs GORM `AutoMigrate` without starting the server.
- `GET /healthz`: verifies API readiness and checks DB/Redis status.

## GitHub Secrets To Create
- `STAGING_DATABASE_URL`
- `RENDER_STAGING_DEPLOY_HOOK_URL`
- `STAGING_HEALTHCHECK_URL`
- `PRODUCTION_DATABASE_URL`
- `PRODUCTION_HEALTHCHECK_URL`
- `AWS_ROLE_TO_ASSUME`
- `AWS_REGION`
- `ECR_REPOSITORY`
- `APP_RUNNER_SERVICE_ARN`

## Render Staging Setup
- Create a Render web service from this repo using the included `Dockerfile`.
- Set the service start command to the image default so it runs `serve`.
- Configure staging env vars in Render: `APP_ENV=staging`, `PORT`, `DATABASE_URL`, `REDIS_*`, `COOKIE_*`, `CORS_ALLOWED_ORIGINS`, `JWT_SECRET`, `GOOGLE_*`, `RAZORPAY_*`.
- Create a deploy hook in Render and store it as `RENDER_STAGING_DEPLOY_HOOK_URL`.

## Environment-Specific YAML Config
- Non-secret defaults live in `config/{development,staging,production}.yaml`.
- On startup, `config/runtime.go` loads `config/{APP_ENV}.yaml` and uses it as defaults.
- Environment variables **always override** YAML values.
- Secrets (`DATABASE_URL`, `JWT_SECRET`, etc.) are never in YAML files — set them via env vars only.
- The YAML files are checked into git and copied into the Docker image at build time.

## AWS Production Setup
- Create an ECR repository for this backend image.
- Configure an App Runner service to pull the `production` tag from that ECR repository.
- Enable GitHub OIDC access to AWS and store the assumable role ARN as `AWS_ROLE_TO_ASSUME`.
- Set the same runtime env vars in App Runner that you use in staging, but with production values.

## Neon And Upstash
- Use separate Neon databases for staging and production.
- Point staging/prod `DATABASE_URL` secrets at the correct Neon connection strings.
- Use Upstash Redis connection details for `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB`, and `REDIS_USE_TLS=true`.

## Rollback Path
- Render: redeploy the previous successful build from the Render dashboard.
- App Runner: retag or redeploy the previous ECR image, then run `start-deployment` again.
- Neon: restore from backup or restore point before rerunning a deployment if a schema change needs to be undone.
