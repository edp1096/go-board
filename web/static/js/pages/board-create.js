// web/static/js/pages/board-create.js
document.addEventListener('DOMContentLoaded', function () {
    // 폼 제출 이벤트 처리
    const form = document.querySelector('form');
    if (form) {
        form.addEventListener('submit', function (e) {
            // form 기본 동작은 Alpine.js에서 처리
        });
    }

    // 필드 유효성 검사 추가 기능
    const validateField = (field) => {
        const fieldType = field.getAttribute('type');
        const isRequired = field.hasAttribute('required');

        if (!isRequired) return true;

        if (fieldType === 'text' || fieldType === 'textarea') {
            return field.value.trim() !== '';
        } else if (fieldType === 'number') {
            return !isNaN(field.value) && field.value !== '';
        }

        return true;
    };

    // 에디터 기능 초기화 (필요한 경우)
    const contentTextarea = document.getElementById('content');
    if (contentTextarea) {
        // 간단한 에디터 기능 추가 가능
        // 여기서는 생략
    }
});