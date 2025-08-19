ARG V_GOLANG=1.24
ARG V_NODE=lts
ARG V_PLAYWIGHT=20
FROM golang:${V_GOLANG} AS golang
FROM node:${V_NODE}-slim AS final
RUN npx -y playwright@$V_PLAYWIGHT install-deps chromium

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    gnupg libc6-dev libnss3-dev libnet-dev build-essential \
    libmagickwand-dev libmagickcore-dev imagemagick libmupdf-dev \
    apt-transport-https ca-certificates && \
    apt-get autoremove -y && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=golang /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:/root/go/bin:${PATH}"

RUN go install github.com/a-h/templ/cmd/templ@latest

COPY ./go.mod .
COPY ./go.sum .
RUN go mod download
