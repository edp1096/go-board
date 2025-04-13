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