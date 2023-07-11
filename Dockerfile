FROM rust:1.70-bullseye AS builder

COPY ./src /build/src
COPY ./Cargo.toml ./Cargo.lock /build/
COPY ./migration /build/migration

WORKDIR /build

RUN cargo build --release

FROM gcr.io/distroless/cc

COPY --from=builder /build/target/release/blog /server

CMD [ "/server" ]
