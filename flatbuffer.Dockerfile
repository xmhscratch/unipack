FROM alpine:3.21 AS fbgen-builder
ENV PATH=${PATH}:/usr/local/bin/
RUN \
    apk update && apk upgrade; \
    apk add \
        curl \
        alpine-sdk \
        cmake \
        flatc \
    ; \
    mkdir -pv ./flatbuffers; \
    git clone --depth=1 --branch=v25.2.10 https://github.com/google/flatbuffers.git ./flatbuffers; \
	cd ./flatbuffers; \
	cmake -G "Unix Makefiles"; \
	make -j$(getconf _NPROCESSORS_ONLN) && make install;

FROM fbgen-builder AS gen-rust
WORKDIR /export/
COPY fixture/*.fbs ./
RUN \
    mkdir -pv /export/fbgen-rust/; \
    flatc --rust \
        --filename-suffix "" \
        -o /export/fbgen-rust/ \
        --rust-module-root-file \
        ./*.fbs;

FROM fbgen-builder AS gen-go
WORKDIR /export/
COPY fixture/*.fbs ./
RUN \
    mkdir -pv /export/fbgen-go/; \
    flatc --go \
        --filename-suffix "" \
        -o /export/fbgen-go/ \
        ./*.fbs;

FROM scratch AS export-gen
COPY --from=gen-rust /export/fbgen-rust/ /fbgen-rust/src/schema/
COPY --from=gen-go /export/fbgen-go/ /fbgen-go/schema/
ENTRYPOINT ["/export/"]
