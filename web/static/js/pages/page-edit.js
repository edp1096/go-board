// static/js/pages/page-edit.js
let editor;

document.addEventListener('DOMContentLoaded', function () {
    // 에디터 초기화
    initEditor();
});

// 에디터 초기화 함수
function initEditor() {
    const editorContainer = document.querySelector(`editor-other#editor`)?.shadowRoot;
    if (!editorContainer) return;

    const editorEL = editorContainer.querySelector("#editor-shadow-dom");
    if (!editorEL) return;

    const contentField = document.getElementById('content');
    if (!contentField) return;

    // URL에서 페이지 ID 추출
    const path = window.location.pathname;
    const matches = path.match(/\/admin\/pages\/(\d+)\/edit/);
    let pageId = 0;

    if (matches && matches.length > 1) {
        pageId = matches[1];
    }

    // 에디터 옵션 설정
    const editorOptions = {
        uploadInputName: "upload-files[]",
        uploadActionURI: `/api/pages/${pageId}/upload`, // 페이지 이미지 업로드용 API 경로
        uploadAccessURI: `/uploads/pages/${pageId}/medias`, // 페이지 이미지 접근 경로
        uploadCallback: function (response) {
            // console.log("업로드 완료:", response);
        },
        uploadErrorCallback: async function (response) {
            const message = response.message;
            console.error("업로드 오류:", response);
            alert(`업로드 오류: ${message}`);
        },
    };

    // 에디터 초기화 (기존 내용으로)
    const content = contentField.value || "";
    editor = new MyEditor(content, editorEL, editorOptions);
}

// 에디터 내용 저장 함수
function saveContent() {
    const contentField = document.getElementById('content');
    if (contentField && editor) {
        contentField.value = editor.getHTML();
    }
}