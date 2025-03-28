// web/static/js/pages/admin-board-edit.js
document.addEventListener('DOMContentLoaded', function () {
    // 기존 필드 처리 기능
    function setupExistingFields() {
        // 필드 정보는 Alpine.js에서 관리하므로 여기서는 추가 기능만 구현

        // 필드 타입에 따른 UI 조정
        document.querySelectorAll('[id^="field_type_"]').forEach(select => {
            select.addEventListener('change', function () {
                const index = this.id.split('_').pop();
                const optionsContainer = document.querySelector(`#options_container_${index}`);

                if (this.value === 'select' && optionsContainer) {
                    optionsContainer.style.display = 'block';
                } else if (optionsContainer) {
                    optionsContainer.style.display = 'none';
                }
            });

            // 초기 상태 설정
            select.dispatchEvent(new Event('change'));
        });
    }

    // 폼 제출 전 유효성 검사
    function validateEditForm() {
        const name = document.getElementById('name');
        if (!name || name.value.trim() === '') {
            alert('게시판 이름을 입력해주세요.');
            return false;
        }

        return true;
    }

    // 초기화
    setupExistingFields();

    // 나머지 기능은 Alpine.js에서 처리
});