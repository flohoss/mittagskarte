ARG V_NODE=25
ARG V_GOLANG=1.26
ARG V_AIR=1.64.5
FROM golang:${V_GOLANG} AS golang
FROM node:${V_NODE}-slim AS final
WORKDIR /app

RUN npx -y playwright@v1.58.0 install-deps chromium > /dev/null 2>&1

RUN apt-get update > /dev/null 2>&1 && apt-get install -y --no-install-recommends \
    git gnupg libc6-dev libnss3-dev libnet-dev build-essential \
    libmagickwand-dev libmagickcore-dev imagemagick libmupdf-dev \
    apt-transport-https ca-certificates curl > /dev/null 2>&1 && \
    apt-get autoremove -y > /dev/null 2>&1 && \
    apt-get clean > /dev/null 2>&1 &&  \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=golang /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:/root/go/bin:${PATH}"

ARG V_AIR
RUN go install github.com/air-verse/air@v${V_AIR}

ENV APP_VERSION=v0.0.0.0-dev

COPY ./go.mod ./go.sum ./
RUN go mod download
