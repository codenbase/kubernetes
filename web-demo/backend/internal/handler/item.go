package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/codenbase/kubernetes/web-demo/internal/biz"
)

// ItemHandler 控制器
type ItemHandler struct {
	biz *biz.ItemBiz
}

func NewItemHandler(biz *biz.ItemBiz) *ItemHandler {
	return &ItemHandler{biz: biz}
}

// 辅助方法：返回 JSON 响应
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

// 辅助方法：返回错误 JSON
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// ListItems 处理 GET /api/items
func (h *ItemHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	items, err := h.biz.GetItems(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	respondJSON(w, http.StatusOK, items)
}

// CreateItem 处理 POST /api/items
func (h *ItemHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if payload.Content == "" {
		respondError(w, http.StatusBadRequest, "Content cannot be empty")
		return
	}

	item, err := h.biz.CreateItem(r.Context(), payload.Content)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, item)
}

// DeleteItem 处理 DELETE /api/items/{id}
func (h *ItemHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	// 从 PathValue 获取参数 (Go 1.22+)
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid ID parameter")
		return
	}

	if err := h.biz.DeleteItem(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
