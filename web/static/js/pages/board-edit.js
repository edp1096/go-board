// board-edit.js

let editor;

// 파일 크기 포맷팅 함수
function formatFileSize(bytes) {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

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
            }
        };

        const content = document.querySelector("#content").value;
        editor = new MyEditor(content, editorEL, editorOptions);

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

document.addEventListener('alpine:init', () => {
    Alpine.data('postEditor', function () {
        return {
            submitting: false,
            fileNames: [],

            submitForm(form) {
                const boardId = document.getElementById('board-id').value;
                const postId = document.getElementById('post-id').value;

                // 에디터 내용을 hidden input에 설정
                // const editorContent = document.querySelector('editor-other').shadowRoot.querySelector('#editor-shadow-dom').innerHTML;
                const editorContent = editor.getHTML();
                document.getElementById('content').value = editorContent;

                this.submitting = true;

                // FormData 객체 생성 (파일 업로드를 위해)
                const formData = new FormData(form);

                fetch(`/boards/${boardId}/posts/${postId}`, {
                    method: 'PUT',
                    body: formData, // FormData 사용
                    headers: {
                        'Accept': 'application/json'
                        // Content-Type 헤더는 자동으로 설정됨
                    }
                })
                    .then(res => res.json())
                    .then(data => {
                        if (data.success) {
                            window.location.href = `/boards/${boardId}/posts/${postId}`;
                        } else {
                            alert(data.message);
                            this.submitting = false;
                        }
                    })
                    .catch(err => {
                        alert('오류가 발생했습니다: ' + err);
                        this.submitting = false;
                    });
            }
        };
    });
});