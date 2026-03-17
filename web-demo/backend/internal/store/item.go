package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Item 定义数据库实体
type Item struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
}

// ItemStore 定义数据库操作接口及实现
type ItemStore struct {
	db *pgxpool.Pool
}

func NewItemStore(db *pgxpool.Pool) *ItemStore {
	return &ItemStore{db: db}
}

// List 获取所有 item
func (s *ItemStore) List(ctx context.Context) ([]Item, error) {
	rows, err := s.db.Query(ctx, "SELECT id, content FROM items ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 初始化为了避免 JSON 序列化为 null
	items := make([]Item, 0)
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Content); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// Create 插入一条新记录
func (s *ItemStore) Create(ctx context.Context, content string) (Item, error) {
	var item Item
	err := s.db.QueryRow(ctx, "INSERT INTO items (content) VALUES ($1) RETURNING id, content", content).Scan(&item.ID, &item.Content)
	return item, err
}

// Delete 删除指定记录
func (s *ItemStore) Delete(ctx context.Context, id int) error {
	_, err := s.db.Exec(ctx, "DELETE FROM items WHERE id = $1", id)
	return err
}
