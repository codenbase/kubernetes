# 使用 Debian 12 (bookworm) 的 slim 版本作为基础
FROM debian:12-slim

# 设置环境变量，避免 apt-get 在安装过程中进行交互式提问
ENV DEBIAN_FRONTEND=noninteractive

# 更新软件包列表，并安装 systemd 和 K8s 依赖的核心工具
# 为了减小镜像体积，我们将所有步骤合并在一个 RUN 指令中
# 首先，执行一次完整的系统更新与升级，确保安全补丁和基础软件都处于最新稳定版
RUN apt-get update && apt-get upgrade -y \
    # --- 安装所有软件包 ---
    && apt-get install -y \
        openssh-server openssh-client \
        chrony vim curl wget jq tree sysstat kmod gpg \
        ca-certificates apt-transport-https \
        iputils-ping iproute2 iptables ipset \
        ipvsadm socat conntrack ethtool \
    # --- "改造" systemd，清理不需要的服务 ---
    && find /lib/systemd/system/sysinit.target.wants/ -name "systemd-tmpfiles-setup.service" -delete \
    && rm -f /lib/systemd/system/multi-user.target.wants/* \
    && rm -f /etc/systemd/system/*.wants/* \
    && rm -f /lib/systemd/system/local-fs.target.wants/* \
    && rm -f /lib/systemd/system/sockets.target.wants/*udev* \
    && rm -f /lib/systemd/system/sockets.target.wants/*initctl* \
    && rm -f /lib/systemd/system/basic.target.wants/* \
    && echo "ReadKMsg=no" >> /etc/systemd/journald.conf \
    # # --- 清理临时文件 ---
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    # --- 清理 ssh 主机密钥，保留 sshd_config 等核心配置文件，确保 sshd 服务正常启动
    && rm -rf /etc/ssh/ssh_host_*_key*

# 将我们的入口脚本复制到镜像中
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
# 设置入口点为我们的脚本
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
# 容器启动时默认执行的命令，这个命令会传递给 entrypoint.sh
CMD ["/usr/bin/systemd"]