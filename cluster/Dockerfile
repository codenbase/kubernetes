# Dockerfile

# 基于 Debian 12 镜像
FROM debian:12
FROM kindest/node:v1.27.1

# 维护者信息
LABEL maintainer="Your Name <your.email@example.com>"

# 设置非交互模式，避免 apt 弹出确认提示
ENV DEBIAN_FRONTEND=noninteractive

# 更新系统并安装所有必要的软件包
# 包括: ping, ssh, jq, chrony, curl, wget, vim, socat, conntrack, ipset, ipvsadm, sysstat
RUN apt update && \
    apt upgrade -y && \
    apt install -y \
        iputils-ping \
        openssh-server \
        openssh-client \
        jq \
        chrony \
        curl \
        wget \
        vim \
        socat \
        conntrack \
        ipset \
        ipvsadm \
        haveged \
        sysstat && \
    # 删除自动生成的 ssh 配置和密钥
    rm -rf /etc/ssh/* && \
    # 清理 apt 缓存，减小镜像大小
    rm -rf /var/lib/apt/lists/*

# 配置 Kubernetes 要求的内核参数 (sysctl)
# 永久加载 br_netfilter 和 overlay 模块 (仅写入配置文件)
RUN echo "br_netfilter" > /etc/modules-load.d/k8s.conf && \
    echo "overlay" >> /etc/modules-load.d/k8s.conf
# 配置 sysctl 参数，实现 IP 转发和桥接流量通过 iptables (仅写入配置文件，在 entrypoint.sh 中执行 `sysctl --system` 应用配置)
RUN echo "net.bridge.bridge-nf-call-iptables = 1" > /etc/sysctl.d/k8s.conf && \
    echo "net.bridge.bridge-nf-call-ip6tables = 1" >> /etc/sysctl.d/k8s.conf && \
    echo "net.ipv4.ip_forward = 1" >> /etc/sysctl.d/k8s.conf && \
    echo "net.ipv4.tcp_tw_recycle=0" >> /etc/sysctl.d/k8s.conf

# 设置系统时区
RUN echo "Asia/Shanghai" > /etc/timezone && \
    dpkg-reconfigure -f noninteractive tzdata

# 拷贝二进制文件
COPY bin/* /usr/local/bin/

# 拷贝 containerd 安装文件
RUN mkdir -p /opt/cni/bin
COPY containerd/cni-plugins/* /opt/cni/bin/
COPY containerd/bin/* containerd/crictl /usr/local/bin/
COPY containerd/runc.amd64 /usr/local/sbin/runc
COPY containerd/crictl.yaml /etc/crictl.yaml
RUN chmod +x /usr/local/sbin/runc

RUN helm repo add cilium https://helm.cilium.io/
RUN helm repo update

# 设置 PATH
ENV PATH="/opt/cni/bin:$PATH"

# 拷贝 kubectl 配置文件
COPY kubectl.kubeconfig /root/.kube/config

# 设置root密码为root，方便调试，请在实际使用中修改或移除！
RUN echo "root:root" | chpasswd
# 暴露 SSH 端口
EXPOSE 22

# 定义容器的入口点和默认命令
COPY entrypoint.sh service.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/entrypoint.sh /usr/local/bin/service.sh
# ENTRYPOINT 总是执行 entrypoint.sh 脚本，它是容器启动后第一个且唯一执行的程序
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
# CMD 为 ENTRYPOINT 提供默认参数。这里只提供一个默认的 bash shell，
# 这样如果 docker run 命令没有指定其他参数，容器会默认启动一个交互式 shell。
CMD ["bash"]
