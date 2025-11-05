#!/bin/sh
set -e

cp /app/node_modules/htmx.org/dist/htmx.min.js /app/assets/js/htmx.min.js

cp /app/node_modules/@floating-ui/core/dist/floating-ui.core.umd.min.js /app/assets/js/floating-ui.core.umd.min.js
cp /app/node_modules/@floating-ui/dom/dist/floating-ui.dom.umd.min.js /app/assets/js/floating-ui.dom.umd.min.js

cat /logo.txt

templ generate --watch --proxybind="0.0.0.0" --proxy="http://localhost:8156" --cmd="go run ." --open-browser=false
