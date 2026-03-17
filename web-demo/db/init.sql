-- 创建 items 表
CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    content VARCHAR(255) NOT NULL
);

-- 插入初始测试数据
INSERT INTO items (content) VALUES ('学习 Kubernetes 基础概念');
INSERT INTO items (content) VALUES ('部署第一个 Go 语言 Web 极简容器应用');
