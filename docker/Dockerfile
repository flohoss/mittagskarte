ARG V_GOLANG
ARG V_DEBIAN
ARG V_NODE
ARG V_ALPINE
FROM golang:${V_GOLANG} AS goBuilder
RUN apt-get update -qq
RUN apt-get install -y libtesseract-dev libleptonica-dev

WORKDIR /app
COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

COPY . .
RUN go build -tags ocr -ldflags="-s -w" cmd/mittag/mittag.go

FROM node:${V_NODE}-alpine AS nodeBuilder
WORKDIR /web

COPY ./web .
RUN yarn install --frozen-lockfile
RUN yarn run tw:build

FROM alpine:${V_ALPINE} AS logo
RUN apk add figlet
RUN figlet Mittag > logo.txt

FROM debian:${V_DEBIAN}-slim AS final
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

ENTRYPOINT ["/app/entrypoint.sh"]
