FROM oven/bun AS bun_builder

WORKDIR /app

COPY ./web/package.json ./web/bun.lock /app/
RUN bun install --frozen-lockfile

COPY ./web .
RUN bun run build

FROM golang:1 AS builder

ARG VERSION

WORKDIR /build

ENV DEBUG=0
ENV CGO_ENABLED=1
ENV SKIPTESTS=1
ENV OUTPUT=bin/blog
ENV GOTAGS=embed

RUN go env -w GOCACHE=/go-cache
RUN go env -w GOMODCACHE=/gomod-cache

COPY . .
COPY --from=bun_builder /app/dist /build/web/dist
COPY --from=bun_builder /app/.source-map.json /build/web/.source-map.json

RUN --mount=type=cache,target=/gomod-cache \
    --mount=type=cache,target=/go-cache \
    make

FROM gcr.io/distroless/cc-debian12

COPY --from=builder /build/bin/blog /blog

ENTRYPOINT [ "/blog" ]
CMD [ "--migrate" ]
