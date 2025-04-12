# Claude Conversation

## Prompt:
migrations 폴더에 mysql, postgres 폴더에 goose 를 이용한 table creation 코드가 있는데 만들어진 테이블들이 죄다 created_at, update_at 이 앞쪽에 있어서 가독성이 떨어진다. 005_timestamp.sql 식으로 created_at, update_at을 제일 마지막에 놓이게 하고 싶다.

실제로 004_anonymous_access.sql은 굳이 단독파일로 놓일 필요가 없어 보이는데 저렇게 있어서 위 필드 뒤쪽에 배치가 되어버리잖아? comments 허용 필드도 마찬가지고.

필요하면 전체적으로 다시 재구성 해줘.

대답 항상 한글로.
제안 코드는 항상 생략 없는 완전한 코드로.

## Response:

요청 내용을 명확히 이해해보겠습니다. 사용자는:

1. 모든 데이터베이스 테이블에서 created_at, updated_at 필드가 앞쪽에 위치해서 가독성이 떨어진다고 지적하며, 이 필드들을 테이블 정의의 마지막으로 이동하길 원합니다.

2. 004_anonymous_access.sql과 같은 파일이 단독 파일로 존재하는 것이 비효율적이라고 생각하며, 특히 이런 작은 변경사항들(allow_anonymous 필드나 comments_enabled 필드 추가)을 별도 파일로 분리하는 것이 적절하지 않다고 봅니다.

3. 전체 마이그레이션 파일을 재구성해달라고 요청하고 있습니다.

이 요청에 따라 마이그레이션 파일을 재구성하려면:

1. 기존 마이그레이션 파일들의 구조와 내용을 분석합니다.
2. 각 테이블에서 created_at, updated_at 필드를 마지막으로 이동시킵니다.
3. 작은 변경사항들(anonymous_access, comments_enabled 등)을 별도 파일로 유지하지 않고 관련 테이블 생성 시에 한꺼번에 포함시킵니다.
4. goose 마이그레이션 포맷(Up/Down)을 유지하며 새 파일 구조를 만듭니다.

현재 마이그레이션 파일 구조는:
- 001_create_tables.sql: 기본 테이블(users, boards, board_fields) 생성
- 002_comments.sql: comments_enabled 필드 추가 및 comments 테이블 생성
- 003_file_uploads.sql: attachments 테이블 생성
- 004_anonymous_access.sql: allow_anonymous 필드 추가

재구성 계획:
1. 001_create_tables.sql: created_at, updated_at을 마지막에 위치시키고, comments_enabled와 allow_anonymous 필드를 boards 테이블 정의에 통합합니다.
2. 002_comments.sql: comments 테이블 생성, created_at과 updated_at을 마지막에 위치시킵니다.
3. 003_file_uploads.sql: attachments 테이블 생성, created_at을 마지막에 위치시킵니다.
4. 004_anonymous_access.sql 파일은 제거하고 내용을 001_create_tables.sql로 통합합니다.

MySQL과 PostgreSQL 모두에 대해 동일한 방식으로 재구성합니다.

이제 파일별로 구체적인 변경 내용을 작성해 보겠습니다.

마이그레이션 파일의 구조를 개선하고, `created_at`, `updated_at` 필드를 테이블 정의의 마지막으로 이동시키려는 요청을 이해했습니다. 작은 변경사항들이 별도 파일로 분리되어 있는 것도 비효율적으로 보이네요. 다음과 같이 마이그레이션 파일들을 재구성하겠습니다.

## 1. MySQL 마이그레이션 파일 재구성

### 001_create_tables.sql (MySQL)

```sql
-- migrations/mysql/001_create_tables.sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(100) NOT NULL,
    full_name VARCHAR(100),
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS boards (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    board_type VARCHAR(20) NOT NULL,
    table_name VARCHAR(50) NOT NULL UNIQUE,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    comments_enabled BOOLEAN NOT NULL DEFAULT true,
    allow_anonymous BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS board_fields (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    board_id INT NOT NULL,
    name VARCHAR(50) NOT NULL,
    column_name VARCHAR(50) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    field_type VARCHAR(20) NOT NULL,
    required BOOLEAN NOT NULL DEFAULT FALSE,
    sortable BOOLEAN NOT NULL DEFAULT FALSE,
    searchable BOOLEAN NOT NULL DEFAULT FALSE,
    options TEXT,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (board_id, name),
    UNIQUE (board_id, column_name),
    FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS board_fields;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS boards;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
```

### 002_comments.sql (MySQL)

```sql
-- migrations/mysql/002_comments.sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE comments (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    post_id INT NOT NULL,
    board_id INT NOT NULL,
    user_id INT NOT NULL,
    content TEXT NOT NULL,
    parent_id INT DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose StatementBegin
-- Add index for post_id
CREATE INDEX idx_comments_post_id ON comments (post_id);
-- +goose StatementEnd

-- +goose StatementBegin
-- Add index for board_id
CREATE INDEX idx_comments_board_id ON comments (board_id);
-- +goose StatementEnd

-- +goose StatementBegin
-- Add index for parent_id
CREATE INDEX idx_comments_parent_id ON comments (parent_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS comments;
-- +goose StatementEnd
```

### 003_file_uploads.sql (MySQL)

```sql
-- migrations/mysql/003_file_uploads.sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE attachments (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    board_id BIGINT NOT NULL,
    post_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    storage_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    is_image BOOLEAN NOT NULL DEFAULT false,
    download_url VARCHAR(255) NOT NULL,
    download_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_attachments_post ON attachments(board_id, post_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS attachments;
-- +goose StatementEnd
```

## 2. PostgreSQL 마이그레이션 파일 재구성

### 001_create_tables.sql (PostgreSQL)

```sql
-- migrations/postgres/001_create_tables.sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(100) NOT NULL,
    full_name VARCHAR(100),
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS boards (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    board_type VARCHAR(20) NOT NULL,
    table_name VARCHAR(50) NOT NULL UNIQUE,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    comments_enabled BOOLEAN NOT NULL DEFAULT true,
    allow_anonymous BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS board_fields (
    id SERIAL PRIMARY KEY,
    board_id INT NOT NULL,
    name VARCHAR(50) NOT NULL,
    column_name VARCHAR(50) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    field_type VARCHAR(20) NOT NULL,
    required BOOLEAN NOT NULL DEFAULT FALSE,
    sortable BOOLEAN NOT NULL DEFAULT FALSE,
    searchable BOOLEAN NOT NULL DEFAULT FALSE,
    options TEXT,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (board_id, name),
    UNIQUE (board_id, column_name),
    FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS board_fields CASCADE;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS boards CASCADE;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS users CASCADE;
-- +goose StatementEnd
```

### 002_comments.sql (PostgreSQL)

```sql
-- migrations/postgres/002_comments.sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE comments (
    id SERIAL PRIMARY KEY,
    post_id INT NOT NULL,
    board_id INT NOT NULL,
    user_id INT NOT NULL,
    content TEXT NOT NULL,
    parent_id INT DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose StatementBegin
-- Add index for post_id
CREATE INDEX idx_comments_post_id ON comments (post_id);
-- +goose StatementEnd

-- +goose StatementBegin
-- Add index for board_id
CREATE INDEX idx_comments_board_id ON comments (board_id);
-- +goose StatementEnd

-- +goose StatementBegin
-- Add index for parent_id
CREATE INDEX idx_comments_parent_id ON comments (parent_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS comments;
-- +goose StatementEnd
```

### 003_file_uploads.sql (PostgreSQL)

```sql
-- migrations/postgres/003_file_uploads.sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE attachments (
    id SERIAL PRIMARY KEY,
    board_id BIGINT NOT NULL,
    post_id BIGINT NOT NULL, 
    user_id BIGINT NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    storage_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    is_image BOOLEAN NOT NULL DEFAULT false,
    download_url VARCHAR(255) NOT NULL,
    download_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_attachments_post ON attachments(board_id, post_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS attachments;
-- +goose StatementEnd
```

## 변경 사항 요약

1. 모든 테이블에서 `created_at`과 `updated_at` 필드를 테이블 정의의 마지막으로 이동했습니다.
2. `004_anonymous_access.sql` 파일을 제거하고 `allow_anonymous` 필드를 `boards` 테이블 생성 스크립트에 직접 포함시켰습니다.
3. `comments_enabled` 필드도 `002_comments.sql`에서 분리하여 `boards` 테이블 생성 시에 함께 추가되도록 했습니다.
4. 마이그레이션 파일 수를 3개로 줄였습니다.

이렇게 하면 테이블 구조가 더 명확해지고, 불필요한 마이그레이션 파일을 줄여서 관리가 편해질 것입니다. 특히 `created_at`과 `updated_at` 필드가 마지막에 위치하여 가독성이 향상되었습니다.

## Prompt:
굳! 이제  `q&a 유형`도 완성하자.

## Response:

Q&A(질문 답변) 유형의 게시판을 완성하려면 몇 가지 핵심 기능이 필요합니다:

1. 질문 작성 및 조회 기능
2. 답변 작성 및 조회 기능
3. 채택된 답변 표시 기능
4. 답변 상태 관리 (해결됨/미해결됨) 

제공된 코드를 살펴보면 이미 "board_type"에 "qna"라는 유형이 있는 것을 확인할 수 있습니다. 그러나 실제 템플릿은 일반 게시판과 갤러리 게시판 템플릿만 있고 Q&A 전용 템플릿은 완성되지 않은 상태인 것 같습니다.

다음 부분을 구현해야 합니다:

1. Q&A 게시판 전용 템플릿 (`web/templates/board/qna_posts.html`)
2. Q&A 게시물 조회 템플릿 (`web/templates/board/qna_view.html`)
3. 답변 채택 기능을 위한 API 핸들러
4. Q&A 게시물 상태 관리를 위한 필드 추가
5. 채택된 답변 표시 기능

먼저 Q&A 게시판을 위한 테이블 구조 변경이 필요한지 살펴봅시다. 다음과 같은 필드를 추가하는 것이 좋을 것 같습니다:
- 게시물에 'is_solved' 필드 (질문이 해결되었는지 여부)
- 댓글에 'is_accepted' 필드 (채택된 답변인지 여부)

이를 위해 기존 마이그레이션 파일에 필드를 추가하거나 새 마이그레이션 파일을 만들어야 합니다.

지금 정리된 방식에서 테이블을 수정하는 것보다는 Q&A 기능을 위한 새로운 마이그레이션 파일을 추가하는 것이 좋겠습니다.

Q&A 게시판 유형을 완성하겠습니다. 이를 위해 다음 작업들이 필요합니다:

1. Q&A 게시판을 위한 테이블 변경 마이그레이션
2. Q&A 게시판 전용 템플릿
3. 질문 해결 상태 및 답변 채택 기능 구현

## 1. 마이그레이션 파일 추가 - Q&A 기능 추가

### MySQL 마이그레이션 (004_qna_features.sql)

```sql
-- migrations/mysql/004_qna_features.sql
-- +goose Up
-- +goose StatementBegin
-- comments 테이블에 답변 채택 기능 추가
ALTER TABLE comments ADD COLUMN is_accepted BOOLEAN NOT NULL DEFAULT false;
-- +goose StatementEnd

-- +goose StatementBegin
-- 게시물 테이블 변경을 위한 함수 생성
DROP PROCEDURE IF EXISTS add_is_solved_to_all_board_tables;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE PROCEDURE add_is_solved_to_all_board_tables()
BEGIN
    DECLARE done INT DEFAULT FALSE;
    DECLARE table_name VARCHAR(100);
    DECLARE cur CURSOR FOR 
        SELECT b.table_name 
        FROM boards b 
        WHERE b.table_name IS NOT NULL AND b.table_name != '';
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = TRUE;
    
    OPEN cur;
    read_loop: LOOP
        FETCH cur INTO table_name;
        IF done THEN
            LEAVE read_loop;
        END IF;
        
        SET @alter_sql = CONCAT('ALTER TABLE ', table_name, ' ADD COLUMN is_solved BOOLEAN NOT NULL DEFAULT false');
        PREPARE stmt FROM @alter_sql;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
    END LOOP;
    CLOSE cur;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CALL add_is_solved_to_all_board_tables();
-- +goose StatementEnd

-- +goose StatementBegin
DROP PROCEDURE IF EXISTS add_is_solved_to_all_board_tables;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE comments DROP COLUMN IF EXISTS is_accepted;
-- +goose StatementEnd

-- +goose StatementBegin
-- 동적으로 모든 게시판 테이블에서 is_solved 컬럼 제거
DROP PROCEDURE IF EXISTS remove_is_solved_from_all_board_tables;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE PROCEDURE remove_is_solved_from_all_board_tables()
BEGIN
    DECLARE done INT DEFAULT FALSE;
    DECLARE table_name VARCHAR(100);
    DECLARE cur CURSOR FOR 
        SELECT b.table_name 
        FROM boards b 
        WHERE b.table_name IS NOT NULL AND b.table_name != '';
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = TRUE;
    
    OPEN cur;
    read_loop: LOOP
        FETCH cur INTO table_name;
        IF done THEN
            LEAVE read_loop;
        END IF;
        
        SET @alter_sql = CONCAT('ALTER TABLE ', table_name, ' DROP COLUMN IF EXISTS is_solved');
        PREPARE stmt FROM @alter_sql;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
    END LOOP;
    CLOSE cur;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CALL remove_is_solved_from_all_board_tables();
-- +goose StatementEnd

-- +goose StatementBegin
DROP PROCEDURE IF EXISTS remove_is_solved_from_all_board_tables;
-- +goose StatementEnd
```

### PostgreSQL 마이그레이션 (004_qna_features.sql)

```sql
-- migrations/postgres/004_qna_features.sql
-- +goose Up
-- +goose StatementBegin
-- comments 테이블에 답변 채택 기능 추가
ALTER TABLE comments ADD COLUMN is_accepted BOOLEAN NOT NULL DEFAULT false;
-- +goose StatementEnd

-- +goose StatementBegin
-- 동적으로 모든 게시판 테이블에 is_solved 컬럼 추가
CREATE OR REPLACE FUNCTION add_is_solved_to_all_board_tables() RETURNS void AS $$
DECLARE
    r RECORD;
BEGIN
    FOR r IN SELECT table_name FROM boards WHERE table_name IS NOT NULL AND table_name != '' LOOP
        EXECUTE format('ALTER TABLE %I ADD COLUMN is_solved BOOLEAN NOT NULL DEFAULT false', r.table_name);
    END LOOP;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
SELECT add_is_solved_to_all_board_tables();
-- +goose StatementEnd

-- +goose StatementBegin
DROP FUNCTION IF EXISTS add_is_solved_to_all_board_tables();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE comments DROP COLUMN IF EXISTS is_accepted;
-- +goose StatementEnd

-- +goose StatementBegin
-- 동적으로 모든 게시판 테이블에서 is_solved 컬럼 제거
CREATE OR REPLACE FUNCTION remove_is_solved_from_all_board_tables() RETURNS void AS $$
DECLARE
    r RECORD;
BEGIN
    FOR r IN SELECT table_name FROM boards WHERE table_name IS NOT NULL AND table_name != '' LOOP
        EXECUTE format('ALTER TABLE %I DROP COLUMN IF EXISTS is_solved', r.table_name);
    END LOOP;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
SELECT remove_is_solved_from_all_board_tables();
-- +goose StatementEnd

-- +goose StatementBegin
DROP FUNCTION IF EXISTS remove_is_solved_from_all_board_tables();
-- +goose StatementEnd
```

## 2. Q&A 게시판 템플릿 - 게시물 목록

```html
<!-- web/templates/board/qna_posts.html -->
<div>
    <div class="flex justify-between items-center mb-6">
        <h1 class="text-2xl font-bold">{{.board.Name}}</h1>

        {{if .user}}
        <a href="/boards/{{.board.ID}}/posts/create" class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
            질문하기
        </a>
        {{end}}
    </div>

    {{if .board.Description}}
    <div class="bg-white shadow rounded-lg p-4 mb-6">
        <p class="text-gray-600">{{.board.Description}}</p>
    </div>
    {{end}}

    <!-- 검색 폼 -->
    <div class="mb-6">
        <form action="/boards/{{.board.ID}}/posts" method="GET" class="flex">
            <input type="text" name="q" value="{{.query}}" placeholder="검색어를 입력하세요" class="shadow-sm focus:ring-blue-500 focus:border-blue-500 block w-full sm:text-sm border-gray-300 rounded-md">
            <button type="submit" class="ml-2 inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                검색
            </button>
        </form>
    </div>

    <!-- 필터 옵션 -->
    <div class="mb-6 flex space-x-4">
        <a href="/boards/{{.board.ID}}/posts?filter=all{{if .query}}&q={{.query}}{{end}}" class="inline-flex items-center px-3 py-1 border border-transparent text-sm font-medium rounded-md {{if eq .filter "all"}}bg-blue-100 text-blue-800{{else}}bg-gray-100 text-gray-800{{end}}">
            전체
        </a>
        <a href="/boards/{{.board.ID}}/posts?filter=solved{{if .query}}&q={{.query}}{{end}}" class="inline-flex items-center px-3 py-1 border border-transparent text-sm font-medium rounded-md {{if eq .filter "solved"}}bg-green-100 text-green-800{{else}}bg-gray-100 text-gray-800{{end}}">
            해결됨
        </a>
        <a href="/boards/{{.board.ID}}/posts?filter=unsolved{{if .query}}&q={{.query}}{{end}}" class="inline-flex items-center px-3 py-1 border border-transparent text-sm font-medium rounded-md {{if eq .filter "unsolved"}}bg-yellow-100 text-yellow-800{{else}}bg-gray-100 text-gray-800{{end}}">
            미해결
        </a>
    </div>

    {{if not .posts}}
    <div class="bg-white shadow rounded-lg p-6 text-center">
        <p class="text-gray-500">등록된 질문이 없습니다.</p>
    </div>
    {{else}}
    <!-- 게시물 목록 -->
    <div class="bg-white shadow overflow-hidden sm:rounded-md">
        <ul class="divide-y divide-gray-200">
            {{range .posts}}
            <li>
                <a href="/boards/{{$.board.ID}}/posts/{{.ID}}" class="block hover:bg-gray-50">
                    <div class="px-4 py-4 sm:px-6">
                        <div class="flex items-center justify-between">
                            <div class="flex-1 min-w-0">
                                <div class="flex items-center">
                                    {{if .RawData.is_solved}}
                                    <span class="mr-2 px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                                        해결됨
                                    </span>
                                    {{else}}
                                    <span class="mr-2 px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800">
                                        미해결
                                    </span>
                                    {{end}}
                                    <p class="text-sm font-medium text-blue-600 truncate">
                                        {{toUTF8 .Title}}
                                    </p>
                                </div>
                                <p class="mt-1 text-xs text-gray-500 truncate">
                                    {{.Username}} · 조회수 {{.ViewCount}}
                                </p>
                            </div>
                            <div class="ml-2 flex-shrink-0 flex">
                                <div class="flex items-center space-x-2">
                                    <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-blue-100 text-blue-800">
                                        답변 {{.RawData.answer_count|default "0"}}
                                    </span>
                                    <p class="px-2 inline-flex text-xs leading-5 font-semibold">
                                        {{.CreatedAt.Format "2006-01-02 15:04"}}
                                    </p>
                                </div>
                            </div>
                        </div>
                    </div>
                </a>
            </li>
            {{end}}
        </ul>
    </div>

    <!-- 페이지네이션 -->
    {{if gt .totalPages 1}}
    <div class="mt-6">
        <nav class="flex justify-center">
            <ul class="flex">
                <!-- 이전 페이지 -->
                {{if gt .page 1}}
                <li>
                    <a href="/boards/{{.board.ID}}/posts?page={{add .page -1}}{{if .query}}&q={{.query}}{{end}}{{if .filter}}&filter={{.filter}}{{end}}" class="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
                        이전
                    </a>
                </li>
                {{end}}

                <!-- 페이지 번호 -->
                {{range $i := iterate 1 .totalPages}}
                <li>
                    <a href="/boards/{{$.board.ID}}/posts?page={{$i}}{{if $.query}}&q={{$.query}}{{end}}{{if $.filter}}&filter={{$.filter}}{{end}}" class="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md {{if eq $i $.page}}bg-blue-50 text-blue-600 border-blue-500{{else}}text-gray-700 bg-white hover:bg-gray-50{{end}}">
                        {{$i}}
                    </a>
                </li>
                {{end}}

                <!-- 다음 페이지 -->
                {{if lt .page .totalPages}}
                <li>
                    <a href="/boards/{{.board.ID}}/posts?page={{add .page 1}}{{if .query}}&q={{$.query}}{{end}}{{if $.filter}}&filter={{$.filter}}{{end}}" class="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
                        다음
                    </a>
                </li>
                {{end}}
            </ul>
        </nav>
    </div>
    {{end}}
    {{end}}
</div>
```

## 3. Q&A 게시물 조회 템플릿

```html
<!-- web/templates/board/qna_view.html -->
<input type="hidden" id="boardId" value="{{.board.ID}}">
<input type="hidden" id="postId" value="{{.post.ID}}">
<input type="hidden" id="csrfToken" value="{{.csrf}}">

<div>
    <div class="mb-4">
        <nav class="flex" aria-label="Breadcrumb">
            <ol class="flex items-center space-x-4">
                <li>
                    <div>
                        <a href="/boards" class="text-gray-400 hover:text-gray-500">게시판 목록</a>
                    </div>
                </li>
                <li>
                    <div class="flex items-center">
                        <svg class="flex-shrink-0 h-5 w-5 text-gray-300" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 20 20" aria-hidden="true">
                            <path d="M5.555 17.776l8-16 .894.448-8 16-.894-.448z" />
                        </svg>
                        <a href="/boards/{{.board.ID}}/posts" class="ml-4 text-gray-400 hover:text-gray-500">{{.board.Name}}</a>
                    </div>
                </li>
            </ol>
        </nav>
    </div>

    <div class="bg-white shadow rounded-lg overflow-hidden">
        <!-- 질문 상태 표시 -->
        <div class="bg-gray-50 px-6 py-2 flex items-center justify-between">
            <div class="flex items-center">
                {{if .post.RawData.is_solved}}
                <span class="px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                    해결된 질문
                </span>
                {{else}}
                <span class="px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800">
                    미해결 질문
                </span>
                {{end}}
            </div>

            <!-- 질문 해결 마크/해제 버튼 (질문 작성자 또는 관리자만) -->
            {{if .user}}
            {{if or (eq .user.ID .post.UserID) (eq .user.Role "admin")}}
            <div x-data="{ isSolved: {{if .post.RawData.is_solved}}true{{else}}false{{end}} }">
                <button 
                    @click="markQuestionSolved({{.post.ID}}, !isSolved)" 
                    class="inline-flex items-center px-3 py-1 text-sm font-medium rounded-md"
                    :class="isSolved ? 'text-yellow-700 bg-yellow-100 hover:bg-yellow-200' : 'text-green-700 bg-green-100 hover:bg-green-200'"
                >
                    <svg class="w-4 h-4 mr-1" :class="isSolved ? 'text-yellow-500' : 'text-green-500'" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                    </svg>
                    <span x-text="isSolved ? '미해결로 표시' : '해결됨으로 표시'"></span>
                </button>
            </div>
            {{end}}
            {{end}}
        </div>

        <!-- 게시물 헤더 -->
        <div class="border-b border-gray-200 px-6 py-4">
            <h1 class="text-2xl font-bold text-gray-900">{{.post.Title}}</h1>
            <div class="mt-2 flex justify-between items-center">
                <div class="flex items-center">
                    <div class="text-sm text-gray-500">
                        <span>{{.post.Username}}</span>
                        <span class="mx-1">·</span>
                        <span>{{.post.CreatedAt.Format "2006-01-02 15:04"}}</span>
                        {{if ne .post.UpdatedAt .post.CreatedAt}}
                        <span class="mx-1">·</span>
                        <span>수정됨: {{.post.UpdatedAt.Format "2006-01-02 15:04"}}</span>
                        {{end}}
                    </div>
                </div>
                <div class="text-sm text-gray-500">
                    <span>조회수 {{.post.ViewCount}}</span>
                </div>
            </div>
        </div>

        <!-- 게시물 내용 -->
        <div class="px-6 py-4">
            <!-- 본문 -->
            <div class="prose max-w-none">
                <!-- HTML 태그를 그대로 렌더링하기 위해 안전한 방식으로 출력 -->
                <div id="post-content"></div>
                <script>
                    // 서버에서 받은 HTML 내용을 안전하게 설정
                    document.getElementById('post-content').innerHTML = `{{.post.Content}}`;
                </script>
            </div>

            <!-- 동적 필드 정보 -->
            {{if .board.Fields}}
            <div class="mt-6 grid grid-cols-2 gap-4 bg-gray-50 p-4 rounded-lg">
                {{range $fieldIndex, $boardField := .board.Fields}}
                {{with index $.post.Fields $boardField.Name}}
                <div>
                    <div class="text-sm font-medium text-gray-500">{{$boardField.DisplayName}}</div>
                    <div class="mt-1 text-sm text-gray-900">
                        {{if eq $boardField.FieldType "checkbox"}}
                        {{if .Value}}
                        <svg class="h-5 w-5 text-green-500" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                        </svg>
                        {{else}}
                        <svg class="h-5 w-5 text-gray-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
                        </svg>
                        {{end}}
                        {{else if eq $boardField.FieldType "file"}}
                        <!-- 파일 필드 표시 -->
                        <div id="attachments-container"></div>
                        {{else}}
                        {{.Value}}
                        {{end}}
                    </div>
                </div>
                {{end}}
                {{end}}
            </div>
            {{end}}

            <!-- 첨부 파일 목록 -->
            <div class="mt-6" x-data="{ attachments: [] }" x-init="
                fetch('/api/boards/{{.board.ID}}/posts/{{.post.ID}}/attachments')
                    .then(res => res.json())
                    .then(data => {
                        if (data.success) attachments = data.attachments || [];
                    })
            ">
                <h3 class="text-lg font-medium text-gray-900">첨부 파일</h3>
                <div x-show="attachments.length === 0" class="mt-2 text-sm text-gray-500">첨부된 파일이 없습니다.</div>
                <ul x-show="attachments.length > 0" class="mt-2 border border-gray-200 rounded-md divide-y divide-gray-200">
                    <template x-for="file in attachments" :key="file.id">
                        <li class="pl-3 pr-4 py-3 flex items-center justify-between text-sm">
                            <div class="w-0 flex-1 flex items-center">
                                <svg class="flex-shrink-0 h-5 w-5 text-gray-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                                    <path fill-rule="evenodd" d="M8 4a3 3 0 00-3 3v4a5 5 0 0010 0V7a1 1 0 112 0v4a7 7 0 11-14 0V7a5 5 0 0110 0v4a3 3 0 11-6 0V7a1 1 0 012 0v4a1 1 0 102 0V7a3 3 0 00-3-3z" clip-rule="evenodd" />
                                </svg>
                                <span class="ml-2 flex-1 w-0 truncate" x-text="file.fileName"></span>
                            </div>
                            <div class="ml-4 flex-shrink-0">
                                <a :href="'/attachments/' + file.id + '/download'" class="font-medium text-blue-600 hover:text-blue-500">
                                    다운로드
                                </a>
                            </div>
                        </li>
                    </template>
                </ul>
            </div>
        </div>

        <!-- 게시물 푸터 -->
        <div class="px-6 py-4 bg-gray-50 border-t border-gray-200">
            <div class="flex justify-between">
                <div>
                    <a href="/boards/{{.board.ID}}/posts" class="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
                        목록
                    </a>
                </div>

                {{if .user}}
                {{if or (eq .user.ID .post.UserID) (eq .user.Role "admin")}}
                <div class="flex space-x-2">
                    <a href="/boards/{{.board.ID}}/posts/{{.post.ID}}/edit" class="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
                        수정
                    </a>
                    <button x-data="postActions()" @click="deletePost()" class="inline-flex items-center px-4 py-2 border border-red-300 text-sm font-medium rounded-md text-red-700 bg-white hover:bg-red-50">
                        삭제
                    </button>
                </div>
                {{end}}
                {{end}}
            </div>
        </div>
    </div>
</div>

<!-- 답변 섹션 -->
<div x-data="qnaCommentSystem({{.board.CommentsEnabled}})" class="mt-8 border-t border-gray-200 pt-8">
    <h3 class="text-lg font-medium text-gray-900 flex items-center">
        <span>답변</span>
        <span class="ml-2 text-gray-500 text-sm" x-text="comments.length + ' 개'"></span>
    </h3>

    <!-- 로딩 표시 -->
    <div x-show="loading" class="text-center py-4">
        <svg class="animate-spin h-5 w-5 text-gray-500 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <p class="mt-2 text-sm text-gray-500">답변을 불러오는 중...</p>
    </div>

    <!-- 오류 메시지 -->
    <div x-show="error" class="bg-red-50 p-4 rounded-md">
        <div class="flex">
            <div class="flex-shrink-0">
                <svg class="h-5 w-5 text-red-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
                </svg>
            </div>
            <div class="ml-3">
                <p class="text-sm text-red-700" x-text="error"></p>
            </div>
        </div>
    </div>

    <!-- 답변 없음 메시지 -->
    <div x-show="!loading && !error && comments.length === 0" class="text-center py-10 bg-gray-50 rounded-lg mt-4">
        <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z"></path>
        </svg>
        <h3 class="mt-2 text-sm font-medium text-gray-900">아직 답변이 없습니다</h3>
        <p class="mt-1 text-sm text-gray-500">이 질문에 첫 답변을 작성해보세요.</p>
    </div>

    <!-- 답변 목록 -->
    <div class="mt-4 space-y-6">
        <template x-for="comment in comments" :key="comment.id">
            <div class="bg-white rounded-lg shadow overflow-hidden"
                :class="{'border-2 border-green-500': comment.is_accepted}">
                <!-- 채택된 답변 배지 -->
                <div x-show="comment.is_accepted" class="bg-green-100 px-4 py-2 border-b border-green-200">
                    <div class="flex items-center">
                        <svg class="h-5 w-5 text-green-500 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                        </svg>
                        <span class="font-medium text-green-800">채택된 답변</span>
                    </div>
                </div>

                <!-- 답변 내용 -->
                <div class="p-4">
                    <div class="flex space-x-3">
                        <div class="flex-shrink-0">
                            <div class="h-10 w-10 rounded-full bg-gray-200 flex items-center justify-center">
                                <span class="text-sm font-medium text-gray-500" x-text="comment.user ? comment.user.username.charAt(0) : '?'"></span>
                            </div>
                        </div>
                        <div class="min-w-0 flex-1">
                            <div class="flex items-center justify-between">
                                <div class="text-sm font-medium text-gray-900" x-text="comment.user ? comment.user.username : '알 수 없음'"></div>
                                
                                <!-- 질문 작성자만 답변 채택 가능 -->
                                <div x-show="{{if .user}}{{if eq .user.ID .post.UserID}}true{{else}}false{{end}}{{else}}false{{end}} && !comment.is_accepted && !{{.post.RawData.is_solved}}">
                                    <button 
                                        @click="acceptAnswer(comment.id)"
                                        class="inline-flex items-center px-3 py-1 border border-transparent text-xs font-medium rounded-md text-green-700 bg-green-100 hover:bg-green-200"
                                    >
                                        <svg class="mr-1.5 h-4 w-4 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
                                        </svg>
                                        답변 채택
                                    </button>
                                </div>
                                
                                <!-- 관리자/글 작성자가 채택된 답변을 취소할 수 있는 버튼 -->
                                <div x-show="comment.is_accepted && {{if .user}}{{if or (eq .user.ID .post.UserID) (eq .user.Role "admin")}}true{{else}}false{{end}}{{else}}false{{end}}">
                                    <button 
                                        @click="unacceptAnswer(comment.id)"
                                        class="inline-flex items-center px-3 py-1 border border-transparent text-xs font-medium rounded-md text-yellow-700 bg-yellow-100 hover:bg-yellow-200"
                                    >
                                        <svg class="mr-1.5 h-4 w-4 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                                        </svg>
                                        채택 취소
                                    </button>
                                </div>
                            </div>
                            
                            <div class="mt-1 text-sm text-gray-700" x-html="comment.content.replace(/\n/g, '<br>')"></div>
                            
                            <div class="mt-2 text-xs text-gray-500 flex items-center justify-between">
                                <span x-text="new Date(comment.createdAt).toLocaleString()"></span>
                                
                                <!-- 수정/삭제 버튼 (작성자/관리자만) -->
                                <template x-if="{{if .user}}{{.user.ID}}{{else}}0{{end}} === comment.userId || {{if eq .user.Role "admin"}}true{{else}}false{{end}}">
                                    <div class="flex items-center space-x-2">
                                        <button class="text-gray-600 hover:text-gray-800" @click="editComment(comment.id, comment.content)">
                                            수정
                                        </button>
                                        <button class="text-red-600 hover:text-red-800" @click="deleteComment(comment.id)">
                                            삭제
                                        </button>
                                    </div>
                                </template>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </template>
    </div>

    <!-- 답변 작성 폼 -->
    {{if .user}}
    <div class="mt-8">
        <h3 class="text-lg font-medium text-gray-900">답변 작성</h3>
        <div class="mt-3">
            <form @submit.prevent="submitComment()">
                <div>
                    <label for="comment" class="sr-only">답변 내용</label>
                    <textarea id="comment" name="comment" rows="5" x-model="commentContent" class="block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm" placeholder="답변을 입력하세요..." required></textarea>
                </div>
                <div class="mt-3 flex items-center justify-end">
                    <button type="submit" class="inline-flex items-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2">
                        답변 등록
                    </button>
                </div>
            </form>
        </div>
    </div>
    {{else}}
    <div class="mt-8 bg-gray-50 p-4 rounded-md text-center">
        <p class="text-gray-600">답변을 작성하려면 <a href="/auth/login?redirect=/boards/{{.board.ID}}/posts/{{.post.ID}}" class="text-blue-600 hover:text-blue-500">로그인</a>이 필요합니다.</p>
    </div>
    {{end}}

    <!-- Q&A 관련 스크립트 -->
    <script>
        // QnA 전용 댓글 시스템
        function qnaCommentSystem(commentsEnabled) {
            return {
                comments: [],
                commentContent: '',
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
                            this.error = '답변을 불러오는 중 오류가 발생했습니다.';
                            this.loading = false;
                        });
                },

                submitComment() {
                    const boardId = document.getElementById('boardId').value;
                    const postId = document.getElementById('postId').value;
                    const csrfToken = document.getElementById('csrfToken').value;

                    fetch(`/api/boards/${boardId}/posts/${postId}/comments`, {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                            'X-CSRF-Token': csrfToken
                        },
                        body: JSON.stringify({
                            content: this.commentContent
                        })
                    })
                        .then(res => res.json())
                        .then(data => {
                            if (data.success) {
                                this.comments.push(data.comment);
                                this.commentContent = '';
                            } else {
                                alert(data.message);
                            }
                        });
                },

                editComment(commentId, content) {
                    const csrfToken = document.getElementById('csrfToken').value;
                    const newContent = prompt('답변을 수정합니다:', content);

                    if (newContent && newContent !== content) {
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
                                    const index = this.comments.findIndex(c => c.id === commentId);
                                    if (index !== -1) {
                                        this.comments[index].content = data.comment.content;
                                        this.comments[index].updatedAt = data.comment.updatedAt;
                                    }
                                } else {
                                    alert(data.message);
                                }
                            });
                    }
                },

                deleteComment(commentId) {
                    const csrfToken = document.getElementById('csrfToken').value;

                    if (confirm('이 답변을 삭제하시겠습니까?')) {
                        fetch(`/api/comments/${commentId}`, {
                            method: 'DELETE',
                            headers: {
                                'X-CSRF-Token': csrfToken
                            }
                        })
                            .then(res => res.json())
                            .then(data => {
                                if (data.success) {
                                    // 목록에서 제거
                                    this.comments = this.comments.filter(c => c.id !== commentId);
                                } else {
                                    alert(data.message);
                                }
                            });
                    }
                },

                acceptAnswer(commentId) {
                    const csrfToken = document.getElementById('csrfToken').value;
                    const boardId = document.getElementById('boardId').value;
                    const postId = document.getElementById('postId').value;

                    fetch(`/api/boards/${boardId}/posts/${postId}/comments/${commentId}/accept`, {
                        method: 'PUT',
                        headers: {
                            'Content-Type': 'application/json',
                            'X-CSRF-Token': csrfToken
                        }
                    })
                        .then(res => res.json())
                        .then(data => {
                            if (data.success) {
                                // 모든 답변의 채택 상태 초기화
                                this.comments.forEach(c => c.is_accepted = false);
                                
                                // 채택된 답변 표시
                                const index = this.comments.findIndex(c => c.id === commentId);
                                if (index !== -1) {
                                    this.comments[index].is_accepted = true;
                                }
                                
                                // 페이지 새로고침 (질문 상태도 업데이트)
                                window.location.reload();
                            } else {
                                alert(data.message);
                            }
                        });
                },

                unacceptAnswer(commentId) {
                    const csrfToken = document.getElementById('csrfToken').value;
                    const boardId = document.getElementById('boardId').value;
                    const postId = document.getElementById('postId').value;

                    fetch(`/api/boards/${boardId}/posts/${postId}/comments/${commentId}/unaccept`, {
                        method: 'PUT',
                        headers: {
                            'Content-Type': 'application/json',
                            'X-CSRF-Token': csrfToken
                        }
                    })
                        .then(res => res.json())
                        .then(data => {
                            if (data.success) {
                                // 채택 취소
                                const index = this.comments.findIndex(c => c.id === commentId);
                                if (index !== -1) {
                                    this.comments[index].is_accepted = false;
                                }
                                
                                // 페이지 새로고침 (질문 상태도 업데이트)
                                window.location.reload();
                            } else {
                                alert(data.message);
                            }
                        });
                }
            };
        }

        // 질문 해결/미해결 표시 함수
        function markQuestionSolved(postId, isSolved) {
            const csrfToken = document.getElementById('csrfToken').value;
            const boardId = document.getElementById('boardId').value;

            fetch(`/api/boards/${boardId}/posts/${postId}/solved`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': csrfToken
                },
                body: JSON.stringify({ solved: isSolved })
            })
                .then(res => res.json())
                .then(data => {
                    if (data.success) {
                        window.location.reload();
                    } else {
                        alert(data.message);
                    }
                });
        }
    </script>
</div>
```

## 4. CommentHandler에 추가할 메서드 (채택 관련 API)

이제 답변 채택 API를 구현하기 위해 CommentHandler에 다음 메서드를 추가해야 합니다:

```go
// 댓글 채택 API
func (h *CommentHandler) AcceptComment(c *fiber.Ctx) error {
    boardID, err := strconv.ParseInt(c.Params("boardID"), 10, 64)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "잘못된 게시판 ID입니다",
        })
    }

    postID, err := strconv.ParseInt(c.Params("postID"), 10, 64)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "잘못된 게시물 ID입니다",
        })
    }

    commentID, err := strconv.ParseInt(c.Params("commentID"), 10, 64)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "잘못된 댓글 ID입니다",
        })
    }

    // 현재 로그인한 사용자 정보
    user := c.Locals("user").(*models.User)

    // 질문 조회 (작성자만 답변 채택 가능)
    post, err := h.boardService.GetPost(c.Context(), boardID, postID)
    if err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "success": false,
            "message": "게시물을 찾을 수 없습니다",
        })
    }

    // 작성자 또는 관리자만 채택 가능
    if user.ID != post.UserID && user.Role != models.RoleAdmin {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
            "success": false,
            "message": "질문 작성자만 답변을 채택할 수 있습니다",
        })
    }

    // 댓글 채택 처리
    err = h.commentService.AcceptComment(c.Context(), commentID, postID)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "success": false,
            "message": "답변 채택에 실패했습니다: " + err.Error(),
        })
    }

    // 게시물 해결됨으로 표시
    post.RawData["is_solved"] = true
    err = h.boardService.UpdatePost(c.Context(), boardID, post)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "success": false,
            "message": "게시물 상태 업데이트에 실패했습니다: " + err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "success": true,
        "message": "답변이 채택되었습니다",
    })
}

// 댓글 채택 취소 API
func (h *CommentHandler) UnacceptComment(c *fiber.Ctx) error {
    boardID, err := strconv.ParseInt(c.Params("boardID"), 10, 64)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "잘못된 게시판 ID입니다",
        })
    }

    postID, err := strconv.ParseInt(c.Params("postID"), 10, 64)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "잘못된 게시물 ID입니다",
        })
    }

    commentID, err := strconv.ParseInt(c.Params("commentID"), 10, 64)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "잘못된 댓글 ID입니다",
        })
    }

    // 현재 로그인한 사용자 정보
    user := c.Locals("user").(*models.User)

    // 질문 조회 (작성자만 답변 채택 취소 가능)
    post, err := h.boardService.GetPost(c.Context(), boardID, postID)
    if err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "success": false,
            "message": "게시물을 찾을 수 없습니다",
        })
    }

    // 작성자 또는 관리자만 채택 취소 가능
    if user.ID != post.UserID && user.Role != models.RoleAdmin {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
            "success": false,
            "message": "질문 작성자만 답변 채택을 취소할 수 있습니다",
        })
    }

    // 댓글 채택 취소 처리
    err = h.commentService.UnacceptComment(c.Context(), commentID)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "success": false,
            "message": "답변 채택 취소에 실패했습니다: " + err.Error(),
        })
    }

    // 게시물 미해결로 표시
    post.RawData["is_solved"] = false
    err = h.boardService.UpdatePost(c.Context(), boardID, post)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "success": false,
            "message": "게시물 상태 업데이트에 실패했습니다: " + err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "success": true,
        "message": "답변 채택이 취소되었습니다",
    })
}

// 질문 해결 상태 변경 API
func (h *CommentHandler) SetQuestionSolved(c *fiber.Ctx) error {
    boardID, err := strconv.ParseInt(c.Params("boardID"), 10, 64)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "잘못된 게시판 ID입니다",
        })
    }

    postID, err := strconv.ParseInt(c.Params("postID"), 10, 64)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "잘못된 게시물 ID입니다",
        })
    }

    // 요청 본문 파싱
    var req struct {
        Solved bool `json:"solved"`
    }

    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "요청 데이터 파싱 실패",
        })
    }

    // 현재 로그인한 사용자 정보
    user := c.Locals("user").(*models.User)

    // 질문 조회 (작성자만 상태 변경 가능)
    post, err := h.boardService.GetPost(c.Context(), boardID, postID)
    if err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "success": false,
            "message": "게시물을 찾을 수 없습니다",
        })
    }

    // 작성자 또는 관리자만 상태 변경 가능
    if user.ID != post.UserID && user.Role != models.RoleAdmin {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
            "success": false,
            "message": "질문 작성자만 상태를 변경할 수 있습니다",
        })
    }

    // 게시물 해결 상태 업데이트
    post.RawData["is_solved"] = req.Solved
    err = h.boardService.UpdatePost(c.Context(), boardID, post)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "success": false,
            "message": "게시물 상태 업데이트에 실패했습니다: " + err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "success": true,
        "message": req.Solved ? "질문이 해결됨으로 표시되었습니다" : "질문이 미해결로 표시되었습니다",
    })
}
```

## 5. CommentService에 추가할 메서드

```go
// AcceptComment - 댓글을 채택된 답변으로 표시
func (s *commentService) AcceptComment(ctx context.Context, commentID, postID int64) error {
    // 이전에 채택된 답변이 있으면 채택 취소
    existingAccepted, err := s.commentRepo.GetAcceptedCommentByPostID(ctx, postID)
    if err == nil && existingAccepted != nil {
        // 채택 취소
        existingAccepted.IsAccepted = false
        err = s.commentRepo.Update(ctx, existingAccepted)
        if err != nil {
            return fmt.Errorf("기존 채택된 답변 상태 변경 실패: %w", err)
        }
    }

    // 댓글 조회
    comment, err := s.commentRepo.GetByID(ctx, commentID)
    if err != nil {
        return ErrCommentNotFound
    }

    // 답변 채택
    comment.IsAccepted = true
    return s.commentRepo.Update(ctx, comment)
}

// UnacceptComment - 채택된 답변 취소
func (s *commentService) UnacceptComment(ctx context.Context, commentID int64) error {
    // 댓글 조회
    comment, err := s.commentRepo.GetByID(ctx, commentID)
    if err != nil {
        return ErrCommentNotFound
    }

    // 채택 취소
    comment.IsAccepted = false
    return s.commentRepo.Update(ctx, comment)
}
```

## 6. CommentRepository에 추가할 메서드

```go
// GetAcceptedCommentByPostID - 게시물의 채택된 답변 조회
func (r *commentRepository) GetAcceptedCommentByPostID(ctx context.Context, postID int64) (*models.Comment, error) {
    comment := new(models.Comment)
    err := r.db.NewSelect().
        Model(comment).
        Where("post_id = ? AND is_accepted = ?", postID, true).
        Scan(ctx)
    if err != nil {
        return nil, err
    }
    return comment, nil
}
```

## 7. 라우트 추가 (cmd/main.go의 setupRoutes 함수에 추가)

```go
// Q&A 관련 API 라우트
commentsAPI.Put("/:commentID/accept", authMiddleware.RequireAuth, commentHandler.AcceptComment)
commentsAPI.Put("/:commentID/unaccept", authMiddleware.RequireAuth, commentHandler.UnacceptComment)
boardsAuthWithID.Put("/posts/:postID/solved", boardHandler.SetQuestionSolved)
```

## 8. board_handler.go에 추가할 메서드 (Q&A 목록 조회 및 해결 상태 변경)

```go
// SetQuestionSolved 질문 해결 상태 변경
func (h *BoardHandler) SetQuestionSolved(c *fiber.Ctx) error {
    boardID, err := strconv.ParseInt(c.Params("boardID"), 10, 64)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "잘못된 게시판 ID입니다",
        })
    }

    postID, err := strconv.ParseInt(c.Params("postID"), 10, 64)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "잘못된 게시물 ID입니다",
        })
    }

    // 요청 본문 파싱
    var req struct {
        Solved bool `json:"solved"`
    }

    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "요청 데이터 파싱 실패",
        })
    }

    // 현재 로그인한 사용자 정보
    user := c.Locals("user").(*models.User)

    // 질문 조회 (작성자만 상태 변경 가능)
    post, err := h.boardService.GetPost(c.Context(), boardID, postID)
    if err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "success": false,
            "message": "게시물을 찾을 수 없습니다",
        })
    }

    // 작성자 또는 관리자만 상태 변경 가능
    if user.ID != post.UserID && user.Role != models.RoleAdmin {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
            "success": false,
            "message": "질문 작성자만 상태를 변경할 수 있습니다",
        })
    }

    // 게시물 해결 상태 업데이트
    post.RawData["is_solved"] = req.Solved
    err = h.boardService.UpdatePost(c.Context(), boardID, post)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "success": false,
            "message": "게시물 상태 업데이트에 실패했습니다: " + err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "success": true,
        "message": req.Solved ? "질문이 해결됨으로 표시되었습니다" : "질문이 미해결로 표시되었습니다",
    })
}
```

이상으로 Q&A 게시판 기능을 완성했습니다. 이렇게 구현된 기능들을 통해 다음 기능이 가능해졌습니다:

1. Q&A 전용 게시판 템플릿 보기
2. 질문과 답변 작성 및 조회
3. 질문의 해결/미해결 상태 관리
4. 답변 채택 및 채택 취소
5. 해결됨/미해결 필터링 기능

이 기능들은 사용자들이 질문과 답변을 효과적으로 관리할 수 있게 해주며, 특히 해결된 질문들을 쉽게 구분할 수 있게 합니다.



---
Powered by [Claude Exporter](https://www.claudexporter.com)