/* comments.js */
// 댓글 기능 관련 JavaScript

document.addEventListener('DOMContentLoaded', function () {
    // 댓글 영역으로 스크롤 (URL 해시가 #comments인 경우)
    scrollToCommentsIfNeeded();

    // 댓글 작성 폼 자동 포커스 (URL 파라미터에 comment=new가 있는 경우)
    focusCommentFormIfNeeded();

    // 댓글 내용 자동 저장 (localStorage) 설정
    setupCommentAutosave();

    // 댓글 글자 수 표시 설정
    setupCommentCharCounter();
});

// 댓글 영역으로 스크롤하는 함수
function scrollToCommentsIfNeeded() {
    if (window.location.hash === '#comments') {
        const commentsSection = document.querySelector('.comment-section');
        if (commentsSection) {
            commentsSection.scrollIntoView({ behavior: 'smooth' });
        }
    }
}

// 댓글 폼에 자동 포커스하는 함수
function focusCommentFormIfNeeded() {
    const urlParams = new URLSearchParams(window.location.search);
    if (urlParams.get('comment') === 'new') {
        const commentTextarea = document.getElementById('comment');
        if (commentTextarea) {
            commentTextarea.focus();
        }
    }
}

// 댓글 자동 저장 기능 설정
function setupCommentAutosave() {
    const commentTextarea = document.getElementById('comment');
    if (!commentTextarea) return;

    // 페이지 로드 시 저장된 내용이 있으면 복원
    const boardId = commentTextarea.getAttribute('data-board-id');
    const postId = commentTextarea.getAttribute('data-post-id');

    if (!boardId || !postId) return;

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

// 댓글 글자 수 표시 기능
function setupCommentCharCounter() {
    const commentTextarea = document.getElementById('comment');
    const commentCharCounter = document.getElementById('comment-char-counter');

    if (!commentTextarea || !commentCharCounter) return;

    // 글자 수 업데이트 함수
    const updateCounter = function () {
        const maxLength = parseInt(commentTextarea.getAttribute('maxlength') || '1000');
        const currentLength = commentTextarea.value.length;
        commentCharCounter.textContent = `${currentLength}/${maxLength}`;

        // 남은 글자 수가 적으면 경고 스타일 적용
        if (currentLength > maxLength * 0.9) {
            commentCharCounter.classList.add('text-error');
        } else {
            commentCharCounter.classList.remove('text-error');
        }
    };

    commentTextarea.addEventListener('input', updateCounter);
    updateCounter(); // 초기화
}

// 새 댓글 추가 후 이미지 처리
async function processNewCommentImages(commentId) {
    const commentsContainer = document.querySelector('#comments-container');
    if (!commentsContainer) return;

    // commentId가 있으면 해당 댓글만 처리, 없으면 마지막 댓글 처리
    setTimeout(() => {
        let commentElement;

        if (commentId) {
            commentElement = commentsContainer.querySelector(`[data-new-comment-id="${commentId}"]`);
        } else {
            // 마지막 댓글 선택 (대안)
            const allComments = document.querySelectorAll('#comments-container .comment-item');
            if (allComments.length > 0) {
                commentElement = allComments[allComments.length - 1];
            }
        }

        if (commentElement && typeof setupSpecificImages === 'function') {
            setupSpecificImages(commentElement);
        }
    }, 100); // DOM 업데이트 후 처리하기 위한 지연
}