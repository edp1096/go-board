/* web/static/js/comments.js */
// 댓글 기능 관련 JavaScript

document.addEventListener('DOMContentLoaded', function () {
    // Alpine.js를 사용하기 때문에 기본적인 기능은 템플릿에 구현되어 있습니다.
    // 여기서는 추가적인 고급 기능을 구현합니다.

    // 1. 댓글 영역으로 스크롤 (URL 해시가 #comments인 경우)
    if (window.location.hash === '#comments') {
        const commentsSection = document.querySelector('.comment-section');
        if (commentsSection) {
            commentsSection.scrollIntoView({ behavior: 'smooth' });
        }
    }

    // 2. 댓글 작성 폼 자동 포커스 (URL 파라미터에 comment=new가 있는 경우)
    const urlParams = new URLSearchParams(window.location.search);
    if (urlParams.get('comment') === 'new') {
        const commentTextarea = document.getElementById('comment');
        if (commentTextarea) {
            commentTextarea.focus();
        }
    }

    // 3. 댓글 내용 자동 저장 (localStorage)
    const commentTextarea = document.getElementById('comment');
    if (commentTextarea) {
        // 페이지 로드 시 저장된 내용이 있으면 복원
        const boardId = commentTextarea.getAttribute('data-board-id');
        const postId = commentTextarea.getAttribute('data-post-id');
        const storageKey = `comment_draft_${boardId}_${postId}`;
        const savedContent = localStorage.getItem(storageKey);

        if (savedContent) {
            // Alpine.js 모델에 직접 접근할 수 없으므로 이벤트를 통해 값을 설정
            commentTextarea.value = savedContent;
            commentTextarea.dispatchEvent(new Event('input'));
        }

        // 내용 변경 시 자동 저장
        commentTextarea.addEventListener('input', function () {
            localStorage.setItem(storageKey, this.value);
        });

        // 댓글 제출 시 저장된 내용 삭제
        const commentForm = commentTextarea.closest('form');
        if (commentForm) {
            commentForm.addEventListener('submit', function () {
                localStorage.removeItem(storageKey);
            });
        }
    }

    // 4. 댓글 글자 수 표시
    const commentCharCounter = document.getElementById('comment-char-counter');
    if (commentTextarea && commentCharCounter) {
        const updateCounter = function () {
            const maxLength = parseInt(commentTextarea.getAttribute('maxlength') || '1000');
            const currentLength = commentTextarea.value.length;
            commentCharCounter.textContent = `${currentLength}/${maxLength}`;

            // 남은 글자 수가 적으면 경고 스타일 적용
            if (currentLength > maxLength * 0.9) {
                commentCharCounter.classList.add('text-red-500');
            } else {
                commentCharCounter.classList.remove('text-red-500');
            }
        };

        commentTextarea.addEventListener('input', updateCounter);
        updateCounter(); // 초기화
    }

    // 5. 댓글에 멘션 기능 (@ 입력 시 사용자 목록 표시)
    // 이 기능은 서버 API 및 사용자 인터페이스 추가 구현이 필요하므로 여기서는 생략합니다.
});