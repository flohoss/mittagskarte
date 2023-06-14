ARG FEDORA_VERSION
ARG NODE_VERSION
ARG ALPINE_VERSION
FROM fedora:${FEDORA_VERSION} AS goBuilder
ENV GOPROXY=https://proxy.golang.org

RUN dnf -y upgrade --refresh
RUN dnf -y install golang

WORKDIR /src
COPY ./src/go.mod .
COPY ./src/go.sum .
RUN go mod download

COPY ./src/ .
RUN go build -ldflags="-s -w"

FROM node:${NODE_VERSION}-alpine AS nodeBuilder

COPY package.json .
COPY yarn.lock .
RUN yarn install --frozen-lockfile

COPY tailwind.config.js .
COPY ./src/templates/ ./src/templates/
COPY ./src/static/ ./src/static/
RUN yarn run tailwind:build

FROM alpine:${ALPINE_VERSION} AS logo
RUN apk add figlet
RUN figlet Mittag > logo.txt

FROM fedora:${FEDORA_VERSION} AS final
RUN dnf -y upgrade --refresh
RUN dnf -y install tesseract tesseract-osd tesseract-langpack-deu tesseract-langpack-eng
RUN dnf -y install ImageMagick poppler-utils libwebp

# Mittag
WORKDIR /src
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

COPY --from=logo /logo.txt .
COPY --from=nodeBuilder /src/static/ ./static/
COPY --from=nodeBuilder /src/templates ./templates/
COPY --from=goBuilder /src/mittag .

# Envs
ARG VERSION
ENV VERSION=$VERSION
ARG BUILDTIME
ENV BUILDTIME=$BUILDTIME

ENTRYPOINT ["/src/entrypoint.sh"]
