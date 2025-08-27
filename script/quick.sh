#!/bin/bash

# 当前工作目录应该是项目的根目录
working_directory=$(cd "$(dirname "$0")/.."; pwd)

# ---- 编译阶段 ------
# 构建 agent 后端核心
cd $working_directory/agent
make clean
make build

# 构建 agent-cli
cd $working_directory/agent-cli
make clean
make build

# ---- 运行阶段 ------

pkill -f "obsidian-agent"

bin_directory=$working_directory/agent/bin
# 找到这个文件夹下以 obsidian-agent 开头的可执行文件
executable_file=$(find $bin_directory -name "obsidian-agent*")
# 运行可执行文件
cd $bin_directory/..
exec $executable_file &





