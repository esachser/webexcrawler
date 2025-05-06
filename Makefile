# Makefile for building ./cmd/webexcrawler for multiple platforms and architectures
BINARY_NAME=webexcrawler
BUILD_DIR=build

PLATFORMS=windows linux darwin
ARCH=amd64

.PHONY: all clean

all: $(PLATFORMS)

$(PLATFORMS):
	@echo "Building for $@..."
	@mkdir -p $(BUILD_DIR)
	@if [ "$@" = "windows" ]; then \
		GOOS=$@ GOARCH=$(ARCH) go build -o $(BUILD_DIR)/$(BINARY_NAME)-$@-$(ARCH).exe ./cmd/webexcrawler; \
	else \
		GOOS=$@ GOARCH=$(ARCH) go build -o $(BUILD_DIR)/$(BINARY_NAME)-$@-$(ARCH) ./cmd/webexcrawler; \
	fi

clean:
	@echo "Cleaning build directory..."
	@rm -rf $(BUILD_DIR)
