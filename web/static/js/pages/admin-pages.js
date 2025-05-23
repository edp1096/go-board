// web/static/js/pages/admin-pages.js
function pageManagement() {
    return {
        csrf: document.querySelector('meta[name="csrf-token"]')?.content || '',
        submitting: false, // submitting 변수 추가

        deletePage(id) {
            if (!confirm('정말 페이지를 삭제하시겠습니까? 이 작업은 되돌릴 수 없습니다.')) return;

            fetch(`/admin/pages/${id}`, {
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
                                window.location.href = '/admin/pages';
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
        }
    };
}