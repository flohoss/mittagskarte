ARG FEDORA_VERSION
ARG NODE_VERSION
ARG ALPINE_VERSION
FROM fedora:${FEDORA_VERSION} AS goBuilder
ENV GOPROXY=https://proxy.golang.org

RUN dnf -y upgrade --refresh
RUN dnf -y install golang

WORKDIR /app

COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" cmd/mittag/mittag.go

FROM node:${NODE_VERSION}-alpine AS nodeBuilder

COPY ./package.json .
COPY ./yarn.lock .
RUN yarn install --frozen-lockfile

COPY ./web ./web
COPY ./tailwind.config.js .
RUN yarn run tailwind:build

FROM alpine:${ALPINE_VERSION} AS logo
RUN apk add figlet
RUN figlet Mittag > logo.txt

FROM fedora:${FEDORA_VERSION} AS final
RUN dnf -y upgrade --refresh
RUN dnf -y install tesseract tesseract-osd tesseract-langpack-deu tesseract-langpack-eng
RUN dnf -y install ImageMagick poppler-utils libwebp

# Mittag
WORKDIR /app
COPY scripts/entrypoint.sh .

COPY --from=logo /logo.txt .
COPY --from=nodeBuilder /web/static/ ./web/static/
COPY --from=nodeBuilder /web/templates ./web/templates/
COPY --from=goBuilder /app/mittag .

# Envs
ARG VERSION
ENV VERSION=$VERSION
ARG BUILDTIME
ENV BUILDTIME=$BUILDTIME

ENTRYPOINT ["/app/entrypoint.sh"]
