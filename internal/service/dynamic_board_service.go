// internal/service/dynamic_board_service.go
package service

import (
	"context"
	"database/sql"
	"dynamic-board/internal/models"
	"errors"
	"fmt"
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

func (s *dynamicBoardService) CreateBoardTable(ctx context.Context, board *models.Board, fields []*models.BoardField) error {
	// 테이블 이름 유효성 검사
	if board.TableName == "" {
		return errors.New("테이블 이름이 비어 있습니다")
	}

	// SQL 생성 - 기본 필드
	columns := []string{
		"id SERIAL PRIMARY KEY",
		"title VARCHAR(200) NOT NULL",
		"content TEXT NOT NULL",
		"user_id INTEGER NOT NULL REFERENCES users(id)",
		"view_count INTEGER NOT NULL DEFAULT 0",
	}

	// 동적 필드 SQL 생성
	for _, field := range fields {
		var columnDef string

		// 필드 유형에 따른 컬럼 정의
		switch field.FieldType {
		case models.FieldTypeText:
			columnDef = fmt.Sprintf("%s VARCHAR(255)", field.ColumnName)
		case models.FieldTypeTextarea:
			columnDef = fmt.Sprintf("%s TEXT", field.ColumnName)
		case models.FieldTypeNumber:
			columnDef = fmt.Sprintf("%s INTEGER", field.ColumnName)
		case models.FieldTypeDate:
			columnDef = fmt.Sprintf("%s DATE", field.ColumnName)
		case models.FieldTypeSelect:
			columnDef = fmt.Sprintf("%s VARCHAR(100)", field.ColumnName)
		case models.FieldTypeCheckbox:
			columnDef = fmt.Sprintf("%s BOOLEAN DEFAULT FALSE", field.ColumnName)
		case models.FieldTypeFile:
			columnDef = fmt.Sprintf("%s VARCHAR(255)", field.ColumnName)
		default:
			columnDef = fmt.Sprintf("%s VARCHAR(255)", field.ColumnName)
		}

		// NOT NULL 옵션 추가
		if field.Required {
			columnDef += " NOT NULL"
		}

		columns = append(columns, columnDef)
	}

	// 타임스탬프 필드 추가
	columns = append(columns, "created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP")
	columns = append(columns, "updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP")

	// CREATE TABLE 쿼리 생성
	query := fmt.Sprintf(
		"CREATE TABLE %s (%s);",
		board.TableName,
		strings.Join(columns, ", "),
	)

	// 쿼리 실행
	_, err := s.db.Exec(query)
	return err
}

func (s *dynamicBoardService) AlterBoardTable(ctx context.Context, board *models.Board, addFields, modifyFields []*models.BoardField, dropFields []string) error {
	// 트랜잭션 시작
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. 필드 추가
	for _, field := range addFields {
		var columnDef string

		// 필드 유형에 따른 컬럼 정의
		switch field.FieldType {
		case models.FieldTypeText:
			columnDef = fmt.Sprintf("%s VARCHAR(255)", field.ColumnName)
		case models.FieldTypeTextarea:
			columnDef = fmt.Sprintf("%s TEXT", field.ColumnName)
		case models.FieldTypeNumber:
			columnDef = fmt.Sprintf("%s INTEGER", field.ColumnName)
		case models.FieldTypeDate:
			columnDef = fmt.Sprintf("%s DATE", field.ColumnName)
		case models.FieldTypeSelect:
			columnDef = fmt.Sprintf("%s VARCHAR(100)", field.ColumnName)
		case models.FieldTypeCheckbox:
			columnDef = fmt.Sprintf("%s BOOLEAN DEFAULT FALSE", field.ColumnName)
		case models.FieldTypeFile:
			columnDef = fmt.Sprintf("%s VARCHAR(255)", field.ColumnName)
		default:
			columnDef = fmt.Sprintf("%s VARCHAR(255)", field.ColumnName)
		}

		// NOT NULL 옵션 추가
		if field.Required {
			columnDef += " NOT NULL"
		}

		// ADD COLUMN 쿼리 실행
		query := fmt.Sprintf(
			"ALTER TABLE %s ADD COLUMN %s;",
			board.TableName,
			columnDef,
		)

		_, err := tx.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("필드 추가 실패 (%s): %w", field.Name, err)
		}
	}

	// 2. 필드 수정 (PostgreSQL과 MySQL/MariaDB 방식이 다름)
	for _, field := range modifyFields {
		var columnDef string

		// 필드 유형에 따른 컬럼 정의
		switch field.FieldType {
		case models.FieldTypeText:
			columnDef = "VARCHAR(255)"
		case models.FieldTypeTextarea:
			columnDef = "TEXT"
		case models.FieldTypeNumber:
			columnDef = "INTEGER"
		case models.FieldTypeDate:
			columnDef = "DATE"
		case models.FieldTypeSelect:
			columnDef = "VARCHAR(100)"
		case models.FieldTypeCheckbox:
			columnDef = "BOOLEAN"
		case models.FieldTypeFile:
			columnDef = "VARCHAR(255)"
		default:
			columnDef = "VARCHAR(255)"
		}

		// 데이터베이스 드라이버에 따라 다른 ALTER COLUMN 구문 사용
		var query string
		dialectName := s.db.Dialect().Name()

		// 수정된 부분: 타입 안전한 방식으로 dialect 비교
		if dialectName.String() == "pg" || dialectName.String() == "postgres" {
			// PostgreSQL용 ALTER COLUMN
			query = fmt.Sprintf(
				"ALTER TABLE %s ALTER COLUMN %s TYPE %s;",
				board.TableName,
				field.ColumnName,
				columnDef,
			)

			// NOT NULL 제약 조건 처리
			if field.Required {
				query += fmt.Sprintf(
					"ALTER TABLE %s ALTER COLUMN %s SET NOT NULL;",
					board.TableName,
					field.ColumnName,
				)
			} else {
				query += fmt.Sprintf(
					"ALTER TABLE %s ALTER COLUMN %s DROP NOT NULL;",
					board.TableName,
					field.ColumnName,
				)
			}
		} else {
			// MySQL/MariaDB용 MODIFY COLUMN
			if field.Required {
				columnDef += " NOT NULL"
			}

			query = fmt.Sprintf(
				"ALTER TABLE %s MODIFY COLUMN %s %s;",
				board.TableName,
				field.ColumnName,
				columnDef,
			)
		}

		_, err := tx.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("필드 수정 실패 (%s): %w", field.Name, err)
		}
	}

	// 3. 필드 삭제
	for _, columnName := range dropFields {
		query := fmt.Sprintf(
			"ALTER TABLE %s DROP COLUMN %s;",
			board.TableName,
			columnName,
		)

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

func (s *dynamicBoardService) DropBoardTable(ctx context.Context, tableName string) error {
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName)
	_, err := s.db.ExecContext(ctx, query)
	return err
}

func (s *dynamicBoardService) GetBoardTableSchema(ctx context.Context, tableName string) ([]*models.BoardField, error) {
	var fields []*models.BoardField

	// 데이터베이스 드라이버에 따라 다른 스키마 조회 방법 사용
	var query string
	var rows *sql.Rows
	var err error

	dialectName := s.db.Dialect().Name()

	// 수정된 부분: 타입 안전한 방식으로 dialect 비교
	if dialectName.String() == "pg" || dialectName.String() == "postgres" {
		// PostgreSQL 스키마 조회
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
                AND column_name NOT IN ('id', 'title', 'content', 'user_id', 'view_count', 'created_at', 'updated_at')
            ORDER BY 
                ordinal_position;
        `
		rows, err = s.db.QueryContext(ctx, query, tableName)
	} else {
		// MySQL/MariaDB 스키마 조회
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
                AND column_name NOT IN ('id', 'title', 'content', 'user_id', 'view_count', 'created_at', 'updated_at')
            ORDER BY 
                ordinal_position;
        `
		rows, err = s.db.QueryContext(ctx, query, tableName)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 결과 처리
	for rows.Next() {
		var columnName, dataType, isNullable, columnDefault sql.NullString
		err := rows.Scan(&columnName, &dataType, &isNullable, &columnDefault)
		if err != nil {
			return nil, err
		}

		// 필드 유형 변환
		var fieldType models.FieldType
		switch dataType.String {
		case "character varying", "varchar":
			fieldType = models.FieldTypeText
		case "text":
			fieldType = models.FieldTypeTextarea
		case "integer", "int":
			fieldType = models.FieldTypeNumber
		case "date":
			fieldType = models.FieldTypeDate
		case "boolean", "bool":
			fieldType = models.FieldTypeCheckbox
		default:
			fieldType = models.FieldTypeText
		}

		field := &models.BoardField{
			ColumnName:  columnName.String,
			Name:        columnName.String, // 실제 이름은 DB에 저장되어 있지 않음
			DisplayName: columnName.String, // 실제 표시 이름은 DB에 저장되어 있지 않음
			FieldType:   fieldType,
			Required:    isNullable.String == "NO",
		}

		fields = append(fields, field)
	}

	return fields, nil
}
