# Makefile 用于本地快速构建和推送到 GHCR

# 配置相关变量
GITHUB_OWNER ?= codenbase
REGISTRY := ghcr.io
IMAGE_PREFIX := $(REGISTRY)/$(GITHUB_OWNER)
TAG ?= latest

BACKEND_IMAGE := $(IMAGE_PREFIX)/go-web-demo-backend:$(TAG)
FRONTEND_IMAGE := $(IMAGE_PREFIX)/go-web-demo-frontend:$(TAG)

.PHONY: all build-backend build-frontend push-backend push-frontend push login update-k8s

all: build push update-k8s

build: build-backend build-frontend

build-backend:
	docker build -t $(BACKEND_IMAGE) ./backend

build-frontend:
	docker build -t $(FRONTEND_IMAGE) ./frontend

push-backend: build-backend
	docker push $(BACKEND_IMAGE)

push-frontend: build-frontend
	docker push $(FRONTEND_IMAGE)

push: push-backend push-frontend

# 本地登录 GHCR (需要提供你的 GitHub PAT(Personal Access Token)，且该 Token 需要有 write:packages 权限)
login:
	@echo "请执行以下命令登录 GHCR："
	@echo "export CR_PAT=YOUR_TOKEN"
	@echo "echo \$$CR_PAT | docker login ghcr.io -u $(GITHUB_OWNER) --password-stdin"

# 将 K8s manifests 里的镜像统一改为 GHCR 的地址
update-k8s:
	sed -i.bak 's|image: go-web-demo-backend:latest|image: $(BACKEND_IMAGE)|g' k8s/backend.yaml
	sed -i.bak 's|image: go-web-demo-frontend:latest|image: $(FRONTEND_IMAGE)|g' k8s/frontend.yaml
	rm -f k8s/*.bak
	@echo "K8s manifests 已更新！现在它们使用的是 $(IMAGE_PREFIX) 下的镜像。"
