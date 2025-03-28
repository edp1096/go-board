// web/static/js/pages/board-edit.js
document.addEventListener('DOMContentLoaded', function () {
    // 폼 제출 이벤트 처리
    const form = document.querySelector('form');
    if (form) {
        // 대부분의 처리는 Alpine.js에서 관리
    }

    // 에디터 기능 초기화 (필요한 경우)
    const contentTextarea = document.getElementById('content');
    if (contentTextarea) {
        // 간단한 에디터 기능 추가 가능
        // 여기서는 생략
    }

    // 삭제 버튼 이벤트 관리
    const deleteButton = document.querySelector('button[data-action="delete"]');
    if (deleteButton) {
        deleteButton.addEventListener('click', function () {
            if (confirm('정말 삭제하시겠습니까?')) {
                // 삭제 로직 (Alpine.js에서 처리하므로 여기서는 생략)
            }
        });
    }
});