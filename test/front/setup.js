// /test/front/setup.js
const { chromium } = require('playwright');
const fs = require('fs');
const path = require('path');

/**
 * Go 게시판 애플리케이션의 초기 설정 및 게시물 생성 자동화
 */
async function setupToyBoardApplication() {
    // 로그 디렉토리 생성
    const logDir = path.join(__dirname, 'log');
    if (!fs.existsSync(logDir)) {
        fs.mkdirSync(logDir, { recursive: true });
        console.log('로그 디렉토리를 생성했습니다:', logDir);
    }

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

        // 관리자 설정 페이지인 경우 계정 생성
        if (currentUrl.includes('/admin/setup')) {
            console.log('초기 설정 페이지로 리디렉션되었습니다. 관리자 계정을 설정합니다...');
            await createAdminAccount(page);
        }

        // 로그인 시도
        console.log('로그인 시도합니다...');
        const loginSuccess = await loginAsAdmin(page);

        if (!loginSuccess) {
            throw new Error('로그인에 실패했습니다.');
        }

        console.log('로그인 성공 확인. 게시판 설정을 진행합니다.');

        // 관리자 대시보드로 이동
        console.log('관리자 대시보드로 이동합니다...');
        await page.goto('http://localhost:3000/admin');
        await page.waitForLoadState('networkidle');

        // 게시판 생성 또는 기존 게시판 선택
        await createOrSelectBoard(page);

        // 게시물 작성
        await createFirstPost(page);

        // 로그아웃
        await logout(page);

        console.log('모든 과정이 완료되었습니다!');

    } catch (error) {
        console.error('자동화 과정 중 오류가 발생했습니다:', error);

        // 현재 URL 출력 및 디버깅 정보 저장
        try {
            console.log(`오류 발생 시점의 URL: ${page.url()}`);
            const screenshotPath = path.join(logDir, 'error-screenshot.png');
            await page.screenshot({ path: screenshotPath });
            console.log('오류 스크린샷이 저장되었습니다:', screenshotPath);

            const htmlPath = path.join(logDir, 'error-page.html');
            const html = await page.content();
            fs.writeFileSync(htmlPath, html);
            console.log('오류 페이지 HTML이 저장되었습니다:', htmlPath);
        } catch (e) {
            console.error('디버깅 정보 저장 중 오류:', e);
        }
    } finally {
        // 1초 후 브라우저 종료
        await new Promise(resolve => setTimeout(resolve, 1000));
        await browser.close();
    }
}

/**
 * 관리자 계정 생성
 */
async function createAdminAccount(page) {
    console.log('관리자 계정 설정 페이지에서 계정을 생성합니다...');

    // 관리자 계정 정보 입력
    await page.fill('#username', 'admin');

    // 이메일 필드가 표시될 때까지 대기 후 입력
    await page.waitForSelector('#email', { state: 'visible', timeout: 1000 });
    await page.fill('#email', 'admin@example.com');
    console.log('이메일 입력 완료: admin@example.com');

    await page.fill('#password', 'Admin123!');
    await page.fill('#password_confirm', 'Admin123!');
    await page.fill('#full_name', '관리자');

    // 디버깅: 입력 필드 상태 확인
    const emailValue = await page.$eval('#email', el => el.value);
    console.log(`이메일 필드 입력값: ${emailValue}`);

    // 관리자 계정 생성 버튼 클릭
    await page.click('button[type="submit"]');
    await page.waitForLoadState('networkidle');
    console.log('관리자 계정이 생성되었습니다. 현재 URL:', page.url());
}

/**
 * 관리자로 로그인
 */
async function loginAsAdmin(page) {
    // 현재 로그인 페이지가 아니면 로그인 페이지로 이동
    if (!page.url().includes('/auth/login')) {
        console.log('로그인 페이지로 이동합니다...');
        await page.goto('http://localhost:3000/auth/login');
        await page.waitForLoadState('networkidle');
    }

    console.log('관리자 계정으로 로그인합니다...');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'Admin123!');

    // 로그인 버튼 클릭
    await page.click('button[type="submit"]');
    await page.waitForLoadState('networkidle');

    // 로그인 성공 여부 확인 (로그인 페이지를 벗어났는지)
    const loginSuccess = !page.url().includes('/auth/login');

    if (loginSuccess) {
        console.log('로그인 성공! 현재 URL:', page.url());
        return true;
    } else {
        console.log('로그인 실패했습니다. 현재 URL:', page.url());
        return false;
    }
}

/**
 * 게시판 생성 또는 기존 게시판 선택
 */
async function createOrSelectBoard(page) {
    const logDir = path.join(__dirname, 'log');

    // 게시판 관리 페이지로 이동
    console.log('게시판 관리 페이지로 이동합니다...');
    await page.goto('http://localhost:3000/admin/boards');
    await page.waitForLoadState('networkidle');
    console.log('게시판 관리 페이지 URL:', page.url());

    // 기존 '일반 게시판' 확인
    console.log('기존 게시판을 확인합니다...');
    const boardName = page.locator('text=일반 게시판').first();

    // 게시판이 이미 존재하는 경우
    if (await boardName.count() > 0) {
        console.log('일반 게시판이 이미 존재합니다. 바로 게시판으로 이동합니다.');

        // 해당 게시판 행에서 보기 버튼 찾기
        const row = await boardName.locator('..').locator('..'); // 상위 요소로 이동하여 행 찾기
        const viewButton = row.locator('text=보기');

        if (await viewButton.count() === 0) {
            // 대안 시도: 페이지 내 모든 '보기' 버튼 중 첫 번째 클릭
            const allViewButtons = page.locator('text=보기');
            if (await allViewButtons.count() > 0) {
                console.log('보기 버튼을 찾았습니다.');
                await allViewButtons.first().click();
            } else {
                console.log('보기 버튼을 찾을 수 없습니다. 게시판 목록 페이지의 스크린샷을 저장합니다.');
                const screenshotPath = path.join(logDir, 'boards-list.png');
                await page.screenshot({ path: screenshotPath, fullPage: true });
                throw new Error('보기 버튼을 찾을 수 없습니다.');
            }
        } else {
            await viewButton.click();
        }

        await page.waitForLoadState('networkidle');
        return;
    }

    // 게시판이 존재하지 않는 경우 새로 생성
    console.log('일반 게시판이 존재하지 않습니다. 새 게시판을 생성합니다...');

    // 새 게시판 만들기 버튼 확인 및 클릭
    const newBoardButton = page.locator('text=새 게시판 만들기');

    if (await newBoardButton.count() === 0) {
        console.error('새 게시판 만들기 버튼을 찾을 수 없습니다.');
        throw new Error('새 게시판 만들기 버튼을 찾을 수 없습니다.');
    }

    console.log('새 게시판 만들기 버튼을 클릭합니다...');
    await newBoardButton.click();
    await page.waitForLoadState('networkidle');

    // 현재 페이지가 새 게시판 생성 페이지인지 확인
    console.log('현재 URL:', page.url());

    // 게시판 정보 입력
    console.log('게시판 정보를 입력합니다...');
    await page.fill('#name', '일반 게시판');
    await page.fill('#description', '일반적인 게시물을 작성할 수 있는 게시판입니다.');

    // 게시판 유형 선택 (select 요소)
    const boardTypeSelect = page.locator('#board_type');
    if (await boardTypeSelect.count() > 0) {
        await boardTypeSelect.selectOption('normal');
    } else {
        console.log('게시판 유형 선택 요소를 찾을 수 없습니다.');
    }

    // 체크박스 선택
    const commentsEnabled = page.locator('#comments_enabled');
    const allowAnonymous = page.locator('#allow_anonymous');

    if (await commentsEnabled.count() > 0) {
        await commentsEnabled.check();
    } else {
        console.log('댓글 활성화 체크박스를 찾을 수 없습니다.');
    }

    if (await allowAnonymous.count() > 0) {
        await allowAnonymous.check();
    } else {
        console.log('익명 접근 허용 체크박스를 찾을 수 없습니다.');
    }

    if (await allowPrivate.count() > 0) {
        await allowPrivate.check();
    } else {
        console.log('비밀글 허용 체크박스를 찾을 수 없습니다.');
    }

    // 페이지의 모든 버튼 텍스트 로깅 (디버깅용)
    const buttons = await page.locator('button').all();
    console.log('페이지의 버튼들:');
    for (let i = 0; i < buttons.length; i++) {
        const text = await buttons[i].textContent();
        console.log(`버튼 ${i + 1}: "${text.trim()}"`);
    }

    // 등록 버튼 찾기 및 클릭
    console.log('등록 버튼을 찾아 클릭합니다...');

    // 다양한 선택자로 시도
    const submitButtonSelectors = [
        'button[type="submit"]',
        'button:has-text("등록")',
        'button:has-text("저장")',
        'button:has-text("생성")',
        'button.primary',
        'input[type="submit"]'
    ];

    let buttonClicked = false;

    for (const selector of submitButtonSelectors) {
        const button = page.locator(selector);
        if (await button.count() > 0) {
            console.log(`선택자 "${selector}"로 버튼을 찾았습니다.`);
            await button.scrollIntoViewIfNeeded();

            const buttonScreenshotPath = path.join(logDir, 'found-button.png');
            await button.screenshot({ path: buttonScreenshotPath });
            console.log('버튼 스크린샷을 저장했습니다:', buttonScreenshotPath);

            try {
                await button.click({ timeout: 1000 });
                buttonClicked = true;
                console.log('버튼을 클릭했습니다.');
                break;
            } catch (e) {
                console.log(`버튼 클릭 실패: ${e.message}`);
            }
        }
    }

    if (!buttonClicked) {
        console.error('어떤 등록 버튼도 찾거나 클릭할 수 없습니다.');
        // 페이지 스크린샷 저장
        const formScreenshotPath = path.join(logDir, 'form-page.png');
        await page.screenshot({ path: formScreenshotPath, fullPage: true });
        console.log('전체 페이지 스크린샷을 저장했습니다:', formScreenshotPath);
        throw new Error('등록 버튼을 찾을 수 없습니다.');
    }

    // 페이지 전환 대기
    await page.waitForLoadState('networkidle');
    console.log('등록 버튼 클릭 후 현재 URL:', page.url());

    // 게시판 목록 페이지로 돌아갔는지 확인
    if (page.url().includes('/admin/boards')) {
        console.log('게시판이 성공적으로 생성되었습니다.');

        // 생성한 게시판의 보기 버튼 찾기 및 클릭
        console.log('생성한 게시판의 보기 버튼을 찾습니다...');
        const viewButton = page.locator('text=보기').first();

        if (await viewButton.count() === 0) {
            console.log('보기 버튼을 찾을 수 없습니다. 게시판 목록 페이지의 스크린샷을 저장합니다.');
            const boardsScreenshotPath = path.join(logDir, 'boards-list.png');
            await page.screenshot({ path: boardsScreenshotPath, fullPage: true });
            throw new Error('보기 버튼을 찾을 수 없습니다.');
        }

        console.log('보기 버튼을 클릭합니다...');
        await viewButton.click();
        await page.waitForLoadState('networkidle');
    } else {
        console.log('알림: 게시판 생성 후 예상 페이지로 이동하지 않았습니다.');
    }
}

/**
 * 첫 게시물 작성
 */
async function createFirstPost(page) {
    const logDir = path.join(__dirname, 'log');

    // 게시판 페이지에서 글쓰기 버튼 클릭
    console.log('글쓰기 버튼을 찾습니다...');
    const writeButton = page.locator('text=글쓰기').first();

    if (await writeButton.count() === 0) {
        console.log('글쓰기 버튼을 찾을 수 없습니다. 현재 페이지 스크린샷을 저장합니다.');
        const boardScreenshotPath = path.join(logDir, 'board-page.png');
        await page.screenshot({ path: boardScreenshotPath, fullPage: true });
        throw new Error('글쓰기 버튼을 찾을 수 없습니다.');
    }

    console.log('글쓰기 버튼을 클릭합니다...');
    await writeButton.click();
    await page.waitForLoadState('networkidle');

    // 게시물 제목 입력
    console.log('게시물 제목을 입력합니다...');
    await page.fill('#title', '첫 번째 게시물');

    // 현재 시간을 포함한 게시물 내용 생성
    const timestamp = new Date().toLocaleString('ko-KR');
    const postContent = `<p>안녕하세요! 첫 번째 게시물입니다.</p>
<p>이 게시물은 Playwright를 사용하여 자동으로 생성되었습니다.</p>
<p>생성 시간: ${timestamp}</p>
<p>이 게시물은 자동 테스트를 위해 작성되었습니다.</p>`;

    // 에디터 내용 입력 시도
    console.log('에디터에 내용을 입력합니다...');
    let contentInputSuccess = false;

    // 에디터 내용 입력 - 자연스러운 입력 방식
    console.log('에디터에 자연스러운 방식으로 내용을 입력합니다...');

    try {
        // 현재 시간을 포함한 게시물 내용 준비
        const timestamp = new Date().toLocaleString('ko-KR');
        const paragraphs = [
            '안녕하세요! 첫 번째 게시물입니다.',
            '이 게시물은 Playwright를 사용하여 자동으로 생성되었습니다.',
            `생성 시간: ${timestamp}`,
            '이 게시물은 자동 테스트를 위해 작성되었습니다.'
        ];

        // 1. ProseMirror 편집 영역 찾기 (이미지에서 보이는 에디터 영역)
        const editorArea = page.locator('[contenteditable="true"]').first();

        if (await editorArea.count() > 0) {
            console.log('편집 가능한 영역을 찾았습니다. 클릭합니다...');

            // 에디터 영역 클릭하여 포커스 맞추기
            await editorArea.click();

            // 잠시 대기 (실제 사용자처럼 포커스 후 타이핑 시작까지 시간 간격)
            await page.waitForTimeout(50);

            // 자연스러운 타이핑으로 단락별 입력
            for (let i = 0; i < paragraphs.length; i++) {
                // 타이핑 시작 지연 (인간 사용자 시뮬레이션)
                await page.waitForTimeout(180);

                // 약간의 지연으로 자연스러운 타이핑 효과
                await page.keyboard.type(paragraphs[i], { delay: 15 });

                // 마지막 단락이 아니면 줄바꿈 추가
                if (i < paragraphs.length - 1) {
                    await page.keyboard.press('Enter');
                    await page.keyboard.press('Enter');
                    await page.waitForTimeout(100);  // 단락 입력 후 잠시 대기
                }
            }

            console.log('자연스러운 방식으로 모든 내용을 입력했습니다.');

            // 편집 영역에서 포커스 해제 (blur 이벤트 발생)
            await page.keyboard.press('Tab');

            // 내용 입력 성공
            contentInputSuccess = true;
        } else {
            console.log('contenteditable 영역을 찾을 수 없습니다. 대체 방법을 시도합니다...');

            // ProseMirror 에디터의 iframe 또는 특수 선택자 확인
            const possibleEditors = [
                page.locator('.ProseMirror'),
                page.locator('.ProseMirror-content'),
                page.locator('.editor-content'),
                page.locator('editor-other#editor')
            ];

            let editorFound = false;

            for (const editor of possibleEditors) {
                if (await editor.count() > 0) {
                    console.log('에디터 요소를 찾았습니다. 클릭합니다...');
                    await editor.click();

                    // 필요한 경우 Shadow DOM 내부 요소 처리
                    if ((await editor.evaluate(el => el.tagName)) === 'EDITOR-OTHER') {
                        console.log('Shadow DOM 에디터 요소입니다.');

                        // Shadow DOM 내부에 접근하여 클릭
                        await page.evaluate(() => {
                            const editorOther = document.querySelector('editor-other#editor');
                            if (editorOther && editorOther.shadowRoot) {
                                const editorDom = editorOther.shadowRoot.querySelector('#editor-shadow-dom');
                                if (editorDom) {
                                    editorDom.focus();
                                    editorDom.click();
                                }
                            }
                        });
                    }

                    // 잠시 대기
                    await page.waitForTimeout(50);

                    // 자연스러운 타이핑으로 단락별 입력
                    for (let i = 0; i < paragraphs.length; i++) {
                        await page.waitForTimeout(180);
                        await page.keyboard.type(paragraphs[i], { delay: 15 });

                        if (i < paragraphs.length - 1) {
                            await page.keyboard.press('Enter');
                            await page.keyboard.press('Enter');
                            await page.waitForTimeout(100);
                        }
                    }

                    console.log('대체 에디터에 내용을 입력했습니다.');
                    editorFound = true;
                    contentInputSuccess = true;
                    break;
                }
            }

            // 에디터를 찾지 못한 경우 textarea 시도
            if (!editorFound) {
                const textarea = page.locator('textarea').first();
                if (await textarea.count() > 0) {
                    console.log('textarea를 찾았습니다. 내용을 입력합니다...');
                    await textarea.click();

                    // 전체 내용을 한 번에 입력
                    await textarea.fill(paragraphs.join('\n\n'));
                    console.log('textarea에 내용을 입력했습니다.');
                    contentInputSuccess = true;
                } else {
                    console.log('어떤 입력 필드도 찾지 못했습니다.');
                }
            }
        }

        // 내용 입력 후 스크린샷 (확인용)
        const contentScreenshotPath = path.join(logDir, 'content-input.png');
        await page.screenshot({ path: contentScreenshotPath });
        console.log('내용 입력 후 스크린샷을 저장했습니다:', contentScreenshotPath);

    } catch (error) {
        console.log('에디터 내용 입력 중 오류 발생:', error);
    }

    if (!contentInputSuccess) {
        console.log('다른 방법으로 내용 입력을 시도합니다...');

        try {
            // CKEditor 확인
            const ckEditorFrame = page.frameLocator('.ck-editor__editable').first();
            if (await ckEditorFrame.count() > 0) {
                console.log('CKEditor를 찾았습니다.');
                await ckEditorFrame.locator('body').fill(postContent);
                console.log('CKEditor에 내용을 입력했습니다.');
                contentInputSuccess = true;
            }
        } catch (error) {
            console.log('CKEditor 접근 중 오류 발생:', error);
        }

        if (!contentInputSuccess) {
            try {
                // content 필드 또는 textarea 찾기
                const contentField = page.locator('#content');
                if (await contentField.count() > 0) {
                    console.log('content 필드를 찾았습니다.');
                    await page.evaluate((content) => {
                        document.getElementById('content').value = content;
                    }, postContent);
                    console.log('content 필드에 내용을 설정했습니다.');
                    contentInputSuccess = true;
                } else {
                    // textarea 찾기
                    const textarea = page.locator('textarea').first();
                    if (await textarea.count() > 0) {
                        console.log('textarea를 찾았습니다.');
                        await textarea.fill(postContent);
                        console.log('textarea에 내용을 설정했습니다.');
                        contentInputSuccess = true;
                    }
                }
            } catch (error) {
                console.log('일반 필드 접근 중 오류 발생:', error);
            }
        }

        if (!contentInputSuccess) {
            console.log('모든 내용 입력 방법이 실패했습니다. 마지막 방법을 시도합니다...');

            try {
                // 마지막 방법: 제출 직전에 강제로 hidden input 설정
                await page.evaluate((content) => {
                    // 1. ProseMirror 에디터 -> hidden input 값 복사 시도
                    try {
                        const editorOther = document.querySelector('editor-other#editor');
                        if (editorOther && editorOther.shadowRoot) {
                            const editorDom = editorOther.shadowRoot.querySelector('#editor-shadow-dom');
                            if (editorDom) {
                                console.log('Shadow DOM 직접 접근 성공');

                                // 이 내용을 hidden input에 강제 설정
                                const hiddenContent = document.querySelector('#content');
                                if (hiddenContent) {
                                    hiddenContent.value = editorDom.innerHTML || content;
                                    console.log('에디터 내용을 hidden input에 강제 복사');
                                }
                            }
                        }
                    } catch (e) {
                        console.log('Shadow DOM -> hidden input 복사 실패:', e);
                    }

                    // 2. 직접 hidden input에 내용 설정 (가장 확실한 방법)
                    const hiddenContent = document.querySelector('#content');
                    if (hiddenContent) {
                        hiddenContent.value = content;
                        console.log('Hidden input에 최종 내용 강제 설정 완료');
                        return true;
                    }

                    // 3. form 요소 찾아서 hidden input 추가 시도
                    const form = document.querySelector('form');
                    if (form && !hiddenContent) {
                        const newHiddenInput = document.createElement('input');
                        newHiddenInput.type = 'hidden';
                        newHiddenInput.id = 'content';
                        newHiddenInput.name = 'content';
                        newHiddenInput.value = content;
                        form.appendChild(newHiddenInput);
                        console.log('Hidden input이 없어서 새로 생성하여 추가했습니다');
                        return true;
                    }

                    return false;
                }, postContent);

                console.log('가능한 모든 요소에 내용 입력을 시도했습니다.');
            } catch (e) {
                console.log('대체 방법도 실패:', e);
            }
        }
    }

    // 현재 페이지 스크린샷 저장 (내용 입력 확인용)
    const contentScreenshotPath = path.join(logDir, 'content-input.png');
    await page.screenshot({ path: contentScreenshotPath });
    console.log('내용 입력 후 스크린샷을 저장했습니다:', contentScreenshotPath);

    // 게시물 등록 버튼 클릭 직전에 content 필드 설정 확인
    console.log('게시물 등록 직전에 content 값을 확인합니다...');
    const contentFieldValue = await page.evaluate(() => {
        const contentField = document.querySelector('#content');
        return contentField ? contentField.value : '내용 필드가 없습니다';
    });

    console.log(`Content 필드 값 (처음 20자): ${contentFieldValue.substring(0, 20)}...`);

    // content 필드가 비어있거나 값이 없으면 마지막으로 설정 시도
    if (!contentFieldValue || contentFieldValue === '내용 필드가 없습니다') {
        console.log('등록 직전 Content 필드가 비어있습니다. 마지막으로 설정 시도합니다.');
        await page.evaluate((content) => {
            // 폼을 찾아서 hidden input 추가 또는 수정
            const form = document.querySelector('form');
            if (form) {
                let contentField = document.querySelector('#content');
                if (!contentField) {
                    contentField = document.createElement('input');
                    contentField.type = 'hidden';
                    contentField.id = 'content';
                    contentField.name = 'content';
                    form.appendChild(contentField);
                }
                contentField.value = content;
                console.log('폼 제출 직전에 content 필드 값 설정 완료');
            }
        }, postContent);
    }

    // 게시물 등록 버튼 클릭
    console.log('게시물 등록 버튼을 클릭합니다...');
    const postButton = page.locator('button[type="submit"]');

    if (await postButton.count() === 0) {
        console.log('게시물 등록 버튼을 찾을 수 없습니다. 다른 등록 버튼을 찾아봅니다...');

        // 등록 또는 저장 버튼 찾기 (다양한 선택자로 시도)
        const submitButtons = [
            page.locator('button:has-text("등록")'),
            page.locator('button:has-text("저장")'),
            page.locator('button:has-text("작성")'),
            page.locator('button:has-text("글쓰기")'),
            page.locator('button.primary'),
            page.locator('input[type="submit"]')
        ];

        let buttonFound = false;
        for (const button of submitButtons) {
            if (await button.count() > 0) {
                console.log('대체 등록 버튼을 찾았습니다.');
                await button.click();
                buttonFound = true;
                break;
            }
        }

        if (!buttonFound) {
            const postFormScreenshotPath = path.join(logDir, 'post-form.png');
            await page.screenshot({ path: postFormScreenshotPath, fullPage: true });
            console.log('어떤 등록 버튼도 찾을 수 없습니다. 스크린샷 저장:', postFormScreenshotPath);

            // 마지막 수단: 폼 직접 제출 시도
            await page.evaluate(() => {
                const form = document.querySelector('form');
                if (form) {
                    console.log('폼을 직접 제출합니다.');
                    form.submit();
                }
            });
        }
    } else {
        await postButton.click();
    }
    await page.waitForLoadState('networkidle');

    // 게시물 등록 확인 후 삭제 기능 추가
    try {
        const postTitle = await page.textContent('h1');
        console.log(`작성된 게시물 제목: ${postTitle}`);

        // 게시물 확인 스크린샷
        const postViewScreenshotPath = path.join(logDir, 'post-view.png');
        await page.screenshot({ path: postViewScreenshotPath });
        console.log('게시물 확인 스크린샷을 저장했습니다:', postViewScreenshotPath);

        console.log('첫 게시물이 성공적으로 작성되었습니다.');

        // 게시물 삭제 프로세스 시작
        console.log('작성된 게시물을 삭제합니다...');

        // 대화상자(alert/confirm) 처리를 위한 이벤트 핸들러 등록
        // 반드시 삭제 버튼 클릭 전에 등록해야 함
        page.once('dialog', async dialog => {
            console.log(`대화상자 감지: ${dialog.type()}, 메시지: "${dialog.message()}"`);
            // 대화상자 스크린샷은 불가능함 (브라우저 네이티브 UI임)

            // 확인 버튼 클릭 (dialog.dismiss()는 취소 버튼)
            await dialog.accept();
            console.log('대화상자 확인 버튼을 클릭했습니다.');
        });

        // 삭제 버튼 찾기 (더 구체적인 선택자 사용)
        const deleteButtons = [
            page.locator('button:has-text("삭제")'),
            page.locator('a:has-text("삭제")'),
            page.locator('text=삭제').first()
        ];

        let deleteButtonFound = false;
        for (const button of deleteButtons) {
            if (await button.count() > 0) {
                console.log('삭제 버튼을 찾았습니다. 스크린샷을 저장합니다...');
                const beforeDeleteScreenshotPath = path.join(logDir, 'before-delete.png');
                await page.screenshot({ path: beforeDeleteScreenshotPath, fullPage: true });

                // 버튼의 스크린샷도 저장
                try {
                    const buttonScreenshotPath = path.join(logDir, 'delete-button.png');
                    await button.screenshot({ path: buttonScreenshotPath });
                    console.log('삭제 버튼 스크린샷 저장:', buttonScreenshotPath);
                } catch (e) {
                    console.log('삭제 버튼 스크린샷 저장 실패:', e.message);
                }

                console.log('삭제 버튼을 클릭합니다...');
                await button.click();
                deleteButtonFound = true;

                // 충분한 대기 시간 추가 (대화상자 처리 및 삭제 작업 완료 대기)
                console.log('삭제 처리를 위해 대기 중...');
                await page.waitForTimeout(1000);
                break;
            }
        }

        if (!deleteButtonFound) {
            console.log('삭제 버튼을 찾을 수 없습니다. 페이지 스크린샷을 저장합니다...');
            const noDeleteButtonScreenshotPath = path.join(logDir, 'no-delete-button.png');
            await page.screenshot({ path: noDeleteButtonScreenshotPath, fullPage: true });

            // 페이지 소스 저장
            const pageSource = await page.content();
            fs.writeFileSync(path.join(logDir, 'post-page-source.html'), pageSource);
            console.log('페이지 소스를 저장했습니다:', path.join(logDir, 'post-page-source.html'));
        } else {
            // 현재 URL 확인
            const currentUrl = page.url();
            console.log('삭제 후 현재 URL:', currentUrl);

            // 삭제 후 페이지 스크린샷
            const afterDeleteScreenshotPath = path.join(logDir, 'after-delete.png');
            await page.screenshot({ path: afterDeleteScreenshotPath });
            console.log('삭제 후 페이지 스크린샷을 저장했습니다:', afterDeleteScreenshotPath);

            // 삭제 확인을 위해 게시판 목록으로 이동
            console.log('게시판 목록으로 이동하여 삭제 여부를 확인합니다...');
            await page.goto('http://localhost:3000/boards/1/posts');
            await page.waitForLoadState('networkidle');

            // 게시판 목록 스크린샷
            const boardListScreenshotPath = path.join(logDir, 'board-list.png');
            await page.screenshot({ path: boardListScreenshotPath });
            console.log('게시판 목록 스크린샷을 저장했습니다:', boardListScreenshotPath);

            // "첫 번째 게시물" 텍스트가 있는지 확인
            const hasPost = await page.locator('text="첫 번째 게시물"').count() > 0;
            if (hasPost) {
                console.log('경고: 게시물이 여전히 목록에 보입니다. 삭제가 실패했을 수 있습니다.');
            } else {
                console.log('게시물이 목록에서 보이지 않습니다. 삭제가 성공했을 가능성이 높습니다.');
            }
        }
    } catch (e) {
        console.log('게시물 제목을 찾을 수 없거나 삭제 과정 중 오류 발생:', e);
        console.log('현재 URL:', page.url());
    }
}

/**
 * 로그아웃
 */
async function logout(page) {
    const logDir = path.join(__dirname, 'log');
    console.log('로그아웃합니다...');

    try {
        // 헤더의 사용자 메뉴 클릭
        const userMenu = page.locator('text=admin').first();

        if (await userMenu.count() === 0) {
            console.log('사용자 메뉴를 찾을 수 없습니다. 현재 페이지 스크린샷을 저장합니다.');
            const logoutScreenshotPath = path.join(logDir, 'before-logout.png');
            await page.screenshot({ path: logoutScreenshotPath, fullPage: true });
            throw new Error('사용자 메뉴를 찾을 수 없습니다.');
        }

        await userMenu.click();

        // 로그아웃 링크 클릭
        const logoutLink = page.locator('text=로그아웃');

        if (await logoutLink.count() === 0) {
            console.log('로그아웃 링크를 찾을 수 없습니다.');
            throw new Error('로그아웃 링크를 찾을 수 없습니다.');
        }

        await logoutLink.click();
        await page.waitForLoadState('networkidle');

        // 로그아웃 확인 스크린샷
        const afterLogoutScreenshotPath = path.join(logDir, 'after-logout.png');
        await page.screenshot({ path: afterLogoutScreenshotPath });
        console.log('로그아웃 후 스크린샷을 저장했습니다:', afterLogoutScreenshotPath);

        console.log('로그아웃 완료. 현재 URL:', page.url());
    } catch (e) {
        console.log('로그아웃 중 오류 발생:', e);
    }
}

// 스크립트 실행
setupToyBoardApplication().catch(console.error);