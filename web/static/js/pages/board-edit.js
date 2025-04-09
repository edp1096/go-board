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
            },
            uploadErrorCallback: async function (response) {
                const message = response.message;
                console.error("업로드 오류:", response);
                alert(`업로드 오류: ${message}`);
            },
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


    // 업로드 허용치 초과 체크
    const fileInputs = document.querySelectorAll('input[type="file"]');

    // 각 파일 입력 요소에 이벤트 리스너 추가
    fileInputs.forEach(input => {
        // 파일 유형별 크기 제한 설정 (서버 설정과 일치시켜야 함)
        const maxFileSizeMB = input.dataset.fileType === 'image' ? 20 : 10; // 이미지는 20MB, 일반 파일은 10MB
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
                let message = `다음 파일이 최대 허용 크기(${maxFileSizeMB}MB)를 초과했습니다:\n\n`;

                oversizedFiles.forEach(file => {
                    message += `- ${file.name}: ${file.size}MB\n`;
                });

                alert(message);
                this.value = ''; // 파일 선택 초기화
            }
        });
    });
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