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

# 获取容器主机名和 IP 地址
HOSTNAME=$(hostname)
IP=$(hostname -I | awk '{print $1}')
echo "Resolved Container Hostname: ${HOSTNAME}"
echo "Resolved Container IP: ${IP}"

# 启动 Etcd 服务
# 创建 Etcd 数据和 WAL 目录
mkdir -p /var/lib/etcd/data /var/lib/etcd/wal
# /etc/kubernetes/etcd/etcd-args.template 是已经挂载到容器内的模板文件
ETCD_SETUP_ARGS_TEMPLATE_FILE="/etc/kubernetes/etcd/etcd-setup-args.template"
# 检查模板文件是否存在
if [ ! -f "${ETCD_SETUP_ARGS_TEMPLATE_FILE}" ]; then
    echo "Error: Etcd arguments template file not found at ${ETCD_SETUP_ARGS_TEMPLATE_FILE}"
    exit 1
fi
# 替换模板中的 ##NODE_NAME## 和 ##NODE_IP## 占位符
GENERATED_ETCD_SETUP_ARGS=$(sed -e "s/##NODE_NAME##/${HOSTNAME}/" -e "s/##NODE_IP##/${IP}/" "${ETCD_SETUP_ARGS_TEMPLATE_FILE}")
/usr/local/bin/etcd ${GENERATED_ETCD_SETUP_ARGS} & 
# 简短的等待，确保 Etcd 有时间启动（可选，但对于依赖 Etcd 的服务很有用）
sleep 5

# 启动 chrony 服务（设置系统时钟同步）
service chrony start

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

# 检查 K8S_HOSTS 环境变量是否存在
# 如果存在，则将其内容作为额外的 hostname-IP 映射追加到 /etc/hosts 文件中。
# 我们使用 -e K8S_HOSTS="..." 的方式在 docker run 命令中传入这些动态信息。
if [ -n "$K8S_HOSTS" ]; then
    echo -e "$K8S_HOSTS" >> /etc/hosts
fi

# 检查 NODE_TYPE 环境变量是否存在
# 如果存在，则根据 NODE_TYPE = master 或 node 来启动不同的服务
if [ -n "$NODE_TYPE" ]; then
    if [ "$NODE_TYPE" = "master" ]; then
        # 创建各个组件的数据目录
        mkdir -p /var/lib/data/kube-apiserver /var/lib/data/kube-controller-manager /var/lib/data/kube-scheduler

        # kube-apiserver 启动参数模板文件
        KUBE_APISERVER_SETUP_ARGS_TEMPLATE_FILE="/etc/kubernetes/kube-apiserver-setup-args.template"
        # # 检查模板文件是否存在
        # if [ ! -f "${KUBE_APISERVER_SETUP_ARGS_TEMPLATE_FILE}" ]; then
        #     echo "Error: Etcd arguments template file not found at ${KUBE_APISERVER_SETUP_ARGS_TEMPLATE_FILE}"
        #     exit 1
        # fi
        # 替换模板中的 ##NODE_IP## 占位符
        GENERATED_KUBE_APISERVER_SETUP_ARGS=$(sed -e "s/##NODE_IP##/${IP}/" "${KUBE_APISERVER_SETUP_ARGS_TEMPLATE_FILE}")
        # 启动 kube-apiserver
        /usr/local/bin/kube-apiserver ${GENERATED_KUBE_APISERVER_SETUP_ARGS} &

        # kube-controller-manager 启动参数模板文件
        KUBE_CONTROLLER_MANAGER_SETUP_ARGS_TEMPLATE_FILE="/etc/kubernetes/kube-controller-manager-setup-args.template"
        # 替换模板中的 ##NODE_IP## 占位符
        GENERATED_KUBE_CONTROLLER_MANAGER_SETUP_ARGS=$(sed -e "s/##NODE_IP##/${IP}/" "${KUBE_CONTROLLER_MANAGER_SETUP_ARGS_TEMPLATE_FILE}")
        # 启动 kube-controller-manager
        /usr/local/bin/kube-controller-manager ${GENERATED_KUBE_CONTROLLER_MANAGER_SETUP_ARGS} &

        # kube-scheduler 启动参数模板文件
        KUBE_SCHEDULER_SETUP_ARGS_TEMPLATE_FILE="/etc/kubernetes/kube-scheduler-setup-args.template"
        # 替换模板中的 ##NODE_IP## 占位符
        GENERATED_KUBE_SCHEDULER_SETUP_ARGS=$(sed -e "s/##NODE_IP##/${IP}/" "${KUBE_SCHEDULER_SETUP_ARGS_TEMPLATE_FILE}")
        # 启动 kube-scheduler
        /usr/local/bin/kube-scheduler ${GENERATED_KUBE_SCHEDULER_SETUP_ARGS} &

        echo "Master 节点服务启动完成"
    elif [ "$NODE_TYPE" = "node" ]; then
        # 启动 xxx
        echo "Node 类型服务启动逻辑待实现"
    fi
fi

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
