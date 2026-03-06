# Monolith Stress Lab

这是一个极简的单体架构（Web 应用与数据库同节点部署）压力测试沙盒项目。其核心目标是在资源受限的环境下（例如 4C4G），测试并监控 Go 应用服务器与底层数据库在面对高并发、高 CPU 消耗（如 `bcrypt` 密码验证）场景时的极限性能。

## 🏗️ 架构概览

- **业务逻辑层**: Go 1.25, Gin Web 框架, JWT 鉴权。
- **持久化层**: PostgreSQL 16, GORM。
- **压测工具**: K6 (运行在独立的辅助容器中)。
- **部署方式**: Docker Compose。应用与数据库被打包运行在同一个名为 `stress-app` 的容器内模拟单台物理机，容器资源被严格限制为 4 核 CPU 与 4GB 内存。

## 🚀 快速部署与测试指南

### 1. 构建与启动环境

在项目根目录下执行以下命令：

```bash
docker-compose build --no-cache app
docker-compose up -d
```
> 这将编译 Go 代码并把 Postgres 数据库打包运行，全过程可能需要 1~2 分钟来初始化数据库。

### 2. 验证运行状态

确保 `stress-app` 和 `k6-loadtest` 容器均处于 Up 状态：
```bash
docker-compose ps
```

您可以查看应用日志确保其已经成功监听 8080 端口且没有报错：
```bash
docker logs -f stress-app
```

### 3. 一键发起压力测试

我们在项目内提供了一份默认的 K6 压测脚本 (`loadtest/login_test.js`)，它将模拟 50 个并发用户持续 30 秒进行 `/login` 登录接口的轰炸。

直接利用后台已经部署好的 `k6-loadtest` 容器执行压测：
```bash
docker exec -it k6-loadtest k6 run /loadtest/login_test.js
```

**如何看懂 K6 的压测指标？**
压测结束后，控制台会输出一张报表，重点关注：
- `http_reqs`: 显示总请求数以及最重要的 **QPS（每秒请求数）**。
- `http_req_duration`: 显示接口耗时。关注 `avg` (平均延迟，反映整体健康度) 和 `p(99)` (99%的尾部延迟，反映极端卡顿用户的体验)。

---

## 🛠️ 原生 Linux 性能瓶颈监控指南

压测的精髓不在于看最终数字，而在于**在压测进行中寻找系统的瓶颈**。我们在 `stress-app` 容器中预装了原生的 Linux 诊断工具。

请打开一个**新终端窗口**，进入目标服务器的 Bash 环境：
```bash
docker exec -it stress-app bash
```

在另一个窗口启动 K6 压测，同时在这个 Bash 中使用以下工具观察：

### 1. 全局资源概览 (`top`)
最经典的命令，用于查看系统整体 CPU 和内存的损耗：
```bash
top
```
- **核心观测点**:
  - `%Cpu(s) - us`: 用户态 CPU 占用。如果它接近 100%，说明代码逻辑（例如海量的 `bcrypt` 哈希计算）吃光了算力。
  - `MiB Mem`: 关注可用内存 (`free`) 是否持续下降，判断是否有严重的内存泄漏。
  - 💡 **进阶用法**：在 `top` 界面中按键盘 `1` 键，可以展开查看 4 个限制核心中每一个的负载分配是否均衡。

### 2. 精确监控进程层面的阻力 (`pidstat`)
`top` 只能看大盘，而 `pidstat` 能够精确拆解到底是 Go Web 程序（`main`）成了瓶颈，还是 PostgreSQL 数据库（`postgres`）被压趴下了。

**监控各个进程的 CPU 抢占（每 2 秒刷新一次）：**
```bash
pidstat -u 2
```
查看 `%CPU` 和 `Command` 列的对应关系。

**监控各个进程的内存占用情况：**
```bash
pidstat -r 2
```
关注 `%MEM` 指标的变化。

### 3. 磁盘 I/O 阻塞排查 (`iostat`)
随着并发上升，PostgreSQL 的刷盘和查询可能会遇到极其严重的磁盘读写瓶颈。

**监控磁盘详细 I/O 指标（每 2 秒刷新一次，显示详细列）：**
```bash
iostat -x 2
```
- **核心观测点**:
  - `%util`: 磁盘利用率。如果这个值长期维持在 `80% - 100%`，说明磁盘已经满载，IO 是最大的瓶颈。
  - `await`: 每个 I/O 请求的平均等待时间（毫秒）。过高的数值意味着 SQL 读写请求正在排队枯等。

---

## 🧹 关于环境的重置与清理

如果由于测试需要，您修改了数据库构建相关的代码（例如 `db/init.sql` 中的表结构，或者修改了其内部密码签发的 Bcrypt Cost 重试极限测试），请**务必注意**：简单的 `docker-compose restart` 或者 `build` 是不够的，因为旧的数据卷依然挂载着。

**正确的环境彻底销毁与重建步骤为：**
```bash
# 连同数据库持久化卷一起销毁
docker-compose down -v

# 重新构建与启动
docker-compose build app
docker-compose up -d
```
