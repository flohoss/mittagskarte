# Mittag

[![pipeline status](https://gitlab.unjx.de/flohoss/mittag/badges/main/pipeline.svg)](https://gitlab.unjx.de/flohoss/mittag/-/commits/main)

## Run development server

```sh
./scripts/dev.sh
docker compose run --rm yarn install --frozen-lockfile
docker compose --profile dev up
```

## Build mittag image

```sh
./scripts/dev.sh
docker compose --profile build build
```
