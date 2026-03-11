document.addEventListener('DOMContentLoaded', () => {
    const itemInput = document.getElementById('itemInput');
    const addBtn = document.getElementById('addBtn');
    const itemList = document.getElementById('itemList');

    // 初始化加载列表
    fetchItems();

    // 绑定添加事件
    addBtn.addEventListener('click', addItem);
    itemInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            addItem();
        }
    });

    // 获取并渲染列表
    async function fetchItems() {
        try {
            const response = await fetch('/api/items');
            if (!response.ok) throw new Error('网络请求失败');
            const items = await response.json();
            renderList(items);
        } catch (error) {
            console.error('获取列表失败:', error);
        }
    }

    // 添加记录
    async function addItem() {
        const content = itemInput.value.trim();
        if (!content) return;

        // UI 状态反馈
        const originalText = addBtn.textContent;
        addBtn.textContent = '添加中...';
        addBtn.disabled = true;

        try {
            const response = await fetch('/api/items', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ content })
            });

            if (!response.ok) throw new Error('添加失败');
            
            // 清空输入并重新加载列表
            itemInput.value = '';
            await fetchItems();
        } catch (error) {
            console.error('添加记录失败:', error);
            alert('添加失败，请检查网络和后端服务');
        } finally {
            addBtn.textContent = originalText;
            addBtn.disabled = false;
        }
    }

    // 删除记录
    window.deleteItem = async function(id) {
        if (!confirm('确定要删除这条记录吗？')) return;

        try {
            const response = await fetch(`/api/items/${id}`, {
                method: 'DELETE'
            });

            if (!response.ok) throw new Error('删除失败');
            
            // 重新加载列表
            await fetchItems();
        } catch (error) {
            console.error('删除记录失败:', error);
            alert('删除失败');
        }
    }

    // 渲染列表 DOM
    function renderList(items) {
        itemList.innerHTML = '';
        if (!items || items.length === 0) {
            itemList.innerHTML = '<li style="justify-content:center;color:#999;">暂无数据</li>';
            return;
        }

        items.forEach(item => {
            const li = document.createElement('li');
            
            const contentSpan = document.createElement('span');
            contentSpan.className = 'item-content';
            contentSpan.textContent = item.content;

            const deleteBtn = document.createElement('button');
            deleteBtn.className = 'delete-btn';
            deleteBtn.textContent = '删除';
            deleteBtn.onclick = () => window.deleteItem(item.id);

            li.appendChild(contentSpan);
            li.appendChild(deleteBtn);
            itemList.appendChild(li);
        });
    }
});
