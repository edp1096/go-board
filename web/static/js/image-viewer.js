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
 * 본문 내 이미지를 처리하는 함수
 */
function setupContentImages() {
    const postContent = document.getElementById('post-content');
    if (!postContent) return;

    // 본문 내 모든 이미지에 이벤트 리스너 추가
    const images = postContent.querySelectorAll('img');

    images.forEach(img => {
        // 원본 이미지 URL 저장
        const originalSrc = img.src;

        // 썸네일이 아닌 경우에만 썸네일로 변경
        if (!originalSrc.includes('/thumbs/')) {
            // 썸네일 URL로 변경
            img.src = convertToThumbnailUrl(originalSrc);

            // 원본 이미지 URL을 data 속성에 저장
            img.dataset.originalSrc = originalSrc;

            // 이미지에 클릭 이벤트 추가
            img.addEventListener('click', function () {
                // 이미지 뷰어 컴포넌트 호출
                const imageViewer = document.getElementById('image-viewer');
                if (imageViewer && typeof Alpine !== 'undefined') {
                    const viewerData = Alpine.$data(imageViewer);
                    if (viewerData) {
                        viewerData.openImageViewer(originalSrc, img.alt);
                    }
                }
            });

            // 스타일 및 클래스 추가
            img.classList.add('cursor-pointer', 'hover:opacity-90');
            img.title = '클릭하면 원본 이미지를 볼 수 있습니다';
        }
    });
}

/**
 * 댓글 내 이미지를 처리하는 함수
 */
function setupCommentImages() {
    // Alpine의 변경 감지를 이용해 댓글 로드 후 이미지 처리
    document.addEventListener('comments-loaded', function () {
        // 댓글 컨테이너
        const commentsContainer = document.querySelector('.space-y-6');
        if (!commentsContainer) return;

        // 모든 댓글 내 이미지 찾기
        const commentImages = commentsContainer.querySelectorAll('.text-sm.text-gray-700 img');

        commentImages.forEach(img => {
            const originalSrc = img.src;

            if (!originalSrc.includes('/thumbs/')) {
                // 썸네일 URL로 변경
                img.src = convertToThumbnailUrl(originalSrc);

                // 원본 이미지 URL을 data 속성에 저장
                img.dataset.originalSrc = originalSrc;

                // 이미지에 클릭 이벤트 추가
                img.addEventListener('click', function () {
                    // 이미지 뷰어 컴포넌트 호출
                    const imageViewer = document.getElementById('image-viewer');
                    if (imageViewer && typeof Alpine !== 'undefined') {
                        const viewerData = Alpine.$data(imageViewer);
                        if (viewerData) {
                            viewerData.openImageViewer(originalSrc, img.alt);
                        }
                    }
                });

                // 스타일 및 클래스 추가
                img.classList.add('cursor-pointer', 'hover:opacity-90');
                img.title = '클릭하면 원본 이미지를 볼 수 있습니다';
            }
        });
    });

    // 댓글이 로드되었을 때 이벤트 발생
    // Alpine 데이터가 변경될 때 감시
    if (typeof Alpine !== 'undefined') {
        // 댓글 관련 컴포넌트 초기화 후 처리
        const checkComments = function () {
            const commentsSystem = document.querySelector('[x-data="commentSystem"]');
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