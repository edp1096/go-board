/* web/static/js/pages/admin-board-create.js */
document.addEventListener('alpine:init', () => {
    Alpine.data('boardCreateForm', () => ({
        submitting: false,
        fields: [],
        fieldCount: 0,
        board_type: 'normal',
        previousCommentCheckbox: false,
        previousPrivateCheckbox: false, // 비밀글 설정 상태 저장 변수
        previousVotesCheckbox: false,  // 좋아요/싫어요 상태 저장

        // 매니저 관련 속성
        managers: [],
        searchResults: [],
        showSearchResults: false,

        // 소모임 참여자 관련 속성
        participants: [],
        participantResults: [],
        showParticipantResults: false,

        init() {
            // fields 배열의 변경을 감시
            this.$watch('fields', value => {
                this.fieldCount = value.length;
            });

            // board_type의 변경을 감시
            this.$watch('board_type', value => {
                const commentsCheckbox = document.getElementById('comments_enabled');
                const privateCheckbox = document.getElementById('allow_private'); // 비밀글 설정 체크박스
                const votesCheckbox = document.getElementById('votes_enabled'); // 좋아요/싫어요 체크박스

                if (value === 'qna') {
                    // 댓글 기능 설정 - 체크 및 비활성화
                    this.previousCommentCheckbox = commentsCheckbox.checked;
                    commentsCheckbox.checked = true;
                    commentsCheckbox.disabled = true;

                    // 비밀글 설정 - 체크 해제 및 비활성화
                    this.previousPrivateCheckbox = privateCheckbox.checked;
                    privateCheckbox.checked = false;
                    privateCheckbox.disabled = true;

                    // 좋아요/싫어요 설정 - 상태 저장
                    if (votesCheckbox) {
                        this.previousVotesCheckbox = votesCheckbox.checked;
                        // QnA 게시판에서는 좋아요/싫어요 체크박스 상태 유지
                    }
                } else {
                    // 댓글 기능 설정 - 이전 상태로 복원
                    commentsCheckbox.checked = this.previousCommentCheckbox;
                    commentsCheckbox.disabled = false;

                    // 비밀글 설정 - 이전 상태로 복원
                    privateCheckbox.checked = this.previousPrivateCheckbox;
                    privateCheckbox.disabled = false;

                    // 좋아요/싫어요 설정 - 이전 상태로 복원
                    if (votesCheckbox) {
                        votesCheckbox.checked = this.previousVotesCheckbox;
                        votesCheckbox.disabled = false;
                    }
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

        // 참여자 검색
        async searchParticipants() {
            const searchTerm = document.getElementById('participant_search').value;
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
                    this.participantResults = data.users || [];
                    this.showParticipantResults = true;
                }
            } catch (error) {
                console.error('사용자 검색 오류:', error);
                alert('사용자 검색 중 오류가 발생했습니다.');
            }
        },

        // 참여자 추가
        addParticipant(user) {
            const exists = this.participants.some(p => p.id === user.id);
            if (exists) {
                alert('이미 참여자로 추가된 사용자입니다.');
                return;
            }

            this.participants.push({
                ...user,
                role: 'member'  // 기본값
            });
            this.showParticipantResults = false;
            document.getElementById('participant_search').value = '';
        },

        // 참여자 제거
        removeParticipant(index) {
            this.participants.splice(index, 1);
        },

        // 폼 제출 메소드
        submitForm() {
            this.submitting = true;

            // 폼 요소를 ID로 가져오기
            const form = document.getElementById('board-create-form');
            const formData = new FormData(form);

            // 참여자 정보 추가
            if (this.board_type === 'group') {
                const participantData = this.participants.map(p => ({
                    id: p.id,
                    role: p.role
                }));
                formData.append('participants', JSON.stringify(participantData));
            }

            const commentsCheckbox = document.getElementById('comments_enabled');
            const privateCheckbox = document.getElementById('allow_private');
            const votesCheckbox = document.getElementById('votes_enabled');
            const wasCommentsDisabled = commentsCheckbox.disabled;
            const wasPrivateDisabled = privateCheckbox.disabled;
            const wasVotesDisabled = votesCheckbox ? votesCheckbox.disabled : false;

            // 비활성화된 체크박스를 일시적으로 활성화해서 값이 전송되도록 함
            if (wasCommentsDisabled) {
                commentsCheckbox.disabled = false;
            }
            if (wasPrivateDisabled) {
                privateCheckbox.disabled = false;
            }
            if (votesCheckbox && wasVotesDisabled) {
                votesCheckbox.disabled = false;
            }

            formData.append('field_count', this.fields.length);

            // 비활성화 상태 복원
            if (wasCommentsDisabled) {
                commentsCheckbox.disabled = true;
            }
            if (wasPrivateDisabled) {
                privateCheckbox.disabled = true;
            }
            if (votesCheckbox && wasVotesDisabled) {
                votesCheckbox.disabled = true;
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