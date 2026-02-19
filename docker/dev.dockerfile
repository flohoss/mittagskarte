ARG V_NODE=25
ARG V_GOLANG=1.25
FROM golang:${V_GOLANG} AS golang
FROM node:${V_NODE}-slim AS final
WORKDIR /app

RUN npx -y playwright@v1.57.0 install-deps chromium

RUN apt-get update > /dev/null 2>&1 && apt-get install -y --no-install-recommends \
    git gnupg libc6-dev libnss3-dev libnet-dev build-essential \
    libmagickwand-dev libmagickcore-dev imagemagick libmupdf-dev \
    apt-transport-https ca-certificates curl > /dev/null 2>&1 && \
    apt-get autoremove -y > /dev/null 2>&1 && \
    apt-get clean > /dev/null 2>&1 &&  \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=golang /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:/root/go/bin:${PATH}"

RUN go install github.com/a-h/templ/cmd/templ@v0.3.960

COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

# https://patorjk.com/software/taag/#p=display&f=Coder+Mini&t=Mittagskarte&x=none&v=4&h=4&w=80&we=false
COPY ./docker/coder-mini.txt /coder-mini.txt

ENTRYPOINT [ "/app/docker/dev.entrypoint.sh" ]
