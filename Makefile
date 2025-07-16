.PHONY: FORCE

ROOT_DIR=./
CMD_DIR=cmd
FIXTURE_DIR=fixture
DIST_DIR=dist
PKG_DIR=pkg
TESTAPP_DIR=test-app

DBUILD_CMD = buildx build
DBUILD_ARGS = --progress=plain
DBUILD_REPO = localhost:5000
DBUILD_VERS = latest

DOCK_ROOT_CTX = $(ROOT_DIR)

# BINARY = $(DIST_DIR)/

# $(BINARY):

ifdef
endif

fbgen_schema:
	docker $(DBUILD_CMD) $(DBUILD_ARGS) --file=./flatbuffer.Dockerfile --target=export-gen --output type=local,dest=$(PKG_DIR) $(DOCK_ROOT_CTX)

wasm_viewer: clean_docker clean_build
	rm -rf $(DIST_DIR)/viewer/*
	mkdir -pv $(DIST_DIR)/viewer/
	docker $(DBUILD_CMD) $(DBUILD_ARGS) --file=./viewer.Dockerfile --target=export-viewer --output type=local,dest=$(DIST_DIR)/viewer/ $(DOCK_ROOT_CTX)

build_lorem:
	if [[ ! -d $(DIST_DIR)/lorem ]]; then \
		go build -ldflags="-s -w" -mod=vendor -o $(DIST_DIR)/lorem/lorem.bin $(TESTAPP_DIR)/lorem/main.go; \
		cp -vr $(FIXTURE_DIR)/lorem $(DIST_DIR)/lorem/fixture; \
	fi;

pack_lorem: build_lorem
	if [[ ! -f $(DIST_DIR)/lorem.tar.gz ]]; then \
		tar -C $(DIST_DIR)/lorem/ -c fixture lorem.bin | gzip -9n > $(DIST_DIR)/lorem.tar.gz; \
	fi;

gorun_lorem: pack_lorem
	go run $(CMD_DIR)/main.go run --main-file=lorem.bin -- $(DIST_DIR)/lorem.tar.gz

go.mod:
	go mod tidy;

go.sum:
	go mod vendor;

clean_go:
	go clean -cache;
	go clean -modcache;

clean_docker:
	echo -e 'y' | docker system prune
.PHONY: clean_docker

clean_build:
	rm -rf $(DIST_DIR)/*
	mkdir -pv $(DIST_DIR)
.PHONY: clean_build

clean_lorem:
	rm -rf $(DIST_DIR)/*;

clean: clean_build clean_docker clean_lorem clean_go
.PHONY: clean
