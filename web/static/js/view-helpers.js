// view-helpers.js
// 게시물 조회 페이지에서 사용되는 유틸리티 함수들을 모은 파일

// 파일 크기 포맷팅 함수
function formatFileSize(bytes) {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// 첨부 파일 삭제 함수
async function deleteAttachment(id) {
    if (!confirm('첨부 파일을 삭제하시겠습니까?')) return;

    // CSRF 토큰 가져오기
    const csrfToken = document.querySelector('meta[name="csrf-token"]').content;

    try {
        const response = await fetch(`/api/attachments/${id}`, {
            method: 'DELETE',
            headers: {
                'X-CSRF-Token': csrfToken,
            }
        });

        const data = await response.json();

        if (data.success) {
            // 페이지 새로고침
            window.location.reload();
        } else {
            alert('첨부 파일 삭제 실패: ' + data.message);
        }
    } catch (err) {
        console.error('첨부 파일 삭제 중 오류:', err);
        alert('첨부 파일 삭제 중 오류가 발생했습니다.');
    }
}

// 이미지 뷰어 컴포넌트 초기화
document.addEventListener('alpine:init', () => {
    // 이미지 뷰어 컴포넌트
    Alpine.data('imageViewer', () => ({
        show: false,
        imageSrc: '',
        imageAlt: '',

        openImageViewer(src, alt) {
            this.imageSrc = src;
            this.imageAlt = alt || 'Image';
            this.show = true;
            // 스크롤 방지
            document.body.style.overflow = 'hidden';
        },

        closeImageViewer() {
            this.show = false;
            // 스크롤 복원
            document.body.style.overflow = '';
        }
    }));

    // 첨부 파일 관리자 컴포넌트 - 조회용 (게시글 보기 페이지)
    Alpine.data('attachmentManager', (boardId, postId) => ({
        attachments: [],

        init() {
            this.loadAttachments(boardId, postId);
        },

        async loadAttachments(boardId, postId) {
            try {
                const response = await fetch(`/api/boards/${boardId}/posts/${postId}/attachments`);
                const data = await response.json();

                if (data.success) {
                    this.attachments = data.attachments || [];
                }
            } catch (error) {
                console.error('첨부 파일 로딩 중 오류:', error);
            }
        }
    }));

    // 첨부 파일 생성 컴포넌트 - 새 게시글 작성 페이지용
    Alpine.data('newAttachmentManager', () => ({
        attachments: [],
        deletedFiles: [],

        init() {
            // 새 게시글에는 첨부 파일이 없으므로 빈 배열로 초기화
            this.attachments = [];
        },

        markForDeletion(file) {
            this.deletedFiles.push(file);
            this.attachments = this.attachments.filter(f => f.id !== file.id);
        },

        restoreFile(fileId) {
            const fileToRestore = this.deletedFiles.find(file => file.id === fileId);
            if (fileToRestore) {
                this.attachments.push(fileToRestore);
                this.attachments.sort((a, b) => a.id - b.id);
                this.deletedFiles = this.deletedFiles.filter(file => file.id !== fileId);
            }
        }
    }));

    // 첨부 파일 에디터 컴포넌트 - 수정용 (게시글 수정 페이지)
    Alpine.data('attachmentEditor', (boardId, postId) => ({
        attachments: [],
        deletedFiles: [],

        init() {
            this.loadAttachments(boardId, postId);
        },

        async loadAttachments(boardId, postId) {
            try {
                const response = await fetch(`/api/boards/${boardId}/posts/${postId}/attachments`);
                const data = await response.json();

                if (data.success) {
                    this.attachments = data.attachments || [];
                }
            } catch (error) {
                console.error('첨부 파일 로딩 중 오류:', error);
            }
        },

        markForDeletion(file) {
            this.deletedFiles.push(file);
            this.attachments = this.attachments.filter(f => f.id !== file.id);
        },

        restoreFile(fileId) {
            const fileToRestore = this.deletedFiles.find(file => file.id === fileId);
            if (fileToRestore) {
                this.attachments.push(fileToRestore);
                // 정렬 (필요시)
                this.attachments.sort((a, b) => a.id - b.id);
                // 삭제 목록에서 제거
                this.deletedFiles = this.deletedFiles.filter(file => file.id !== fileId);
            }
        }
    }));
});

// 에디터 커스텀 요소 정의를 등록하는 함수
function registerEditorElements() {
    // 댓글 작성 에디터 요소
    if (!customElements.get('editor-comment')) {
        class CommentEditorElement extends HTMLElement {
            constructor() {
                super();
                const templateContent = document.querySelector("#comment-editor-template")?.content;
                if (templateContent) {
                    const shadowRoot = this.attachShadow({ mode: "open" });
                    shadowRoot.appendChild(templateContent.cloneNode(true));
                }
            }
        }
        customElements.define("editor-comment", CommentEditorElement);
    }

    // 댓글 수정 에디터 요소
    if (!customElements.get('editor-edit-comment')) {
        class EditCommentEditorElement extends HTMLElement {
            constructor() {
                super();
                const templateContent = document.querySelector("#edit-comment-editor-template")?.content;
                if (templateContent) {
                    const shadowRoot = this.attachShadow({ mode: "open" });
                    shadowRoot.appendChild(templateContent.cloneNode(true));
                }
            }
        }
        customElements.define("editor-edit-comment", EditCommentEditorElement);
    }
}

// 페이지 로드 시 초기화
document.addEventListener('DOMContentLoaded', function () {
    registerEditorElements();
});