#!/bin/bash

# --- SERVICE 脚本 ---
# 这个脚本负责定义公共函数，供 entrypoint.sh 调用

# 防止重复加载
if [ -n "$SERVICE_LOADED" ]; then
  return
fi
SERVICE_LOADED=1

# 打印信息日志
log:info() {
    echo "[INFO] $1"
}

# 打印错误日志
log:error() {
    echo "[ERROR] $1"
}

# 获取容器 主机名 和 IP
HOSTNAME=$(hostname)
IP=$(hostname -I | awk '{print $1}')
log:info "Resolved Container Hostname: ${HOSTNAME}"
log:info "Resolved Container IP: ${IP}"

# 配置 HOSTS
configure:hosts() {
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
}

start:system() {
    # 启动 chrony 服务（设置系统时钟同步）
    service chrony start

    # 启动 SSH 服务
    service ssh start
}

start:etcd() {
    # 启动 Etcd 服务
    # 创建 Etcd 数据和 WAL 目录
    mkdir -p /var/lib/etcd/data /var/lib/etcd/wal
    # /etc/kubernetes/etcd/etcd-args.template 是已经挂载到容器内的模板文件
    ETCD_SETUP_ARGS_TEMPLATE_FILE="/etc/kubernetes/etcd/etcd-setup-args.template"
    # 检查模板文件是否存在
    if [ ! -f "${ETCD_SETUP_ARGS_TEMPLATE_FILE}" ]; then
        log:error "Etcd arguments template file not found at ${ETCD_SETUP_ARGS_TEMPLATE_FILE}"
        exit 1
    fi
    # 替换模板中的 ##NODE_NAME## 和 ##NODE_IP## 占位符
    GENERATED_ETCD_SETUP_ARGS=$(sed -e "s/##NODE_NAME##/${HOSTNAME}/" -e "s/##NODE_IP##/${IP}/" "${ETCD_SETUP_ARGS_TEMPLATE_FILE}")
    etcd ${GENERATED_ETCD_SETUP_ARGS} & 
    # 简短的等待，确保 Etcd 有时间启动（可选，但对于依赖 Etcd 的服务很有用）
    sleep 5
}

start:master() {
    # 检查 NODE_TYPE 环境变量是否存在，仅启动 NODE_TYPE = master 的服务
    if [ -n "$NODE_TYPE" ] && [ "$NODE_TYPE" = "master" ]; then
        start:kube-apiserver
        start:kube-controller-manager
        start:kube-scheduler
        log:info "Master 节点服务启动完成"
    fi
}

start:worker() {
    start:containerd
}

start:kube-apiserver() {
    # kube-apiserver 启动参数模板文件
    KUBE_APISERVER_SETUP_ARGS_TEMPLATE_FILE="/etc/kubernetes/kube-apiserver-setup-args.template"
    # 检查模板文件是否存在
    if [ ! -f "${KUBE_APISERVER_SETUP_ARGS_TEMPLATE_FILE}" ]; then
        log:error "kube-apiserver arguments template file not found at ${KUBE_APISERVER_SETUP_ARGS_TEMPLATE_FILE}"
        exit 1
    fi
    # 替换模板中的 ##NODE_IP## 占位符
    GENERATED_KUBE_APISERVER_SETUP_ARGS=$(sed -e "s/##NODE_IP##/${IP}/" "${KUBE_APISERVER_SETUP_ARGS_TEMPLATE_FILE}")
    # 启动 kube-apiserver
    kube-apiserver ${GENERATED_KUBE_APISERVER_SETUP_ARGS} &
}

start:kube-controller-manager() {
    # kube-controller-manager 启动参数模板文件
    KUBE_CONTROLLER_MANAGER_SETUP_ARGS_TEMPLATE_FILE="/etc/kubernetes/kube-controller-manager-setup-args.template"
    # 检查模板文件是否存在
    if [ ! -f "${KUBE_CONTROLLER_MANAGER_SETUP_ARGS_TEMPLATE_FILE}" ]; then
        log:error "kube-controller-manager arguments template file not found at ${KUBE_CONTROLLER_MANAGER_SETUP_ARGS_TEMPLATE_FILE}"
        exit 1
    fi
    # 替换模板中的 ##NODE_IP## 占位符
    GENERATED_KUBE_CONTROLLER_MANAGER_SETUP_ARGS=$(sed -e "s/##NODE_IP##/${IP}/" "${KUBE_CONTROLLER_MANAGER_SETUP_ARGS_TEMPLATE_FILE}")
    # 启动 kube-controller-manager
    kube-controller-manager ${GENERATED_KUBE_CONTROLLER_MANAGER_SETUP_ARGS} &
}

start:kube-scheduler() {
    # kube-scheduler 启动参数模板文件
    KUBE_SCHEDULER_SETUP_ARGS_TEMPLATE_FILE="/etc/kubernetes/kube-scheduler-setup-args.template"
    # 检查模板文件是否存在
    if [ ! -f "${KUBE_SCHEDULER_SETUP_ARGS_TEMPLATE_FILE}" ]; then
        log:error "kube-scheduler arguments template file not found at ${KUBE_SCHEDULER_SETUP_ARGS_TEMPLATE_FILE}"
        exit 1
    fi
    # 替换模板中的 ##NODE_IP## 占位符
    GENERATED_KUBE_SCHEDULER_SETUP_ARGS=$(sed -e "s/##NODE_IP##/${IP}/" "${KUBE_SCHEDULER_SETUP_ARGS_TEMPLATE_FILE}")
    # 启动 kube-scheduler
    kube-scheduler ${GENERATED_KUBE_SCHEDULER_SETUP_ARGS} &
}

start:containerd() {
    CONTAINERD_CONFIG_FILE="/etc/kubernetes/containerd/config.toml"
    # 检查配置文件是否存在
    if [ ! -f "${CONTAINERD_CONFIG_FILE}" ]; then
        log:error "containerd config file not found at ${CONTAINERD_CONFIG_FILE}"
        exit 1
    fi
    # 启动 containerd 服务
    containerd --config ${CONTAINERD_CONFIG_FILE} &
}
