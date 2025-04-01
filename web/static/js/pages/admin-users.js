// 사용자 역할 전환 함수
function toggleUserRole(userId, currentRole) {
    if (!confirm('이 사용자의 역할을 변경하시겠습니까?')) {
        return;
    }

    const userIdInt = parseInt(userId); // 변수명 변경 주의

    const newRole = currentRole === 'admin' ? 'user' : 'admin';
    
    fetch(`/admin/users/${userIdInt}/role`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
            'X-CSRF-Token': '{{.csrf}}',
        },
        body: JSON.stringify({ role: newRole }),
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            window.location.reload();
        } else {
            alert('오류가 발생했습니다: ' + data.message);
        }
    })
    .catch(error => {
        console.error('Error:', error);
        alert('요청 처리 중 오류가 발생했습니다.');
    });
}

// 사용자 상태 전환 함수
function toggleUserStatus(userId, currentStatus) {
    if (!confirm(currentStatus ? '이 사용자의 계정을 비활성화하시겠습니까?' : '이 사용자의 계정을 활성화하시겠습니까?')) {
        return;
    }

    const userIdInt = parseInt(userId); // 변수명 변경 주의
    const parsedStatus = parseBool(currentStatus);

    const newStatus = !parsedStatus;
    
    fetch(`/admin/users/${userIdInt}/status`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
            'X-CSRF-Token': '{{.csrf}}',
        },
        body: JSON.stringify({ active: newStatus }),
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            window.location.reload();
        } else {
            alert('오류가 발생했습니다: ' + data.message);
        }
    })
    .catch(error => {
        console.error('Error:', error);
        alert('요청 처리 중 오류가 발생했습니다.');
    });
}
