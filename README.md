# Mittagskarte

[https://schniddzl.de](https://schniddzl.de) - Deine Mittagskarte

## Dev

```bash
# Upgrade node packages
docker compose run --rm node yarn upgrade
# Upgrade golang packages
docker compose run --rm backend go get -u && go mod tidy
# Run dev server
docker compose up --build --force-recreate
````
