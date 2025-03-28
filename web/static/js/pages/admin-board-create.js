// web/static/js/pages/admin-board-create.js
document.addEventListener('DOMContentLoaded', function () {
    // 필드 관리 관련 기능
    const fieldManager = {
        nextId: -1,

        createOptionUI: function (options) {
            if (!options || options.length === 0) {
                return [
                    { value: 'option1', label: '옵션 1' },
                    { value: 'option2', label: '옵션 2' }
                ];
            }

            try {
                return JSON.parse(options);
            } catch (e) {
                console.error('옵션 파싱 오류:', e);
                return [
                    { value: 'option1', label: '옵션 1' },
                    { value: 'option2', label: '옵션 2' }
                ];
            }
        },

        getFieldTypes: function () {
            return [
                { value: 'text', label: '텍스트' },
                { value: 'textarea', label: '텍스트 영역' },
                { value: 'number', label: '숫자' },
                { value: 'date', label: '날짜' },
                { value: 'select', label: '선택 옵션' },
                { value: 'checkbox', label: '체크박스' },
                { value: 'file', label: '파일 업로드' }
            ];
        }
    };

    // 폼 유효성 검사 기능
    function validateForm() {
        const name = document.getElementById('name');
        if (!name || name.value.trim() === '') {
            alert('게시판 이름을 입력해주세요.');
            return false;
        }

        // Alpine.js에서 처리되는 필드 수 확인
        const fieldCount = document.querySelector('input[name="field_count"]');
        if (!fieldCount || parseInt(fieldCount.value, 10) === 0) {
            // 필드가 없어도 계속 진행할지 확인
            return confirm('추가 필드 없이 게시판을 생성하시겠습니까?');
        }

        return true;
    }

    // 기존에 알파인으로 처리하는 부분은 생략
});