# Monolith Stress Lab

本仓库是用于极其严苛条件下的高并发单机性能压测实验室。被测服务当前主要运行在一台 4C4G 的远程 AWS EC2 实例上。
本篇文档为 **压测人员专属使用手册**，包含了目前实验室支持的所有压测场景、执行命令以及可调参数。

## � 压测总指挥 (施压端执行)

请在配备了 K6 环境的施压服务器上执行以下命令。所有的压测脚本都在 `loadtest/` 目录下。

### 核心通用参数
所有的压测脚本都支持以下环境变量进行“战前换挡”：
- `-e BASE_URL=http://x.x.x.x:8080`: 被测应用服务器的内网/外网地址。
- `-e VUS=200`: 并发虚拟用户数 (Virtual Users)，决定火力大小。
- `-e DURATION=30s`: 持续火力时间 (如 `30s`, `1m`)。

---

### 实验一：登录鉴权与 CPU 极限消耗 (`login_test.js`)
**场景**：模拟用户并发登录。后端需要消耗大量的 CPU 算力来计算 Bcrypt 密码哈希。
**特有参数**：
- `-e COST=10`: [动态特权参数] K6 会在压测开始前，自动调用服务端的上帝接口，强行把全库几千名验证用户的密码计算成本篡改为指定值 (例如 4 是极速，12 是地狱难度)。

**压测命令**：
```bash
sudo docker run --rm -it -v /home/ubuntu/loadtest:/loadtest \
  -e VUS=200 -e DURATION=30s -e COST=10 \
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
