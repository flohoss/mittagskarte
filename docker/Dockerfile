ARG V_GOLANG=1.23
ARG V_NODE=20
ARG V_ALPINE=3
FROM golang:${V_GOLANG} AS golang
WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    apt-transport-https ca-certificates curl gnupg libmagickwand-dev libmagickcore-dev imagemagick libmupdf-dev && \
    apt-get autoremove -y && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

WORKDIR /app
COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o mittag main.go

FROM node:${V_NODE}-alpine AS node
WORKDIR /web

COPY ./web .
RUN yarn install --frozen-lockfile --network-timeout 30000
RUN yarn run build

FROM alpine:${V_ALPINE} AS logo
WORKDIR /app
RUN apk add figlet
RUN figlet Mittag > logo.txt

FROM mcr.microsoft.com/playwright:v1.47.2-noble AS final

RUN apt-get update && apt-get install -y --no-install-recommends \
    dumb-init \
    apt-transport-https ca-certificates curl gnupg libmagickwand-dev libmagickcore-dev imagemagick libmupdf-dev && \
    apt-get autoremove -y && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

WORKDIR /app

COPY --from=logo /app/logo.txt .
COPY --from=node /web/dist/spa/ ./web/
COPY --from=golang /app/mittag .
COPY ./data/ ./data/

ARG APP_VERSION
ENV APP_VERSION=$APP_VERSION

RUN chown -R ubuntu:ubuntu /app

ENTRYPOINT ["dumb-init", "--"]
USER ubuntu
CMD ["/app/godash"]
