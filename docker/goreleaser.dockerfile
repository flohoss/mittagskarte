ARG V_GOLANG=1.25
ARG V_NODE=lts
ARG V_ALPINE=3
ARG V_PLAYWRIGHT=20
FROM alpine:${V_ALPINE} AS logo
WORKDIR /app
RUN apk add figlet
RUN figlet Mittagskarte > logo.txt

FROM node:${V_NODE}-alpine AS node-builder
WORKDIR /app

COPY package.json yarn.lock ./
RUN yarn install --frozen-lockfile --network-timeout 30000

COPY ./views/ ./views/
COPY ./assets/ ./assets/
RUN yarn run tw:build

FROM node:${V_NODE}-slim AS final
WORKDIR /app

RUN npx -y playwright@$V_PLAYWRIGHT install-deps chromium

RUN apt-get update && apt-get install -y --no-install-recommends \
    libmagickwand-6.q16-6 imagemagick libmupdf-dev libnss3 libnet1 \
    apt-transport-https ca-certificates curl dumb-init && \
    apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=logo /app/logo.txt .
COPY --from=node-builder /app/assets/favicon/ ./assets/favicon/
COPY --from=node-builder /app/assets/js/ ./assets/js/
COPY --from=node-builder /app/assets/css/style.css ./assets/css/style.css

# Copy compiled Go binary from GoReleaser
COPY mittagskarte ./mittagskarte

EXPOSE 8156

ARG APP_VERSION
ENV APP_VERSION=$APP_VERSION
ARG BUILD_TIME
ENV BUILD_TIME=$BUILD_TIME
ARG REPO
ENV REPO=$REPO

RUN chown -R 1000:1000 /app

ENTRYPOINT ["dumb-init", "--"]
USER node
CMD ["/app/mittagskarte"]
