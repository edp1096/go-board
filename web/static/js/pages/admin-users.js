// 사용자 역할 전환 함수
function toggleUserRole(userId, currentRole) {
    if (!confirm('이 사용자의 역할을 변경하시겠습니까?')) {
        return;
    }

    const userIdInt = parseInt(userId);
    const newRole = currentRole === 'admin' ? 'user' : 'admin';

    // CSRF 토큰 가져오기
    const csrfToken = document.querySelector('meta[name="csrf-token"]').content;

    fetch(`/admin/users/${userIdInt}/role`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
            'X-CSRF-Token': csrfToken,
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

// 문자열을 불리언으로 변환하는 함수
function parseBool(value) {
    // 불리언인 경우 그대로 반환
    if (typeof value === 'boolean') {
        return value;
    }

    // 문자열인 경우 "true"면 true, 그 외에는 false 반환
    if (typeof value === 'string') {
        return value.toLowerCase() === 'true';
    }

    // 숫자인 경우 0이 아니면 true
    if (typeof value === 'number') {
        return value !== 0;
    }

    // 그 외 경우는 !!value로 불리언 변환
    return !!value;
}

// 사용자 상태 전환 함수
function toggleUserStatus(userId, currentStatus) {
    // parseBool 함수로 문자열 "true"/"false"를 불리언으로 변환
    const isActive = parseBool(currentStatus);

    if (!confirm(isActive ? '이 사용자의 계정을 비활성화하시겠습니까?' : '이 사용자의 계정을 활성화하시겠습니까?')) {
        return;
    }

    const userIdInt = parseInt(userId);
    // 현재 상태의 반대값으로 변경
    const newStatus = !isActive;

    // CSRF 토큰 가져오기
    const csrfToken = document.querySelector('meta[name="csrf-token"]').content;

    fetch(`/api/admin/users/${userIdInt}/status`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
            'X-CSRF-Token': csrfToken,
        },
        body: JSON.stringify({
            active: newStatus // 불리언 값으로 변환된 상태 전송
        }),
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