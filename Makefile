CMD_DIR=cmd
FIXTURE_DIR=fixture
TEST_DIR=test

build_lorem:
	go build -ldflags="-s -w" -mod=vendor -o $(TEST_DIR)/lorem $(CMD_DIR)/lorem/main.go
	tar -cvzf $(TEST_DIR)/lorem.tar --directory=./$(TEST_DIR) ./lorem ../$(FIXTURE_DIR)
.phony: build_lorem
