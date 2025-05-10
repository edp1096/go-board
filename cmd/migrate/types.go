package main

import (
	"database/sql"

	"github.com/edp1096/toy-board/config"
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

// FieldInfo는 모델 필드의 메타데이터를 저장합니다
type FieldInfo struct {
	fieldName string
	fieldType string
	tags      map[string]string
}
