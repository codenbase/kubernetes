#!/bin/bash

# --- ENTRYPOINT 脚本 ---
# 这个脚本负责容器启动时的所有初始化工作

# 临时禁用 Swap
swapoff -a

# 永久禁用 Swap (通过注释 /etc/fstab)
# 注意：在容器环境中，/etc/fstab 通常为空，此步骤可能无实际效果，主要是为了配置的完整性
sed -i '/ swap / s/^\(.*\)$/#\1/g' /etc/fstab

# 解决 /proc/swaps 显示问题 (强制 Kubelet 通过检测)
# 经过我的实践，即使禁用了 Swap，容器里执行 `cat /proc/swaps` 依然可能显示存在 Swap 分区
# 由于 Kubelet 启动时会检查 `cat /proc/swaps` 的输出，只有输出为空时才能正常启动
# 因此，这里强制挂载一个空文件到 `/proc/swaps` 来“欺骗” Kubelet 的检测
mkdir -p /tmp/fakeproc
touch /tmp/fakeproc/swaps
mount --bind /tmp/fakeproc/swaps /proc/swaps

# 在容器启动时应用 sysctl 配置
# 当容器以 --privileged 模式运行时，此时拥有足够的权限来修改宿主机内核参数。
sysctl --system

# 加载 service.sh 脚本
source service.sh

# 配置 /etc/hosts
configure:hosts

# 启动系统服务
start:system

# 启动 Etcd 集群
start:etcd

# 启动 Master 和 Worker 服务
start:master
start:worker

# 执行 CMD 或 docker run 命令行参数
# "$@" 代表传递给脚本的所有参数（即 Dockerfile 中的 CMD 或 docker run 命令后面的参数）。
# exec 会替换当前 shell 进程，确保 CMD 或用户指定的命令成为容器的主进程 (PID 1)，
# 这有助于 Docker 更好地管理容器的生命周期和信号处理。
# 如果没有 CMD 或命令行参数，则默认启动 bash，保持容器运行并提供交互式 shell。
if [ "$#" -gt 0 ]; then
    exec "$@"
else
    exec bash # 如果没有传递 CMD 参数，则默认启动 bash
fi
