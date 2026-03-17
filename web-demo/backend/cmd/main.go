package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/codenbase/kubernetes/web-demo/internal/biz"
	"github.com/codenbase/kubernetes/web-demo/internal/handler"
	"github.com/codenbase/kubernetes/web-demo/internal/store"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// 从环境变量获取数据库连接串
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// 本地开发回退地址
		dsn = "postgres://postgres:postgres@localhost:5432/demo?sslmode=disable"
	}

	// 建立数据库连接池
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("无法连接数据库: %v", err)
	}
	defer pool.Close()

	// 依赖注入
	itemStore := store.NewItemStore(pool)
	itemBiz := biz.NewItemBiz(itemStore)
	itemHandler := handler.NewItemHandler(itemBiz)

	// 使用 Go 1.22+ 增强的路由匹配功能
	mux := http.NewServeMux()
	
	// API 路由: 带有 /api 前缀
	mux.HandleFunc("GET /api/items", itemHandler.ListItems)
	mux.HandleFunc("POST /api/items", itemHandler.CreateItem)
	mux.HandleFunc("DELETE /api/items/{id}", itemHandler.DeleteItem)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("后端服务已启动，监听端口: %s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), mux); err != nil {
		log.Fatalf("服务运行失败: %v", err)
	}
}
