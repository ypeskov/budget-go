# Your Finance Tracker
An API for tracking personal finances.
Main features include:
- Multi-user support
- Expense and income tracking
- Budget management (monthly, yearly, etc.)
- Category management (e.g., food, rent, entertainment)
- Reporting and analytics
- Multi-currency support (for accounts, transactions, budgets, etc.)

Client (VueJS 3) can be found [here](https://github.com/ypeskov/budget-tracker/tree/master/src-front)

Kubernetes manifests can be found [here](https://github.com/ypeskov/k8s-orgfin)

Previous Python/FastAPI version can be found [here](https://github.com/ypeskov/budget-tracker/tree/master/back-fastapi)


## Tech Stack
- Go/Echo
- PostgreSQL
- Docker/Kubernetes (optional)
- JWT for authentication
- Goose for database migrations

## Database migrations

We use [goose](https://github.com/pressly/goose) for migrations.

### Install
```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

## Build and Push Docker Image
The next command builds and pushes a Docker image to your Docker registry.
It build all three images: API, worker, and scheduler (Asynq)
```bash
./build-and-push.sh push VersionTag
```