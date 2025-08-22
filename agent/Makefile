# Makefile for obsidian-agent

# Variables
APP_NAME := obsidian-agent
SRC_DIR := ./cmd
BIN_DIR := bin
MAIN_FILE := agentd.go

.PHONY: all build clean run

all: build

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) $(SRC_DIR)/$(MAIN_FILE)

run: build
	$(BIN_DIR)/$(APP_NAME)

clean:
	rm -rf $(BIN_DIR)