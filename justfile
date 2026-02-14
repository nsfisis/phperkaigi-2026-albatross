docker_compose := "docker compose -f compose.local.yaml"

default: down build up

build:
    {{ docker_compose }} build
    npm install

up:
    {{ docker_compose }} up -d
    npm -w frontend run dev

down:
    {{ docker_compose }} down --remove-orphans

logs:
    {{ docker_compose }} logs

logsf:
    {{ docker_compose }} logs -f

psql:
    {{ docker_compose }} up --wait db
    {{ docker_compose }} exec db psql --user=postgres albatross

psql-query:
    {{ docker_compose }} up --wait db
    {{ docker_compose }} exec --no-TTY db psql --user=postgres albatross

sqldef-dryrun: down
    {{ docker_compose }} build db
    {{ docker_compose }} up --wait db
    {{ docker_compose }} run --no-TTY tools psqldef --dry-run < ./backend/schema.sql

sqldef: down
    {{ docker_compose }} build db
    {{ docker_compose }} up --wait db
    {{ docker_compose }} run --no-TTY tools psqldef < ./backend/schema.sql

asynq:
    {{ docker_compose }} up --wait task-db
    {{ docker_compose }} run tools go run github.com/hibiken/asynq/tools/asynq --uri task-db:6379 dash

init: build initdb

initdb:
    just psql-query < ./backend/schema.sql
    just psql-query < ./backend/fixtures/dev.sql

gen:
    npm -w typespec run build
    cd backend; just gen
    npm -w frontend run openapi-typescript

check:
    cd backend; just check
    cd worker/swift; just check
    npm -w frontend run check
