ARG V_GOLANG=1.26
ARG V_NODE=25
ARG REPO_URL

FROM golang:${V_GOLANG} AS golang
FROM node:${V_NODE}-slim AS backend-builder
WORKDIR /app/backend

RUN apt-get update > /dev/null 2>&1 && apt-get install -y --no-install-recommends \
    gnupg libc6-dev libnss3-dev libnet-dev build-essential \
    libmagickwand-dev libmagickcore-dev imagemagick libmupdf-dev \
    apt-transport-https ca-certificates > /dev/null 2>&1 && \
    apt-get autoremove -y > /dev/null 2>&1 && \
    apt-get clean > /dev/null 2>&1 && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=golang /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:/root/go/bin:${PATH}"

COPY ./backend/go.mod ./backend/go.sum ./
RUN go mod download

COPY ./backend/ ./
RUN go build -ldflags="-s -w" -o /out/mittag .

FROM node:${V_NODE}-slim AS frontend-builder
WORKDIR /app/frontend

ARG APP_VERSION
ENV VITE_APP_VERSION=${APP_VERSION}
ARG REPO_URL
ENV VITE_REPO_URL=${REPO_URL}

COPY ./frontend/package.json ./frontend/yarn.lock ./
RUN yarn install --frozen-lockfile --network-timeout 30000 --silent

COPY ./frontend/ ./
RUN yarn build

FROM node:${V_NODE}-slim AS final
WORKDIR /app

RUN npx -y playwright@v1.59.1 install-deps chromium > /dev/null 2>&1

RUN apt-get update > /dev/null 2>&1 && apt-get install -y --no-install-recommends \
    libnss3 libnet1 dumb-init ca-certificates curl tzdata \
    libmagickwand-6.q16-6 imagemagick libmupdf-dev > /dev/null 2>&1 && \
    apt-get autoremove -y > /dev/null 2>&1 && \
    apt-get clean > /dev/null 2>&1 && \
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
