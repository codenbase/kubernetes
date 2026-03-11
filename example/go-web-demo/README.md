# Kubernetes Basics & Go Web Demo

这是一个前后端分离的极简 CRUD Web 应用程序，主要用于演示在 Docker 和 Kubernetes 中的服务部署及编排。

## 架构

* **前端**: 纯 HTML/CSS/Vanilla JS 页面，打包在 Nginx 中提供静态资源服务，同时 Nginx 支持反向代理 `/api` 到后端服务。
* **后端**: Go 1.25.1 原生 `net/http` 服务（使用 1.22+ 增强路由），利用 `github.com/jackc/pgx/v5` 存取数据库。
* **数据库**: PostgreSQL 15，存储 `items` 清单。

```text
外部请求 -> K8s Ingress (demo.local) 
   -> /    -> Frontend (Nginx 静态文件)
   -> /api -> Backend (Go 服务)
                 -> Postgres 数据库
```

## 运行与部署

### 方式一：本地 Docker Compose 运行

直接在项目根目录下运行：

```bash
docker-compose up --build -d
```

等待构建完成后，在浏览器访问 **[http://localhost:8080](http://localhost:8080)** 即可看到完整交互页面。（注：前端 Nginx 容器将内部的 `80` 端口映射到了外部的 `8080`）

### 方式二：Kubernetes 部署

1. **构建镜像并加载到本地集群 (以 minikube 为例)**

   ```bash
   docker build -t go-web-demo-backend:latest ./backend
   docker build -t go-web-demo-frontend:latest ./frontend
   minikube image load go-web-demo-backend:latest
   minikube image load go-web-demo-frontend:latest
   ```

2. **应用 K8s 清单配置**

   ```bash
   kubectl apply -f k8s/postgres.yaml
   kubectl apply -f k8s/backend.yaml
   kubectl apply -f k8s/frontend.yaml
   kubectl apply -f k8s/ingress.yaml
   ```

3. **设置 Hosts (假设你使用 Nginx Ingress Controller)**

   获取集群 Ingress IP：`kubectl get ingress`
   然后在系统 `/etc/hosts` 中加入一条映射：
   `<YOUR_CLUSTER_IP> demo.local`

4. **访问服务**

   在浏览器中访问 **[http://demo.local](http://demo.local)**
