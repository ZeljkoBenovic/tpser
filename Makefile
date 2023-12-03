all: build

.PHONY: build
build:
	@echo "Building tpser binary"
	go build -o build/tpser ./cmd
