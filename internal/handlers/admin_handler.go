// internal/handlers/admin_handler.go

package handlers

import (
	"dynamic-board/internal/models"
	"dynamic-board/internal/service"
	"dynamic-board/internal/utils" // utils 패키지 임포트 추가
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gosimple/slug"
)

type AdminHandler struct {
	dynamicBoardService service.DynamicBoardService
	boardService        service.BoardService
}

func NewAdminHandler(dynamicBoardService service.DynamicBoardService, boardService service.BoardService) *AdminHandler {
	return &AdminHandler{
		dynamicBoardService: dynamicBoardService,
		boardService:        boardService,
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
		"title":  "게시판 관리",
		"boards": boards,
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
		Name:        name,
		Description: description,
		BoardType:   boardType,
		Slug:        slugStr,
		TableName:   tableName,
		Active:      true,
	}

	// 필드 정보 파싱
	fieldCount, _ := strconv.Atoi(c.FormValue("field_count", "0"))
	fields := make([]*models.BoardField, 0, fieldCount)

	for i := 0; i < fieldCount; i++ {
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

	// JSON 요청인 경우
	if c.Get("Accept") == "application/json" {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "게시판이 생성되었습니다",
			"id":      board.ID,
		})
	}

	// 웹 요청인 경우 게시판 관리 페이지로 리다이렉트
	return c.Redirect("/admin/boards")
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

	return utils.RenderWithUser(c, "admin/board_edit", fiber.Map{
		"title": "게시판 수정",
		"board": board,
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

	for i := 0; i < fieldCount; i++ {
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

			field := &models.BoardField{
				ID:          fieldID,
				BoardID:     boardID,
				Name:        fieldName,
				DisplayName: displayName,
				FieldType:   fieldType,
				Required:    required,
				Sortable:    sortable,
				Searchable:  searchable,
				Options:     options,
				SortOrder:   i + 1,
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

	// JSON 요청인 경우
	if c.Get("Accept") == "application/json" {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "게시판이 수정되었습니다",
		})
	}

	// 웹 요청인 경우 게시판 관리 페이지로 리다이렉트
	return c.Redirect("/admin/boards")
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

	// JSON 요청인 경우
	if c.Get("Accept") == "application/json" {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "게시판이 삭제되었습니다",
		})
	}

	// 웹 요청인 경우 게시판 관리 페이지로 리다이렉트
	return c.Redirect("/admin/boards")
}
