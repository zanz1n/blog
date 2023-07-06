FROM rust:1.70-bullseye AS builder

RUN cargo install sea-orm-cli

COPY migration/migrate.sh /usr/bin/migrate

RUN chmod +x /usr/bin/migrate

WORKDIR /build

COPY Cargo.toml Cargo.lock /build/
COPY migration/Cargo.toml /build/migration/
COPY migration/src /build/migration/src

RUN mkdir src

RUN cat>./src/main.rs <<EOF
fn main() {
    println!("Hello World");
}
EOF

RUN sh -c "cargo run --manifest-path ./migration/Cargo.toml; exit 0"

CMD [ "bash" ]
