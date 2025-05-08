// static/js/pages/page-create.js

let editor;
let sessionId = ''; // 사용자 세션 ID 저장용

document.addEventListener('DOMContentLoaded', function () {
    // 세션 ID 생성 (16자리 랜덤 문자열)
    sessionId = generateSessionId();

    // 세션 ID를 hidden input에 저장 (서버로 전송하기 위함)
    const sessionIdInput = document.createElement('input');
    sessionIdInput.type = 'hidden';
    sessionIdInput.id = 'editorSessionId';
    sessionIdInput.name = 'editorSessionId';
    sessionIdInput.value = sessionId;
    document.querySelector('form').appendChild(sessionIdInput);

    // 에디터 초기화
    initEditor();
});

// 16자리 랜덤 문자열 생성 함수
function generateSessionId() {
    return Array.from(Array(16), () => Math.floor(Math.random() * 36).toString(36)).join('');
}

// 에디터 초기화 함수
function initEditor() {
    const editorContainer = document.querySelector(`editor-other#editor`)?.shadowRoot;
    if (!editorContainer) return;

    const editorEL = editorContainer.querySelector("#editor-shadow-dom");
    if (!editorEL) return;

    const contentField = document.getElementById('content');
    if (!contentField) return;

    // 에디터 옵션 설정 (세션 ID 사용)
    const editorOptions = {
        uploadInputName: "upload-files[]",
        uploadActionURI: `/api/pages/upload?sessionId=${sessionId}`, // 세션 ID 전달
        uploadAccessURI: `/uploads/pages/temp/${sessionId}/medias`, // 세션별 임시 경로
        uploadCallback: function (response) {
            // 업로드 파일 추적을 위한 처리는 서버에서 함
        },
        uploadErrorCallback: async function (response) {
            const message = response.message;
            console.error("업로드 오류:", response);
            alert(`업로드 오류: ${message}`);
        },
    };

    // 에디터 초기화 (빈 내용으로)
    editor = new MyEditor("", editorEL, editorOptions);
}

// 에디터 내용 저장 함수
function saveContent() {
    const contentField = document.getElementById('content');
    if (contentField && editor) {
        contentField.value = editor.getHTML();
    }
}