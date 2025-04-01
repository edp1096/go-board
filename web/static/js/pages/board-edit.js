// board-edit.js
document.addEventListener('alpine:init', () => {
    Alpine.data('postEditor', function () {
        return {
            submitting: false,

            submitForm(form) {
                const boardId = document.getElementById('boardId').value;
                const postId = document.getElementById('postId').value;

                this.submitting = true;

                fetch(`/boards/${boardId}/posts/${postId}`, {
                    method: 'PUT',
                    headers: {
                        'Content-Type': 'application/x-www-form-urlencoded',
                        'Accept': 'application/json'
                    },
                    body: new URLSearchParams(new FormData(form))
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