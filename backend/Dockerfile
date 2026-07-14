ARG V_NODE
ARG V_GOLANG

FROM golang:${V_GOLANG} AS golang
FROM node:${V_NODE}-slim AS final
WORKDIR /app

RUN apt-get update > /dev/null 2>&1 && apt-get install -y --no-install-recommends \
    git build-essential libc6-dev \
    curl ca-certificates tzdata > /dev/null 2>&1 && \
    apt-get autoremove -y > /dev/null 2>&1 && \
    apt-get clean > /dev/null 2>&1 &&  \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=golang /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:/root/go/bin:${PATH}"

ARG V_AIR
RUN go install github.com/air-verse/air@v${V_AIR}
ARG V_OGEN
RUN go install github.com/ogen-go/ogen/cmd/ogen@v${V_OGEN}

ENV APP_VERSION=v0.0.0.0-dev

COPY ./go.mod ./go.sum ./
RUN go mod download

ARG V_PLAYWRIGHT
RUN go run github.com/mxschmitt/playwright-go/cmd/playwright@v${V_PLAYWRIGHT} install --with-deps chromium > /dev/null 2>&1
