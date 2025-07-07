FROM alpine:3.21 AS build-viewer
USER root
WORKDIR /export
ENV DEV=0
COPY . .
RUN \
    apk update && apk upgrade; \
    apk add \
        curl \
        rustup \
        rust-wasm \
        wasm-bindgen \
        wasm-pack \
        trunk \
    ; \
    rustup-init -y --profile=default; \
    . "$HOME/.cargo/env"; \
    rustup target add wasm32-unknown-unknown; \
    cargo install --path .; \
    trunk build \
        --config=/export/viewer/Trunk.toml \
        --dist=/export/dist/viewer/ \
        --cargo-profile=$(if [ ! -z "$DEV" ] && [ "$DEV" -ne 0 ]; then echo "release"; else echo "dev"; fi) \
        /export/viewer/index.html;

FROM scratch AS export-viewer
COPY --from=build-viewer /export/dist/viewer/ /
ENTRYPOINT ["/export/dist/viewer/"]
