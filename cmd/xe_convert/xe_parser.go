package main

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"
)

// XEParser XE 데이터 파싱 클래스
type XEParser struct {
	db     *sql.DB
	config *Config
}

// NewXEParser XEParser 생성자
func NewXEParser(db *sql.DB, config *Config) *XEParser {
	return &XEParser{
		db:     db,
		config: config,
	}
}

// XEBoard 제로보드XE 게시판 정보
type XEBoard struct {
	ModuleSrl    int64
	Mid          string
	BrowserTitle string
	Description  string
	IsPublic     bool  // true: 모든 사용자 접근 가능, false: 제한된 사용자만 접근 가능
	FileCount    int64 // 첨부파일 최대 갯수
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// XEDocument 제로보드XE 게시물 정보
type XEDocument struct {
	DocumentSrl   int64
	ModuleSrl     int64
	CategorySrl   int64
	Title         string
	Content       string
	UserID        string
	UserName      string
	NickName      string
	Email         string
	Homepage      string
	Tags          string
	RegDate       time.Time
	UpdatedAt     time.Time
	Voted         int
	Blamed        int
	CommentCount  int
	ViewCount     int
	ExtraVars     map[string]string
	IsSecret      bool
	StatusCode    string
	CommentStatus string
}

// XEComment 제로보드XE 댓글 정보
type XEComment struct {
	CommentSrl  int64
	DocumentSrl int64
	ModuleSrl   int64
	ParentSrl   int64
	Content     string
	UserID      string
	UserName    string
	NickName    string
	Email       string
	Homepage    string
	RegDate     time.Time
	Voted       int
	Blamed      int
	IsSecret    bool
}

// XEFile 제로보드XE 첨부파일 정보
type XEFile struct {
	FileSrl       int64
	ModuleSrl     int64
	DocumentSrl   int64
	UploadPath    string // 계산된 경로 (DB에 없음)
	FileName      string // uploaded_filename
	OriginName    string // source_filename
	FileSize      int64
	DownloadCount int
	IsImage       bool
	Width         int
	Height        int
	RegDate       time.Time
}

// GetBoards 게시판 목록 조회
func (p *XEParser) GetBoards() ([]XEBoard, error) {
	var boards []XEBoard

	// modules 테이블에서 게시판 모듈만 조회
	query := fmt.Sprintf(`
		SELECT module_srl, mid, browser_title, description, 
			   regdate, regdate
		FROM %smodules
		WHERE module = 'board'
	`, p.config.XEPrefix)

	if p.config.Verbose {
		log.Printf("실행 쿼리: %s", query)
	}

	rows, err := p.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("게시판 목록 조회 실패: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var board XEBoard
		var regDateStr, updateDateStr string
		var description sql.NullString

		err := rows.Scan(
			&board.ModuleSrl,
			&board.Mid,
			&board.BrowserTitle,
			&description,
			&regDateStr,
			&updateDateStr,
		)
		if err != nil {
			return nil, fmt.Errorf("게시판 정보 스캔 실패: %w", err)
		}

		// NULL 처리
		if description.Valid {
			board.Description = description.String
		} else {
			board.Description = ""
		}

		// 날짜 파싱
		if regDateStr != "" {
			// 제로보드XE는 YYYYMMDDHHMMSS 형식으로 날짜 저장
			t, err := time.Parse("20060102150405", regDateStr)
			if err == nil {
				board.CreatedAt = t
			} else {
				// 다른 형식 시도
				t, err = time.Parse("2006-01-02 15:04:05", regDateStr)
				if err != nil {
					log.Printf("날짜 파싱 실패: %v", err)
					board.CreatedAt = time.Now()
				} else {
					board.CreatedAt = t
				}
			}
		} else {
			board.CreatedAt = time.Now()
		}
		board.UpdatedAt = board.CreatedAt // 업데이트일이 없으면 생성일과 동일하게

		// 접근 권한 확인 (모든 사용자 = 일반 게시판, 그 외 = 소모임 게시판)
		isPublic, err := p.isPublicBoard(board.ModuleSrl)
		if err != nil {
			log.Printf("게시판 접근 권한 확인 실패: %v - 기본값 true 사용", err)
			isPublic = true // 오류 발생 시 기본값
		}
		board.IsPublic = isPublic

		// 첨부파일 설정 확인
		fileCount, err := p.getBoardFileCount(board.ModuleSrl)
		if err != nil {
			// 오류가 발생해도 기본값 사용
			fileCount = 10
		}
		board.FileCount = fileCount

		boards = append(boards, board)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("게시판 목록 조회 중 오류: %w", err)
	}

	return boards, nil
}

// isPublicBoard 게시판이 공개 게시판인지 확인
func (p *XEParser) isPublicBoard(moduleSrl int64) (bool, error) {
	query := fmt.Sprintf(`
		SELECT count(*) 
		FROM %smodule_grants 
		WHERE module_srl = ? AND name = 'access' AND group_srl = -1
	`, p.config.XEPrefix)

	var count int
	err := p.db.QueryRow(query, moduleSrl).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// getBoardFileCount 게시판의 첨부파일 갯수 설정 조회
func (p *XEParser) getBoardFileCount(moduleSrl int64) (int64, error) {
	query := fmt.Sprintf(`
		SELECT value 
		FROM %smodule_extra_vars 
		WHERE module_srl = ? AND name = 'file_count'
	`, p.config.XEPrefix)

	var countStr string
	err := p.db.QueryRow(query, moduleSrl).Scan(&countStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return 10, nil // 기본값 10
		}
		return 0, err
	}

	var count int64
	fmt.Sscanf(countStr, "%d", &count)
	if count <= 0 {
		count = 10 // 기본값 10
	}

	return count, nil
}

// GetDocuments 게시물 목록 조회
func (p *XEParser) GetDocuments(moduleSrl int64, offset, limit int) ([]XEDocument, error) {
	var documents []XEDocument

	// email 필드 제거 및 homepage 필드도 확인 필요
	query := fmt.Sprintf(`
        SELECT document_srl, module_srl, category_srl, title, content, 
               member_srl, user_id, user_name, nick_name, 
               homepage, tags, regdate, regdate as last_update, 
               voted_count, blamed_count, comment_count, 
               readed_count, status
        FROM %sdocuments
        WHERE module_srl = ?
        ORDER BY document_srl
        LIMIT ? OFFSET ?
    `, p.config.XEPrefix)

	if p.config.Verbose {
		log.Printf("게시물 조회 쿼리: %s", query)
	}

	rows, err := p.db.Query(query, moduleSrl, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("게시물 목록 조회 실패: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var doc XEDocument
		var memberSrl sql.NullInt64
		var categorySrl sql.NullInt64
		var userID, userName, nickName, homepage, tags sql.NullString
		var regDateStr, updateDateStr string
		var status, isSecretStr sql.NullString
		var title, content sql.NullString

		err := rows.Scan(
			&doc.DocumentSrl,
			&doc.ModuleSrl,
			&categorySrl,
			&title,
			&content,
			&memberSrl,
			&userID,
			&userName,
			&nickName,
			&homepage,
			&tags,
			&regDateStr,
			&updateDateStr,
			&doc.Voted,
			&doc.Blamed,
			&doc.CommentCount,
			&doc.ViewCount,
			&status,
		)
		if err != nil {
			return nil, fmt.Errorf("게시물 정보 스캔 실패: %w", err)
		}

		// NULL 처리
		if title.Valid {
			doc.Title = title.String
		} else {
			doc.Title = "(제목 없음)"
		}

		if content.Valid {
			doc.Content = content.String
		} else {
			doc.Content = ""
		}

		if categorySrl.Valid {
			doc.CategorySrl = categorySrl.Int64
		}

		if userID.Valid {
			doc.UserID = userID.String
		}
		if userName.Valid {
			doc.UserName = userName.String
		}
		if nickName.Valid {
			doc.NickName = nickName.String
		} else {
			doc.NickName = "회원" // 기본값
		}
		// email 필드 제거
		doc.Email = "" // 빈 문자열로 설정

		if homepage.Valid {
			doc.Homepage = homepage.String
		}
		if tags.Valid {
			doc.Tags = tags.String
		}

		if status.Valid {
			doc.StatusCode = status.String
		} else {
			doc.StatusCode = "PUBLIC" // 기본값
		}

		// 비밀글 여부 확인
		doc.IsSecret = false
		if isSecretStr.Valid && isSecretStr.String == "Y" {
			doc.IsSecret = true
		}

		// 날짜 파싱
		if regDateStr != "" {
			// 제로보드XE는 YYYYMMDDHHMMSS 형식으로 날짜 저장
			t, err := time.Parse("20060102150405", regDateStr)
			if err == nil {
				doc.RegDate = t
			} else {
				// 다른 형식 시도
				t, err = time.Parse("2006-01-02 15:04:05", regDateStr)
				if err != nil {
					log.Printf("게시물 작성일 파싱 실패: %v", err)
					doc.RegDate = time.Now()
				} else {
					doc.RegDate = t
				}
			}
		} else {
			doc.RegDate = time.Now()
		}

		if updateDateStr != "" {
			// 제로보드XE는 YYYYMMDDHHMMSS 형식으로 날짜 저장
			t, err := time.Parse("20060102150405", updateDateStr)
			if err == nil {
				doc.UpdatedAt = t
			} else {
				// 다른 형식 시도
				t, err = time.Parse("2006-01-02 15:04:05", updateDateStr)
				if err != nil {
					log.Printf("게시물 수정일 파싱 실패: %v", err)
					doc.UpdatedAt = doc.RegDate
				} else {
					doc.UpdatedAt = t
				}
			}
		} else {
			doc.UpdatedAt = doc.RegDate
		}

		// 확장 변수 조회
		extraVars, err := p.getDocumentExtraVars(doc.DocumentSrl)
		if err == nil {
			doc.ExtraVars = extraVars
		} else {
			doc.ExtraVars = make(map[string]string)
		}

		// 댓글 허용 여부 확인
		doc.CommentStatus = "ALLOW" // 기본값

		documents = append(documents, doc)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("게시물 목록 조회 중 오류: %w", err)
	}

	return documents, nil
}

// GetTotalDocumentCount 게시물 총 개수 조회
func (p *XEParser) GetTotalDocumentCount(moduleSrl int64) (int, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM %sdocuments 
		WHERE module_srl = ?
	`, p.config.XEPrefix)

	var count int
	err := p.db.QueryRow(query, moduleSrl).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("게시물 개수 조회 실패: %w", err)
	}

	return count, nil
}

// getDocumentExtraVars 게시물 확장 변수 조회
func (p *XEParser) getDocumentExtraVars(documentSrl int64) (map[string]string, error) {
	query := fmt.Sprintf(`
		SELECT name, value 
		FROM %sdocument_extra_vars 
		WHERE document_srl = ?
	`, p.config.XEPrefix)

	rows, err := p.db.Query(query, documentSrl)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	extraVars := make(map[string]string)
	for rows.Next() {
		var name, value string
		if err := rows.Scan(&name, &value); err != nil {
			return nil, err
		}
		extraVars[name] = value
	}

	return extraVars, rows.Err()
}

// GetComments 댓글 목록 조회
func (p *XEParser) GetComments(documentSrl int64) ([]XEComment, error) {
	var comments []XEComment

	// email 필드 제거
	query := fmt.Sprintf(`
        SELECT comment_srl, document_srl, module_srl, parent_srl, content, 
               member_srl, user_id, user_name, nick_name, 
               homepage, regdate, voted_count, blamed_count
        FROM %scomments
        WHERE document_srl = ?
        ORDER BY comment_srl
    `, p.config.XEPrefix)

	rows, err := p.db.Query(query, documentSrl)
	if err != nil {
		return nil, fmt.Errorf("댓글 목록 조회 실패: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var comment XEComment
		var memberSrl sql.NullInt64
		var parentSrl sql.NullInt64
		var userID, userName, nickName, homepage sql.NullString
		var regDateStr string
		var isSecretStr sql.NullString
		var content sql.NullString

		err := rows.Scan(
			&comment.CommentSrl,
			&comment.DocumentSrl,
			&comment.ModuleSrl,
			&parentSrl,
			&content,
			&memberSrl,
			&userID,
			&userName,
			&nickName,
			&homepage,
			&regDateStr,
			&comment.Voted,
			&comment.Blamed,
		)
		if err != nil {
			return nil, fmt.Errorf("댓글 정보 스캔 실패: %w", err)
		}

		// NULL 처리
		if content.Valid {
			comment.Content = content.String
		} else {
			comment.Content = ""
		}

		if parentSrl.Valid {
			comment.ParentSrl = parentSrl.Int64
		}

		if userID.Valid {
			comment.UserID = userID.String
		}
		if userName.Valid {
			comment.UserName = userName.String
		}
		if nickName.Valid {
			comment.NickName = nickName.String
		} else {
			comment.NickName = "회원" // 기본값
		}
		// email 필드 제거
		comment.Email = "" // 빈 문자열로 설정

		if homepage.Valid {
			comment.Homepage = homepage.String
		}

		// 비밀 댓글 여부
		comment.IsSecret = false
		if isSecretStr.Valid && isSecretStr.String == "Y" {
			comment.IsSecret = true
		}

		// 날짜 파싱
		if regDateStr != "" {
			// 제로보드XE는 YYYYMMDDHHMMSS 형식으로 날짜 저장
			t, err := time.Parse("20060102150405", regDateStr)
			if err == nil {
				comment.RegDate = t
			} else {
				// 다른 형식 시도
				t, err = time.Parse("2006-01-02 15:04:05", regDateStr)
				if err != nil {
					log.Printf("댓글 작성일 파싱 실패: %v", err)
					comment.RegDate = time.Now()
				} else {
					comment.RegDate = t
				}
			}
		} else {
			comment.RegDate = time.Now()
		}

		comments = append(comments, comment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("댓글 목록 조회 중 오류: %w", err)
	}

	return comments, nil
}

// GetFiles 첨부파일 목록 조회
func (p *XEParser) GetFiles(documentSrl int64) ([]XEFile, error) {
	var files []XEFile

	query := fmt.Sprintf(`
        SELECT file_srl, module_srl, upload_target_srl, uploaded_filename, 
               source_filename, file_size, download_count, direct_download, 
               cover_image, regdate, isvalid
        FROM %sfiles
        WHERE upload_target_srl = ?
        ORDER BY file_srl
    `, p.config.XEPrefix)

	rows, err := p.db.Query(query, documentSrl)
	if err != nil {
		return nil, fmt.Errorf("첨부파일 목록 조회 실패: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var file XEFile
		var directDownload, coverImage, isValid string
		var regDateStr string

		err := rows.Scan(
			&file.FileSrl,
			&file.ModuleSrl,
			&file.DocumentSrl,
			&file.FileName,   // uploaded_filename을 FileName으로 사용
			&file.OriginName, // source_filename을 OriginName으로 사용
			&file.FileSize,
			&file.DownloadCount,
			&directDownload,
			&coverImage,
			&regDateStr,
			&isValid,
		)
		if err != nil {
			return nil, fmt.Errorf("첨부파일 정보 스캔 실패: %w", err)
		}

		// XE에서는 upload_path가 없고 파일명만 저장됨 - 경로는 별도로 계산
		// 기본 경로를 구성 (대체로 module_srl 기반)
		mod_srl_str := fmt.Sprintf("%d", file.ModuleSrl)
		if len(mod_srl_str) >= 3 {
			// XE 파일 저장 방식: module_srl을 3개씩 끊어서 역순으로 경로 구성
			// 예: module_srl=123456 -> 경로: 654/321/
			mod_srl_parts := make([]string, 0)
			for i := len(mod_srl_str) - 1; i >= 0; i -= 3 {
				start := i - 2
				if start < 0 {
					start = 0
				}
				mod_srl_parts = append(mod_srl_parts, mod_srl_str[start:i+1])
			}
			file.UploadPath = filepath.Join("attach", getFileType(file.FileName), strings.Join(mod_srl_parts, "/"))
		} else {
			file.UploadPath = filepath.Join("attach", getFileType(file.FileName))
		}

		// 날짜 파싱
		file.RegDate, _ = time.Parse("20060102150405", regDateStr)

		// 이미지 여부 확인
		file.IsImage = isImageFile(file.FileName)

		// 이미지 파일 크기 정보 (DB에 저장되어 있지 않으므로 0으로 설정)
		file.Width = 0
		file.Height = 0

		files = append(files, file)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("첨부파일 목록 조회 중 오류: %w", err)
	}

	return files, nil
}

// getFileType은 파일 타입(images/binary)을 결정합니다
func getFileType(filename string) string {
	if isImageFile(filename) {
		return "images"
	}
	return "binary"
}

// isImageFile 파일이 이미지인지 확인
func isImageFile(filename string) bool {
	ext := getLowerExtension(filename)
	imageExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".webp": true,
		".svg":  true,
	}
	return imageExts[ext]
}

// getLowerExtension 파일 확장자 추출 (소문자)
func getLowerExtension(filename string) string {
	for i := len(filename) - 1; i >= 0 && i > len(filename)-6; i-- {
		if filename[i] == '.' {
			return strings.ToLower(filename[i:])
		}
	}
	return ""
}
