/* web/static/js/pages/admin-board-edit.js */
document.addEventListener('alpine:init', () => {
    Alpine.data('boardEditForm', (board) => ({
        board: board,
        submitting: false,
        fields: board.fields || [],
        fieldCount: 0,
        nextId: -1,
        isQnaBoard: false,
        board_type: document.getElementById('board_type').value || 'normal',

        // 매니저 관련 속성
        managers: [],
        searchResults: [],
        showSearchResults: false,

        // 소모임 참여자 관련 속성
        participants: [],
        participantResults: [],
        showParticipantResults: false,
        originalParticipants: [],

        init() {
            // 게시판 유형이 QnA인지 확인
            const boardTypeSelect = document.getElementById('board_type');
            if (boardTypeSelect) {
                this.isQnaBoard = boardTypeSelect.value === 'qna';
            }

            // QnA 게시판이면 댓글 기능 체크박스 해제 및 비활성화
            if (this.isQnaBoard) {
                const commentsCheckbox = document.getElementById('comments_enabled');
                if (commentsCheckbox) {
                    commentsCheckbox.checked = document.getElementById('comments_enabled').checked ? true : false;
                    commentsCheckbox.disabled = true;
                }

                // QnA 게시판이면 비밀글 설정 체크박스 해제 및 비활성화
                const privateCheckbox = document.getElementById('allow_private');
                if (privateCheckbox) {
                    privateCheckbox.checked = false;
                    privateCheckbox.disabled = true;
                }
            }

            // // 스크립트 태그에서 필드 데이터 초기화
            // try {
            //     const initialFieldsScript = document.getElementById('initial-field-data');
            //     if (initialFieldsScript && initialFieldsScript.textContent) {
            //         // initialFieldsScript.textContent가 이미 문자열이므로 직접 확인
            //         let parsedFields;
            //         const content = initialFieldsScript.textContent.trim();

            //         // 이미 JSON 문자열인지 확인 (따옴표로 시작하는지)
            //         if (content.startsWith('"') && content.endsWith('"')) {
            //             // 이미 문자열화된 JSON인 경우 이스케이프된 따옴표 처리
            //             const unescapedContent = content.slice(1, -1).replace(/\\"/g, '"');
            //             parsedFields = JSON.parse(unescapedContent);
            //         } else {
            //             // 일반 JSON인 경우
            //             parsedFields = JSON.parse(content);
            //         }

            //         // 배열이 아닌 경우 배열로 변환
            //         if (!Array.isArray(parsedFields)) {
            //             if (typeof parsedFields === 'object' && parsedFields !== null) {
            //                 parsedFields = [parsedFields]; // 객체인 경우 배열로 변환
            //             } else {
            //                 parsedFields = []; // 그 외의 경우 빈 배열로
            //             }
            //         }

            //         // 기존 필드에 isNew 속성 추가
            //         this.fields = parsedFields.map(field => ({
            //             ...field,
            //             isNew: false,
            //             // columnName이 없는 경우 name으로 설정
            //             columnName: field.columnName || field.name
            //         }));
            //     } else {
            //         this.fields = [];
            //     }
            // } catch (e) {
            //     this.fields = [];
            // }

            // 매니저 데이터 초기화
            try {
                const initialManagersScript = document.getElementById('initial-managers-data');
                if (initialManagersScript && initialManagersScript.textContent) {
                    const content = initialManagersScript.textContent.trim();

                    // JSON 파싱
                    if (content.startsWith('"') && content.endsWith('"')) {
                        const unescapedContent = content.slice(1, -1).replace(/\\"/g, '"');
                        this.managers = JSON.parse(unescapedContent) || [];
                    } else {
                        this.managers = JSON.parse(content) || [];
                    }

                    // 배열이 아닌 경우 처리
                    if (!Array.isArray(this.managers)) {
                        this.managers = [];
                    }
                } else {
                    this.managers = [];
                }
            } catch (e) {
                console.error('매니저 데이터 파싱 오류:', e);
                this.managers = [];
            }

            // 다음 ID 설정 (새 필드용)
            this.nextId = -1;

            // fields 배열의 변경을 감시
            this.$watch('fields', value => {
                this.fieldCount = value.length;
            });

            // 소모임 게시판인 경우 참여자 목록 로드
            if (this.board_type === 'group') {
                this.loadParticipants();
            }
        },

        // 필드 추가 메소드
        addField() {
            this.fields.push({
                id: this.nextId--,
                isNew: true,
                name: '',
                columnName: '', // 명시적으로 columnName 속성 추가
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

        // 참여자 목록 로드
        async loadParticipants() {
            try {
                const csrfToken = document.querySelector('meta[name="csrf-token"]').content;
                const response = await fetch(`/api/admin/boards/${this.board.id}/participants`, {
                    headers: {
                        'X-CSRF-Token': csrfToken
                    }
                });

                if (response.ok) {
                    const data = await response.json();
                    this.participants = data.participants || [];
                    this.originalParticipants = JSON.parse(JSON.stringify(this.participants));
                }
            } catch (error) {
                console.error('참여자 목록 로드 오류:', error);
            }
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
            const exists = this.participants.some(p =>
                (p.id === user.id) || (p.userId === user.id) || (p.user?.id === user.id)
            );

            if (exists) {
                alert('이미 참여자로 추가된 사용자입니다.');
                return;
            }

            this.participants.push({
                id: user.id,
                userId: user.id,
                username: user.username,
                email: user.email,
                role: 'member',
                isNew: true
            });

            this.showParticipantResults = false;
            document.getElementById('participant_search').value = '';
        },

        // 참여자 제거
        removeParticipant(index) {
            this.participants.splice(index, 1);
        },

        async updateParticipants(boardId, csrfToken) {
            // 삭제할 참여자 처리
            for (const original of this.originalParticipants) {
                const exists = this.participants.some(p => p.userId === original.userId);
                if (!exists) {
                    await fetch(`/api/admin/boards/${boardId}/participants/${original.userId}`, {
                        method: 'DELETE',
                        headers: {
                            'X-CSRF-Token': csrfToken
                        }
                    });
                }
            }

            // 추가할 참여자 처리
            for (const participant of this.participants) {
                const isNew = !this.originalParticipants.some(o => o.userId === participant.userId);
                const roleChanged = this.originalParticipants.some(o =>
                    o.userId === participant.userId && o.role !== participant.role
                );

                if (isNew) {
                    await fetch(`/api/admin/boards/${boardId}/participants`, {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                            'X-CSRF-Token': csrfToken
                        },
                        body: JSON.stringify({
                            userId: participant.id || participant.userId,
                            role: participant.role
                        })
                    });
                } else if (roleChanged) {
                    await fetch(`/api/admin/boards/${boardId}/participants/${participant.userId}/role`, {
                        method: 'PUT',
                        headers: {
                            'Content-Type': 'application/json',
                            'X-CSRF-Token': csrfToken
                        },
                        body: JSON.stringify({
                            role: participant.role
                        })
                    });
                }
            }
        },

        // 폼 제출 메소드
        async submitForm() {
            this.submitting = true;

            // 폼 요소 가져오기
            const form = document.getElementById('board-edit-form');

            const commentsCheckbox = document.getElementById('comments_enabled');
            const privateCheckbox = document.getElementById('allow_private');
            const votesCheckbox = document.getElementById('votes_enabled'); // 추가: 좋아요/싫어요 체크박스

            const wasCommentsDisabled = commentsCheckbox.disabled;
            const wasPrivateDisabled = privateCheckbox.disabled;
            const wasVotesDisabled = votesCheckbox ? votesCheckbox.disabled : false; // 추가: 좋아요/싫어요 체크박스 상태

            // 비활성화된 체크박스를 일시적으로 활성화해서 값이 전송되도록 함
            if (wasCommentsDisabled) {
                commentsCheckbox.disabled = false;
            }
            if (wasPrivateDisabled) {
                privateCheckbox.disabled = false;
            }
            // 추가: 좋아요/싫어요 체크박스 활성화
            if (votesCheckbox && wasVotesDisabled) {
                votesCheckbox.disabled = false;
            }

            // FormData 객체 생성
            const formData = new FormData(form);
            formData.append('field_count', this.fields.length);

            // 비활성화 상태 복원
            if (wasCommentsDisabled) {
                commentsCheckbox.disabled = true;
            }
            if (wasPrivateDisabled) {
                privateCheckbox.disabled = true;
            }
            // 추가: 좋아요/싫어요 체크박스 상태 복원
            if (votesCheckbox && wasVotesDisabled) {
                votesCheckbox.disabled = true;
            }

            // 매니저 ID를 폼에 추가
            const managerIds = this.managers.map(m => m.id).join(',');
            if (managerIds) {
                formData.append('manager_ids', managerIds);
            }

            // 필드 데이터 디버깅용 객체
            const debugFields = {};

            // 각 필드에 대한 더 자세한 정보 수집
            const fieldsDetails = this.fields.map((field, index) => {
                return {
                    index,
                    id: field.id,
                    name: field.name,
                    columnName: field.columnName || field.name,
                    displayName: field.displayName,
                    isNew: field.isNew,
                    fieldType: field.fieldType
                };
            });

            this.fields.forEach((field, index) => {
                // 컬럼명 결정 - 기존 필드는 columnName을, 새 필드는 name을 사용
                const columnName = field.isNew ? field.name : (field.columnName || field.name);

                debugFields[`field_${index}`] = {
                    id: field.id,
                    name: field.name,
                    columnName: columnName,
                    isNew: field.isNew
                };

                // 폼 데이터에 필드 정보 명시적으로 추가
                formData.set(`field_id_${index}`, field.id);
                formData.set(`field_name_${index}`, columnName); // 이 부분을 columnName으로 설정
                formData.set(`display_name_${index}`, field.displayName);
                formData.set(`field_type_${index}`, field.fieldType);

                // 체크박스는 체크된 경우에만 값이 전송되므로 명시적으로 설정
                formData.set(`required_${index}`, field.required ? "on" : "off");
                formData.set(`sortable_${index}`, field.sortable ? "on" : "off");
                formData.set(`searchable_${index}`, field.searchable ? "on" : "off");

                // select 필드의 경우 옵션 추가
                if (field.fieldType === 'select' && field.options) {
                    formData.set(`options_${index}`, field.options);
                }
            });

            // FormData의 모든 키-값 쌍을 로깅
            let formEntries = {};
            for (let [key, value] of formData.entries()) {
                formEntries[key] = value;
            }

            // 폼 액션 URL 가져오기
            const actionUrl = form.getAttribute('action');

            // CSRF 토큰 가져오기
            const csrfToken = document.querySelector('meta[name="csrf-token"]').content;
            const boardId = this.board.id;

            // 서버에 데이터 전송
            const r = await fetch(actionUrl, {
                method: 'PUT',
                headers: {
                    'X-CSRF-Token': csrfToken,
                    'Accept': 'application/json'
                },
                body: formData
            });

            if (r.ok) {
                // 참여자 업데이트
                if (this.board_type === 'group') {
                    await this.updateParticipants(boardId, csrfToken);
                }
                this.handleResponse(r);

                return;
            }

            this.handleError('서버 응답 오류: ' + r.statusText);
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
                    alert('처리 중 오류가 발생했습니다. 다시 시도해 주세요.');
                    this.submitting = false;
                });
            }
        },

        // 오류 처리 메소드
        handleError(err) {
            alert('오류가 발생했습니다: ' + err);
            this.submitting = false;
        }
    }));
});