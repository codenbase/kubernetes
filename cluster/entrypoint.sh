#!/bin/bash

# --- ENTRYPOINT 脚本 ---
# 这个脚本负责容器启动时的所有初始化工作

# 禁用 Swap。虽然 fstab 已配置，但容器运行时仍需执行此命令确保禁用。
swapoff -a
# 默认 Kubelet 启动时会检查 `cat /proc/swaps` 的输出，只有输出为空时才能正常启动。
# 然而，经过我的实践，即使执行 `swapoff -a` 禁用了 Swap，容器里执行 `cat /proc/swaps` 依然会有输出。
# 因此，这里强制挂载一个空文件到 `/proc/swaps` 即可解决此问题
mkdir -p /tmp/fakeproc
touch /tmp/fakeproc/swaps
mount --bind /tmp/fakeproc/swaps /proc/swaps

# 在容器启动时应用 sysctl 配置
# 当容器以 --privileged 模式运行时，此时拥有足够的权限来修改宿主机内核参数。
sysctl --system

# 启动 Containerd (后台运行)
containerd &

# 启动 SSH 服务
service ssh start

# 动态配置 /etc/hosts
# 首先，清空 /etc/hosts 并添加 localhost 及 IPv6 的默认条目，确保基础网络功能正常。
echo "127.0.0.1 localhost" > /etc/hosts
echo "::1     localhost ip6-localhost ip6-loopback" >> /etc/hosts
echo "fe00::  ip6-localnet" >> /etc/hosts
echo "ff00::  ip6-mcastprefix" >> /etc/hosts
echo "ff02::1 ip6-allnodes" >> /etc/hosts
echo "ff02::2 ip6-allrouters" >> /etc/hosts
echo "" >> /etc/hosts # 添加一个空行，美观

# 检查 K8S_HOSTS 环境变量是否存在。
# 如果存在，则将其内容作为额外的 hostname-IP 映射追加到 /etc/hosts 文件中。
# 我们使用 -e K8S_HOSTS="..." 的方式在 docker run 命令中传入这些动态信息。
if [ -n "$K8S_HOSTS" ]; then
    echo -e "$K8S_HOSTS" >> /etc/hosts
fi

# 执行 CMD 或 docker run 命令行参数。
# "$@" 代表传递给脚本的所有参数（即 Dockerfile 中的 CMD 或 docker run 命令后面的参数）。
# exec 会替换当前 shell 进程，确保 CMD 或用户指定的命令成为容器的主进程 (PID 1)，
# 这有助于 Docker 更好地管理容器的生命周期和信号处理。
# 如果没有 CMD 或命令行参数，则默认启动 bash，保持容器运行并提供交互式 shell。
if [ "$#" -gt 0 ]; then
    exec "$@"
else
    exec bash # 如果没有传递 CMD 参数，则默认启动 bash
fi
