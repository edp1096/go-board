document.addEventListener('DOMContentLoaded', function() {
    const form = document.getElementById('createUserForm');
    const errorMessage = document.getElementById('error-message');
    const successMessage = document.getElementById('success-message');

    form.addEventListener('submit', async function(e) {
        e.preventDefault();
        
        // 오류 메시지 초기화
        errorMessage.classList.add('hidden');
        successMessage.classList.add('hidden');
        
        // 비밀번호 일치 확인
        const password = document.getElementById('password').value;
        const passwordConfirm = document.getElementById('password_confirm').value;
        
        if (password !== passwordConfirm) {
            errorMessage.textContent = '비밀번호가 일치하지 않습니다.';
            errorMessage.classList.remove('hidden');
            return;
        }
        
        // 폼 데이터 생성
        const formData = new FormData(form);
        const active = document.getElementById('active').checked;
        
        // JSON 객체로 변환
        const data = {
            username: formData.get('username'),
            email: formData.get('email'),
            password: formData.get('password'),
            full_name: formData.get('full_name'),
            role: formData.get('role'),
            active: active
        };
        
        try {
            // 서버에 데이터 전송
            const response = await fetch('/admin/users', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': formData.get('csrf'),
                },
                body: JSON.stringify(data)
            });
            
            const result = await response.json();
            
            if (result.success) {
                // 성공 메시지 표시
                successMessage.textContent = '사용자가 성공적으로 추가되었습니다.';
                successMessage.classList.remove('hidden');
                
                // 폼 초기화
                form.reset();
                
                // 3초 후 사용자 목록 페이지로 리다이렉트
                setTimeout(() => {
                    window.location.href = '/admin/users';
                }, 2000);
            } else {
                // 오류 메시지 표시
                errorMessage.textContent = result.message || '사용자 추가 중 오류가 발생했습니다.';
                errorMessage.classList.remove('hidden');
            }
        } catch (error) {
            console.error('Error:', error);
            errorMessage.textContent = '서버 통신 중 오류가 발생했습니다.';
            errorMessage.classList.remove('hidden');
        }
    });
});
