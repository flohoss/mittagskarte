#!/bin/sh
set -e

cat /app/coder-mini.txt
echo -n "Version: ${APP_VERSION} Build Time: ${BUILD_TIME}"

exec /app/mittagskarte
