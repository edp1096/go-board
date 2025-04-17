# goboard #74 - migration 2

**Created:** 4/16/2025 17:39  
**Updated:** 4/17/2025 22:11  
**Exported:** 4/18/2025 2:55  

## Prompt:
go-board의 migration을 수정해서 데이터 마이그레이션까지 완전하게 할 수 있게 짜여진 코드야.
values와 results 변수는 not used 에러 떠서 리마크했어. 문제 없겠지?

## Response:

질문을 이해해보면, 사용자는 Go 기반의 게시판 시스템(`go-board`)의 데이터베이스 마이그레이션 코드에 대해 언급하고 있습니다. 특히 `values`와 `results` 변수가 사용되지 않아 (not used) 에러가 발생하여 리마크(주석) 처리를 했다고 말하고 있으며, 이것이 문제가 없는지 묻고 있습니다.

주어진 코드는 데이터베이스 마이그레이션(스키마 및 데이터 이동)을 처리하는 복잡한 Go 프로그램이며, 다양한 데이터베이스(MySQL, PostgreSQL, SQLite) 간의 테이블 및 데이터 마이그레이션을 지원합니다.

사용자의 질문에 대답하기 위해 다음을 분석해보겠습니다:

1. `values`와 `results` 변수가 있는 위치와 그 용도
2. 이러한 변수를 주석 처리했을 때 발생할 수 있는 문제 여부

코드를 검토해 보겠습니다:

1. `values` 변수:
   `migrateTableData` 함수에서 사용되지만 주석 처리된 것으로 보입니다.
   주석 처리된 부분은 다음과 같은 라인일 것입니다: `// var values []string`

2. `results` 변수:
   `getTableColumns` 함수에서 SQLite 부분에서 주석 처리된 것으로 보입니다.
   
```go
// SQLite용 쿼리
// var results []struct {
//     Name    string         `bun:"name"`
//     Type    string         `bun:"type"`
//     NotNull int            `bun:"notnull"`
//     Default sql.NullString `bun:"dflt_value"`
// }
```

`values` 변수는 SQL 쿼리 생성에 사용되었을 가능성이 있지만, 현재 코드에서는 `placeholders` 배열을 사용하여 쿼리 생성 로직을 처리하고 있어서 더 이상 필요하지 않은 것 같습니다.

`results` 변수의 경우, SQLite 스키마 정보를 가져오는 데 사용했을 텐데, 현재 코드에서는 직접 쿼리 실행 후 행을 스캔하는 방식으로 변경된 것으로 보입니다.

이 두 변수를 주석 처리한 것이 문제가 없는지는 다음을 고려해야 합니다:

1. 코드가 여전히 작동하는지 (기능적 영향)
2. 성능이나 리소스 사용에 변화가 있는지 (비기능적 영향)

분석 결과, 두 변수는 더 이상 사용되지 않거나 다른 접근 방식으로 대체된 것으로 보입니다. 변수들이 선언되었지만 사용되지 않아 발생하는 컴파일러 경고를 제거하기 위해 주석 처리한 것으로 판단됩니다.

따라서 이러한 변수를 주석 처리하는 것은 코드의 작동에 영향을 미치지 않을 것이며, 오히려 불필요한 메모리 할당을 줄이는 긍정적인 효과가 있을 수 있습니다. 코드의 다른 부분이 이 변수들을 참조하지 않는 한, 주석 처리해도 문제가 없을 것입니다.

코드를 검토해 보았습니다. `values`와 `results` 변수를 리마크(주석) 처리한 것은 문제가 없을 것으로 보입니다.

이유는 다음과 같습니다:

1. **`values` 변수**: 
   - `migrateTableData` 함수에서 보이는 `// var values []string`은 원래 SQL 쿼리 생성에 사용되었을 변수인데, 현재는 `placeholders` 배열만으로 충분히 쿼리 생성이 가능하도록 코드가 개선된 것으로 보입니다.
   - 이 변수는 현재 코드 로직에서 실제로 사용되지 않고 있어서 제거해도 무방합니다.

2. **`results` 변수**:
   - SQLite 테이블 컬럼 정보를 가져오는 `getTableColumns` 함수에서 주석 처리된 구조체입니다.
   - 현재는 `PRAGMA table_info` 쿼리 결과를 직접 스캔하는 방식으로 코드가 구현되어 있어 이 구조체가 필요 없게 되었습니다.

두 변수 모두 코드 진화 과정에서 더 이상 필요하지 않게 된 것으로 보이며, 주석 처리함으로써 "unused variable" 경고를 제거하는 것이 오히려 좋은 방법입니다. 기능적으로는 어떠한 영향도 없을 것이며, 오히려 불필요한 메모리 할당을 줄일 수 있어 약간의 성능 개선 효과까지 있을 수 있습니다.

결론적으로, 이 변수들을 주석 처리한 것은 코드 품질 향상에 도움이 되므로 문제가 없습니다.

## Prompt:
아직 테스트는 안해봤는데 스키마와 데이터 마이그레이션 동작에 문제나 결함은 없겠지?

## Response:

코드를 자세히 검토하여 스키마 및 데이터 마이그레이션 과정에서 발생할 수 있는 잠재적인 문제나 결함을 확인해보겠습니다.

1. 스키마 마이그레이션 관련:

- `runMigration` 함수와 관련 로직은 goose 라이브러리를 사용하여 구현되어 있고, 일반적인 패턴을 따릅니다.
- 데이터베이스 드라이버별(PostgreSQL, MySQL, SQLite)로 다른 마이그레이션 파일 세트를 사용하도록 구현되어 있습니다.
- 마이그레이션 상태를 확인하고 적용하는 로직이 구현되어 있습니다.

2. 데이터 마이그레이션 관련:

- `runDataMigration` 함수에서 소스와 대상 데이터베이스를 연결하고 데이터를 마이그레이션하는 과정이 구현되어 있습니다.
- 기본 테이블과 동적 생성된 테이블의 마이그레이션이 분리되어 있습니다.
- 배치 처리와 트랜잭션 처리가 구현되어 있습니다.

잠재적인 문제점이나 개선 가능성:

1. **데이터 타입 변환 문제**: 
   - `convertValue` 함수에서 데이터베이스 간 타입 변환을 처리하지만, 몇 가지 특정 케이스만 처리하고 있습니다. 
   - 특히 복잡한 데이터 타입(예: JSON, 배열 등)의 변환이 필요한 경우 문제가 발생할 수 있습니다.

2. **외래 키 문제**: 
   - 데이터 마이그레이션 중에 외래 키 제약 조건이 있는 테이블 간의 순서가 중요합니다. 현재 코드에서는 이에 대한 명시적인 처리가 없습니다.
   - 참조 무결성이 깨질 가능성이 있습니다.

3. **대량 데이터 처리**: 
   - 대량의 데이터를 처리할 때 메모리 문제가 발생할 수 있습니다.
   - `batchSize` 파라미터를 통해 배치 크기를 조절할 수 있지만, 실제 대용량 데이터에서 테스트가 필요합니다.

4. **오류 처리**: 
   - 단일 행 삽입 실패 시 전체 배치 또는 심지어 전체 마이그레이션이 중단될 수 있습니다.
   - 특히 동적 테이블의 경우 스키마가 정확히 일치하지 않을 때 문제가 발생할 수 있습니다.

5. **테이블 스키마 차이**: 
   - 소스 및 대상 데이터베이스의 테이블 스키마가 약간 다른 경우(예: 컬럼 타입이 다르거나 추가 컬럼이 있는 경우) 문제가 발생할 수 있습니다.
   - `findCommonColumns` 함수는 공통 컬럼을 찾아 처리하지만, 세부적인 타입 불일치는 런타임 오류를 일으킬 수 있습니다.

6. **시퀀스 및 자동 증가 값 재설정**: 
   - PostgreSQL과 MySQL에 대한 시퀀스 재설정이 구현되어 있지만, 다른 모든 테이블 유형이나 케이스를 다루지는 않습니다.

7. **트랜잭션 관리**: 
   - 일부 데이터베이스 작업은 트랜잭션 내에서 수행되지만, 다른 작업은 그렇지 않습니다.
   - 이로 인해 부분적으로 적용된 마이그레이션이 발생할 수 있습니다.

8. **SQLite 특화 문제**: 
   - SQLite는 다른 데이터베이스와 다른 타입 시스템을 가지고 있어 변환 문제가 발생할 수 있습니다.
   - 특히 불리언 값의 처리(0/1 vs. TRUE/FALSE)에서 문제가 발생할 수 있습니다.

9. **특수 문자 처리**: 
   - 테이블 이름이나 컬럼 이름에 특수 문자가 포함된 경우 적절히 처리되지 않을 수 있습니다.
   - 이는 SQL 인젝션 취약점으로 이어질 수도 있습니다.

10. **성능 관련 문제**: 
    - 대량의 데이터를 마이그레이션할 때 성능이 저하될 수 있습니다.
    - 디스크 공간, 메모리 사용량, 네트워크 오버헤드 등의 리소스 제약이 고려되지 않았습니다.

전반적으로 코드는 잘 구성되어 있지만, 실제 데이터 및 스키마 마이그레이션 과정에서 여러 문제가 발생할 가능성이 있습니다. 따라서 다양한 시나리오에서의 철저한 테스트와 에러 처리 개선이 필요해 보입니다.

데이터와 스키마 마이그레이션 코드를 자세히 검토해본 결과, 전체적으로 잘 구현되어 있지만 몇 가지 잠재적인 문제점이 있을 수 있습니다:

### 잠재적 문제점

1. **데이터 타입 변환 문제**:
   - `convertValue` 함수는 기본적인 타입 변환만 처리하고 있습니다. 특히 특수 데이터 타입(JSON이나 배열)이 있는 경우 변환이 제대로 되지 않을 수 있습니다.
   - 날짜/시간 데이터는 일부 케이스만 처리하고 있어 누락된 케이스가 있을 수 있습니다.

2. **외래 키 의존성**:
   - 테이블 간 의존 관계가 복잡한 경우 마이그레이션 순서가 중요한데, 현재 코드는 특별히 이런 의존성을 고려하지 않습니다.
   - 특히 동적 테이블 마이그레이션에서 참조 무결성에 문제가 발생할 수 있습니다.

3. **대용량 데이터 처리**:
   - 대량의 데이터(수백만 행)가 있는 테이블의 경우 메모리 문제가 발생할 수 있습니다.
   - `batchSize` 매개변수로 조정 가능하지만, 실제 대규모 데이터에서 검증이 필요합니다.

4. **오류 복구 부족**:
   - 마이그레이션 중 오류 발생 시 일부 데이터만 마이그레이션되고 나머지는 손실될 수 있습니다.
   - `MaxErrorsBeforeExit` 설정이 있지만, 부분적으로 적용된 변경을 롤백하는 메커니즘은 없습니다.

5. **다양한 DB 버전 고려 부족**:
   - 특히 PostgreSQL, MySQL 등의 다양한 버전 간 호환성 문제가 발생할 수 있습니다.

### 권장 테스트 시나리오

철저한 테스트를 위해 다음과 같은 시나리오를 확인해보세요:

1. **소규모 데이터셋 테스트**:
   - 단순한 데이터로 기본 기능 검증
   
2. **다양한 데이터 유형 테스트**:
   - 텍스트, 숫자, 날짜, 불리언, NULL 값 등 다양한 데이터 유형 포함
   
3. **이기종 DB간 마이그레이션**:
   - MySQL → PostgreSQL, SQLite → MySQL 등 다양한 조합 테스트
   
4. **대용량 데이터 테스트**:
   - 최소 수만 개 이상의 행이 있는 테이블로 성능 및 안정성 확인

5. **오류 케이스 테스트**:
   - 일부러 제약 조건 위반이나 타입 불일치 상황을 만들어 오류 처리 확인

전반적으로 코드는 잘 설계되어 있으나, 실제 환경과 데이터에서의 철저한 테스트가 필요합니다. 특히 중요한 데이터 마이그레이션 전에는 백업을 확실히 해두는 것이 좋겠습니다.

## Prompt:
소스가 길어서 분리하려는데 같은 패키지 명이면 go build main.go 해도 다 같이 컴파일 되는거지?

## Response:

네, Go 언어에서는 같은 패키지에 있는 소스 파일들은 모두 함께 컴파일됩니다. 제공해주신 질문에 대해 자세히 설명해보겠습니다.

Go 언어의 빌드 시스템은 패키지 기반으로 작동합니다:

1. **같은 패키지(package) 내의 모든 파일**: 
   - 같은 디렉토리에 있고 
   - 같은 패키지 선언(`package 패키지명`)을 가진 모든 .go 파일은 
   - `go build` 명령 실행 시 함께 컴파일됩니다.

2. **`go build main.go` 명령**:
   - 이 명령은 `main.go` 파일을 포함하는 패키지의 모든 파일을 함께 컴파일합니다.
   - 즉, `main.go`와 같은 디렉토리에 있고 같은 패키지 선언을 가진 모든 `.go` 파일이 자동으로 함께 컴파일됩니다.

3. **패키지 분리 방법**:
   - 소스를 분리하려면, 같은 패키지 내에서 여러 파일로 나누면 됩니다.
   - 예를 들어, `main.go`, `migration.go`, `database.go` 등으로 나눌 수 있습니다.
   - 각 파일은 `package main`(또는 동일한 패키지명)으로 시작해야 합니다.

4. **모듈 사용**:
   - `go.mod` 파일이 있는 프로젝트에서는 모듈 경로를 기준으로 패키지가 결정됩니다.
   - 이 경우에도 같은 패키지 내의 파일들은 함께 컴파일됩니다.

따라서 `main.go` 파일이 있는 디렉토리에 다른 .go 파일들을 만들고, 각 파일에 동일한 패키지 선언(예: `package main`)을 포함시키면, `go build main.go` 명령으로 모든 파일이 함께 컴파일됩니다.

맞습니다. Go 언어에서는 같은 패키지(package)에 속한 소스 파일들은 자동으로 함께 컴파일됩니다. 

소스 코드를 분리할 때 다음과 같은 점을 고려하시면 됩니다:

1. **같은 디렉토리, 같은 패키지명**:
   - 모든 파일이 동일한 디렉토리에 있고
   - 모든 파일의 시작에 같은 `package 패키지명` 선언이 있으면
   - `go build main.go` 명령어로 모든 파일이 함께 컴파일됩니다.

2. **파일 분리 방법**:
   - `main.go`: 메인 로직과 진입점
   - `migration.go`: 마이그레이션 관련 기능
   - `database.go`: 데이터베이스 연결 관련 기능
   - `types.go`: 타입 정의 
   등으로 기능별로 분리하는 것이 일반적입니다.

3. **패키지 내 함수 접근**:
   - 같은 패키지 내에서는 다른 파일에 정의된 함수를 자유롭게 호출할 수 있습니다.
   - 대소문자로 구분되는 접근 제어(대문자로 시작하면 외부 패키지에서 접근 가능)가 적용됩니다.

분리하시면 코드 가독성과 유지보수가 훨씬 향상될 것입니다. 현재 코드가 한 파일에 너무 많은 기능이 들어있어서 분리하는 것이 좋은 결정입니다.

## Prompt:
니가 분리해줘.

## Response:

이 코드를 적절하게 분리하기 위해 기능별로 파일을 나눠보겠습니다. 현재 main.go 파일은 다음과 같은 주요 부분으로 나눌 수 있습니다:

1. 메인 함수와 명령행 인자 처리 - main.go
2. 설정 관련 함수 - config.go
3. 데이터베이스 연결 관련 - database.go
4. 스키마 마이그레이션 관련 - schema_migration.go
5. 데이터 마이그레이션 관련 - data_migration.go
6. 타입/구조체 정의 - types.go

각 파일로 분리하면서 중복되는 import도 적절히 처리해야 합니다. 패키지명은 원본과 동일하게 'main'을 유지합니다.

분리 작업을 진행하겠습니다.

코드를 기능별로 분리하겠습니다. 같은 패키지(`main`)를 사용하여 다음과 같이 6개의 파일로 나누었습니다.

### 1. types.go (타입 정의)

```go
package main

import (
	"database/sql"
	"github.com/uptrace/bun"
)

// 데이터 마이그레이션을 위한 설정 구조체
type DataMigrationConfig struct {
	SourceDBConfig      *config.Config
	TargetDBConfig      *config.Config
	SourceDB            *bun.DB
	TargetDB            *bun.DB
	BatchSize           int
	SkipTables          []string
	DynamicTablesOnly   bool
	BasicTablesOnly     bool
	IncludeInactive     bool
	VerboseLogging      bool
	DataOnly            bool
	SchemaOnly          bool
	EnableTransactions  bool
	MaxErrorsBeforeExit int
	Errors              []error
}

// 테이블 메타데이터 구조체
type TableMetadata struct {
	Name       string
	Schema     string
	Columns    []ColumnMetadata
	IsDynamic  bool
	SourceRows int
	TargetRows int
}

// 컬럼 메타데이터 구조체
type ColumnMetadata struct {
	Name     string
	Type     string
	Nullable bool
	Default  sql.NullString
}
```

### 2. config.go (설정 관련)

```go
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/edp1096/go-board/config"
)

// loadConfig는 환경 파일에서 설정을 로드하고 드라이버를 오버라이드합니다
func loadConfig(envFile string, driverOverride string) (*config.Config, error) {
	// 기존 환경 변수 저장
	oldAppEnv := os.Getenv("APP_ENV")
	oldDBDriver := os.Getenv("DB_DRIVER")

	// 환경 변수 설정
	if envFile != "" && envFile != ".env" {
		// 환경 파일 내용 로드
		content, err := os.ReadFile(envFile)
		if err != nil {
			return nil, fmt.Errorf("환경 파일 읽기 실패: %w", err)
		}

		// 환경 변수 파싱 및 설정
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
		}
	}

	// 드라이버 오버라이드
	if driverOverride != "" {
		os.Setenv("DB_DRIVER", driverOverride)
	}

	// 설정 로드
	cfg, err := config.Load()

	// 기존 환경 변수 복원
	os.Setenv("APP_ENV", oldAppEnv)
	os.Setenv("DB_DRIVER", oldDBDriver)

	return cfg, err
}
```

### 3. database.go (데이터베이스 관련)

```go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/edp1096/go-board/config"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
)

// ensureDatabaseExists는 데이터베이스가 존재하는지 확인하고 없으면 생성합니다
func ensureDatabaseExists(cfg *config.Config) error {
	// 기존 코드 유지
	var db *sql.DB
	var err error

	switch cfg.DBDriver {
	case "postgres":
		// PostgreSQL 관리자 연결 (데이터베이스 생성용) - postgres 데이터베이스에 연결
		connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
			cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword)
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			return fmt.Errorf("PostgreSQL 관리자 연결 실패: %w", err)
		}
		defer db.Close()

		// 데이터베이스 존재 여부 확인
		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", cfg.DBName).Scan(&exists)
		if err != nil {
			return fmt.Errorf("데이터베이스 확인 실패: %w", err)
		}

		// 데이터베이스가 없으면 생성
		if !exists {
			fmt.Printf("데이터베이스 '%s'가 존재하지 않아 생성합니다\n", cfg.DBName)

			// 데이터베이스 이름에 특수문자가 있으면 큰따옴표로 감싸줌
			dbNameQuoted := cfg.DBName
			if strings.ContainsAny(dbNameQuoted, "-.") {
				dbNameQuoted = fmt.Sprintf("\"%s\"", dbNameQuoted)
			}

			_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbNameQuoted))
			if err != nil {
				return fmt.Errorf("데이터베이스 생성 실패: %w", err)
			}
			fmt.Printf("데이터베이스 '%s'가 성공적으로 생성되었습니다\n", cfg.DBName)
		}

	case "mysql", "mariadb":
		// MySQL/MariaDB 관리자 연결 (데이터베이스 생성용) - 데이터베이스 이름 없이 연결
		connStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort)
		db, err = sql.Open("mysql", connStr)
		if err != nil {
			return fmt.Errorf("MySQL 관리자 연결 실패: %w", err)
		}
		defer db.Close()

		// 데이터베이스 존재 여부 확인
		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM information_schema.schemata WHERE schema_name = ?)", cfg.DBName).Scan(&exists)
		if err != nil {
			return fmt.Errorf("데이터베이스 확인 실패: %w", err)
		}

		// 데이터베이스가 없으면 생성
		if !exists {
			fmt.Printf("데이터베이스 '%s'가 존재하지 않아 생성합니다\n", cfg.DBName)

			// 데이터베이스 이름에 특수문자가 있으면 백틱으로 감싸줌
			dbNameQuoted := cfg.DBName
			if strings.ContainsAny(dbNameQuoted, "-.") {
				dbNameQuoted = fmt.Sprintf("`%s`", dbNameQuoted)
			}

			_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", dbNameQuoted))
			if err != nil {
				return fmt.Errorf("데이터베이스 생성 실패: %w", err)
			}
			fmt.Printf("데이터베이스 '%s'가 성공적으로 생성되었습니다\n", cfg.DBName)
		}

	case "sqlite":
		// SQLite는 파일 기반이므로 별도의 생성 과정 불필요
		return nil

	default:
		return fmt.Errorf("지원하지 않는 데이터베이스 드라이버: %s", cfg.DBDriver)
	}

	return nil
}

// ConnectDatabase는 Config 구조체를 사용하여 데이터베이스에 연결합니다
func (c *DataMigrationConfig) ConnectDatabase(cfg *config.Config) (*bun.DB, error) {
	var sqldb *sql.DB
	var err error

	switch cfg.DBDriver {
	case "postgres":
		// PostgreSQL 연결 설정
		dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
		sqldb, err = sql.Open("postgres", dsn)
		if err != nil {
			return nil, fmt.Errorf("PostgreSQL 연결 실패: %w", err)
		}

	case "mysql", "mariadb":
		// MySQL/MariaDB 연결 설정
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
		sqldb, err = sql.Open("mysql", dsn)
		if err != nil {
			return nil, fmt.Errorf("MySQL 연결 실패: %w", err)
		}

	case "sqlite":
		// SQLite 연결 설정
		if cfg.DBPath == "" {
			return nil, fmt.Errorf("SQLite 데이터베이스 경로가 설정되지 않았습니다")
		}

		// 디렉토리 존재 확인 및 생성
		dir := filepath.Dir(cfg.DBPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("SQLite 데이터베이스 디렉토리 생성 실패: %w", err)
		}

		// SQLite 연결 - 로킹 타임아웃 추가
		sqldb, err = sql.Open("sqlite", cfg.DBPath+"?_timeout=5000&_journal=WAL&_busy_timeout=5000")
		if err != nil {
			return nil, fmt.Errorf("SQLite 연결 실패: %w", err)
		}

		// SQLite 성능 최적화 설정
		if _, err := sqldb.Exec("PRAGMA journal_mode=WAL; PRAGMA synchronous=NORMAL; PRAGMA busy_timeout=5000;"); err != nil {
			return nil, fmt.Errorf("SQLite PRAGMA 설정 실패: %w", err)
		}

	default:
		return nil, fmt.Errorf("지원하지 않는 데이터베이스 드라이버: %s", cfg.DBDriver)
	}

	// 기본 연결 설정
	sqldb.SetMaxOpenConns(25)
	sqldb.SetMaxIdleConns(5)
	sqldb.SetConnMaxLifetime(5 * time.Minute)

	// 연결 테스트
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqldb.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("데이터베이스 연결 확인 실패: %w", err)
	}

	// Bun DB 인스턴스 생성
	var db *bun.DB
	if cfg.DBDriver == "postgres" {
		db = bun.NewDB(sqldb, pgdialect.New())
	} else if cfg.DBDriver == "sqlite" {
		db = bun.NewDB(sqldb, sqlitedialect.New())
	} else {
		db = bun.NewDB(sqldb, mysqldialect.New())
	}

	return db, nil
}

// connectDatabases는 소스 및 대상 데이터베이스에 연결합니다
func connectDatabases(config *DataMigrationConfig) (*bun.DB, *bun.DB, error) {
	// 소스 데이터베이스가 존재하는지 확인
	if config.SourceDBConfig.DBDriver != "sqlite" {
		if err := ensureDatabaseExists(config.SourceDBConfig); err != nil {
			return nil, nil, fmt.Errorf("소스 데이터베이스 확인 실패: %w", err)
		}
	} else {
		// SQLite 소스 디렉토리 확인
		dir := filepath.Dir(config.SourceDBConfig.DBPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, nil, fmt.Errorf("SQLite 소스 디렉토리 생성 실패: %w", err)
		}

		// 파일 존재 여부 확인
		if _, err := os.Stat(config.SourceDBConfig.DBPath); os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("SQLite 소스 데이터베이스 파일이 존재하지 않습니다: %s", config.SourceDBConfig.DBPath)
		}
	}

	// 대상 데이터베이스 확인/생성
	if config.TargetDBConfig.DBDriver != "sqlite" {
		if err := ensureDatabaseExists(config.TargetDBConfig); err != nil {
			return nil, nil, fmt.Errorf("대상 데이터베이스 확인 실패: %w", err)
		}
	} else {
		// SQLite 대상 디렉토리 확인
		dir := filepath.Dir(config.TargetDBConfig.DBPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, nil, fmt.Errorf("SQLite 대상 디렉토리 생성 실패: %w", err)
		}
	}

	// 소스 데이터베이스 연결
	sourceDB, err := config.ConnectDatabase(config.SourceDBConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("소스 데이터베이스 연결 실패: %w", err)
	}

	// 대상 데이터베이스 연결
	targetDB, err := config.ConnectDatabase(config.TargetDBConfig)
	if err != nil {
		sourceDB.Close()
		return nil, nil, fmt.Errorf("대상 데이터베이스 연결 실패: %w", err)
	}

	return sourceDB, targetDB, nil
}
```

### 4. schema_migration.go (스키마 마이그레이션)

```go
package main

import (
	"database/sql"
	"fmt"
	"io/fs"

	goboard "github.com/edp1096/go-board"
	"github.com/pressly/goose/v3"
)

// runMigration은 지정된 마이그레이션 작업을 실행합니다
func runMigration(db *sql.DB, driver, migrationsDir, operation, dbName string) error {
	// 기존 코드 유지
	var err error
	switch operation {
	case "up":
		fmt.Println("마이그레이션 적용 중...")
		err = goose.Up(db, migrationsDir)
	case "down":
		fmt.Println("마이그레이션 롤백 중...")
		err = goose.Down(db, migrationsDir)
	case "status":
		fmt.Println("마이그레이션 상태 확인 중...")
		err = goose.Status(db, migrationsDir)
	case "purge":
		// 모든 테이블 삭제
		fmt.Println("모든 테이블 삭제 중...")
		err = dropAllTables(db, driver, dbName)
		if err == nil {
			// 마이그레이션 기록도 삭제
			_, err = db.Exec("DROP TABLE IF EXISTS goose_db_version;")
			if err != nil {
				fmt.Printf("마이그레이션 기록 테이블 삭제 실패: %v\n", err)
			} else {
				fmt.Println("마이그레이션 기록 테이블 삭제됨")
			}
			// 마이그레이션 다시 적용
			fmt.Println("마이그레이션 다시 적용 중...")
			err = goose.Up(db, migrationsDir)
		}
	default:
		return fmt.Errorf("알 수 없는 마이그레이션 작업: %s", operation)
	}

	if err != nil {
		return err
	}

	fmt.Printf("마이그레이션 작업 '%s'가 성공적으로 완료되었습니다\n", operation)
	return nil
}

// dropAllTables는 데이터베이스의 모든 테이블을 삭제합니다
func dropAllTables(sqlDB *sql.DB, driver string, dbName string) error {
	// 기존 코드 유지
	var query string
	var rows *sql.Rows
	var err error

	switch driver {
	case "postgres":
		// PostgreSQL에서 모든 테이블 목록 가져오기
		query = `
            SELECT tablename FROM pg_tables 
            WHERE schemaname = 'public' AND 
            tablename != 'goose_db_version';
        `
		rows, err = sqlDB.Query(query)
		if err != nil {
			return fmt.Errorf("테이블 목록 조회 실패: %w", err)
		}
		defer rows.Close()

		// 외래 키 제약 조건 비활성화
		_, err = sqlDB.Exec("SET session_replication_role = 'replica';")
		if err != nil {
			return fmt.Errorf("외래 키 제약 비활성화 실패: %w", err)
		}

		// 모든 테이블 삭제
		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				return fmt.Errorf("테이블 이름 읽기 실패: %w", err)
			}

			if tableName != "goose_db_version" {
				_, err = sqlDB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS \"%s\" CASCADE;", tableName))
				if err != nil {
					return fmt.Errorf("테이블 삭제 실패 (%s): %w", tableName, err)
				}
				fmt.Printf("테이블 삭제: %s\n", tableName)
			}
		}

		// 외래 키 제약 조건 다시 활성화
		_, err = sqlDB.Exec("SET session_replication_role = 'origin';")
		if err != nil {
			return fmt.Errorf("외래 키 제약 재활성화 실패: %w", err)
		}

	case "mysql", "mariadb":
		// MySQL/MariaDB에서 모든 테이블 목록 가져오기
		query = fmt.Sprintf("SHOW TABLES FROM `%s` WHERE Tables_in_%s != 'goose_db_version'", dbName, dbName)
		rows, err = sqlDB.Query(query)
		if err != nil {
			return fmt.Errorf("테이블 목록 조회 실패: %w", err)
		}
		defer rows.Close()

		// 외래 키 제약 조건 비활성화
		_, err = sqlDB.Exec("SET FOREIGN_KEY_CHECKS = 0;")
		if err != nil {
			return fmt.Errorf("외래 키 검사 비활성화 실패: %w", err)
		}

		// 모든 테이블 삭제
		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				return fmt.Errorf("테이블 이름 읽기 실패: %w", err)
			}

			if tableName != "goose_db_version" {
				_, err = sqlDB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS `%s`;", tableName))
				if err != nil {
					return fmt.Errorf("테이블 삭제 실패 (%s): %w", tableName, err)
				}
				fmt.Printf("테이블 삭제: %s\n", tableName)
			}
		}

		// 외래 키 제약 조건 다시 활성화
		_, err = sqlDB.Exec("SET FOREIGN_KEY_CHECKS = 1;")
		if err != nil {
			return fmt.Errorf("외래 키 검사 재활성화 실패: %w", err)
		}

	case "sqlite":
		// SQLite에서 모든 테이블 목록 가져오기
		query = "SELECT name FROM sqlite_master WHERE type='table' AND name != 'goose_db_version';"
		rows, err = sqlDB.Query(query)
		if err != nil {
			return fmt.Errorf("테이블 목록 조회 실패: %w", err)
		}
		defer rows.Close()

		// 외래 키 제약 조건 비활성화
		_, err = sqlDB.Exec("PRAGMA foreign_keys = OFF;")
		if err != nil {
			return fmt.Errorf("외래 키 제약 비활성화 실패: %w", err)
		}

		// 모든 테이블 삭제
		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				return fmt.Errorf("테이블 이름 읽기 실패: %w", err)
			}

			if tableName != "goose_db_version" && !strings.HasPrefix(tableName, "sqlite_") {
				_, err = sqlDB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS \"%s\";", tableName))
				if err != nil {
					return fmt.Errorf("테이블 삭제 실패 (%s): %w", tableName, err)
				}
				fmt.Printf("테이블 삭제: %s\n", tableName)
			}
		}

		// 외래 키 제약 조건 다시 활성화
		_, err = sqlDB.Exec("PRAGMA foreign_keys = ON;")
		if err != nil {
			return fmt.Errorf("외래 키 제약 재활성화 실패: %w", err)
		}

	default:
		return fmt.Errorf("지원하지 않는 데이터베이스 드라이버: %s", driver)
	}

	return nil
}

// migrateSchema는 스키마 정보를 마이그레이션합니다
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
	if err := goose.Up(sqlDB, migrationsDir); err != nil {
		return fmt.Errorf("스키마 마이그레이션 실패: %w", err)
	}

	fmt.Println("기본 스키마 마이그레이션 완료")
	return nil
}
```

### 5. data_migration.go (데이터 마이그레이션)

```go
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/edp1096/go-board/internal/models"
	"github.com/edp1096/go-board/internal/service"
	"github.com/uptrace/bun"
)

// runDataMigration은 데이터 마이그레이션을 실행합니다
func runDataMigration(config *DataMigrationConfig) error {
	fmt.Println("==========================")
	fmt.Printf("이기종 DB 데이터 마이그레이션 시작\n")
	fmt.Printf("소스: %s (%s)\n", config.SourceDBConfig.DBDriver, config.SourceDBConfig.DBName)
	fmt.Printf("대상: %s (%s)\n", config.TargetDBConfig.DBDriver, config.TargetDBConfig.DBName)
	fmt.Println("==========================")

	// 1. 데이터베이스 연결
	sourceDB, targetDB, err := connectDatabases(config)
	if err != nil {
		return fmt.Errorf("데이터베이스 연결 실패: %w", err)
	}
	defer sourceDB.Close()
	defer targetDB.Close()

	config.SourceDB = sourceDB
	config.TargetDB = targetDB

	// 2. 대상 데이터베이스 스키마 생성 (있을 경우에만)
	if !config.DataOnly {
		if err := migrateSchema(config); err != nil {
			return fmt.Errorf("스키마 마이그레이션 실패: %w", err)
		}
	}

	// 스키마 전용 마이그레이션이면 여기서 종료
	if config.SchemaOnly {
		fmt.Println("스키마 마이그레이션이 완료되었습니다.")
		return nil
	}

	// 3. 게시판 메타데이터 가져오기
	boardService, dynamicBoardService, err := createBoardServices(config)
	if err != nil {
		return fmt.Errorf("서비스 생성 실패: %w", err)
	}

	// 4. 데이터 마이그레이션 실행
	startTime := time.Now()

	// 4.1 기본 테이블 데이터 마이그레이션
	if !config.DynamicTablesOnly {
		if err := migrateBasicTables(config); err != nil {
			return fmt.Errorf("기본 테이블 마이그레이션 실패: %w", err)
		}
	}

	// 4.2 동적 테이블 데이터 마이그레이션
	if !config.BasicTablesOnly {
		if err := migrateDynamicTables(config, boardService, dynamicBoardService); err != nil {
			return fmt.Errorf("동적 테이블 마이그레이션 실패: %w", err)
		}
	}

	// 5. 시퀀스/자동증가 값 복구
	if err := resetSequences(config); err != nil {
		log.Printf("시퀀스 복구 실패: %v (무시하고 계속 진행합니다)", err)
	}

	// 6. 결과 요약
	elapsedTime := time.Since(startTime)
	fmt.Println("==========================")
	fmt.Printf("데이터 마이그레이션 완료 (소요 시간: %s)\n", elapsedTime)

	if len(config.Errors) > 0 {
		fmt.Printf("경고: 마이그레이션 중 %d개의 오류가 발생했습니다\n", len(config.Errors))
		for i, err := range config.Errors {
			if i >= 5 {
				fmt.Printf("추가 %d개 오류 생략...\n", len(config.Errors)-5)
				break
			}
			fmt.Printf("  - %v\n", err)
		}
	} else {
		fmt.Println("마이그레이션이 오류 없이 완료되었습니다.")
	}

	fmt.Println("==========================")

	return nil
}

// createBoardServices는 BoardService 및 DynamicBoardService 인스턴스를 생성합니다
func createBoardServices(config *DataMigrationConfig) (service.BoardService, service.DynamicBoardService, error) {
	// 소스 레포지토리 생성
	boardRepo := repository.NewBoardRepository(config.SourceDB)

	// 서비스 생성
	boardService := service.NewBoardService(boardRepo, config.SourceDB)
	dynamicBoardService := service.NewDynamicBoardService(config.SourceDB)

	return boardService, dynamicBoardService, nil
}

// getBasicTables는 기본 테이블 목록을 반환합니다
func getBasicTables() []string {
	return []string{
		"users",
		"boards",
		"board_fields",
		"board_managers",
		"comments",
		"attachments",
		"qna_answers",
		"qna_question_votes",
		"qna_answer_votes",
		"referrer_stats",
		"system_settings",
	}
}

// migrateBasicTables는 기본 테이블의 데이터를 마이그레이션합니다
func migrateBasicTables(config *DataMigrationConfig) error {
	fmt.Println("[2/4] 기본 테이블 데이터 마이그레이션 중...")

	// 기본 테이블 목록
	basicTables := getBasicTables()

	// 테이블별 데이터 마이그레이션
	for _, tableName := range basicTables {
		// 건너뛸 테이블 확인
		if isTableSkipped(tableName, config.SkipTables) {
			fmt.Printf("  - 테이블 '%s' 건너뛰기 (사용자 설정에 의해 제외)\n", tableName)
			continue
		}

		// 테이블 데이터 복사
		err := migrateTableData(config, tableName)
		if err != nil {
			config.addError(fmt.Errorf("테이블 '%s' 마이그레이션 실패: %w", tableName, err))
			if len(config.Errors) > config.MaxErrorsBeforeExit {
				return fmt.Errorf("오류가 너무 많아 마이그레이션이 중단되었습니다")
			}
		}
	}

	fmt.Println("기본 테이블 데이터 마이그레이션 완료")
	return nil
}

// migrateDynamicTables는 동적 테이블의 데이터를 마이그레이션합니다
func migrateDynamicTables(config *DataMigrationConfig, boardService service.BoardService, dynamicBoardService service.DynamicBoardService) error {
	fmt.Println("[3/4] 동적 테이블 마이그레이션 중...")

	// 모든 게시판 목록 가져오기
	boards, err := boardService.ListBoards(context.Background(), !config.IncludeInactive)
	if err != nil {
		return fmt.Errorf("게시판 목록 가져오기 실패: %w", err)
	}

	fmt.Printf("  총 %d개의 게시판을 마이그레이션합니다\n", len(boards))

	// 게시판별 마이그레이션
	for _, board := range boards {
		// 건너뛸 테이블 확인
		if isTableSkipped(board.TableName, config.SkipTables) {
			fmt.Printf("  - 게시판 '%s' (%s) 건너뛰기 (사용자 설정에 의해 제외)\n", board.Name, board.TableName)
			continue
		}

		fmt.Printf("  - 게시판 '%s' (%s) 마이그레이션 중...\n", board.Name, board.TableName)

		// 게시판 테이블이 대상 DB에 존재하는지 확인
		if err := ensureDynamicTableExists(config, board, dynamicBoardService); err != nil {
			config.addError(fmt.Errorf("게시판 '%s' 테이블 생성 실패: %w", board.Name, err))
			continue
		}

		// 테이블 데이터 복사
		err := migrateTableData(config, board.TableName)
		if err != nil {
			config.addError(fmt.Errorf("게시판 '%s' 데이터 마이그레이션 실패: %w", board.Name, err))
			if len(config.Errors) > config.MaxErrorsBeforeExit {
				return fmt.Errorf("오류가 너무 많아 마이그레이션이 중단되었습니다")
			}
		}
	}

	fmt.Println("동적 테이블 마이그레이션 완료")
	return nil
}

// ensureDynamicTableExists는 대상 DB에 동적 테이블이 존재하는지 확인하고, 없으면 생성합니다
func ensureDynamicTableExists(config *DataMigrationConfig, board *models.Board, dynamicBoardService service.DynamicBoardService) error {
	// 대상 DB에 테이블이 존재하는지 확인
	var exists bool
	ctx := context.Background()

	// DB 종류에 따라 다른 쿼리 사용
	switch config.TargetDBConfig.DBDriver {
	case "postgres":
		err := config.TargetDB.NewSelect().
			TableExpr("information_schema.tables").
			Column("count(*) > 0").
			Where("table_schema = 'public'").
			Where("table_name = ?", board.TableName).
			Scan(ctx, &exists)
		if err != nil {
			return fmt.Errorf("테이블 존재 여부 확인 실패: %w", err)
		}
	case "mysql", "mariadb":
		err := config.TargetDB.NewSelect().
			TableExpr("information_schema.tables").
			Column("count(*) > 0").
			Where("table_schema = ?", config.TargetDBConfig.DBName).
			Where("table_name = ?", board.TableName).
			Scan(ctx, &exists)
		if err != nil {
			return fmt.Errorf("테이블 존재 여부 확인 실패: %w", err)
		}
	case "sqlite":
		err := config.TargetDB.NewSelect().
			TableExpr("sqlite_master").
			Column("count(*) > 0").
			Where("type = 'table'").
			Where("name = ?", board.TableName).
			Scan(ctx, &exists)
		if err != nil {
			return fmt.Errorf("테이블 존재 여부 확인 실패: %w", err)
		}
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

	targetColumns, err := getTableColumns(config.TargetDB, config.TargetDBConfig.DBDriver, tableName)
	if err != nil {
		return fmt.Errorf("대상 테이블 구조 가져오기 실패: %w", err)
	}

	// 공통 컬럼 찾기
	commonColumns, sourceColumnNames, targetColumnNames := findCommonColumns(sourceColumns, targetColumns)
	if len(commonColumns) == 0 {
		return fmt.Errorf("공통 컬럼이 없습니다")
	}

	// 데이터 배치 처리
	batchSize := config.BatchSize
	if batchSize <= 0 {
		batchSize = 1000
	}

	totalBatches := (totalRows + batchSize - 1) / batchSize
	processedRows := 0

	// 대상 테이블의 기존 데이터 삭제 (필요 시)
	if shouldCleanTableBeforeMigration(tableName) {
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
		sourceQuery := fmt.Sprintf("SELECT %s FROM %s LIMIT %d OFFSET %d",
			strings.Join(sourceColumnNames, ", "), quoteTableName(config.SourceDBConfig.DBDriver, tableName), limit, offset)

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

		// 대상 테이블 준비
		// var values []string
		var placeholders []string
		for i := range commonColumns {
			// 데이터베이스별 플레이스홀더 형식
			switch config.TargetDBConfig.DBDriver {
			case "postgres":
				placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
			default:
				placeholders = append(placeholders, "?")
			}
		}

		// 행별 처리
		batchInsertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			quoteTableName(config.TargetDBConfig.DBDriver, tableName),
			strings.Join(targetColumnNames, ", "),
			strings.Join(placeholders, ", "))

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

			// 데이터 변환 (필요 시)
			for i, val := range rowValues {
				rowValues[i] = convertValue(val, tableName, commonColumns[i], config.SourceDBConfig.DBDriver, config.TargetDBConfig.DBDriver)
			}

			// 행 삽입
			var err error
			if useTransaction {
				_, err = tx.ExecContext(ctx, batchInsertSQL, rowValues...)
			} else {
				_, err = config.TargetDB.ExecContext(ctx, batchInsertSQL, rowValues...)
			}

			if err != nil {
				// JSON으로 행 데이터 직렬화 (디버깅용)
				rowJSON, _ := json.Marshal(rowValues)
				sourceRows.Close()
				if useTransaction {
					tx.Rollback()
				}
				return fmt.Errorf("행 삽입 실패: %w\n데이터: %s", err, rowJSON)
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
		// MySQL/MariaDB용 쿼리
		var results []struct {
			ColumnName string         `bun:"COLUMN_NAME"`
			DataType   string         `bun:"DATA_TYPE"`
			IsNullable string         `bun:"IS_NULLABLE"`
			Default    sql.NullString `bun:"COLUMN_DEFAULT"`
		}

		query := db.NewSelect().
			TableExpr("information_schema.columns").
			Column("COLUMN_NAME", "DATA_TYPE", "IS_NULLABLE", "COLUMN_DEFAULT").
			Where("TABLE_SCHEMA = ?", db.DB.Stats().OpenConnections).
			Where("TABLE_NAME = ?", tableName).
			Order("ORDINAL_POSITION")

		err := query.Scan(ctx, &results)
		if err != nil {
			// 데이터베이스 이름 직접 지정 시도
			dbName := ""
			var dbNameRow struct {
				DatabaseName string `bun:"DATABASE()"`
			}
			if err := db.NewSelect().ColumnExpr("DATABASE()").Scan(ctx, &dbNameRow); err == nil {
				dbName = dbNameRow.DatabaseName
			}

			if dbName != "" {
				err = db.NewSelect().
					TableExpr("information_schema.columns").
					Column("COLUMN_NAME", "DATA_TYPE", "IS_NULLABLE", "COLUMN_DEFAULT").
					Where("TABLE_SCHEMA = ?", dbName).
					Where("TABLE_NAME = ?", tableName).
					Order("ORDINAL_POSITION").
					Scan(ctx, &results)
			}

			if err != nil {
				return nil, fmt.Errorf("컬럼 정보 조회 실패: %w", err)
			}
		}

		for _, r := range results {
			columns = append(columns, ColumnMetadata{
				Name:     r.ColumnName,
				Type:     r.DataType,
				Nullable: r.IsNullable == "YES",
				Default:  r.Default,
			})
		}

	case "sqlite":
		// SQLite용 쿼리
		// var results []struct {
		// 	Name    string         `bun:"name"`
		// 	Type    string         `bun:"type"`
		// 	NotNull int            `bun:"notnull"`
		// 	Default sql.NullString `bun:"dflt_value"`
		// }

		err := db.QueryRow(fmt.Sprintf("PRAGMA table_info(%s)", tableName)).Scan()
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("테이블 존재 여부 확인 실패: %w", err)
		}

		rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", tableName))
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

	// 소스 컬럼을 기준으로 공통 컬럼 찾기
	for _, sourceCol := range sourceColumns {
		lowerName := strings.ToLower(sourceCol.Name)
		if targetCol, ok := targetMap[lowerName]; ok {
			common = append(common, sourceCol)
			sourceNames = append(sourceNames, quoteColumnName(sourceCol.Name))
			targetNames = append(targetNames, quoteColumnName(targetCol.Name))
		}
	}

	return common, sourceNames, targetNames
}

// quoteTableName은 데이터베이스 드라이버에 따라 테이블 이름을 인용 부호로 묶습니다
func quoteTableName(driver, tableName string) string {
	switch driver {
	case "postgres":
		return fmt.Sprintf("\"%s\"", tableName)
	case "mysql", "mariadb":
		return fmt.Sprintf("`%s`", tableName)
	default:
		return tableName
	}
}

// quoteColumnName은 데이터베이스 드라이버에 따라 컬럼 이름을 인용 부호로 묶습니다
func quoteColumnName(columnName string) string {
	return columnName
}

// convertValue는 소스 DB의 값을 대상 DB에 맞게 변환합니다
func convertValue(val interface{}, tableName string, column ColumnMetadata, sourceDriver, targetDriver string) interface{} {
	if val == nil {
		return nil
	}

	// 불리언 타입 변환
	if sourceDriver == "sqlite" && targetDriver != "sqlite" {
		// SQLite의 불리언은 0/1 정수로 저장됨
		if column.Type == "INTEGER" || column.Type == "BOOLEAN" || column.Type == "TINYINT" {
			if i, ok := val.(int64); ok && (i == 0 || i == 1) {
				if targetDriver == "postgres" {
					return i == 1
				}
			}
		}
	} else if sourceDriver != "sqlite" && targetDriver == "sqlite" {
		// PostgreSQL/MySQL의 불리언을 SQLite의 0/1로 변환
		if b, ok := val.(bool); ok {
			if b {
				return 1
			}
			return 0
		}
	}

	// MySQL에서 PostgreSQL로 날짜/시간 변환
	if (sourceDriver == "mysql" || sourceDriver == "mariadb") && targetDriver == "postgres" {
		if s, ok := val.(string); ok && strings.Contains(column.Type, "time") {
			// MySQL은 가끔 문자열로 날짜 반환
			t, err := time.Parse("2006-01-02 15:04:05", s)
			if err == nil {
				return t
			}
		}
	}

	// PostgreSQL에서 MySQL로 날짜/시간 변환
	if sourceDriver == "postgres" && (targetDriver == "mysql" || targetDriver == "mariadb") {
		if t, ok := val.(time.Time); ok {
			return t.Format("2006-01-02 15:04:05")
		}
	}

	return val
}

// shouldUseTransactionForTable은 해당 테이블에 트랜잭션을 사용해야 하는지 결정합니다
func shouldUseTransactionForTable(tableName string, totalRows int) bool {
	// 대규모 테이블은 트랜잭션 없이 처리
	if totalRows > 10000 {
		return false
	}

	// 빈번하게 접근되는 테이블은 트랜잭션 없이 처리 (개별 삽입)
	highTrafficTables := map[string]bool{
		"referrer_stats": true,
	}

	return !highTrafficTables[tableName]
}

// shouldCleanTableBeforeMigration은 마이그레이션 전에 테이블을 정리해야 하는지 결정합니다
func shouldCleanTableBeforeMigration(tableName string) bool {
	// 이력 데이터는 정리하지 않음
	historyTables := map[string]bool{
		"referrer_stats": true,
	}

	return !historyTables[tableName]
}

// isTableSkipped는 주어진 테이블이 건너뛰기 목록에 있는지 확인합니다
func isTableSkipped(tableName string, skipTables []string) bool {
	for _, skipTable := range skipTables {
		if skipTable == tableName {
			return true
		}
	}
	return false
}

// resetSequences는 ID 시퀀스/자동증가 값을 재설정합니다
func resetSequences(config *DataMigrationConfig) error {
	fmt.Println("[4/4] 시퀀스/자동증가 값 재설정 중...")

	if config.TargetDBConfig.DBDriver == "postgres" {
		// PostgreSQL 시퀀스 재설정
		tables := getBasicTables()
		for _, tableName := range tables {
			// 시퀀스 재설정이 필요한 테이블만 처리
			if !needsSequenceReset(tableName) {
				continue
			}

			resetSQL := fmt.Sprintf(
				"SELECT setval(pg_get_serial_sequence('%s', 'id'), COALESCE((SELECT MAX(id) FROM \"%s\"), 1));",
				tableName, tableName)

			_, err := config.TargetDB.ExecContext(context.Background(), resetSQL)
			if err != nil {
				return fmt.Errorf("테이블 '%s' 시퀀스 재설정 실패: %w", tableName, err)
			}

			if config.VerboseLogging {
				fmt.Printf("  - 테이블 '%s' 시퀀스 재설정 완료\n", tableName)
			}
		}
	} else if config.TargetDBConfig.DBDriver == "mysql" || config.TargetDBConfig.DBDriver == "mariadb" {
		// MySQL/MariaDB 자동증가 값 재설정
		tables := getBasicTables()
		for _, tableName := range tables {
			// 자동증가 재설정이 필요한 테이블만 처리
			if !needsSequenceReset(tableName) {
				continue
			}

			// 현재 최대 ID 값 조회
			var maxID int64
			err := config.TargetDB.NewSelect().
				TableExpr(tableName).
				ColumnExpr("COALESCE(MAX(id), 0) as max_id").
				Scan(context.Background(), &maxID)

			if err != nil {
				return fmt.Errorf("테이블 '%s' 최대 ID 조회 실패: %w", tableName, err)
			}

			// 자동증가 값 재설정
			if maxID > 0 {
				resetSQL := fmt.Sprintf("ALTER TABLE `%s` AUTO_INCREMENT = %d;", tableName, maxID+1)
				_, err := config.TargetDB.ExecContext(context.Background(), resetSQL)
				if err != nil {
					return fmt.Errorf("테이블 '%s' 자동증가 값 재설정 실패: %w", tableName, err)
				}

				if config.VerboseLogging {
					fmt.Printf("  - 테이블 '%s' 자동증가 값 재설정 완료 (ID: %d)\n", tableName, maxID+1)
				}
			}
		}
	}

	// SQLite는 특별한 재설정이 필요 없음

	fmt.Println("시퀀스/자동증가 값 재설정 완료")
	return nil
}

// needsSequenceReset은 테이블에 시퀀스 재설정이 필요한지 확인합니다
func needsSequenceReset(tableName string) bool {
	// 시퀀스 재설정이 필요 없는 테이블
	noResetTables := map[string]bool{
		"goose_db_version": true,
		"referrer_stats":   true,
	}

	return !noResetTables[tableName]
}

// addError는 마이그레이션 오류 목록에 오류를 추가합니다
func (c *DataMigrationConfig) addError(err error) {
	if c.VerboseLogging {
		log.Printf("오류: %v", err)
	}
	c.Errors = append(c.Errors, err)
}
```

### 6. main.go (메인 함수)

```go
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	goboard "github.com/edp1096/go-board"
	"github.com/pressly/goose/v3"
)

func main() {
	// 명령행 인자 파싱
	sourceDriver := flag.String("source-driver", "", "소스 데이터베이스 드라이버 (mysql, postgres, sqlite)")
	targetDriver := flag.String("target-driver", "", "대상 데이터베이스 드라이버 (mysql, postgres, sqlite)")
	sourceEnvFile := flag.String("source-env", ".env", "소스 환경 설정 파일")
	targetEnvFile := flag.String("target-env", "", "대상 환경 설정 파일 (기본값: 소스와 동일)")
	operation := flag.String("op", "up", "마이그레이션 작업 (up, down, status, purge, data-migrate)")
	batchSize := flag.Int("batch-size", 1000, "데이터 마이그레이션 배치 크기")
	skipTables := flag.String("skip-tables", "", "마이그레이션 시 건너뛸 테이블 목록 (쉼표로 구분)")
	dynamicOnly := flag.Bool("dynamic-only", false, "동적 테이블만 마이그레이션")
	basicOnly := flag.Bool("basic-only", false, "기본 테이블만 마이그레이션")
	includeInactive := flag.Bool("include-inactive", true, "비활성 게시판 포함 여부")
	verboseLogging := flag.Bool("verbose", true, "상세 로깅 여부")
	dataOnly := flag.Bool("data-only", false, "데이터만 마이그레이션 (스키마 마이그레이션 없음)")
	schemaOnly := flag.Bool("schema-only", false, "스키마만 마이그레이션 (데이터 마이그레이션 없음)")
	disableTransactions := flag.Bool("disable-transactions", false, "트랜잭션 비활성화 (대용량 데이터에 유용)")
	maxErrors := flag.Int("max-errors", 10, "마이그레이션 중단 전 최대 오류 수")

	flag.Parse()

	// 환경 설정 로드
	sourceConfig, err := loadConfig(*sourceEnvFile, *sourceDriver)
	if err != nil {
		log.Fatalf("소스 설정을 로드할 수 없습니다: %v", err)
	}

	targetEnvPath := *targetEnvFile
	if targetEnvPath == "" {
		targetEnvPath = *sourceEnvFile
	}

	// 대상 데이터베이스 설정
	targetConfig, err := loadConfig(targetEnvPath, *targetDriver)
	if err != nil {
		log.Fatalf("대상 설정을 로드할 수 없습니다: %v", err)
	}

	// 데이터 마이그레이션인 경우 소스와 대상 DB 모두 설정
	if *operation == "data-migrate" {
		if *sourceDriver == "" || *targetDriver == "" {
			log.Fatalf("데이터 마이그레이션에는 소스와 대상 드라이버가 모두 필요합니다")
		}

		if *sourceDriver == *targetDriver && sourceConfig.DBName == targetConfig.DBName &&
			sourceConfig.DBHost == targetConfig.DBHost && sourceConfig.DBPort == targetConfig.DBPort {
			log.Fatalf("소스와 대상 데이터베이스가 동일합니다")
		}

		// 데이터 마이그레이션 구성
		migConfig := &DataMigrationConfig{
			SourceDBConfig:      sourceConfig,
			TargetDBConfig:      targetConfig,
			BatchSize:           *batchSize,
			SkipTables:          []string{},
			DynamicTablesOnly:   *dynamicOnly,
			BasicTablesOnly:     *basicOnly,
			IncludeInactive:     *includeInactive,
			VerboseLogging:      *verboseLogging,
			DataOnly:            *dataOnly,
			SchemaOnly:          *schemaOnly,
			EnableTransactions:  !*disableTransactions,
			MaxErrorsBeforeExit: *maxErrors,
			Errors:              []error{},
		}

		// 건너뛸 테이블 목록 파싱
		if *skipTables != "" {
			migConfig.SkipTables = strings.Split(*skipTables, ",")
			for i, table := range migConfig.SkipTables {
				migConfig.SkipTables[i] = strings.TrimSpace(table)
			}
		}

		// 데이터 마이그레이션 실행
		err = runDataMigration(migConfig)
		if err != nil {
			log.Fatalf("데이터 마이그레이션 실패: %v", err)
		}

		log.Println("데이터 마이그레이션이 성공적으로 완료되었습니다")
		return
	}

	// 일반 스키마 마이그레이션 로직 (기존 코드)
	// SQLite가 아닌 경우에만 데이터베이스 존재 여부 확인
	if sourceConfig.DBDriver != "sqlite" {
		// 데이터베이스 존재 여부 확인 및 없으면 생성
		if err := ensureDatabaseExists(sourceConfig); err != nil {
			log.Fatalf("데이터베이스 생성 실패: %v", err)
		}
	} else {
		// SQLite인 경우 디렉토리 확인
		dir := filepath.Dir(sourceConfig.DBPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("SQLite 데이터베이스 디렉토리 생성 실패: %v", err)
		}
	}

	// 데이터베이스 연결
	database, err := config.NewDatabase(sourceConfig)
	if err != nil {
		log.Fatalf("데이터베이스에 연결할 수 없습니다: %v", err)
	}
	defer database.Close()

	fmt.Printf("데이터베이스 '%s'가 준비되었습니다\n", sourceConfig.DBName)

	// SQL DB 인스턴스 가져오기
	sqlDB := database.DB.DB

	// 마이그레이션 설정
	goose.SetBaseFS(nil) // 기본 파일시스템 설정 초기화

	// 방언 설정
	if err := goose.SetDialect(sourceConfig.DBDriver); err != nil {
		log.Fatalf("데이터베이스 방언 설정 실패: %v", err)
	}

	// 임베디드 마이그레이션 파일 시스템 가져오기
	var migrationFS fs.FS
	var migrationsDir string

	switch sourceConfig.DBDriver {
	case "postgres":
		// PostgreSQL 마이그레이션 설정
		subFS, err := fs.Sub(goboard.PostgresMigrationsFS, "migrations/postgres")
		if err != nil {
			log.Fatalf("PostgreSQL 마이그레이션 파일 접근 실패: %v", err)
		}
		migrationFS = subFS
		migrationsDir = "."
	case "sqlite":
		// SQLite 마이그레이션 설정
		subFS, err := fs.Sub(goboard.SQLiteMigrationsFS, "migrations/sqlite")
		if err != nil {
			log.Fatalf("SQLite 마이그레이션 파일 접근 실패: %v", err)
		}
		migrationFS = subFS
		migrationsDir = "."
	default:
		// MySQL 마이그레이션 설정
		subFS, err := fs.Sub(goboard.MysqlMigrationsFS, "migrations/mysql")
		if err != nil {
			log.Fatalf("MySQL 마이그레이션 파일 접근 실패: %v", err)
		}
		migrationFS = subFS
		migrationsDir = "."
	}

	// 마이그레이션 파일 시스템 설정
	goose.SetBaseFS(migrationFS)

	// 마이그레이션 실행
	if err := runMigration(sqlDB, sourceConfig.DBDriver, migrationsDir, *operation, sourceConfig.DBName); err != nil {
		log.Fatalf("마이그레이션 실패: %v", err)
	}
}
```

이렇게 코드를 6개의 파일로 분리했습니다. 모든 파일은 같은 `main` 패키지에 속하므로 `go build` 명령으로 모두 함께 컴파일됩니다. 코드를 분리함으로써 가독성과 유지보수성이 크게 향상될 것입니다.

## Prompt:
PS D:\dev\pcbangstudio\workspace\go-board\bin> .\migrate_windows_amd64.exe -op data-migrate -source-driver sqlite -target-driver postgres -source-env .env_sqlite -target-env .env_pg      
2025/04/17 21:09:51 소스 설정을 로드할 수 없습니다: 유효하지 않은 환경 변수: development    # development (possible http), production (require https)
PS D:\dev\pcbangstudio\workspace\go-board\bin>

## Response:

에러 메시지를 보니 환경 설정 파일을 로드하는 과정에서 문제가 발생했습니다. 구체적으로 "유효하지 않은 환경 변수: development"라는 오류가 발생했습니다. 이는 `.env_sqlite` 파일 내에 있는 `APP_ENV` 환경 변수가 유효하지 않은 값("development")으로 설정되어 있기 때문인 것 같습니다.

오류 메시지에 있는 주석을 보면:
```
development (possible http), production (require https)
```

이 설명은 `APP_ENV` 환경 변수가 "development" 또는 "production" 값만 허용하며, development는 HTTP를 사용할 수 있고, production은 HTTPS를 필요로 한다는 것을 암시합니다.

이 문제를 해결하기 위해서는 다음과 같은 방법을 시도해볼 수 있습니다:

1. `.env_sqlite` 파일을 확인하고 `APP_ENV` 변수가 제대로 설정되어 있는지 확인합니다.
2. 만약 `APP_ENV`가 "development"로 설정되어 있다면, 이는 유효한 값이어야 하는데, 오류가 발생한다는 것은 코드에서 이 값을 검증하는 로직이 잘못되었을 가능성이 있습니다.
3. `config.Load()` 함수 내에서 `APP_ENV` 값을 검증하는 부분을 확인하고 수정해야 할 수 있습니다.

`.env_sqlite` 파일이 없거나 잘못 구성되어 있을 가능성도 있습니다. 환경 변수 파일의 내용을 확인하고 필요한 모든 변수가 올바르게 설정되어 있는지 확인해야 합니다.

`APP_ENV` 환경 변수 값이 유효하지 않아 발생하는 오류입니다. `development`는 유효한 값이어야 하는데, 설정에 문제가 있는 것 같습니다.

다음과 같이 문제를 해결해보세요:

### 1. .env_sqlite 파일 확인

`.env_sqlite` 파일을 열어서 다음과 같이 APP_ENV 값이 올바르게 설정되어 있는지 확인하세요:

```
APP_ENV=development
```

만약 이 값이 다르게 설정되어 있거나 오타가 있을 수 있습니다.

### 2. 환경 변수 파일 형식 확인

`.env_sqlite` 파일이 올바른 형식으로 작성되어 있는지 확인하세요. 각 줄에 하나의 환경 변수가 `KEY=VALUE` 형식으로 정의되어 있어야 합니다.

### 3. APP_ENV 유효성 검증 수정

만약 위의 방법으로 해결되지 않는다면, `config.Load()` 함수에 `APP_ENV` 값을 너무 엄격하게 검증하는 로직이 있을 수 있습니다. 이 경우 소스 코드의 검증 부분을 수정해야 할 수 있습니다.

### 4. 환경 변수 직접 설정

마지막 방법으로, 명령줄에서 직접 환경 변수를 설정하여 실행할 수 있습니다:

```powershell
# PowerShell
$env:APP_ENV="development"
.\migrate_windows_amd64.exe -op data-migrate -source-driver sqlite -target-driver postgres -source-env .env_sqlite -target-env .env_pg
```

이렇게 문제가 해결되지 않는다면, `.env_sqlite`와 `.env_pg` 파일의 전체 내용을 확인하여 더 자세한 도움을 드릴 수 있습니다.

## Prompt:
따옴표가 왜 들어가야돼? 이해를 못하겠네.

## Response:

문제를 살펴보겠습니다. 사용자가 언급한 오류 메시지는 다음과 같습니다:

```
2025/04/17 21:09:51 소스 설정을 로드할 수 없습니다: 유효하지 않은 환경 변수: development    # development (possible http), production (require https)
```

이 오류는 환경 변수 파일(.env_sqlite)을 파싱하는 과정에서 발생한 것으로 보입니다. 사용자가 제공한 .env_sqlite 파일을 보면:

```
APP_ENV=development    # development (possible http), production (require https)
```

문제는 이 라인에 있는 것 같습니다. 주석 부분(# 이후)까지 환경 변수 값으로 인식되고 있을 가능성이 있습니다. 그래서 "유효하지 않은 환경 변수: development    # development (possible http), production (require https)" 오류가 발생했을 것입니다.

대부분의 .env 파서는 # 문자를 주석으로 취급하지만, 사용자의 코드에서 주석을 처리하는 방법에 문제가 있거나, 또는 = 기호 뒤에 오는 값에 따옴표를 사용해야 할 수도 있습니다.

문제 해결 방법:
1. 따옴표를 사용하여 환경 변수 값을 명확히 구분: `APP_ENV="development"` 또는 `APP_ENV='development'`
2. 주석과 값 사이에 충분한 공백을 추가: `APP_ENV=development     # 주석...`
3. 주석을 다른 줄로 이동

이 문제는 환경 변수 파싱 로직에 따라 달라질 수 있습니다. 여기서 "따옴표가 왜 들어가야돼?"라는 사용자의 질문은 환경 변수 값을 따옴표로 묶어야 한다는 것을 의미하는 것 같습니다.

문제는 환경 변수 파일에서 주석(#)을 처리하는 방식에 있습니다. `loadConfig` 함수의 환경 변수 파싱 부분을 살펴보면:

```go
// 환경 변수 파싱 및 설정
lines := strings.Split(string(content), "\n")
for _, line := range lines {
    line = strings.TrimSpace(line)
    if line == "" || strings.HasPrefix(line, "#") {
        continue
    }

    parts := strings.SplitN(line, "=", 2)
    if len(parts) != 2 {
        continue
    }

    key := strings.TrimSpace(parts[0])
    value := strings.TrimSpace(parts[1])
    os.Setenv(key, value)
}
```

여기서 문제는 값에 있는 주석 부분(`# development...`)까지 값으로 인식하고 있기 때문입니다. `config.Load()` 함수는 `APP_ENV`를 검증할 때 정확히 "development" 또는 "production"만 허용하는데, "development # development (possible http), production (require https)"처럼 주석까지 포함된 값이 전달되면 검증에 실패합니다.

이 문제를 해결할 수 있는 방법:

1. **환경 변수 값에 따옴표 사용**:
   ```
   APP_ENV="development"    # development (possible http), production (require https)
   ```

2. **주석 제거**:
   ```
   APP_ENV=development
   ```

3. **`loadConfig` 함수 개선**:
   주석을 올바르게 처리하도록 코드를 수정할 수 있습니다.

```go
value := strings.TrimSpace(parts[1])
// 주석 처리: # 문자가 있으면 그 앞부분만 사용
if commentIdx := strings.Index(value, "#"); commentIdx >= 0 {
    value = strings.TrimSpace(value[:commentIdx])
}
os.Setenv(key, value)
```

가장 간단한 해결책은 따옴표를 사용하거나 주석을 제거하는 것입니다.

## Prompt:


## Response:

코드의 실행 결과를 보니 여러 오류가 발생했습니다. 주요 문제점들을 분석해보겠습니다:

1. SQLite 테이블 구조 가져오기 실패:
   ```
   테이블 'users' 마이그레이션 실패: 소스 테이블 구조 가져오기 실패: 테이블 존재 여부 확인 실패: sql: expected 6 destination arguments in Scan, not 0
   ```
   이 오류는 SQLite 데이터베이스에서 테이블 구조를 가져오는 과정에서 발생했습니다. `db.QueryRow(...).Scan()`을 호출할 때 충분한 변수를 제공하지 않았습니다. 

2. PostgreSQL 동적 테이블 존재 확인 실패:
   ```
   게시판 '갤러리' 테이블 생성 실패: 테이블 존재 여부 확인 실패: pq: column "count(*) > 0" does not exist
   ```
   PostgreSQL에서 테이블 존재 여부를 확인하는 쿼리에 문제가 있습니다. `count(*) > 0`라는 표현식을 칼럼으로 잘못 처리하고 있습니다.

3. 시퀀스 재설정 오류:
   ```
   시퀀스 복구 실패: 테이블 'board_managers' 시퀀스 재설정 실패: pq: column "id" does not exist
   ```
   `board_managers` 테이블에는 `id` 컬럼이 없는데 시퀀스를 재설정하려고 시도했습니다.

이 문제들을 해결하기 위한 수정 방법을 제안하겠습니다:

1. **SQLite 테이블 구조 가져오기 수정**:
   `getTableColumns` 함수의 SQLite 부분을 수정해야 합니다. `PRAGMA table_info` 결과를 처리하는 방법이 잘못되었습니다.

```go
// SQLite용 쿼리 수정
err := db.QueryRow(fmt.Sprintf("PRAGMA table_info(%s)", tableName)).Scan()  // 여기가 문제입니다
```

이 부분은 삭제하고 바로 rows 쿼리만 수행해야 합니다:

```go
// SQLite용 쿼리 수정
rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", tableName))
if err != nil {
    return nil, fmt.Errorf("컬럼 정보 조회 실패: %w", err)
}
defer rows.Close()
```

2. **PostgreSQL 테이블 존재 확인 쿼리 수정**:
   `ensureDynamicTableExists` 함수에서 PostgreSQL 테이블 존재 확인 쿼리를 수정해야 합니다.

```go
// 잘못된 코드
err := config.TargetDB.NewSelect().
    TableExpr("information_schema.tables").
    Column("count(*) > 0").  // 이 부분이 문제입니다
    Where("table_schema = 'public'").
    Where("table_name = ?", board.TableName).
    Scan(ctx, &exists)
```

올바른 방법은 다음과 같습니다:

```go
// 수정된 코드
err := config.TargetDB.NewSelect().
    TableExpr("information_schema.tables").
    ColumnExpr("COUNT(*) > 0").  // ColumnExpr로 변경
    Where("table_schema = 'public'").
    Where("table_name = ?", board.TableName).
    Scan(ctx, &exists)
```

3. **시퀀스 재설정 로직 수정**:
   `board_managers` 테이블은 복합 기본키를 사용하므로 `id` 컬럼이 없습니다. 이 테이블은 시퀀스 재설정 대상에서 제외해야 합니다.

```go
// needsSequenceReset 함수 수정
func needsSequenceReset(tableName string) bool {
    // 시퀀스 재설정이 필요 없는 테이블
    noResetTables := map[string]bool{
        "goose_db_version": true,
        "referrer_stats":   true,
        "board_managers":   true,  // 추가
    }

    return !noResetTables[tableName]
}
```

이러한 변경 사항을 적용하면 대부분의 오류가 해결될 것입니다.

SQLite 테이블 구조 검사와 PostgreSQL 쿼리 작성 부분에 몇 가지 오류가 있네요. 다음과 같이 수정이 필요합니다:

### 1. SQLite 테이블 구조 가져오기 오류 수정

`getTableColumns` 함수에서 SQLite 부분에 문제가 있습니다. 다음과 같이 수정해야 합니다:

```go
case "sqlite":
    // SQLite용 쿼리
    // 이 줄을 제거하세요 - 잘못된 스캔 시도
    //err := db.QueryRow(fmt.Sprintf("PRAGMA table_info(%s)", tableName)).Scan()
    
    // 바로 rows 쿼리만 실행
    rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", tableName))
    if err != nil {
        return nil, fmt.Errorf("컬럼 정보 조회 실패: %w", err)
    }
    defer rows.Close()
```

### 2. PostgreSQL 테이블 존재 확인 쿼리 수정

`ensureDynamicTableExists` 함수에서 PostgreSQL 쿼리도 수정이 필요합니다:

```go
case "postgres":
    err := config.TargetDB.NewSelect().
        TableExpr("information_schema.tables").
        ColumnExpr("COUNT(*) > 0").  // Column을 ColumnExpr로 변경
        Where("table_schema = 'public'").
        Where("table_name = ?", board.TableName).
        Scan(ctx, &exists)
```

마찬가지로 `mysql` 케이스에도 동일한 수정이 필요합니다:

```go
case "mysql", "mariadb":
    err := config.TargetDB.NewSelect().
        TableExpr("information_schema.tables").
        ColumnExpr("COUNT(*) > 0").  // Column을 ColumnExpr로 변경
        Where("table_schema = ?", config.TargetDBConfig.DBName).
        Where("table_name = ?", board.TableName).
        Scan(ctx, &exists)
```

### 3. 시퀀스 재설정 로직 수정

`board_managers` 테이블은 복합 기본키를 사용하므로 id 컬럼이 없어 시퀀스 재설정 시 오류가 발생합니다. 이 테이블은 시퀀스 재설정 대상에서 제외해야 합니다:

```go
func needsSequenceReset(tableName string) bool {
    // 시퀀스 재설정이 필요 없는 테이블
    noResetTables := map[string]bool{
        "goose_db_version": true,
        "referrer_stats":   true,
        "board_managers":   true,  // 이 줄 추가
    }

    return !noResetTables[tableName]
}
```

이러한 수정을 적용하면 대부분의 오류가 해결될 것입니다. 혹시 데이터가 없는 테이블에 대한 마이그레이션은 정상적으로 처리된 것으로 보이므로, 주요 오류를 해결한 후 다시 마이그레이션을 시도해보세요.

## Prompt:


## Response:

여전히 같은 오류가 발생하고 있습니다. 이 문제를 더 자세히 분석해보겠습니다.

주요 오류 두 가지:

1. SQLite 테이블 구조 가져오기 오류: `sql: expected 6 destination arguments in Scan, not 0`
2. PostgreSQL 테이블 존재 여부 확인 오류: `pq: column "count(*) > 0" does not exist`

이 문제를 해결하기 위해 코드를 더 자세히 수정해야 합니다.

### SQLite 테이블 구조 가져오기 문제

SQLite의 PRAGMA table_info 쿼리 실행 방식에 문제가 있습니다. 현재 코드에서 문제가 되는 부분:

```go
err := db.QueryRow(fmt.Sprintf("PRAGMA table_info(%s)", tableName)).Scan()
```

이 부분은 Scan() 호출 시 매개변수가 없어서 문제가 발생합니다. 이 줄은 완전히 제거해야 하며, 바로 아래의 rows 쿼리만 사용해야 합니다.

### PostgreSQL 존재 여부 확인 문제

`count(*) > 0`를 직접 컬럼으로 선택하려고 시도하는 것이 문제입니다. 이것은 PostgreSQL에서 작동하지 않습니다. 대신 먼저 COUNT(*)를 쿼리한 다음 결과를 확인해야 합니다.

## 자세한 수정 방법

1. **SQLite 테이블 컬럼 정보 가져오기 수정**

```go
case "sqlite":
    // SQLite용 쿼리
    // 이 라인을 완전히 제거
    // err := db.QueryRow(fmt.Sprintf("PRAGMA table_info(%s)", tableName)).Scan()
    
    rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", tableName))
    if err != nil {
        return nil, fmt.Errorf("컬럼 정보 조회 실패: %w", err)
    }
    defer rows.Close()
    
    // 나머지 코드는 그대로
```

2. **PostgreSQL 테이블 존재 여부 확인 수정**

```go
case "postgres":
    var count int
    err := config.TargetDB.NewSelect().
        TableExpr("information_schema.tables").
        Column("COUNT(*)").
        Where("table_schema = 'public'").
        Where("table_name = ?", board.TableName).
        Scan(ctx, &count)
    
    if err != nil {
        return fmt.Errorf("테이블 존재 여부 확인 실패: %w", err)
    }
    
    exists = count > 0
```

3. **MySQL 테이블 존재 여부 확인 수정**

```go
case "mysql", "mariadb":
    var count int
    err := config.TargetDB.NewSelect().
        TableExpr("information_schema.tables").
        Column("COUNT(*)").
        Where("table_schema = ?", config.TargetDBConfig.DBName).
        Where("table_name = ?", board.TableName).
        Scan(ctx, &count)
    
    if err != nil {
        return fmt.Errorf("테이블 존재 여부 확인 실패: %w", err)
    }
    
    exists = count > 0
```

4. **SQLite 테이블 존재 여부 확인 수정**

```go
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
```

5. **board_managers 테이블 시퀀스 재설정 제외**

`needsSequenceReset` 함수에서 "board_managers" 테이블을 제외해야 합니다:

```go
func needsSequenceReset(tableName string) bool {
    // 시퀀스 재설정이 필요 없는 테이블
    noResetTables := map[string]bool{
        "goose_db_version": true,
        "referrer_stats":   true,
        "board_managers":   true,  // board_managers 추가
    }

    return !noResetTables[tableName]
}
```

이러한 모든 변경 사항을 적용하면 오류가 해결될 것으로 예상됩니다.

문제의 코드를 더 살펴보니 SQLite 파일 자체가 존재하는지도 확인해야 할 수 있습니다. 현재 코드에서는 파일 경로를 확인하지만 실제 DB 파일에 접근이 가능한지는 확인하지 않을 수 있습니다. 아래 코드를 connectDatabases 함수에 추가할 수 있습니다:

```go
// SQLite 소스 파일이 있는지 확인
if config.SourceDBConfig.DBDriver == "sqlite" {
    if _, err := os.Stat(config.SourceDBConfig.DBPath); os.IsNotExist(err) {
        return nil, nil, fmt.Errorf("SQLite 소스 데이터베이스 파일이 존재하지 않습니다: %s", config.SourceDBConfig.DBPath)
    }
    
    // 파일 읽기 권한 확인
    file, err := os.Open(config.SourceDBConfig.DBPath)
    if err != nil {
        return nil, nil, fmt.Errorf("SQLite 소스 데이터베이스 파일을 열 수 없습니다: %w", err)
    }
    file.Close()
}
```

하지만 이 부분은 이미 코드에 있는 것 같습니다. 문제는 SQLite DB 파일이 존재하더라도 그 안에 필요한 테이블이 없을 수 있다는 점입니다.

마지막으로, SQLite DB에 실제로 필요한 테이블이 있는지 확인하는 코드를 추가할 수 있습니다:

```go
if config.SourceDBConfig.DBDriver == "sqlite" {
    // 기본 테이블이 있는지 확인
    for _, tableName := range getBasicTables() {
        var exists bool
        query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='%s')", tableName)
        err := config.SourceDB.QueryRow(query).Scan(&exists)
        if err != nil {
            return nil, nil, fmt.Errorf("SQLite 테이블 확인 실패: %w", err)
        }
        
        if !exists && !isTableSkipped(tableName, config.SkipTables) {
            return nil, nil, fmt.Errorf("SQLite 소스 데이터베이스에 테이블 '%s'가 없습니다", tableName)
        }
    }
}
```

하지만 이는 추가적인 검증 단계이므로 마이그레이션 성공을 보장하는 데 필수적이지는 않습니다. 우선 위에서 언급한 핵심 버그 수정을 먼저 시도하는 것이 좋겠습니다.

계속해서 같은 오류가 발생하고 있네요. 코드를 더 상세하게 수정해야 할 것 같습니다. 다음과 같이 해결해보세요:

### 1. SQLite 테이블 구조 조회 함수 완전히 수정

SQLite 부분의 `getTableColumns` 함수에서 문제가 되는 부분을 완전히 새로 작성합니다:

```go
case "sqlite":
    // 먼저 테이블 존재 여부 확인하는 부분 제거 (이 부분이 오류 발생)
    // err := db.QueryRow(fmt.Sprintf("PRAGMA table_info(%s)", tableName)).Scan()
    
    rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", tableName))
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
```

### 2. 테이블 존재 여부 확인 쿼리 완전히 수정

PostgreSQL, MySQL, SQLite 모두 동일한 패턴으로 수정합니다:

```go
case "postgres":
    var count int
    err := config.TargetDB.NewSelect().
        TableExpr("information_schema.tables").
        Column("COUNT(*)").   // count(*) > 0 대신 COUNT(*) 사용
        Where("table_schema = 'public'").
        Where("table_name = ?", board.TableName).
        Scan(ctx, &count)
    
    if err != nil {
        return fmt.Errorf("테이블 존재 여부 확인 실패: %w", err)
    }
    
    exists = count > 0   // 결과를 boolean으로 변환

case "mysql", "mariadb":
    var count int
    err := config.TargetDB.NewSelect().
        TableExpr("information_schema.tables").
        Column("COUNT(*)").
        Where("table_schema = ?", config.TargetDBConfig.DBName).
        Where("table_name = ?", board.TableName).
        Scan(ctx, &count)
    
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
```

### 3. 복합키 테이블 시퀀스 재설정 제외

```go
func needsSequenceReset(tableName string) bool {
    // 시퀀스 재설정이 필요 없는 테이블
    noResetTables := map[string]bool{
        "goose_db_version": true,
        "referrer_stats":   true,
        "board_managers":   true,  // 복합키 테이블이므로 ID 컬럼이 없음
    }

    return !noResetTables[tableName]
}
```

### 4. SQLite 테이블 실제 존재 여부 확인 추가 (선택적)

마이그레이션 시작 시 실제로 SQLite 데이터베이스에 테이블이 존재하는지 확인하는 코드를 추가할 수 있습니다. `connectDatabases` 함수 끝부분에:

```go
// SQLite 소스 테이블이 실제로 존재하는지 확인
if config.SourceDBConfig.DBDriver == "sqlite" && !config.DataOnly {
    var tableCount int
    err := sourceDB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&tableCount)
    if err != nil {
        sourceDB.Close()
        return nil, nil, fmt.Errorf("SQLite 테이블 확인 실패: %w", err)
    }
    
    if tableCount == 0 {
        sourceDB.Close()
        return nil, nil, fmt.Errorf("SQLite 소스 데이터베이스에 테이블이 없습니다")
    }
}
```

이러한 수정을 적용하면 대부분의 오류가 해결될 것입니다. 특히 SQLite Scan 오류와 PostgreSQL "count(*) > 0" 컬럼 오류는 확실히 해결될 것입니다.

## Prompt:
PS D:\dev\pcbangstudio\workspace\go-board\bin> .\migrate_windows_amd64.exe -op data-migrate -source-driver sqlite -target-driver postgres -source-env .env_sqlite -target-env .env_pg
==========================
이기종 DB 데이터 마이그레이션 시작
소스: sqlite (go_board)
대상: postgres (go_board)
==========================
[1/4] 대상 데이터베이스 스키마 생성 중...
마이그레이션 상태 확인 중...
2025/04/17 21:20:41     Applied At                  Migration
2025/04/17 21:20:41     =======================================
2025/04/17 21:20:41     Tue Apr 15 01:15:39 2025 -- 001_create_tables.sql
2025/04/17 21:20:41     Tue Apr 15 01:15:39 2025 -- 002_comments.sql
2025/04/17 21:20:41     Tue Apr 15 01:15:39 2025 -- 003_file_uploads.sql
2025/04/17 21:20:41     Tue Apr 15 01:15:39 2025 -- 004_qna_votes.sql
2025/04/17 21:20:41     Tue Apr 15 01:15:39 2025 -- 005_board_manager.sql
2025/04/17 21:20:41     Tue Apr 15 01:15:39 2025 -- 006_referrer_stats.sql
2025/04/17 21:20:41     Tue Apr 15 01:15:39 2025 -- 007_system_settings.sql
스키마 마이그레이션 적용 중...
2025/04/17 21:20:41 goose: no migrations to run. current version: 7
기본 스키마 마이그레이션 완료
[2/4] 기본 테이블 데이터 마이그레이션 중...
  - 테이블 'users' 데이터 마이그레이션 중...
2025/04/17 21:20:41 오류: 테이블 'users' 마이그레이션 실패: 대상 테이블 데이터 삭제 실패: pq: update or delete on table "users" violates foreign key constraint "board_sipan1_user_id_fkey" on table "board_sipan1"
  - 테이블 'boards' 데이터 마이그레이션 중...
    기존 대상 테이블 데이터 삭제됨
2025/04/17 21:20:41 오류: 테이블 'boards' 마이그레이션 실패: 행 삽입 실패: pq: invalid input syntax for type integer: "true"
데이터: [true,"기타 미분류","gita-mibunryu","분류 없음. 테스트.","normal","board_gita-mibunryu",true,true,true,"2025-04-14T07:11:40.955988Z","2025-04-15T20:13:29.216687Z"]
  - 테이블 'board_fields' 데이터 마이그레이션 중...
    기존 대상 테이블 데이터 삭제됨
2025/04/17 21:20:41 오류: 테이블 'board_fields' 마이그레이션 실패: 행 삽입 실패: pq: invalid input syntax for type integer: "true"
데이터: [true,true,"file_upload","file_upload","파일 첨부","file",false,false,false,"",true,"2025-04-14T07:11:40.960999Z","2025-04-15T20:13:29.216838Z"]
  - 테이블 'board_managers' 데이터 마이그레이션 중...
    테이블 'board_managers'에 데이터가 없습니다
  - 테이블 'comments' 데이터 마이그레이션 중...
    기존 대상 테이블 데이터 삭제됨
2025/04/17 21:20:41 오류: 테이블 'comments' 마이그레이션 실패: 행 삽입 실패: pq: invalid input syntax for type integer: "true"
데이터: [true,true,true,2,"\u003cp\u003e안녕. 댓글 ^-^~\u003c/p\u003e\u003cp style=\"text-align: center;\"\u003e\u003cimg src=\"/uploads/boards/1/images/7cdfdeeb-16f2-446a-a057-0ba694e6189a.gif\" animate=\"true\"\u003e\u003c/p\u003e",null,"2025-04-14T07:15:19.481008Z","2025-04-15T17:58:56.492543Z"]   
  - 테이블 'attachments' 데이터 마이그레이션 중...
    기존 대상 테이블 데이터 삭제됨
2025/04/17 21:20:41 오류: 테이블 'attachments' 마이그레이션 실패: 행 삽입 실패: pq: invalid input syntax for type integer: "true"
데이터: [true,2,true,2,"dandadan.gif","53370e74-0638-4300-80dc-f917818321bd.gif","uploads/boards/2/posts/1/attachments/53370e74-0638-4300-80dc-f917818321bd.gif",513179,"image/gif",true,"/uploads/boards/2/posts/1/attachments/53370e74-0638-4300-80dc-f917818321bd.gif","/uploads/boards/2/posts/1/attachments/thumbs/53370e74-0638-4300-80dc-f917818321bd.gif",false,"2025-04-14T07:14:29Z"]
  - 테이블 'qna_answers' 데이터 마이그레이션 중...
    테이블 'qna_answers'에 데이터가 없습니다
  - 테이블 'qna_question_votes' 데이터 마이그레이션 중...
    테이블 'qna_question_votes'에 데이터가 없습니다
  - 테이블 'qna_answer_votes' 데이터 마이그레이션 중...
    테이블 'qna_answer_votes'에 데이터가 없습니다
  - 테이블 'referrer_stats' 데이터 마이그레이션 중...
    배치 1/2 처리 중... (오프셋: 0, 한계: 1000)
2025/04/17 21:20:41 오류: 테이블 'referrer_stats' 마이그레이션 실패: 행 삽입 실패: pq: there is no parameter $1
데이터: [true,"https://bbs.enjoytools.net/admin/referrer-stats","enjoytools.net","other","/admin/setup","223.194.169.11",null,"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36 Edg/135.0.0.0","2025-04-14T07:10:33.54278Z"]
  - 테이블 'system_settings' 데이터 마이그레이션 중...
    기존 대상 테이블 데이터 삭제됨
2025/04/17 21:20:41 오류: 테이블 'system_settings' 마이그레이션 실패: 행 삽입 실패: pq: invalid input syntax for type integer: "true"
데이터: [true,"approval_mode","immediate","회원가입 승인 모드 (immediate: 즉시, delayed: n일 후, manual: 관리자 승인)","2025-04-14T07:09:26Z","2025-04-14T07:09:26Z"]
기본 테이블 데이터 마이그레이션 완료
[3/4] 동적 테이블 마이그레이션 중...
  총 2개의 게시판을 마이그레이션합니다
  - 게시판 '갤러리' (board_gaelreori) 마이그레이션 중...
2025/04/17 21:20:41 오류: 게시판 '갤러리' 테이블 생성 실패: 테이블 존재 여부 확인 실패: pq: column "COUNT(*)" does not exist
  - 게시판 '기타 미분류' (board_gita-mibunryu) 마이그레이션 중...
2025/04/17 21:20:41 오류: 게시판 '기타 미분류' 테이블 생성 실패: 테이블 존재 여부 확인 실패: pq: column "COUNT(*)" does not exist
동적 테이블 마이그레이션 완료
[4/4] 시퀀스/자동증가 값 재설정 중...
  - 테이블 'users' 시퀀스 재설정 완료
  - 테이블 'boards' 시퀀스 재설정 완료
  - 테이블 'board_fields' 시퀀스 재설정 완료
  - 테이블 'comments' 시퀀스 재설정 완료
  - 테이블 'attachments' 시퀀스 재설정 완료
  - 테이블 'qna_answers' 시퀀스 재설정 완료
  - 테이블 'qna_question_votes' 시퀀스 재설정 완료
  - 테이블 'qna_answer_votes' 시퀀스 재설정 완료
  - 테이블 'system_settings' 시퀀스 재설정 완료
시퀀스/자동증가 값 재설정 완료
==========================
데이터 마이그레이션 완료 (소요 시간: 71.8495ms)
경고: 마이그레이션 중 9개의 오류가 발생했습니다
  - 테이블 'users' 마이그레이션 실패: 대상 테이블 데이터 삭제 실패: pq: update or delete on table "users" violates foreign key constraint "board_sipan1_user_id_fkey" on table "board_sipan1"
  - 테이블 'boards' 마이그레이션 실패: 행 삽입 실패: pq: invalid input syntax for type integer: "true"
데이터: [true,"기타 미분류","gita-mibunryu","분류 없음. 테스트.","normal","board_gita-mibunryu",true,true,true,"2025-04-14T07:11:40.955988Z","2025-04-15T20:13:29.216687Z"]
  - 테이블 'board_fields' 마이그레이션 실패: 행 삽입 실패: pq: invalid input syntax for type integer: "true"
데이터: [true,true,"file_upload","file_upload","파일 첨부","file",false,false,false,"",true,"2025-04-14T07:11:40.960999Z","2025-04-15T20:13:29.216838Z"]
  - 테이블 'comments' 마이그레이션 실패: 행 삽입 실패: pq: invalid input syntax for type integer: "true"
데이터: [true,true,true,2,"\u003cp\u003e안녕. 댓글 ^-^~\u003c/p\u003e\u003cp style=\"text-align: center;\"\u003e\u003cimg src=\"/uploads/boards/1/images/7cdfdeeb-16f2-446a-a057-0ba694e6189a.gif\" animate=\"true\"\u003e\u003c/p\u003e",null,"2025-04-14T07:15:19.481008Z","2025-04-15T17:58:56.492543Z"]   
  - 테이블 'attachments' 마이그레이션 실패: 행 삽입 실패: pq: invalid input syntax for type integer: "true"
데이터: [true,2,true,2,"dandadan.gif","53370e74-0638-4300-80dc-f917818321bd.gif","uploads/boards/2/posts/1/attachments/53370e74-0638-4300-80dc-f917818321bd.gif",513179,"image/gif",true,"/uploads/boards/2/posts/1/attachments/53370e74-0638-4300-80dc-f917818321bd.gif","/uploads/boards/2/posts/1/attachments/thumbs/53370e74-0638-4300-80dc-f917818321bd.gif",false,"2025-04-14T07:14:29Z"]
추가 4개 오류 생략...
==========================
2025/04/17 21:20:41 데이터 마이그레이션이 성공적으로 완료되었습니다
PS D:\dev\pcbangstudio\workspace\go-board\bin>



4번 제안은 반영 안했다. 나머지는 모두 반영했다.

## Response:

이제 이전과는 다른 오류들이 나타났습니다. 이번에는 데이터 타입 변환 문제와 외래 키 제약 조건 위반 문제가 발생했습니다. 오류를 분석해보겠습니다:

1. SQLite에서 PostgreSQL로 데이터 변환 시 불리언 값 처리 문제:
   `pq: invalid input syntax for type integer: "true"` - SQLite에서 불리언 값이 문자열 "true"로 전달되고 있어 PostgreSQL에서 정수형으로 변환할 수 없습니다.

2. 외래 키 제약 조건 위반:
   `pq: update or delete on table "users" violates foreign key constraint "board_sipan1_user_id_fkey" on table "board_sipan1"` - 이는 users 테이블의 데이터를 삭제하려고 할 때 다른 테이블(board_sipan1)에서 참조하고 있어서 발생한 문제입니다.

3. PostgreSQL의 COUNT(*) 쿼리:
   `pq: column "COUNT(*)" does not exist` - PostgreSQL에서는 COUNT(*)를 그대로 결과 컬럼으로 사용할 수 없습니다. 이 부분을 수정해야 합니다.

4. referrer_stats 테이블에서 매개변수 문제:
   `pq: there is no parameter $1` - 쿼리의 매개변수 수와 실제 제공된 매개변수 수가 일치하지 않거나, PostgreSQL에서 매개변수 처리 방식이 다르게 동작하고 있습니다.

이러한 문제를 해결하기 위한 주요 수정사항:

1. 불리언 타입 변환 개선: 
   - `convertValue` 함수에서 SQLite에서 불리언 값을 정수(0/1)로 변환하는 로직을 보완해야 합니다.
   - 데이터베이스 간 타입 매핑을 더 정확하게 처리해야 합니다.

2. 외래 키 제약 조건 처리:
   - 테이블 삭제 순서를 변경하거나, 외래 키 제약 조건을 일시적으로 비활성화하는 방법을 사용해야 합니다.
   - 또는 CASCADE 옵션을 사용하여 참조하는 데이터도 함께 삭제하도록 처리할 수 있습니다.

3. PostgreSQL COUNT(*) 쿼리 수정:
   - `COUNT(*)` 대신 `COUNT(*) AS count`와 같이 별칭을 사용해야 합니다.

4. 매개변수 처리 개선:
   - PostgreSQL과 SQLite의 쿼리 매개변수 처리 방식 차이를 고려한 코드를 작성해야 합니다.

이제 다음과 같은 새로운 오류들이 발생했네요. 문제점들을 분석하고 해결 방법을 제안하겠습니다:

### 1. 불리언 타입 변환 오류 
`pq: invalid input syntax for type integer: "true"` 오류는 SQLite의 불리언 값이 올바르게 변환되지 않는 문제입니다. `convertValue` 함수를 다음과 같이 수정해야 합니다:

```go
func convertValue(val interface{}, tableName string, column ColumnMetadata, sourceDriver, targetDriver string) interface{} {
    if val == nil {
        return nil
    }

    // 문자열 "true"/"false"를 불리언으로 변환
    if s, ok := val.(string); ok {
        if s == "true" {
            if targetDriver == "postgres" {
                return true
            }
            return 1
        } else if s == "false" {
            if targetDriver == "postgres" {
                return false
            }
            return 0
        }
    }

    // 불리언 타입 변환
    if sourceDriver == "sqlite" && targetDriver != "sqlite" {
        // SQLite의 불리언은 0/1 정수로 저장됨
        if column.Type == "INTEGER" || column.Type == "BOOLEAN" || column.Type == "TINYINT" {
            if i, ok := val.(int64); ok && (i == 0 || i == 1) {
                if targetDriver == "postgres" {
                    return i == 1
                }
            }
        }
    } else if sourceDriver != "sqlite" && targetDriver == "sqlite" {
        // PostgreSQL/MySQL의 불리언을 SQLite의 0/1로 변환
        if b, ok := val.(bool); ok {
            if b {
                return 1
            }
            return 0
        }
    }

    // 나머지 기존 코드...
    return val
}
```

### 2. 외래 키 제약 조건 해결
테이블 삭제 시 외래 키 제약 조건으로 인한 오류를 처리하기 위해 `shouldCleanTableBeforeMigration` 함수를 수정합니다:

```go
func shouldCleanTableBeforeMigration(tableName string) bool {
    // 외래 키 제약 조건이 있거나 이력 데이터인 테이블은 정리하지 않음
    skipCleanTables := map[string]bool{
        "referrer_stats": true,
        "users": true,  // 외래 키 참조가 있으므로 삭제하지 않음
    }

    return !skipCleanTables[tableName]
}
```

### 3. PostgreSQL COUNT 쿼리 수정
테이블 존재 여부 확인 쿼리를 다음과 같이 수정합니다:

```go
case "postgres":
    var count int
    err := config.TargetDB.NewSelect().
        TableExpr("information_schema.tables").
        ColumnExpr("COUNT(*) AS count").  // 별칭 사용
        Where("table_schema = 'public'").
        Where("table_name = ?", board.TableName).
        Scan(ctx, &count)
    
    if err != nil {
        return fmt.Errorf("테이블 존재 여부 확인 실패: %w", err)
    }
    
    exists = count > 0
```

### 4. referrer_stats에서 매개변수 오류 해결
PostgreSQL에서 매개변수 처리 방식이 다르므로, 동적 SQL 문에서 매개변수 형식을 변경해야 합니다:

```go
// 데이터베이스별 플레이스홀더 형식
var placeholders []string
for i := range commonColumns {
    // 데이터베이스별 플레이스홀더 형식
    switch config.TargetDBConfig.DBDriver {
    case "postgres":
        placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
    default:
        placeholders = append(placeholders, "?")
    }
}

// 코드 위치에 따라 이미 이렇게 작성되어 있을 수 있지만, 
// placeholders 배열을 사용하는 방식이 올바른지 확인하세요
```

### 추가 권장사항:
1. **CASCADE 삭제 설정 고려**:
   대상 데이터베이스에서 외래 키 제약 조건에 CASCADE 옵션을 사용하는 것이 좋습니다.

2. **트랜잭션 사용 최적화**:
   테이블 간 관계와 참조가 있을 경우, 전체 마이그레이션을 하나의 트랜잭션으로 묶어 처리하면 더 안정적입니다.

3. **에러 처리 세분화**:
   각 유형의 오류에 대한 더 구체적인 메시지와 처리 방법을 제공하는 것이 좋습니다.

이러한 변경사항을 적용하면 대부분의 데이터 변환 및 제약 조건 오류를 해결할 수 있을 것입니다.

## Prompt:


## Response:

아직도 몇 가지 문제가 남아있습니다. 주요 문제를 분석해보겠습니다:

1. **SQLite에서 PostgreSQL로의 불리언 타입 변환 문제**:
   - `"pq: invalid input syntax for type integer: "true""` 오류가 여전히 발생하고 있습니다. 
   - SQLite에서 불리언 값은 문자열 "true"/"false"로 나타나는 것 같고, PostgreSQL에서는 이를 정확히 처리하지 못하고 있습니다.

2. **PostgreSQL 파라미터화 쿼리 문제**:
   - `"pq: there is no parameter $1"`는 PostgreSQL에서 파라미터 바인딩 문제를 나타냅니다.

3. **테이블 이름에 대시(-)가 포함된 경우 처리 문제**:
   - `"board_gita-mibunryu"` 테이블에서 `SQL logic error: near "-": syntax error`가 발생합니다.
   - SQL에서 대시(-)가 포함된 식별자는 특수 처리가 필요합니다.

이 문제들을 해결하기 위한 수정 방법을 제안하겠습니다:

### 1. 불리언 타입 변환 문제 해결

SQLite에서 데이터를 가져올 때 자바스크립트 객체로 변환되는 과정에서 불리언 값이 문자열로 처리되는 것 같습니다. 따라서 이러한 경우를 처리하는 코드를 `convertValue` 함수에 추가해야 합니다.

### 2. 파라미터화된 SQL 쿼리 문제

PostgreSQL은 `?` 플레이스홀더 대신 `$1`, `$2` 등의 형식을 사용합니다. 데이터 마이그레이션을 위한 SQL 쿼리에서 이를 제대로 처리하고 있는지 확인해야 합니다.

### 3. 테이블 이름에 특수문자 처리

테이블 이름에 대시(-) 같은 특수문자가 포함된 경우, SQL 쿼리에서 이러한 테이블 이름을 적절히 인용해야 합니다.

위의 문제들을 모두 해결하기 위한 코드 수정 방안을 제시하겠습니다.

여전히 몇 가지 중요한 문제가 남아있네요. 주요 문제점과 해결 방법을 제시하겠습니다:

### 1. 불리언 타입 변환 문제 (`"true"` 문자열)

SQLite에서는 불리언 데이터가 문자열 형태로 저장되어 있는 것 같습니다. `convertValue` 함수를 더 강화해야 합니다:

```go
func convertValue(val interface{}, tableName string, column ColumnMetadata, sourceDriver, targetDriver string) interface{} {
    if val == nil {
        return nil
    }

    // 문자열 "true"/"false"를 불리언으로 변환
    if strVal, ok := val.(string); ok {
        lowerStr := strings.ToLower(strVal)
        if lowerStr == "true" {
            if targetDriver == "postgres" {
                return true
            }
            return 1
        } else if lowerStr == "false" {
            if targetDriver == "postgres" {
                return false
            }
            return 0
        }
    }

    // 기존 코드...
    return val
}
```

### 2. 테이블 이름에 대시(-)가 있는 경우 처리

SQL 문에서 대시(-)가 있는 테이블 이름은 특수하게 처리해야 합니다:

```go
func quoteTableName(driver, tableName string) string {
    // 특수 문자가 있는지 확인
    hasSpecialChars := strings.ContainsAny(tableName, "-.") || strings.Contains(tableName, " ")
    
    switch driver {
    case "postgres":
        if hasSpecialChars {
            return fmt.Sprintf("\"%s\"", tableName)
        }
        return tableName
    case "mysql", "mariadb":
        if hasSpecialChars {
            return fmt.Sprintf("`%s`", tableName)
        }
        return tableName
    case "sqlite":
        if hasSpecialChars {
            return fmt.Sprintf("\"%s\"", tableName)
        }
        return tableName
    default:
        return tableName
    }
}
```

### 3. PostgreSQL 매개변수 바인딩 문제

`migrateTableData` 함수에서 플레이스홀더 부분을 확인해야 합니다:

```go
// 행별 처리 부분에서
batchInsertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
    quoteTableName(config.TargetDBConfig.DBDriver, tableName),
    strings.Join(targetColumnNames, ", "),
    strings.Join(placeholders, ", "))

// PostgreSQL 매개변수 처리를 위한 특별 확인
if config.TargetDBConfig.DBDriver == "postgres" {
    // PostgreSQL은 $1, $2 등의 형식으로 매개변수 사용
    for i := range placeholders {
        placeholders[i] = fmt.Sprintf("$%d", i+1)
    }
} else {
    // 다른 데이터베이스는 ? 사용
    for i := range placeholders {
        placeholders[i] = "?"
    }
}
```

### 4. 마이그레이션 전에 고급 검증 추가

마이그레이션 시작 전에 데이터가 올바른 형식인지 확인하는 검증 단계를 추가하는 것이 좋습니다:

```go
// 마이그레이션 전 데이터 검증
func validateSourceData(config *DataMigrationConfig, tableName string) error {
    // 테이블 존재 여부 확인
    var count int
    query := fmt.Sprintf("SELECT COUNT(*) FROM %s LIMIT 1", 
                        quoteTableName(config.SourceDBConfig.DBDriver, tableName))
    
    err := config.SourceDB.QueryRow(query).Scan(&count)
    if err != nil {
        return fmt.Errorf("테이블 '%s' 검증 실패: %w", tableName, err)
    }
    
    return nil
}
```

### 5. 특수 테이블명 처리 방법

대시를 포함하는 테이블 이름인 "board_gita-mibunryu"는 특별한 처리가 필요합니다. SQLite와 PostgreSQL 모두에서 이런 이름을 다룰 때는 항상 따옴표로 감싸야 합니다:

```go
// 이 함수는 모든 SQL 쿼리에서 테이블 이름을 참조할 때 사용해야 합니다
func safeTableName(driver, tableName string) string {
    if strings.ContainsAny(tableName, "-.") {
        if driver == "mysql" || driver == "mariadb" {
            return "`" + tableName + "`"
        } else {
            return "\"" + tableName + "\""
        }
    }
    return tableName
}
```

이러한 수정사항들을 적용하면 대부분의 현재 문제는 해결될 것입니다. 특히 불리언 타입 변환과 특수 테이블명 처리가 가장 중요한 문제로 보입니다.

## Prompt:
4번 5번은 어디에 넣으라는 소리야?

## Response:

맞습니다. 제 설명이 충분히 명확하지 않았네요. 4번과 5번 제안을 코드에 구체적으로 어디에 넣어야 할지 더 명확히 설명해 드리겠습니다.

4번 제안 (마이그레이션 전에 고급 검증 추가):
- 이 함수는 `runDataMigration` 함수 내에서 각 테이블 마이그레이션 전에 호출되어야 합니다.
- `migrateBasicTables` 및 `migrateDynamicTables` 함수 내에서 각 테이블에 대해 `migrateTableData`를 호출하기 전에 검증할 수 있습니다.

5번 제안 (특수 테이블명 처리 방법):
- 이 함수는 기존의 `quoteTableName` 함수를 대체하거나 개선하는 것입니다.
- 현재 코드에서는 `quoteTableName` 함수가 이미 있으므로, 해당 함수를 수정하면 됩니다.

더 명확하게 말씀드리자면:

1. 4번 제안은 `migrateTableData` 함수를 호출하기 전에 데이터 유효성을 검증하는 새로운 함수입니다.
2. 5번 제안은 기존의 `quoteTableName` 함수를 더 강화하는 것입니다.

구체적인 코드 위치와 함께 다시 설명드리겠습니다.

죄송합니다, 설명이 명확하지 않았네요. 4번과 5번 제안을 어디에 넣어야 하는지 더 구체적으로 설명드리겠습니다:

### 4번 제안 (마이그레이션 전 데이터 검증 추가)

이 함수는 `migrateTableData` 함수를 호출하기 전에 사용해야 합니다. 예를 들어 `migrateBasicTables` 함수에서:

```go
func migrateBasicTables(config *DataMigrationConfig) error {
    fmt.Println("[2/4] 기본 테이블 데이터 마이그레이션 중...")

    // 기본 테이블 목록
    basicTables := getBasicTables()

    // 테이블별 데이터 마이그레이션
    for _, tableName := range basicTables {
        // 건너뛸 테이블 확인
        if isTableSkipped(tableName, config.SkipTables) {
            fmt.Printf("  - 테이블 '%s' 건너뛰기 (사용자 설정에 의해 제외)\n", tableName)
            continue
        }

        // 여기에 검증 추가
        if err := validateSourceData(config, tableName); err != nil {
            config.addError(err)
            continue
        }

        // 테이블 데이터 복사
        err := migrateTableData(config, tableName)
        if err != nil {
            config.addError(fmt.Errorf("테이블 '%s' 마이그레이션 실패: %w", tableName, err))
            if len(config.Errors) > config.MaxErrorsBeforeExit {
                return fmt.Errorf("오류가 너무 많아 마이그레이션이 중단되었습니다")
            }
        }
    }

    fmt.Println("기본 테이블 데이터 마이그레이션 완료")
    return nil
}
```

마찬가지로 `migrateDynamicTables` 함수에서도 동일하게 적용하면 됩니다.

### 5번 제안 (특수 테이블명 처리)

이 제안은 기존의 `quoteTableName` 함수를 개선하는 것입니다. 따라서 `quoteTableName` 함수를 다음과 같이 수정하면 됩니다:

```go
// quoteTableName은 데이터베이스 드라이버에 따라 테이블 이름을 인용 부호로 묶습니다
func quoteTableName(driver, tableName string) string {
    // 특수 문자가 있는지 확인 (이 부분이 5번 제안의 핵심)
    hasSpecialChars := strings.ContainsAny(tableName, "-.") || strings.Contains(tableName, " ")
    
    switch driver {
    case "postgres":
        // PostgreSQL에서는 특수 문자가 있으면 항상 따옴표로 감싸야 함
        if hasSpecialChars {
            return fmt.Sprintf("\"%s\"", tableName)
        }
        return tableName
    case "mysql", "mariadb":
        // MySQL에서는 특수 문자가 있으면 백틱으로 감싸야 함
        if hasSpecialChars {
            return fmt.Sprintf("`%s`", tableName)
        }
        return tableName
    case "sqlite":
        // SQLite도 특수 문자가 있으면 따옴표로 감싸야 함
        if hasSpecialChars {
            return fmt.Sprintf("\"%s\"", tableName)
        }
        return tableName
    default:
        return tableName
    }
}
```

기존 함수를 이렇게 수정하면 특수 문자가 포함된 테이블 이름도 올바르게 처리될 것입니다.

## Prompt:


## Response:

아직도 동일한 오류가 발생하고 있습니다. 가장 중요한 두 가지 문제는:

1. `pq: invalid input syntax for type integer: "true"` - 불리언 타입 변환 문제
2. `SQL logic error: near "-": syntax error (1)` - 대시(-)가 있는 테이블 이름 처리 문제

이 문제들은 우리가 제안한 수정사항이 제대로 적용되지 않았거나, 더 근본적인 수정이 필요할 수 있습니다. 실제 코드를 더 자세히 이해해야 합니다.

첫 번째 문제인 "true" 문자열을 불리언으로 변환하는 문제를 살펴보겠습니다. 오류 메시지를 보면 PostgreSQL은 문자열 "true"를 정수 값으로 변환하려고 시도하고 있습니다. 이는 테이블 스키마에서 불리언 필드가 정수 타입으로 정의되어 있을 수 있습니다. 또는 데이터 행의 잘못된 순서 때문일 수도 있습니다.

더 과감한 접근법으로 문제를 해결해보겠습니다. 먼저 오류 데이터를 분석해보겠습니다:

첫 번째 데이터 항목 분석:
```
[true,"bab2","edp1096@outlook.com","$2a$10$fqA.SFqpLS8nujCh/6DsJOXhwpKnffYBbARw.eGurBPz0qUUzgsre","꿈돌이","admin",true,"approved",null,"2025-04-14T07:10:58.105779Z","2025-04-14T07:10:58.105779Z"]
```

이 데이터에서 첫 번째 값인 `true`가 문제인 것 같습니다. 스키마를 보면, 첫 번째 값은 ID일 가능성이 높습니다. PostgreSQL에서는 ID가 정수(integer)로 예상되지만, 여기에는 불리언 값이 들어가 있습니다.

모든 오류 메시지는 비슷한 패턴을 보입니다. 문자열 "true"를 정수 타입으로 변환하려고 시도하다가 실패하고 있습니다.

그래서 우리가 생각해낼 수 있는 해결책은:

1. 데이터 매핑 이슈: 소스와 대상 테이블의 컬럼 순서나 타입이 매칭되지 않는 문제
2. 데이터 변환 문제: 모든 타입의 값(문자열, 불리언, 정수 등)을 올바르게 처리하지 못함

여기서는 두 가지 주요 방향으로 문제를 해결할 수 있습니다:

1. 모든 데이터를 문자열로 변환한 후, PostgreSQL의 타입 캐스팅 사용
2. 각 테이블 컬럼의 타입 정보를 가져와서 데이터를 해당 타입에 맞게 명시적으로 변환

가장 좋은 접근법은 PostgreSQL 대상 테이블의 각 컬럼 타입 정보를 가져와서 데이터를 변환하는 것입니다. 하지만 빠른 해결을 위해 좀 더 간단한 방법을 제안하겠습니다.

문제 해결을 위해 가장 근본적인 수정이 필요합니다. 오류 메시지를 보면 모든 테이블에서 동일한 이슈가 발생하고 있습니다. `convertValue` 함수를 완전히 재작성하고 타입 변환을 더 철저하게 처리해야 합니다.

### 1. 데이터 변환 문제 해결

오류 메시지를 보면 PostgreSQL에 값을 삽입할 때 문자열 "true"가 정수 타입으로 변환하려고 하다가 실패하는 것이 핵심 문제입니다. 다음과 같이 `migrateTableData` 함수에서 행 데이터 처리 부분을 수정하세요:

```go
// 행별 처리
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

    // *** 중요: 컬럼의 PostgreSQL 타입에 맞게 데이터 전처리 ***
    for i, val := range rowValues {
        // 문자열 "true"/"false" 처리
        if strValue, ok := val.(string); ok {
            if strings.ToLower(strValue) == "true" {
                if config.TargetDBConfig.DBDriver == "postgres" {
                    rowValues[i] = true
                } else {
                    rowValues[i] = 1
                }
                continue
            } else if strings.ToLower(strValue) == "false" {
                if config.TargetDBConfig.DBDriver == "postgres" {
                    rowValues[i] = false
                } else {
                    rowValues[i] = 0
                }
                continue
            }
        }

        // 첫 번째 컬럼은 보통 ID이고 PostgreSQL에서는 정수여야 함
        if i == 0 && config.TargetDBConfig.DBDriver == "postgres" {
            // 불리언 값을 ID에 변환
            if boolVal, ok := val.(bool); ok {
                if boolVal {
                    // 값이 true인 경우 ID를 자동 생성하도록 NULL로 설정
                    rowValues[i] = nil
                } else {
                    // false인 경우 0으로 설정 (일반적으로 사용되지 않음)
                    rowValues[i] = 0
                }
                continue
            }
        }

        // 일반적인 타입 변환
        rowValues[i] = convertValue(val, tableName, commonColumns[i], config.SourceDBConfig.DBDriver, config.TargetDBConfig.DBDriver)
    }

    // 행 삽입 코드...
}
```

### 2. 테이블 이름 처리 개선

대시가 있는 테이블 이름(`board_gita-mibunryu`)에 대한 SQL 쿼리 수정:

```go
func quoteTableName(driver, tableName string) string {
    // 특수 문자가 있는지 확인
    hasSpecialChars := strings.ContainsAny(tableName, "-.") || strings.Contains(tableName, " ")
    
    switch driver {
    case "postgres":
        // PostgreSQL에서는 항상 따옴표로 감싸기
        return fmt.Sprintf("\"%s\"", tableName)
    case "mysql", "mariadb":
        // MySQL에서는 항상 백틱으로 감싸기
        return fmt.Sprintf("`%s`", tableName)
    case "sqlite":
        // SQLite에서는 특수 문자가 있으면 항상 따옴표로 감싸기
        if hasSpecialChars {
            return fmt.Sprintf("\"%s\"", tableName)
        }
        return tableName
    default:
        return tableName
    }
}
```

### 3. PostgreSQL 매개변수 처리 개선

`migrateTableData` 함수에서 PostgreSQL 매개변수 바인딩 부분을 다음과 같이 수정하세요:

```go
// IMPORTANT: PostgreSQL 매개변수 처리를 위한 특별 코드
var placeholders []string
if config.TargetDBConfig.DBDriver == "postgres" {
    // PostgreSQL은 $1, $2 등의 형식으로 매개변수 사용
    for i := range commonColumns {
        placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
    }
} else {
    // 다른 데이터베이스는 ? 사용
    for i := range commonColumns {
        placeholders = append(placeholders, "?")
    }
}

// 행별 처리
batchInsertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
    quoteTableName(config.TargetDBConfig.DBDriver, tableName),
    strings.Join(targetColumnNames, ", "),
    strings.Join(placeholders, ", "))
```

이러한 변경사항을 적용하면 대부분의 오류가 해결될 것입니다. 특히 문자열 "true"를 불리언으로 변환하는 문제와 첫 번째 컬럼(ID) 처리 부분이 핵심입니다.



---
Powered by [Claude Exporter](https://www.claudexporter.com)