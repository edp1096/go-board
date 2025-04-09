document.addEventListener('DOMContentLoaded', function () {
    // 에디터 초기화
    const editorContainer = document.querySelector(`editor-other#editor`).shadowRoot
    const editorEL = editorContainer.querySelector("#editor-shadow-dom")
    if (editorContainer) {
        const boardId = document.getElementById('board-id').value;
        const contentField = document.getElementById('content');
        const editorOptions = {
            uploadInputName: "upload-files[]",
            uploadActionURI: `/api/boards/${boardId}/upload`,
            uploadAccessURI: `/uploads/boards/${boardId}/images`,
            uploadCallback: function (response) {
                console.log("업로드 완료:", response);

                // 애니메이션 이미지 처리
                if (response && response.files && response.files.length > 0) {
                    response.files.forEach(file => {
                        if (file.animation) {
                            // 에디터에 이미지가 삽입된 후 처리
                            setTimeout(() => {
                                // 최근 삽입된 이미지를 찾기 (URL로 식별)
                                const images = editorEL.querySelectorAll('img');
                                images.forEach(img => {
                                    if (img.src.includes(file.storagename)) {
                                        // 애니메이션 이미지로 표시
                                        img.setAttribute('data-animate', 'true');
                                        console.log(`애니메이션 이미지 태그 처리 완료: ${file.storagename}`);
                                    }
                                });
                            }, 100); // 이미지가 DOM에 삽입될 시간을 조금 주기 위한 지연
                        }
                    });
                }

                setTimeout(() => {
                    console.log(editor.getHTML());
                }, 1000);
            }
        };
        // const editor = new MyEditor(contentField.value, editorContainer, editorOptions);
        const editor = new MyEditor("", editorEL, editorOptions);

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