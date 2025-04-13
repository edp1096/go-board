// board-create.js

let editor;

document.addEventListener('DOMContentLoaded', function () {
    // 에디터 초기화
    initEditor();

    // 파일 업로드 프리뷰 기능 설정
    setupFileUploadPreview();

    // 파일 크기 제한 검사 설정
    setupFileSizeValidation();
});

// 에디터 초기화 함수
function initEditor() {
    const editorContainer = document.querySelector(`editor-other#editor`)?.shadowRoot;
    if (!editorContainer) return;

    const editorEL = editorContainer.querySelector("#editor-shadow-dom");
    if (!editorEL) return;

    const boardId = document.getElementById('board-id')?.value;
    const contentField = document.getElementById('content');

    if (!boardId || !contentField) return;

    const editorOptions = {
        uploadInputName: "upload-files[]",
        uploadActionURI: `/api/boards/${boardId}/upload`,
        uploadAccessURI: `/uploads/boards/${boardId}/images`,
        uploadCallback: function (response) {
            // console.log("업로드 완료:", response);
        },
        uploadErrorCallback: async function (response) {
            const message = response.message;
            console.error("업로드 오류:", response);
            alert(`업로드 오류: ${message}`);
        },
    };

    // 에디터 초기화 (빈 내용으로)
    editor = new MyEditor("", editorEL, editorOptions);

    // 폼 제출 이벤트 핸들러
    const form = document.querySelector('form');
    if (form) {
        form.addEventListener('submit', function () {
            // 에디터 내용을 hidden input에 설정
            contentField.value = editor.getHTML();
        });
    }
}

// 파일 업로드 프리뷰 기능
function setupFileUploadPreview() {
    const fileInput = document.getElementById('files');
    if (!fileInput) return;

    fileInput.addEventListener('change', function () {
        const filesList = document.getElementById('files-list');
        if (!filesList) return;

        filesList.innerHTML = '';
        if (this.files.length > 0) {
            for (let i = 0; i < this.files.length; i++) {
                const li = document.createElement('li');
                li.textContent = this.files[i].name;
                filesList.appendChild(li);
            }
        } else {
            const li = document.createElement('li');
            li.textContent = '선택된 파일 없음';
            filesList.appendChild(li);
        }
    });
}

// 파일 크기 제한 검사
function setupFileSizeValidation() {
    const fileInputs = document.querySelectorAll('input[type="file"]');

    // 각 파일 입력 요소에 이벤트 리스너 추가
    fileInputs.forEach(input => {
        // 파일 유형별 크기 제한 설정 (서버 설정과 일치시켜야 함)
        const maxFileSizeMB = input.dataset.fileType === 'image' ? 20 : 20; // 이미지는 20MB, 일반 파일은 20MB
        const maxFileSizeBytes = maxFileSizeMB * 1024 * 1024;

        // 입력값 변경 시 이벤트
        input.addEventListener('change', function () {
            const files = this.files;
            let oversizedFiles = [];

            // 모든 선택된 파일 검사
            for (let i = 0; i < files.length; i++) {
                if (files[i].size > maxFileSizeBytes) {
                    oversizedFiles.push({
                        name: files[i].name,
                        size: (files[i].size / (1024 * 1024)).toFixed(2)
                    });
                }
            }

            // 용량 초과 파일이 있으면 알림
            if (oversizedFiles.length > 0) {
                let message = `다음 파일이 최대 허용 크기를 초과했습니다:\n\n`;

                oversizedFiles.forEach(file => {
                    message += `- ${file.name}: ${file.size}MB\n`;
                });

                alert(message);
                this.value = ''; // 파일 선택 초기화
            }
        });
    });
}