#!/bin/sh

go build -gcflags="all=-N -l" -o tmp/mittag -tags ocr cmd/mittag/mittag.go
exec /go/bin/dlv --listen=:4001 --headless=true --api-version=2 --accept-multiclient exec tmp/mittag
