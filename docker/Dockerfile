ARG GOLANG_VERSION
ARG DEBIAN_VERSION
ARG NODE_VERSION
ARG ALPINE_VERSION
FROM golang:${GOLANG_VERSION} AS goBuilder
RUN apt-get update -qq
RUN apt-get install -y libtesseract-dev libleptonica-dev

WORKDIR /app
COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

COPY . .
RUN go build -tags ocr -ldflags="-s -w" cmd/mittag/mittag.go

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

FROM debian:${DEBIAN_VERSION}-slim AS final
RUN apt-get update -qq
RUN apt-get install -y imagemagick poppler-utils wv unrtf tidy
RUN sed -i '/disable ghostscript format types/,+6d' /etc/ImageMagick-6/policy.xml
RUN apt-get install -y libtesseract-dev libleptonica-dev
RUN apt-get install -y tesseract-ocr-eng tesseract-ocr-deu

# Mittag
WORKDIR /app
COPY scripts/entrypoint.sh .

COPY --from=logo /logo.txt .
COPY --from=nodeBuilder /web/static/ ./web/static/
COPY --from=nodeBuilder /web/templates ./web/templates/
COPY --from=goBuilder /app/mittag .
COPY ./configs/ ./configs/

# Envs
ARG APP_VERSION
ENV APP_VERSION=$APP_VERSION
ARG BUILD_TIME
ENV BUILD_TIME=$BUILD_TIME

ENTRYPOINT ["/app/entrypoint.sh"]
