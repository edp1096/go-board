package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"slices"

	"github.com/edp1096/toy-board/internal/models"
	"github.com/edp1096/toy-board/internal/repository"
	"github.com/edp1096/toy-board/internal/service"
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

	// 4. 외래 키 제약 조건 비활성화
	if err := disableForeignKeyConstraints(config); err != nil {
		log.Printf("경고: 외래 키 제약 조건 비활성화 실패: %v (계속 진행합니다)", err)
	} else {
		fmt.Println("외래 키 제약 조건 비활성화됨")
		// 함수 종료 시 제약 조건 다시 활성화
		defer func() {
			if err := enableForeignKeyConstraints(config); err != nil {
				log.Printf("경고: 외래 키 제약 조건 재활성화 실패: %v", err)
			} else {
				fmt.Println("외래 키 제약 조건 재활성화됨")
			}
		}()
	}

	// 5. 데이터 마이그레이션 실행
	startTime := time.Now()

	// 5.1 기본 테이블 데이터 마이그레이션
	if !config.DynamicTablesOnly {
		if err := migrateBasicTables(config); err != nil {
			return fmt.Errorf("기본 테이블 마이그레이션 실패: %w", err)
		}
	}

	// 5.2 동적 테이블 데이터 마이그레이션
	if !config.BasicTablesOnly {
		if err := migrateDynamicTables(config, boardService, dynamicBoardService); err != nil {
			return fmt.Errorf("동적 테이블 마이그레이션 실패: %w", err)
		}
	}

	// 6. 댓글 수 업데이트
	if !config.SchemaOnly {
		if err := updateCommentCounts(config); err != nil {
			log.Printf("댓글 수 업데이트 실패: %v (무시하고 계속 진행합니다)", err)
		}
	}

	// 7. 게시물 좋아요/싫어요 수 업데이트
	if !config.SchemaOnly {
		if err := updatePostVoteCounts(config); err != nil {
			log.Printf("게시물 좋아요/싫어요 수 업데이트 실패: %v (무시하고 계속 진행합니다)", err)
		}
	}

	// 8. 댓글 좋아요/싫어요 수 업데이트
	if !config.SchemaOnly {
		if err := updateCommentVoteCounts(config); err != nil {
			log.Printf("댓글 좋아요/싫어요 수 업데이트 실패: %v (무시하고 계속 진행합니다)", err)
		}
	}

	// 9. 시퀀스/자동증가 값 복구
	if err := resetSequences(config); err != nil {
		log.Printf("시퀀스 복구 실패: %v (무시하고 계속 진행합니다)", err)
	}

	// 10. 결과 요약
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
	participantRepo := repository.NewBoardParticipantRepository(config.SourceDB)

	// 서비스 생성
	boardService := service.NewBoardService(boardRepo, participantRepo, config.SourceDB)
	dynamicBoardService := service.NewDynamicBoardService(config.SourceDB)

	return boardService, dynamicBoardService, nil
}

// getBasicTables는 기본 테이블 목록을 반환합니다 (외래 키 의존성 순서로 정렬됨)
func getBasicTables() []string {
	return []string{
		"users",              // 다른 테이블이 참조하는 기본 테이블
		"boards",             // 게시판 테이블
		"board_fields",       // boards 참조
		"board_managers",     // boards와 users 참조
		"board_participants", // boards와 users 참조 (소모임 참여자)
		"system_settings",    // 독립 테이블
		"comments",           // users와 boards 참조
		"attachments",        // users와 boards 참조
		"post_votes",         // users와 boards 참조 (좋아요/싫어요)
		"comment_votes",      // users, boards, comments 참조 (좋아요/싫어요)
		"qna_answers",        // users와 boards 참조
		"qna_question_votes", // users와 boards 참조
		"qna_answer_votes",   // users와 boards와 qna_answers 참조
		"referrer_stats",     // users 참조 (선택 사항)
		"pages",              // 독립 테이블 (페이지 데이터)
		"categories",         // 독립 테이블 (카테고리 데이터)
		"category_items",     // categories, boards, pages 참조 (카테고리-아이템 관계)
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

		// 테이블 데이터 검증
		if err := validateSourceData(config, tableName); err != nil {
			config.addError(err)
			continue
		}

		// 테이블 데이터 복사 - 기본 테이블은 필드 타입 정보 없음
		err := migrateTableData(config, tableName, nil)
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

		// 테이블 데이터 검증
		if err := validateSourceData(config, board.TableName); err != nil {
			config.addError(err)
			continue
		}

		fmt.Printf("  - 게시판 '%s' (%s) 마이그레이션 중...\n", board.Name, board.TableName)

		// 게시판 테이블이 대상 DB에 존재하는지 확인
		if err := ensureDynamicTableExists(config, board, dynamicBoardService); err != nil {
			config.addError(fmt.Errorf("게시판 '%s' 테이블 생성 실패: %w", board.Name, err))
			continue
		}

		// 게시판 필드 정보 가져오기
		var fields []*models.BoardField
		err := config.SourceDB.NewSelect().
			Model(&fields).
			Where("board_id = ?", board.ID).
			Order("sort_order ASC").
			Scan(context.Background())
		if err != nil {
			config.addError(fmt.Errorf("게시판 '%s' 필드 정보 가져오기 실패: %w", board.Name, err))
			continue
		}

		// 필드 타입 정보를 맵으로 변환 (컬럼명 -> 필드타입)
		fieldTypeMap := make(map[string]models.FieldType)
		for _, field := range fields {
			fieldTypeMap[field.ColumnName] = field.FieldType
		}

		// 테이블 데이터 복사 - 필드 타입 정보 전달
		err = migrateTableData(config, board.TableName, fieldTypeMap)
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
		// SQLite용 직접 SQL 쿼리로 변경 - 수정된 부분
		var count int
		// 테이블 이름이 특수 문자를 포함하는 경우를 고려한 쿼리
		query := fmt.Sprintf("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='%s'",
			strings.Replace(board.TableName, "'", "''", -1))

		err := config.TargetDB.QueryRow(query).Scan(&count)
		if err != nil {
			return fmt.Errorf("테이블 존재 여부 확인 실패: %w", err)
		}

		exists = count > 0

		if !exists && config.VerboseLogging {
			fmt.Printf("    SQLite: 테이블 '%s'가 존재하지 않음\n", board.TableName)
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

		if config.VerboseLogging {
			fmt.Printf("    테이블 '%s'를 위한 %d개 필드 로드됨\n", board.TableName, len(fields))
		}

		// 대상 DB에 테이블 생성 - SQLite 특별 처리 추가
		if config.TargetDBConfig.DBDriver == "sqlite" {
			// SQLite용 테이블 생성 직접 수행
			if err := createSQLiteDynamicTable(config, board, fields); err != nil {
				return fmt.Errorf("SQLite 게시판 테이블 생성 실패: %w", err)
			}
		} else {
			// 다른 DB는 기존 서비스 사용
			targetDynamicService := service.NewDynamicBoardService(config.TargetDB)
			if err := targetDynamicService.CreateBoardTable(ctx, board, fields); err != nil {
				return fmt.Errorf("게시판 테이블 생성 실패: %w", err)
			}
		}

		fmt.Printf("    게시판 테이블 '%s' 생성됨\n", board.TableName)
	}

	return nil
}

// createSQLiteDynamicTable은 SQLite에 동적 게시판 테이블을 생성합니다
func createSQLiteDynamicTable(config *DataMigrationConfig, board *models.Board, fields []*models.BoardField) error {
	// 테이블 이름 준비 (이스케이프 처리)
	tableName := board.TableName
	safeTableName := tableName
	if strings.Contains(tableName, "-") || strings.Contains(tableName, ".") {
		safeTableName = fmt.Sprintf("\"%s\"", tableName)
	}

	// 기본 컬럼 정의
	columns := []string{
		"id INTEGER PRIMARY KEY AUTOINCREMENT",
		"title TEXT NOT NULL",
		"content TEXT NOT NULL",
		"user_id INTEGER NOT NULL",
		"view_count INTEGER NOT NULL DEFAULT 0",
		"comment_count INTEGER NOT NULL DEFAULT 0",
		"like_count INTEGER NOT NULL DEFAULT 0",
		"dislike_count INTEGER NOT NULL DEFAULT 0",
		"is_private TINYINT NOT NULL DEFAULT 0",
		"created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP",
		"updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP",
	}

	// 사용자 정의 필드 추가
	for _, field := range fields {
		var colType string
		switch field.FieldType {
		case models.FieldTypeNumber:
			colType = "INTEGER"
		case models.FieldTypeDate:
			colType = "TIMESTAMP"
		case models.FieldTypeCheckbox:
			colType = "TINYINT" // 0 또는 1
		default:
			colType = "TEXT" // text, textarea, select, file 등
		}

		// 컬럼 이름 처리
		safeColumnName := field.ColumnName
		if strings.Contains(field.ColumnName, "-") || strings.Contains(field.ColumnName, ".") {
			safeColumnName = fmt.Sprintf("\"%s\"", field.ColumnName)
		}

		// 필수 여부에 따라 NOT NULL 추가
		if field.Required {
			columns = append(columns, fmt.Sprintf("%s %s NOT NULL", safeColumnName, colType))
		} else {
			columns = append(columns, fmt.Sprintf("%s %s", safeColumnName, colType))
		}
	}

	// CREATE TABLE 쿼리 구성
	createQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", safeTableName, strings.Join(columns, ", "))

	// 테이블 생성 실행
	_, err := config.TargetDB.Exec(createQuery)
	if err != nil {
		return fmt.Errorf("테이블 생성 쿼리 실행 실패: %w", err)
	}

	// 인덱스 이름 및 테이블 이름 안전하게 처리
	// SQLite에서 하이픈이 포함된 이름 처리
	safeIndexName := fmt.Sprintf("idx_%s_user_id", strings.Replace(tableName, "-", "_", -1))

	// 인덱스 생성
	createIndexQuery := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (user_id)",
		safeIndexName, safeTableName)
	_, err = config.TargetDB.Exec(createIndexQuery)
	if err != nil {
		return fmt.Errorf("인덱스 생성 실패: %w", err)
	}

	return nil
}

// migrateTableData는 테이블 데이터를 마이그레이션합니다
func migrateTableData(config *DataMigrationConfig, tableName string, fieldTypeMap map[string]models.FieldType) error {
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

	// 테이블 컬럼 타입 정보 준비
	columnDataTypeMap := make(map[string]string)
	for _, col := range targetColumns {
		columnDataTypeMap[col.Name] = col.Type
	}

	// 알려진 boolean 필드 - 동적 테이블에서 일반적으로 boolean으로 처리해야 하는 필드
	booleanFields := map[string]bool{
		"is_private": true,
	}

	// 데이터 배치 처리
	batchSize := config.BatchSize
	if batchSize <= 0 {
		batchSize = 1000
	}

	totalBatches := (totalRows + batchSize - 1) / batchSize
	processedRows := 0

	// 외래 키 제약 조건 처리
	if config.TargetDBConfig.DBDriver == "mysql" || config.TargetDBConfig.DBDriver == "mariadb" {
		_, err = config.TargetDB.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 0;")
		if err != nil {
			return fmt.Errorf("외래 키 제약 비활성화 실패: %w", err)
		}
		defer func() {
			_, _ = config.TargetDB.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 1;")
		}()
	} else if config.TargetDBConfig.DBDriver == "postgres" {
		_, err = config.TargetDB.ExecContext(ctx, "SET session_replication_role = 'replica';")
		if err != nil {
			return fmt.Errorf("외래 키 제약 비활성화 실패: %w", err)
		}
		defer func() {
			_, _ = config.TargetDB.ExecContext(ctx, "SET session_replication_role = 'origin';")
		}()
	} else if config.TargetDBConfig.DBDriver == "sqlite" {
		_, err = config.TargetDB.ExecContext(ctx, "PRAGMA foreign_keys = OFF;")
		if err != nil {
			return fmt.Errorf("외래 키 제약 비활성화 실패: %w", err)
		}
		defer func() {
			_, _ = config.TargetDB.ExecContext(ctx, "PRAGMA foreign_keys = ON;")
		}()
	}

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

	// 테이블 특성에 따른 정렬 키 설정
	var orderByColumn string

	// 복합 기본 키를 가진 테이블들 처리
	if tableName == "board_participants" || tableName == "board_managers" {
		orderByColumn = "board_id, user_id"
	} else if tableName == "qna_question_votes" {
		orderByColumn = "question_id, user_id"
	} else if tableName == "qna_answer_votes" {
		orderByColumn = "answer_id, user_id"
	} else {
		// 기본값은 id 컬럼 사용
		orderByColumn = "id"
	}

	for batchNum := 0; batchNum < totalBatches; batchNum++ {
		offset := batchNum * batchSize
		limit := batchSize

		if config.VerboseLogging && totalBatches > 1 {
			fmt.Printf("    배치 %d/%d 처리 중... (오프셋: %d, 한계: %d)\n", batchNum+1, totalBatches, offset, limit)
		}

		// 소스 데이터 쿼리 구성
		sourceQuery := fmt.Sprintf("SELECT %s FROM %s ORDER BY %s LIMIT %d OFFSET %d",
			strings.Join(sourceColumnNames, ", "),
			quoteTableName(config.SourceDBConfig.DBDriver, tableName),
			orderByColumn,
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

		// 배치 내 오류 추적
		batchError := false

		// 행별 처리
		rowsInBatch := 0
		for sourceRows.Next() {
			// 행 데이터 변수 준비
			rowValues := make([]any, len(commonColumns))
			valuePtrs := make([]any, len(commonColumns))
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

				// 대상 DB 드라이버에 따라 다른 파라미터 바인딩 사용
				var checkQuery string
				if config.TargetDBConfig.DBDriver == "postgres" {
					checkQuery = fmt.Sprintf("SELECT 1 FROM %s WHERE id = $1 LIMIT 1",
						quoteTableName(config.TargetDBConfig.DBDriver, tableName))
				} else {
					checkQuery = fmt.Sprintf("SELECT 1 FROM %s WHERE id = ? LIMIT 1",
						quoteTableName(config.TargetDBConfig.DBDriver, tableName))
				}

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

				// MySQL/PostgreSQL에서 가져온 인코딩된 데이터 처리
				if config.SourceDBConfig.DBDriver == "mysql" || config.SourceDBConfig.DBDriver == "postgres" {
					// 문자열 데이터 처리
					if strVal, ok := val.(string); ok && strings.HasSuffix(strVal, "==") {
						// Base64 인코딩된 문자열로 보이는 경우 디코딩 시도
						decoded, err := tryBase64Decode(strVal)
						if err == nil {
							val = decoded
						}
					} else if bytes, ok := val.([]byte); ok {
						// 바이트 배열인 경우 문자열로 변환
						val = string(bytes)
					}
				}

				// ===== 필드 타입 처리 로직 수정 =====
				// 대상 DB가 PostgreSQL이고 필드 타입이 boolean인 경우 처리
				if config.TargetDBConfig.DBDriver == "postgres" {
					targetColumnType := columnDataTypeMap[colName]
					isBooleanField := false

					// 1. 컬럼 타입이 boolean인지 확인
					if strings.Contains(targetColumnType, "bool") {
						isBooleanField = true
					}

					// 2. 알려진 boolean 필드인지 확인
					if booleanFields[colName] {
						isBooleanField = true
					}

					// 3. 동적 테이블의 경우 필드 타입 정보 활용
					if fieldTypeMap != nil {
						if fieldType, exists := fieldTypeMap[colName]; exists && fieldType == models.FieldTypeCheckbox {
							isBooleanField = true
						}
					}

					// boolean 필드면 적절히 변환
					if isBooleanField {
						switch v := val.(type) {
						case int64:
							if v == 1 {
								columnValues = append(columnValues, "TRUE")
							} else {
								columnValues = append(columnValues, "FALSE")
							}
							continue
						case int:
							if v == 1 {
								columnValues = append(columnValues, "TRUE")
							} else {
								columnValues = append(columnValues, "FALSE")
							}
							continue
						case int32:
							if v == 1 {
								columnValues = append(columnValues, "TRUE")
							} else {
								columnValues = append(columnValues, "FALSE")
							}
							continue
						case int16:
							if v == 1 {
								columnValues = append(columnValues, "TRUE")
							} else {
								columnValues = append(columnValues, "FALSE")
							}
							continue
						case int8:
							if v == 1 {
								columnValues = append(columnValues, "TRUE")
							} else {
								columnValues = append(columnValues, "FALSE")
							}
							continue
						case uint64, uint32, uint16, uint8, uint:
							// 숫자를 문자열로 변환하여 비교
							valStr := fmt.Sprintf("%v", v)
							if valStr == "1" {
								columnValues = append(columnValues, "TRUE")
							} else {
								columnValues = append(columnValues, "FALSE")
							}
							continue
						case string:
							lowerVal := strings.ToLower(v)
							if lowerVal == "1" || lowerVal == "true" || lowerVal == "t" || lowerVal == "y" || lowerVal == "yes" {
								columnValues = append(columnValues, "TRUE")
							} else {
								columnValues = append(columnValues, "FALSE")
							}
							continue
						case bool:
							if v {
								columnValues = append(columnValues, "TRUE")
							} else {
								columnValues = append(columnValues, "FALSE")
							}
							continue
						}
					}
				}

				// 모델 기반 타입 변환
				fieldInfo, hasField := modelInfo[colName]

				// 데이터베이스별 타입 처리
				switch config.TargetDBConfig.DBDriver {
				case "postgres":
					// PostgreSQL 타입 처리
					if hasField {
						switch fieldInfo.fieldType {
						case "bool":
							// SQLite나 MySQL의 boolean 값을 PostgreSQL boolean으로 변환
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
								// 문자열로 된 시간 포맷팅 시도
								t, err := parseTimeString(v)
								if err == nil {
									columnValues = append(columnValues, fmt.Sprintf("'%s'", t.Format("2006-01-02 15:04:05")))
								} else {
									columnValues = append(columnValues, fmt.Sprintf("'%s'", v))
								}
							default:
								columnValues = append(columnValues, fmt.Sprintf("'%v'", v))
							}
							continue
						}
					}

				case "mysql", "mariadb":
					// MySQL/MariaDB 타입 처리
					if hasField {
						switch fieldInfo.fieldType {
						case "bool":
							// Boolean 값을 MySQL의 1/0으로 변환
							switch v := val.(type) {
							case int64:
								if v == 1 {
									columnValues = append(columnValues, "1")
								} else {
									columnValues = append(columnValues, "0")
								}
							case int:
								if v == 1 {
									columnValues = append(columnValues, "1")
								} else {
									columnValues = append(columnValues, "0")
								}
							case string:
								if strings.ToLower(v) == "true" || v == "1" {
									columnValues = append(columnValues, "1")
								} else {
									columnValues = append(columnValues, "0")
								}
							case bool:
								if v {
									columnValues = append(columnValues, "1")
								} else {
									columnValues = append(columnValues, "0")
								}
							default:
								columnValues = append(columnValues, "0")
							}
							continue
						case "time.Time":
							// 시간 처리
							switch v := val.(type) {
							case time.Time:
								columnValues = append(columnValues, fmt.Sprintf("'%s'", v.Format("2006-01-02 15:04:05")))
							case string:
								// 문자열로 된 시간 포맷팅 시도
								t, err := parseTimeString(v)
								if err == nil {
									columnValues = append(columnValues, fmt.Sprintf("'%s'", t.Format("2006-01-02 15:04:05")))
								} else {
									columnValues = append(columnValues, fmt.Sprintf("'%s'", v))
								}
							default:
								columnValues = append(columnValues, fmt.Sprintf("'%v'", v))
							}
							continue
						}
					}

				case "sqlite":
					// SQLite 타입 처리
					if hasField {
						switch fieldInfo.fieldType {
						case "bool":
							// Boolean 값을 SQLite의 0/1로 변환
							switch v := val.(type) {
							case int64:
								if v == 1 {
									columnValues = append(columnValues, "1")
								} else {
									columnValues = append(columnValues, "0")
								}
							case int:
								if v == 1 {
									columnValues = append(columnValues, "1")
								} else {
									columnValues = append(columnValues, "0")
								}
							case string:
								if strings.ToLower(v) == "true" || v == "1" {
									columnValues = append(columnValues, "1")
								} else {
									columnValues = append(columnValues, "0")
								}
							case bool:
								if v {
									columnValues = append(columnValues, "1")
								} else {
									columnValues = append(columnValues, "0")
								}
							default:
								columnValues = append(columnValues, "0")
							}
							continue
						}
					}
				}

				// 일반 데이터 타입 변환 - 이스케이프 로직 강화
				switch v := val.(type) {
				case string:
					// 문자열은 작은 따옴표로 감싸고 내부 작은 따옴표는 이스케이프
					escapedVal := strings.Replace(v, "'", "''", -1)

					// 대상 DB가 SQLite이고 HTML 내용이 있는 경우 추가 처리
					if config.TargetDBConfig.DBDriver == "sqlite" && (strings.Contains(v, "<") || strings.Contains(v, ">")) {
						columnValues = append(columnValues, fmt.Sprintf("'%s'", escapedVal))
					} else {
						columnValues = append(columnValues, fmt.Sprintf("'%s'", escapedVal))
					}
				case []byte:
					// 바이트 배열을 문자열로 변환 후 이스케이프
					strVal := string(v)
					escapedVal := strings.Replace(strVal, "'", "''", -1)
					columnValues = append(columnValues, fmt.Sprintf("'%s'", escapedVal))
				case bool:
					// DB 종류에 따라 불리언 처리
					if config.TargetDBConfig.DBDriver == "postgres" {
						if v {
							columnValues = append(columnValues, "TRUE")
						} else {
							columnValues = append(columnValues, "FALSE")
						}
					} else {
						// MySQL, SQLite는 1/0 사용
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
			var execErr error
			if useTransaction {
				_, execErr = tx.ExecContext(ctx, directSQL)
			} else {
				_, execErr = config.TargetDB.ExecContext(ctx, directSQL)
			}

			if execErr != nil {
				// 행 실패 기록
				batchError = true
				if config.VerboseLogging {
					rowJSON, _ := json.Marshal(rowValues)
					log.Printf("행 삽입 실패: %v\n데이터: %s\n쿼리: %s", execErr, rowJSON, directSQL)
				}

				// 에러 수집하되 진행 계속
				config.addError(fmt.Errorf("행 삽입 실패 (테이블: %s): %w", tableName, execErr))

				// 최대 오류 수 초과 시 중단
				if len(config.Errors) > config.MaxErrorsBeforeExit {
					sourceRows.Close()
					if useTransaction {
						tx.Rollback()
					}
					return fmt.Errorf("최대 오류 수 초과로 마이그레이션 중단: %w", execErr)
				}

				// 트랜잭션 사용 중인 경우 롤백 후 계속
				if useTransaction {
					tx.Rollback()

					// 새 트랜잭션 시작
					tx, err = config.TargetDB.DB.Begin()
					if err != nil {
						sourceRows.Close()
						return fmt.Errorf("트랜잭션 재시작 실패: %w", err)
					}
				}

				continue
			}

			rowsInBatch++
		}

		sourceRows.Close()

		// 트랜잭션 커밋 (필요 시 그리고 오류가 없을 때만)
		if useTransaction && !batchError {
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

// getModelInfo는 테이블 이름으로부터 해당 모델의 필드 정보를 가져옵니다
func getModelInfo(tableName string) map[string]FieldInfo {
	// DB 컬럼 이름 -> 필드 정보 맵핑
	fieldInfoMap := make(map[string]FieldInfo)

	// 테이블 이름에 따라 적절한 모델 구조체 선택
	var modelType reflect.Type
	switch tableName {
	case "users":
		modelType = reflect.TypeOf(models.User{})
	case "boards":
		modelType = reflect.TypeOf(models.Board{})
	case "board_fields":
		modelType = reflect.TypeOf(models.BoardField{})
	case "board_managers":
		modelType = reflect.TypeOf(models.BoardManager{})
	case "board_participants":
		modelType = reflect.TypeOf(models.BoardParticipant{})
	case "comments":
		modelType = reflect.TypeOf(models.Comment{})
	case "attachments":
		modelType = reflect.TypeOf(models.Attachment{})
	case "post_votes":
		modelType = reflect.TypeOf(models.PostVote{})
	case "comment_votes":
		modelType = reflect.TypeOf(models.CommentVote{})
	case "qna_answers":
		modelType = reflect.TypeOf(models.Answer{})
	case "qna_question_votes":
		modelType = reflect.TypeOf(models.QuestionVote{})
	case "qna_answer_votes":
		modelType = reflect.TypeOf(models.AnswerVote{})
	case "referrer_stats":
		modelType = reflect.TypeOf(models.ReferrerStat{})
	case "system_settings":
		modelType = reflect.TypeOf(models.SystemSetting{})
	default:
		// 필수 기본 필드 추가
		fieldInfoMap["id"] = FieldInfo{fieldName: "ID", fieldType: "int64"}
		fieldInfoMap["title"] = FieldInfo{fieldName: "Title", fieldType: "string"}
		fieldInfoMap["content"] = FieldInfo{fieldName: "Content", fieldType: "string"}
		fieldInfoMap["user_id"] = FieldInfo{fieldName: "UserID", fieldType: "int64"}
		fieldInfoMap["view_count"] = FieldInfo{fieldName: "ViewCount", fieldType: "int"}
		fieldInfoMap["comment_count"] = FieldInfo{fieldName: "CommentCount", fieldType: "int"}
		fieldInfoMap["like_count"] = FieldInfo{fieldName: "LikeCount", fieldType: "int"}
		fieldInfoMap["dislike_count"] = FieldInfo{fieldName: "DislikeCount", fieldType: "int"}
		fieldInfoMap["created_at"] = FieldInfo{fieldName: "CreatedAt", fieldType: "time.Time"}
		fieldInfoMap["updated_at"] = FieldInfo{fieldName: "UpdatedAt", fieldType: "time.Time"}

		return fieldInfoMap
	}

	// 모델 필드 분석
	for i := range modelType.NumField() {
		field := modelType.Field(i)

		// 임베디드 필드는 건너뛰기 (bun.BaseModel 등)
		if field.Anonymous {
			continue
		}

		// bun 태그 파싱
		bunTag := field.Tag.Get("bun")
		if bunTag == "" || bunTag == "-" {
			continue
		}

		// 태그 파싱
		tagMap := parseTags(bunTag)

		// 컬럼 이름 가져오기
		columnName := ""
		if tagName, ok := tagMap["column"]; ok {
			columnName = tagName
		} else {
			// 기본적으로 필드 이름을 스네이크 케이스로 변환
			columnName = toSnakeCase(field.Name)
		}

		// 필드 타입 가져오기
		fieldType := field.Type.String()

		// FieldInfo 생성 및 맵에 추가
		fieldInfoMap[columnName] = FieldInfo{
			fieldName: field.Name,
			fieldType: fieldType,
			tags:      tagMap,
		}
	}

	return fieldInfoMap
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
		// SQLite용 테이블 존재 여부 확인 및 컬럼 정보 조회
		var count int
		checkQuery := fmt.Sprintf("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='%s'",
			strings.Replace(tableName, "'", "''", -1))
		err := db.QueryRow(checkQuery).Scan(&count)
		if err != nil {
			return nil, fmt.Errorf("테이블 존재 여부 확인 실패: %w", err)
		}

		if count == 0 {
			return nil, fmt.Errorf("테이블이 존재하지 않습니다: %s", tableName)
		}

		// 테이블 이름 준비
		safeTableName := tableName
		if strings.Contains(tableName, "-") || strings.Contains(tableName, ".") {
			safeTableName = fmt.Sprintf("\"%s\"", tableName)
		}

		// PRAGMA table_info 쿼리 실행
		rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", safeTableName))
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

// validateSourceData는 마이그레이션 전 데이터 검증
func validateSourceData(config *DataMigrationConfig, tableName string) error {
	// 테이블 존재 여부 확인
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s LIMIT 1", quoteTableName(config.SourceDBConfig.DBDriver, tableName))

	err := config.SourceDB.QueryRow(query).Scan(&count)
	if err != nil {
		return fmt.Errorf("테이블 '%s' 검증 실패: %w", tableName, err)
	}

	return nil
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
	// 외래 키 제약 조건이 있거나 이력 데이터인 테이블은 정리하지 않음
	skipCleanTables := map[string]bool{
		"referrer_stats": true,  // 대량 데이터이므로 건너뜀
		"users":          false, // 사용자 데이터는 초기화
	}

	if skip, exists := skipCleanTables[tableName]; exists {
		return !skip
	}

	// 기본값: 테이블 데이터 정리
	return true
}

// isTableSkipped는 주어진 테이블이 건너뛰기 목록에 있는지 확인합니다
func isTableSkipped(tableName string, skipTables []string) bool {
	return slices.Contains(skipTables, tableName)
}

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

// needsSequenceReset은 테이블에 시퀀스 재설정이 필요한지 확인합니다
func needsSequenceReset(tableName string) bool {
	// 시퀀스 재설정이 필요 없는 테이블
	noResetTables := map[string]bool{
		"goose_db_version":   true,
		"referrer_stats":     true,
		"board_managers":     true,
		"board_participants": true,
		"qna_question_votes": true,
		"qna_answer_votes":   true,
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

// updateCommentCounts는 모든 게시판의 게시물에 대해 댓글 수를 업데이트합니다
func updateCommentCounts(config *DataMigrationConfig) error {
	fmt.Println("[5/5] 댓글 수 업데이트 중...")

	ctx := context.Background()

	// 1. 모든 게시판 목록 가져오기
	var boards []*models.Board
	err := config.TargetDB.NewSelect().
		Model(&boards).
		Column("id", "table_name").
		Scan(ctx)

	if err != nil {
		return fmt.Errorf("게시판 목록 조회 실패: %w", err)
	}

	// 2. 각 게시판별로 댓글 수 업데이트
	for _, board := range boards {
		// 게시판의 모든 게시물 ID 가져오기
		var postIDs []int64
		query := fmt.Sprintf("SELECT id FROM %s", quoteTableName(config.TargetDBConfig.DBDriver, board.TableName))

		// Scan 사용 시 rows 변수에 결과를 받아야 함
		rows, err := config.TargetDB.QueryContext(ctx, query)
		if err != nil {
			fmt.Printf("게시판 '%s' 게시물 ID 조회 실패: %v\n", board.TableName, err)
			continue
		}
		defer rows.Close()

		for rows.Next() {
			var postID int64
			if err := rows.Scan(&postID); err != nil {
				fmt.Printf("게시물 ID 스캔 실패: %v\n", err)
				continue
			}
			postIDs = append(postIDs, postID)
		}

		// 각 게시물의 댓글 수 계산 및 업데이트
		for _, postID := range postIDs {
			// 댓글 수 계산
			count, err := countCommentsForPost(config, board.ID, postID)
			if err != nil {
				fmt.Printf("게시물 ID %d의 댓글 수 계산 실패: %v\n", postID, err)
				continue
			}

			// 댓글 수 업데이트
			updateQuery := fmt.Sprintf("UPDATE %s SET comment_count = ? WHERE id = ?",
				quoteTableName(config.TargetDBConfig.DBDriver, board.TableName))
			_, err = config.TargetDB.ExecContext(ctx, updateQuery, count, postID)
			if err != nil {
				fmt.Printf("게시물 ID %d의 댓글 수 업데이트 실패: %v\n", postID, err)
			}
		}

		fmt.Printf("게시판 '%s'의 댓글 수 업데이트 완료\n", board.TableName)
	}

	fmt.Println("댓글 수 업데이트 완료")
	return nil
}

// countCommentsForPost는 특정 게시물의 댓글 수를 계산합니다
func countCommentsForPost(config *DataMigrationConfig, boardID, postID int64) (int, error) {
	count, err := config.TargetDB.NewSelect().
		Table("comments").
		Where("board_id = ? AND post_id = ?", boardID, postID).
		Count(context.Background())

	return count, err
}

// updatePostVoteCounts는 게시물의 좋아요/싫어요 수를 업데이트합니다
func updatePostVoteCounts(config *DataMigrationConfig) error {
	fmt.Println("[6/6] 게시물 좋아요/싫어요 수 업데이트 중...")

	ctx := context.Background()

	// 1. 모든 게시판 목록 가져오기
	var boards []*models.Board
	err := config.TargetDB.NewSelect().
		Model(&boards).
		Column("id", "table_name", "votes_enabled").
		Scan(ctx)

	if err != nil {
		return fmt.Errorf("게시판 목록 조회 실패: %w", err)
	}

	// 2. 각 게시판별로 게시물 좋아요/싫어요 수 업데이트
	for _, board := range boards {
		if !board.VotesEnabled {
			if config.VerboseLogging {
				fmt.Printf("게시판 '%s'는 좋아요/싫어요 기능이 비활성화되어 있습니다\n", board.TableName)
			}
			continue
		}

		// 게시판의 모든 게시물 ID 가져오기
		// var postIDs []int64
		query := fmt.Sprintf("SELECT id FROM %s", quoteTableName(config.TargetDBConfig.DBDriver, board.TableName))

		rows, err := config.TargetDB.QueryContext(ctx, query)
		if err != nil {
			fmt.Printf("게시판 '%s' 게시물 ID 조회 실패: %v\n", board.TableName, err)
			continue
		}

		var postsToUpdate []int64
		for rows.Next() {
			var postID int64
			if err := rows.Scan(&postID); err != nil {
				fmt.Printf("게시물 ID 스캔 실패: %v\n", err)
				continue
			}
			postsToUpdate = append(postsToUpdate, postID)
		}
		rows.Close()

		// 각 게시물의 좋아요/싫어요 수 계산 및 업데이트
		for _, postID := range postsToUpdate {
			// 좋아요 수 계산
			likeCount, err := countPostVotes(config, postID, 1)
			if err != nil {
				fmt.Printf("게시물 ID %d의 좋아요 수 계산 실패: %v\n", postID, err)
				continue
			}

			// 싫어요 수 계산
			dislikeCount, err := countPostVotes(config, postID, -1)
			if err != nil {
				fmt.Printf("게시물 ID %d의 싫어요 수 계산 실패: %v\n", postID, err)
				continue
			}

			// 좋아요/싫어요 수 업데이트
			updateQuery := fmt.Sprintf("UPDATE %s SET like_count = ?, dislike_count = ? WHERE id = ?",
				quoteTableName(config.TargetDBConfig.DBDriver, board.TableName))
			_, err = config.TargetDB.ExecContext(ctx, updateQuery, likeCount, dislikeCount, postID)
			if err != nil {
				fmt.Printf("게시물 ID %d의 좋아요/싫어요 수 업데이트 실패: %v\n", postID, err)
			}
		}

		if config.VerboseLogging {
			fmt.Printf("게시판 '%s'의 게시물 좋아요/싫어요 수 업데이트 완료 (%d개)\n", board.TableName, len(postsToUpdate))
		}
	}

	fmt.Println("게시물 좋아요/싫어요 수 업데이트 완료")
	return nil
}

// countPostVotes는 특정 게시물의 좋아요 또는 싫어요 수를 계산합니다
func countPostVotes(config *DataMigrationConfig, postID int64, value int) (int, error) {
	count, err := config.TargetDB.NewSelect().
		Table("post_votes").
		Where("post_id = ? AND value = ?", postID, value).
		Count(context.Background())

	return count, err
}

// updateCommentVoteCounts는 댓글의 좋아요/싫어요 수를 업데이트합니다
func updateCommentVoteCounts(config *DataMigrationConfig) error {
	fmt.Println("[7/7] 댓글 좋아요/싫어요 수 업데이트 중...")

	ctx := context.Background()

	// 모든 댓글 ID 가져오기
	var commentIDs []int64
	err := config.TargetDB.NewSelect().
		Model((*models.Comment)(nil)).
		Column("id").
		Scan(ctx, &commentIDs)

	if err != nil {
		return fmt.Errorf("댓글 ID 목록 조회 실패: %w", err)
	}

	// 각 댓글의 좋아요/싫어요 수 계산 및 업데이트
	for _, commentID := range commentIDs {
		// 좋아요 수 계산
		likeCount, err := countCommentVotes(config, commentID, 1)
		if err != nil {
			fmt.Printf("댓글 ID %d의 좋아요 수 계산 실패: %v\n", commentID, err)
			continue
		}

		// 싫어요 수 계산
		dislikeCount, err := countCommentVotes(config, commentID, -1)
		if err != nil {
			fmt.Printf("댓글 ID %d의 싫어요 수 계산 실패: %v\n", commentID, err)
			continue
		}

		// 좋아요/싫어요 수 업데이트
		_, err = config.TargetDB.NewUpdate().
			Model((*models.Comment)(nil)).
			Set("like_count = ?", likeCount).
			Set("dislike_count = ?", dislikeCount).
			Where("id = ?", commentID).
			Exec(ctx)

		if err != nil {
			fmt.Printf("댓글 ID %d의 좋아요/싫어요 수 업데이트 실패: %v\n", commentID, err)
		}
	}

	fmt.Println("댓글 좋아요/싫어요 수 업데이트 완료")
	return nil
}

// countCommentVotes는 특정 댓글의 좋아요 또는 싫어요 수를 계산합니다
func countCommentVotes(config *DataMigrationConfig, commentID int64, value int) (int, error) {
	count, err := config.TargetDB.NewSelect().
		Table("comment_votes").
		Where("comment_id = ? AND value = ?", commentID, value).
		Count(context.Background())

	return count, err
}
