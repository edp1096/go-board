package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"dynamic-board/config"
	"dynamic-board/internal/handlers"
	"dynamic-board/internal/middleware"
	"dynamic-board/internal/repository"
	"dynamic-board/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
)

func main() {
	// 환경 설정 로드
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("설정을 로드할 수 없습니다: %v", err)
	}

	// 데이터베이스 연결
	db, err := config.ConnectDatabase(cfg)
	if err != nil {
		log.Fatalf("데이터베이스에 연결할 수 없습니다: %v", err)
	}
	defer db.Close()

	// HTML 템플릿 엔진 설정
	templateDir := "./web/templates"
	engine := html.New(templateDir, ".html")
	engine.Reload(cfg.Debug) // 디버그 모드에서 템플릿 자동 리로드 활성화

	// 템플릿 함수 등록
	engine.AddFunc("json", func(v interface{}) string {
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return "Error: Could not marshal JSON"
		}
		return string(jsonBytes)
	})

	// 다른 템플릿 함수 등록
	engine.AddFunc("add", func(a, b int) int {
		return a + b
	})

	engine.AddFunc("parseJSON", func(s string) interface{} {
		var result any
		err := json.Unmarshal([]byte(s), &result)
		if err != nil {
			return nil
		}
		return result
	})

	engine.AddFunc("iterate", func(start, end int) []int {
		var result []int
		for i := start; i <= end; i++ {
			result = append(result, i)
		}
		return result
	})

	// Fiber 앱 생성
	app := fiber.New(fiber.Config{
		Views:       engine,
		ViewsLayout: "layouts/base",
	})

	// 저장소 초기화
	userRepo := repository.NewUserRepository(db)
	boardRepo := repository.NewBoardRepository(db)

	// 서비스 초기화
	authService := service.NewAuthService(userRepo)
	boardService := service.NewBoardService(boardRepo, db)
	dynamicBoardService := service.NewDynamicBoardService(db)

	// 핸들러 초기화
	authHandler := handlers.NewAuthHandler(authService)
	boardHandler := handlers.NewBoardHandler(boardService)
	adminHandler := handlers.NewAdminHandler(dynamicBoardService, boardService)

	// 인증 미들웨어
	authMiddleware := middleware.NewAuthMiddleware(authService)
	adminMiddleware := middleware.NewAdminMiddleware(authService)

	// 미들웨어 설정
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New())

	// 디버깅 미들웨어 추가 (가장 먼저 실행되도록)
	app.Use(middleware.RequestDebugger())

	// 전역 인증 미들웨어 (모든 요청에서 인증 정보 확인)
	app.Use(middleware.GlobalAuth(authService))

	// 정적 파일 제공
	app.Static("/static", "./web/static")

	// 라우트 설정

	// 인증 관련 라우트
	auth := app.Group("/auth")
	auth.Get("/login", authHandler.LoginPage)
	auth.Post("/login", authHandler.Login)
	auth.Get("/register", authHandler.RegisterPage)
	auth.Post("/register", authHandler.Register)
	auth.Get("/logout", authHandler.Logout)

	// 사용자 프로필 라우트 (인증 필요)
	user := app.Group("/user", authMiddleware.RequireAuth)
	user.Get("/profile", authHandler.ProfilePage)
	user.Post("/profile", authHandler.UpdateProfile)

	// 게시판 라우트
	// boards := app.Group("/boards")
	// boards.Get("/", boardHandler.ListBoards)
	// boards.Get("/:boardID", boardHandler.GetBoard)
	// boards.Get("/:boardID/posts", boardHandler.ListPosts)
	// boards.Get("/:boardID/posts/:postID", boardHandler.GetPost)

	// 수정된 버전: 모든 게시판 관련 경로에 인증 요구 (테스트용)
	boards := app.Group("/boards", authMiddleware.RequireAuth) // 모든 게시판 경로에 인증 필요
	boards.Get("/", boardHandler.ListBoards)
	boards.Get("/:boardID", boardHandler.GetBoard)
	boards.Get("/:boardID/posts", boardHandler.ListPosts)
	boards.Get("/:boardID/posts/:postID", boardHandler.GetPost)

	// 게시물 작성/수정/삭제 (인증 필요)
	boardsAuth := boards.Group("", authMiddleware.RequireAuth)
	boardsAuth.Get("/:boardID/posts/create", boardHandler.CreatePostPage)
	boardsAuth.Post("/:boardID/posts", boardHandler.CreatePost)
	boardsAuth.Get("/:boardID/posts/:postID/edit", boardHandler.EditPostPage)
	boardsAuth.Put("/:boardID/posts/:postID", boardHandler.UpdatePost)
	boardsAuth.Delete("/:boardID/posts/:postID", boardHandler.DeletePost)

	// 관리자 라우트 (관리자 권한 필요)
	admin := app.Group("/admin", authMiddleware.RequireAuth, adminMiddleware.RequireAdmin)
	admin.Get("/", adminHandler.Dashboard)
	admin.Get("/boards", adminHandler.ListBoards)
	admin.Get("/boards/create", adminHandler.CreateBoardPage)
	admin.Post("/boards", adminHandler.CreateBoard)
	admin.Get("/boards/:boardID/edit", adminHandler.EditBoardPage)
	admin.Put("/boards/:boardID", adminHandler.UpdateBoard)
	admin.Delete("/boards/:boardID", adminHandler.DeleteBoard)

	// 루트 라우트
	app.Get("/", func(c *fiber.Ctx) error {
		// 사용자 정보가 있는지 디버깅 로그
		if user := c.Locals("user"); user != nil {
			log.Printf("루트 경로 - 사용자 정보 있음")
		} else {
			log.Printf("루트 경로 - 사용자 정보 없음")
		}

		return c.Redirect("/boards")
	})

	// 서버 시작
	go func() {
		if err := app.Listen(cfg.ServerAddress); err != nil {
			log.Fatalf("서버 시작 실패: %v", err)
		}
	}()

	// 종료 시그널 처리
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Println("서버를 종료합니다...")
	if err := app.Shutdown(); err != nil {
		log.Fatalf("서버 종료 실패: %v", err)
	}
}
