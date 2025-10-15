ARG V_GOLANG=1.25
ARG V_NODE=lts
ARG V_PLAYWIGHT=20
FROM golang:${V_GOLANG} AS golang
FROM node:${V_NODE}-slim AS final
WORKDIR /app

RUN npx -y playwright@$V_PLAYWIGHT install-deps chromium

RUN apt-get -qq update && apt-get -qq install -y --no-install-recommends \
    gnupg libc6-dev libnss3-dev libnet-dev build-essential \
    libmagickwand-dev libmagickcore-dev imagemagick libmupdf-dev \
    apt-transport-https ca-certificates curl git docker.io > /dev/null 2>&1 && \
    apt-get autoremove -y && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=golang /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:/root/go/bin:${PATH}"

RUN go install github.com/a-h/templ/cmd/templ@latest
RUN go install github.com/goreleaser/goreleaser/v2@latest

COPY ./go.mod .
COPY ./go.sum .
RUN go mod download
