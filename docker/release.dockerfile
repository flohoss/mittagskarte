ARG V_GOLANG=1.24
ARG V_NODE=lts
ARG V_ALPINE=3
ARG V_PLAYWIGHT=20
FROM alpine:${V_ALPINE} AS logo
WORKDIR /app
RUN apk add figlet
RUN figlet GoDash > logo.txt

FROM node:${V_NODE}-slim AS golang-builder
WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    gnupg libc6-dev libnss3-dev libnet-dev build-essential \
    libmagickwand-dev libmagickcore-dev imagemagick libmupdf-dev \
    apt-transport-https ca-certificates && \
    apt-get autoremove -y && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=golang /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:/root/go/bin:${PATH}"

RUN go install github.com/a-h/templ/cmd/templ@latest

COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

COPY . .
RUN templ generate
RUN go build -ldflags="-s -w" -o mittag main.go

FROM node:${V_NODE}-alpine AS node-builder
WORKDIR /app

COPY package.json yarn.lock ./
RUN yarn install --frozen-lockfile --network-timeout 30000

COPY ./views/ ./views/
COPY ./assets/ ./assets/
RUN yarn run tw:build

FROM node:${V_NODE}-slim AS final
WORKDIR /app

RUN npx -y playwright@$V_PLAYWIGHT install-deps chromium

RUN apt-get update && apt-get install -y --no-install-recommends \
    gnupg libc6-dev libnss3-dev libnet-dev build-essential \
    libmagickwand-dev libmagickcore-dev imagemagick libmupdf-dev \
    apt-transport-https ca-certificates curl dumb-init && \
    apt-get autoremove -y && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=logo /app/logo.txt .
COPY --from=node-builder /app/assets/favicon/ ./assets/favicon/
COPY --from=node-builder /app/assets/js/ ./assets/js/
COPY --from=node-builder /app/assets/css/style.css ./assets/css/style.css
COPY --from=golang-builder /app/mittag .

EXPOSE 8156

ARG APP_VERSION
ENV APP_VERSION=$APP_VERSION

RUN chown -R 1000:1000 /app

ENTRYPOINT ["dumb-init", "--"]
USER node
CMD ["/app/mittag"]
