ARG V_GOLANG=1.23
ARG V_NODE=lts
ARG V_ALPINE=3
FROM golang:${V_GOLANG} AS golang
RUN apt-get update -qq && apt-get install -y --no-install-recommends \
    gnupg libc6-dev libnss3-dev libnet-dev build-essential \
    libmagickwand-dev libmagickcore-dev imagemagick libmupdf-dev \
    > /dev/null && \
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

FROM node:${V_NODE}-slim AS final
RUN npx -y playwright@$V_PLAYWIGHT install-deps chromium
RUN apt-get update -qq && apt-get install -y --no-install-recommends \
    gnupg libc6-dev libnss3-dev libnet-dev build-essential \
    libmagickwand-dev libmagickcore-dev imagemagick libmupdf-dev \
    apt-transport-https ca-certificates dumb-init \
    > /dev/null && \
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

EXPOSE 4000

RUN chown -R 1000:1000 /app

ENTRYPOINT ["dumb-init", "--"]
USER node
CMD ["/app/mittag"]
