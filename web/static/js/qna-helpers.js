// qna-helpers.js
// Q&A 게시판 기능을 위한 유틸리티 함수

// 질문 투표 함수
async function voteQuestion(direction) {
    const boardId = document.getElementById('boardId').value;
    const postId = document.getElementById('postId').value;

    // CSRF 토큰 가져오기
    const csrfToken = document.querySelector('meta[name="csrf-token"]').content;

    try {
        const response = await fetch(`/api/boards/${boardId}/qnas/${postId}/vote`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': csrfToken
            },
            body: JSON.stringify({ direction: direction })
        });

        const data = await response.json();

        if (data.success) {
            // 페이지 새로고침 대신 투표 수 표시 요소만 업데이트
            document.getElementById('question-vote-count').textContent = data.voteCount;
        } else {
            alert('투표 실패: ' + data.message);
        }
    } catch (error) {
        console.error('투표 처리 중 오류:', error);
        alert('투표 처리 중 오류가 발생했습니다.');
    }
}

// 질문 상태 변경 함수
async function changeQuestionStatus(status) {
    const boardId = document.getElementById('boardId').value;
    const postId = document.getElementById('postId').value;

    // CSRF 토큰 가져오기
    const csrfToken = document.querySelector('meta[name="csrf-token"]').content;

    try {
        const response = await fetch(`/api/boards/${boardId}/qnas/${postId}/status`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': csrfToken
            },
            body: JSON.stringify({ status: status })
        });

        const data = await response.json();

        if (data.success) {
            // 페이지 새로고침하여 상태 변경 반영
            window.location.reload();
        } else {
            alert('상태 변경 실패: ' + data.message);
        }
    } catch (error) {
        console.error('상태 변경 중 오류:', error);
        alert('상태 변경 중 오류가 발생했습니다.');
    }
}

// 투표 수 조회 함수
async function getQuestionVoteCount() {
    const boardId = document.getElementById('boardId').value;
    const postId = document.getElementById('postId').value;

    try {
        const response = await fetch(`/api/boards/${boardId}/qnas/${postId}/vote-count`);
        const data = await response.json();

        if (data.success) {
            // 투표 수 업데이트
            document.getElementById('question-vote-count').textContent = data.voteCount;
        }
    } catch (error) {
        console.error('투표 수 조회 오류:', error);
    }
}

// 페이지 로드 시 투표 수 조회
document.addEventListener('DOMContentLoaded', function () {
    getQuestionVoteCount();
});

// 이미지 뷰어 컴포넌트
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
});