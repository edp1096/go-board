package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite" // CGO 없는 SQLite 드라이버
)

func main() {
	startTime := time.Now()

	// 1. 환경 설정 로드
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("환경 설정 로드 실패: %v", err)
	}

	// 2. 소스 데이터베이스 연결 (XE)
	sourceDB, err := connectDatabase(config.SourceDriver, config.GetSourceDSN())
	if err != nil {
		log.Fatalf("소스 데이터베이스 연결 실패: %v", err)
	}
	defer sourceDB.Close()

	// 3. 대상 데이터베이스 연결 (Toy-Board)
	targetDB, err := connectDatabase(config.TargetDriver, config.GetTargetDSN())
	if err != nil {
		log.Fatalf("대상 데이터베이스 연결 실패: %v", err)
	}
	defer targetDB.Close()

	// 4. 마이그레이션 실행
	migration := NewMigration(sourceDB, targetDB, config)

	fmt.Println("마이그레이션을 시작합니다...")
	if err := migration.Run(); err != nil {
		log.Fatalf("마이그레이션 실패: %v", err)
	}

	// 소요 시간 계산
	duration := time.Since(startTime)
	fmt.Printf("마이그레이션이 완료되었습니다. (소요 시간: %s)\n", duration)
}

// connectDatabase 데이터베이스 연결 함수
func connectDatabase(driver, dsn string) (*sql.DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("데이터베이스 열기 실패: %w", err)
	}

	// 연결 테스트
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("데이터베이스 연결 실패: %w", err)
	}

	// 연결 최적화 설정
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}
