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

	"github.com/edp1096/go-board/internal/models"
	"github.com/edp1096/go-board/internal/repository"
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

		// 테이블 데이터 검증
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

// FieldInfo는 모델 필드의 메타데이터를 저장합니다
type FieldInfo struct {
	fieldName string
	fieldType string
	tags      map[string]string
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
	case "comments":
		modelType = reflect.TypeOf(models.Comment{})
	case "attachments":
		modelType = reflect.TypeOf(models.Attachment{})
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
		// 동적 테이블은 기본 PostCommon 구조체 사용
		// modelType = reflect.TypeOf(models.PostCommon{})
		// 필수 기본 필드 추가
		fieldInfoMap["id"] = FieldInfo{fieldName: "ID", fieldType: "int64"}
		fieldInfoMap["title"] = FieldInfo{fieldName: "Title", fieldType: "string"}
		fieldInfoMap["content"] = FieldInfo{fieldName: "Content", fieldType: "string"}
		fieldInfoMap["user_id"] = FieldInfo{fieldName: "UserID", fieldType: "int64"}
		fieldInfoMap["view_count"] = FieldInfo{fieldName: "ViewCount", fieldType: "int"}
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

// parseTags는 bun 태그를 파싱하여 맵으로 반환합니다
func parseTags(tag string) map[string]string {
	tagMap := make(map[string]string)

	// 태그 파싱
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		// column:name 형식 처리
		if strings.Contains(part, ":") {
			kv := strings.SplitN(part, ":", 2)
			tagMap[kv[0]] = kv[1]
		} else if strings.Contains(part, "=") {
			// column=name 형식 처리
			kv := strings.SplitN(part, "=", 2)
			tagMap[kv[0]] = kv[1]
		} else {
			// 단일 태그 처리 (pk, notnull 등)
			tagMap[part] = "true"
		}
	}

	return tagMap
}

// toSnakeCase는 캐멀 케이스 문자열을 스네이크 케이스로 변환합니다
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
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

// quoteColumnName은 컬럼 이름을 인용 부호로 묶습니다
func quoteColumnName(columnName string) string {
	return columnName
}

// validateSourceData는 마이그레이션 전 데이터 검증
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
		"referrer_stats": true,
		// "users":          true,
	}

	return !skipCleanTables[tableName]
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
		"goose_db_version": true,
		"referrer_stats":   true,
		"board_managers":   true,
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
