BINARY_NAME := k8slogretrievalapi
BUILD_DIR=bin


.PHONY: test, build, clean, build-static


build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/api/main.go

build-static:
	CGO_ENABLED=0 go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/api/main.go

clean:
	rm -rf $(BUILD_DIR)

test:
	go test ./...
	go test -race -v internal/util/concurrent_map_test.go internal/util/concurrent_map.go