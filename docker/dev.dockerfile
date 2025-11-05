ARG V_GOLANG=1.25
ARG V_NODE=lts
ARG V_ALPINE=3
ARG V_PLAYWIGHT=20
FROM alpine:${V_ALPINE} AS logo
WORKDIR /app
RUN apk add figlet > /dev/null 2>&1
RUN figlet Mittagskarte > logo.txt

FROM golang:${V_GOLANG} AS golang
FROM node:${V_NODE}-slim AS final
WORKDIR /app

RUN npx -y playwright@$V_PLAYWIGHT install-deps chromium

RUN apt-get update > /dev/null 2>&1 && apt-get install -y --no-install-recommends \
    gnupg libc6-dev libnss3-dev libnet-dev build-essential \
    libmagickwand-dev libmagickcore-dev imagemagick libmupdf-dev \
    apt-transport-https ca-certificates curl > /dev/null 2>&1 && \
    apt-get autoremove -y > /dev/null 2>&1 && \
    apt-get clean > /dev/null 2>&1 &&  \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=golang /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:/root/go/bin:${PATH}"

RUN go install github.com/a-h/templ/cmd/templ@latest > /dev/null 2>&1

COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

COPY package.json yarn.lock ./
RUN yarn install --frozen-lockfile --network-timeout 30000 --silent

COPY --from=logo /app/logo.txt /logo.txt
