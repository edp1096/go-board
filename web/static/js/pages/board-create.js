document.addEventListener('DOMContentLoaded', function() {
    const editorContainer = document.getElementById('content-editor');
    if (!editorContainer) return;
    const boardId = document.getElementById('boardId').value;
    const contentField = document.getElementById('content');
    const editorOptions = {
        uploadInputName: "upload-files[]",
        uploadActionURI: `/api/boards/${boardId}/upload`,
        uploadAccessURI: "/uploads",
        uploadCallback: function(response) {
            console.log("업로드 완료:", response);
        }
    };
    const editor = new MyEditor(contentField.value, editorContainer, editorOptions);
    const form = document.querySelector('form');
    form.addEventListener('submit', function() {
        contentField.value = editor.getHTML();
    });
});
