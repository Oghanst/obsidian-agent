# Makefile for agent-cli

# Variables
APP_NAME := agent-cli
SRC_DIR := ./cmd/agent-cli
BIN_DIR := bin
MAIN_FILE := main.go
.PHONY: all build clean 

all: build

build:
	# 从 VERSION 文件中读取版本号
	@echo "Building $(APP_NAME)..."
	@echo "Version: $(shell cat ../VERSION)"
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME).$(shell cat ../VERSION) $(SRC_DIR)/$(MAIN_FILE)

clean:
	rm -rf $(BIN_DIR)