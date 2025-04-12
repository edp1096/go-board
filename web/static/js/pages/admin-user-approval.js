// web/static/js/pages/admin-user-approval.js
// 사용자 승인 처리 함수
function approveUser(userId) {
    if (!confirm('이 사용자를 승인하시겠습니까?')) {
        return;
    }

    // CSRF 토큰 가져오기
    const csrfToken = document.querySelector('meta[name="csrf-token"]').content;

    fetch(`/admin/users/${userId}/approve`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
            'X-CSRF-Token': csrfToken,
        }
    })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                // 성공 시 해당 행 제거
                const userRow = document.getElementById(`user-row-${userId}`);
                if (userRow) {
                    userRow.remove();
                }

                // 테이블 내용 확인 후 메시지 표시
                checkEmptyTable();

                alert('사용자가 성공적으로 승인되었습니다.');
            } else {
                alert('오류가 발생했습니다: ' + data.message);
            }
        })
        .catch(error => {
            console.error('Error:', error);
            alert('요청 처리 중 오류가 발생했습니다.');
        });
}

// 사용자 승인 거부 함수
function rejectUser(userId) {
    if (!confirm('이 사용자의 승인을 거부하시겠습니까? 이 작업은 되돌릴 수 없습니다.')) {
        return;
    }

    // CSRF 토큰 가져오기
    const csrfToken = document.querySelector('meta[name="csrf-token"]').content;

    fetch(`/admin/users/${userId}/reject`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
            'X-CSRF-Token': csrfToken,
        }
    })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                // 성공 시 해당 행 제거
                const userRow = document.getElementById(`user-row-${userId}`);
                if (userRow) {
                    userRow.remove();
                }

                // 테이블 내용 확인 후 메시지 표시
                checkEmptyTable();

                alert('사용자 승인이 거부되었습니다.');
            } else {
                alert('오류가 발생했습니다: ' + data.message);
            }
        })
        .catch(error => {
            console.error('Error:', error);
            alert('요청 처리 중 오류가 발생했습니다.');
        });
}

// 테이블이 비어있는지 확인하고 메시지 표시
function checkEmptyTable() {
    const tableBody = document.querySelector('tbody');
    const rows = tableBody.querySelectorAll('tr');

    if (rows.length === 0) {
        // 모든 행이 제거된 경우 메시지 표시
        const emptyRow = document.createElement('tr');
        emptyRow.innerHTML = `
            <td colspan="5" class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-center">
                승인 대기 중인 사용자가 없습니다.
            </td>
        `;
        tableBody.appendChild(emptyRow);
    }
}