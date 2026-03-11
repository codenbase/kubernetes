package biz

import (
	"context"

	"go-web-demo/internal/store"
)

// ItemBiz 处理业务逻辑
type ItemBiz struct {
	store *store.ItemStore
}

func NewItemBiz(store *store.ItemStore) *ItemBiz {
	return &ItemBiz{store: store}
}

func (b *ItemBiz) GetItems(ctx context.Context) ([]store.Item, error) {
	return b.store.List(ctx)
}

func (b *ItemBiz) CreateItem(ctx context.Context, content string) (store.Item, error) {
	return b.store.Create(ctx, content)
}

func (b *ItemBiz) DeleteItem(ctx context.Context, id int) error {
	return b.store.Delete(ctx, id)
}
