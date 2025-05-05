// web/static/js/pages/admin-categories.js
function categoryManagement() {
    return {
        csrf: document.querySelector('meta[name="csrf-token"]')?.content || '',
        boardOptions: [],
        pageOptions: [],
        submitting: false, // submitting 변수 추가

        init() {
            this.loadBoardOptions();
            this.loadPageOptions();
        },

        loadBoardOptions() {
            fetch('/api/admin/boards?active=true', {
                headers: {
                    'Accept': 'application/json'
                }
            })
                .then(res => res.json())
                .then(data => {
                    if (data.success) {
                        this.boardOptions = data.boards || [];
                    }
                })
                .catch(err => {
                    console.error('게시판 목록 로딩 실패:', err);
                });
        },

        loadPageOptions() {
            fetch('/api/admin/pages?active=true', {
                headers: {
                    'Accept': 'application/json'
                }
            })
                .then(res => res.json())
                .then(data => {
                    if (data.success) {
                        this.pageOptions = data.pages || [];
                    }
                })
                .catch(err => {
                    console.error('페이지 목록 로딩 실패:', err);
                });
        },

        deleteCategory(id) {
            if (!confirm('정말 카테고리를 삭제하시겠습니까? 이 작업은 되돌릴 수 없습니다.')) return;

            fetch(`/admin/categories/${id}`, {
                method: 'DELETE',
                headers: {
                    'Accept': 'application/json',
                    'X-CSRF-Token': this.csrf,
                }
            })
                .then(res => {
                    const contentType = res.headers.get('Content-Type');
                    if (contentType && contentType.includes('application/json')) {
                        return res.json().then(data => {
                            if (data.success) {
                                window.location.href = '/admin/categories';
                            } else {
                                alert(data.message);
                            }
                        });
                    } else {
                        return res.text().then(html => {
                            console.error('서버에서 HTML 응답이 반환되었습니다', html);
                            alert('처리 중 오류가 발생했습니다. 다시 시도해 주세요.');
                        });
                    }
                })
                .catch(err => {
                    console.error('요청 처리 중 오류:', err);
                    alert('오류가 발생했습니다: ' + err);
                });
        },

        addBoardItem(categoryId) {
            const boardId = this.$refs.boardSelect.value;
            if (!boardId) {
                alert('게시판을 선택해주세요.');
                return;
            }

            fetch(`/api/admin/categories/${categoryId}/items`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Accept': 'application/json',
                    'X-CSRF-Token': this.csrf,
                },
                body: JSON.stringify({
                    itemId: parseInt(boardId),
                    itemType: 'board'
                })
            })
                .then(res => res.json())
                .then(data => {
                    if (data.success) {
                        window.location.reload();
                    } else {
                        alert(data.message);
                    }
                })
                .catch(err => {
                    console.error('항목 추가 중 오류:', err);
                    alert('오류가 발생했습니다: ' + err);
                });
        },

        addPageItem(categoryId) {
            const pageId = this.$refs.pageSelect.value;
            if (!pageId) {
                alert('페이지를 선택해주세요.');
                return;
            }

            fetch(`/api/admin/categories/${categoryId}/items`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Accept': 'application/json',
                    'X-CSRF-Token': this.csrf,
                },
                body: JSON.stringify({
                    itemId: parseInt(pageId),
                    itemType: 'page'
                })
            })
                .then(res => res.json())
                .then(data => {
                    if (data.success) {
                        window.location.reload();
                    } else {
                        alert(data.message);
                    }
                })
                .catch(err => {
                    console.error('항목 추가 중 오류:', err);
                    alert('오류가 발생했습니다: ' + err);
                });
        },

        removeItem(categoryId, itemId, itemType) {
            if (!confirm('이 항목을 카테고리에서 제거하시겠습니까?')) return;

            fetch(`/api/admin/categories/${categoryId}/items/${itemId}/${itemType}`, {
                method: 'DELETE',
                headers: {
                    'Accept': 'application/json',
                    'X-CSRF-Token': this.csrf,
                }
            })
                .then(res => res.json())
                .then(data => {
                    if (data.success) {
                        window.location.reload();
                    } else {
                        alert(data.message);
                    }
                })
                .catch(err => {
                    console.error('항목 제거 중 오류:', err);
                    alert('오류가 발생했습니다: ' + err);
                });
        }
    };
}