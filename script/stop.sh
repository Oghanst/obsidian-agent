#!/bin/bash

# 当前工作目录应该是项目的根目录
working_directory=$(cd "$(dirname "$0")/.."; pwd)

# 找到任何以 obsidian-agent 开头的进程并终止它们
pkill -f "obsidian-agent"