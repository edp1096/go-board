// board-view.js

let commentEditor;
let editCommentEditor; // 댓글 수정용 에디터
let isEditEditorInitialized = false; // 수정 에디터 초기화 여부 플래그
let inlineReplyEditor = null; // 인라인 답글 에디터

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
        uploadAccessURI: `/uploads/boards/${boardId}/medias`,
        uploadCallback: function (response) {
            // console.log("댓글 이미지 업로드 완료:", response);
        }
    };

    // 에디터 초기화
    commentEditor = new MyEditor("", editorEl, editorOptions);
}

// 인라인 답글 에디터 초기화 함수
function initInlineReplyEditor(containerId) {
    // 이전 인라인 에디터 참조 제거 (destroy 메소드 사용하지 않음)
    inlineReplyEditor = null;

    // 새 에디터 컨테이너 찾기
    const editorContainer = document.querySelector(`editor-inline-reply#inline-reply-editor-${containerId}`);
    if (!editorContainer) return null;

    const shadowRoot = editorContainer.shadowRoot;
    const editorEl = shadowRoot.querySelector("#inline-reply-editor-container");
    if (!editorEl) return null;

    const boardId = document.getElementById('boardId').value;

    // 에디터 옵션 설정
    const editorOptions = {
        uploadInputName: "upload-files[]",
        uploadActionURI: `/api/boards/${boardId}/upload`,
        uploadAccessURI: `/uploads/boards/${boardId}/medias`,
        uploadCallback: function (response) {
            // console.log("이미지 업로드 완료:", response);
        }
    };

    // 에디터 초기화
    inlineReplyEditor = new MyEditor("", editorEl, editorOptions);
    return inlineReplyEditor;
}

// 인라인 답글 에디터에 초기 내용 설정
function setInlineReplyEditorContent(username, editorId) {
    if (inlineReplyEditor && typeof inlineReplyEditor.setHTML === 'function') {
        const replyText = `<span>@${username}</span>&nbsp;`;
        inlineReplyEditor.setHTML(replyText);

        // 에디터 포커스 설정
        setTimeout(() => {
            if (inlineReplyEditor && typeof inlineReplyEditor.focus === 'function') {
                inlineReplyEditor.focus();
                return;
            }

            // ProseMirror 편집 가능 영역 포커스 시도
            const editorContainer = document.querySelector(`editor-inline-reply#inline-reply-editor-${editorId}`);
            if (!editorContainer) return;

            const shadowRoot = editorContainer.shadowRoot;
            const editableDiv = shadowRoot.querySelector('[contenteditable="true"]');
            if (editableDiv) {
                editableDiv.focus();

                // 커서를 끝으로 이동
                const selection = window.getSelection();
                const range = document.createRange();
                range.selectNodeContents(editableDiv);
                range.collapse(false);
                selection.removeAllRanges();
                selection.addRange(range);
            }
        }, 50);
    }
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
        uploadCallback: function (response) {
            // console.log("댓글 이미지 업로드 완료:", response);
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

// Alpine.js 컴포넌트 수정
document.addEventListener('alpine:init', () => {
    // 댓글 시스템 컴포넌트
    Alpine.data('commentSystem', function (commentsEnabled) {
        return {
            comments: [],
            commentContent: '',
            editCommentContent: '',
            replyToId: null,
            replyToUser: '',
            inlineReplyToId: null,  // 인라인 답글 대상 ID
            inlineReplyToUser: '',  // 인라인 답글 대상 사용자명
            inlineReplyContent: '', // 인라인 답글 내용
            inlineEditorId: null,   // 현재 활성화된 인라인 에디터 ID
            loading: true,
            submitting: false,
            inlineSubmitting: false, // 인라인 폼 제출 중 상태
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
                }, 250);

                // 인라인 에디터 요소 등록
                this.registerInlineEditorElement();
            },

            // 인라인 에디터용 커스텀 엘리먼트 등록
            registerInlineEditorElement() {
                if (!customElements.get('editor-inline-reply')) {
                    class InlineReplyEditorElement extends HTMLElement {
                        constructor() {
                            super();
                            const templateContent = document.querySelector("#inline-reply-editor-template")?.content;
                            if (templateContent) {
                                const shadowRoot = this.attachShadow({ mode: "open" });
                                shadowRoot.appendChild(templateContent.cloneNode(true));
                            }
                        }
                    }
                    customElements.define("editor-inline-reply", InlineReplyEditorElement);
                }
            },

            // 댓글 좋아요/싫어요 처리
            async voteComment(commentId, value) {
                const currentUserId = document.getElementById('currentUserId');
                if (!currentUserId || !currentUserId.value) {
                    alert('로그인이 필요합니다.');
                    return;
                }

                const boardId = document.getElementById('boardId').value;
                const csrfToken = document.querySelector('meta[name="csrf-token"]').content;

                try {
                    const response = await fetch(`/api/comments/${commentId}/vote`, {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                            'X-CSRF-Token': csrfToken
                        },
                        body: JSON.stringify({
                            boardId: boardId,
                            value: value
                        })
                    });

                    const data = await response.json();

                    if (data.success) {
                        // 댓글 목록 업데이트
                        this.updateCommentVotes(commentId, data.likes, data.dislikes, data.userVote);
                    } else {
                        alert(data.message);
                    }
                } catch (error) {
                    console.error('댓글 투표 오류:', error);
                    alert('댓글 투표 중 오류가 발생했습니다.');
                }
            },

            // 댓글 투표 상태 업데이트 (로컬 상태 관리)
            updateCommentVotes(commentId, likes, dislikes, userVote) {
                // 최상위 댓글 먼저 확인
                const comment = this.comments.find(c => c.id === commentId);
                if (comment) {
                    comment.likeCount = likes;
                    comment.dislikeCount = dislikes;
                    comment.userVote = userVote;
                    return;
                }

                // 답글 확인
                this.comments.forEach(c => {
                    if (c.children && c.children.length > 0) {
                        const reply = c.children.find(r => r.id === commentId);
                        if (reply) {
                            reply.likeCount = likes;
                            reply.dislikeCount = dislikes;
                            reply.userVote = userVote;
                        }
                    }
                });
            },

            async loadCommentVoteStatuses() {
                const boardId = document.getElementById('boardId').value;
                const commentIds = this.getAllCommentIds();

                if (commentIds.length === 0) return;

                try {
                    const response = await fetch(`/api/comments/vote-statuses`, {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json'
                        },
                        body: JSON.stringify({
                            boardId: boardId,
                            commentIds: commentIds
                        })
                    });

                    const data = await response.json();

                    if (data.success) {
                        this.applyVoteStatuses(data.voteStatuses);
                    }
                } catch (error) {
                    console.error('댓글 투표 상태 로딩 오류:', error);
                }
            },

            // 모든 댓글 ID 수집
            getAllCommentIds() {
                const ids = [];

                this.comments.forEach(comment => {
                    ids.push(comment.id);

                    if (comment.children && comment.children.length > 0) {
                        comment.children.forEach(reply => {
                            ids.push(reply.id);
                        });
                    }
                });

                return ids;
            },

            // 투표 상태 적용
            applyVoteStatuses(voteStatuses) {
                if (!voteStatuses) return;

                this.comments.forEach(comment => {
                    if (voteStatuses[comment.id]) {
                        comment.userVote = voteStatuses[comment.id];
                    }

                    if (comment.children && comment.children.length > 0) {
                        comment.children.forEach(reply => {
                            if (voteStatuses[reply.id]) {
                                reply.userVote = voteStatuses[reply.id];
                            }
                        });
                    }
                });
            },

            async loadComments() {
                const boardId = document.getElementById('boardId').value;
                const postId = document.getElementById('postId').value;

                try {
                    const response = await fetch(`/api/boards/${boardId}/posts/${postId}/comments`);
                    const data = await response.json();

                    if (data.success) {
                        this.comments = data.comments;

                        // // 로그인 사용자인 경우 댓글 투표 상태 로드 - 필요없음
                        // const currentUserId = document.getElementById('currentUserId');
                        // if (!currentUserId || !currentUserId.value) {
                        //     await this.loadCommentVoteStatuses();
                        // }
                    } else {
                        this.error = data.message;
                    }
                } catch (err) {
                    this.error = '댓글을 불러오는 중 오류가 발생했습니다.';
                    console.error('댓글 로딩 오류:', err);
                } finally {
                    this.loading = false;
                }
            },

            // 기존 에디터 포커스 함수들
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

            // 인라인 답글 토글 함수
            toggleInlineReply(id, username, isReply, parentCommentId) {
                // 에디터 ID 결정 (댓글이면 comment-{id}, 답글이면 reply-{id})
                const newEditorId = isReply ? `reply-${id}` : `comment-${id}`;

                // 현재 작업 상황 저장
                const currentEditorId = this.inlineEditorId;
                const currentReplyToId = this.inlineReplyToId;

                // 일단 현재 에디터 정리
                this.closeInlineReply();

                // 같은 에디터를 다시 열려고 하는 경우는 여기서 종료 (토글)
                if (currentEditorId === newEditorId && currentReplyToId === id) {
                    return;
                }

                // 여기까지 오면 항상 새 에디터를 열어야 함
                this.inlineReplyToId = id;
                this.inlineReplyToUser = username || '알 수 없음';
                this.inlineEditorId = newEditorId;

                // 인라인 답글 대상 설정 (부모 댓글 ID가 있으면 그것 사용, 아니면 현재 ID 사용)
                if (isReply && parentCommentId) {
                    // 이 경우 답글에 답글을 다는 것이지만, API는 부모 댓글 ID에 달아야 함
                    this.inlineReplyToId = parentCommentId;
                }

                // 에디터 초기화 (DOM 업데이트 후)
                setTimeout(() => {
                    const editor = initInlineReplyEditor(newEditorId);
                    if (editor) {
                        setInlineReplyEditorContent(this.inlineReplyToUser, this.inlineEditorId);
                    }
                }, 10);
            },

            // 인라인 답글 폼 닫기
            // 인라인 답글 폼 닫기 함수 수정
            closeInlineReply() {
                if (!this.inlineEditorId) return;

                this.inlineReplyToId = null;
                this.inlineReplyToUser = '';
                this.inlineReplyContent = '';
                this.inlineEditorId = null;

                // 인라인 에디터 정리 - destroy 메소드가 없으므로 단순히 null 처리만 합니다
                if (inlineReplyEditor) {
                    // MyEditor 클래스에 destroy 메소드가 없으므로 참조만 제거
                    inlineReplyEditor = null;
                }
            },

            // 인라인 답글 제출
            async submitInlineReply() {
                const boardId = document.getElementById('boardId').value;
                const postId = document.getElementById('postId').value;

                // 에디터에서 HTML 내용 가져오기
                if (inlineReplyEditor) {
                    this.inlineReplyContent = inlineReplyEditor.getHTML();

                    // 빈 댓글 체크
                    if (!this.inlineReplyContent || this.inlineReplyContent === '<p></p>' || this.inlineReplyContent === '<p><br></p>') {
                        alert('답글 내용을 입력해주세요.');
                        return;
                    }
                }

                // CSRF 토큰 가져오기
                const csrfToken = document.querySelector('meta[name="csrf-token"]').content;

                this.inlineSubmitting = true;

                try {
                    const response = await fetch(`/api/boards/${boardId}/posts/${postId}/comments`, {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                            'X-CSRF-Token': csrfToken
                        },
                        body: JSON.stringify({
                            content: this.inlineReplyContent,
                            parentId: this.inlineReplyToId
                        })
                    });

                    const data = await response.json();

                    if (data.success) {
                        const newCommentId = data.comment.id; // 새 댓글 ID

                        // 부모 댓글 찾아서 답글 추가
                        const parentIndex = this.comments.findIndex(c => c.id === this.inlineReplyToId);
                        if (parentIndex !== -1) {
                            if (!this.comments[parentIndex].children) {
                                this.comments[parentIndex].children = [];
                            }
                            this.comments[parentIndex].children.push(data.comment);
                        }

                        // 인라인 답글 폼 닫기
                        this.closeInlineReply();

                        // DOM 업데이트 후 새 댓글의 이미지 처리
                        setTimeout(() => {
                            if (typeof processNewCommentImages === 'function') {
                                processNewCommentImages(newCommentId);
                            }
                        }, 200);
                    } else {
                        alert(data.message);
                    }
                } catch (err) {
                    console.error('답글 등록 중 오류:', err);
                    alert('답글 등록 중 오류가 발생했습니다.');
                } finally {
                    this.inlineSubmitting = false;
                }
            },

            // 기존 페이지 하단 댓글 폼용 메소드들 (유지)
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
                }, 50);

                // 답글 표시 추가
                if (this.replyToUser) {
                    const replyText = `<span>@${this.replyToUser}</span>&nbsp;`;
                    if (commentEditor && typeof commentEditor.setHTML === 'function') {
                        commentEditor.setHTML(replyText);
                    }
                }
            },

            async submitComment() {
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

                try {
                    const response = await fetch(`/api/boards/${boardId}/posts/${postId}/comments`, {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                            'X-CSRF-Token': csrfToken
                        },
                        body: JSON.stringify({
                            content: this.commentContent,
                            parentId: this.replyToId
                        })
                    });

                    const data = await response.json();

                    if (data.success) {
                        const newCommentId = data.comment.id; // 새 댓글 ID

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

                        // DOM 업데이트 후 새 댓글의 이미지 처리
                        setTimeout(() => {
                            if (typeof processNewCommentImages === 'function') {
                                processNewCommentImages(newCommentId);
                            }
                        }, 200);
                    } else {
                        alert(data.message);
                    }
                } catch (err) {
                    console.error('댓글 등록 중 오류:', err);
                    alert('댓글 등록 중 오류가 발생했습니다.');
                } finally {
                    this.submitting = false;
                }
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
                        }, 10);
                    }, 10);
                } else {
                    // 이미 초기화된 에디터면 내용만 업데이트
                    setTimeout(() => {
                        updateEditCommentContent(content);
                        this.focusEditCommentEditor();
                    }, 10);
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
            async submitEditComment() {
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

                try {
                    const response = await fetch(`/api/comments/${this.editingCommentId}`, {
                        method: 'PUT',
                        headers: {
                            'Content-Type': 'application/json',
                            'X-CSRF-Token': csrfToken
                        },
                        body: JSON.stringify({
                            content: this.editCommentContent
                        })
                    });

                    const data = await response.json();

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
                } catch (err) {
                    console.error('댓글 수정 중 오류:', err);
                    alert('댓글 수정 중 오류가 발생했습니다.');
                }
            },

            // 이전 방식의 댓글 수정 함수 (이제 새 모달 열기로 변경)
            editComment(commentId, content) {
                this.openEditModal(commentId, content);
            },

            async deleteComment(commentId, isReply = false, parentId = null) {
                // CSRF 토큰 가져오기
                const csrfToken = document.querySelector('meta[name="csrf-token"]').content;

                if (confirm('댓글을 삭제하시겠습니까?')) {
                    try {
                        const response = await fetch(`/api/comments/${commentId}`, {
                            method: 'DELETE',
                            headers: {
                                'X-CSRF-Token': csrfToken
                            }
                        });

                        const data = await response.json();

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
                    } catch (err) {
                        console.error('댓글 삭제 중 오류:', err);
                        alert('댓글 삭제 중 오류가 발생했습니다.');
                    }
                }
            }
        };
    });

    // 게시물 액션 컴포넌트
    Alpine.data('postActions', function () {
        return {
            async deletePost() {
                const boardId = document.getElementById('boardId').value;
                const postId = document.getElementById('postId').value;

                if (confirm('정말 삭제하시겠습니까?')) {
                    try {
                        const response = await fetch(`/boards/${boardId}/posts/${postId}`, {
                            method: 'DELETE',
                            headers: { 'Accept': 'application/json' }
                        });

                        const data = await response.json();

                        if (data.success) {
                            window.location.href = `/boards/${boardId}/posts`;
                        } else {
                            alert(data.message);
                        }
                    } catch (err) {
                        console.error('게시물 삭제 중 오류:', err);
                        alert('게시물 삭제 중 오류가 발생했습니다.');
                    }
                }
            }
        };
    });

    // Alpine.js 컴포넌트 - 게시물 좋아요/싫어요
    Alpine.data('postVotes', function () {
        return {
            likeCount: 0,
            dislikeCount: 0,
            userVote: 0, // 0: 투표 안함, 1: 좋아요, -1: 싫어요

            init() {
                // 데이터 속성에서 초기값 로드
                const container = document.getElementById('post-votes-container');
                if (container) {
                    this.likeCount = parseInt(container.dataset.likeCount) || 0;
                    this.dislikeCount = parseInt(container.dataset.dislikeCount) || 0;
                }

                this.loadUserVoteStatus();
            },

            async loadUserVoteStatus() {
                // 로그인 체크
                const currentUserId = document.getElementById('currentUserId');
                if (!currentUserId || !currentUserId.value) return;

                const boardId = document.getElementById('boardId').value;
                const postId = document.getElementById('postId').value;

                try {
                    const response = await fetch(`/api/boards/${boardId}/posts/${postId}/vote-status`);
                    const data = await response.json();

                    if (data.success) {
                        this.userVote = data.voteValue;
                    }
                } catch (error) {
                    console.error('투표 상태 로딩 오류:', error);
                }
            },

            async votePost(value) {
                // 로그인 체크
                const currentUserId = document.getElementById('currentUserId');
                if (!currentUserId || !currentUserId.value) {
                    alert('로그인이 필요합니다.');
                    return;
                }

                const boardId = document.getElementById('boardId').value;
                const postId = document.getElementById('postId').value;
                const csrfToken = document.querySelector('meta[name="csrf-token"]').content;

                try {
                    const response = await fetch(`/api/boards/${boardId}/posts/${postId}/vote`, {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                            'X-CSRF-Token': csrfToken
                        },
                        body: JSON.stringify({
                            value: value
                        })
                    });

                    const data = await response.json();

                    if (data.success) {
                        this.likeCount = data.likes;
                        this.dislikeCount = data.dislikes;
                        this.userVote = data.userVote;
                    } else {
                        alert(data.message);
                    }
                } catch (error) {
                    console.error('투표 오류:', error);
                    alert('투표 중 오류가 발생했습니다.');
                }
            }
        };
    });

});