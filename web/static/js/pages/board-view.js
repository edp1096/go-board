// board-view.js

let commentEditor;
let editCommentEditor; // 댓글 수정용 에디터
let isEditEditorInitialized = false; // 수정 에디터 초기화 여부 플래그

document.addEventListener('DOMContentLoaded', function () {
    initCommentEditor(); // 댓글 에디터 초기화
});

// 댓글 에디터 초기화 함수
function initCommentEditor() {
    // 에디터 요소가 있는지 확인
    const editorContainer = document.querySelector('editor-comment#comment-editor');
    if (!editorContainer) return;

    const shadowRoot = editorContainer.shadowRoot;
    const editorEl = shadowRoot.querySelector("#comment-editor-container");

    const boardId = document.getElementById('boardId').value;

    // 에디터 옵션 설정
    const editorOptions = {
        uploadInputName: "upload-files[]",
        uploadActionURI: `/api/boards/${boardId}/upload`,
        uploadAccessURI: `/uploads/boards/${boardId}/images`,
        placeholder: '댓글을 입력하세요...',
        uploadCallback: function (response) {
            console.log("댓글 이미지 업로드 완료:", response);
        }
    };

    // 에디터 초기화
    commentEditor = new MyEditor("", editorEl, editorOptions);
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
        uploadAccessURI: `/uploads/boards/${boardId}/images`,
        placeholder: '댓글을 수정하세요...',
        uploadCallback: function (response) {
            console.log("댓글 이미지 업로드 완료:", response);
        }
    };

    try {
        // 에디터 초기화 (빈 내용으로)
        editCommentEditor = new MyEditor("", editorEl, editorOptions);
        isEditEditorInitialized = true;
        console.log("수정 에디터가 초기화되었습니다.");
    } catch (error) {
        console.error("수정 에디터 초기화 중 오류:", error);
    }
}

// 에디터 내용 업데이트 함수
function updateEditCommentContent(content) {
    if (editCommentEditor && typeof editCommentEditor.setHTML === 'function') {
        try {
            editCommentEditor.setHTML(content);
        } catch (error) {
            console.error("에디터 내용 업데이트 중 오류:", error);
        }
    }
}

// Alpine.js 컴포넌트 수정
document.addEventListener('alpine:init', () => {
    Alpine.data('commentSystem', function (commentsEnabled) {
        return {
            comments: [],
            commentContent: '',
            editCommentContent: '',
            replyToId: null,
            replyToUser: '',
            loading: true,
            submitting: false,
            commentsEnabled: commentsEnabled,
            error: null,
            editingCommentId: null,
            showEditModal: false,

            init() {
                if (this.commentsEnabled) {
                    this.loadComments();
                } else {
                    this.loading = false;
                }

                // 한 번만 수정 에디터 초기화 (DOMContentLoaded 이후)
                setTimeout(() => {
                    if (!isEditEditorInitialized) {
                        initEditCommentEditor();
                    }
                }, 1000);
            },

            loadComments() {
                const boardId = document.getElementById('boardId').value;
                const postId = document.getElementById('postId').value;

                fetch(`/api/boards/${boardId}/posts/${postId}/comments`)
                    .then(res => res.json())
                    .then(data => {
                        if (data.success) {
                            this.comments = data.comments;
                        } else {
                            this.error = data.message;
                        }
                        this.loading = false;
                    })
                    .catch(err => {
                        this.error = '댓글을 불러오는 중 오류가 발생했습니다.';
                        this.loading = false;
                    });
            },

            focusCommentEditor() {
                try {
                    // 1. 먼저 직접 에디터 메서드로 시도
                    if (commentEditor && typeof commentEditor.focus === 'function') {
                        commentEditor.focus();
                        return;
                    }

                    // 2. ProseMirror 편집 가능 영역 찾아서 시도
                    const editorContainer = document.querySelector('editor-comment#comment-editor');
                    if (editorContainer) {
                        const shadowRoot = editorContainer.shadowRoot;
                        if (shadowRoot) {
                            // ProseMirror 편집 가능 영역 찾기
                            const editableDiv = shadowRoot.querySelector('[contenteditable="true"]');
                            if (editableDiv) {
                                editableDiv.focus();

                                // 커서를 끝으로 이동
                                const selection = window.getSelection();
                                const range = document.createRange();
                                range.selectNodeContents(editableDiv);
                                range.collapse(false); // false는 끝으로 이동
                                selection.removeAllRanges();
                                selection.addRange(range);
                                return;
                            }
                        }
                    }

                    console.log('에디터 영역을 찾을 수 없습니다.');
                } catch (error) {
                    console.error('에디터 포커스 설정 중 오류:', error);
                }
            },

            focusEditCommentEditor() {
                try {
                    // 1. 먼저 직접 에디터 메서드로 시도
                    if (editCommentEditor && typeof editCommentEditor.focus === 'function') {
                        editCommentEditor.focus();
                        return;
                    }

                    // 2. ProseMirror 편집 가능 영역 찾아서 시도
                    const editorContainer = document.querySelector('editor-edit-comment#edit-comment-editor');
                    if (editorContainer) {
                        const shadowRoot = editorContainer.shadowRoot;
                        if (shadowRoot) {
                            // ProseMirror 편집 가능 영역 찾기
                            const editableDiv = shadowRoot.querySelector('[contenteditable="true"]');
                            if (editableDiv) {
                                editableDiv.focus();

                                // 커서를 끝으로 이동
                                const selection = window.getSelection();
                                const range = document.createRange();
                                range.selectNodeContents(editableDiv);
                                range.collapse(false); // false는 끝으로 이동
                                selection.removeAllRanges();
                                selection.addRange(range);
                                return;
                            }
                        }
                    }

                    console.log('수정 에디터 영역을 찾을 수 없습니다.');
                } catch (error) {
                    console.error('수정 에디터 포커스 설정 중 오류:', error);
                }
            },

            setReplyInfo(id, username) {
                this.replyToId = id;
                this.replyToUser = username || '알 수 없음';

                // 댓글 영역으로 스크롤 이동
                const commentForm = document.querySelector('.comment-form-area');
                if (commentForm) {
                    commentForm.scrollIntoView({ behavior: 'smooth' });
                }

                // 250ms 후에 포커스 시도 (DOM이 완전히 업데이트된 후)
                setTimeout(() => {
                    this.focusCommentEditor();
                }, 250);

                // 답글 표시 추가
                if (this.replyToUser) {
                    const replyText = `<span>@${this.replyToUser}</span>&nbsp;`;
                    if (commentEditor && typeof commentEditor.setHTML === 'function') {
                        commentEditor.setHTML(replyText);
                    }
                }
            },

            submitComment() {
                const boardId = document.getElementById('boardId').value;
                const postId = document.getElementById('postId').value;

                // 에디터에서 HTML 내용 가져오기
                if (commentEditor) {
                    this.commentContent = commentEditor.getHTML();

                    // 빈 댓글 체크
                    if (!this.commentContent || this.commentContent === '<p></p>' || this.commentContent === '<p><br></p>') {
                        alert('댓글 내용을 입력해주세요.');
                        return;
                    }
                }

                // CSRF 토큰 가져오기
                const csrfToken = document.querySelector('meta[name="csrf-token"]').content;

                this.submitting = true;

                fetch(`/api/boards/${boardId}/posts/${postId}/comments`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'X-CSRF-Token': csrfToken
                    },
                    body: JSON.stringify({
                        content: this.commentContent,
                        parentId: this.replyToId
                    })
                })
                    .then(res => res.json())
                    .then(data => {
                        this.submitting = false;
                        if (data.success) {
                            if (this.replyToId === null) {
                                this.comments.push(data.comment);
                            } else {
                                // 부모 댓글 찾아서 답글 추가
                                const parentIndex = this.comments.findIndex(c => c.id === this.replyToId);
                                if (parentIndex !== -1) {
                                    if (!this.comments[parentIndex].children) {
                                        this.comments[parentIndex].children = [];
                                    }
                                    this.comments[parentIndex].children.push(data.comment);
                                }
                            }

                            // 입력 필드 초기화
                            this.commentContent = '';
                            this.replyToId = null;
                            this.replyToUser = '';

                            // 에디터 내용 초기화
                            if (commentEditor) {
                                commentEditor.setHTML('');
                            }
                        } else {
                            alert(data.message);
                        }
                    })
                    .catch(err => {
                        this.submitting = false;
                        console.error('댓글 등록 중 오류:', err);
                        alert('댓글 등록 중 오류가 발생했습니다.');
                    });
            },

            cancelReply() {
                this.replyToId = null;
                this.replyToUser = '';

                // 에디터 내용 초기화
                if (commentEditor) {
                    commentEditor.setHTML('');
                }
            },

            // 댓글 수정 모달 열기 (에디터 재사용 방식)
            openEditModal(commentId, content) {
                // 현재 수정할 댓글 정보 설정
                this.editingCommentId = commentId;
                this.editCommentContent = content;

                // 모달 표시 (인스턴스 생성 없이)
                this.showEditModal = true;

                // 에디터가 초기화되지 않았다면 초기화
                if (!isEditEditorInitialized) {
                    setTimeout(() => {
                        initEditCommentEditor();
                        setTimeout(() => {
                            // 에디터 내용 업데이트
                            updateEditCommentContent(content);
                            this.focusEditCommentEditor();
                        }, 100);
                    }, 100);
                } else {
                    // 이미 초기화된 에디터면 내용만 업데이트
                    setTimeout(() => {
                        updateEditCommentContent(content);
                        this.focusEditCommentEditor();
                    }, 50);
                }
            },

            // 댓글 수정 모달 닫기 (에디터 유지)
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
                fetch(`/api/comments/${this.editingCommentId}`, {
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
                            // 댓글 목록에서 해당 댓글 찾아서 업데이트
                            const updateComment = (comments) => {
                                for (let i = 0; i < comments.length; i++) {
                                    if (comments[i].id === this.editingCommentId) {
                                        comments[i].content = data.comment.content;
                                        comments[i].updatedAt = data.comment.updatedAt;
                                        return true;
                                    }
                                    if (comments[i].children) {
                                        if (updateComment(comments[i].children)) {
                                            return true;
                                        }
                                    }
                                }
                                return false;
                            };

                            updateComment(this.comments);

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

            // 이전 방식의 댓글 수정 함수 (이제 새 모달 열기로 변경)
            editComment(commentId, content) {
                this.openEditModal(commentId, content);
            },

            deleteComment(commentId, isReply = false, parentId = null) {
                // 기존 코드 유지
                // CSRF 토큰 가져오기
                const csrfToken = document.querySelector('meta[name="csrf-token"]').content;

                if (confirm('댓글을 삭제하시겠습니까?')) {
                    fetch(`/api/comments/${commentId}`, {
                        method: 'DELETE',
                        headers: {
                            'X-CSRF-Token': csrfToken
                        }
                    })
                        .then(res => res.json())
                        .then(data => {
                            if (data.success) {
                                if (isReply && parentId) {
                                    // 부모 댓글에서 답글 삭제
                                    const parentComment = this.comments.find(c => c.id === parentId);
                                    if (parentComment && parentComment.children) {
                                        parentComment.children = parentComment.children.filter(r => r.id !== commentId);
                                    }
                                } else {
                                    // 기본 댓글 삭제
                                    this.comments = this.comments.filter(c => c.id !== commentId);
                                }
                            } else {
                                alert(data.message);
                            }
                        });
                }
            }
        };
    });

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