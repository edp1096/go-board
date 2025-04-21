/* web/static/js/alpine-init.js */
// Alpine.js 초기화 및 전역 컴포넌트 정의

document.addEventListener('alpine:init', () => {
    // Notification 컴포넌트
    Alpine.data('notification', (message = '', type = 'info', autoClose = true) => ({
        visible: message !== '',
        message,
        type,

        init() {
            this.visible = this.message !== '';
            if (this.visible && autoClose) {
                setTimeout(() => {
                    this.visible = false;
                }, 250);
            }
        },

        close() {
            this.visible = false;
        }
    }));

    // 모달 컴포넌트
    Alpine.data('modal', () => ({
        open: false,

        toggle() {
            this.open = !this.open;
        },

        close() {
            this.open = false;
        }
    }));

    // 게시판 필드 관리 컴포넌트
    Alpine.data('boardFieldManager', (initialFields = []) => ({
        fields: initialFields,
        nextId: -1,

        init() {
            // 기존 필드에 isNew 속성 추가
            this.fields.forEach(field => {
                field.isNew = false;
            });
        },

        addField() {
            this.fields.push({
                id: this.nextId--,
                isNew: true,
                name: '',
                displayName: '',
                fieldType: 'text',
                required: false,
                sortable: false,
                searchable: false,
                options: ''
            });
        },

        removeField(index) {
            this.fields.splice(index, 1);
        }
    }));

    // 게시물 편집기 컴포넌트
    Alpine.data('postEditor', (boardID, postID = null, isEdit = false) => ({
        content: '',
        title: '',
        fields: {},
        submitting: false,
        errors: {},

        submitPost() {
            this.submitting = true;
            this.errors = {};

            // 폼 데이터 가져오기
            const form = this.$el;
            const formData = new FormData(form);

            // 요청 설정
            const method = isEdit ? 'PUT' : 'POST';
            const url = isEdit
                ? `/boards/${boardID}/posts/${postID}`
                : `/boards/${boardID}/posts`;

            // 요청 보내기
            fetch(url, {
                method,
                body: formData,
                headers: {
                    'Accept': 'application/json'
                }
            })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        if (isEdit) {
                            window.location.href = `/boards/${boardID}/posts/${postID}`;
                        } else {
                            window.location.href = `/boards/${boardID}/posts/${data.id}`;
                        }
                    } else {
                        this.errors = data.errors || { general: data.message };
                        this.submitting = false;
                    }
                })
                .catch(error => {
                    this.errors = { general: '요청 처리 중 오류가 발생했습니다.' };
                    this.submitting = false;
                    console.error('Error:', error);
                });
        }
    }));
});

// 사용자 정의 유틸리티 함수

// 날짜 포맷 유틸리티
function formatDate(dateString, format = 'YYYY-MM-DD HH:mm') {
    const date = new Date(dateString);
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');

    return format
        .replace('YYYY', year)
        .replace('MM', month)
        .replace('DD', day)
        .replace('HH', hours)
        .replace('mm', minutes);
}