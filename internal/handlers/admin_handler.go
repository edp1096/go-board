// internal/handlers/admin_handler.go

package handlers

import (
	"fmt"
	"go-board/internal/models"
	"go-board/internal/service"
	"go-board/internal/utils" // utils 패키지 임포트 추가
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gosimple/slug"
)

type AdminHandler struct {
	dynamicBoardService service.DynamicBoardService
	boardService        service.BoardService
	authService         service.AuthService
}

func NewAdminHandler(dynamicBoardService service.DynamicBoardService, boardService service.BoardService, authService service.AuthService) *AdminHandler {
	return &AdminHandler{
		dynamicBoardService: dynamicBoardService,
		boardService:        boardService,
		authService:         authService,
	}
}

// Dashboard 관리자 대시보드 페이지
func (h *AdminHandler) Dashboard(c *fiber.Ctx) error {
	return utils.RenderWithUser(c, "admin/dashboard", fiber.Map{
		"title": "관리자 대시보드",
	})
}

// ListBoards 게시판 관리 목록 페이지
func (h *AdminHandler) ListBoards(c *fiber.Ctx) error {
	// 모든 게시판 조회 (비활성 게시판 포함)
	boards, err := h.boardService.ListBoards(c.Context(), false)
	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "게시판 목록을 불러오는데 실패했습니다",
			"error":   err.Error(),
		})
	}

	return utils.RenderWithUser(c, "admin/boards", fiber.Map{
		"title":          "게시판 관리",
		"boards":         boards,
		"pageScriptPath": "/static/js/pages/admin-boards.js",
	})
}

// CreateBoardPage 게시판 생성 페이지
func (h *AdminHandler) CreateBoardPage(c *fiber.Ctx) error {
	return utils.RenderWithUser(c, "admin/board_create", fiber.Map{
		"title": "게시판 생성",
		"boardTypes": []models.BoardType{
			models.BoardTypeNormal,
			models.BoardTypeGallery,
			models.BoardTypeQnA,
		},
		"fieldTypes": []models.FieldType{
			models.FieldTypeText,
			models.FieldTypeTextarea,
			models.FieldTypeNumber,
			models.FieldTypeDate,
			models.FieldTypeSelect,
			models.FieldTypeCheckbox,
			models.FieldTypeFile,
		},
		"pageScriptPath": "/static/js/pages/admin-board-create.js",
	})
}

// CreateBoard 게시판 생성 처리
func (h *AdminHandler) CreateBoard(c *fiber.Ctx) error {
	// 폼 데이터 파싱
	name := c.FormValue("name")
	description := c.FormValue("description")
	boardTypeStr := c.FormValue("board_type")
	boardType := models.BoardType(boardTypeStr)
	slugStr := c.FormValue("slug")
	commentsEnabled := c.FormValue("comments_enabled") == "on"
	allowAnonymous := c.FormValue("allow_anonymous") == "on"

	// 유효성 검사
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "게시판 이름을 입력해주세요",
		})
	}

	// 슬러그 자동 생성
	if slugStr == "" {
		slugStr = slug.Make(name)
	}

	// 테이블 이름 생성
	tableName := "board_" + slug.Make(name)

	// 게시판 객체 생성
	board := &models.Board{
		Name:            name,
		Description:     description,
		BoardType:       boardType,
		Slug:            slugStr,
		TableName:       tableName,
		Active:          true,
		CommentsEnabled: commentsEnabled,
		AllowAnonymous:  allowAnonymous,
	}

	// 필드 정보 파싱
	fieldCount, _ := strconv.Atoi(c.FormValue("field_count", "0"))
	fields := make([]*models.BoardField, 0, fieldCount)

	for i := range fieldCount {
		fieldName := c.FormValue(fmt.Sprintf("field_name_%d", i))
		displayName := c.FormValue(fmt.Sprintf("display_name_%d", i))
		fieldTypeStr := c.FormValue(fmt.Sprintf("field_type_%d", i))
		fieldType := models.FieldType(fieldTypeStr)
		required := c.FormValue(fmt.Sprintf("required_%d", i)) == "on"
		sortable := c.FormValue(fmt.Sprintf("sortable_%d", i)) == "on"
		searchable := c.FormValue(fmt.Sprintf("searchable_%d", i)) == "on"
		options := c.FormValue(fmt.Sprintf("options_%d", i))

		// 필드 유효성 검사
		if fieldName == "" {
			continue
		}

		// 컬럼명 생성
		columnName := slug.Make(fieldName)

		// 필드 객체 생성
		field := &models.BoardField{
			Name:        fieldName,
			DisplayName: displayName,
			ColumnName:  columnName,
			FieldType:   fieldType,
			Required:    required,
			Sortable:    sortable,
			Searchable:  searchable,
			Options:     options,
			SortOrder:   i + 1,
		}

		fields = append(fields, field)
	}

	// 갤러리 타입일 경우 자동으로 파일 업로드 필드 추가
	var galleryImagesField *models.BoardField = nil

	if boardType == models.BoardTypeGallery {
		// 갤러리 게시판을 위한 이미지 필드가 있는지 확인
		for _, field := range fields {
			if field.FieldType == models.FieldTypeFile {
				galleryImagesField = field
				// 기존 파일 필드를 갤러리 이미지 필드로 수정
				field.Name = "gallery_images"
				field.DisplayName = "이미지"
				field.ColumnName = "gallery_images"
				field.Required = true
				break
			}
		}

		// 이미지 필드가 없으면 추가
		if galleryImagesField == nil {
			galleryImagesField = &models.BoardField{
				Name:        "gallery_images",
				DisplayName: "이미지",
				ColumnName:  "gallery_images",
				FieldType:   models.FieldTypeFile,
				// Required:    true,
				Required:   false,
				Sortable:   false,
				Searchable: false,
				Options:    "",
				SortOrder:  len(fields) + 1,
			}
			fields = append(fields, galleryImagesField)
		}
	} else if boardType == models.BoardTypeQnA {
		// Q&A 게시판을 위한 필드 추가
		// 1. 상태 필드
		statusField := &models.BoardField{
			Name:        "status",
			DisplayName: "상태",
			ColumnName:  "status",
			FieldType:   models.FieldTypeSelect,
			Required:    true,
			Sortable:    true,
			Searchable:  true,
			Options:     `[{"value":"unsolved","label":"미해결"},{"value":"solved","label":"해결됨"}]`,
			SortOrder:   len(fields) + 1,
		}
		fields = append(fields, statusField)

		// 2. 태그 필드
		tagsField := &models.BoardField{
			Name:        "tags",
			DisplayName: "태그",
			ColumnName:  "tags",
			FieldType:   models.FieldTypeText,
			Required:    false,
			Sortable:    false,
			Searchable:  true,
			Options:     "",
			SortOrder:   len(fields) + 2,
		}
		fields = append(fields, tagsField)

		// 3. 답변 수 필드 (시스템이 자동 관리)
		answerCountField := &models.BoardField{
			Name:        "answer_count",
			DisplayName: "답변 수",
			ColumnName:  "answer_count",
			FieldType:   models.FieldTypeNumber,
			Required:    false,
			Sortable:    true,
			Searchable:  false,
			Options:     "",
			SortOrder:   len(fields) + 3,
		}
		fields = append(fields, answerCountField)

		// 4. 투표 수 필드 (시스템이 자동 관리)
		voteCountField := &models.BoardField{
			Name:        "vote_count",
			DisplayName: "투표 수",
			ColumnName:  "vote_count",
			FieldType:   models.FieldTypeNumber,
			Required:    false,
			Sortable:    true,
			Searchable:  false,
			Options:     "",
			SortOrder:   len(fields) + 4,
		}
		fields = append(fields, voteCountField)

		// 5. 베스트 답변 ID 필드 (시스템이 자동 관리)
		bestAnswerField := &models.BoardField{
			Name:        "best_answer_id",
			DisplayName: "베스트 답변 ID",
			ColumnName:  "best_answer_id",
			FieldType:   models.FieldTypeNumber,
			Required:    false,
			Sortable:    false,
			Searchable:  false,
			Options:     "",
			SortOrder:   len(fields) + 5,
		}
		fields = append(fields, bestAnswerField)
	}

	// 트랜잭션 처리가 필요하지만 간단히 처리

	// 1. 게시판 생성
	err := h.boardService.CreateBoard(c.Context(), board)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "게시판 생성에 실패했습니다: " + err.Error(),
		})
	}

	// 2. 게시판 테이블 생성
	err = h.dynamicBoardService.CreateBoardTable(c.Context(), board, fields)
	if err != nil {
		// 롤백이 필요하지만 생략
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "게시판 테이블 생성에 실패했습니다: " + err.Error(),
		})
	}

	// 3. 게시판 필드 생성
	for _, field := range fields {
		field.BoardID = board.ID
		err = h.boardService.AddBoardField(c.Context(), board.ID, field)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "게시판 필드 생성에 실패했습니다: " + err.Error(),
			})
		}
	}

	// 매니저 ID 처리
	managerIDs := c.FormValue("manager_ids[]")
	if managerIDs != "" {
		managerIDList := strings.SplitSeq(managerIDs, ",")
		for idStr := range managerIDList {
			managerID, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				continue
			}

			// 매니저 추가
			h.boardService.AddBoardManager(c.Context(), board.ID, managerID)
		}
	}

	// 폼이 fetch를 통해 제출되었으므로 항상 JSON 응답을 반환
	c.Set("Content-Type", "application/json")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "게시판이 생성되었습니다",
		"id":      board.ID,
	})
}

// EditBoardPage 게시판 수정 페이지
func (h *AdminHandler) EditBoardPage(c *fiber.Ctx) error {
	boardID, err := strconv.ParseInt(c.Params("boardID"), 10, 64)
	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "잘못된 게시판 ID입니다",
			"error":   err.Error(),
		})
	}

	// 게시판 정보 조회
	board, err := h.boardService.GetBoardByID(c.Context(), boardID)
	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "게시판을 찾을 수 없습니다",
			"error":   err.Error(),
		})
	}

	// 게시판 정보와 함께 매니저 정보도 조회
	managers, err := h.boardService.GetBoardManagers(c.Context(), boardID)
	if err != nil {
		// 오류 처리하지만 계속 진행
		managers = []*models.User{}
	}

	return utils.RenderWithUser(c, "admin/board_edit", fiber.Map{
		"title":    "게시판 수정",
		"board":    board,
		"managers": managers,
		"boardTypes": []models.BoardType{
			models.BoardTypeNormal,
			models.BoardTypeGallery,
			models.BoardTypeQnA,
		},
		"fieldTypes": []models.FieldType{
			models.FieldTypeText,
			models.FieldTypeTextarea,
			models.FieldTypeNumber,
			models.FieldTypeDate,
			models.FieldTypeSelect,
			models.FieldTypeCheckbox,
			models.FieldTypeFile,
		},
		"pageScriptPath": "/static/js/pages/admin-board-edit.js",
	})
}

// UpdateBoard 게시판 수정 처리
func (h *AdminHandler) UpdateBoard(c *fiber.Ctx) error {
	boardID, err := strconv.ParseInt(c.Params("boardID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 게시판 ID입니다",
		})
	}

	// 게시판 정보 조회
	board, err := h.boardService.GetBoardByID(c.Context(), boardID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "게시판을 찾을 수 없습니다",
		})
	}

	// 폼 데이터 파싱
	name := c.FormValue("name")
	description := c.FormValue("description")
	boardTypeStr := c.FormValue("board_type")
	active := c.FormValue("active") == "on"
	commentsEnabled := c.FormValue("comments_enabled") == "on"
	allowAnonymous := c.FormValue("allow_anonymous") == "on"

	// 유효성 검사
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "게시판 이름을 입력해주세요",
		})
	}

	// 게시판 업데이트
	board.Name = name
	board.Description = description
	board.BoardType = models.BoardType(boardTypeStr)
	board.Active = active
	board.CommentsEnabled = commentsEnabled
	board.AllowAnonymous = allowAnonymous

	// 필드 정보 파싱
	fieldCount, _ := strconv.Atoi(c.FormValue("field_count", "0"))
	addFields := make([]*models.BoardField, 0)
	modifyFields := make([]*models.BoardField, 0)

	// 기존 필드 ID 목록
	existingFieldIDs := make(map[int64]bool)
	for _, field := range board.Fields {
		existingFieldIDs[field.ID] = true
	}

	// 제출된 필드 ID 목록
	submittedFieldIDs := make(map[int64]bool)

	// Q&A 게시판 특수 필드 이름들
	qnaSpecialFields := map[string]bool{
		"status":         true,
		"tags":           true,
		"answer_count":   true,
		"vote_count":     true,
		"best_answer_id": true,
	}

	// 기존 특수 필드 정보 저장용 맵
	existingQnaFields := make(map[string]*models.BoardField)
	if board.BoardType == models.BoardTypeQnA {
		for _, field := range board.Fields {
			if qnaSpecialFields[field.Name] {
				existingQnaFields[field.Name] = field
			}
		}
	}

	for i := range fieldCount {
		fieldIDStr := c.FormValue(fmt.Sprintf("field_id_%d", i))
		fieldID, _ := strconv.ParseInt(fieldIDStr, 10, 64)

		fieldName := c.FormValue(fmt.Sprintf("field_name_%d", i))
		displayName := c.FormValue(fmt.Sprintf("display_name_%d", i))
		fieldTypeStr := c.FormValue(fmt.Sprintf("field_type_%d", i))
		fieldType := models.FieldType(fieldTypeStr)
		required := c.FormValue(fmt.Sprintf("required_%d", i)) == "on"
		sortable := c.FormValue(fmt.Sprintf("sortable_%d", i)) == "on"
		searchable := c.FormValue(fmt.Sprintf("searchable_%d", i)) == "on"
		options := c.FormValue(fmt.Sprintf("options_%d", i))

		// 필드 유효성 검사
		if fieldName == "" {
			continue
		}

		// 새 필드 또는 기존 필드 수정
		if fieldID > 0 {
			// 기존 필드 수정
			submittedFieldIDs[fieldID] = true

			// 기존 필드에서 columnName 찾기
			var columnName string
			var createdAt time.Time

			for _, existingField := range board.Fields {
				if existingField.ID == fieldID {
					columnName = existingField.ColumnName
					createdAt = existingField.CreatedAt
					break
				}
			}

			field := &models.BoardField{
				ID:          fieldID,
				BoardID:     boardID,
				Name:        fieldName,
				ColumnName:  columnName,
				DisplayName: displayName,
				FieldType:   fieldType,
				Required:    required,
				Sortable:    sortable,
				Searchable:  searchable,
				Options:     options,
				SortOrder:   i + 1,
				CreatedAt:   createdAt,  // 기존 생성 시간 유지
				UpdatedAt:   time.Now(), // 현재 시간으로 업데이트
			}

			modifyFields = append(modifyFields, field)
		} else {
			// 새 필드 추가
			columnName := slug.Make(fieldName)

			field := &models.BoardField{
				BoardID:     boardID,
				Name:        fieldName,
				ColumnName:  columnName,
				DisplayName: displayName,
				FieldType:   fieldType,
				Required:    required,
				Sortable:    sortable,
				Searchable:  searchable,
				Options:     options,
				SortOrder:   i + 1,
			}

			addFields = append(addFields, field)
		}
	}

	// 삭제할 필드 ID 목록
	dropFieldIDs := make([]int64, 0)
	dropFieldColumns := make([]string, 0)

	for id := range existingFieldIDs {
		if !submittedFieldIDs[id] {
			// Q&A 게시판인 경우 특수 필드는 삭제 목록에서 제외
			var isQnaSpecialField bool = false

			for _, field := range board.Fields {
				if field.ID == id && board.BoardType == models.BoardTypeQnA {
					if qnaSpecialFields[field.Name] {
						isQnaSpecialField = true
						break
					}
				}
			}

			if !isQnaSpecialField {
				// 삭제할 필드 찾기
				for _, field := range board.Fields {
					if field.ID == id {
						dropFieldIDs = append(dropFieldIDs, id)
						dropFieldColumns = append(dropFieldColumns, field.ColumnName)
						break
					}
				}
			}
		}
	}

	// Q&A 게시판인 경우 특수 필드 재추가 (사용자가 지웠거나 없는 경우)
	if board.BoardType == models.BoardTypeQnA {
		// 제출된 필드에서 특수 필드 이름들 확인
		submittedSpecialFields := make(map[string]bool)
		for _, field := range modifyFields {
			if qnaSpecialFields[field.Name] {
				submittedSpecialFields[field.Name] = true
			}
		}
		for _, field := range addFields {
			if qnaSpecialFields[field.Name] {
				submittedSpecialFields[field.Name] = true
			}
		}

		// 누락된 특수 필드 다시 추가
		nextSortOrder := len(modifyFields) + len(addFields) + 1

		// 1. 상태 필드
		if !submittedSpecialFields["status"] {
			field := existingQnaFields["status"]
			if field == nil {
				// 기존에 없었으면 새로 생성
				field = &models.BoardField{
					BoardID:     boardID,
					Name:        "status",
					ColumnName:  "status",
					DisplayName: "상태",
					FieldType:   models.FieldTypeSelect,
					Required:    true,
					Sortable:    true,
					Searchable:  true,
					Options:     `[{"value":"unsolved","label":"미해결"},{"value":"solved","label":"해결됨"}]`,
					SortOrder:   nextSortOrder,
				}
				nextSortOrder++
				addFields = append(addFields, field)
			} else {
				// 기존 필드 유지
				field.SortOrder = nextSortOrder
				nextSortOrder++
				modifyFields = append(modifyFields, field)
				submittedFieldIDs[field.ID] = true // 삭제 목록에서 제외하기 위해
			}
		}

		// 2. 태그 필드
		if !submittedSpecialFields["tags"] {
			field := existingQnaFields["tags"]
			if field == nil {
				field = &models.BoardField{
					BoardID:     boardID,
					Name:        "tags",
					ColumnName:  "tags",
					DisplayName: "태그",
					FieldType:   models.FieldTypeText,
					Required:    false,
					Sortable:    false,
					Searchable:  true,
					Options:     "",
					SortOrder:   nextSortOrder,
				}
				nextSortOrder++
				addFields = append(addFields, field)
			} else {
				field.SortOrder = nextSortOrder
				nextSortOrder++
				modifyFields = append(modifyFields, field)
				submittedFieldIDs[field.ID] = true
			}
		}

		// 3. 답변 수 필드
		if !submittedSpecialFields["answer_count"] {
			field := existingQnaFields["answer_count"]
			if field == nil {
				field = &models.BoardField{
					BoardID:     boardID,
					Name:        "answer_count",
					ColumnName:  "answer_count",
					DisplayName: "답변 수",
					FieldType:   models.FieldTypeNumber,
					Required:    false,
					Sortable:    true,
					Searchable:  false,
					Options:     "",
					SortOrder:   nextSortOrder,
				}
				nextSortOrder++
				addFields = append(addFields, field)
			} else {
				field.SortOrder = nextSortOrder
				nextSortOrder++
				modifyFields = append(modifyFields, field)
				submittedFieldIDs[field.ID] = true
			}
		}

		// 4. 투표 수 필드
		if !submittedSpecialFields["vote_count"] {
			field := existingQnaFields["vote_count"]
			if field == nil {
				field = &models.BoardField{
					BoardID:     boardID,
					Name:        "vote_count",
					ColumnName:  "vote_count",
					DisplayName: "투표 수",
					FieldType:   models.FieldTypeNumber,
					Required:    false,
					Sortable:    true,
					Searchable:  false,
					Options:     "",
					SortOrder:   nextSortOrder,
				}
				nextSortOrder++
				addFields = append(addFields, field)
			} else {
				field.SortOrder = nextSortOrder
				nextSortOrder++
				modifyFields = append(modifyFields, field)
				submittedFieldIDs[field.ID] = true
			}
		}

		// 5. 베스트 답변 ID 필드
		if !submittedSpecialFields["best_answer_id"] {
			field := existingQnaFields["best_answer_id"]
			if field == nil {
				field = &models.BoardField{
					BoardID:     boardID,
					Name:        "best_answer_id",
					ColumnName:  "best_answer_id",
					DisplayName: "베스트 답변 ID",
					FieldType:   models.FieldTypeNumber,
					Required:    false,
					Sortable:    false,
					Searchable:  false,
					Options:     "",
					SortOrder:   nextSortOrder,
				}
				nextSortOrder++
				addFields = append(addFields, field)
			} else {
				field.SortOrder = nextSortOrder
				nextSortOrder++
				modifyFields = append(modifyFields, field)
				submittedFieldIDs[field.ID] = true
			}
		}

		// dropFieldIDs와 dropFieldColumns 재구성 (특수 필드 제외)
		newDropFieldIDs := make([]int64, 0)
		newDropFieldColumns := make([]string, 0)

		for i, id := range dropFieldIDs {
			var isQnaSpecialField bool = false
			for _, field := range board.Fields {
				if field.ID == id && qnaSpecialFields[field.Name] {
					isQnaSpecialField = true
					break
				}
			}

			if !isQnaSpecialField {
				newDropFieldIDs = append(newDropFieldIDs, id)
				newDropFieldColumns = append(newDropFieldColumns, dropFieldColumns[i])
			}
		}

		dropFieldIDs = newDropFieldIDs
		dropFieldColumns = newDropFieldColumns
	}

	// 트랜잭션 처리가 필요하지만 간단히 처리

	// 1. 게시판 업데이트
	err = h.boardService.UpdateBoard(c.Context(), board)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "게시판 업데이트에 실패했습니다: " + err.Error(),
		})
	}

	// 2. 게시판 테이블 변경
	err = h.dynamicBoardService.AlterBoardTable(c.Context(), board, addFields, modifyFields, dropFieldColumns)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "게시판 테이블 변경에 실패했습니다: " + err.Error(),
		})
	}

	// 3. 게시판 필드 추가
	for _, field := range addFields {
		err = h.boardService.AddBoardField(c.Context(), boardID, field)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "게시판 필드 추가에 실패했습니다: " + err.Error(),
			})
		}
	}

	// 4. 게시판 필드 수정
	for _, field := range modifyFields {
		err = h.boardService.UpdateBoardField(c.Context(), field)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "게시판 필드 수정에 실패했습니다: " + err.Error(),
			})
		}
	}

	// 5. 게시판 필드 삭제
	for _, fieldID := range dropFieldIDs {
		err = h.boardService.DeleteBoardField(c.Context(), fieldID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "게시판 필드 삭제에 실패했습니다: " + err.Error(),
			})
		}
	}

	// 매니저 처리
	h.boardService.RemoveAllBoardManagers(c.Context(), boardID) // 기존 매니저 목록 제거
	managerIDs := c.FormValue("manager_ids")
	if managerIDs != "" {
		managerIDList := strings.SplitSeq(managerIDs, ",") // 새 매니저 추가
		for idStr := range managerIDList {
			managerID, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				continue
			}

			h.boardService.AddBoardManager(c.Context(), boardID, managerID)
		}
	}

	// 폼이 fetch를 통해 제출되었으므로 항상 JSON 응답을 반환
	c.Set("Content-Type", "application/json")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "게시판이 수정되었습니다",
	})
}

// DeleteBoard 게시판 삭제 처리
func (h *AdminHandler) DeleteBoard(c *fiber.Ctx) error {
	boardID, err := strconv.ParseInt(c.Params("boardID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 게시판 ID입니다",
		})
	}

	// 게시판 정보 조회
	board, err := h.boardService.GetBoardByID(c.Context(), boardID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "게시판을 찾을 수 없습니다",
		})
	}

	// 트랜잭션 처리가 필요하지만 간단히 처리

	// 1. 게시판 테이블 삭제
	err = h.dynamicBoardService.DropBoardTable(c.Context(), board.TableName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "게시판 테이블 삭제에 실패했습니다: " + err.Error(),
		})
	}

	// 2. 게시판 삭제
	err = h.boardService.DeleteBoard(c.Context(), boardID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "게시판 삭제에 실패했습니다: " + err.Error(),
		})
	}

	// 폼이 fetch를 통해 제출되었으므로 항상 JSON 응답을 반환
	c.Set("Content-Type", "application/json")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "게시판이 삭제되었습니다",
	})
}

// ListUsers 사용자 관리 목록 페이지
func (h *AdminHandler) ListUsers(c *fiber.Ctx) error {
	// 페이지네이션 파라미터
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize := 10

	// 검색어 파라미터
	search := c.Query("search", "")

	// 오프셋 계산
	offset := (page - 1) * pageSize

	// 사용자 목록 조회
	users, count, err := h.authService.ListUsers(c.Context(), offset, pageSize, search)
	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "사용자 목록을 불러오는데 실패했습니다",
			"error":   err.Error(),
		})
	}

	// 총 페이지 수 계산
	totalPages := (count + pageSize - 1) / pageSize

	return utils.RenderWithUser(c, "admin/users", fiber.Map{
		"title":          "사용자 관리",
		"users":          users,
		"currentPage":    page,
		"totalPages":     totalPages,
		"totalUsers":     count,
		"search":         search,
		"pageScriptPath": "/static/js/pages/admin-users.js",
	})
}

// UpdateUserRole 사용자 역할 변경
func (h *AdminHandler) UpdateUserRole(c *fiber.Ctx) error {
	userID, err := strconv.ParseInt(c.Params("userID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 사용자 ID입니다",
		})
	}

	// 현재 로그인한 사용자가 자신의 역할을 변경하려고 시도하는 경우 방지
	currentUser := c.Locals("user").(*models.User)
	if currentUser.ID == userID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "자신의 역할은 변경할 수 없습니다",
		})
	}

	// 요청 본문 파싱
	var body struct {
		Role models.Role `json:"role"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "요청 데이터 파싱에 실패했습니다",
		})
	}

	// 역할 검증
	if body.Role != models.RoleAdmin && body.Role != models.RoleUser {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "유효하지 않은 역할입니다",
		})
	}

	// 사용자 정보 조회
	user, err := h.authService.GetUserByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "사용자를 찾을 수 없습니다",
		})
	}

	// 역할 업데이트
	user.Role = body.Role
	err = h.authService.UpdateUser(c.Context(), user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "사용자 역할 업데이트에 실패했습니다: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "사용자 역할이 업데이트되었습니다",
	})
}

// UpdateUserStatus 사용자 상태 변경
func (h *AdminHandler) UpdateUserStatus(c *fiber.Ctx) error {
	userID, err := strconv.ParseInt(c.Params("userID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 사용자 ID입니다",
		})
	}

	// 현재 로그인한 사용자가 자신의 상태를 변경하려고 시도하는 경우 방지
	currentUser := c.Locals("user").(*models.User)
	if currentUser.ID == userID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "자신의 계정 상태는 변경할 수 없습니다",
		})
	}

	// 요청 본문 파싱
	var body struct {
		Active bool `json:"active"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "요청 데이터 파싱에 실패했습니다",
		})
	}

	// 사용자 정보 조회
	user, err := h.authService.GetUserByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "사용자를 찾을 수 없습니다",
		})
	}

	// 명시적으로 Active 필드 설정
	user.Active = body.Active
	user.UpdatedAt = time.Now()

	// 특정 필드만 업데이트하도록 수정
	err = h.authService.UpdateUserActiveStatus(c.Context(), user.ID, body.Active)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "사용자 상태 업데이트에 실패했습니다: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "사용자 상태가 업데이트되었습니다",
	})
}

// CreateUserPage 사용자 추가 페이지 렌더링
func (h *AdminHandler) CreateUserPage(c *fiber.Ctx) error {
	return utils.RenderWithUser(c, "admin/user_create", fiber.Map{
		"title":          "새 사용자 추가",
		"pageScriptPath": "/static/js/pages/admin-user-create.js",
	})
}

// CreateUser 사용자 추가 처리
func (h *AdminHandler) CreateUser(c *fiber.Ctx) error {
	// 요청 본문 파싱
	var body struct {
		Username string      `json:"username"`
		Email    string      `json:"email"`
		Password string      `json:"password"`
		FullName string      `json:"full_name"`
		Role     models.Role `json:"role"`
		Active   bool        `json:"active"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "요청 데이터 파싱에 실패했습니다",
		})
	}

	// 필수 필드 검증
	if body.Username == "" || body.Email == "" || body.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "사용자명, 이메일, 비밀번호는 필수 항목입니다",
		})
	}

	// 역할 검증
	if body.Role != models.RoleAdmin && body.Role != models.RoleUser {
		body.Role = models.RoleUser // 기본값으로 설정
	}

	// 사용자 등록
	_, err := h.authService.Register(c.Context(), body.Username, body.Email, body.Password, body.FullName)
	if err != nil {
		errorMessage := "사용자 등록에 실패했습니다"
		if err == service.ErrUsernameTaken {
			errorMessage = "이미 사용 중인 사용자명입니다"
		} else if err == service.ErrEmailTaken {
			errorMessage = "이미 사용 중인 이메일입니다"
		}

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": errorMessage,
		})
	}

	// 생성된 사용자 조회
	user, err := h.authService.GetUserByUsername(c.Context(), body.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "사용자 정보 조회에 실패했습니다",
		})
	}

	// 역할과 활성 상태 업데이트 (RegisterService에서는 기본값이 설정되기 때문)
	user.Role = body.Role
	user.Active = body.Active
	err = h.authService.UpdateUser(c.Context(), user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "사용자 정보 업데이트에 실패했습니다",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "사용자가 성공적으로 추가되었습니다",
		"id":      user.ID,
	})
}

// EditUserPage 사용자 수정 페이지 렌더링
func (h *AdminHandler) EditUserPage(c *fiber.Ctx) error {
	userID, err := strconv.ParseInt(c.Params("userID"), 10, 64)
	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "잘못된 사용자 ID입니다",
			"error":   err.Error(),
		})
	}

	// 사용자 정보 조회
	user, err := h.authService.GetUserByID(c.Context(), userID)
	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "사용자를 찾을 수 없습니다",
			"error":   err.Error(),
		})
	}

	return utils.RenderWithUser(c, "admin/user_edit", fiber.Map{
		"title":          "사용자 정보 수정",
		"user":           user,
		"pageScriptPath": "/static/js/pages/admin-user-edit.js",
	})
}

// UpdateUser 사용자 정보 수정 처리
func (h *AdminHandler) UpdateUser(c *fiber.Ctx) error {
	userID, err := strconv.ParseInt(c.Params("userID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 사용자 ID입니다",
		})
	}

	// 요청 본문 파싱
	var body struct {
		Username string      `json:"username"`
		Email    string      `json:"email"`
		Password string      `json:"password"`
		FullName string      `json:"full_name"`
		Role     models.Role `json:"role"`
		Active   bool        `json:"active"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "요청 데이터 파싱에 실패했습니다",
		})
	}

	// 필수 필드 검증
	if body.Username == "" || body.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "사용자명과 이메일은 필수 항목입니다",
		})
	}

	// 역할 검증
	if body.Role != models.RoleAdmin && body.Role != models.RoleUser {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "유효하지 않은 역할입니다",
		})
	}

	// 현재 로그인한 사용자가 관리자 -> 일반 사용자로 자신의 역할을 변경하려는 경우 방지
	currentUser := c.Locals("user").(*models.User)
	if currentUser.ID == userID && currentUser.Role == models.RoleAdmin && body.Role == models.RoleUser {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "자신의 관리자 권한을 해제할 수 없습니다",
		})
	}

	// 사용자 정보 조회
	user, err := h.authService.GetUserByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "사용자를 찾을 수 없습니다",
		})
	}

	// 다른 사용자와 사용자명/이메일 중복 확인
	if user.Username != body.Username {
		existingUser, _ := h.authService.GetUserByUsername(c.Context(), body.Username)
		if existingUser != nil && existingUser.ID != userID {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "이미 사용 중인 사용자명입니다",
			})
		}
	}

	// 사용자 정보 업데이트
	user.Username = body.Username
	user.Email = body.Email
	user.FullName = body.FullName
	user.Role = body.Role
	user.Active = body.Active
	user.UpdatedAt = time.Now()

	err = h.authService.UpdateUser(c.Context(), user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "사용자 정보 업데이트에 실패했습니다: " + err.Error(),
		})
	}

	// 비밀번호 변경 요청이 있는 경우
	if body.Password != "" {
		// 관리자에 의한 비밀번호 변경이므로 현재 비밀번호 검증 없이 변경
		err = h.authService.AdminChangePassword(c.Context(), userID, body.Password)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "비밀번호 변경에 실패했습니다: " + err.Error(),
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "사용자 정보가 업데이트되었습니다",
	})
}

// DeleteUser 사용자 삭제 처리
func (h *AdminHandler) DeleteUser(c *fiber.Ctx) error {
	userID, err := strconv.ParseInt(c.Params("userID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 사용자 ID입니다",
		})
	}

	// 현재 로그인한 사용자가 자신을 삭제하려는 경우 방지
	currentUser := c.Locals("user").(*models.User)
	if currentUser.ID == userID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "자신의 계정은 삭제할 수 없습니다",
		})
	}

	// 사용자 삭제
	err = h.authService.DeleteUser(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "사용자 삭제에 실패했습니다: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "사용자가 성공적으로 삭제되었습니다",
	})
}

// SearchUsers 사용자 검색
func (h *AdminHandler) SearchUsers(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "검색어를 입력해주세요",
		})
	}

	// 사용자 검색
	users, err := h.authService.SearchUsers(c.Context(), query, 0, 10)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "사용자 검색에 실패했습니다",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"users":   users,
	})
}

// UserApprovalPage 사용자 승인 관리 페이지 렌더링
func (h *AdminHandler) UserApprovalPage(c *fiber.Ctx) error {
	// 승인 대기 중인 사용자 목록 조회
	users, err := h.authService.GetPendingApprovalUsers(c.Context())
	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "사용자 목록을 불러오는데 실패했습니다",
			"error":   err.Error(),
		})
	}

	return utils.RenderWithUser(c, "admin/user_approval", fiber.Map{
		"title":          "사용자 승인 관리",
		"users":          users,
		"pageScriptPath": "/static/js/pages/admin-user-approval.js",
	})
}

// ApproveUser 사용자 승인 처리
func (h *AdminHandler) ApproveUser(c *fiber.Ctx) error {
	userID, err := strconv.ParseInt(c.Params("userID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 사용자 ID입니다",
		})
	}

	// 사용자 정보 조회
	// user, err := h.authService.GetUserByID(c.Context(), userID)
	_, err = h.authService.GetUserByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "사용자를 찾을 수 없습니다",
		})
	}

	// 승인 상태 업데이트
	err = h.authService.UpdateUserApprovalStatus(c.Context(), userID, models.ApprovalApproved)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "사용자 승인에 실패했습니다: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "사용자가 성공적으로 승인되었습니다",
	})
}

// RejectUser 사용자 승인 거부 처리
func (h *AdminHandler) RejectUser(c *fiber.Ctx) error {
	userID, err := strconv.ParseInt(c.Params("userID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 사용자 ID입니다",
		})
	}

	// 사용자 정보 조회
	// user, err := h.authService.GetUserByID(c.Context(), userID)
	_, err = h.authService.GetUserByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "사용자를 찾을 수 없습니다",
		})
	}

	// 승인 상태 업데이트
	err = h.authService.UpdateUserApprovalStatus(c.Context(), userID, models.ApprovalRejected)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "사용자 승인 거부에 실패했습니다: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "사용자 승인이 거부되었습니다",
	})
}
