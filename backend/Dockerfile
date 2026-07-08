ARG V_NODE
ARG V_GOLANG
ARG V_AIR

FROM golang:${V_GOLANG} AS golang
FROM node:${V_NODE}-slim AS final
WORKDIR /app

RUN apt-get update > /dev/null 2>&1 && apt-get install -y --no-install-recommends \
    git gnupg libc6-dev libnss3-dev libnet-dev build-essential \
    libmagickwand-7.q16-dev libmagickcore-7.q16-dev imagemagick libmupdf-dev \
    libheif1 libheif-plugin-libde265 libheif-plugin-dav1d \
    apt-transport-https ca-certificates curl tzdata > /dev/null 2>&1 && \
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

ARG V_PLAYWRIGHT
RUN go run github.com/mxschmitt/playwright-go/cmd/playwright@v${V_PLAYWRIGHT} install --with-deps chromium > /dev/null 2>&1
