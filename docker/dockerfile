ARG V_GOLANG=1.25
ARG V_NODE=lts
ARG V_ALPINE=3
ARG V_PLAYWRIGHT=20
FROM alpine:${V_ALPINE} AS logo
WORKDIR /app
RUN apk add figlet > /dev/null 2>&1
RUN figlet Mittagskarte > logo.txt

FROM node:${V_NODE}-slim AS golang-builder
WORKDIR /app

RUN apt-get update > /dev/null 2>&1 && apt-get install -y --no-install-recommends \
    gnupg libc6-dev libnss3-dev libnet-dev build-essential \
    libmagickwand-dev libmagickcore-dev imagemagick libmupdf-dev \
    apt-transport-https ca-certificates > /dev/null 2>&1 && \
    apt-get autoremove -y > /dev/null 2>&1 && \
    apt-get clean > /dev/null 2>&1 && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=golang /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:/root/go/bin:${PATH}"

RUN go install github.com/a-h/templ/cmd/templ@latest > /dev/null 2>&1

COPY ./go.mod .
COPY ./go.sum .
RUN go mod download > /dev/null 2>&1

COPY . .
RUN templ generate
RUN go build -ldflags="-s -w" -o mittagskarte main.go

FROM node:${V_NODE}-alpine AS node-builder
WORKDIR /app

COPY package.json yarn.lock ./
RUN yarn install --frozen-lockfile --network-timeout 30000 --silent

COPY ./views/ ./views/
COPY ./assets/ ./assets/
RUN yarn run tw:build

FROM node:${V_NODE}-slim AS final
WORKDIR /app

RUN npx -y playwright@$V_PLAYWRIGHT install-deps chromium > /dev/null 2>&1

RUN apt-get update > /dev/null 2>&1 && apt-get install -y --no-install-recommends \
    libnss3 libnet1 dumb-init ca-certificates curl \
    libmagickwand-6.q16-6 imagemagick libmupdf-dev > /dev/null 2>&1 \
    && apt-get clean > /dev/null 2>&1 \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=logo /app/logo.txt .
COPY --from=node-builder /app/assets/favicon/ ./assets/favicon/
COPY --from=node-builder /app/assets/js/ ./assets/js/
COPY --from=node-builder /app/assets/css/style.css ./assets/css/style.css
COPY --from=golang-builder /app/mittagskarte .
COPY ./docker/entrypoint.sh .

EXPOSE 8156

ARG APP_VERSION
ENV APP_VERSION=$APP_VERSION
ARG BUILD_TIME
ENV BUILD_TIME=$BUILD_TIME
ARG REPO
ENV REPO=$REPO

RUN chown -R 1000:1000 /app

USER node
ENTRYPOINT ["dumb-init", "--", "/app/entrypoint.sh"]
