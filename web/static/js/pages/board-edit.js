// board-edit.js
document.addEventListener('alpine:init', () => {
    Alpine.data('postEditor', function () {
        return {
            submitting: false,

            submitForm(form) {
                const boardId = document.getElementById('boardId').value;
                const postId = document.getElementById('postId').value;

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