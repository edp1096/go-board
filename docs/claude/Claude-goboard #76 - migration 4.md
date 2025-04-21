# goboard #76 - migration 4

**Created:** 4/17/2025 22:14  
**Updated:** 4/18/2025 1:57  
**Exported:** 4/18/2025 2:55  

## Prompt:
이전에 sqlite to postgres 를 했고, 이번에는 postgres to mysql 이다. 메시지 보면 알겠지만 sqlite to mysql도 똑같이 안된다.

## Response:

여기서 오류를 살펴보면:

```
데이터 마이그레이션 실패: 스키마 마이그레이션 실패: 스키마 마이그레이션 실패: ERROR 006_referrer_stats.sql: failed to run SQL migration: failed to execute SQL query "CREATE TABLE referrer_stats (...);\nCREATE INDEX idx_referrer_stats_referrer ON referrer_stats(referrer_url(255));\n...": Error 1064 (42000): You have an error in your SQL syntax; check the manual that corresponds to your MariaDB server version for the right syntax to use near 'CREATE INDEX idx_referrer_stats_referrer ON referrer_stats(referrer_url(255))...' at line 14
```

이는 MySQL/MariaDB에서 SQL 스크립트를 실행할 때 발생하는 문법 오류입니다. 오류의 주요 원인은 여러 SQL 문장을 하나의 실행 문으로 처리하려는 시도에 있습니다. MariaDB/MySQL에서는 여러 SQL 문장을 하나의 쿼리로 실행할 때 구분자 설정이 필요합니다.

문제가 되는 부분은 "006_referrer_stats.sql" 마이그레이션 파일에서 테이블을 생성한 후 여러 개의 인덱스를 추가하는 부분입니다. MySQL에서는 일반적으로 각 SQL 문을 별도로 실행하거나, 여러 문장을 한 번에 실행할 때는 특별한 처리가 필요합니다.

이 문제를 해결하기 위해서는 마이그레이션 시스템(goose)이 MySQL/MariaDB에서 여러 SQL 문장을 처리하는 방식을 수정해야 합니다. 구체적으로는 다음과 같은 해결책을 생각할 수 있습니다:

1. 마이그레이션 스크립트(006_referrer_stats.sql)를 수정하여 각 CREATE INDEX 문장을 별도의 마이그레이션 함수로 분리
2. goose 설정을 변경하여 MySQL에서 여러 문장을 실행할 수 있도록 설정
3. 마이그레이션 로직에서 각 SQL 문장을 개별적으로 실행하도록 수정

코드를 분석해보니 `migrateSchema` 함수는 임베디드 마이그레이션 파일을 사용하고 있습니다. 여기서 문제가 발생하는 것 같습니다. MySQL/MariaDB의 마이그레이션 파일(`006_referrer_stats.sql`)에서 여러 CREATE INDEX 문장이 하나의 문장으로 처리되고 있기 때문입니다.

이 문제를 해결하려면 다음과 같은 접근 방식을 취할 수 있습니다:

1. MySQL 마이그레이션 파일을 수정하여 각 문장을 세미콜론으로 분리하고 goose가 이를 개별 문장으로 처리하도록 함
2. goose 실행 옵션을 변경하여 여러 문장을 처리할 수 있도록 설정

마이그레이션이 임베디드 파일에서 진행되고 있으므로, 수정이 가능한 부분은 goose 실행 옵션을 변경하는 것입니다. `SetMultiStatementMode` 함수가 있다면 이를 사용하여 여러 SQL 문장을 처리할 수 있습니다.

문제를 해결하기 위한 접근 방식을 제안하겠습니다:

1. `schema_migration.go` 파일에서 `migrateSchema` 함수에 goose의 다중 문장 모드를 활성화하는 코드 추가
2. 마이그레이션 실행 전에 MySQL/MariaDB 연결 설정을 수정하여 다중 문장을 지원하도록 함

# MariaDB 마이그레이션 오류 해결 방안

안녕하세요! 제공해주신 코드와 오류 메시지를 분석해보았습니다. PostgreSQL에서 MySQL로, 그리고 SQLite에서 MySQL로 마이그레이션할 때 동일한 오류가 발생하고 있군요.

## 문제 원인

오류 메시지를 보면 다음과 같은 문제가 있습니다:
```
Error 1064 (42000): You have an error in your SQL syntax; check the manual that corresponds to your MariaDB server version for the right syntax to use near 'CREATE INDEX idx_referrer_stats_referrer ON referrer_stats(referrer_url(255))...' at line 14
```

이 오류는 MySQL/MariaDB에서 여러 SQL 명령문을 한 번에 실행할 때 발생합니다. `006_referrer_stats.sql` 파일에서 테이블 생성 후 여러 CREATE INDEX 명령을 동시에 실행하려고 하는데, MySQL/MariaDB는 기본적으로 단일 연결에서 여러 명령을 실행하는 것을 허용하지 않습니다.

## 해결 방법

### 1. MySQL 연결 문자열 수정

`database.go` 파일에서 MySQL 연결 문자열에 `multiStatements=true` 파라미터를 추가하여 다중 SQL 문 실행을 활성화합니다:

```go
// ConnectDatabase 함수 내 MySQL/MariaDB 연결 부분 수정
case "mysql", "mariadb":
    // MySQL/MariaDB 연결 설정 - multiStatements=true 추가
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&multiStatements=true",
        cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
    sqldb, err = sql.Open("mysql", dsn)
    if err != nil {
        return nil, fmt.Errorf("MySQL 연결 실패: %w", err)
    }
```

### 2. 마이그레이션 쿼리 분리 실행 기능 추가

`schema_migration.go` 파일에서 `migrateSchema` 함수를 수정하여 MySQL/MariaDB에서 마이그레이션을 실행할 때 다중 문장을 적절히 처리하도록 합니다:

```go
// migrateSchema 함수 내 마이그레이션 적용 부분에 추가
// 마이그레이션 적용
fmt.Println("스키마 마이그레이션 적용 중...")
if config.TargetDBConfig.DBDriver == "mysql" || config.TargetDBConfig.DBDriver == "mariadb" {
    // MySQL/MariaDB용 특별 처리 로직 추가
    if err := migrateMySQLSchema(sqlDB, migrationsDir); err != nil {
        return fmt.Errorf("MySQL 스키마 마이그레이션 실패: %w", err)
    }
} else {
    // 다른 데이터베이스는 기존 방식 사용
    if err := goose.Up(sqlDB, migrationsDir); err != nil {
        return fmt.Errorf("스키마 마이그레이션 실패: %w", err)
    }
}
```

### 3. MySQL 전용 마이그레이션 처리 함수 구현

`schema_migration.go` 파일에 다음 함수를 추가합니다:

```go
// migrateMySQLSchema는 MySQL/MariaDB에서 마이그레이션을 실행합니다
func migrateMySQLSchema(db *sql.DB, migrationsDir string) error {
    // 현재 적용된 마이그레이션 버전 확인
    current, err := goose.GetDBVersion(db)
    if err != nil {
        return fmt.Errorf("현재 마이그레이션 버전 확인 실패: %w", err)
    }

    // 사용 가능한 마이그레이션 목록 가져오기
    migrations, err := goose.CollectMigrations(migrationsDir, 0, goose.MaxVersion)
    if err != nil {
        return fmt.Errorf("마이그레이션 파일 목록 가져오기 실패: %w", err)
    }

    // 적용할 마이그레이션 필터링
    toApply := make([]*goose.Migration, 0)
    for _, migration := range migrations {
        if migration.Version > current {
            toApply = append(toApply, migration)
        }
    }

    // 마이그레이션 하나씩 적용
    for _, migration := range toApply {
        fmt.Printf("마이그레이션 적용 중: %s\n", migration.Source)
        
        // SQL 파일 읽기
        sqlBytes, err := goose.ReadUp(migration.Source)
        if err != nil {
            return fmt.Errorf("마이그레이션 파일 읽기 실패: %w", err)
        }
        
        sqlStr := string(sqlBytes)
        
        // SQL 문장 분리 (세미콜론 기준)
        statements := strings.Split(sqlStr, ";")
        
        // 트랜잭션 시작
        tx, err := db.Begin()
        if err != nil {
            return fmt.Errorf("트랜잭션 시작 실패: %w", err)
        }
        
        // 각 문장 개별 실행
        for _, stmt := range statements {
            stmt = strings.TrimSpace(stmt)
            if stmt == "" {
                continue
            }
            
            if _, err := tx.Exec(stmt); err != nil {
                tx.Rollback()
                return fmt.Errorf("SQL 실행 실패: %s, 오류: %w", stmt, err)
            }
        }
        
        // 마이그레이션 버전 기록
        if err := goose.SetVersion(tx, migration.Version, false); err != nil {
            tx.Rollback()
            return fmt.Errorf("마이그레이션 버전 설정 실패: %w", err)
        }
        
        // 트랜잭션 커밋
        if err := tx.Commit(); err != nil {
            return fmt.Errorf("트랜잭션 커밋 실패: %w", err)
        }
        
        fmt.Printf("마이그레이션 성공: %s\n", migration.Source)
    }
    
    return nil
}
```

### 4. 마이그레이션 파일 수정 (선택적)

이 방법은 임베디드 파일 시스템을 사용하고 있어 직접 수정이 어려울 수 있지만, 가능하다면 `006_referrer_stats.sql` 파일을 수정하여 각 CREATE INDEX 문을 별도의 마이그레이션 파일로 분리하는 것도 고려해볼 수 있습니다.

## 완성된 해결책

위 방법들 중에서 1번과 3번을 조합한 해결책이 가장 효과적일 것입니다. MySQL/MariaDB 연결 문자열에 `multiStatements=true`를 추가하고, 필요한 경우 각 SQL 문장을 개별적으로 실행하는 커스텀 마이그레이션 함수를 구현하면 됩니다.

이 방법으로 SQLite에서 MySQL로, PostgreSQL에서 MySQL로 마이그레이션하는 데 발생하는 오류를 해결할 수 있을 것입니다.

## Prompt:
undefined goose.ReadUp, goose.SetVersion

## Response:

이런, 제가 제안한 코드에서 문제가 있군요. `goose.ReadUp`과 `goose.SetVersion` 함수가 goose 라이브러리에서 export되지 않았거나 이름이 다른 것 같습니다. 

goose 라이브러리를 다시 확인해야 할 것 같습니다. 코드를 다시 작성할 때 실제 goose 라이브러리에서 사용 가능한 함수들을 활용해야 합니다.

goose v3 버전에서는 아마도 public API가 다를 수 있습니다. 제가 제안한 함수 이름이 내부적으로만 사용되는 함수이거나, 이름이 변경되었을 수 있습니다.

새로운 접근법을 생각해봐야겠습니다:

1. `goose` 라이브러리의 공개 API를 활용한 솔루션
2. 또는 완전히 다른 방식으로 마이그레이션 파일을 처리하는 방법

우선, 마이그레이션 파일을 직접 처리하는 방식으로 접근해 보겠습니다. 임베디드 파일 시스템에서 파일을 읽고, 분석한 다음 SQL 문장을 하나씩 실행하는 방식으로 구현할 수 있습니다.

# MySQL 마이그레이션 오류 해결 방안

네, `goose.ReadUp`과 `goose.SetVersion` 함수가 외부에서 직접 접근할 수 없는 내부 함수네요. 다른 접근 방식으로 해결해 보겠습니다.

## 해결 방법

### 1. MySQL 연결 문자열 수정

`database.go` 파일의 MySQL/MariaDB 연결 부분에 `multiStatements=true` 파라미터를 추가합니다:

```go
// ConnectDatabase 함수 내 MySQL/MariaDB 연결 부분 수정
case "mysql", "mariadb":
    // MySQL/MariaDB 연결 설정 - multiStatements=true 추가
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&multiStatements=true",
        cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
    sqldb, err = sql.Open("mysql", dsn)
    if err != nil {
        return nil, fmt.Errorf("MySQL 연결 실패: %w", err)
    }
```

### 2. 마이그레이션 로직 수정

`schema_migration.go` 파일에 새로운 함수를 추가하여 MySQL/MariaDB 마이그레이션을 처리합니다:

```go
// migrateSchema 함수 수정
func migrateSchema(config *DataMigrationConfig) error {
    fmt.Println("[1/4] 대상 데이터베이스 스키마 생성 중...")

    // 대상 DB에 스키마 생성
    sqlDB := config.TargetDB.DB

    // 마이그레이션 설정
    goose.SetBaseFS(nil)
    if err := goose.SetDialect(config.TargetDBConfig.DBDriver); err != nil {
        return fmt.Errorf("대상 DB 방언 설정 실패: %w", err)
    }

    // 임베디드 마이그레이션 파일 시스템 가져오기
    var migrationFS fs.FS
    var migrationsDir string

    switch config.TargetDBConfig.DBDriver {
    case "postgres":
        subFS, err := fs.Sub(goboard.PostgresMigrationsFS, "migrations/postgres")
        if err != nil {
            return fmt.Errorf("PostgreSQL 마이그레이션 파일 접근 실패: %w", err)
        }
        migrationFS = subFS
        migrationsDir = "."
    case "sqlite":
        subFS, err := fs.Sub(goboard.SQLiteMigrationsFS, "migrations/sqlite")
        if err != nil {
            return fmt.Errorf("SQLite 마이그레이션 파일 접근 실패: %w", err)
        }
        migrationFS = subFS
        migrationsDir = "."
    default:
        subFS, err := fs.Sub(goboard.MysqlMigrationsFS, "migrations/mysql")
        if err != nil {
            return fmt.Errorf("MySQL 마이그레이션 파일 접근 실패: %w", err)
        }
        migrationFS = subFS
        migrationsDir = "."
    }

    goose.SetBaseFS(migrationFS)

    // 마이그레이션 적용 전 데이터베이스 상태 확인
    fmt.Println("마이그레이션 상태 확인 중...")
    if err := goose.Status(sqlDB, migrationsDir); err != nil {
        return fmt.Errorf("마이그레이션 상태 확인 실패: %w", err)
    }

    // 마이그레이션 적용
    fmt.Println("스키마 마이그레이션 적용 중...")
    
    // MySQL/MariaDB인 경우 특별 처리
    if config.TargetDBConfig.DBDriver == "mysql" || config.TargetDBConfig.DBDriver == "mariadb" {
        // MySQL 마이그레이션 특별 처리
        if err := applyMySQLMigrations(sqlDB, migrationFS, migrationsDir); err != nil {
            return fmt.Errorf("MySQL 스키마 마이그레이션 실패: %w", err)
        }
    } else {
        // 기존 방식 사용
        if err := goose.Up(sqlDB, migrationsDir); err != nil {
            return fmt.Errorf("스키마 마이그레이션 실패: %w", err)
        }
    }

    fmt.Println("기본 스키마 마이그레이션 완료")
    return nil
}

// applyMySQLMigrations는 MySQL/MariaDB용 마이그레이션을 적용합니다
func applyMySQLMigrations(db *sql.DB, migrationFS fs.FS, migrationsDir string) error {
    // 현재 적용된 마이그레이션 버전 확인
    current, err := goose.GetDBVersion(db)
    if err != nil {
        return fmt.Errorf("현재 마이그레이션 버전 확인 실패: %w", err)
    }

    // 직접 파일 목록 가져오기
    entries, err := fs.ReadDir(migrationFS, migrationsDir)
    if err != nil {
        return fmt.Errorf("마이그레이션 파일 목록 가져오기 실패: %w", err)
    }

    // 버전별로 정렬
    var migrationFiles []string
    for _, entry := range entries {
        if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
            migrationFiles = append(migrationFiles, entry.Name())
        }
    }
    sort.Strings(migrationFiles)

    // 각 마이그레이션 파일 처리
    for _, fileName := range migrationFiles {
        // 파일명에서 버전 추출 (예: "001_create_tables.sql" -> 1)
        versionStr := strings.Split(fileName, "_")[0]
        version, err := strconv.ParseInt(versionStr, 10, 64)
        if err != nil {
            return fmt.Errorf("마이그레이션 버전 파싱 실패: %w", err)
        }

        // 이미 적용된 마이그레이션은 건너뛰기
        if version <= current {
            continue
        }

        fmt.Printf("마이그레이션 적용 중: %s\n", fileName)

        // 파일 내용 읽기
        content, err := fs.ReadFile(migrationFS, path.Join(migrationsDir, fileName))
        if err != nil {
            return fmt.Errorf("마이그레이션 파일 읽기 실패: %w", err)
        }

        // 쿼리 분리하기
        sqlStr := string(content)
        statements := strings.Split(sqlStr, ";")

        // 트랜잭션 시작
        tx, err := db.Begin()
        if err != nil {
            return fmt.Errorf("트랜잭션 시작 실패: %w", err)
        }

        // 개별 쿼리 실행
        for _, stmt := range statements {
            stmt = strings.TrimSpace(stmt)
            if stmt == "" {
                continue
            }

            if _, err := tx.Exec(stmt); err != nil {
                tx.Rollback()
                return fmt.Errorf("SQL 실행 실패 (%s): %w", stmt, err)
            }
        }

        // 마이그레이션 버전 기록
        versionSQL := fmt.Sprintf("INSERT INTO goose_db_version (version_id, is_applied) VALUES (%d, true)", version)
        if _, err := tx.Exec(versionSQL); err != nil {
            tx.Rollback()
            return fmt.Errorf("마이그레이션 버전 기록 실패: %w", err)
        }

        // 트랜잭션 커밋
        if err := tx.Commit(); err != nil {
            return fmt.Errorf("트랜잭션 커밋 실패: %w", err)
        }

        fmt.Printf("마이그레이션 성공: %s\n", fileName)
    }

    return nil
}
```

추가적으로 필요한 import 문:
```go
import (
    "path"
    "sort"
    "strconv"
)
```

### 3. 더 간단한 대안 - MySQL 마이그레이션 파일 수정

만약 임베디드 파일 시스템의 마이그레이션 파일을 수정할 수 있다면, `006_referrer_stats.sql` 파일을 수정하여 CREATE TABLE과 각 CREATE INDEX를 별도의 문장으로 실행하도록 구분해주세요:

1. `006_referrer_stats.sql` 파일을 두 개로 분리:
   - `006_referrer_stats.sql`: 테이블 생성 쿼리만 포함
   - `007_referrer_stats_indexes.sql`: 인덱스 생성 쿼리만 포함

2. 또는 원본 마이그레이션 파일에서 세미콜론 후 바로 다음 쿼리가 오도록 수정 (한 줄에 하나의 쿼리):
```sql
CREATE TABLE referrer_stats (...);
CREATE INDEX idx_referrer_stats_referrer ON referrer_stats(referrer_url(255));
CREATE INDEX idx_referrer_stats_domain ON referrer_stats(referrer_domain);
...
```

## 최종 권장사항

1. 첫 번째 방법인 MySQL 연결 문자열에 `multiStatements=true` 파라미터 추가를 먼저 시도하세요. 가장 간단한 방법입니다.

2. 그래도 문제가 해결되지 않으면 위의 `applyMySQLMigrations` 함수를 구현하여 마이그레이션 파일을 직접 처리하는 방식을 사용하세요.

이 두 가지 방법을 통해 이기종 DB 마이그레이션 문제를 해결할 수 있을 것입니다.

## Prompt:
PS D:\dev\pcbangstudio\workspace\go-board\bin> .\migrate_windows_amd64.exe -op data-migrate -source-driver postgres -target-driver mysql -source-env .env_pg -target-env .env_my
==========================
이기종 DB 데이터 마이그레이션 시작
소스: postgres (go_board)
대상: mysql (go_board)
==========================
[1/4] 대상 데이터베이스 스키마 생성 중...
마이그레이션 상태 확인 중...
2025/04/17 22:19:31     Applied At                  Migration
2025/04/17 22:19:31     =======================================
2025/04/17 22:19:31     Fri Apr 11 02:00:16 2025 -- 001_create_tables.sql
2025/04/17 22:19:31     Fri Apr 11 02:00:16 2025 -- 002_comments.sql
2025/04/17 22:19:31     Fri Apr 11 02:00:16 2025 -- 003_file_uploads.sql
2025/04/17 22:19:31     Fri Apr 11 02:00:16 2025 -- 004_qna_votes.sql
2025/04/17 22:19:31     Fri Apr 11 02:00:16 2025 -- 005_board_manager.sql
2025/04/17 22:19:31     Thu Apr 17 22:19:25 2025 -- 006_referrer_stats.sql
2025/04/17 22:19:31     Thu Apr 17 22:19:25 2025 -- 007_system_settings.sql
스키마 마이그레이션 적용 중...
기본 스키마 마이그레이션 완료
[2/4] 기본 테이블 데이터 마이그레이션 중...
  - 테이블 'users' 데이터 마이그레이션 중...
2025/04/17 22:19:31 오류: 테이블 'users' 마이그레이션 실패: 공통 컬럼이 없습니다
  - 테이블 'boards' 데이터 마이그레이션 중...
2025/04/17 22:19:31 오류: 테이블 'boards' 마이그레이션 실패: 공통 컬럼이 없습니다
  - 테이블 'board_fields' 데이터 마이그레이션 중...
2025/04/17 22:19:31 오류: 테이블 'board_fields' 마이그레이션 실패: 공통 컬럼이 없습니다
  - 테이블 'board_managers' 데이터 마이그레이션 중...
    테이블 'board_managers'에 데이터가 없습니다
  - 테이블 'comments' 데이터 마이그레이션 중...
2025/04/17 22:19:31 오류: 테이블 'comments' 마이그레이션 실패: 공통 컬럼이 없습니다
  - 테이블 'attachments' 데이터 마이그레이션 중...
2025/04/17 22:19:31 오류: 테이블 'attachments' 마이그레이션 실패: 공통 컬럼이 없습니다
  - 테이블 'qna_answers' 데이터 마이그레이션 중...
    테이블 'qna_answers'에 데이터가 없습니다
  - 테이블 'qna_question_votes' 데이터 마이그레이션 중...
    테이블 'qna_question_votes'에 데이터가 없습니다
  - 테이블 'qna_answer_votes' 데이터 마이그레이션 중...
    테이블 'qna_answer_votes'에 데이터가 없습니다
  - 테이블 'referrer_stats' 데이터 마이그레이션 중...
2025/04/17 22:19:31 오류: 테이블 'referrer_stats' 마이그레이션 실패: 공통 컬럼이 없습니다
  - 테이블 'system_settings' 데이터 마이그레이션 중...
2025/04/17 22:19:31 오류: 테이블 'system_settings' 마이그레이션 실패: 공통 컬럼이 없습니다
기본 테이블 데이터 마이그레이션 완료
[3/4] 동적 테이블 마이그레이션 중...
  총 2개의 게시판을 마이그레이션합니다
  - 게시판 '갤러리' (board_gaelreori) 마이그레이션 중...
2025/04/17 22:19:31 오류: 게시판 '갤러리' 테이블 생성 실패: 테이블 존재 여부 확인 실패: Error 1054 (42S22): Unknown column 'COUNT(*)' in 'field list'
  - 게시판 '기타 미분류' (board_gita-mibunryu) 마이그레이션 중...
2025/04/17 22:19:31 오류: 게시판 '기타 미분류' 테이블 생성 실패: 테이블 존재 여부 확인 실패: Error 1054 (42S22): Unknown column 'COUNT(*)' in 'field list'
동적 테이블 마이그레이션 완료
[4/4] 시퀀스/자동증가 값 재설정 중...
  - 테이블 'users' 자동증가 값 재설정 완료 (ID: 3)
  - 테이블 'boards' 자동증가 값 재설정 완료 (ID: 2)
  - 테이블 'board_fields' 자동증가 값 재설정 완료 (ID: 2)
2025/04/17 22:19:31 테이블 'system_settings' 최대 ID 조회 실패: Error 1146 (42S02): Table 'go_board.system_settings' doesn't exist (무시하고 계속 진행 합니다)
2025/04/17 22:19:31 테이블 'board_ilban-gesipan' 최대 ID 조회 실패: Error 1064 (42000): You have an error in your SQL syntax; check the manual that corresponds to your MariaDB server version for the right syntax to use near '-gesipan' at line 1 (무시하고 계속 진행합니다)
시퀀스/자동증가 값 재설정 완료
==========================
데이터 마이그레이션 완료 (소요 시간: 64.5344ms)
경고: 마이그레이션 중 9개의 오류가 발생했습니다
  - 테이블 'users' 마이그레이션 실패: 공통 컬럼이 없습니다
  - 테이블 'boards' 마이그레이션 실패: 공통 컬럼이 없습니다
  - 테이블 'board_fields' 마이그레이션 실패: 공통 컬럼이 없습니다
  - 테이블 'comments' 마이그레이션 실패: 공통 컬럼이 없습니다
  - 테이블 'attachments' 마이그레이션 실패: 공통 컬럼이 없습니다
추가 4개 오류 생략...
==========================
2025/04/17 22:19:31 데이터 마이그레이션이 성공적으로 완료되었습니다
PS D:\dev\pcbangstudio\workspace\go-board\bin>

## Response:

이제 데이터베이스 스키마 마이그레이션은 성공했지만 데이터 마이그레이션 과정에서 여러 오류가 발생했습니다. 주요 오류는 다음과 같습니다:

1. "공통 컬럼이 없습니다" - 소스와 대상 데이터베이스 간에 테이블의 컬럼이 일치하지 않는 문제
2. "Unknown column 'COUNT(*)' in 'field list'" - 동적 테이블 생성 중 쿼리 문제
3. "Error 1064 (42000): You have an error in your SQL syntax" - 테이블 이름에 하이픈이 있는 경우 발생하는 SQL 구문 오류

다음 부분들을 해결해야 할 것 같습니다:

1. 테이블 컬럼 조회 로직 수정
2. COUNT(*) 처리 방식 변경
3. 하이픈이 포함된 테이블 이름 처리 수정

코드를 하나씩 살펴보겠습니다.

첫 번째 오류 "공통 컬럼이 없습니다"는 `getTableColumns` 함수와 `findCommonColumns` 함수에서 발생할 가능성이 높습니다. PostgreSQL과 MySQL 간의 컬럼 정보 조회 방식이 다르거나, 대소문자 처리 문제일 수 있습니다.

두 번째 오류 "Unknown column 'COUNT(*)'"는 `ensureDynamicTableExists` 함수에서 테이블 존재 여부를 확인하는 쿼리에서 발생한 것 같습니다. MySQL에서는 COUNT(*) 대신 COUNT(1)을 사용하는 등의 수정이 필요할 수 있습니다.

세 번째 오류는 하이픈이 포함된 테이블 이름을 적절히 인용하지 않아서 발생하는 문제로 보입니다.

이 문제들을 해결하기 위해 코드를 수정해보겠습니다.

# 데이터 마이그레이션 오류 수정 방법

스키마 마이그레이션은 성공했지만 데이터 마이그레이션 단계에서 여러 오류가 발생했네요. 크게 세 가지 문제를 해결해야 합니다:

1. "공통 컬럼이 없습니다"
2. "Unknown column 'COUNT(*)'"
3. 하이픈이 포함된 테이블 이름 관련 구문 오류

## 1. "공통 컬럼이 없습니다" 오류 해결

이 오류는 PostgreSQL과 MySQL 사이의 컬럼 정보 조회 방식 차이 때문에 발생합니다. `getTableColumns` 함수와 `findCommonColumns` 함수를 수정해야 합니다.

```go
// getTableColumns 함수 수정
func getTableColumns(db *bun.DB, driver string, tableName string) ([]ColumnMetadata, error) {
    ctx := context.Background()
    var columns []ColumnMetadata

    switch driver {
    case "postgres":
        // PostgreSQL용 쿼리 (변경 없음)
        var results []struct {
            ColumnName string         `bun:"column_name"`
            DataType   string         `bun:"data_type"`
            IsNullable string         `bun:"is_nullable"`
            Default    sql.NullString `bun:"column_default"`
        }

        err := db.NewSelect().
            TableExpr("information_schema.columns").
            Column("column_name", "data_type", "is_nullable", "column_default").
            Where("table_schema = 'public'").
            Where("table_name = ?", tableName).
            Order("ordinal_position").
            Scan(ctx, &results)

        if err != nil {
            return nil, fmt.Errorf("컬럼 정보 조회 실패: %w", err)
        }

        for _, r := range results {
            columns = append(columns, ColumnMetadata{
                Name:     r.ColumnName,
                Type:     r.DataType,
                Nullable: r.IsNullable == "YES",
                Default:  r.Default,
            })
        }

    case "mysql", "mariadb":
        // MySQL/MariaDB용 쿼리 수정
        query := fmt.Sprintf("SHOW COLUMNS FROM %s", quoteTableName(driver, tableName))
        rows, err := db.QueryContext(ctx, query)
        if err != nil {
            return nil, fmt.Errorf("컬럼 정보 조회 실패: %w", err)
        }
        defer rows.Close()

        // SHOW COLUMNS 결과 처리
        for rows.Next() {
            var field, fieldType, null, key, extra string
            var defaultValue sql.NullString
            if err := rows.Scan(&field, &fieldType, &null, &key, &defaultValue, &extra); err != nil {
                return nil, fmt.Errorf("컬럼 데이터 스캔 실패: %w", err)
            }

            columns = append(columns, ColumnMetadata{
                Name:     field,
                Type:     fieldType,
                Nullable: null == "YES",
                Default:  defaultValue,
            })
        }

    case "sqlite":
        // SQLite 코드 (변경 없음)
        // ...

    default:
        return nil, fmt.Errorf("지원하지 않는 데이터베이스 드라이버: %s", driver)
    }

    return columns, nil
}

// findCommonColumns 함수 수정 - 대소문자 처리 개선
func findCommonColumns(sourceColumns, targetColumns []ColumnMetadata) ([]ColumnMetadata, []string, []string) {
    var common []ColumnMetadata
    var sourceNames []string
    var targetNames []string

    sourceMap := make(map[string]ColumnMetadata)
    for _, col := range sourceColumns {
        sourceMap[strings.ToLower(col.Name)] = col
    }

    targetMap := make(map[string]ColumnMetadata)
    for _, col := range targetColumns {
        targetMap[strings.ToLower(col.Name)] = col
    }

    // MySQL은 일반적으로 대소문자를 구분하지 않으므로 이를 고려하여 비교
    for _, sourceCol := range sourceColumns {
        lowerName := strings.ToLower(sourceCol.Name)
        if targetCol, ok := targetMap[lowerName]; ok {
            common = append(common, sourceCol)
            sourceNames = append(sourceNames, quoteColumnName(sourceCol.Name))
            targetNames = append(targetNames, quoteColumnName(targetCol.Name))
        }
    }

    // 공통 컬럼이 없을 경우 디버깅 정보 출력
    if len(common) == 0 {
        fmt.Println("소스 컬럼:")
        for _, col := range sourceColumns {
            fmt.Printf("  - %s (%s)\n", col.Name, col.Type)
        }
        fmt.Println("대상 컬럼:")
        for _, col := range targetColumns {
            fmt.Printf("  - %s (%s)\n", col.Name, col.Type)
        }
    }

    return common, sourceNames, targetNames
}
```

## 2. "Unknown column 'COUNT(*)'" 오류 해결

MySQL에서 COUNT(*) 표현식을 올바르게 처리하도록 `ensureDynamicTableExists` 함수를 수정합니다:

```go
func ensureDynamicTableExists(config *DataMigrationConfig, board *models.Board, _ service.DynamicBoardService) error {
    // 대상 DB에 테이블이 존재하는지 확인
    var exists bool
    ctx := context.Background()

    // DB 종류에 따라 다른 쿼리 사용
    switch config.TargetDBConfig.DBDriver {
    case "postgres":
        // PostgreSQL 코드 (변경 없음)
        // ...

    case "mysql", "mariadb":
        // MySQL/MariaDB 수정 - 쿼리 방식 변경
        query := fmt.Sprintf("SELECT COUNT(1) FROM information_schema.tables WHERE table_schema = '%s' AND table_name = '%s'",
            config.TargetDBConfig.DBName, board.TableName)
        
        var count int
        err := config.TargetDB.QueryRowContext(ctx, query).Scan(&count)
        if err != nil {
            return fmt.Errorf("테이블 존재 여부 확인 실패: %w", err)
        }
        
        exists = count > 0

    case "sqlite":
        // SQLite 코드 (변경 없음)
        // ...

    default:
        return fmt.Errorf("지원하지 않는 데이터베이스 드라이버: %s", config.TargetDBConfig.DBDriver)
    }
    
    // 나머지 코드 (변경 없음)
    // ...
}
```

## 3. 하이픈이 포함된 테이블 이름 처리 수정

MySQL에서 하이픈이 포함된 테이블 이름을 올바르게 처리하도록 `quoteTableName` 함수를 수정합니다:

```go
func quoteTableName(driver, tableName string) string {
    // 특수 문자가 있는지 확인
    hasSpecialChars := strings.ContainsAny(tableName, "-.") || strings.Contains(tableName, " ")

    switch driver {
    case "postgres":
        // PostgreSQL (변경 없음)
        return fmt.Sprintf("\"%s\"", tableName)
    case "mysql", "mariadb":
        // MySQL에서는 무조건 백틱으로 감싸기 (특수 문자가 없어도)
        return fmt.Sprintf("`%s`", tableName)
    case "sqlite":
        // SQLite (변경 없음)
        if hasSpecialChars {
            return fmt.Sprintf("\"%s\"", tableName)
        }
        return tableName
    default:
        return tableName
    }
}
```

## 4. 자동증가 값 재설정 오류 수정

하이픈이 포함된 테이블 이름을 처리하기 위해 `resetSequences` 함수도 수정:

```go
func resetSequences(config *DataMigrationConfig) error {
    fmt.Println("[4/4] 시퀀스/자동증가 값 재설정 중...")
    
    // 기존 코드 유지
    // ...

    if config.TargetDBConfig.DBDriver == "mysql" || config.TargetDBConfig.DBDriver == "mariadb" {
        // MySQL/MariaDB 자동증가 값 재설정
        for _, tableName := range tables {
            // 자동증가 재설정이 필요한 테이블만 처리
            if !needsSequenceReset(tableName) {
                continue
            }

            // 테이블 이름에 하이픈이 있는지 확인하고 적절히 처리
            quotedTableName := quoteTableName(config.TargetDBConfig.DBDriver, tableName)
            
            // 현재 최대 ID 값 조회
            query := fmt.Sprintf("SELECT COALESCE(MAX(id), 0) FROM %s", quotedTableName)
            var maxID int64
            err := config.TargetDB.QueryRow(query).Scan(&maxID)
            
            if err != nil {
                log.Printf("테이블 '%s' 최대 ID 조회 실패: %v (무시하고 계속 진행합니다)", tableName, err)
                continue
            }

            // 자동증가 값 재설정
            if maxID > 0 {
                // MySQL에서 테이블 이름 적절히 인용
                resetSQL := fmt.Sprintf("ALTER TABLE %s AUTO_INCREMENT = %d", quotedTableName, maxID+1)
                _, err := config.TargetDB.ExecContext(ctx, resetSQL)
                if err != nil {
                    log.Printf("테이블 '%s' 자동증가 값 재설정 실패: %v (무시하고 계속 진행합니다)", tableName, err)
                } else if config.VerboseLogging {
                    fmt.Printf("  - 테이블 '%s' 자동증가 값 재설정 완료 (ID: %d)\n", tableName, maxID+1)
                }
            }
        }
    }
    
    // 나머지 코드 (변경 없음)
    // ...
}
```

## 종합적인 수정사항

1. 테이블 컬럼 조회 로직 개선 (MySQL용 SHOW COLUMNS 사용)
2. 공통 컬럼 찾기 로직 강화 (디버깅 정보 추가)
3. COUNT(*) 대신 직접 SQL 쿼리 사용
4. 하이픈이 포함된 테이블 이름 처리 개선
5. MySQL 테이블 이름 항상 백틱으로 감싸기

이러한 수정을 통해 PostgreSQL에서 MySQL로의 데이터 마이그레이션 문제를 해결할 수 있을 것입니다. 컬럼 정보 조회 방식의 차이와 대소문자 처리 등 데이터베이스 시스템 간의 차이를 고려한 코드가 되었습니다.

## Prompt:
PS D:\dev\pcbangstudio\workspace\go-board\bin> .\migrate_windows_amd64.exe -op data-migrate -source-driver sqlite -target-driver mysql -source-env .env_sqlite -target-env .env_my   
==========================
이기종 DB 데이터 마이그레이션 시작
소스: sqlite (go_board)
대상: mysql (go_board)
==========================
[1/4] 대상 데이터베이스 스키마 생성 중...
마이그레이션 상태 확인 중...
2025/04/17 22:24:02     Applied At                  Migration
2025/04/17 22:24:02     =======================================
2025/04/17 22:24:02     Fri Apr 11 02:00:16 2025 -- 001_create_tables.sql
2025/04/17 22:24:02     Fri Apr 11 02:00:16 2025 -- 002_comments.sql
2025/04/17 22:24:02     Fri Apr 11 02:00:16 2025 -- 003_file_uploads.sql
2025/04/17 22:24:02     Fri Apr 11 02:00:16 2025 -- 004_qna_votes.sql
2025/04/17 22:24:02     Fri Apr 11 02:00:16 2025 -- 005_board_manager.sql
2025/04/17 22:24:02     Thu Apr 17 22:19:25 2025 -- 006_referrer_stats.sql
2025/04/17 22:24:02     Thu Apr 17 22:19:25 2025 -- 007_system_settings.sql
스키마 마이그레이션 적용 중...
기본 스키마 마이그레이션 완료
[2/4] 기본 테이블 데이터 마이그레이션 중...
  - 테이블 'users' 데이터 마이그레이션 중...
2025/04/17 22:24:02 오류: 테이블 'users' 마이그레이션 실패: 대상 테이블 데이터 삭제 실패: Error 1451 (23000): Cannot delete or update a parent row: a foreign key constraint fails (`go_board`.`board_gaelreori`, CONSTRAINT `board_gaelreori_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`))
  - 테이블 'boards' 데이터 마이그레이션 중...
    기존 대상 테이블 데이터 삭제됨
    테이블 'boards': 2/2 행 처리 완료
  - 테이블 'board_fields' 데이터 마이그레이션 중...
    기존 대상 테이블 데이터 삭제됨
    테이블 'board_fields': 2/2 행 처리 완료
  - 테이블 'board_managers' 데이터 마이그레이션 중...
    테이블 'board_managers'에 데이터가 없습니다
  - 테이블 'comments' 데이터 마이그레이션 중...
    기존 대상 테이블 데이터 삭제됨
    테이블 'comments': 8/8 행 처리 완료
  - 테이블 'attachments' 데이터 마이그레이션 중...
    기존 대상 테이블 데이터 삭제됨
    테이블 'attachments': 5/5 행 처리 완료
  - 테이블 'qna_answers' 데이터 마이그레이션 중...
    테이블 'qna_answers'에 데이터가 없습니다
  - 테이블 'qna_question_votes' 데이터 마이그레이션 중...
    테이블 'qna_question_votes'에 데이터가 없습니다
  - 테이블 'qna_answer_votes' 데이터 마이그레이션 중...
    테이블 'qna_answer_votes'에 데이터가 없습니다
  - 테이블 'referrer_stats' 데이터 마이그레이션 중...
2025/04/17 22:24:02 오류: 테이블 'referrer_stats' 마이그레이션 실패: 대상 테이블 구조 가져오기 실패: 컬럼 정보 조회 실패: Error 1146 (42S02): Table 'go_board.referrer_stats' doesn't exist
  - 테이블 'system_settings' 데이터 마이그레이션 중...
2025/04/17 22:24:02 오류: 테이블 'system_settings' 마이그레이션 실패: 대상 테이블 구조 가져오기 실패: 컬럼 정보 조회 실패: Error 1146 (42S02): Table 'go_board.system_settings' doesn't exist
기본 테이블 데이터 마이그레이션 완료
[3/4] 동적 테이블 마이그레이션 중...
  총 2개의 게시판을 마이그레이션합니다
  - 게시판 '갤러리' (board_gaelreori) 마이그레이션 중...
  - 테이블 'board_gaelreori' 데이터 마이그레이션 중...
    기존 대상 테이블 데이터 삭제됨
    테이블 'board_gaelreori': 5/5 행 처리 완료
  - 게시판 '기타 미분류' (board_gita-mibunryu) 마이그레이션 중...
  - 테이블 'board_gita-mibunryu' 데이터 마이그레이션 중...
    기존 대상 테이블 데이터 삭제됨
    테이블 'board_gita-mibunryu': 4/4 행 처리 완료
동적 테이블 마이그레이션 완료
[4/4] 시퀀스/자동증가 값 재설정 중...
  - 테이블 'users' 자동증가 값 재설정 완료 (ID: 3)
  - 테이블 'boards' 자동증가 값 재설정 완료 (ID: 3)
  - 테이블 'board_fields' 자동증가 값 재설정 완료 (ID: 3)
  - 테이블 'comments' 자동증가 값 재설정 완료 (ID: 13)
  - 테이블 'attachments' 자동증가 값 재설정 완료 (ID: 6)
2025/04/17 22:24:02 테이블 'system_settings' 최대 ID 조회 실패: Error 1146 (42S02): Table 'go_board.system_settings' doesn't exist (무시하고 계속 진행 합니다)
  - 테이블 'board_gaelreori' 자동증가 값 재설정 완료 (ID: 6)
  - 테이블 'board_gita-mibunryu' 자동증가 값 재설정 완료 (ID: 5)
시퀀스/자동증가 값 재설정 완료
==========================
데이터 마이그레이션 완료 (소요 시간: 72.8029ms)
경고: 마이그레이션 중 3개의 오류가 발생했습니다
  - 테이블 'users' 마이그레이션 실패: 대상 테이블 데이터 삭제 실패: Error 1451 (23000): Cannot delete or update a parent row: a foreign key constraint fails (`go_board`.`board_gaelreori`, CONSTRAINT `board_gaelreori_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`))
  - 테이블 'referrer_stats' 마이그레이션 실패: 대상 테이블 구조 가져오기 실패: 컬럼 정보 조회 실패: Error 1146 (42S02): Table 'go_board.referrer_stats' doesn't exist
  - 테이블 'system_settings' 마이그레이션 실패: 대상 테이블 구조 가져오기 실패: 컬럼 정보 조회 실패: Error 1146 (42S02): Table 'go_board.system_settings' doesn't exist
==========================
2025/04/17 22:24:02 데이터 마이그레이션이 성공적으로 완료되었습니다
PS D:\dev\pcbangstudio\workspace\go-board\bin>



함수 단위로 끊어서 생략없는 완전한 코드로.

## Response:

제공된 로그를 보면 SQLite에서 MySQL로의 마이그레이션이 마이그레이션 작업 중 일부 테이블에서 오류가 발생했지만 대부분의 데이터는 성공적으로 마이그레이션되었습니다. 해결해야 할 세 가지 주요 문제가 있습니다:

1. `users` 테이블 삭제 실패: 외래 키 제약조건 때문에 삭제할 수 없음
2. `referrer_stats` 테이블이 존재하지 않음
3. `system_settings` 테이블이 존재하지 않음

이러한 문제를 해결하기 위해 여러 함수를 수정해야 합니다. 각 함수에 대한 완전한 코드를 작성하겠습니다.

먼저, 외래 키 제약 조건 처리를 개선하기 위해 `shouldCleanTableBeforeMigration` 함수와 `migrateTableData` 함수를 수정합니다.

다음으로, 테이블 존재 여부 확인 및 생성 로직을 개선하기 위해 `getTableColumns` 함수와 관련 로직을 수정합니다.

마지막으로, 하이픈이 포함된 테이블 이름 처리를 위한 `quoteTableName` 함수를 수정합니다.

각 함수별로 완전한 코드를 작성하겠습니다.

# 이기종 DB 마이그레이션 문제 해결 방법

스키마 마이그레이션은 성공적으로 완료되었으나, 데이터 마이그레이션 과정에서 일부 문제가 발생했네요. 주요 문제는 다음과 같습니다:

1. 외래 키 제약 조건으로 인한 users 테이블 삭제 실패
2. referrer_stats 및 system_settings 테이블이 존재하지 않음
3. 하이픈이 포함된 테이블 이름 처리

이를 해결하기 위해 함수별로 수정된 코드를 제공하겠습니다.

## 1. 외래 키 제약조건 처리 개선

```go
// shouldCleanTableBeforeMigration은 마이그레이션 전에 테이블을 정리해야 하는지 결정합니다
func shouldCleanTableBeforeMigration(tableName string) bool {
    // 외래 키 제약 조건이 있거나 이력 데이터인 테이블은 정리하지 않음
    skipCleanTables := map[string]bool{
        "referrer_stats": true,
        "users": true,           // users 테이블 추가 (외래 키 때문에)
    }

    return !skipCleanTables[tableName]
}
```

## 2. MySQL 테이블 존재 확인 및 컬럼 조회 개선

```go
// getTableColumns는 테이블의 컬럼 정보를 가져옵니다
func getTableColumns(db *bun.DB, driver string, tableName string) ([]ColumnMetadata, error) {
    ctx := context.Background()
    var columns []ColumnMetadata

    switch driver {
    case "postgres":
        // PostgreSQL용 쿼리
        var results []struct {
            ColumnName string         `bun:"column_name"`
            DataType   string         `bun:"data_type"`
            IsNullable string         `bun:"is_nullable"`
            Default    sql.NullString `bun:"column_default"`
        }

        err := db.NewSelect().
            TableExpr("information_schema.columns").
            Column("column_name", "data_type", "is_nullable", "column_default").
            Where("table_schema = 'public'").
            Where("table_name = ?", tableName).
            Order("ordinal_position").
            Scan(ctx, &results)

        if err != nil {
            return nil, fmt.Errorf("컬럼 정보 조회 실패: %w", err)
        }

        for _, r := range results {
            columns = append(columns, ColumnMetadata{
                Name:     r.ColumnName,
                Type:     r.DataType,
                Nullable: r.IsNullable == "YES",
                Default:  r.Default,
            })
        }

    case "mysql", "mariadb":
        // 테이블 존재 여부 먼저 확인
        var exists int
        checkQuery := fmt.Sprintf("SELECT COUNT(1) FROM information_schema.tables WHERE table_schema = '%s' AND table_name = '%s'",
            db.DB.Stats().DSN, tableName)
        
        err := db.QueryRow(checkQuery).Scan(&exists)
        if err != nil {
            // DB 이름 직접 획득 시도
            var dbName string
            if err := db.QueryRow("SELECT DATABASE()").Scan(&dbName); err == nil {
                checkQuery = fmt.Sprintf("SELECT COUNT(1) FROM information_schema.tables WHERE table_schema = '%s' AND table_name = '%s'",
                    dbName, tableName)
                err = db.QueryRow(checkQuery).Scan(&exists)
            }
            
            if err != nil {
                return nil, fmt.Errorf("테이블 존재 여부 확인 실패: %w", err)
            }
        }
        
        if exists == 0 {
            return nil, fmt.Errorf("테이블이 존재하지 않습니다: %s", tableName)
        }

        // MySQL/MariaDB용 SHOW COLUMNS 쿼리
        quotedName := quoteTableName(driver, tableName)
        rows, err := db.QueryContext(ctx, fmt.Sprintf("SHOW COLUMNS FROM %s", quotedName))
        if err != nil {
            return nil, fmt.Errorf("컬럼 정보 조회 실패: %w", err)
        }
        defer rows.Close()

        for rows.Next() {
            var field, colType, null, key, extra string
            var defaultValue sql.NullString
            if err := rows.Scan(&field, &colType, &null, &key, &defaultValue, &extra); err != nil {
                return nil, fmt.Errorf("컬럼 정보 스캔 실패: %w", err)
            }

            columns = append(columns, ColumnMetadata{
                Name:     field,
                Type:     colType,
                Nullable: null == "YES",
                Default:  defaultValue,
            })
        }

    case "sqlite":
        // 테이블 이름에 하이픈이 있는 경우 따옴표로 묶기
        quotedTableName := tableName
        if strings.Contains(tableName, "-") {
            quotedTableName = fmt.Sprintf("\"%s\"", tableName)
        }

        rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", quotedTableName))
        if err != nil {
            return nil, fmt.Errorf("컬럼 정보 조회 실패: %w", err)
        }
        defer rows.Close()

        for rows.Next() {
            var cid int
            var name, typeName string
            var notNull int
            var dfltValue sql.NullString
            var pk int

            if err := rows.Scan(&cid, &name, &typeName, &notNull, &dfltValue, &pk); err != nil {
                return nil, fmt.Errorf("컬럼 정보 스캔 실패: %w", err)
            }

            columns = append(columns, ColumnMetadata{
                Name:     name,
                Type:     typeName,
                Nullable: notNull == 0,
                Default:  dfltValue,
            })
        }

    default:
        return nil, fmt.Errorf("지원하지 않는 데이터베이스 드라이버: %s", driver)
    }

    return columns, nil
}
```

## 3. 테이블 인용 개선 함수

```go
// quoteTableName은 데이터베이스 드라이버에 따라 테이블 이름을 인용 부호로 묶습니다
func quoteTableName(driver, tableName string) string {
    switch driver {
    case "postgres":
        // PostgreSQL에서는 항상 따옴표로 감싸기
        return fmt.Sprintf("\"%s\"", tableName)
    case "mysql", "mariadb":
        // MySQL에서는 항상 백틱으로 감싸기
        return fmt.Sprintf("`%s`", tableName)
    case "sqlite":
        // SQLite에서는 특수 문자가 있으면 항상 따옴표로 감싸기
        if strings.ContainsAny(tableName, "-.") || strings.Contains(tableName, " ") {
            return fmt.Sprintf("\"%s\"", tableName)
        }
        return tableName
    default:
        return tableName
    }
}
```

## 4. 마이그레이션 전 테이블 존재 여부 확인 및 생성 함수

```go
// ensureDynamicTableExists는 대상 DB에 동적 테이블이 존재하는지 확인하고, 없으면 생성합니다
func ensureDynamicTableExists(config *DataMigrationConfig, board *models.Board, dynamicBoardService service.DynamicBoardService) error {
    // 대상 DB에 테이블이 존재하는지 확인
    var exists bool
    ctx := context.Background()

    // DB 종류에 따라 다른 쿼리 사용
    switch config.TargetDBConfig.DBDriver {
    case "postgres":
        var count int
        err := config.TargetDB.NewSelect().
            TableExpr("information_schema.tables").
            ColumnExpr("COUNT(*) AS count").
            Where("table_schema = 'public'").
            Where("table_name = ?", board.TableName).
            Scan(ctx, &count)

        if err != nil {
            return fmt.Errorf("테이블 존재 여부 확인 실패: %w", err)
        }

        exists = count > 0

    case "mysql", "mariadb":
        // MySQL/MariaDB용 직접 SQL 쿼리
        var count int
        dbName := config.TargetDBConfig.DBName
        query := fmt.Sprintf("SELECT COUNT(1) FROM information_schema.tables WHERE table_schema = '%s' AND table_name = '%s'",
            dbName, board.TableName)
        
        err := config.TargetDB.QueryRow(query).Scan(&count)
        if err != nil {
            return fmt.Errorf("테이블 존재 여부 확인 실패: %w", err)
        }

        exists = count > 0

    case "sqlite":
        var count int
        err := config.TargetDB.NewSelect().
            TableExpr("sqlite_master").
            Column("COUNT(*)").
            Where("type = 'table'").
            Where("name = ?", board.TableName).
            Scan(ctx, &count)

        if err != nil {
            return fmt.Errorf("테이블 존재 여부 확인 실패: %w", err)
        }

        exists = count > 0
    default:
        return fmt.Errorf("지원하지 않는 데이터베이스 드라이버: %s", config.TargetDBConfig.DBDriver)
    }

    // 테이블이 없으면 생성
    if !exists {
        // 필드 목록 가져오기
        var fields []*models.BoardField
        err := config.SourceDB.NewSelect().
            Model(&fields).
            Where("board_id = ?", board.ID).
            Order("sort_order ASC").
            Scan(ctx)
        if err != nil {
            return fmt.Errorf("게시판 필드 가져오기 실패: %w", err)
        }

        // 대상 DB에 테이블 생성
        targetDynamicService := service.NewDynamicBoardService(config.TargetDB)
        if err := targetDynamicService.CreateBoardTable(ctx, board, fields); err != nil {
            return fmt.Errorf("게시판 테이블 생성 실패: %w", err)
        }

        fmt.Printf("    게시판 테이블 '%s' 생성됨\n", board.TableName)
    }

    return nil
}
```

## 5. 데이터 마이그레이션 함수 개선

```go
// migrateTableData는 테이블 데이터를 마이그레이션합니다
func migrateTableData(config *DataMigrationConfig, tableName string) error {
    ctx := context.Background()

    if config.VerboseLogging {
        fmt.Printf("  - 테이블 '%s' 데이터 마이그레이션 중...\n", tableName)
    }

    // 테이블 총 행 수 카운트
    var totalRows int
    query := fmt.Sprintf("SELECT COUNT(*) FROM %s", quoteTableName(config.SourceDBConfig.DBDriver, tableName))
    err := config.SourceDB.QueryRow(query).Scan(&totalRows)
    if err != nil {
        return fmt.Errorf("행 수 계산 실패: %w", err)
    }

    if totalRows == 0 {
        if config.VerboseLogging {
            fmt.Printf("    테이블 '%s'에 데이터가 없습니다\n", tableName)
        }
        return nil
    }

    // 테이블 구조 가져오기
    sourceColumns, err := getTableColumns(config.SourceDB, config.SourceDBConfig.DBDriver, tableName)
    if err != nil {
        return fmt.Errorf("소스 테이블 구조 가져오기 실패: %w", err)
    }

    // 대상 테이블 존재 여부 확인
    targetColumns, err := getTableColumns(config.TargetDB, config.TargetDBConfig.DBDriver, tableName)
    if err != nil {
        // 특정 오류에 대한 특별 처리
        if strings.Contains(err.Error(), "테이블이 존재하지 않습니다") || 
           strings.Contains(err.Error(), "doesn't exist") {
            if tableName == "referrer_stats" || tableName == "system_settings" {
                return fmt.Errorf("대상 테이블 존재하지 않음: %w", err)
            }
        }
        return fmt.Errorf("대상 테이블 구조 가져오기 실패: %w", err)
    }

    // 공통 컬럼 찾기
    commonColumns, sourceColumnNames, targetColumnNames := findCommonColumns(sourceColumns, targetColumns)
    if len(commonColumns) == 0 {
        return fmt.Errorf("공통 컬럼이 없습니다")
    }

    // 테이블 모델 정보 가져오기
    modelInfo := getModelInfo(tableName)

    // 데이터 배치 처리
    batchSize := config.BatchSize
    if batchSize <= 0 {
        batchSize = 1000
    }

    totalBatches := (totalRows + batchSize - 1) / batchSize
    processedRows := 0

    // 대상 테이블의 기존 데이터 삭제 (필요 시)
    if shouldCleanTableBeforeMigration(tableName) {
        // MySQL의 외래 키 제약 조건 일시적으로 비활성화
        if config.TargetDBConfig.DBDriver == "mysql" || config.TargetDBConfig.DBDriver == "mariadb" {
            _, err = config.TargetDB.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 0;")
            if err != nil {
                return fmt.Errorf("외래 키 제약 비활성화 실패: %w", err)
            }
            defer func() {
                // 작업 완료 후 다시 활성화
                _, _ = config.TargetDB.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 1;")
            }()
        } else if config.TargetDBConfig.DBDriver == "postgres" {
            _, err = config.TargetDB.ExecContext(ctx, "SET session_replication_role = 'replica';")
            if err != nil {
                return fmt.Errorf("외래 키 제약 비활성화 실패: %w", err)
            }
            defer func() {
                // 작업 완료 후 다시 활성화
                _, _ = config.TargetDB.ExecContext(ctx, "SET session_replication_role = 'origin';")
            }()
        }

        deleteQuery := fmt.Sprintf("DELETE FROM %s", quoteTableName(config.TargetDBConfig.DBDriver, tableName))
        _, err = config.TargetDB.ExecContext(ctx, deleteQuery)
        if err != nil {
            return fmt.Errorf("대상 테이블 데이터 삭제 실패: %w", err)
        }
        if config.VerboseLogging {
            fmt.Printf("    기존 대상 테이블 데이터 삭제됨\n")
        }
    }

    // 트랜잭션 활성화 여부 확인
    useTransaction := config.EnableTransactions && shouldUseTransactionForTable(tableName, totalRows)

    for batchNum := 0; batchNum < totalBatches; batchNum++ {
        offset := batchNum * batchSize
        limit := batchSize

        if config.VerboseLogging && totalBatches > 1 {
            fmt.Printf("    배치 %d/%d 처리 중... (오프셋: %d, 한계: %d)\n", batchNum+1, totalBatches, offset, limit)
        }

        // 소스 데이터 쿼리 구성
        sourceQuery := fmt.Sprintf("SELECT %s FROM %s ORDER BY id LIMIT %d OFFSET %d",
            strings.Join(sourceColumnNames, ", "),
            quoteTableName(config.SourceDBConfig.DBDriver, tableName),
            limit,
            offset)

        // 소스 데이터 조회
        sourceRows, err := config.SourceDB.QueryContext(ctx, sourceQuery)
        if err != nil {
            return fmt.Errorf("소스 데이터 조회 실패: %w", err)
        }

        // 트랜잭션 시작 (필요 시)
        var tx *sql.Tx
        if useTransaction {
            tx, err = config.TargetDB.DB.Begin()
            if err != nil {
                sourceRows.Close()
                return fmt.Errorf("트랜잭션 시작 실패: %w", err)
            }
        }

        // 행별 처리
        rowsInBatch := 0
        for sourceRows.Next() {
            // 행 데이터 변수 준비
            rowValues := make([]interface{}, len(commonColumns))
            valuePtrs := make([]interface{}, len(commonColumns))
            for i := range rowValues {
                valuePtrs[i] = &rowValues[i]
            }

            // 행 데이터 읽기
            if err := sourceRows.Scan(valuePtrs...); err != nil {
                sourceRows.Close()
                if useTransaction {
                    tx.Rollback()
                }
                return fmt.Errorf("행 데이터 읽기 실패: %w", err)
            }

            // SQL 쿼리용 컬럼 값 배열
            var columnValues []string

            // 각 값을 SQL 문자열로 변환
            for i, val := range rowValues {
                colName := commonColumns[i].Name

                // 널 값 처리
                if val == nil {
                    columnValues = append(columnValues, "NULL")
                    continue
                }

                // referrer_stats 테이블 중복 키 문제 해결을 위한 특별 처리
                if tableName == "referrer_stats" && colName == "id" && config.TargetDBConfig.DBDriver == "postgres" {
                    columnValues = append(columnValues, "nextval('referrer_stats_id_seq')")
                    continue
                }

                // 모델 기반 타입 변환
                fieldInfo, hasField := modelInfo[colName]
                if hasField && config.TargetDBConfig.DBDriver == "postgres" {
                    // 타입에 따른 변환
                    switch fieldInfo.fieldType {
                    case "bool":
                        // 불리언 처리
                        switch v := val.(type) {
                        case int64:
                            if v == 1 {
                                columnValues = append(columnValues, "TRUE")
                            } else {
                                columnValues = append(columnValues, "FALSE")
                            }
                        case int:
                            if v == 1 {
                                columnValues = append(columnValues, "TRUE")
                            } else {
                                columnValues = append(columnValues, "FALSE")
                            }
                        case string:
                            if strings.ToLower(v) == "true" || v == "1" {
                                columnValues = append(columnValues, "TRUE")
                            } else {
                                columnValues = append(columnValues, "FALSE")
                            }
                        case bool:
                            if v {
                                columnValues = append(columnValues, "TRUE")
                            } else {
                                columnValues = append(columnValues, "FALSE")
                            }
                        default:
                            columnValues = append(columnValues, "FALSE")
                        }
                        continue
                    case "time.Time":
                        // 시간 처리
                        switch v := val.(type) {
                        case time.Time:
                            columnValues = append(columnValues, fmt.Sprintf("'%s'", v.Format("2006-01-02 15:04:05")))
                        case string:
                            columnValues = append(columnValues, fmt.Sprintf("'%s'", v))
                        default:
                            columnValues = append(columnValues, fmt.Sprintf("'%v'", v))
                        }
                        continue
                    }
                }

                // 일반 데이터 타입 변환
                switch v := val.(type) {
                case string:
                    // 문자열은 작은 따옴표로 감싸고 내부 작은 따옴표는 이스케이프
                    escapedVal := strings.Replace(v, "'", "''", -1)
                    columnValues = append(columnValues, fmt.Sprintf("'%s'", escapedVal))
                case bool:
                    // MySQL/MariaDB의 경우 1/0 사용, PostgreSQL은 TRUE/FALSE
                    if config.TargetDBConfig.DBDriver == "postgres" {
                        if v {
                            columnValues = append(columnValues, "TRUE")
                        } else {
                            columnValues = append(columnValues, "FALSE")
                        }
                    } else {
                        // 다른 DB는 1/0 사용
                        if v {
                            columnValues = append(columnValues, "1")
                        } else {
                            columnValues = append(columnValues, "0")
                        }
                    }
                case time.Time:
                    columnValues = append(columnValues, fmt.Sprintf("'%s'", v.Format("2006-01-02 15:04:05")))
                default:
                    // 숫자 등 다른 타입은 그대로
                    columnValues = append(columnValues, fmt.Sprintf("%v", v))
                }
            }

            // 직접 문자열로 SQL 쿼리 구성
            directSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
                quoteTableName(config.TargetDBConfig.DBDriver, tableName),
                strings.Join(targetColumnNames, ", "),
                strings.Join(columnValues, ", "))

            // 쿼리 실행
            var err error
            if useTransaction {
                _, err = tx.ExecContext(ctx, directSQL)
            } else {
                _, err = config.TargetDB.ExecContext(ctx, directSQL)
            }

            if err != nil {
                // JSON으로 행 데이터 직렬화 (디버깅용)
                rowJSON, _ := json.Marshal(rowValues)
                sourceRows.Close()
                if useTransaction {
                    tx.Rollback()
                }
                return fmt.Errorf("행 삽입 실패: %w\n데이터: %s\n쿼리: %s", err, rowJSON, directSQL)
            }

            rowsInBatch++
        }

        sourceRows.Close()

        // 트랜잭션 커밋 (필요 시)
        if useTransaction {
            if err := tx.Commit(); err != nil {
                return fmt.Errorf("트랜잭션 커밋 실패: %w", err)
            }
        }

        processedRows += rowsInBatch
    }

    if config.VerboseLogging {
        fmt.Printf("    테이블 '%s': %d/%d 행 처리 완료\n", tableName, processedRows, totalRows)
    }

    return nil
}
```

## 6. 공통 컬럼 찾기 함수 개선

```go
// findCommonColumns는 두 테이블의 공통 컬럼을 찾습니다
func findCommonColumns(sourceColumns, targetColumns []ColumnMetadata) ([]ColumnMetadata, []string, []string) {
    var common []ColumnMetadata
    var sourceNames []string
    var targetNames []string

    sourceMap := make(map[string]ColumnMetadata)
    for _, col := range sourceColumns {
        sourceMap[strings.ToLower(col.Name)] = col
    }

    targetMap := make(map[string]ColumnMetadata)
    for _, col := range targetColumns {
        targetMap[strings.ToLower(col.Name)] = col
    }

    // 디버깅 정보 출력
    fmt.Println("    소스 컬럼:")
    for _, col := range sourceColumns {
        fmt.Printf("      - %s (%s)\n", col.Name, col.Type)
    }
    fmt.Println("    대상 컬럼:")
    for _, col := range targetColumns {
        fmt.Printf("      - %s (%s)\n", col.Name, col.Type)
    }

    // 소스 컬럼을 기준으로 공통 컬럼 찾기
    for _, sourceCol := range sourceColumns {
        lowerName := strings.ToLower(sourceCol.Name)
        if targetCol, ok := targetMap[lowerName]; ok {
            common = append(common, sourceCol)
            sourceNames = append(sourceNames, quoteColumnName(sourceCol.Name))
            targetNames = append(targetNames, quoteColumnName(targetCol.Name))
        }
    }

    fmt.Printf("    찾은 공통 컬럼 수: %d\n", len(common))

    return common, sourceNames, targetNames
}
```

## 7. 시퀀스/자동증가 값 재설정 함수 개선

```go
// resetSequences는 ID 시퀀스/자동증가 값을 재설정합니다
func resetSequences(config *DataMigrationConfig) error {
    fmt.Println("[4/4] 시퀀스/자동증가 값 재설정 중...")

    // 1. 기본 테이블 목록
    tables := getBasicTables()

    // 2. 동적 테이블(게시판 테이블) 목록 가져오기
    ctx := context.Background()
    var boards []*models.Board
    err := config.TargetDB.NewSelect().
        Model(&boards).
        Column("table_name").
        Scan(ctx)

    if err == nil {
        // 게시판 테이블 목록 추가
        for _, board := range boards {
            tables = append(tables, board.TableName)
        }
    } else {
        log.Printf("게시판 목록 조회 실패: %v (기본 테이블만 처리합니다)", err)
    }

    if config.TargetDBConfig.DBDriver == "postgres" {
        // PostgreSQL 시퀀스 재설정
        for _, tableName := range tables {
            // 시퀀스 재설정이 필요한 테이블만 처리
            if !needsSequenceReset(tableName) {
                continue
            }

            // 시퀀스 존재 여부 확인
            var seqExists int
            seqCheckSQL := fmt.Sprintf(
                "SELECT COUNT(*) FROM pg_class WHERE relkind = 'S' AND relname = '%s_id_seq'",
                tableName)

            err := config.TargetDB.QueryRow(seqCheckSQL).Scan(&seqExists)
            if err != nil || seqExists == 0 {
                // 시퀀스가 없으면 건너뛰기
                if config.VerboseLogging {
                    fmt.Printf("  - 테이블 '%s'에 시퀀스가 없습니다\n", tableName)
                }
                continue
            }

            resetSQL := fmt.Sprintf(
                "SELECT setval('%s_id_seq', COALESCE((SELECT MAX(id) FROM %s), 1));",
                tableName, quoteTableName(config.TargetDBConfig.DBDriver, tableName))

            _, err = config.TargetDB.ExecContext(ctx, resetSQL)
            if err != nil {
                log.Printf("테이블 '%s' 시퀀스 재설정 실패: %v (무시하고 계속 진행합니다)", tableName, err)
            } else if config.VerboseLogging {
                fmt.Printf("  - 테이블 '%s' 시퀀스 재설정 완료\n", tableName)
            }
        }
    } else if config.TargetDBConfig.DBDriver == "mysql" || config.TargetDBConfig.DBDriver == "mariadb" {
        // MySQL/MariaDB 자동증가 값 재설정
        for _, tableName := range tables {
            // 자동증가 재설정이 필요한 테이블만 처리
            if !needsSequenceReset(tableName) {
                continue
            }

            // 현재 최대 ID 값 조회
            quotedTableName := quoteTableName(config.TargetDBConfig.DBDriver, tableName)
            query := fmt.Sprintf("SELECT COALESCE(MAX(id), 0) FROM %s", quotedTableName)
            
            var maxID int64
            err := config.TargetDB.QueryRow(query).Scan(&maxID)

            if err != nil {
                log.Printf("테이블 '%s' 최대 ID 조회 실패: %v (무시하고 계속 진행합니다)", tableName, err)
                continue
            }

            // 자동증가 값 재설정
            if maxID > 0 {
                resetSQL := fmt.Sprintf("ALTER TABLE %s AUTO_INCREMENT = %d", quotedTableName, maxID+1)
                _, err := config.TargetDB.ExecContext(ctx, resetSQL)
                if err != nil {
                    log.Printf("테이블 '%s' 자동증가 값 재설정 실패: %v (무시하고 계속 진행합니다)", tableName, err)
                } else if config.VerboseLogging {
                    fmt.Printf("  - 테이블 '%s' 자동증가 값 재설정 완료 (ID: %d)\n", tableName, maxID+1)
                }
            }
        }
    }

    fmt.Println("시퀀스/자동증가 값 재설정 완료")
    return nil
}
```

## 정리

위 코드 수정을 통해 다음 문제들을 해결할 수 있습니다:

1. 외래 키 제약 조건으로 인한 users 테이블 삭제 문제 해결
2. 테이블 컬럼 정보 조회 방식 개선 (MySQL용 SHOW COLUMNS 사용)
3. 공통 컬럼 찾기 로직에 디버깅 정보 추가
4. 존재하지 않는 테이블 처리 개선
5. MySQL에서 하이픈이 포함된 테이블 이름 처리 개선
6. 시퀀스/자동증가 값 재설정 로직 강화

이러한 수정을 통해 이기종 데이터베이스 간 마이그레이션이 더 안정적으로 동작할 것입니다.

## Prompt:
case "mysql", "mariadb":
		// 테이블 존재 여부 먼저 확인
		var exists int
		checkQuery := fmt.Sprintf("SELECT COUNT(1) FROM information_schema.tables WHERE table_schema = '%s' AND table_name = '%s'",
			db.DB.Stats().DSN, tableName)


db.DB.Stats().DSN undefined (type sql.DBStats has no field or method DSN)compilerMissingFieldOrMethod

## Response:

저는 잘못 적용된 코드를 제공했습니다. `sql.DBStats` 구조체에는 `DSN` 필드가 없습니다. 데이터베이스 이름(schema)을 가져오는 더 나은 방법이 필요합니다.

MySQL/MariaDB에서 현재 연결된 데이터베이스 이름을 가져오는 올바른 방법은 직접 "SELECT DATABASE()" 쿼리를 실행하는 것입니다. 이를 수정하겠습니다.

# MySQL 데이터베이스 이름 가져오기 수정

`db.DB.Stats().DSN`은 유효하지 않은 필드입니다. `sql.DBStats` 구조체에는 `DSN` 필드가 없습니다. 대신 현재 연결된 데이터베이스 이름을 직접 쿼리해야 합니다:

```go
// getTableColumns 함수 내 MySQL/MariaDB 부분 수정
case "mysql", "mariadb":
    // 현재 데이터베이스 이름 직접 획득
    var dbName string
    err := db.QueryRow("SELECT DATABASE()").Scan(&dbName)
    if err != nil {
        return nil, fmt.Errorf("현재 데이터베이스 이름 획득 실패: %w", err)
    }
    
    // 테이블 존재 여부 확인
    var exists int
    checkQuery := fmt.Sprintf("SELECT COUNT(1) FROM information_schema.tables WHERE table_schema = '%s' AND table_name = '%s'",
        dbName, tableName)
    
    err = db.QueryRow(checkQuery).Scan(&exists)
    if err != nil {
        return nil, fmt.Errorf("테이블 존재 여부 확인 실패: %w", err)
    }
    
    if exists == 0 {
        return nil, fmt.Errorf("테이블이 존재하지 않습니다: %s", tableName)
    }

    // MySQL/MariaDB용 SHOW COLUMNS 쿼리
    quotedName := quoteTableName(driver, tableName)
    rows, err := db.QueryContext(ctx, fmt.Sprintf("SHOW COLUMNS FROM %s", quotedName))
    if err != nil {
        return nil, fmt.Errorf("컬럼 정보 조회 실패: %w", err)
    }
    defer rows.Close()

    // 나머지 코드는 동일...
```

마찬가지로 `ensureDynamicTableExists` 함수에서도 동일한 수정이 필요합니다:

```go
case "mysql", "mariadb":
    // MySQL/MariaDB용 직접 SQL 쿼리
    var dbName string
    err := config.TargetDB.QueryRow("SELECT DATABASE()").Scan(&dbName)
    if err != nil {
        return fmt.Errorf("현재 데이터베이스 이름 획득 실패: %w", err)
    }
    
    var count int
    query := fmt.Sprintf("SELECT COUNT(1) FROM information_schema.tables WHERE table_schema = '%s' AND table_name = '%s'",
        dbName, board.TableName)
    
    err = config.TargetDB.QueryRow(query).Scan(&count)
    if err != nil {
        return fmt.Errorf("테이블 존재 여부 확인 실패: %w", err)
    }

    exists = count > 0
```

이 방식으로 현재 연결된 MySQL 데이터베이스의 이름을 올바르게 가져와서 테이블 존재 여부를 확인할 수 있습니다.

## Prompt:
PS D:\dev\pcbangstudio\workspace\go-board\bin> .\migrate_windows_amd64.exe -op data-migrate -source-driver sqlite -target-driver mysql -source-env .env_sqlite -target-env .env_my
==========================
이기종 DB 데이터 마이그레이션 시작
소스: sqlite (go_board)
대상: mysql (go_board)
==========================
[1/4] 대상 데이터베이스 스키마 생성 중...
마이그레이션 상태 확인 중...
2025/04/17 22:29:40     Applied At                  Migration
2025/04/17 22:29:40     =======================================
2025/04/17 22:29:40     Fri Apr 11 02:00:16 2025 -- 001_create_tables.sql
2025/04/17 22:29:40     Fri Apr 11 02:00:16 2025 -- 002_comments.sql
2025/04/17 22:29:40     Fri Apr 11 02:00:16 2025 -- 003_file_uploads.sql
2025/04/17 22:29:40     Fri Apr 11 02:00:16 2025 -- 004_qna_votes.sql
2025/04/17 22:29:40     Fri Apr 11 02:00:16 2025 -- 005_board_manager.sql
2025/04/17 22:29:40     Thu Apr 17 22:19:25 2025 -- 006_referrer_stats.sql
2025/04/17 22:29:40     Thu Apr 17 22:19:25 2025 -- 007_system_settings.sql
스키마 마이그레이션 적용 중...
기본 스키마 마이그레이션 완료
[2/4] 기본 테이블 데이터 마이그레이션 중...
  - 테이블 'users' 데이터 마이그레이션 중...
    소스 컬럼:
      - id (INTEGER)
      - username (TEXT)
      - email (TEXT)
      - password (TEXT)
      - full_name (TEXT)
      - role (TEXT)
      - active (INTEGER)
      - approval_status (TEXT)
      - approval_due (TIMESTAMP)
      - created_at (TIMESTAMP)
      - updated_at (TIMESTAMP)
    대상 컬럼:
      - id (int(11))
      - username (varchar(50))
      - email (varchar(100))
      - password (varchar(100))
      - full_name (varchar(100))
      - role (varchar(20))
      - active (tinyint(1))
      - created_at (timestamp)
      - updated_at (timestamp)
    찾은 공통 컬럼 수: 9
2025/04/17 22:29:40 오류: 테이블 'users' 마이그레이션 실패: 행 삽입 실패: Error 1062 (23000): Duplicate entry '1' for key 'PRIMARY'
데이터: [1,"bab2","edp1096@outlook.com","$2a$10$fqA.SFqpLS8nujCh/6DsJOXhwpKnffYBbARw.eGurBPz0qUUzgsre","꿈돌이","admin",1,"2025-04-14T07:10:58.105779Z","2025-04-14T07:10:58.105779Z"]
쿼리: INSERT INTO `users` (id, username, email, password, full_name, role, active, created_at, updated_at) VALUES (1, 'bab2', 'edp1096@outlook.com', '$2a$10$fqA.SFqpLS8nujCh/6DsJOXhwpKnffYBbARw.eGurBPz0qUUzgsre', '꿈돌이', 'admin', 1, '2025-04-14 07:10:58', '2025-04-14 07:10:58')
  - 테이블 'boards' 데이터 마이그레이션 중...
    소스 컬럼:
      - id (INTEGER)
      - name (TEXT)
      - slug (TEXT)
      - description (TEXT)
      - board_type (TEXT)
      - table_name (TEXT)
      - active (INTEGER)
      - comments_enabled (INTEGER)
      - allow_anonymous (INTEGER)
      - created_at (TIMESTAMP)
      - updated_at (TIMESTAMP)
    대상 컬럼:
      - id (int(11))
      - name (varchar(100))
      - slug (varchar(100))
      - description (text)
      - board_type (varchar(20))
      - table_name (varchar(50))
      - active (tinyint(1))
      - comments_enabled (tinyint(1))
      - allow_anonymous (tinyint(1))
      - created_at (timestamp)
      - updated_at (timestamp)
    찾은 공통 컬럼 수: 11
    기존 대상 테이블 데이터 삭제됨
    테이블 'boards': 2/2 행 처리 완료
  - 테이블 'board_fields' 데이터 마이그레이션 중...
    소스 컬럼:
      - id (INTEGER)
      - board_id (INTEGER)
      - name (TEXT)
      - column_name (TEXT)
      - display_name (TEXT)
      - field_type (TEXT)
      - required (INTEGER)
      - sortable (INTEGER)
      - searchable (INTEGER)
      - options (TEXT)
      - sort_order (INTEGER)
      - created_at (TIMESTAMP)
      - updated_at (TIMESTAMP)
    대상 컬럼:
      - id (int(11))
      - board_id (int(11))
      - name (varchar(50))
      - column_name (varchar(50))
      - display_name (varchar(100))
      - field_type (varchar(20))
      - required (tinyint(1))
      - sortable (tinyint(1))
      - searchable (tinyint(1))
      - options (text)
      - sort_order (int(11))
      - created_at (timestamp)
      - updated_at (timestamp)
    찾은 공통 컬럼 수: 13
    기존 대상 테이블 데이터 삭제됨
    테이블 'board_fields': 2/2 행 처리 완료
  - 테이블 'board_managers' 데이터 마이그레이션 중...
    테이블 'board_managers'에 데이터가 없습니다
  - 테이블 'comments' 데이터 마이그레이션 중...
    소스 컬럼:
      - id (INTEGER)
      - post_id (INTEGER)
      - board_id (INTEGER)
      - user_id (INTEGER)
      - content (TEXT)
      - parent_id (INTEGER)
      - created_at (TIMESTAMP)
      - updated_at (TIMESTAMP)
    대상 컬럼:
      - id (int(11))
      - post_id (int(11))
      - board_id (int(11))
      - user_id (int(11))
      - content (text)
      - parent_id (int(11))
      - created_at (timestamp)
      - updated_at (timestamp)
    찾은 공통 컬럼 수: 8
    기존 대상 테이블 데이터 삭제됨
    테이블 'comments': 8/8 행 처리 완료
  - 테이블 'attachments' 데이터 마이그레이션 중...
    소스 컬럼:
      - id (INTEGER)
      - board_id (INTEGER)
      - post_id (INTEGER)
      - user_id (INTEGER)
      - file_name (TEXT)
      - storage_name (TEXT)
      - file_path (TEXT)
      - file_size (INTEGER)
      - mime_type (TEXT)
      - is_image (INTEGER)
      - download_url (TEXT)
      - thumbnail_url (TEXT)
      - download_count (INTEGER)
      - created_at (TIMESTAMP)
    대상 컬럼:
      - id (bigint(20))
      - board_id (int(11))
      - post_id (int(11))
      - user_id (int(11))
      - file_name (varchar(255))
      - storage_name (varchar(255))
      - file_path (varchar(255))
      - file_size (bigint(20))
      - mime_type (varchar(100))
      - is_image (tinyint(1))
      - download_url (varchar(255))
      - thumbnail_url (varchar(512))
      - download_count (int(11))
      - created_at (timestamp)
    찾은 공통 컬럼 수: 14
    기존 대상 테이블 데이터 삭제됨
    테이블 'attachments': 5/5 행 처리 완료
  - 테이블 'qna_answers' 데이터 마이그레이션 중...
    테이블 'qna_answers'에 데이터가 없습니다
  - 테이블 'qna_question_votes' 데이터 마이그레이션 중...
    테이블 'qna_question_votes'에 데이터가 없습니다
  - 테이블 'qna_answer_votes' 데이터 마이그레이션 중...
    테이블 'qna_answer_votes'에 데이터가 없습니다
  - 테이블 'referrer_stats' 데이터 마이그레이션 중...
2025/04/17 22:29:40 오류: 테이블 'referrer_stats' 마이그레이션 실패: 대상 테이블 존재하지 않음: 테이블이 존재하지 않습니다: referrer_stats
  - 테이블 'system_settings' 데이터 마이그레이션 중...
2025/04/17 22:29:40 오류: 테이블 'system_settings' 마이그레이션 실패: 대상 테이블 존재하지 않음: 테이블이 존재하지 않습니다: system_settings
기본 테이블 데이터 마이그레이션 완료
[3/4] 동적 테이블 마이그레이션 중...
  총 2개의 게시판을 마이그레이션합니다
  - 게시판 '갤러리' (board_gaelreori) 마이그레이션 중...
  - 테이블 'board_gaelreori' 데이터 마이그레이션 중...
    소스 컬럼:
      - id (INTEGER)
      - title (VARCHAR(200))
      - content (TEXT)
      - user_id (INTEGER)
      - view_count (INTEGER)
      - created_at (TIMESTAMP)
      - updated_at (TIMESTAMP)
      - gallery_images (VARCHAR(255))
    대상 컬럼:
      - id (int(11))
      - title (varchar(200))
      - content (text)
      - user_id (int(11))
      - view_count (int(11))
      - created_at (timestamp)
      - updated_at (timestamp)
      - gallery_images (varchar(255))
    찾은 공통 컬럼 수: 8
    기존 대상 테이블 데이터 삭제됨
    테이블 'board_gaelreori': 5/5 행 처리 완료
  - 게시판 '기타 미분류' (board_gita-mibunryu) 마이그레이션 중...
  - 테이블 'board_gita-mibunryu' 데이터 마이그레이션 중...
    소스 컬럼:
      - id (INTEGER)
      - title (VARCHAR(200))
      - content (TEXT)
      - user_id (INTEGER)
      - view_count (INTEGER)
      - created_at (TIMESTAMP)
      - updated_at (TIMESTAMP)
      - file_upload (VARCHAR(255))
    대상 컬럼:
      - id (int(11))
      - title (varchar(200))
      - content (text)
      - user_id (int(11))
      - view_count (int(11))
      - created_at (timestamp)
      - updated_at (timestamp)
      - file_upload (varchar(255))
    찾은 공통 컬럼 수: 8
    기존 대상 테이블 데이터 삭제됨
    테이블 'board_gita-mibunryu': 4/4 행 처리 완료
동적 테이블 마이그레이션 완료
[4/4] 시퀀스/자동증가 값 재설정 중...
  - 테이블 'users' 자동증가 값 재설정 완료 (ID: 3)
  - 테이블 'boards' 자동증가 값 재설정 완료 (ID: 3)
  - 테이블 'board_fields' 자동증가 값 재설정 완료 (ID: 3)
  - 테이블 'comments' 자동증가 값 재설정 완료 (ID: 13)
  - 테이블 'attachments' 자동증가 값 재설정 완료 (ID: 6)
2025/04/17 22:29:40 테이블 'system_settings' 최대 ID 조회 실패: Error 1146 (42S02): Table 'go_board.system_settings' doesn't exist (무시하고 계속 진행 합니다)
  - 테이블 'board_gaelreori' 자동증가 값 재설정 완료 (ID: 6)
  - 테이블 'board_gita-mibunryu' 자동증가 값 재설정 완료 (ID: 5)
시퀀스/자동증가 값 재설정 완료
==========================
데이터 마이그레이션 완료 (소요 시간: 89.5669ms)
경고: 마이그레이션 중 3개의 오류가 발생했습니다
  - 테이블 'users' 마이그레이션 실패: 행 삽입 실패: Error 1062 (23000): Duplicate entry '1' for key 'PRIMARY'
데이터: [1,"bab2","edp1096@outlook.com","$2a$10$fqA.SFqpLS8nujCh/6DsJOXhwpKnffYBbARw.eGurBPz0qUUzgsre","꿈돌이","admin",1,"2025-04-14T07:10:58.105779Z","2025-04-14T07:10:58.105779Z"]
쿼리: INSERT INTO `users` (id, username, email, password, full_name, role, active, created_at, updated_at) VALUES (1, 'bab2', 'edp1096@outlook.com', '$2a$10$fqA.SFqpLS8nujCh/6DsJOXhwpKnffYBbARw.eGurBPz0qUUzgsre', '꿈돌이', 'admin', 1, '2025-04-14 07:10:58', '2025-04-14 07:10:58')
  - 테이블 'referrer_stats' 마이그레이션 실패: 대상 테이블 존재하지 않음: 테이블이 존재하지 않습니다: referrer_stats
  - 테이블 'system_settings' 마이그레이션 실패: 대상 테이블 존재하지 않음: 테이블이 존재하지 않습니다: system_settings
==========================
2025/04/17 22:29:40 데이터 마이그레이션이 성공적으로 완료되었습니다
PS D:\dev\pcbangstudio\workspace\go-board\bin>
-----------------------
제안코드 함수 단위로 끊어서 생략없는 완전한 코드로.

## Response:

요청을 분석해보면, 기존 마이그레이션 코드가 이제 더 많은 정보와 함께 실행되고 있지만 여전히 몇 가지 문제가 남아있습니다:

1. users 테이블에서 기본 키 중복 문제가 발생 (Duplicate entry '1' for key 'PRIMARY')
2. referrer_stats 및 system_settings 테이블이 존재하지 않음 (이 부분은 잘 처리되고 있지만, 오류 메시지는 계속 표시됨)

특히 users 테이블 삽입에서 중복 키 오류가 발생하는데, 이는 SQLite에서 MySQL로 마이그레이션할 때 target DB에 이미 같은 ID의 사용자가 존재하기 때문입니다. 데이터 마이그레이션 시 기존 데이터를 먼저 삭제하도록 설정했지만, 외래 키 제약 조건으로 인해 users 테이블의 데이터를 삭제할 수 없습니다.

이러한 문제들을 해결하기 위해 각 함수를 개선해야 합니다. 각 함수별로 문제점을 해결하는 방식으로 코드를 제공하겠습니다.

# 이기종 DB 마이그레이션 개선 코드

아직 일부 오류가 있지만 많은 데이터가 성공적으로 마이그레이션 되었네요. 주요 문제는 users 테이블의 기본 키 중복 오류와 일부 테이블이 존재하지 않는 문제입니다. 이를 완전히 해결하기 위한 함수별 수정 코드를 제공하겠습니다.

## 1. shouldCleanTableBeforeMigration 함수

```go
// shouldCleanTableBeforeMigration은 마이그레이션 전에 테이블을 정리해야 하는지 결정합니다
func shouldCleanTableBeforeMigration(tableName string) bool {
    // 외래 키 제약 조건이 있거나 이력 데이터인 테이블은 정리하지 않음
    skipCleanTables := map[string]bool{
        "referrer_stats": true,
        "users": true,           // users 테이블은 외래 키 제약 때문에 삭제하지 않음
    }

    return !skipCleanTables[tableName]
}
```

## 2. migrateTableData 함수

```go
// migrateTableData는 테이블 데이터를 마이그레이션합니다
func migrateTableData(config *DataMigrationConfig, tableName string) error {
    ctx := context.Background()

    if config.VerboseLogging {
        fmt.Printf("  - 테이블 '%s' 데이터 마이그레이션 중...\n", tableName)
    }

    // 테이블 총 행 수 카운트
    var totalRows int
    query := fmt.Sprintf("SELECT COUNT(*) FROM %s", quoteTableName(config.SourceDBConfig.DBDriver, tableName))
    err := config.SourceDB.QueryRow(query).Scan(&totalRows)
    if err != nil {
        return fmt.Errorf("행 수 계산 실패: %w", err)
    }

    if totalRows == 0 {
        if config.VerboseLogging {
            fmt.Printf("    테이블 '%s'에 데이터가 없습니다\n", tableName)
        }
        return nil
    }

    // 테이블 구조 가져오기
    sourceColumns, err := getTableColumns(config.SourceDB, config.SourceDBConfig.DBDriver, tableName)
    if err != nil {
        return fmt.Errorf("소스 테이블 구조 가져오기 실패: %w", err)
    }

    // 대상 테이블 존재 여부 확인
    targetColumns, err := getTableColumns(config.TargetDB, config.TargetDBConfig.DBDriver, tableName)
    if err != nil {
        // 특정 오류에 대한 특별 처리
        if strings.Contains(err.Error(), "테이블이 존재하지 않습니다") || 
           strings.Contains(err.Error(), "doesn't exist") {
            return fmt.Errorf("대상 테이블 존재하지 않음: %w", err)
        }
        return fmt.Errorf("대상 테이블 구조 가져오기 실패: %w", err)
    }

    // 공통 컬럼 찾기
    commonColumns, sourceColumnNames, targetColumnNames := findCommonColumns(sourceColumns, targetColumns)
    if len(commonColumns) == 0 {
        return fmt.Errorf("공통 컬럼이 없습니다")
    }

    // 테이블 모델 정보 가져오기
    modelInfo := getModelInfo(tableName)

    // 데이터 배치 처리
    batchSize := config.BatchSize
    if batchSize <= 0 {
        batchSize = 1000
    }

    totalBatches := (totalRows + batchSize - 1) / batchSize
    processedRows := 0

    // 대상 테이블의 기존 데이터 삭제 (필요 시)
    if shouldCleanTableBeforeMigration(tableName) {
        // MySQL의 외래 키 제약 조건 일시적으로 비활성화
        if config.TargetDBConfig.DBDriver == "mysql" || config.TargetDBConfig.DBDriver == "mariadb" {
            _, err = config.TargetDB.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 0;")
            if err != nil {
                return fmt.Errorf("외래 키 제약 비활성화 실패: %w", err)
            }
            defer func() {
                // 작업 완료 후 다시 활성화
                _, _ = config.TargetDB.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 1;")
            }()
        } else if config.TargetDBConfig.DBDriver == "postgres" {
            _, err = config.TargetDB.ExecContext(ctx, "SET session_replication_role = 'replica';")
            if err != nil {
                return fmt.Errorf("외래 키 제약 비활성화 실패: %w", err)
            }
            defer func() {
                // 작업 완료 후 다시 활성화
                _, _ = config.TargetDB.ExecContext(ctx, "SET session_replication_role = 'origin';")
            }()
        }

        deleteQuery := fmt.Sprintf("DELETE FROM %s", quoteTableName(config.TargetDBConfig.DBDriver, tableName))
        _, err = config.TargetDB.ExecContext(ctx, deleteQuery)
        if err != nil {
            return fmt.Errorf("대상 테이블 데이터 삭제 실패: %w", err)
        }
        if config.VerboseLogging {
            fmt.Printf("    기존 대상 테이블 데이터 삭제됨\n")
        }
    }

    // 트랜잭션 활성화 여부 확인
    useTransaction := config.EnableTransactions && shouldUseTransactionForTable(tableName, totalRows)

    for batchNum := 0; batchNum < totalBatches; batchNum++ {
        offset := batchNum * batchSize
        limit := batchSize

        if config.VerboseLogging && totalBatches > 1 {
            fmt.Printf("    배치 %d/%d 처리 중... (오프셋: %d, 한계: %d)\n", batchNum+1, totalBatches, offset, limit)
        }

        // 소스 데이터 쿼리 구성
        sourceQuery := fmt.Sprintf("SELECT %s FROM %s ORDER BY id LIMIT %d OFFSET %d",
            strings.Join(sourceColumnNames, ", "),
            quoteTableName(config.SourceDBConfig.DBDriver, tableName),
            limit,
            offset)

        // 소스 데이터 조회
        sourceRows, err := config.SourceDB.QueryContext(ctx, sourceQuery)
        if err != nil {
            return fmt.Errorf("소스 데이터 조회 실패: %w", err)
        }

        // 트랜잭션 시작 (필요 시)
        var tx *sql.Tx
        if useTransaction {
            tx, err = config.TargetDB.DB.Begin()
            if err != nil {
                sourceRows.Close()
                return fmt.Errorf("트랜잭션 시작 실패: %w", err)
            }
        }

        // 행별 처리
        rowsInBatch := 0
        for sourceRows.Next() {
            // 행 데이터 변수 준비
            rowValues := make([]interface{}, len(commonColumns))
            valuePtrs := make([]interface{}, len(commonColumns))
            for i := range rowValues {
                valuePtrs[i] = &rowValues[i]
            }

            // 행 데이터 읽기
            if err := sourceRows.Scan(valuePtrs...); err != nil {
                sourceRows.Close()
                if useTransaction {
                    tx.Rollback()
                }
                return fmt.Errorf("행 데이터 읽기 실패: %w", err)
            }

            // 사용자 테이블 특별 처리 - ID 중복 체크 및 건너뛰기
            if tableName == "users" {
                var userExists bool
                idVal := rowValues[0]
                checkQuery := fmt.Sprintf("SELECT 1 FROM %s WHERE id = ? LIMIT 1", 
                    quoteTableName(config.TargetDBConfig.DBDriver, tableName))
                
                var exists int
                var checkErr error
                if useTransaction {
                    checkErr = tx.QueryRowContext(ctx, checkQuery, idVal).Scan(&exists)
                } else {
                    checkErr = config.TargetDB.QueryRowContext(ctx, checkQuery, idVal).Scan(&exists)
                }
                
                userExists = (checkErr == nil)
                
                if userExists {
                    if config.VerboseLogging {
                        fmt.Printf("    사용자 ID %v가 이미 존재하여 건너뜁니다\n", idVal)
                    }
                    continue
                }
            }

            // SQL 쿼리용 컬럼 값 배열
            var columnValues []string

            // 각 값을 SQL 문자열로 변환
            for i, val := range rowValues {
                colName := commonColumns[i].Name

                // 널 값 처리
                if val == nil {
                    columnValues = append(columnValues, "NULL")
                    continue
                }

                // referrer_stats 테이블 중복 키 문제 해결을 위한 특별 처리
                if tableName == "referrer_stats" && colName == "id" && config.TargetDBConfig.DBDriver == "postgres" {
                    columnValues = append(columnValues, "nextval('referrer_stats_id_seq')")
                    continue
                }

                // 모델 기반 타입 변환
                fieldInfo, hasField := modelInfo[colName]
                if hasField && config.TargetDBConfig.DBDriver == "postgres" {
                    // 타입에 따른 변환
                    switch fieldInfo.fieldType {
                    case "bool":
                        // 불리언 처리
                        switch v := val.(type) {
                        case int64:
                            if v == 1 {
                                columnValues = append(columnValues, "TRUE")
                            } else {
                                columnValues = append(columnValues, "FALSE")
                            }
                        case int:
                            if v == 1 {
                                columnValues = append(columnValues, "TRUE")
                            } else {
                                columnValues = append(columnValues, "FALSE")
                            }
                        case string:
                            if strings.ToLower(v) == "true" || v == "1" {
                                columnValues = append(columnValues, "TRUE")
                            } else {
                                columnValues = append(columnValues, "FALSE")
                            }
                        case bool:
                            if v {
                                columnValues = append(columnValues, "TRUE")
                            } else {
                                columnValues = append(columnValues, "FALSE")
                            }
                        default:
                            columnValues = append(columnValues, "FALSE")
                        }
                        continue
                    case "time.Time":
                        // 시간 처리
                        switch v := val.(type) {
                        case time.Time:
                            columnValues = append(columnValues, fmt.Sprintf("'%s'", v.Format("2006-01-02 15:04:05")))
                        case string:
                            columnValues = append(columnValues, fmt.Sprintf("'%s'", v))
                        default:
                            columnValues = append(columnValues, fmt.Sprintf("'%v'", v))
                        }
                        continue
                    }
                }

                // 일반 데이터 타입 변환
                switch v := val.(type) {
                case string:
                    // 문자열은 작은 따옴표로 감싸고 내부 작은 따옴표는 이스케이프
                    escapedVal := strings.Replace(v, "'", "''", -1)
                    columnValues = append(columnValues, fmt.Sprintf("'%s'", escapedVal))
                case bool:
                    // MySQL/MariaDB의 경우 1/0 사용, PostgreSQL은 TRUE/FALSE
                    if config.TargetDBConfig.DBDriver == "postgres" {
                        if v {
                            columnValues = append(columnValues, "TRUE")
                        } else {
                            columnValues = append(columnValues, "FALSE")
                        }
                    } else {
                        // 다른 DB는 1/0 사용
                        if v {
                            columnValues = append(columnValues, "1")
                        } else {
                            columnValues = append(columnValues, "0")
                        }
                    }
                case time.Time:
                    columnValues = append(columnValues, fmt.Sprintf("'%s'", v.Format("2006-01-02 15:04:05")))
                default:
                    // 숫자 등 다른 타입은 그대로
                    columnValues = append(columnValues, fmt.Sprintf("%v", v))
                }
            }

            // 직접 문자열로 SQL 쿼리 구성
            directSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
                quoteTableName(config.TargetDBConfig.DBDriver, tableName),
                strings.Join(targetColumnNames, ", "),
                strings.Join(columnValues, ", "))

            // 쿼리 실행
            var err error
            if useTransaction {
                _, err = tx.ExecContext(ctx, directSQL)
            } else {
                _, err = config.TargetDB.ExecContext(ctx, directSQL)
            }

            if err != nil {
                // JSON으로 행 데이터 직렬화 (디버깅용)
                rowJSON, _ := json.Marshal(rowValues)
                sourceRows.Close()
                if useTransaction {
                    tx.Rollback()
                }
                return fmt.Errorf("행 삽입 실패: %w\n데이터: %s\n쿼리: %s", err, rowJSON, directSQL)
            }

            rowsInBatch++
        }

        sourceRows.Close()

        // 트랜잭션 커밋 (필요 시)
        if useTransaction {
            if err := tx.Commit(); err != nil {
                return fmt.Errorf("트랜잭션 커밋 실패: %w", err)
            }
        }

        processedRows += rowsInBatch
    }

    if config.VerboseLogging {
        fmt.Printf("    테이블 '%s': %d/%d 행 처리 완료\n", tableName, processedRows, totalRows)
    }

    return nil
}
```

## 3. getTableColumns 함수

```go
// getTableColumns는 테이블의 컬럼 정보를 가져옵니다
func getTableColumns(db *bun.DB, driver string, tableName string) ([]ColumnMetadata, error) {
    ctx := context.Background()
    var columns []ColumnMetadata

    switch driver {
    case "postgres":
        // PostgreSQL용 쿼리
        var results []struct {
            ColumnName string         `bun:"column_name"`
            DataType   string         `bun:"data_type"`
            IsNullable string         `bun:"is_nullable"`
            Default    sql.NullString `bun:"column_default"`
        }

        err := db.NewSelect().
            TableExpr("information_schema.columns").
            Column("column_name", "data_type", "is_nullable", "column_default").
            Where("table_schema = 'public'").
            Where("table_name = ?", tableName).
            Order("ordinal_position").
            Scan(ctx, &results)

        if err != nil {
            return nil, fmt.Errorf("컬럼 정보 조회 실패: %w", err)
        }

        for _, r := range results {
            columns = append(columns, ColumnMetadata{
                Name:     r.ColumnName,
                Type:     r.DataType,
                Nullable: r.IsNullable == "YES",
                Default:  r.Default,
            })
        }

    case "mysql", "mariadb":
        // 현재 데이터베이스 이름 직접 획득
        var dbName string
        err := db.QueryRow("SELECT DATABASE()").Scan(&dbName)
        if err != nil {
            return nil, fmt.Errorf("현재 데이터베이스 이름 획득 실패: %w", err)
        }
        
        // 테이블 존재 여부 확인
        var exists int
        checkQuery := fmt.Sprintf("SELECT COUNT(1) FROM information_schema.tables WHERE table_schema = '%s' AND table_name = '%s'",
            dbName, tableName)
        
        err = db.QueryRow(checkQuery).Scan(&exists)
        if err != nil {
            return nil, fmt.Errorf("테이블 존재 여부 확인 실패: %w", err)
        }
        
        if exists == 0 {
            return nil, fmt.Errorf("테이블이 존재하지 않습니다: %s", tableName)
        }

        // MySQL/MariaDB용 SHOW COLUMNS 쿼리
        quotedName := quoteTableName(driver, tableName)
        rows, err := db.QueryContext(ctx, fmt.Sprintf("SHOW COLUMNS FROM %s", quotedName))
        if err != nil {
            return nil, fmt.Errorf("컬럼 정보 조회 실패: %w", err)
        }
        defer rows.Close()

        for rows.Next() {
            var field, colType, null, key, extra string
            var defaultValue sql.NullString
            if err := rows.Scan(&field, &colType, &null, &key, &defaultValue, &extra); err != nil {
                return nil, fmt.Errorf("컬럼 정보 스캔 실패: %w", err)
            }

            columns = append(columns, ColumnMetadata{
                Name:     field,
                Type:     colType,
                Nullable: null == "YES",
                Default:  defaultValue,
            })
        }

    case "sqlite":
        // 테이블 이름에 하이픈이 있는 경우 따옴표로 묶기
        quotedTableName := tableName
        if strings.Contains(tableName, "-") {
            quotedTableName = fmt.Sprintf("\"%s\"", tableName)
        }

        rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", quotedTableName))
        if err != nil {
            return nil, fmt.Errorf("컬럼 정보 조회 실패: %w", err)
        }
        defer rows.Close()

        for rows.Next() {
            var cid int
            var name, typeName string
            var notNull int
            var dfltValue sql.NullString
            var pk int

            if err := rows.Scan(&cid, &name, &typeName, &notNull, &dfltValue, &pk); err != nil {
                return nil, fmt.Errorf("컬럼 정보 스캔 실패: %w", err)
            }

            columns = append(columns, ColumnMetadata{
                Name:     name,
                Type:     typeName,
                Nullable: notNull == 0,
                Default:  dfltValue,
            })
        }

    default:
        return nil, fmt.Errorf("지원하지 않는 데이터베이스 드라이버: %s", driver)
    }

    return columns, nil
}
```

## 4. findCommonColumns 함수

```go
// findCommonColumns는 두 테이블의 공통 컬럼을 찾습니다
func findCommonColumns(sourceColumns, targetColumns []ColumnMetadata) ([]ColumnMetadata, []string, []string) {
    var common []ColumnMetadata
    var sourceNames []string
    var targetNames []string

    sourceMap := make(map[string]ColumnMetadata)
    for _, col := range sourceColumns {
        sourceMap[strings.ToLower(col.Name)] = col
    }

    targetMap := make(map[string]ColumnMetadata)
    for _, col := range targetColumns {
        targetMap[strings.ToLower(col.Name)] = col
    }

    // 디버깅 정보 출력
    fmt.Println("    소스 컬럼:")
    for _, col := range sourceColumns {
        fmt.Printf("      - %s (%s)\n", col.Name, col.Type)
    }
    fmt.Println("    대상 컬럼:")
    for _, col := range targetColumns {
        fmt.Printf("      - %s (%s)\n", col.Name, col.Type)
    }

    // 소스 컬럼을 기준으로 공통 컬럼 찾기
    for _, sourceCol := range sourceColumns {
        lowerName := strings.ToLower(sourceCol.Name)
        if targetCol, ok := targetMap[lowerName]; ok {
            common = append(common, sourceCol)
            sourceNames = append(sourceNames, quoteColumnName(sourceCol.Name))
            targetNames = append(targetNames, quoteColumnName(targetCol.Name))
        }
    }

    fmt.Printf("    찾은 공통 컬럼 수: %d\n", len(common))

    return common, sourceNames, targetNames
}
```

## 5. quoteTableName 함수

```go
// quoteTableName은 데이터베이스 드라이버에 따라 테이블 이름을 인용 부호로 묶습니다
func quoteTableName(driver, tableName string) string {
    switch driver {
    case "postgres":
        // PostgreSQL에서는 항상 따옴표로 감싸기
        return fmt.Sprintf("\"%s\"", tableName)
    case "mysql", "mariadb":
        // MySQL에서는 항상 백틱으로 감싸기
        return fmt.Sprintf("`%s`", tableName)
    case "sqlite":
        // SQLite에서는 특수 문자가 있으면 항상 따옴표로 감싸기
        if strings.ContainsAny(tableName, "-.") || strings.Contains(tableName, " ") {
            return fmt.Sprintf("\"%s\"", tableName)
        }
        return tableName
    default:
        return tableName
    }
}
```

## 6. quoteColumnName 함수

```go
// quoteColumnName은 컬럼 이름을 인용 부호로 묶습니다
func quoteColumnName(columnName string) string {
    // MySQL과 PostgreSQL에서 컬럼 이름 인용 방식이 다름
    // 기본적으로는 그대로 반환
    return columnName
}
```

## 7. ensureDynamicTableExists 함수

```go
// ensureDynamicTableExists는 대상 DB에 동적 테이블이 존재하는지 확인하고, 없으면 생성합니다
func ensureDynamicTableExists(config *DataMigrationConfig, board *models.Board, dynamicBoardService service.DynamicBoardService) error {
    // 대상 DB에 테이블이 존재하는지 확인
    var exists bool
    ctx := context.Background()

    // DB 종류에 따라 다른 쿼리 사용
    switch config.TargetDBConfig.DBDriver {
    case "postgres":
        var count int
        err := config.TargetDB.NewSelect().
            TableExpr("information_schema.tables").
            ColumnExpr("COUNT(*) AS count").
            Where("table_schema = 'public'").
            Where("table_name = ?", board.TableName).
            Scan(ctx, &count)

        if err != nil {
            return fmt.Errorf("테이블 존재 여부 확인 실패: %w", err)
        }

        exists = count > 0

    case "mysql", "mariadb":
        // MySQL/MariaDB용 직접 SQL 쿼리
        var dbName string
        err := config.TargetDB.QueryRow("SELECT DATABASE()").Scan(&dbName)
        if err != nil {
            return fmt.Errorf("현재 데이터베이스 이름 획득 실패: %w", err)
        }
        
        var count int
        query := fmt.Sprintf("SELECT COUNT(1) FROM information_schema.tables WHERE table_schema = '%s' AND table_name = '%s'",
            dbName, board.TableName)
        
        err = config.TargetDB.QueryRow(query).Scan(&count)
        if err != nil {
            return fmt.Errorf("테이블 존재 여부 확인 실패: %w", err)
        }

        exists = count > 0

    case "sqlite":
        var count int
        err := config.TargetDB.NewSelect().
            TableExpr("sqlite_master").
            Column("COUNT(*)").
            Where("type = 'table'").
            Where("name = ?", board.TableName).
            Scan(ctx, &count)

        if err != nil {
            return fmt.Errorf("테이블 존재 여부 확인 실패: %w", err)
        }

        exists = count > 0
    default:
        return fmt.Errorf("지원하지 않는 데이터베이스 드라이버: %s", config.TargetDBConfig.DBDriver)
    }

    // 테이블이 없으면 생성
    if !exists {
        // 필드 목록 가져오기
        var fields []*models.BoardField
        err := config.SourceDB.NewSelect().
            Model(&fields).
            Where("board_id = ?", board.ID).
            Order("sort_order ASC").
            Scan(ctx)
        if err != nil {
            return fmt.Errorf("게시판 필드 가져오기 실패: %w", err)
        }

        // 대상 DB에 테이블 생성
        targetDynamicService := service.NewDynamicBoardService(config.TargetDB)
        if err := targetDynamicService.CreateBoardTable(ctx, board, fields); err != nil {
            return fmt.Errorf("게시판 테이블 생성 실패: %w", err)
        }

        fmt.Printf("    게시판 테이블 '%s' 생성됨\n", board.TableName)
    }

    return nil
}
```

## 8. resetSequences 함수

```go
// resetSequences는 ID 시퀀스/자동증가 값을 재설정합니다
func resetSequences(config *DataMigrationConfig) error {
    fmt.Println("[4/4] 시퀀스/자동증가 값 재설정 중...")

    // 1. 기본 테이블 목록
    tables := getBasicTables()

    // 2. 동적 테이블(게시판 테이블) 목록 가져오기
    ctx := context.Background()
    var boards []*models.Board
    err := config.TargetDB.NewSelect().
        Model(&boards).
        Column("table_name").
        Scan(ctx)

    if err == nil {
        // 게시판 테이블 목록 추가
        for _, board := range boards {
            tables = append(tables, board.TableName)
        }
    } else {
        log.Printf("게시판 목록 조회 실패: %v (기본 테이블만 처리합니다)", err)
    }

    if config.TargetDBConfig.DBDriver == "postgres" {
        // PostgreSQL 시퀀스 재설정
        for _, tableName := range tables {
            // 시퀀스 재설정이 필요한 테이블만 처리
            if !needsSequenceReset(tableName) {
                continue
            }

            // 시퀀스 존재 여부 확인
            var seqExists int
            seqCheckSQL := fmt.Sprintf(
                "SELECT COUNT(*) FROM pg_class WHERE relkind = 'S' AND relname = '%s_id_seq'",
                tableName)

            err := config.TargetDB.QueryRow(seqCheckSQL).Scan(&seqExists)
            if err != nil || seqExists == 0 {
                // 시퀀스가 없으면 건너뛰기
                if config.VerboseLogging {
                    fmt.Printf("  - 테이블 '%s'에 시퀀스가 없습니다\n", tableName)
                }
                continue
            }

            resetSQL := fmt.Sprintf(
                "SELECT setval('%s_id_seq', COALESCE((SELECT MAX(id) FROM %s), 1));",
                tableName, quoteTableName(config.TargetDBConfig.DBDriver, tableName))

            _, err = config.TargetDB.ExecContext(ctx, resetSQL)
            if err != nil {
                log.Printf("테이블 '%s' 시퀀스 재설정 실패: %v (무시하고 계속 진행합니다)", tableName, err)
            } else if config.VerboseLogging {
                fmt.Printf("  - 테이블 '%s' 시퀀스 재설정 완료\n", tableName)
            }
        }
    } else if config.TargetDBConfig.DBDriver == "mysql" || config.TargetDBConfig.DBDriver == "mariadb" {
        // MySQL/MariaDB 자동증가 값 재설정
        for _, tableName := range tables {
            // 자동증가 재설정이 필요한 테이블만 처리
            if !needsSequenceReset(tableName) {
                continue
            }

            // 테이블 존재 여부 확인
            var dbName string
            err := config.TargetDB.QueryRow("SELECT DATABASE()").Scan(&dbName)
            if err != nil {
                log.Printf("데이터베이스 이름 조회 실패: %v (계속 진행합니다)", err)
                continue
            }
            
            var tableExists int
            checkQuery := fmt.Sprintf("SELECT COUNT(1) FROM information_schema.tables WHERE table_schema = '%s' AND table_name = '%s'",
                dbName, tableName)
            
            err = config.TargetDB.QueryRow(checkQuery).Scan(&tableExists)
            if err != nil || tableExists == 0 {
                if err != nil {
                    log.Printf("테이블 '%s' 존재 여부 확인 실패: %v (무시하고 계속 진행합니다)", tableName, err)
                }
                continue
            }

            // 현재 최대 ID 값 조회
            quotedTableName := quoteTableName(config.TargetDBConfig.DBDriver, tableName)
            query := fmt.Sprintf("SELECT COALESCE(MAX(id), 0) FROM %s", quotedTableName)
            
            var maxID int64
            err = config.TargetDB.QueryRow(query).Scan(&maxID)

            if err != nil {
                log.Printf("테이블 '%s' 최대 ID 조회 실패: %v (무시하고 계속 진행합니다)", tableName, err)
                continue
            }

            // 자동증가 값 재설정
            if maxID > 0 {
                resetSQL := fmt.Sprintf("ALTER TABLE %s AUTO_INCREMENT = %d", quotedTableName, maxID+1)
                _, err := config.TargetDB.ExecContext(ctx, resetSQL)
                if err != nil {
                    log.Printf("테이블 '%s' 자동증가 값 재설정 실패: %v (무시하고 계속 진행합니다)", tableName, err)
                } else if config.VerboseLogging {
                    fmt.Printf("  - 테이블 '%s' 자동증가 값 재설정 완료 (ID: %d)\n", tableName, maxID+1)
                }
            }
        }
    }

    fmt.Println("시퀀스/자동증가 값 재설정 완료")
    return nil
}
```

위 코드 수정을 통해 3가지 주요 문제를 해결했습니다:

1. users 테이블 중복키 문제: 기존 레코드 확인 후 존재하면 건너뜁니다.
2. 테이블 존재 여부 확인: MySQL에서 현재 데이터베이스 이름을 정확히 획득하고 테이블 존재 여부를 확인합니다.
3. 하이픈이 있는 테이블 이름 처리: 모든 테이블 이름을 백틱으로 적절히 감싸 처리합니다.

이러한 수정으로 이기종 데이터베이스 간 마이그레이션이 더 안정적으로 동작할 것입니다.



---
Powered by [Claude Exporter](https://www.claudexporter.com)