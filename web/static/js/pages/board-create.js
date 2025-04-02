document.addEventListener('DOMContentLoaded', function () {
    // 에디터 초기화
    const editorContainer = document.querySelector(`editor-other#editor`).shadowRoot
    const editorEL = editorContainer.querySelector("#editor-shadow-dom")
    if (editorContainer) {
        const boardId = document.getElementById('boardId').value;
        const contentField = document.getElementById('content');
        const editorOptions = {
            uploadInputName: "upload-files[]",
            uploadActionURI: `/api/boards/${boardId}/images`,
            uploadAccessURI: "/uploads",
            uploadCallback: function (response) {
                console.log("업로드 완료:", response);
            }
        };
        // const editor = new MyEditor(contentField.value, editorContainer, editorOptions);
        const editor = new MyEditor("Hello world!!", editorEL, editorOptions);

        // 폼 제출 이벤트 핸들러
        const form = document.querySelector('form');
        form.addEventListener('submit', function () {
            // 에디터 내용을 hidden input에 설정
            contentField.value = editor.getHTML();
        });
    }

    // 파일 업로드 프리뷰 기능
    const fileInput = document.getElementById('files');
    if (fileInput) {
        fileInput.addEventListener('change', function () {
            const filesList = document.getElementById('files-list');
            if (filesList) {
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
            }
        });
    }
});