# Go parameters
GO := go
BUILD_DIR := bin
EXECUTABLE := x410_utils

.PHONY: all build install clean

all: build

build:
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(EXECUTABLE) .

install: build
	@cp $(BUILD_DIR)/$(EXECUTABLE) /usr/local/bin/

clean:
	@rm -rf $(BUILD_DIR)
