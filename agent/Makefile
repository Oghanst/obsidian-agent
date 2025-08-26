# Makefile for obsidian-agent

# Variables
APP_NAME := obsidian-agent
SRC_DIR := ./cmd/agentd
BIN_DIR := bin
MAIN_FILE := main.go

.PHONY: all build clean run

all: build

build:
# 从 VERSION 文件中读取版本号
	@echo "Building $(APP_NAME)..."
	@echo "Version: $(shell cat VERSION)"
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME).$(shell cat VERSION) $(SRC_DIR)/$(MAIN_FILE)

run: build
	$(BIN_DIR)/$(APP_NAME)

clean:
	rm -rf $(BIN_DIR)