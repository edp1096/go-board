// web/static/js/pages/admin-system-settings.js
document.addEventListener('DOMContentLoaded', function () {
    const approvalSettingsForm = document.getElementById('approvalSettingsForm');
    const approvalModeRadios = document.querySelectorAll('input[name="approval_mode"]');
    const delayedOptions = document.getElementById('delayedOptions');
    const errorMessage = document.getElementById('error-message');
    const successMessage = document.getElementById('success-message');

    // 초기 상태 설정
    toggleDelayedOptions();

    // 라디오 버튼 변경 시 설정 표시 여부 변경
    approvalModeRadios.forEach(radio => {
        radio.addEventListener('change', toggleDelayedOptions);
    });

    // 설정 폼 제출 처리
    approvalSettingsForm.addEventListener('submit', async function (e) {
        e.preventDefault();

        // 메시지 초기화
        errorMessage.classList.add('hidden');
        successMessage.classList.add('hidden');

        // 폼 데이터 수집
        const formData = new FormData(approvalSettingsForm);

        // CSRF 토큰 가져오기
        const csrfToken = document.querySelector('meta[name="csrf-token"]').content;

        try {
            // 서버에 설정 저장 요청
            const response = await fetch('/admin/system-settings', {
                method: 'POST',
                headers: {
                    'X-CSRF-Token': csrfToken
                },
                body: formData
            });

            const result = await response.json();

            if (result.success) {
                successMessage.textContent = result.message;
                successMessage.classList.remove('hidden');
            } else {
                errorMessage.textContent = result.message;
                errorMessage.classList.remove('hidden');
            }
        } catch (error) {
            console.error('Error:', error);
            errorMessage.textContent = '서버 통신 중 오류가 발생했습니다.';
            errorMessage.classList.remove('hidden');
        }
    });

    // 지연 승인 옵션 표시 여부 설정 함수
    function toggleDelayedOptions() {
        const isDelayed = document.getElementById('delayed').checked;

        if (isDelayed) {
            delayedOptions.classList.remove('hidden');
        } else {
            delayedOptions.classList.add('hidden');
        }
    }
});