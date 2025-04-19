document.addEventListener('DOMContentLoaded', function() {
    const form = document.getElementById('editUserForm');
    const errorMessage = document.getElementById('error-message');
    const successMessage = document.getElementById('success-message');
    const deleteBtn = document.getElementById('deleteUserBtn');

    form.addEventListener('submit', async function(e) {
        e.preventDefault();
        
        errorMessage.classList.add('hidden');
        successMessage.classList.add('hidden');
        
        const password = document.getElementById('password').value;
        const passwordConfirm = document.getElementById('password_confirm').value;
        if (password && password !== passwordConfirm) {
            errorMessage.textContent = '비밀번호가 일치하지 않습니다.';
            errorMessage.classList.remove('hidden');
            return;
        }
        
        const formData = new FormData(form);
        const active = document.getElementById('active').checked;
        const userId = formData.get('id');
        
        const data = {
            id: parseInt(userId),
            username: formData.get('username'),
            email: formData.get('email'),
            full_name: formData.get('full_name'),
            role: formData.get('role'),
            active: active
        };
        if (password) {
            data.password = password;
        }

        // CSRF 토큰 가져오기
        const csrfToken = document.querySelector('meta[name="csrf-token"]').content;
        
        try {
            const response = await fetch(`/admin/users/${userId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': csrfToken,
                },
                body: JSON.stringify(data)
            });
            const result = await response.json();
            if (result.success) {
                successMessage.textContent = '사용자 정보가 성공적으로 업데이트되었습니다.';
                successMessage.classList.remove('hidden');
                document.getElementById('password').value = '';
                document.getElementById('password_confirm').value = '';
            } else {
                errorMessage.textContent = result.message || '사용자 정보 업데이트 중 오류가 발생했습니다.';
                errorMessage.classList.remove('hidden');
            }
        } catch (error) {
            console.error('Error:', error);
            errorMessage.textContent = '서버 통신 중 오류가 발생했습니다.';
            errorMessage.classList.remove('hidden');
        }
    });

    deleteBtn.addEventListener('click', async function() {
        if (!confirm('정말로 이 사용자를 삭제하시겠습니까? 이 작업은 되돌릴 수 없습니다.')) return;
        const formData = new FormData(form);
        const userId = formData.get('id');
        try {
            const response = await fetch(`/admin/users/${userId}`, {
                method: 'DELETE',
                headers: {
                    'X-CSRF-Token': formData.get('csrf'),
                }
            });
            const result = await response.json();
            if (result.success) {
                alert('사용자가 성공적으로 삭제되었습니다.');
                window.location.href = '/admin/users';
            } else {
                errorMessage.textContent = result.message || '사용자 삭제 중 오류가 발생했습니다.';
                errorMessage.classList.remove('hidden');
            }
        } catch (error) {
            console.error('Error:', error);
            errorMessage.textContent = '서버 통신 중 오류가 발생했습니다.';
            errorMessage.classList.remove('hidden');
        }
    });
});
