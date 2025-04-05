function boardManagement() {
    return {
        csrf: '{{$.csrf}}', // CSRF 토큰, 필요에 따라 다른 방법으로 설정할 수 있습니다.
        deleteBoard(id) {
            if (!confirm('정말 삭제하시겠습니까? 관련된 모든 게시물이 삭제됩니다.')) return;

            const csrfToken = document.querySelector('meta[name="csrf-token"]').content;

            fetch(`/admin/boards/${id}`, {
                method: 'DELETE',
                headers: {
                    'Accept': 'application/json',
                    'X-CSRF-Token': csrfToken,
                }
            })
            .then(res => {
                const contentType = res.headers.get('Content-Type');
                if (contentType && contentType.includes('application/json')) {
                    return res.json().then(data => {
                        if (data.success) {
                            window.location.reload();
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
    }
}
