// board-view.js
document.addEventListener('alpine:init', () => {
    Alpine.data('commentSystem', function (commentsEnabled) {
        return {
            comments: [],
            commentContent: '',
            replyToId: null,
            replyToUser: '',
            loading: true,
            commentsEnabled: commentsEnabled,
            error: null,

            init() {
                if (this.commentsEnabled) {
                    this.loadComments();
                } else {
                    this.loading = false;
                }
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

            submitComment() {
                const boardId = document.getElementById('boardId').value;
                const postId = document.getElementById('postId').value;
                const csrfToken = document.getElementById('csrfToken').value;

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
                            this.commentContent = '';
                            this.replyToId = null;
                            this.replyToUser = '';
                        } else {
                            alert(data.message);
                        }
                    });
            },

            cancelReply() {
                this.replyToId = null;
                this.replyToUser = '';
            },

            editComment(commentId, content) {
                const csrfToken = document.getElementById('csrfToken').value;
                const newContent = prompt('댓글을 수정합니다:', content);

                if (newContent && newContent !== content) {
                    fetch(`/api/comments/${commentId}`, {
                        method: 'PUT',
                        headers: {
                            'Content-Type': 'application/json',
                            'X-CSRF-Token': csrfToken
                        },
                        body: JSON.stringify({
                            content: newContent
                        })
                    })
                        .then(res => res.json())
                        .then(data => {
                            if (data.success) {
                                // 댓글 목록에서 해당 댓글 찾아서 업데이트
                                const updateComment = (comments) => {
                                    for (let i = 0; i < comments.length; i++) {
                                        if (comments[i].id === commentId) {
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
                            } else {
                                alert(data.message);
                            }
                        });
                }
            },

            deleteComment(commentId, isReply = false, parentId = null) {
                const csrfToken = document.getElementById('csrfToken').value;

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