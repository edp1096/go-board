/* web/static/js/image-viewer.js */
// 이미지 뷰어

// Alpine 컴포넌트 먼저 정의
if (typeof Alpine !== 'undefined') {
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
}

document.addEventListener('DOMContentLoaded', function () {
    if (typeof Alpine === 'undefined') {
        document.addEventListener('alpine:init', function () {
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

            // Alpine 초기화 후 이미지 처리
            setupContentImages();
            setupCommentImages();
        });
    } else {
        // 이미 Alpine이 로드됐다면 바로 처리
        setupContentImages();
        setupCommentImages();
    }
});

/**
 * 레이어팝업 표시를 위한 이미지 처리 함수
 * @param {HTMLElement} container - 이미지를 포함하는 컨테이너 요소
 */
function setupSpecificImages(container) {
    if (!container) return;

    // 아직 처리되지 않은 이미지만 선택
    const images = container.querySelectorAll('img:not([data-viewer-processed="true"])');

    images.forEach(img => {
        const originalSrc = img.src;

        // 애니메이션 webp 이미지인지 확인
        const isAnimatedWebp = originalSrc.toLowerCase().endsWith('.webp') &&
            img.getAttribute('animate') === 'true';

        // 이미지가 이미 처리되었는지 표시
        img.setAttribute('data-viewer-processed', 'true');

        // 썸네일이 아니고 애니메이션 webp가 아닌 경우에만 썸네일로 변경
        if (!originalSrc.includes('/thumbs/') && !isAnimatedWebp) {
            img.src = convertToThumbnailUrl(originalSrc);
            img.dataset.originalSrc = originalSrc;
        }

        // 모든 이미지에 클릭 이벤트 추가 (이미 처리된 이미지는 위에서 필터링됨)
        img.addEventListener('click', function () {
            const imageViewer = document.getElementById('image-viewer');
            if (imageViewer && typeof Alpine !== 'undefined') {
                const viewerData = Alpine.$data(imageViewer);
                if (viewerData) {
                    viewerData.openImageViewer(originalSrc, img.alt);
                }
            }
        });

        img.classList.add('cursor-pointer', 'hover:opacity-90');
        img.title = '클릭하면 원본 이미지를 볼 수 있습니다';
    });
}

/**
 * 본문 내 이미지를 처리하는 함수
 */
function setupContentImages() {
    const postContent = document.getElementById('post-content');
    if (!postContent) { return; }

    setupSpecificImages(postContent);
}

/**
 * 댓글 내 이미지를 처리하는 함수
 */
function setupCommentImages() {
    document.addEventListener('comments-loaded', function () {
        const commentsContainer = document.querySelector('.space-y-6');
        if (!commentsContainer) { return; }

        setupSpecificImages(commentsContainer);
    });

    // 댓글이 로드되었을 때 이벤트 발생
    // Alpine 데이터가 변경될 때 감시
    if (typeof Alpine !== 'undefined') {
        // 댓글 관련 컴포넌트 초기화 후 처리
        const checkComments = function () {
            // const commentsSystem = document.querySelector('[x-data="commentSystem"]');
            const commentsSystem = document.querySelector('#comments-container');
            if (commentsSystem) {
                try {
                    const commentsData = Alpine.$data(commentsSystem);
                    if (commentsData && !commentsData.loading) {
                        // 댓글 로드 완료 이벤트 발생
                        document.dispatchEvent(new CustomEvent('comments-loaded'));
                        return true;
                    }
                } catch (e) {
                    console.warn("댓글 시스템 데이터 접근 오류:", e);
                }
            }
            return false;
        };

        // 처음 한 번 체크
        if (!checkComments()) {
            // 필요하면 폴링
            const commentsInterval = setInterval(() => {
                if (checkComments()) {
                    clearInterval(commentsInterval);
                }
            }, 500);

            // 최대 10초만 시도
            setTimeout(() => {
                clearInterval(commentsInterval);
            }, 10000);
        }
    }
}

/**
 * 새 댓글 추가 후 이미지 처리
 * @param {number} commentId - 새로 추가된 댓글 ID
 */
function processNewCommentImages(commentId) {
    const commentsContainer = document.querySelector('#comments-container');

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

        console.log(commentElement, commentId);
        if (commentElement) {
            setupSpecificImages(commentElement);
        }
    }, 100); // DOM 업데이트 후 처리하기 위한 지연
}

/**
 * 이미지 URL을 썸네일 URL로 변환
 */
function convertToThumbnailUrl(url) {
    if (!url) return url;

    // 이미 썸네일 URL인 경우 그대로 반환
    if (url.includes('/thumbs/')) {
        return url;
    }

    // URL 경로 분해
    const urlParts = url.split('/');
    let filename = urlParts.pop();

    // WebP 파일인 경우 JPG 확장자로 변경
    if (filename.toLowerCase().endsWith('.webp')) {
        filename = filename.substring(0, filename.length - 5) + '.jpg';
    }

    // 썸네일 URL 생성 (filename 앞에 thumbs 폴더 추가)
    urlParts.push('thumbs');
    urlParts.push(filename);

    return urlParts.join('/');
}

/**
 * URL에서 원본 이미지 URL로 변환
 */
function convertToOriginalUrl(url) {
    if (!url) return url;

    // 썸네일 URL이 아닌 경우 그대로 반환
    if (!url.includes('/thumbs/')) {
        return url;
    }

    // /thumbs/ 부분 제거
    return url.replace('/thumbs/', '/');
}