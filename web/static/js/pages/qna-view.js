// qna-view.js

let answerEditor;
let editCommentEditor; // 댓글 수정용 에디터
let isEditEditorInitialized = false; // 수정 에디터 초기화 여부 플래그

document.addEventListener('DOMContentLoaded', function () {
    initAnswerEditor(); // 답변 에디터 초기화
});

// 답변 에디터 초기화 함수
function initAnswerEditor() {
    // 에디터 요소가 있는지 확인
    const editorContainer = document.querySelector('editor-comment#answer-editor');
    if (!editorContainer) return;

    const shadowRoot = editorContainer.shadowRoot;
    const editorEl = shadowRoot.querySelector("#comment-editor-container");

    const boardId = document.getElementById('boardId').value;

    // 에디터 옵션 설정
    const editorOptions = {
        uploadInputName: "upload-files[]",
        uploadActionURI: `/api/boards/${boardId}/upload`,
        uploadAccessURI: `/uploads/boards/${boardId}/medias`,
        placeholder: '답변을 입력하세요...',
        uploadCallback: function (response) {
            console.log("답변 이미지 업로드 완료:", response);
        }
    };

    // 에디터 초기화
    answerEditor = new MyEditor("", editorEl, editorOptions);
}

// 댓글 수정 에디터 초기화 함수 - 최초 한 번만 실행
function initEditCommentEditor() {
    // 이미 초기화되어 있다면 재사용
    if (isEditEditorInitialized && editCommentEditor) {
        return;
    }

    // 에디터 요소가 있는지 확인
    const editorContainer = document.querySelector('editor-edit-comment#edit-comment-editor');
    if (!editorContainer) return;

    const shadowRoot = editorContainer.shadowRoot;
    const editorEl = shadowRoot.querySelector("#edit-comment-editor-container");
    if (!editorEl) return;

    const boardId = document.getElementById('boardId').value;

    // 에디터 옵션 설정
    const editorOptions = {
        uploadInputName: "upload-files[]",
        uploadActionURI: `/api/boards/${boardId}/upload`,
        uploadAccessURI: `/uploads/boards/${boardId}/medias`,
        placeholder: '댓글을 수정하세요...',
        uploadCallback: function (response) {
            console.log("댓글 이미지 업로드 완료:", response);
        }
    };

    // 에디터 초기화 (빈 내용으로)
    editCommentEditor = new MyEditor("", editorEl, editorOptions);
    isEditEditorInitialized = true;
}

// 에디터 내용 업데이트 함수
function updateEditCommentContent(content) {
    if (editCommentEditor && typeof editCommentEditor.setHTML === 'function') {
        editCommentEditor.setHTML(content);
    }
}

// 파일 크기 포맷팅 함수
function formatFileSize(bytes) {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// 베스트 답변 ID를 프론트엔드에서 안전하게 가져오기
function getBestAnswerId() {
    // 서버 측 렌더링된 값에서 안전하게 가져오기
    const bestAnswerIdElement = document.getElementById('best-answer-id');
    if (bestAnswerIdElement) {
        const value = bestAnswerIdElement.value;
        return value ? parseInt(value) : 0;
    }
    return 0;
}

// Alpine.js 코드
document.addEventListener('alpine:init', () => {
    Alpine.data('answerManager', () => ({
        answers: [],
        bestAnswer: null,
        expandReplies: false,
        loading: true,
        error: null,
        content: '',
        submitting: false,
        replyToId: null,
        replyToUser: '',
        replyContent: '',
        submittingReply: false,
        editingCommentId: null,
        showEditModal: false,
        editCommentContent: '',
        editingReplyId: null, // 현재 수정 중인 답글 ID
        editContent: '', // 수정 중인 내용
        submittingEdit: false, // 수정 제출 중

        init() {
            this.loadAnswers();

            // 한 번만 수정 에디터 초기화 (DOMContentLoaded 이후)
            setTimeout(() => {
                if (!isEditEditorInitialized) {
                    initEditCommentEditor();
                }
            }, 250);
        },

        loadAnswers() {
            const boardId = document.getElementById('boardId').value;
            const postId = document.getElementById('postId').value;

            // 베스트 답변 ID 가져오기
            const bestAnswerId = getBestAnswerId();

            fetch(`/api/boards/${boardId}/qnas/${postId}/answers`)
                .then(res => res.json())
                .then(data => {
                    if (data.success) {
                        // 답변 배열이 null이거나 undefined인 경우 빈 배열로 처리
                        const answersWithMeta = (data.answers || []).map(answer => ({
                            ...answer,
                            isBestAnswer: answer.id === bestAnswerId,
                            // displayName: answer.user ? answer.user.username : '알 수 없음'
                            displayName: answer.user ? answer.user.fullName : '알 수 없음'
                        }));

                        // 베스트 답변 찾기
                        if (bestAnswerId) {
                            this.bestAnswer = answersWithMeta.find(a => a.id === bestAnswerId) || null;
                        } else {
                            this.bestAnswer = null;
                        }

                        // 모든 답변 포함 (베스트 답변도 목록에 포함)
                        this.answers = answersWithMeta;
                    } else {
                        // API에서 오류 메시지 반환한 경우
                        this.error = data.message || '답변을 불러오는데 실패했습니다.';
                    }
                    this.loading = false;
                })
                .catch(err => {
                    console.error('답변 로딩 중 오류:', err);
                    this.error = '답변을 불러오는 중 오류가 발생했습니다.';
                    this.loading = false;
                });
        },

        // 베스트 답변 스크롤 함수
        scrollToBestAnswer() {
            if (!this.bestAnswer) return;

            setTimeout(() => {
                const answerElement = document.querySelector(`[data-answer-id="${this.bestAnswer.id}"]`);
                if (answerElement) {
                    answerElement.scrollIntoView({ behavior: 'smooth', block: 'center' });

                    // 스크롤 후 하이라이트 효과
                    answerElement.classList.add('highlight-effect');
                    setTimeout(() => {
                        answerElement.classList.remove('highlight-effect');
                    }, 2000);
                }
            }, 100);
        },

        submitAnswer() {
            const boardId = document.getElementById('boardId').value;
            const postId = document.getElementById('postId').value;

            // 에디터에서 HTML 내용 가져오기
            if (answerEditor) {
                this.content = answerEditor.getHTML();

                // 빈 답변 체크
                if (!this.content || this.content === '<p></p>' || this.content === '<p><br></p>') {
                    alert('답변 내용을 입력해주세요.');
                    return;
                }
            }

            // CSRF 토큰 가져오기
            const csrfToken = document.querySelector('meta[name="csrf-token"]').content;

            this.submitting = true;

            fetch(`/api/boards/${boardId}/qnas/${postId}/answers`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': csrfToken
                },
                body: JSON.stringify({ content: this.content })
            })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        // 페이지 새로고침하여 새 답변 표시
                        window.location.reload();
                    } else {
                        alert('답변 등록 실패: ' + data.message);
                        this.submitting = false;
                    }
                })
                .catch(error => {
                    console.error('답변 등록 중 오류:', error);
                    alert('답변 등록 중 오류가 발생했습니다.');
                    this.submitting = false;
                });
        },

        // 답글 정보 설정
        setReplyInfo(answerId, username) {
            this.replyToId = answerId;
            this.replyToUser = username || '알 수 없음';
            this.replyContent = '';

            // 해당 답변 요소로 스크롤
            setTimeout(() => {
                const answerElement = document.querySelector(`[data-answer-id="${answerId}"]`);
                if (answerElement) {
                    answerElement.scrollIntoView({ behavior: 'smooth', block: 'center' });
                }
            }, 200);
        },

        // 답글 취소
        cancelReply() {
            this.replyToId = null;
            this.replyToUser = '';
            this.replyContent = '';
        },

        // 답글 제출
        submitReply(answerId) {
            if (!this.replyContent.trim()) {
                alert('답글 내용을 입력해주세요.');
                return;
            }

            const boardId = document.getElementById('boardId').value;

            // CSRF 토큰 가져오기
            const csrfToken = document.querySelector('meta[name="csrf-token"]').content;

            this.submittingReply = true;

            fetch(`/api/answers/${answerId}/replies`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': csrfToken
                },
                body: JSON.stringify({
                    content: this.replyContent
                })
            })
                .then(response => response.json())
                .then(data => {
                    this.submittingReply = false;
                    if (data.success) {
                        // 답변 객체 찾기
                        const answer = this.answers.find(a => a.id === answerId);
                        if (answer) {
                            // 자식 배열이 없으면 생성
                            if (!answer.children) {
                                answer.children = [];
                            }

                            // 새 답글 추가
                            answer.children.push(data.reply);

                            // 베스트 답변인 경우 베스트 답변에도 답글 추가
                            if (answer.isBestAnswer && this.bestAnswer) {
                                if (!this.bestAnswer.children) {
                                    this.bestAnswer.children = [];
                                }
                                this.bestAnswer.children.push(data.reply);
                                // 펼치기 상태 활성화
                                this.expandReplies = true;
                            }

                            // 입력 필드 초기화
                            this.replyContent = '';
                            this.replyToId = null;
                            this.replyToUser = '';
                        }
                    } else {
                        alert('답글 등록 실패: ' + data.message);
                    }
                })
                .catch(error => {
                    this.submittingReply = false;
                    console.error('답글 등록 중 오류:', error);
                    alert('답글 등록 중 오류가 발생했습니다.');
                });
        },

        // 댓글 수정 모달 열기
        editComment(commentId, content) {
            // 현재 수정할 댓글 정보 설정
            this.editingCommentId = commentId;
            this.editCommentContent = content;

            // 모달 표시
            this.showEditModal = true;

            // 에디터가 초기화되지 않았다면 초기화
            if (!isEditEditorInitialized) {
                setTimeout(() => {
                    initEditCommentEditor();
                    setTimeout(() => {
                        // 에디터 내용 업데이트
                        updateEditCommentContent(content);
                    }, 100);
                }, 100);
            } else {
                // 이미 초기화된 에디터면 내용만 업데이트
                setTimeout(() => {
                    updateEditCommentContent(content);
                }, 50);
            }
        },

        // 댓글 수정 모달 닫기
        closeEditModal() {
            // 상태만 변경 (에디터 인스턴스는 유지)
            this.showEditModal = false;
            this.editingCommentId = null;
            this.editCommentContent = '';

            // 내용만 비우기 (인스턴스는 유지)
            if (editCommentEditor && typeof editCommentEditor.setHTML === 'function') {
                try {
                    editCommentEditor.setHTML('');
                } catch (error) {
                    console.error("에디터 내용 초기화 중 오류:", error);
                }
            }
        },


        // 수정된 댓글 제출
        submitEditComment() {
            // 에디터에서 HTML 내용 가져오기
            if (editCommentEditor) {
                this.editCommentContent = editCommentEditor.getHTML();

                // 빈 댓글 체크
                if (!this.editCommentContent || this.editCommentContent === '<p></p>' || this.editCommentContent === '<p><br></p>') {
                    alert('댓글 내용을 입력해주세요.');
                    return;
                }
            }

            // CSRF 토큰 가져오기
            const csrfToken = document.querySelector('meta[name="csrf-token"]').content;

            // 수정 API 호출
            fetch(`/api/answers/${this.editingCommentId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': csrfToken
                },
                body: JSON.stringify({
                    content: this.editCommentContent
                })
            })
                .then(res => res.json())
                .then(data => {
                    if (data.success) {
                        // 일반 답변 목록 업데이트
                        const updateContent = (items) => {
                            for (let i = 0; i < items.length; i++) {
                                if (items[i].id === this.editingCommentId) {
                                    items[i].content = data.answer.content;
                                    items[i].updatedAt = data.answer.updatedAt;
                                    return true;
                                }
                                if (items[i].children) {
                                    if (updateContent(items[i].children)) {
                                        return true;
                                    }
                                }
                            }
                            return false;
                        };

                        // 일반 답변 목록 업데이트
                        updateContent(this.answers);

                        // 베스트 답변인 경우 베스트 답변도 업데이트
                        if (this.bestAnswer) {
                            if (this.bestAnswer.id === this.editingCommentId) {
                                this.bestAnswer.content = data.answer.content;
                                this.bestAnswer.updatedAt = data.answer.updatedAt;
                            } else if (this.bestAnswer.children) {
                                updateContent([this.bestAnswer]);
                            }
                        }

                        // 모달 닫기
                        this.closeEditModal();
                    } else {
                        alert(data.message);
                    }
                })
                .catch(err => {
                    console.error('댓글 수정 중 오류:', err);
                    alert('댓글 수정 중 오류가 발생했습니다.');
                });
        },

        deleteAnswer(answerId) {
            if (!confirm('답변을 삭제하시겠습니까?')) return;

            // CSRF 토큰 가져오기
            const csrfToken = document.querySelector('meta[name="csrf-token"]').content;

            fetch(`/api/answers/${answerId}`, {
                method: 'DELETE',
                headers: {
                    'X-CSRF-Token': csrfToken
                }
            })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        // 페이지 새로고침하여 삭제된 답변 반영
                        window.location.reload();
                    } else {
                        alert('답변 삭제 실패: ' + data.message);
                    }
                })
                .catch(error => {
                    console.error('답변 삭제 중 오류:', error);
                    alert('답변 삭제 중 오류가 발생했습니다.');
                });
        },


        // 답글 수정 모드 시작 (인라인 textarea로 변경)
        editReply(commentId, content) {
            // 이미 수정 중인 댓글이 있다면 취소
            this.cancelEdit();

            // 현재 수정할 댓글 정보 설정
            this.editingReplyId = commentId;
            this.editContent = this.extractTextContent(content); // HTML에서 텍스트만 추출

            // 답글 작성 중이면 취소
            this.cancelReply();

            // 수정할 요소로 스크롤
            setTimeout(() => {
                const replyElement = document.querySelector(`[data-reply-id="${commentId}"]`);
                if (replyElement) {
                    const textarea = replyElement.querySelector('.edit-textarea');
                    if (textarea) {
                        textarea.focus();
                    }
                }
            }, 100);
        },

        // HTML에서 텍스트 콘텐츠만 추출하는 헬퍼 함수
        extractTextContent(html) {
            if (!html) return '';
            // 임시 div를 생성하여 HTML 파싱
            const tempDiv = document.createElement('div');
            tempDiv.innerHTML = html;
            return tempDiv.textContent || tempDiv.innerText || '';
        },

        // 수정 취소
        cancelEdit() {
            this.editingReplyId = null;
            this.editContent = '';
            this.submittingEdit = false;
        },

        // 수정된 답글 제출 (인라인 textarea 버전)
        submitEditReply(commentId) {
            if (!this.editContent.trim()) {
                alert('답글 내용을 입력해주세요.');
                return;
            }

            // CSRF 토큰 가져오기
            const csrfToken = document.querySelector('meta[name="csrf-token"]').content;

            this.submittingEdit = true;

            fetch(`/api/answers/${commentId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': csrfToken
                },
                body: JSON.stringify({
                    content: this.editContent
                })
            })
                .then(res => res.json())
                .then(data => {
                    this.submittingEdit = false;
                    if (data.success) {
                        // 일반 답변 목록 업데이트
                        const updateContent = (items) => {
                            for (let i = 0; i < items.length; i++) {
                                if (items[i].id === commentId) {
                                    items[i].content = data.answer.content;
                                    items[i].updatedAt = data.answer.updatedAt;
                                    return true;
                                }
                                if (items[i].children) {
                                    const updated = items[i].children.some(child => {
                                        if (child.id === commentId) {
                                            child.content = data.answer.content;
                                            child.updatedAt = data.answer.updatedAt;
                                            return true;
                                        }
                                        return false;
                                    });
                                    if (updated) return true;
                                }
                            }
                            return false;
                        };

                        // 답변 목록 업데이트
                        updateContent(this.answers);

                        // 베스트 답변도 업데이트
                        if (this.bestAnswer && this.bestAnswer.children) {
                            updateContent([this.bestAnswer]);
                        }

                        // 수정 모드 종료
                        this.cancelEdit();
                    } else {
                        alert(data.message);
                    }
                })
                .catch(err => {
                    this.submittingEdit = false;
                    console.error('답글 수정 중 오류:', err);
                    alert('답글 수정 중 오류가 발생했습니다.');
                });
        },

        // 답글 삭제
        deleteComment(commentId, isReply = false, parentId = null) {
            if (!confirm('답글을 삭제하시겠습니까?')) return;

            // CSRF 토큰 가져오기
            const csrfToken = document.querySelector('meta[name="csrf-token"]').content;

            fetch(`/api/answers/${commentId}`, {
                method: 'DELETE',
                headers: {
                    'X-CSRF-Token': csrfToken
                }
            })
                .then(res => res.json())
                .then(data => {
                    if (data.success) {
                        if (isReply && parentId) {
                            // 부모 답변에서 답글 삭제
                            const parentAnswer = this.answers.find(a => a.id === parentId);
                            if (parentAnswer && parentAnswer.children) {
                                parentAnswer.children = parentAnswer.children.filter(r => r.id !== commentId);

                                // 베스트 답변이라면 베스트 답변에서도 삭제
                                if (parentAnswer.isBestAnswer && this.bestAnswer && this.bestAnswer.children) {
                                    this.bestAnswer.children = this.bestAnswer.children.filter(r => r.id !== commentId);
                                }
                            }
                        } else {
                            // 다른 경우 처리 (필요하다면)
                            window.location.reload();
                        }
                    } else {
                        alert(data.message);
                    }
                })
                .catch(err => {
                    console.error('답글 삭제 중 오류:', err);
                    alert('답글 삭제 중 오류가 발생했습니다.');
                });
        },

        voteAnswer(answerId, direction) {
            // CSRF 토큰 가져오기
            const csrfToken = document.querySelector('meta[name="csrf-token"]').content;

            fetch(`/api/answers/${answerId}/vote`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': csrfToken
                },
                body: JSON.stringify({ direction: direction })
            })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        // 현재 answers 배열에서 해당 답변 찾아 투표 수 업데이트
                        const answer = this.answers.find(a => a.id === answerId);
                        if (answer) {
                            answer.voteCount = data.voteCount;

                            // 베스트 답변인 경우 베스트 답변도 업데이트
                            if (answer.isBestAnswer && this.bestAnswer) {
                                this.bestAnswer.voteCount = data.voteCount;
                            }
                        }
                    } else {
                        alert('투표 실패: ' + data.message);
                    }
                })
                .catch(error => {
                    console.error('투표 처리 중 오류:', error);
                    alert('투표 처리 중 오류가 발생했습니다.');
                });
        },

        selectBestAnswer(answerId) {
            const boardId = document.getElementById('boardId').value;
            const postId = document.getElementById('postId').value;

            // CSRF 토큰 가져오기
            const csrfToken = document.querySelector('meta[name="csrf-token"]').content;

            fetch(`/api/boards/${boardId}/qnas/${postId}/best-answer`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': csrfToken
                },
                body: JSON.stringify({ answerId: answerId })
            })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        // 페이지 새로고침하여 베스트 답변 반영
                        window.location.reload();
                    } else {
                        alert('베스트 답변 선택 실패: ' + data.message);
                    }
                })
                .catch(error => {
                    console.error('베스트 답변 선택 중 오류:', error);
                    alert('베스트 답변 선택 중 오류가 발생했습니다.');
                });
        },

        // 날짜 포맷팅 함수
        formatDate(dateString) {
            if (!dateString) return '';
            const date = new Date(dateString);
            return date.toLocaleString();
        }
    }));

    Alpine.data('postActions', function () {
        return {
            deletePost() {
                const boardId = document.getElementById('boardId').value;
                const postId = document.getElementById('postId').value;

                if (confirm('정말 삭제하시겠습니까?')) {
                    fetch(`/boards/${boardId}/posts/${postId}`, {
                        method: 'DELETE',
                        headers: { 'Accept': 'application/json' }
                    })
                        .then(res => res.json())
                        .then(data => {
                            if (data.success) {
                                window.location.href = `/boards/${boardId}/posts`;
                            } else {
                                alert(data.message);
                            }
                        });
                }
            }
        };
    });
});