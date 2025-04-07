// /test/front/setup.js
const { chromium } = require('playwright');

/**
 * Go 게시판 애플리케이션의 초기 설정 및 게시물 생성 자동화
 */
async function setupGoBoardApplication() {
    // 브라우저 실행
    const browser = await chromium.launch({
        headless: false, // 자동화 과정을 시각적으로 확인하기 위해 headless 모드를 끕니다
        slowMo: 1000 // 각 액션 사이에 1초 지연을 주어 과정을 더 잘 볼 수 있게 합니다
    });

    const context = await browser.newContext();
    const page = await context.newPage();

    try {
        // 애플리케이션 초기 페이지 로드
        console.log('애플리케이션에 접속합니다...');
        await page.goto('http://localhost:3000');
        await page.waitForLoadState('networkidle');

        // 현재 URL 확인
        const currentUrl = page.url();
        console.log(`현재 URL: ${currentUrl}`);

        // 먼저 로그인 시도
        console.log('로그인을 시도합니다...');

        // 로그인 페이지로 이동
        let loginLinkVisible = await page.isVisible('text=로그인');

        if (loginLinkVisible) {
            console.log('로그인 링크가 발견되었습니다. 클릭합니다.');
            await page.click('text=로그인');
            await page.waitForLoadState('networkidle');
        } else {
            // 직접 로그인 페이지로 이동
            console.log('로그인 링크가 보이지 않습니다. 직접 로그인 페이지로 이동합니다.');
            await page.goto('http://localhost:3000/auth/login');
            await page.waitForLoadState('networkidle');
        }

        // 로그인 시도
        console.log('관리자 계정으로 로그인합니다...');
        await page.fill('#username', 'admin');
        await page.fill('#password', 'Admin123!');
        await page.click('button[type="submit"]');

        // 로그인 실패 시(로그인 버튼이 여전히 존재) 관리자 계정 생성이 필요할 수 있음
        await page.waitForLoadState('networkidle');

        // 로그인 성공 여부 확인
        const stillOnLoginPage = page.url().includes('/auth/login');

        if (stillOnLoginPage) {
            console.log('로그인 실패. 관리자 계정 설정이 필요합니다.');

            // 초기 설정 페이지로 이동 시도
            await page.goto('http://localhost:3000/admin/setup');
            await page.waitForLoadState('networkidle');

            // 관리자 계정 설정 페이지인지 확인
            if (page.url().includes('/admin/setup')) {
                console.log('관리자 계정을 설정합니다...');

                // 관리자 계정 정보 입력
                await page.fill('#username', 'admin');
                await page.fill('#email', 'admin@example.com');
                await page.fill('#password', 'Admin123!');
                await page.fill('#password_confirm', 'Admin123!');
                await page.fill('#full_name', '관리자');

                // 관리자 계정 생성 버튼 클릭
                await page.click('button[type="submit"]');

                // 페이지 로딩 대기
                await page.waitForLoadState('networkidle');
                console.log('관리자 계정이 생성되었습니다.');

                // 다시 로그인 시도
                console.log('생성된 관리자 계정으로 로그인합니다...');
                await page.fill('#username', 'admin');
                await page.fill('#password', 'Admin123!');
                await page.click('button[type="submit"]');
                await page.waitForLoadState('networkidle');
            } else {
                throw new Error('관리자 계정 설정 페이지에 접근할 수 없습니다.');
            }
        }

        console.log('로그인 성공! 현재 URL:', page.url());

        // 관리자 대시보드로 이동
        console.log('관리자 대시보드로 이동합니다...');
        await page.goto('http://localhost:3000/admin');
        await page.waitForLoadState('networkidle');

        // 게시판 관리 페이지로 이동
        console.log('게시판 관리 페이지로 이동합니다...');
        const boardManageLink = await page.locator('text=게시판 관리하기').first();
        await boardManageLink.click();
        await page.waitForLoadState('networkidle');

        // 현재 URL 확인
        console.log('게시판 관리 페이지 URL:', page.url());

        // 새 게시판 만들기
        console.log('새 게시판을 생성합니다...');
        const newBoardButton = await page.locator('text=새 게시판 만들기').first();
        await newBoardButton.click();
        await page.waitForLoadState('networkidle');

        // 게시판 정보 입력
        await page.fill('#name', '일반 게시판');
        await page.fill('#description', '일반적인 게시물을 작성할 수 있는 게시판입니다.');
        await page.selectOption('#board_type', 'normal');

        // 게시판 옵션 설정
        await page.check('#comments_enabled'); // 댓글 기능 활성화
        await page.check('#allow_anonymous'); // 익명 접근 허용

        // 게시판 생성
        await page.locator('button:has-text("등록")').click();
        await page.waitForLoadState('networkidle');
        console.log('게시판이 생성되었습니다.');

        // 생성한 게시판 보기
        console.log('생성한 게시판으로 이동합니다...');
        await page.locator('text=보기').first().click();
        await page.waitForLoadState('networkidle');

        // 게시물 작성
        console.log('새 게시물을 작성합니다...');
        await page.locator('text=글쓰기').first().click();
        await page.waitForLoadState('networkidle');

        // 게시물 정보 입력
        await page.fill('#title', '첫 번째 게시물');

        // 에디터가 Shadow DOM 안에 있어 특별한 처리가 필요합니다
        try {
            const editorContainer = await page.$('editor-other#editor');

            if (editorContainer) {
                const shadowRoot = await editorContainer.evaluateHandle(el => el.shadowRoot);
                const editorElement = await shadowRoot.evaluateHandle(root => root.querySelector('#editor-shadow-dom'));

                // 에디터에 내용 입력
                await editorElement.evaluate(editor => {
                    editor.innerHTML = '<p>안녕하세요! 첫 번째 게시물입니다.</p><p>이 게시물은 Playwright를 사용하여 자동으로 생성되었습니다.</p>';
                });
                console.log('에디터에 내용을 입력했습니다.');
            } else {
                console.log('에디터 요소를 찾을 수 없습니다. 다른 방법을 시도합니다...');

                // 대안: 컨텐츠를 hidden input에 직접 설정
                await page.evaluate(() => {
                    document.getElementById('content').value = '<p>안녕하세요! 첫 번째 게시물입니다.</p><p>이 게시물은 Playwright를 사용하여 자동으로 생성되었습니다.</p>';
                });
                console.log('hidden input에 직접 내용을 설정했습니다.');
            }
        } catch (error) {
            console.log('에디터 접근 중 오류 발생:', error);
            console.log('대체 방법으로 시도합니다...');

            // 대안: 컨텐츠를 hidden input에 직접 설정
            await page.evaluate(() => {
                document.getElementById('content').value = '<p>안녕하세요! 첫 번째 게시물입니다.</p><p>이 게시물은 Playwright를 사용하여 자동으로 생성되었습니다.</p>';
            });
            console.log('hidden input에 직접 내용을 설정했습니다.');
        }

        // 게시물 등록
        console.log('게시물을 등록합니다...');
        await page.locator('button[type="submit"]').click();
        await page.waitForLoadState('networkidle');
        console.log('첫 게시물이 성공적으로 작성되었습니다.');

        // 게시물 내용 확인
        const postTitle = await page.textContent('h1');
        console.log(`작성된 게시물 제목: ${postTitle}`);

        // 로그아웃
        console.log('로그아웃 합니다...');

        // 헤더의 사용자 메뉴 클릭
        await page.locator('text=admin').first().click();
        await page.locator('text=로그아웃').click();
        await page.waitForLoadState('networkidle');

        console.log('모든 과정이 완료되었습니다!');

    } catch (error) {
        console.error('자동화 과정 중 오류가 발생했습니다:', error);

        // 현재 URL 출력
        try {
            console.log(`오류 발생 시점의 URL: ${page.url()}`);

            // 현재 페이지의 스크린샷 저장
            await page.screenshot({ path: 'error-screenshot.png' });
            console.log('오류 스크린샷이 저장되었습니다: error-screenshot.png');

            // 페이지 HTML 저장
            const html = await page.content();
            require('fs').writeFileSync('error-page.html', html);
            console.log('오류 페이지 HTML이 저장되었습니다: error-page.html');
        } catch (e) {
            console.error('디버깅 정보 저장 중 오류:', e);
        }
    } finally {
        // 5초 후 브라우저 종료
        await new Promise(resolve => setTimeout(resolve, 5000));
        await browser.close();
    }
}

// 스크립트 실행
setupGoBoardApplication().catch(console.error);