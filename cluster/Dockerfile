# Dockerfile

# 基于 Debian 12 镜像
FROM debian:12

# 维护者信息
LABEL maintainer="Your Name <your.email@example.com>"

# 设置非交互模式，避免 apt 弹出确认提示
ENV DEBIAN_FRONTEND=noninteractive

# 1. 更新系统并安装所有必要的软件包
# 包括：ping, ssh, curl, wget, vim, socat, conntrack, ipset, apt-transport-https, ca-certificates
RUN apt update && \
    apt upgrade -y && \
    apt install -y \
        iputils-ping \
        openssh-server \
        openssh-client \
        curl \
        wget \
        vim \
        socat \
        conntrack \
        ipset \
        apt-transport-https \
        ca-certificates \
        gnupg \
        lsb-release && \
    rm -rf /var/lib/apt/lists/* # 清理 apt 缓存，减小镜像大小

# 2. 禁用 Swap (如果 /etc/fstab 中有 swap 条目则注释掉)
# 对于容器通常不是必需，但为了完整性保留，实际禁用 swap 会在 entrypoint.sh 中执行
RUN sed -i '/ swap / s/^\(.*\)$/#\1/g' /etc/fstab

# 3. 配置 Kubernetes 要求的内核参数 (sysctl)
# 永久加载 br_netfilter 和 overlay 模块 (仅写入配置文件，实际加载在 entrypoint.sh 中执行)
RUN echo "br_netfilter" > /etc/modules-load.d/k8s.conf && \
    echo "overlay" >> /etc/modules-load.d/k8s.conf
# 配置 sysctl 参数，实现 IP 转发和桥接流量通过 iptables (仅写入配置文件，实际应用在 entrypoint.sh 中执行)
RUN echo "net.bridge.bridge-nf-call-iptables = 1" > /etc/sysctl.d/k8s.conf && \
    echo "net.bridge.bridge-nf-call-ip6tables = 1" >> /etc/sysctl.d/k8s.conf && \
    echo "net.ipv4.ip_forward = 1" >> /etc/sysctl.d/k8s.conf

# 4. 安装 Containerd
# 添加 Docker 官方 GPG 密钥
RUN install -m 0755 -d /etc/apt/keyrings && \
    curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg && \
    chmod a+r /etc/apt/keyrings/docker.gpg

# 添加 Docker APT 仓库
RUN echo \
  "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian \
  "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" | \
  tee /etc/apt/sources.list.d/docker.list > /dev/null

# 安装 containerd.io
RUN apt update && \
    apt install -y containerd.io

# 5. 配置 Containerd
# 生成默认配置，并确保 systemd cgroup 驱动被启用
RUN mkdir -p /etc/containerd && \
    containerd config default > /etc/containerd/config.toml && \
    sed -i 's/SystemdCgroup = false/SystemdCgroup = true/g' /etc/containerd/config.toml

# 6. 配置 SSH
# 允许 root 用户 SSH 登录，并禁用PAM，避免密码验证问题（仅限学习环境，生产环境禁用）
RUN sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/g' /etc/ssh/sshd_config && \
    sed -i 's/UsePAM yes/UsePAM no/g' /etc/ssh/sshd_config
# 设置root密码为root，方便调试，请在实际使用中修改或移除！
RUN echo "root:root" | chpasswd

# 暴露 SSH 端口 (在容器内，虽然实际通过 --network 内部通信，但声明更清晰)
EXPOSE 22

# 7. 定义容器的入口点和默认命令
# 将 entrypoint.sh 脚本复制到容器内，并赋予执行权限
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

# ENTRYPOINT 总是执行 entrypoint.sh 脚本，它是容器启动后第一个且唯一执行的程序
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

# CMD 为 ENTRYPOINT 提供默认参数。这里只提供一个默认的 bash shell，
# 这样如果 docker run 命令没有指定其他参数，容器会默认启动一个交互式 shell。
CMD ["bash"]
