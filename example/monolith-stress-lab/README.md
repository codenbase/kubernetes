# Monolith Stress Lab

本仓库是用于极其严苛条件下的高并发单机性能压测实验室。被测服务当前主要运行在一台 4C4G 的远程 AWS EC2 实例上。
本篇文档为 **压测人员专属使用手册**，包含了目前实验室支持的所有压测场景、执行命令以及可调参数。

## � 压测总指挥 (施压端执行)

请在配备了 K6 环境的施压服务器上执行以下命令。所有的压测脚本都在 `loadtest/` 目录下。

### 核心通用参数
所有的压测脚本都支持以下环境变量进行“战前换挡”：
- `-e BASE_URL=http://x.x.x.x:8080`: 被测应用服务器的内网/外网地址。
- `-e VUS=200`: 并发虚拟用户数 (Virtual Users)，决定火力大小（用于闭环测试）。
- `-e DURATION=30s`: 持续火力时间 (如 `30s`, `1m`)。
- `-e QPS=20000`: 目标每秒请求数 (用于**开环极端脉冲测试**)，无视服务器响应强行以指定频率发包。

---

### 实验一：登录鉴权与 CPU 极限消耗 (`login_test.js`)
**场景**：模拟用户并发登录。后端需要消耗大量的 CPU 算力来计算 Bcrypt 密码哈希。
**特有参数**：
- `-e COST=10`: [动态特权参数] K6 会在压测开始前，自动调用服务端的上帝接口，强行把全库几千名验证用户的密码计算成本篡改为指定值 (例如 4 是极速，12 是地狱难度)。

**压测命令**：
```bash
sudo docker run --rm -it -v /home/ubuntu/loadtest:/loadtest \
  -e VUS=200 -e DURATION=30s -e COST=4 \
  -e BASE_URL=http://172.31.44.56:8080 \
  grafana/k6 run /loadtest/login_test.js
```

---

### 实验二：数据库小包高频直读 (`article_small_test.js`)
**场景**：绕过缓存直达数据库，随机读取 100 字节左右的小文章。测试 Postgres 数据库高频查库。
**特有参数**：
- `-e DB_MAX_OPEN=50`: [动态特权参数] 控制被测服务器的数据库最大连接池。

**压测命令**：
```bash
sudo docker run --rm -it -v /home/ubuntu/loadtest:/loadtest \
  -e VUS=200 -e DURATION=30s -e DB_MAX_OPEN=50 \
  -e BASE_URL=http://172.31.44.56:8080 \
  grafana/k6 run /loadtest/article_small_test.js
```

---

### 实验三：数据库大包高频网传 (`article_large_test.js`)
**场景**：直达数据库，随机读取超过 50KB 的超长文本文章。测试 Go 语言的内存大对象 GC 切片开销与磁盘带阅读带宽极限。
**特有参数**：
- `-e DB_MAX_OPEN=50`: [动态特权参数] 控制被测服务器的数据库最大连接池。

**压测命令**：
```bash
sudo docker run --rm -it -v /home/ubuntu/loadtest:/loadtest \
  -e VUS=200 -e DURATION=30s -e DB_MAX_OPEN=50 \
  -e BASE_URL=http://172.31.44.56:8080 \
  grafana/k6 run /loadtest/article_large_test.js
```

---

### 实验四：高频写盘与普通索引裂变 (`comment_test.js`)
**场景**：模拟无解缓冲下的海量用户并发写评论，触发巨大的 Postgres 树形索引更新与磁盘同步刷盘 I/O 等待。
**特有参数**：
- `-e DB_MAX_OPEN=15`: [动态特权参数] K6 会在压测开始前，自动调取服务端上帝接口，把 Go 程序的底层 Postgres 连接池强制缩容或扩容到指定大小，让您感受“线程排队拥堵”引发的惨案。

**压测命令**：
```bash
sudo docker run --rm -it -v /home/ubuntu/loadtest:/loadtest \
  -e VUS=200 -e DURATION=30s -e DB_MAX_OPEN=15 \
  -e BASE_URL=http://172.31.44.56:8080 \
  grafana/k6 run /loadtest/comment_test.js
```

---

### 实验五：混合场景容量探测 (`mixed_workload_test.js`)
**场景**：不再是单一的尖峰负载，而是模拟互联网典型的真实流量漏斗模型（80%浏览、10%活跃写库、10%高计算登录），并采用典型的“**步增模式 (Ramping) 寻找全局拐点**”：
- `0 -> 30s`：从 0 并发缓慢爬升到轻量负载（100 VUs）
- `30s -> 1.5m`：稳步爬升到正常高峰（300 VUs）
- `1.5m -> 2.5m`：暴力拉升测试容量天花板（极限 600 VUs）
- 本脚本包含了 `setup()` 钩子，支持同时下发并发池与算力成本指标。

**特有参数**：
- 此脚本使用自定义内部 `scenarios` 定义了并发规模，**不需要**手动传递 `VUS` 和 `DURATION` 参数。

**压测命令**：
```bash
sudo docker run --rm -it -v /home/ubuntu/loadtest:/loadtest \
  -e COST=4 -e DB_MAX_OPEN=50 \
  -e BASE_URL=http://172.31.44.56:8080 \
  grafana/k6 run /loadtest/mixed_workload_test.js
```

---

### 实验六：开环极端脉冲测试 (`openloop_spike_test.js`)
**场景**：不再是“善意”的闭环测试，而是真正的互联网大考！通过 K6 真正的底层大杀器 `constant-arrival-rate` 执行器，该脚本将“无视被测服务器是否卡死或者超时”，以冷酷无情的**恒力机枪模式**向目标发送固定 QPS 的流量。即使目标服务器的网络栈已经开始丢包报错，这边的施压机依然会生生不息地创建新连接强行灌入。
- **流量漏斗**：依然保持 80% 阅读、10% 评论、10% 登录的真实配比。
- 适合用来观测 Linux 宿主机的网络积压队列 `SYN_RECV`，以及引发各种生动绝望的内核层 Socket Error。

**特有参数**：
- `-e QPS=20000`: 决定每秒钟施加给服务器的真实并发吞吐速率（例如 20000 表示这一秒必须发出 2 万个请求）。

**压测命令**：
```bash
sudo docker run --rm -it -v /home/ubuntu/loadtest:/loadtest \
  -e QPS=20000 -e DURATION=10s \
  -e BASE_URL=http://172.31.44.56:8080 \
  grafana/k6 run /loadtest/openloop_spike_test.js
```
