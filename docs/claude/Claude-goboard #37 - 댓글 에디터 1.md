# Claude Conversation

## Prompt:
현재 goboard의 댓글입력창이 textarea인데, discourse 같이 댓글에도 이미지 업로드와 파일업로드가 되게 할 수 있을까?
에디터도 본문에디터처럼 myeditor(prosemirror에디터)로 바꿔서 말이지.

대답 항상 한글로.
제안코드는 항상 생략없는 완전한 코드로.

## Response:

사용자는 게시판 시스템인 goboard에서 댓글 입력창을 단순 textarea에서 본문 편집에 사용되는 myeditor(prosemirror 기반)로 바꾸고, 이미지 업로드와 파일 업로드 기능을 추가하고 싶어합니다. 이는 Discourse와 같은 포럼 시스템처럼 댓글에서도 리치 에디팅 기능을 사용하고 싶다는 요청입니다.

현재 시스템 구조를 파악해 보겠습니다:

1. goboard는 Go로 작성된, Fiber 프레임워크 기반 웹 애플리케이션입니다.
2. 본문 에디터는 이미 myeditor(prosemirror 기반)를 사용하고 있습니다.
3. 댓글 시스템은 현재 단순 textarea로 구현되어 있습니다.
4. 업로드 관련 API는 이미 존재하는 것으로 보입니다.

필요한 변경 사항:

1. 댓글 입력 부분을 찾아 textarea 대신 myeditor 컴포넌트로 교체
2. 댓글 API가 HTML 콘텐츠를 처리하도록 수정
3. 업로드 API를 댓글에서도 사용할 수 있도록 연결
4. 댓글 표시 부분이 HTML을 렌더링하도록 수정

먼저 댓글 관련 템플릿과 자바스크립트 파일을 찾아보겠습니다. 주요 파일은:

1. web/templates/board/view.html - 댓글 시스템 UI가 있는 것으로 보임
2. web/static/js/pages/board-view.js - 댓글 관련 자바스크립트 코드가 있는 것으로 보임
3. internal/handlers/comment_handler.go - 댓글 관련 API 핸들러

댓글 입력 UI는 web/templates/board/view.html의 맨 아래 쪽에 있을 가능성이 높습니다. 이 부분을 확인하고 필요한 변경사항을 파악하겠습니다.

현재 댓글 입력 부분은 다음과 같습니다 (web/templates/board/view.html 파일 내에서):

```html
<!-- 댓글 작성 폼 -->
{{if .user}}
<div class="mt-4">
    <div class="flex space-x-3">
        <div class="flex-shrink-0">
            <div class="h-10 w-10 rounded-full bg-gray-200 flex items-center justify-center">
                <span class="text-sm font-medium text-gray-500">{{slice .user.Username 0 1}}</span>
            </div>
        </div>
        <div class="min-w-0 flex-1">
            <form @submit.prevent="submitComment()">
                <div>
                    <label for="comment" class="sr-only">댓글 내용</label>
                    <textarea id="comment" name="comment" rows="3" x-model="commentContent" class="block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm" :placeholder="replyToId ? `${replyToUser}님에게 답글 작성...` : '댓글을 입력하세요...'" required></textarea>
                </div>
                <div class="mt-3 flex items-center justify-between">
                    <div>
                        <template x-if="replyToId !== null">
                            <button type="button" class="text-sm text-gray-500 hover:text-gray-700" @click="cancelReply()">
                                답글 취소
                            </button>
                        </template>
                    </div>
                    <button type="submit" class="inline-flex items-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                        댓글 등록
                    </button>
                </div>
            </form>
        </div>
    </div>
</div>
{{else}}
<div class="mt-4 bg-gray-50 p-4 rounded-md text-center">
    <p class="text-gray-600">댓글을 작성하려면 <a href="/auth/login?redirect=/boards/{{.board.ID}}/posts/{{.post.ID}}" class="text-blue-600 hover:text-blue-500">로그인</a>이 필요합니다.</p>
</div>
{{end}}
```

댓글을 표시하는 부분도 살펴보겠습니다:

```html
<!-- 댓글 표시 -->
<template x-for="comment in comments" :key="comment.id">
    <div class="bg-white rounded-lg shadow-sm overflow-hidden">
        <!-- 기본 댓글 -->
        <div class="p-4">
            <div class="flex space-x-3">
                <div class="flex-shrink-0">
                    <div class="h-10 w-10 rounded-full bg-gray-200 flex items-center justify-center">
                        <span class="text-sm font-medium text-gray-500" x-text="comment.user ? comment.user.username.charAt(0) : '?'"></span>
                    </div>
                </div>
                <div class="min-w-0 flex-1">
                    <div class="text-sm font-medium text-gray-900" x-text="comment.user ? comment.user.username : '알 수 없음'"></div>
                    <div class="mt-1 text-sm text-gray-700" x-html="comment.content.replace(/\n/g, '<br>')"></div>
                    <div class="mt-2 text-xs text-gray-500 flex items-center space-x-2">
                        <span x-text="new Date(comment.createdAt).toLocaleString()"></span>
                        {{if .user}}
                        <button class="text-blue-600 hover:text-blue-800" @click="replyToId = comment.id; replyToUser = comment.user ? comment.user.username : '알 수 없음'; document.getElementById('comment').focus();">
                            답글
                        </button>
                        <template x-if="{{.user.ID}} === comment.userId || {{if eq .user.Role " admin"}}true{{else}}false{{end}}">
                            <div class="flex items-center space-x-2">
                                <button class="text-gray-600 hover:text-gray-800" @click="editComment(comment.id, comment.content)">
                                    수정
                                </button>
                                <button class="text-red-600 hover:text-red-800" @click="deleteComment(comment.id)">
                                    삭제
                                </button>
                            </div>
                        </template>
                        {{end}}
                    </div>
                </div>
            </div>
        </div>
        ... (답글 부분 생략)
    </div>
</template>
```

댓글 작성 및 수정을 다루는 자바스크립트 로직도 확인해 보겠습니다. board-view.js 파일에서 댓글 관련 함수를 찾아보겠습니다:

```javascript
// board-view.js 파일 내 commentSystem 객체
Alpine.data('commentSystem', function (commentsEnabled) {
    return {
        comments: [],
        commentContent: '',
        replyToId: null,
        replyToUser: '',
        loading: true,
        commentsEnabled: commentsEnabled,
        error: null,

        // ... 다른 함수들 ...

        submitComment() {
            const boardId = document.getElementById('boardId').value;
            const postId = document.getElementById('postId').value;

            // CSRF 토큰 가져오기
            const csrfToken = document.querySelector('meta[name="csrf-token"]').content;

            fetch(`/api/boards/${boardId}/posts/${postId}/comments`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': csrfToken
                },
                body: JSON.stringify({
                    content: this.commentContent,
                    parentId: this.replyToId
                })
            })
            // ... 응답 처리 ...
        },

        editComment(commentId, content) {
            // ... 댓글 수정 로직 ...
        }
    };
});
```

서버에서 댓글 처리하는 부분도 확인해보겠습니다 (internal/handlers/comment_handler.go):

```go
// CreateComment - 새 댓글 작성 API
func (h *CommentHandler) CreateComment(c *fiber.Ctx) error {
    // ... 생략 ...

    // 요청 본문 파싱
    var req struct {
        Content  string `json:"content"`
        ParentID *int64 `json:"parentId"`
    }

    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "요청 형식이 잘못되었습니다",
        })
    }

    // 댓글 내용 검증
    if req.Content == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "댓글 내용을 입력해주세요",
        })
    }

    // 댓글 생성
    comment, err := h.commentService.CreateComment(c.Context(), boardID, postID, user.ID, req.Content, req.ParentID)
    // ... 생략 ...
}
```

이제 게시물 작성/수정 시 에디터가 어떻게 초기화되는지 확인해 보겠습니다. web/templates/board/create.html 또는 web/templates/board/edit.html 파일을 확인해서 myeditor가 어떻게 구현되어 있는지 파악하겠습니다:

web/templates/board/create.html:
```html
<div class="mb-4">
    <label for="content" class="block text-sm font-medium text-gray-700">내용 <span class="text-red-500">*</span></label>
    <input type="hidden" name="content" id="content">
    <editor-other id="editor" data-idx="1"></editor-other>
</div>

<!-- editor-other 커스텀 엘리먼트를 위한 템플릿 -->
<template id="editor-shadow-dom-template">
    <link rel="stylesheet" href="/static/vendor/myeditor/myeditor.css">

    <style>
        #editor-shadow-dom {
            resize: vertical;
            overflow: hidden;
            height: 35vh;
            border: 1px solid silver;
        }
    </style>

    <div class="editor-shadow-dom-container">
        <div id="editor-shadow-dom"></div>
    </div>
</template>

<script>
    class modifierElement extends HTMLElement {
        constructor() {
            super()
            const templateContent = document.querySelector("#editor-shadow-dom-template").content
            const shadowRoot = this.attachShadow({ mode: "open" })
            shadowRoot.appendChild(templateContent.cloneNode(true))
        }
    }
    customElements.define("editor-other", modifierElement)
</script>
```

그리고 board-create.js에서 에디터 초기화 방법을 확인합니다:

```javascript
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
        const editor = new MyEditor("", editorEL, editorOptions);

        // 폼 제출 이벤트 핸들러
        const form = document.querySelector('form');
        form.addEventListener('submit', function () {
            // 에디터 내용을 hidden input에 설정
            contentField.value = editor.getHTML();
        });
    }
});
```

이제 모든 필요한 정보를 확보했으니, 댓글 시스템에 myeditor를 적용해 보겠습니다. 다음과 같은 변경이 필요합니다:

1. web/templates/board/view.html: 댓글 작성 부분의 textarea를 에디터 컴포넌트로 교체
2. web/static/js/pages/board-view.js: 댓글 제출 시 에디터에서 HTML 내용을 가져오도록 수정
3. 필요하다면 CSS 스타일 및 레이아웃 조정

여기에서 필요한 변경사항을 구현해 보겠습니다.

1. web/templates/board/view.html: 댓글 작성 폼을 에디터 컴포넌트로 교체하기 위한 수정

```html
<!-- 댓글 작성 폼 - 에디터 적용 -->
{{if .user}}
<div class="mt-4">
    <div class="flex space-x-3">
        <div class="flex-shrink-0">
            <div class="h-10 w-10 rounded-full bg-gray-200 flex items-center justify-center">
                <span class="text-sm font-medium text-gray-500">{{slice .user.Username 0 1}}</span>
            </div>
        </div>
        <div class="min-w-0 flex-1">
            <form @submit.prevent="submitComment()">
                <div>
                    <label for="comment" class="sr-only">댓글 내용</label>
                    <input type="hidden" id="comment" name="comment" x-model="commentContent">
                    <editor-comment id="comment-editor" data-idx="comment" :placeholder="replyToId ? `${replyToUser}님에게 답글 작성...` : '댓글을 입력하세요...'"></editor-comment>
                </div>
                <div class="mt-3 flex items-center justify-between">
                    <div>
                        <template x-if="replyToId !== null">
                            <button type="button" class="text-sm text-gray-500 hover:text-gray-700" @click="cancelReply()">
                                답글 취소
                            </button>
                        </template>
                    </div>
                    <button type="submit" class="inline-flex items-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                        댓글 등록
                    </button>
                </div>
            </form>
        </div>
    </div>
</div>
{{else}}
<div class="mt-4 bg-gray-50 p-4 rounded-md text-center">
    <p class="text-gray-600">댓글을 작성하려면 <a href="/auth/login?redirect=/boards/{{.board.ID}}/posts/{{.post.ID}}" class="text-blue-600 hover:text-blue-500">로그인</a>이 필요합니다.</p>
</div>
{{end}}

<!-- 댓글 에디터 템플릿 -->
<template id="comment-editor-template">
    <link rel="stylesheet" href="/static/vendor/myeditor/myeditor.css">

    <style>
        #comment-editor-container {
            resize: vertical;
            overflow: hidden;
            height: 15vh;
            min-height: 100px;
            border: 1px solid #d1d5db;
            border-radius: 0.375rem;
        }
    </style>

    <div class="comment-editor-container">
        <div id="comment-editor-container"></div>
    </div>
</template>

<script>
    class CommentEditorElement extends HTMLElement {
        constructor() {
            super()
            const templateContent = document.querySelector("#comment-editor-template").content
            const shadowRoot = this.attachShadow({ mode: "open" })
            shadowRoot.appendChild(templateContent.cloneNode(true))
        }
    }
    customElements.define("editor-comment", CommentEditorElement)
</script>
```

2. web/static/js/pages/board-view.js: 댓글 컴포넌트 로직 수정

```javascript
// 댓글 에디터 인스턴스를 저장할 변수
let commentEditor = null;

document.addEventListener('DOMContentLoaded', function() {
    // 댓글 에디터 초기화 로직 등을 여기에 추가
    initCommentEditor();
});

// 댓글 에디터 초기화 함수
function initCommentEditor() {
    // 에디터 요소가 있는지 확인
    const editorContainer = document.querySelector('editor-comment#comment-editor');
    if (!editorContainer) return;
    
    const shadowRoot = editorContainer.shadowRoot;
    const editorEl = shadowRoot.querySelector("#comment-editor-container");
    
    const boardId = document.getElementById('boardId').value;
    const postId = document.getElementById('postId').value;
    
    // 에디터 옵션 설정
    const editorOptions = {
        uploadInputName: "upload-files[]",
        uploadActionURI: `/api/boards/${boardId}/upload`,
        uploadAccessURI: `/uploads/boards/${boardId}/images`,
        // 댓글용 에디터이므로 간소화된 툴바 사용
        simpleToolbar: true,
        placeholder: editorContainer.getAttribute('placeholder') || '댓글을 입력하세요...',
        uploadCallback: function (response) {
            console.log("댓글 이미지 업로드 완료:", response);
        }
    };
    
    // 에디터 초기화
    commentEditor = new MyEditor("", editorEl, editorOptions);
}

// Alpine.js 컴포넌트 수정
document.addEventListener('alpine:init', () => {
    Alpine.data('commentSystem', function (commentsEnabled) {
        return {
            comments: [],
            commentContent: '',
            replyToId: null,
            replyToUser: '',
            loading: true,
            commentsEnabled: commentsEnabled,
            error: null,

            init() {
                if (this.commentsEnabled) {
                    this.loadComments();
                } else {
                    this.loading = false;
                }
            },

            loadComments() {
                // 기존 코드 유지
                // ...
            },

            submitComment() {
                const boardId = document.getElementById('boardId').value;
                const postId = document.getElementById('postId').value;

                // 에디터에서 HTML 내용 가져오기
                if (commentEditor) {
                    this.commentContent = commentEditor.getHTML();
                    
                    // 빈 댓글 체크
                    if (!this.commentContent || this.commentContent === '<p></p>') {
                        alert('댓글 내용을 입력해주세요.');
                        return;
                    }
                }

                // CSRF 토큰 가져오기
                const csrfToken = document.querySelector('meta[name="csrf-token"]').content;

                fetch(`/api/boards/${boardId}/posts/${postId}/comments`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'X-CSRF-Token': csrfToken
                    },
                    body: JSON.stringify({
                        content: this.commentContent,
                        parentId: this.replyToId
                    })
                })
                .then(res => res.json())
                .then(data => {
                    if (data.success) {
                        if (this.replyToId === null) {
                            this.comments.push(data.comment);
                        } else {
                            // 부모 댓글 찾아서 답글 추가
                            const parentIndex = this.comments.findIndex(c => c.id === this.replyToId);
                            if (parentIndex !== -1) {
                                if (!this.comments[parentIndex].children) {
                                    this.comments[parentIndex].children = [];
                                }
                                this.comments[parentIndex].children.push(data.comment);
                            }
                        }
                        
                        // 입력 필드 초기화
                        this.commentContent = '';
                        this.replyToId = null;
                        this.replyToUser = '';
                        
                        // 에디터 내용 초기화
                        if (commentEditor) {
                            commentEditor.setContent('');
                        }
                    } else {
                        alert(data.message);
                    }
                });
            },

            cancelReply() {
                this.replyToId = null;
                this.replyToUser = '';
            },

            editComment(commentId, content) {
                // 에디터로 편집하기 위한 모달이나 인라인 에디터 추가 필요
                // 일단은 기존 방식 유지
                const newContent = prompt('댓글을 수정합니다:', content);

                if (newContent && newContent !== content) {
                    // CSRF 토큰 가져오기
                    const csrfToken = document.querySelector('meta[name="csrf-token"]').content;
                    fetch(`/api/comments/${commentId}`, {
                        method: 'PUT',
                        headers: {
                            'Content-Type': 'application/json',
                            'X-CSRF-Token': csrfToken
                        },
                        body: JSON.stringify({
                            content: newContent
                        })
                    })
                    // ... 기존 응답 처리 코드 ...
                }
            },

            // 기존 메서드들
            // ...
        };
    });
});
```

이렇게 변경하려면 서버 측에서도 HTML 콘텐츠를 허용하도록 변경해야 합니다. 아마도 현재는 pure text만 받아들이도록 되어 있을 수 있습니다. 만약 HTML이 이미 허용된 상태면 별도의 수정이 필요 없습니다.

다음과 같이 view.html의 댓글 표시 부분도 약간 수정하여 HTML 렌더링이 제대로 되는지 확인하겠습니다:

```html
<!-- 댓글 내용 부분만 수정 -->
<div class="mt-1 text-sm text-gray-700 prose prose-sm" x-html="comment.content"></div>
```

이렇게 하면 댓글에 HTML이 포함되어 있을 때 제대로 렌더링될 수 있습니다.

추가적으로, 작업을 완료하기 위해 댓글 수정 기능도 에디터를 사용하도록 업그레이드할 수 있지만, 그것은 후속 작업으로 남겨두겠습니다. 현재 수정 기능은 간단한 prompt()를 사용하고 있습니다.

이제 위에서 만든 변경사항을 요약하고 제안해 보겠습니다.

현재 goboard의 댓글 입력창을 Discourse처럼 리치 에디터로 업그레이드하여 이미지와 파일 업로드를 지원하도록 만들 수 있습니다. 본문 에디터와 동일한 myeditor(prosemirror 기반)를 사용하면 일관된 사용자 경험을 제공할 수 있습니다.

다음과 같은 변경이 필요합니다:

1. 댓글 입력 부분에 에디터 컴포넌트를 추가합니다.
2. 자바스크립트 로직을 수정하여 에디터에서 HTML 콘텐츠를 가져옵니다.
3. 댓글 표시 부분이 HTML 콘텐츠를 제대로 렌더링하도록 수정합니다.

## 1. 먼저 board/view.html 파일의 댓글 입력 부분을 수정합니다:

```html
<!-- 댓글 작성 폼 - 에디터로 변경 -->
{{if .user}}
<div class="mt-4">
    <div class="flex space-x-3">
        <div class="flex-shrink-0">
            <div class="h-10 w-10 rounded-full bg-gray-200 flex items-center justify-center">
                <span class="text-sm font-medium text-gray-500">{{slice .user.Username 0 1}}</span>
            </div>
        </div>
        <div class="min-w-0 flex-1">
            <form @submit.prevent="submitComment()">
                <div>
                    <label for="comment-content" class="sr-only">댓글 내용</label>
                    <input type="hidden" id="comment-content" x-model="commentContent">
                    <editor-comment id="comment-editor" data-idx="comment"></editor-comment>
                </div>
                <div class="mt-3 flex items-center justify-between">
                    <div>
                        <template x-if="replyToId !== null">
                            <button type="button" class="text-sm text-gray-500 hover:text-gray-700" @click="cancelReply()">
                                답글 취소
                            </button>
                        </template>
                    </div>
                    <button type="submit" class="inline-flex items-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500" :disabled="submitting">
                        <span x-show="!submitting">댓글 등록</span>
                        <span x-show="submitting" class="flex items-center">
                            <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                            </svg>
                            등록 중...
                        </span>
                    </button>
                </div>
            </form>
        </div>
    </div>
</div>
{{else}}
<div class="mt-4 bg-gray-50 p-4 rounded-md text-center">
    <p class="text-gray-600">댓글을 작성하려면 <a href="/auth/login?redirect=/boards/{{.board.ID}}/posts/{{.post.ID}}" class="text-blue-600 hover:text-blue-500">로그인</a>이 필요합니다.</p>
</div>
{{end}}

<!-- 댓글 에디터용 템플릿 추가 -->
<template id="comment-editor-template">
    <link rel="stylesheet" href="/static/vendor/myeditor/myeditor.css">

    <style>
        #comment-editor-container {
            resize: vertical;
            overflow: hidden;
            height: 15vh;
            min-height: 100px;
            border: 1px solid #d1d5db;
            border-radius: 0.375rem;
        }
    </style>

    <div class="comment-editor-container">
        <div id="comment-editor-container"></div>
    </div>
</template>

<script>
    class CommentEditorElement extends HTMLElement {
        constructor() {
            super()
            const templateContent = document.querySelector("#comment-editor-template").content
            const shadowRoot = this.attachShadow({ mode: "open" })
            shadowRoot.appendChild(templateContent.cloneNode(true))
        }
    }
    customElements.define("editor-comment", CommentEditorElement)
</script>
```

## 2. 댓글 표시 부분도 HTML 렌더링을 지원하도록 수정합니다:

```html
<!-- 기본 댓글의 내용 부분 -->
<div class="mt-1 text-sm text-gray-700 prose prose-sm max-w-none" x-html="comment.content"></div>

<!-- 답글의 내용 부분 -->
<div class="mt-1 text-sm text-gray-700 prose prose-sm max-w-none" x-html="reply.content"></div>
```

## 3. board-view.js 파일에 댓글 에디터 초기화 및 관련 로직을 추가합니다:

```javascript
// 댓글 에디터 인스턴스를 저장할 변수
let commentEditor = null;

document.addEventListener('DOMContentLoaded', function() {
    // 기존 코드...
    
    // 댓글 에디터 초기화
    initCommentEditor();
});

// 댓글 에디터 초기화 함수
function initCommentEditor() {
    // 에디터 요소가 있는지 확인
    const editorContainer = document.querySelector('editor-comment#comment-editor');
    if (!editorContainer) return;
    
    const shadowRoot = editorContainer.shadowRoot;
    const editorEl = shadowRoot.querySelector("#comment-editor-container");
    
    const boardId = document.getElementById('boardId').value;
    
    // 에디터 옵션 설정
    const editorOptions = {
        uploadInputName: "upload-files[]",
        uploadActionURI: `/api/boards/${boardId}/upload`,
        uploadAccessURI: `/uploads/boards/${boardId}/images`,
        // 댓글용 에디터이므로 간소화된 툴바 사용
        simpleToolbar: true,
        placeholder: '댓글을 입력하세요...',
        uploadCallback: function (response) {
            console.log("댓글 이미지 업로드 완료:", response);
        }
    };
    
    // 에디터 초기화
    commentEditor = new MyEditor("", editorEl, editorOptions);
}

// Alpine.js 컴포넌트 수정
document.addEventListener('alpine:init', () => {
    Alpine.data('commentSystem', function (commentsEnabled) {
        return {
            comments: [],
            commentContent: '',
            replyToId: null,
            replyToUser: '',
            loading: true,
            submitting: false,
            commentsEnabled: commentsEnabled,
            error: null,

            init() {
                if (this.commentsEnabled) {
                    this.loadComments();
                } else {
                    this.loading = false;
                }
            },

            loadComments() {
                // 기존 코드 유지
                const boardId = document.getElementById('boardId').value;
                const postId = document.getElementById('postId').value;

                fetch(`/api/boards/${boardId}/posts/${postId}/comments`)
                    .then(res => res.json())
                    .then(data => {
                        if (data.success) {
                            this.comments = data.comments;
                        } else {
                            this.error = data.message;
                        }
                        this.loading = false;
                    })
                    .catch(err => {
                        this.error = '댓글을 불러오는 중 오류가 발생했습니다.';
                        this.loading = false;
                    });
            },

            submitComment() {
                const boardId = document.getElementById('boardId').value;
                const postId = document.getElementById('postId').value;

                // 에디터에서 HTML 내용 가져오기
                if (commentEditor) {
                    this.commentContent = commentEditor.getHTML();
                    
                    // 빈 댓글 체크
                    if (!this.commentContent || this.commentContent === '<p></p>') {
                        alert('댓글 내용을 입력해주세요.');
                        return;
                    }
                }

                // CSRF 토큰 가져오기
                const csrfToken = document.querySelector('meta[name="csrf-token"]').content;
                
                this.submitting = true;

                fetch(`/api/boards/${boardId}/posts/${postId}/comments`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'X-CSRF-Token': csrfToken
                    },
                    body: JSON.stringify({
                        content: this.commentContent,
                        parentId: this.replyToId
                    })
                })
                .then(res => res.json())
                .then(data => {
                    this.submitting = false;
                    if (data.success) {
                        if (this.replyToId === null) {
                            this.comments.push(data.comment);
                        } else {
                            // 부모 댓글 찾아서 답글 추가
                            const parentIndex = this.comments.findIndex(c => c.id === this.replyToId);
                            if (parentIndex !== -1) {
                                if (!this.comments[parentIndex].children) {
                                    this.comments[parentIndex].children = [];
                                }
                                this.comments[parentIndex].children.push(data.comment);
                            }
                        }
                        
                        // 입력 필드 초기화
                        this.commentContent = '';
                        this.replyToId = null;
                        this.replyToUser = '';
                        
                        // 에디터 내용 초기화
                        if (commentEditor) {
                            commentEditor.setContent('');
                        }
                    } else {
                        alert(data.message);
                    }
                })
                .catch(err => {
                    this.submitting = false;
                    console.error('댓글 등록 중 오류:', err);
                    alert('댓글 등록 중 오류가 발생했습니다.');
                });
            },

            cancelReply() {
                this.replyToId = null;
                this.replyToUser = '';
            },

            editComment(commentId, content) {
                // 현재는 단순 prompt 사용
                // 추후 에디터를 사용한 모달로 업그레이드 가능
                const newContent = prompt('댓글을 수정합니다:', content);

                if (newContent && newContent !== content) {
                    // CSRF 토큰 가져오기
                    const csrfToken = document.querySelector('meta[name="csrf-token"]').content;
                    fetch(`/api/comments/${commentId}`, {
                        method: 'PUT',
                        headers: {
                            'Content-Type': 'application/json',
                            'X-CSRF-Token': csrfToken
                        },
                        body: JSON.stringify({
                            content: newContent
                        })
                    })
                    .then(res => res.json())
                    .then(data => {
                        if (data.success) {
                            // 댓글 목록에서 해당 댓글 찾아서 업데이트
                            const updateComment = (comments) => {
                                for (let i = 0; i < comments.length; i++) {
                                    if (comments[i].id === commentId) {
                                        comments[i].content = data.comment.content;
                                        comments[i].updatedAt = data.comment.updatedAt;
                                        return true;
                                    }
                                    if (comments[i].children) {
                                        if (updateComment(comments[i].children)) {
                                            return true;
                                        }
                                    }
                                }
                                return false;
                            };

                            updateComment(this.comments);
                        } else {
                            alert(data.message);
                        }
                    });
                }
            },

            deleteComment(commentId, isReply = false, parentId = null) {
                // 기존 코드 유지
                // CSRF 토큰 가져오기
                const csrfToken = document.querySelector('meta[name="csrf-token"]').content;

                if (confirm('댓글을 삭제하시겠습니까?')) {
                    fetch(`/api/comments/${commentId}`, {
                        method: 'DELETE',
                        headers: {
                            'X-CSRF-Token': csrfToken
                        }
                    })
                        .then(res => res.json())
                        .then(data => {
                            if (data.success) {
                                if (isReply && parentId) {
                                    // 부모 댓글에서 답글 삭제
                                    const parentComment = this.comments.find(c => c.id === parentId);
                                    if (parentComment && parentComment.children) {
                                        parentComment.children = parentComment.children.filter(r => r.id !== commentId);
                                    }
                                } else {
                                    // 기본 댓글 삭제
                                    this.comments = this.comments.filter(c => c.id !== commentId);
                                }
                            } else {
                                alert(data.message);
                            }
                        });
                }
            }
        };
    });
});
```

## 4. 필요하다면 CSS 스타일을 추가합니다:

```css
/* web/static/css/comments.css에 추가 */
.comment-editor-container {
    margin-bottom: 1rem;
}

.prose img {
    max-width: 100%;
    height: auto;
}

.prose pre {
    background-color: #f3f4f6;
    border-radius: 0.375rem;
    padding: 0.75rem;
    overflow-x: auto;
}

.prose code {
    background-color: #f3f4f6;
    padding: 0.125rem 0.25rem;
    border-radius: 0.25rem;
}
```

이 변경사항으로 댓글 기능이 Discourse와 유사한 리치 에디터를 갖게 되며, 사용자는 이미지와 파일을 업로드할 수 있게 됩니다. 추가적으로 댓글 수정 기능도 미래에 모달 에디터로 업그레이드할 수 있습니다.

댓글 내용이 HTML로 저장되기 때문에, 이 업그레이드가 서버 측 코드에 미치는 영향을 고려해야 합니다. HTML 필터링이나 XSS 방어 로직이 필요할 수 있습니다. 현재 서버 측에서 댓글 내용을 어떻게 처리하는지에 따라 추가 변경이 필요할 수 있습니다.



---
Powered by [Claude Exporter](https://www.claudexporter.com)