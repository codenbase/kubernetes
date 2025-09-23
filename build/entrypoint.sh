#!/bin/bash

# --- ENTRYPOINT 脚本 ---
# 这个脚本负责容器启动时的所有初始化工作

# 任何命令失败时立即退出
set -o errexit
# 遇到未定义的变量时立即退出
set -o nounset
# 管道中的任何一个命令失败，整个管道都失败
set -o pipefail

# 定义全局变量，支持颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO] $1${NC}"
}

log_warn() {
    echo -e "${YELLOW}[WARN] $1${NC}"  
}

log_error() {
    echo -e "${RED}[ERROR] $1${NC}"
}

# 检查 cgroup 版本
if [[ -f "/sys/fs/cgroup/cgroup.controllers" ]]; then
    log_info "==> Detected cgroup v2, no need for manual setup."
else
    # 绝大多数人不可能是 cgroup v1，本课程也不支持此环境，这里直接报错
    log_warn "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
    log_warn "!!! WARNING: cgroup v1 is detected.                            !!!"
    log_warn "!!! This environment is not supported by the main tutorial.    !!!"
    log_warn "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
    # 以非 0 状态码退出，表示错误，阻止容器继续运行
    exit 1
fi

# 禁用 Swap
disable_swap() {
    # 在容器环境中，swapoff 很可能会因为没有 swap 而失败，使用 || true 来忽略这个可预期的错误
    swapoff -a || true
    sed -i '/ swap / s/^\(.*\)$/#\1/g' /etc/fstab  || true

    # 在容器环境中，以上两条命令可能无实际效果
    # 因此，这里强制挂载一个空文件到 `/proc/swaps` 可以达到禁用swap的效果
    if [ -f /proc/swaps ]; then
        mkdir -p /tmp/fakeproc
        touch /tmp/fakeproc/swaps
        mount --bind /tmp/fakeproc/swaps /proc/swaps
    fi

    log_info "==> disable swap completed."
}

disable_swap

# 使用 exec 执行传递给脚本的命令 (即 Dockerfile 中的 CMD)
# 这会将 systemd 进程替换当前的 shell 进程，使其成为 PID 1
exec "$@"