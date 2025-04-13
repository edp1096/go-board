/* gallery-helpers.js */
// 갤러리 게시판 전용 헬퍼 함수

document.addEventListener('DOMContentLoaded', function () {
    setupGalleryImages();
    setupLazyLoading();
});

/**
 * 갤러리 이미지 초기화 및 이벤트 핸들러 연결
 */
function setupGalleryImages() {
    const galleryItems = document.querySelectorAll('.gallery-item');

    galleryItems.forEach(item => {
        // 이미지 호버 효과
        const image = item.querySelector('.gallery-image');
        if (image) {
            item.addEventListener('mouseenter', () => {
                image.classList.add('scale-105');
                image.style.transition = 'transform 0.3s ease';
            });

            item.addEventListener('mouseleave', () => {
                image.classList.remove('scale-105');
            });

            // 이미지 클릭 이벤트 - 원본 이미지 보기
            image.addEventListener('click', (e) => {
                // 부모 링크의 기본 동작 방지 (상세 페이지로 이동)하고 이미지 뷰어 열기
                if (e.ctrlKey || e.metaKey) {
                    // ctrl/cmd + 클릭은 새 탭에서 열기 기본 동작 유지
                    return;
                }

                e.preventDefault();
                e.stopPropagation();

                // 원본 이미지 URL이 data-original 속성에 있으면 사용, 없으면 현재 src 사용
                const originalSrc = image.dataset.original || image.src;
                const imageAlt = image.alt || '갤러리 이미지';

                // 이미지 뷰어가 있으면 열기
                const imageViewer = document.getElementById('image-viewer');
                if (imageViewer && typeof Alpine !== 'undefined') {
                    try {
                        const viewerData = Alpine.$data(imageViewer);
                        if (viewerData && typeof viewerData.openImageViewer === 'function') {
                            viewerData.openImageViewer(originalSrc, imageAlt);
                        }
                    } catch (error) {
                        console.error('이미지 뷰어 열기 오류:', error);
                    }
                }
            });
        }
    });
}

/**
 * 이미지 지연 로딩 설정 
 * - 화면에 보이는 이미지만 로드하여 초기 로딩 성능 개선
 */
function setupLazyLoading() {
    // 모던 브라우저에서 지원하는 Intersection Observer API 사용
    if ('IntersectionObserver' in window) {
        const lazyImages = document.querySelectorAll('.lazy-image');

        const imageObserver = new IntersectionObserver((entries, observer) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    const image = entry.target;

                    // 데이터 속성에서 실제 이미지 URL 가져와서 설정
                    const src = image.dataset.src;
                    if (src) {
                        image.src = src;
                        image.classList.remove('lazy-image');
                        image.classList.add('loaded');

                        // 이미지가 로드된 후 observer에서 제거
                        observer.unobserve(image);
                    }
                }
            });
        });

        lazyImages.forEach(image => {
            imageObserver.observe(image);
        });
    } else {
        // Intersection Observer를 지원하지 않는 브라우저를 위한 폴백
        // 간단한 스크롤 이벤트 기반 지연 로딩
        let lazyLoadThrottleTimeout;

        function lazyLoad() {
            if (lazyLoadThrottleTimeout) {
                clearTimeout(lazyLoadThrottleTimeout);
            }

            lazyLoadThrottleTimeout = setTimeout(() => {
                const scrollTop = window.pageYOffset;
                const lazyImages = document.querySelectorAll('.lazy-image');

                lazyImages.forEach(img => {
                    if (img.offsetTop < window.innerHeight + scrollTop) {
                        const src = img.dataset.src;
                        if (src) {
                            img.src = src;
                            img.classList.remove('lazy-image');
                            img.classList.add('loaded');
                        }
                    }
                });

                if (document.querySelectorAll('.lazy-image').length === 0) {
                    document.removeEventListener('scroll', lazyLoad);
                    window.removeEventListener('resize', lazyLoad);
                    window.removeEventListener('orientationChange', lazyLoad);
                }
            }, 20);
        }

        document.addEventListener('scroll', lazyLoad);
        window.addEventListener('resize', lazyLoad);
        window.addEventListener('orientationChange', lazyLoad);

        // 초기 로드
        lazyLoad();
    }
}

/**
 * 갤러리 필터링 함수
 * @param {string} category - 필터링할 카테고리
 */
function filterGallery(category) {
    const galleryItems = document.querySelectorAll('.gallery-item');

    // 모든 필터 버튼에서 활성 클래스 제거
    document.querySelectorAll('.gallery-filter-btn').forEach(btn => {
        btn.classList.remove('active', 'bg-blue-500', 'text-white');
        btn.classList.add('bg-gray-200', 'text-gray-700');
    });

    // 선택된 필터 버튼에 활성 클래스 추가
    const activeBtn = document.querySelector(`.gallery-filter-btn[data-category="${category}"]`);
    if (activeBtn) {
        activeBtn.classList.remove('bg-gray-200', 'text-gray-700');
        activeBtn.classList.add('active', 'bg-blue-500', 'text-white');
    }

    // 갤러리 아이템 필터링
    galleryItems.forEach(item => {
        if (category === 'all' || item.dataset.category === category) {
            item.style.display = '';
            // 부드러운 애니메이션 적용
            setTimeout(() => {
                item.style.opacity = '1';
                item.style.transform = 'scale(1)';
            }, 50);
        } else {
            item.style.opacity = '0';
            item.style.transform = 'scale(0.8)';
            setTimeout(() => {
                item.style.display = 'none';
            }, 300); // 애니메이션 시간과 맞춤
        }
    });
}