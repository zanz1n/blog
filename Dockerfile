FROM golang:1 AS builder

ARG VERSION=""

WORKDIR /build

ENV CGO_ENABLED=0
ENV SKIPTESTS=1
ENV OUT=bin/blog

RUN go env -w GOCACHE=/go-cache
RUN go env -w GOMODCACHE=/gomod-cache

COPY . .

RUN --mount=type=cache,target=/gomod-cache \
    --mount=type=cache,target=/go-cache \
    make

FROM gcr.io/distroless/static-debian12

COPY --from=builder /build/bin/blog /usr/bin/blog

ENTRYPOINT [ "/usr/bin/blog" ]
