// internal/service/dynamic_board_service.go
package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-board/internal/models"
	"go-board/internal/utils"
	"regexp"
	"strings"
	"time"

	"github.com/uptrace/bun"
)

type DynamicBoardService interface {
	// 게시판 테이블 생성
	CreateBoardTable(ctx context.Context, board *models.Board, fields []*models.BoardField) error

	// 게시판 테이블 변경 (필드 추가/수정/삭제)
	AlterBoardTable(ctx context.Context, board *models.Board, addFields, modifyFields []*models.BoardField, dropFields []string) error

	// 게시판 테이블 삭제
	DropBoardTable(ctx context.Context, tableName string) error

	// 게시판 스키마 정보 조회
	GetBoardTableSchema(ctx context.Context, tableName string) ([]*models.BoardField, error)
}

type dynamicBoardService struct {
	db *bun.DB
}

func NewDynamicBoardService(db *bun.DB) DynamicBoardService {
	return &dynamicBoardService{db: db}
}

// 필드 타입에 따른 SQL 데이터 타입 반환
func (s *dynamicBoardService) getColumnType(fieldType models.FieldType) string {
	// 유효한 필드 타입만 허용
	switch fieldType {
	case models.FieldTypeText:
		return "VARCHAR(255)"
	case models.FieldTypeTextarea:
		return "TEXT"
	case models.FieldTypeNumber:
		return "INTEGER"
	case models.FieldTypeDate:
		return "DATE"
	case models.FieldTypeSelect:
		return "VARCHAR(100)"
	case models.FieldTypeCheckbox:
		return "BOOLEAN DEFAULT FALSE"
	case models.FieldTypeFile:
		return "VARCHAR(255)"
	default:
		// 안전한 기본값으로 처리
		return "VARCHAR(255)"
	}
}

// 열 정의 문자열 생성
func (s *dynamicBoardService) getColumnDefinition(field *models.BoardField) string {
	columnType := s.getColumnType(field.FieldType)
	columnDef := fmt.Sprintf("%s %s", field.ColumnName, columnType)

	// NOT NULL 옵션 추가
	if field.Required {
		columnDef += " NOT NULL"
	}

	return columnDef
}

// 기본 테이블 컬럼 정의 반환
func (s *dynamicBoardService) getBaseColumns() []string {
	var idType string
	switch {
	case utils.IsPostgres(s.db):
		idType = "SERIAL PRIMARY KEY"
	case utils.IsSQLite(s.db):
		idType = "INTEGER PRIMARY KEY AUTOINCREMENT"
	case utils.IsMySQL(s.db):
		idType = "INT AUTO_INCREMENT PRIMARY KEY"
	default:
		// 기본적으로 MySQL/MariaDB로 처리
		idType = "INT AUTO_INCREMENT PRIMARY KEY"
	}

	return []string{
		fmt.Sprintf("id %s", idType),
		"title VARCHAR(200) NOT NULL",
		"content TEXT NOT NULL",
		"user_id INTEGER NOT NULL REFERENCES users(id)",
		"view_count INTEGER NOT NULL DEFAULT 0",
		"created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP",
		"updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP",
	}
}

// 테이블 이름 유효성 검사를 위한 정규식
var tableNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// 컬럼 이름 유효성 검사를 위한 정규식
var columnNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// 시스템 예약 컬럼 이름 목록
var reservedColumnNames = []string{
	"id", "title", "content", "user_id", "view_count", "created_at", "updated_at",
}

// 컬럼 이름이 예약되어 있는지 확인하는 함수
func isReservedColumnName(name string) bool {
	for _, reserved := range reservedColumnNames {
		if reserved == name {
			return true
		}
	}
	return false
}

func (s *dynamicBoardService) CreateBoardTable(ctx context.Context, board *models.Board, fields []*models.BoardField) error {
	// 테이블 이름 유효성 검사
	if board.TableName == "" {
		return errors.New("테이블 이름이 비어 있습니다")
	}

	// SQL 인젝션 방지를 위한 테이블 이름 검증
	if !tableNameRegex.MatchString(board.TableName) {
		return fmt.Errorf("유효하지 않은 테이블 이름입니다: %s (영문자, 숫자, 언더스코어만 허용)", board.TableName)
	}

	// 필드 이름 유효성 검사
	for _, field := range fields {
		if !columnNameRegex.MatchString(field.ColumnName) {
			return fmt.Errorf("유효하지 않은 컬럼 이름입니다: %s (영문자, 숫자, 언더스코어만 허용)", field.ColumnName)
		}

		// 예약된 컬럼 이름 검사
		if isReservedColumnName(field.ColumnName) {
			return fmt.Errorf("'%s'는 시스템에서 예약된 필드 이름입니다. 다른 이름을 사용해주세요", field.ColumnName)
		}
	}

	// 기본 컬럼 정의
	columns := s.getBaseColumns()

	// 동적 필드 SQL 생성
	for _, field := range fields {
		columnDef := s.getColumnDefinition(field)
		columns = append(columns, columnDef)
	}

	// CREATE TABLE 쿼리 생성
	var query string
	if utils.IsPostgres(s.db) {
		// PostgreSQL에서는 큰따옴표로 테이블 이름을 감싸서 예약어와의 충돌 방지
		query = fmt.Sprintf(
			"CREATE TABLE \"%s\" (%s);",
			board.TableName,
			strings.Join(columns, ", "),
		)
	} else {
		// MariaDB/MySQL에서는 백틱(`)으로 테이블 이름을 감싸서 예약어와의 충돌 방지
		query = fmt.Sprintf(
			"CREATE TABLE `%s` (%s);",
			board.TableName,
			strings.Join(columns, ", "),
		)
	}

	// 쿼리 실행
	_, err := s.db.ExecContext(ctx, query)
	return err
}

func (s *dynamicBoardService) AlterBoardTable(ctx context.Context, board *models.Board, addFields, modifyFields []*models.BoardField, dropFields []string) error {
	// SQL 인젝션 방지를 위한 테이블 이름 검증
	if !tableNameRegex.MatchString(board.TableName) {
		return fmt.Errorf("유효하지 않은 테이블 이름입니다: %s (영문자, 숫자, 언더스코어만 허용)", board.TableName)
	}

	// SQLite인 경우 다른 방식으로 처리
	if utils.IsSQLite(s.db) {
		return s.alterBoardTableSQLite(ctx, board, addFields, modifyFields, dropFields)
	}

	// 트랜잭션 시작
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. 필드 추가
	for _, field := range addFields {
		// 컬럼 이름 유효성 검사
		if !columnNameRegex.MatchString(field.ColumnName) {
			return fmt.Errorf("유효하지 않은 컬럼 이름입니다: %s (영문자, 숫자, 언더스코어만 허용)", field.ColumnName)
		}

		// 예약된 컬럼 이름 검사
		if isReservedColumnName(field.ColumnName) {
			return fmt.Errorf("'%s'는 시스템에서 예약된 필드 이름입니다. 다른 이름을 사용해주세요", field.ColumnName)
		}

		columnDef := s.getColumnDefinition(field)
		var query string
		if utils.IsPostgres(s.db) {
			query = fmt.Sprintf("ALTER TABLE \"%s\" ADD COLUMN %s;", board.TableName, columnDef)
		} else {
			query = fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN %s;", board.TableName, columnDef)
		}

		_, err := tx.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("필드 추가 실패 (%s): %w", field.Name, err)
		}
	}

	// 2. 필드 수정 (PostgreSQL과 MySQL/MariaDB 방식이 다름)
	for _, field := range modifyFields {
		// 컬럼 이름 유효성 검사
		if field.ID <= 0 && !columnNameRegex.MatchString(field.ColumnName) {
			return fmt.Errorf("유효하지 않은 컬럼 이름입니다: %s (영문자, 숫자, 언더스코어만 허용)", field.ColumnName)
		}

		var query string
		columnType := s.getColumnType(field.FieldType)

		if utils.IsPostgres(s.db) {
			// PostgreSQL용 ALTER COLUMN
			query = fmt.Sprintf(
				"ALTER TABLE \"%s\" ALTER COLUMN \"%s\" TYPE %s;",
				board.TableName,
				field.ColumnName,
				columnType,
			)

			// NOT NULL 제약 조건 처리 (별도 쿼리)
			if field.Required {
				query += fmt.Sprintf(
					"ALTER TABLE \"%s\" ALTER COLUMN \"%s\" SET NOT NULL;",
					board.TableName,
					field.ColumnName,
				)
			} else {
				query += fmt.Sprintf(
					"ALTER TABLE \"%s\" ALTER COLUMN \"%s\" DROP NOT NULL;",
					board.TableName,
					field.ColumnName,
				)
			}
		} else {
			// MySQL/MariaDB용 MODIFY COLUMN
			if field.Required {
				columnType += " NOT NULL"
			}

			query = fmt.Sprintf(
				"ALTER TABLE `%s` MODIFY COLUMN `%s` %s;",
				board.TableName,
				field.ColumnName,
				columnType,
			)
		}

		_, err := tx.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("필드 수정 실패 (%s): %w", field.Name, err)
		}
	}

	// 3. 필드 삭제
	for _, columnName := range dropFields {
		// 컬럼 이름 유효성 검사
		if !columnNameRegex.MatchString(columnName) {
			return fmt.Errorf("유효하지 않은 컬럼 이름입니다: %s (영문자, 숫자, 언더스코어만 허용)", columnName)
		}

		var query string
		if utils.IsPostgres(s.db) {
			query = fmt.Sprintf("ALTER TABLE \"%s\" DROP COLUMN \"%s\";", board.TableName, columnName)
		} else {
			query = fmt.Sprintf("ALTER TABLE `%s` DROP COLUMN `%s`;", board.TableName, columnName)
		}
		_, err := tx.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("필드 삭제 실패 (%s): %w", columnName, err)
		}
	}

	// 4. 게시판 업데이트 시간 갱신
	board.UpdatedAt = time.Now()
	_, err = tx.ExecContext(ctx,
		"UPDATE boards SET updated_at = ? WHERE id = ?",
		board.UpdatedAt, board.ID,
	)
	if err != nil {
		return fmt.Errorf("게시판 업데이트 실패: %w", err)
	}

	// 트랜잭션 커밋
	return tx.Commit()
}

// SQLite용 AlterBoardTable 메소드 추가
func (s *dynamicBoardService) alterBoardTableSQLite(ctx context.Context, board *models.Board, addFields, modifyFields []*models.BoardField, dropFields []string) error {
	// 트랜잭션 시작
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. 필드 추가 (SQLite는 ADD COLUMN 지원)
	for _, field := range addFields {
		// 컬럼 이름 유효성 검사
		if !columnNameRegex.MatchString(field.ColumnName) {
			return fmt.Errorf("유효하지 않은 컬럼 이름입니다: %s (영문자, 숫자, 언더스코어만 허용)", field.ColumnName)
		}

		// 예약된 컬럼 이름 검사
		if isReservedColumnName(field.ColumnName) {
			return fmt.Errorf("'%s'는 시스템에서 예약된 필드 이름입니다. 다른 이름을 사용해주세요", field.ColumnName)
		}

		columnDef := s.getColumnDefinition(field)
		query := fmt.Sprintf("ALTER TABLE \"%s\" ADD COLUMN %s;", board.TableName, columnDef)

		_, err := tx.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("필드 추가 실패 (%s): %w", field.Name, err)
		}
	}

	// 2. 필드 수정 및 삭제가 있는 경우 - SQLite는 직접 수정할 수 없으므로 테이블 재생성
	if len(modifyFields) > 0 || len(dropFields) > 0 {
		// 기존 테이블 스키마 조회
		var existingColumns []struct {
			Name         string
			Type         string
			NotNull      int
			DefaultValue sql.NullString
			PrimaryKey   int
		}

		// SQLite pragmas을 사용하여 테이블 스키마 가져오기
		rows, err := tx.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(\"%s\");", board.TableName))
		if err != nil {
			return fmt.Errorf("테이블 스키마 조회 실패: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var col struct {
				Name         string
				Type         string
				NotNull      int
				DefaultValue sql.NullString
				PrimaryKey   int
			}
			var cid int
			if err := rows.Scan(&cid, &col.Name, &col.Type, &col.NotNull, &col.DefaultValue, &col.PrimaryKey); err != nil {
				return fmt.Errorf("컬럼 정보 스캔 실패: %w", err)
			}
			existingColumns = append(existingColumns, col)
		}

		if err := rows.Err(); err != nil {
			return err
		}

		// 수정된 컬럼 정보 맵 생성
		modifiedColumns := make(map[string]*models.BoardField)
		for _, field := range modifyFields {
			modifiedColumns[field.ColumnName] = field
		}

		// 삭제할 컬럼 맵 생성
		dropColumnsMap := make(map[string]bool)
		for _, col := range dropFields {
			dropColumnsMap[col] = true
		}

		// 임시 테이블 이름
		tempTableName := board.TableName + "_temp"

		// CREATE TABLE 문 생성
		var createTableSQL strings.Builder
		createTableSQL.WriteString(fmt.Sprintf("CREATE TABLE \"%s\" (\n", tempTableName))

		// 컬럼 정의 추가
		var columnDefs []string
		var columnNames []string
		for _, col := range existingColumns {
			// 삭제할 컬럼은 제외
			if dropColumnsMap[col.Name] {
				continue
			}

			columnNames = append(columnNames, col.Name)

			// 수정할 컬럼인 경우 변경된 정의 사용
			if field, ok := modifiedColumns[col.Name]; ok {
				columnType := s.getColumnType(field.FieldType)
				nullDef := ""
				if field.Required {
					nullDef = " NOT NULL"
				}

				columnDefs = append(columnDefs, fmt.Sprintf("\"%s\" %s%s", col.Name, columnType, nullDef))
			} else {
				// 기존 컬럼 정의 유지
				nullDef := ""
				if col.NotNull == 1 {
					nullDef = " NOT NULL"
				}

				defaultDef := ""
				if col.DefaultValue.Valid {
					defaultDef = fmt.Sprintf(" DEFAULT %s", col.DefaultValue.String)
				}

				pkDef := ""
				if col.PrimaryKey == 1 {
					pkDef = " PRIMARY KEY"
				}

				columnDefs = append(columnDefs, fmt.Sprintf("\"%s\" %s%s%s%s", col.Name, col.Type, nullDef, defaultDef, pkDef))
			}
		}

		createTableSQL.WriteString(strings.Join(columnDefs, ",\n"))
		createTableSQL.WriteString("\n);")

		// 임시 테이블 생성
		_, err = tx.ExecContext(ctx, createTableSQL.String())
		if err != nil {
			return fmt.Errorf("임시 테이블 생성 실패: %w", err)
		}

		// 데이터 복사
		columnNamesStr := "\"" + strings.Join(columnNames, "\", \"") + "\""
		_, err = tx.ExecContext(ctx, fmt.Sprintf("INSERT INTO \"%s\" (%s) SELECT %s FROM \"%s\";",
			tempTableName, columnNamesStr, columnNamesStr, board.TableName))
		if err != nil {
			return fmt.Errorf("데이터 복사 실패: %w", err)
		}

		// 기존 테이블 삭제
		_, err = tx.ExecContext(ctx, fmt.Sprintf("DROP TABLE \"%s\";", board.TableName))
		if err != nil {
			return fmt.Errorf("기존 테이블 삭제 실패: %w", err)
		}

		// 임시 테이블 이름 변경
		_, err = tx.ExecContext(ctx, fmt.Sprintf("ALTER TABLE \"%s\" RENAME TO \"%s\";",
			tempTableName, board.TableName))
		if err != nil {
			return fmt.Errorf("테이블 이름 변경 실패: %w", err)
		}
	}

	// 게시판 업데이트 시간 갱신
	board.UpdatedAt = time.Now()
	_, err = tx.ExecContext(ctx,
		"UPDATE boards SET updated_at = ? WHERE id = ?",
		board.UpdatedAt, board.ID,
	)
	if err != nil {
		return fmt.Errorf("게시판 업데이트 실패: %w", err)
	}

	// 트랜잭션 커밋
	return tx.Commit()
}

func (s *dynamicBoardService) DropBoardTable(ctx context.Context, tableName string) error {
	if tableName == "" {
		return errors.New("테이블 이름이 비어 있습니다")
	}

	// SQL 인젝션 방지를 위한 테이블 이름 검증
	if !tableNameRegex.MatchString(tableName) {
		return fmt.Errorf("유효하지 않은 테이블 이름입니다: %s (영문자, 숫자, 언더스코어만 허용)", tableName)
	}

	var query string
	if utils.IsPostgres(s.db) {
		query = fmt.Sprintf("DROP TABLE IF EXISTS \"%s\";", tableName)
	} else {
		query = fmt.Sprintf("DROP TABLE IF EXISTS `%s`;", tableName)
	}
	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("테이블 삭제 실패 (%s): %w", tableName, err)
	}
	return nil
}

func (s *dynamicBoardService) GetBoardTableSchema(ctx context.Context, tableName string) ([]*models.BoardField, error) {
	if tableName == "" {
		return nil, errors.New("테이블 이름이 비어 있습니다")
	}

	// SQL 인젝션 방지를 위한 테이블 이름 검증
	if !tableNameRegex.MatchString(tableName) {
		return nil, fmt.Errorf("유효하지 않은 테이블 이름입니다: %s (영문자, 숫자, 언더스코어만 허용)", tableName)
	}

	var fields []*models.BoardField

	// 공통 시스템 필드 목록 (제외할 필드)
	systemFields := []string{"id", "title", "content", "user_id", "view_count", "created_at", "updated_at"}
	systemFieldsStr := "'" + strings.Join(systemFields, "','") + "'"

	// SQL 쿼리 생성 - parameterized query로 변경
	var query string
	if utils.IsPostgres(s.db) {
		query = `
			SELECT 
				column_name, 
				data_type, 
				is_nullable,
				column_default
			FROM 
				information_schema.columns 
			WHERE 
				table_name = $1
				AND column_name NOT IN (` + systemFieldsStr + `)
			ORDER BY 
				ordinal_position;
		`
	} else {
		query = `
			SELECT 
				column_name, 
				data_type, 
				is_nullable,
				column_default
			FROM 
				information_schema.columns 
			WHERE 
				table_name = ?
				AND column_name NOT IN (` + systemFieldsStr + `)
			ORDER BY 
				ordinal_position;
		`
	}

	// 쿼리 실행
	rows, err := s.db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("스키마 조회 실패: %w", err)
	}
	defer rows.Close()

	// 결과 처리
	for rows.Next() {
		var columnName, dataType, isNullable, columnDefault sql.NullString
		err := rows.Scan(&columnName, &dataType, &isNullable, &columnDefault)
		if err != nil {
			return nil, fmt.Errorf("스키마 데이터 처리 실패: %w", err)
		}

		// 필드 유형 변환
		field := &models.BoardField{
			ColumnName:  columnName.String,
			Name:        columnName.String, // 실제 이름은 DB에 저장되어 있지 않음
			DisplayName: columnName.String, // 실제 표시 이름은 DB에 저장되어 있지 않음
			FieldType:   s.mapDatabaseTypeToFieldType(dataType.String),
			Required:    isNullable.String == "NO",
		}

		fields = append(fields, field)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("결과 처리 중 오류: %w", err)
	}

	return fields, nil
}

// 데이터베이스 타입을 필드 타입으로 매핑
func (s *dynamicBoardService) mapDatabaseTypeToFieldType(dbType string) models.FieldType {
	dbType = strings.ToLower(dbType)

	switch dbType {
	case "character varying", "varchar", "nvarchar":
		return models.FieldTypeText
	case "text", "longtext":
		return models.FieldTypeTextarea
	case "integer", "int", "bigint", "number":
		return models.FieldTypeNumber
	case "date", "datetime", "timestamp":
		return models.FieldTypeDate
	case "boolean", "bool", "bit":
		return models.FieldTypeCheckbox
	default:
		return models.FieldTypeText
	}
}
