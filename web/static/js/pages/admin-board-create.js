/* web/static/js/pages/admin-board-create.js */
document.addEventListener('alpine:init', () => {
    Alpine.data('boardCreateForm', () => ({
        submitting: false,
        fields: [],
        fieldCount: 0,
        board_type: 'normal',
        previousCommentCheckbox: false,

        // 매니저 관련 속성 추가
        managers: [],
        searchResults: [],
        showSearchResults: false,

        init() {
            // fields 배열의 변경을 감시
            this.$watch('fields', value => {
                this.fieldCount = value.length;
            });

            // board_type의 변경을 감시
            this.$watch('board_type', value => {
                const commentsCheckbox = document.getElementById('comments_enabled');
                if (value === 'qna') {
                    this.previousCommentCheckbox = commentsCheckbox.checked;
                    commentsCheckbox.checked = true;
                    commentsCheckbox.disabled = true;
                } else {
                    commentsCheckbox.checked = this.previousCommentCheckbox;
                    commentsCheckbox.disabled = false;
                }
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

        // 사용자 검색 메서드 추가
        async searchUsers() {
            const searchTerm = document.getElementById('manager_search').value;
            if (!searchTerm || searchTerm.length < 2) {
                alert('검색어는 2글자 이상 입력해주세요.');
                return;
            }

            try {
                const csrfToken = document.querySelector('meta[name="csrf-token"]').content;
                const response = await fetch(`/api/admin/users/search?q=${encodeURIComponent(searchTerm)}`, {
                    headers: {
                        'X-CSRF-Token': csrfToken
                    }
                });

                if (response.ok) {
                    const data = await response.json();
                    this.searchResults = data.users || [];
                    this.showSearchResults = true;
                } else {
                    alert('사용자 검색 중 오류가 발생했습니다.');
                }
            } catch (error) {
                console.error('사용자 검색 오류:', error);
                alert('사용자 검색 중 오류가 발생했습니다.');
            }
        },

        // 매니저 추가 메서드 추가
        addManager(user) {
            // 이미 추가된 매니저인지 확인
            const exists = this.managers.some(m => m.id === user.id);
            if (exists) {
                alert('이미 매니저로 추가된 사용자입니다.');
                return;
            }

            this.managers.push(user);
            this.showSearchResults = false;
            document.getElementById('manager_search').value = '';
        },

        // 매니저 제거 메서드 추가
        removeManager(index) {
            this.managers.splice(index, 1);
        },

        // 폼 제출 메소드
        submitForm() {
            this.submitting = true;

            // 폼 요소를 ID로 가져오기
            const form = document.getElementById('board-create-form');

            const commentsCheckbox = document.getElementById('comments_enabled');
            const wasDisabled = commentsCheckbox.disabled;
            if (wasDisabled) {
                commentsCheckbox.disabled = false;
            }

            // FormData 객체 생성
            const formData = new FormData(form);
            formData.append('field_count', this.fields.length);

            if (wasDisabled) {
                commentsCheckbox.disabled = true;
            }

            // 매니저 ID를 폼에 추가
            const managerIds = this.managers.map(m => m.id).join(',');
            if (managerIds) {
                formData.append('manager_ids', managerIds);
            }

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