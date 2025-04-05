/* web/static/js/alpine-components/board-create-form.js */
document.addEventListener('alpine:init', () => {
    Alpine.data('boardCreateForm', () => ({
        submitting: false,
        fields: [],
        fieldCount: 0,

        init() {
            // fields 배열의 변경을 감시
            this.$watch('fields', value => {
                this.fieldCount = value.length;
            });
        },

        // 필드 추가 메소드
        addField() {
            this.fields.push({
                name: '',
                displayName: '',
                fieldType: 'text',
                required: false,
                sortable: false,
                searchable: false,
                options: ''
            });
        },

        // 필드 제거 메소드
        removeField(index) {
            this.fields.splice(index, 1);
        },

        // 폼 제출 메소드
        submitForm() {
            this.submitting = true;

            // 폼 요소를 ID로 가져오기
            const form = document.getElementById('board-create-form');

            // FormData 객체 생성
            const formData = new FormData(form);
            formData.append('field_count', this.fields.length);

            // CSRF 토큰 가져오기
            const csrfToken = document.querySelector('meta[name="csrf-token"]').content;

            // 서버에 데이터 전송
            fetch('/admin/boards', {
                method: 'POST',
                headers: {
                    'X-CSRF-Token': csrfToken,
                    'Accept': 'application/json'
                },
                body: formData
            })
                .then(res => this.handleResponse(res))
                .catch(err => this.handleError(err));
        },

        // 응답 처리 메소드
        handleResponse(res) {
            // Content-Type 확인
            const contentType = res.headers.get('Content-Type');

            // JSON 응답인 경우
            if (contentType && contentType.includes('application/json')) {
                return res.json().then(data => {
                    if (data.success) {
                        window.location.href = '/admin/boards';
                    } else {
                        alert(data.message);
                        this.submitting = false;
                    }
                });
            }
            // HTML 응답 (오류 페이지)인 경우
            else {
                return res.text().then(html => {
                    console.error('서버에서 HTML 응답이 반환되었습니다', html);
                    alert('처리 중 오류가 발생했습니다. 다시 시도해 주세요.');
                    this.submitting = false;
                });
            }
        },

        // 오류 처리 메소드
        handleError(err) {
            console.error('요청 처리 중 오류:', err);
            alert('오류가 발생했습니다: ' + err);
            this.submitting = false;
        }
    }));
});