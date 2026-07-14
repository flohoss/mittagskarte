ARG V_GOLANG
ARG V_NODE

FROM golang:${V_GOLANG} AS golang
FROM node:${V_NODE}-slim AS backend-builder
WORKDIR /app/backend

RUN apt-get update > /dev/null 2>&1 && apt-get install -y --no-install-recommends \
    git build-essential libc6-dev \
    ca-certificates > /dev/null 2>&1 && \
    apt-get autoremove -y > /dev/null 2>&1 && \
    apt-get clean > /dev/null 2>&1 && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=golang /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:/root/go/bin:${PATH}"

ARG V_OGEN
RUN go install github.com/ogen-go/ogen/cmd/ogen@v${V_OGEN}

COPY ./backend/go.mod ./backend/go.sum ./
RUN go mod download

COPY ./backend/ ./
RUN go generate ./pkg/snapotter
RUN go build -ldflags="-s -w" -o /out/mittag .

ARG V_PLAYWRIGHT
RUN go install github.com/mxschmitt/playwright-go/cmd/playwright@v${V_PLAYWRIGHT}

FROM node:${V_NODE}-slim AS frontend-builder
WORKDIR /app/frontend

ARG APP_VERSION
ENV VITE_APP_VERSION=${APP_VERSION}
ARG REPO_URL
ENV VITE_REPO_URL=${REPO_URL}

COPY ./frontend/package.json ./frontend/package-lock.json ./
RUN npm ci --silent

COPY ./frontend/ ./
RUN npm run build

FROM node:${V_NODE}-slim AS final
WORKDIR /app

RUN apt-get update > /dev/null 2>&1 && apt-get install -y --no-install-recommends \
    dumb-init ca-certificates curl tzdata > /dev/null 2>&1 && \
    apt-get autoremove -y > /dev/null 2>&1 && \
    apt-get clean > /dev/null 2>&1 && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=backend-builder /root/go/bin/playwright /usr/local/bin/playwright
RUN playwright install --with-deps chromium > /dev/null 2>&1 && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

ARG APP_VERSION
ENV APP_VERSION=${APP_VERSION}
ARG BUILD_TIME
ENV BUILD_TIME=${BUILD_TIME}
ARG REPO_URL
ENV REPO_URL=${REPO_URL}

COPY --from=backend-builder /out/mittag /app/mittag
COPY --from=frontend-builder /app/frontend/dist /app/dist

EXPOSE 8090

ENTRYPOINT ["dumb-init", "--", "/app/mittag", "serve", "--http=0.0.0.0:8090"]
