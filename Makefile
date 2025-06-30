.PHONY: FORCE

CMD_DIR=cmd
FIXTURE_DIR=fixture
DIST_DIR=dist
TESTAPP_DIR=test-app

clean_lorem:
	rm -rf $(DIST_DIR)/*;
.PHONY: clean_lorem

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
