ARG V_GOLANG
ARG V_DEBIAN
ARG V_NODE
ARG V_ALPINE
FROM golang:${V_GOLANG} AS goBuilder
RUN apt-get update -qq
# Image to webp convertion
RUN apt-get install -y libwebp-dev

WORKDIR /app
COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w"

FROM node:${V_NODE}-alpine AS nodeBuilder
WORKDIR /web

COPY ./web .
RUN yarn install --frozen-lockfile
RUN yarn run build

FROM alpine:${V_ALPINE} AS logo
RUN apk add figlet
RUN figlet Mittag > logo.txt

FROM debian:${V_DEBIAN}-slim AS final
RUN apt-get update -qq
# Image to webp convertion
RUN apt-get install -y libwebp-dev ca-certificates

# Mittag
WORKDIR /app
COPY scripts/entrypoint.sh .

COPY --from=logo /logo.txt .
COPY --from=nodeBuilder /web/dist/spa/ ./web/
COPY --from=goBuilder /app/mittag .
COPY ./internal/config/restaurants ./internal/config/restaurants
COPY ./internal/config/thumbnails ./internal/config/thumbnails

ENTRYPOINT ["/app/entrypoint.sh"]
