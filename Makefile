.PHONY: FORCE

CMD_DIR=cmd
FIXTURE_DIR=fixture
DIST_DIR=dist
TESTAPP_DIR=test-app

DBUILD_CMD = buildx build
DBUILD_ARGS = --progress=plain
DBUILD_REPO = localhost:5000
DBUILD_VERS = latest

DOCK_ROOT_CTX = ./

# BINARY = $(DIST_DIR)/

# $(BINARY):

ifdef
endif

wasm_viewer: clean_docker clean_build
	rm -rf $(DIST_DIR)/viewer/*
	mkdir -pv $(DIST_DIR)/viewer/
	docker $(DBUILD_CMD) $(DBUILD_ARGS) --file=./viewer.Dockerfile --target=export-viewer --output type=local,dest=$(DIST_DIR)/viewer/ $(DOCK_ROOT_CTX)
.PHONY: wasm_viewer

build_lorem:
	if [[ ! -d $(DIST_DIR)/lorem ]]; then \
		go build -ldflags="-s -w" -mod=vendor -o $(DIST_DIR)/lorem/lorem.bin $(TESTAPP_DIR)/lorem/main.go; \
	fi;
.PHONY: build_lorem

pack_lorem: build_lorem
	if [[ ! -d $(DIST_DIR)/lorem ]]; then \
		cp -vrf $(FIXTURE_DIR) $(DIST_DIR)/lorem; \
	fi;
	if [[ ! -f $(DIST_DIR)/lorem.tar.gz ]]; then \
		tar -C $(DIST_DIR)/lorem -c fixture lorem.bin | gzip -9n > $(DIST_DIR)/lorem.tar.gz; \
	fi;
.PHONY: pack_lorem

gorun_lorem: pack_lorem
	go run $(CMD_DIR)/main.go run --main-file=lorem.bin -- $(DIST_DIR)/lorem.tar.gz
.PHONY: gorun_lorem

clean_docker:
	echo -e 'y' | docker system prune
.PHONY: clean_docker

clean_build:
	rm -rf $(DIST_DIR)/*
	mkdir -pv $(DIST_DIR)
.PHONY: clean_build

clean_lorem:
	rm -rf $(DIST_DIR)/*;
.PHONY: clean_lorem

clean: clean_build clean_docker clean_lorem
.PHONY: clean
