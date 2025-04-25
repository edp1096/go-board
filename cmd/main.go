package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	goboard "github.com/edp1096/go-board"
	"github.com/edp1096/go-board/config"
	"github.com/edp1096/go-board/internal/handlers"
	"github.com/edp1096/go-board/internal/middleware"
	"github.com/edp1096/go-board/internal/repository"
	"github.com/edp1096/go-board/internal/service"
	"github.com/edp1096/go-board/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	flogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	"github.com/microcosm-cc/bluemonday"
)

const APP_VERSION = "v0.0.22"

func main() {
	// // 시작 시간 기록
	// startTime := time.Now()

	// 명령행 인자 처리 (내보내기, 도움말 등)
	shouldExit, err := handleCommandLineArgs()
	if err != nil {
		os.Exit(1)
	}
	if shouldExit {
		os.Exit(0)
	}

	// 환경 설정 로드
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("설정을 로드할 수 없습니다: %v", err)
	}

	// 데이터베이스 연결
	database, err := config.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("데이터베이스에 연결할 수 없습니다: %v", err)
	}
	defer database.Close()
	db := database.DB

	// HTML 템플릿 엔진 설정 - 단순화된 자동 감지 방식
	var engine *html.Engine

	// 현재 디렉토리 기준으로 템플릿 디렉토리 확인
	if _, err := os.Stat("./web/templates"); err == nil {
		// 실제 파일 시스템 사용
		engine = html.New("./web/templates", ".html")
	} else {
		// 임베디드 파일 시스템 사용
		engine = html.NewFileSystem(goboard.GetTemplatesFS(), ".html")
	}

	// 공통 설정
	engine.Reload(cfg.Debug)

	/* 템플릿 함수 */
	// sanitizer
	sanitizer := bluemonday.NewPolicy()
	sanitizer.AllowElements(
		"p", "br", "h1", "h2", "h3", "h4", "h5", "h6",
		"blockquote", "pre", "code", "em", "strong", "del",
		"ul", "ol", "li", "a", "img", "table", "thead", "tbody",
		"tr", "th", "td", "hr", "div", "span", "iframe",
	)
	sanitizer.AllowAttrs("style").OnElements("p")
	sanitizer.AllowAttrs("href").OnElements("a")
	sanitizer.AllowAttrs("src", "alt", "title").OnElements("img")
	sanitizer.AllowAttrs("class").Globally()
	sanitizer.AllowAttrs("animate").OnElements("img")

	sanitizer.AllowElements("iframe")
	sanitizer.AllowAttrs("class").Matching(bluemonday.Number).OnElements("iframe")
	sanitizer.AllowAttrs("width").Matching(bluemonday.Number).OnElements("iframe")
	sanitizer.AllowAttrs("height").Matching(bluemonday.Number).OnElements("iframe")
	sanitizer.AllowAttrs("src").OnElements("iframe")
	sanitizer.AllowAttrs("frameborder").Matching(bluemonday.Number).OnElements("iframe")
	sanitizer.AllowAttrs("allow").Matching(regexp.MustCompile(`[a-z; -]*`)).OnElements("iframe")
	sanitizer.AllowAttrs("allowfullscreen").OnElements("iframe")

	engine.AddFunc("unescape", func(s any) template.HTML {
		sanitizedString := sanitizer.Sanitize(s.(string))
		return template.HTML(sanitizedString)
	})

	// Extract plain text to HTML
	engine.AddFunc("plaintext", utils.PlainText)

	// json
	engine.AddFunc("json", func(v any) string {
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return "Error: Could not marshal JSON"
		}
		return string(jsonBytes)
	})

	// 문자열 분할 함수
	engine.AddFunc("split", func(s, sep string) []string {
		return utils.Split(s, sep)
	})

	// 바이트 배열을 문자열로 변환하는 도우미 함수 추가
	engine.AddFunc("toUTF8", func(v any) string {
		switch value := v.(type) {
		case []byte:
			return string(value)
		case string:
			return value
		default:
			return fmt.Sprintf("%v", value)
		}
	})

	engine.AddFunc("add", func(a, b int) int {
		return a + b
	})

	engine.AddFunc("parseJSON", func(s string) any {
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

	// 경로 관련 함수
	engine.AddFunc("jsPath", func(fileName string) string {
		return "/static/js/" + fileName
	})

	engine.AddFunc("cssPath", func(fileName string) string {
		return "/static/css/" + fileName
	})

	// NullString 처리 헬퍼 함수
	engine.AddFunc("getNullString", func(v any) string {
		if v == nil {
			return ""
		}

		// sql.NullString 처리
		if ns, ok := v.(sql.NullString); ok {
			if ns.Valid {
				return ns.String
			}
			return ""
		}

		// 일반 문자열 처리
		return fmt.Sprintf("%v", v)
	})

	// 레퍼러 봇 확인 함수
	engine.AddFunc("checkBot", func(userAgents string) bool {
		// results := make([]bool, len(userAgents))

		ua := strings.ToLower(userAgents)
		result := strings.Contains(ua, "bot") ||
			strings.Contains(ua, "crawler") ||
			strings.Contains(ua, "spider") ||
			strings.Contains(ua, "slurp") ||
			strings.Contains(ua, "search")

		if result {
			return true
		}

		// for i, ua := range userAgents {
		// 	ua = strings.ToLower(ua)
		// 	results[i] = strings.Contains(ua, "bot") ||
		// 		strings.Contains(ua, "crawler") ||
		// 		strings.Contains(ua, "spider") ||
		// 		strings.Contains(ua, "slurp") ||
		// 		strings.Contains(ua, "search")

		// 	return true
		// }

		return false
	})

	// Fiber 앱 생성
	app := fiber.New(fiber.Config{
		IdleTimeout:           5 * time.Second,
		Views:                 engine,
		ViewsLayout:           "layouts/base",
		DisableStartupMessage: true,
		StreamRequestBody:     true,
		ReadBufferSize:        8192,
		ProxyHeader:           "X-Forwarded-For",
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
		// BodyLimit:             cfg.MaxBodyLimit,
		// Prefork:               true,
		// Immutable:             true,
		// EnablePrintRoutes:     true,
	})

	// 업로드 디렉토리 확인
	uploadDir := "uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			log.Fatalf("업로드 디렉토리 생성 실패: %v", err)
		}
		log.Printf("업로드 디렉토리 생성됨: %s", uploadDir)
	}

	// 계층 구성 (의존성 주입)
	// 저장소 초기화
	systemSettingsRepo := repository.NewSystemSettingsRepository(db)
	userRepo := repository.NewUserRepository(db)
	boardRepo := repository.NewBoardRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	attachmentRepo := repository.NewAttachmentRepository(db)
	participantRepo := repository.NewBoardParticipantRepository(db)
	referrerRepo := repository.NewReferrerRepository(db)

	// 서비스 초기화
	setupService := service.NewSetupService(userRepo)
	systemSettingsService := service.NewSystemSettingsService(systemSettingsRepo)
	authService := service.NewAuthService(userRepo, systemSettingsService)
	boardService := service.NewBoardService(boardRepo, participantRepo, db)
	dynamicBoardService := service.NewDynamicBoardService(db)
	commentService := service.NewCommentService(commentRepo, boardRepo)
	qnaService := service.NewQnAService(db, boardRepo, boardService)
	uploadService := service.NewUploadService(attachmentRepo)
	sitemapService := service.NewSitemapService(boardRepo, boardService)
	referrerService := service.NewReferrerService(referrerRepo)

	// 핸들러 초기화
	setupHandler := handlers.NewSetupHandler(authService, setupService)
	systemSettingsHandler := handlers.NewSystemSettingsHandler(systemSettingsService)
	authHandler := handlers.NewAuthHandler(authService)
	boardHandler := handlers.NewBoardHandler(boardService, commentService, uploadService, cfg)
	commentHandler := handlers.NewCommentHandler(commentService, boardService)
	qnaHandler := handlers.NewQnAHandler(boardService, qnaService)
	uploadHandler := handlers.NewUploadHandler(uploadService, boardService, cfg)
	adminHandler := handlers.NewAdminHandler(dynamicBoardService, boardService, authService)
	sitemapHandler := handlers.NewSitemapHandler(sitemapService)
	robotsHandler := handlers.NewRobotsHandler()
	referrerHandler := handlers.NewReferrerHandler(referrerService)
	whoisHandler := handlers.NewWhoisHandler()

	// 미들웨어
	authMiddleware := middleware.NewAuthMiddleware(authService)
	boardAccessMiddleware := middleware.NewBoardAccessMiddleware(boardService)
	adminMiddleware := middleware.NewAdminMiddleware(authService)
	referrerMiddleware := middleware.NewReferrerMiddleware(referrerService, cfg)

	// 미들웨어 설정
	setupMiddleware(app, cfg, setupService, authService)
	app.Use(referrerMiddleware.CaptureReferrer)

	// 정적 파일 제공 - 단순화된 자동 감지 방식
	var staticFS http.FileSystem

	// 현재 디렉토리 기준으로 정적 파일 디렉토리 확인
	if _, err := os.Stat("./web/static"); err == nil {
		// 실제 파일 시스템 사용
		staticFS = http.Dir("./web/static")
		// log.Println("실제 정적 파일 디렉토리 사용: ./web/static")
	} else {
		// 임베디드 파일 시스템 사용
		staticFS = goboard.GetStaticFS()
		// log.Println("임베디드 정적 파일 시스템 사용 (실제 디렉토리 없음)")
	}

	// 정적 파일 제공 - 하이브리드 파일시스템 사용
	app.Use("/static", filesystem.New(filesystem.Config{
		// Root:         goboard.GetStaticFS(),
		Root:         staticFS,
		Browse:       true,
		Index:        "index.html",
		NotFoundFile: "404.html",
	}))

	// 라우트 설정
	setupRoutes(
		app,
		setupHandler,
		systemSettingsHandler,
		authHandler,
		boardHandler,
		commentHandler,
		qnaHandler,
		uploadHandler,
		adminHandler,
		sitemapHandler,
		robotsHandler,
		referrerHandler,
		whoisHandler,
		authMiddleware,
		boardAccessMiddleware,
		adminMiddleware,
	)

	// 서버 시작
	go startServer(app, cfg.ServerAddress)

	// // 준비 시간 계산 및 출력
	// readyTime := time.Since(startTime)
	// log.Printf("서버가 %.2f초 만에 준비되었습니다", readyTime.Seconds())

	// 종료 시그널 처리
	handleShutdown(app)
}

// 명령행 인자 처리 및 도움말 출력
func handleCommandLineArgs() (bool, error) {
	// 도움말 표시 처리
	helpCmd := flag.NewFlagSet("help", flag.ExitOnError)

	// export-web 명령어 처리
	exportWebCmd := flag.NewFlagSet("export-web", flag.ExitOnError)
	exportPath := exportWebCmd.String("path", "./web", "웹 콘텐츠를 내보낼 경로")

	// export-env 명령어 추가
	exportEnvCmd := flag.NewFlagSet("export-env", flag.ExitOnError)
	exportEnvPath := exportEnvCmd.String("path", "./.env.example", ".env.example 파일을 내보낼 경로")

	// 명령행 인자가 없으면 정상 실행
	if len(os.Args) < 2 {
		return false, nil
	}

	// 명령 처리
	switch os.Args[1] {
	case "help", "--help", "-h":
		helpCmd.Parse(os.Args[2:])
		printHelp()
		return true, nil

	case "version", "--version", "-v":
		fmt.Println(APP_VERSION)
		return true, nil

	case "export-web":
		exportWebCmd.Parse(os.Args[2:])
		if err := goboard.ExportWebContent(*exportPath); err != nil {
			fmt.Fprintf(os.Stderr, "오류: %s\n", err)
			return true, err
		}
		fmt.Printf("웹 콘텐츠가 %s 경로에 성공적으로 내보내졌습니다\n", *exportPath)
		return true, nil

	case "export-env":
		exportEnvCmd.Parse(os.Args[2:])
		if err := goboard.ExportEnvExample(*exportEnvPath); err != nil {
			fmt.Fprintf(os.Stderr, "오류: %s\n", err)
			return true, err
		}
		return true, nil

	default:
		// 알 수 없는 명령 처리
		fmt.Fprintf(os.Stderr, "알 수 없는 명령: %s\n", os.Args[1])
		printHelp()
		return true, fmt.Errorf("알 수 없는 명령: %s", os.Args[1])
	}
}

// 도움말 출력
func printHelp() {
	progName := filepath.Base(os.Args[0])
	fmt.Printf("사용법: %s [명령] [옵션]\n\n", progName)
	fmt.Println("명령:")
	fmt.Println("  help\t\t이 도움말을 표시합니다")
	fmt.Println("  version\t버전 정보를 표시합니다")
	fmt.Println("  export-web\t웹 콘텐츠를 내보냅니다")
	fmt.Println("    옵션:")
	fmt.Println("      -path string\t내보낼 경로를 지정합니다 (기본값: \"./web\")")
	fmt.Println("  export-env\t.env.example 파일을 내보냅니다")
	fmt.Println("    옵션:")
	fmt.Println("      -path string\t내보낼 경로를 지정합니다 (기본값: \"./.env.example\")")
	fmt.Println()
}

// setupMiddleware 앱에 필요한 미들웨어를 설정
func setupMiddleware(
	app *fiber.App,
	cfg *config.Config,
	setupService service.SetupService,
	authService service.AuthService,
) {
	// 기본 미들웨어
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE",
	}))

	app.Use(middleware.BodyLimitMiddleware(cfg))

	app.Use(func(c *fiber.Ctx) error {
		/*
			# NginX 설정 예시
			server {
				listen 80;
				server_name www.example.com;

				location / {
					proxy_set_header Host $host;
					proxy_set_header X-Real-IP $remote_addr;
					proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
					proxy_set_header X-Forwarded-Proto $scheme;
					proxy_set_header X-Forwarded-Host $host;
					proxy_pass http://localhost:3000;
				}
			}
		*/
		// X-Forwarded-Proto와 X-Forwarded-Host 헤더 확인
		proto := c.Get("X-Forwarded-Proto")
		if proto == "" {
			proto = "http"
		}

		host := c.Get("X-Forwarded-Host")
		if host == "" {
			host = c.Hostname()
		}

		baseURL := proto + "://" + host
		c.Locals("baseURL", baseURL)

		return c.Next()
	})

	// // 로거 설정
	// app.Use(flogger.New(flogger.Config{
	// 	Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
	// }))
	// 로거 설정 (디버그 모드일 때만)
	if cfg.Debug {
		app.Use(flogger.New(flogger.Config{
			Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
		}))
	}

	// UTF-8 인코딩 강제 적용 미들웨어
	app.Use(func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.Next()
	})

	// CSRF 미들웨어
	app.Use(middleware.CSRF())

	// 전역 인증 미들웨어 (모든 요청에서 인증 정보 확인)
	app.Use(middleware.GlobalAuth(authService))

	if cfg.RequireSetup {
		app.Use(middleware.SetupMiddleware(setupService))
	}

	// MIME 타입
	app.Use("/uploads/*", func(c *fiber.Ctx) error {
		err := c.Next()

		// 파일을 성공적으로 찾았을 때만 MIME 타입 설정
		if err == nil {
			utils.SetMimeTypeHeader(c, c.Path())
		}

		return err
	})
}

// setupRoutes는 앱의 라우트를 설정합니다
func setupRoutes(
	app *fiber.App,
	setupHandler *handlers.SetupHandler,
	systemSettingsHandler *handlers.SystemSettingsHandler,
	authHandler *handlers.AuthHandler,
	boardHandler *handlers.BoardHandler,
	commentHandler *handlers.CommentHandler,
	qnaHandler *handlers.QnAHandler,
	uploadHandler *handlers.UploadHandler,
	adminHandler *handlers.AdminHandler,
	sitemapHandler *handlers.SitemapHandler,
	robotsHandler *handlers.RobotsHandler,
	referrerHandler *handlers.ReferrerHandler,
	whoisHandler *handlers.WhoisHandler,
	authMiddleware middleware.AuthMiddleware,
	boardAccessMiddleware middleware.BoardAccessMiddleware,
	adminMiddleware middleware.AdminMiddleware,
) {
	// favicon
	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		return c.SendFile("web/static/favicon.ico")
	})

	// 초기 설정 라우트
	app.Get("/admin/setup", setupHandler.SetupPage)
	app.Post("/admin/setup", setupHandler.SetupAdmin)

	app.Get("/sitemap.xml", sitemapHandler.GetSitemapIndex)
	app.Get("/sitemap_:index.xml", sitemapHandler.GetSitemapFile)
	app.Get("/robots.txt", robotsHandler.GetRobots)

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

	// 게시판 라우트 (열람은 인증 없이 가능)
	boards := app.Group("/boards")
	boards.Get("/", boardHandler.ListBoards)

	// /boards/:boardID 라우트
	boardsWithID := boards.Group("/:boardID", boardAccessMiddleware.CheckBoardAccess)
	boardsWithID.Get("", boardHandler.GetBoard)
	boardsWithID.Get("/posts", boardHandler.ListPosts)
	boards.Get("/:boardID/posts/create", authMiddleware.RequireAuth, boardHandler.CreatePostPage)
	boardsWithID.Get("/posts/:postID", boardHandler.GetPost)

	// 게시물 작성/수정/삭제 (인증 필요)
	boardsAuthWithID := boards.Group("/:boardID", authMiddleware.RequireAuth)
	boardsAuthWithID.Post("/posts", boardHandler.CreatePost)
	boardsAuthWithID.Get("/posts/create", boardHandler.CreatePostPage)
	boardsAuthWithID.Get("/posts/:postID/edit", boardHandler.EditPostPage)
	boardsAuthWithID.Put("/posts/:postID", boardHandler.UpdatePost)
	boardsAuthWithID.Delete("/posts/:postID", boardHandler.DeletePost)

	// 관리자 라우트 (관리자 권한 필요)
	admin := app.Group("/admin", authMiddleware.RequireAuth, adminMiddleware.RequireAdmin)
	admin.Get("/system-settings", systemSettingsHandler.SettingsPage)
	admin.Post("/system-settings", systemSettingsHandler.UpdateSettings)
	admin.Get("/referrer-stats", referrerHandler.ReferrerStatsPage)

	admin.Get("/", adminHandler.Dashboard)

	// 게시판 관리 라우트
	admin.Get("/boards", adminHandler.ListBoards)
	admin.Get("/boards/create", adminHandler.CreateBoardPage)
	admin.Post("/boards", adminHandler.CreateBoard)
	admin.Get("/boards/:boardID/edit", adminHandler.EditBoardPage)
	admin.Put("/boards/:boardID", adminHandler.UpdateBoard)
	admin.Delete("/boards/:boardID", adminHandler.DeleteBoard)

	// 사용자 관리 라우트
	admin.Get("/users", adminHandler.ListUsers)
	admin.Get("/user-approval", adminHandler.UserApprovalPage)
	admin.Put("/users/:userID/approve", adminHandler.ApproveUser)
	admin.Put("/users/:userID/reject", adminHandler.RejectUser)
	admin.Get("/users/create", adminHandler.CreateUserPage)
	admin.Post("/users", adminHandler.CreateUser)
	admin.Get("/users/:userID/edit", adminHandler.EditUserPage)
	admin.Put("/users/:userID", adminHandler.UpdateUser)
	admin.Delete("/users/:userID", adminHandler.DeleteUser)
	admin.Put("/users/:userID/role", adminHandler.UpdateUserRole)

	// API 라우트 (댓글 기능)
	api := app.Group("/api")

	// Whois 정보 조회 라우트
	api.Get("/whois", whoisHandler.GetWhoisInfo)

	// 업로드 관련 라우트 추가
	api.Post("/boards/:boardID/upload", authMiddleware.RequireAuth, uploadHandler.UploadImages)
	api.Post("/boards/:boardID/posts/:postID/attachments", authMiddleware.RequireAuth, uploadHandler.UploadAttachments)
	api.Get("/boards/:boardID/posts/:postID/attachments", uploadHandler.GetAttachments)
	app.Get("/attachments/:attachmentID/download", uploadHandler.DownloadAttachment)
	api.Delete("/attachments/:attachmentID", authMiddleware.RequireAuth, uploadHandler.DeleteAttachment)

	// 댓글 관련 API 라우트
	commentsAPI := api.Group("/boards/:boardID/posts/:postID/comments")
	commentsAPI.Get("/", commentHandler.GetComments)
	commentsAPI.Post("/", authMiddleware.RequireAuth, commentHandler.CreateComment)

	// 댓글 수정/삭제 API 라우트
	commentActions := api.Group("/comments/:commentID", authMiddleware.RequireAuth)
	commentActions.Put("/", commentHandler.UpdateComment)
	commentActions.Delete("/", commentHandler.DeleteComment)

	// Q&A 관련 API 라우트
	qnaAPI := api.Group("/boards/:boardID/posts/:postID")
	qnaAPI.Get("/answers", qnaHandler.GetAnswers)
	qnaAPI.Post("/answers", authMiddleware.RequireAuth, qnaHandler.CreateAnswer)
	qnaAPI.Put("/status", authMiddleware.RequireAuth, qnaHandler.UpdateQuestionStatus)
	qnaAPI.Put("/best-answer", authMiddleware.RequireAuth, qnaHandler.SetBestAnswer)
	qnaAPI.Post("/vote", authMiddleware.RequireAuth, qnaHandler.VoteQuestion)
	qnaAPI.Get("/vote-count", qnaHandler.GetQuestionVoteCount)

	// 답변 관련 API 라우트
	answerAPI := api.Group("/answers/:answerID")
	answerAPI.Put("/", authMiddleware.RequireAuth, qnaHandler.UpdateAnswer)
	answerAPI.Delete("/", authMiddleware.RequireAuth, qnaHandler.DeleteAnswer)
	answerAPI.Post("/replies", authMiddleware.RequireAuth, qnaHandler.CreateAnswerReply)
	answerAPI.Post("/vote", authMiddleware.RequireAuth, qnaHandler.VoteAnswer)

	adminAPI := api.Group("/admin", authMiddleware.RequireAuth, adminMiddleware.RequireAdmin)
	adminAPI.Put("/boards/:boardID/order", adminHandler.ChangeOrder)
	adminAPI.Get("/boards/:boardID/participants", adminHandler.GetBoardParticipants)
	adminAPI.Post("/boards/:boardID/participants", adminHandler.AddBoardParticipant)
	adminAPI.Put("/boards/:boardID/participants/:userID/role", adminHandler.UpdateBoardParticipantRole)
	adminAPI.Delete("/boards/:boardID/participants/:userID", adminHandler.RemoveBoardParticipant)
	adminAPI.Get("/users/search", adminHandler.SearchUsers)
	adminAPI.Put("/users/:userID/status", adminHandler.UpdateUserStatus)

	// 레퍼러
	adminAPI.Get("/referrer-stats", referrerHandler.GetReferrerData)

	// 첨부파일 다운로드
	app.Get("/attachments/:attachmentID/download", uploadHandler.DownloadAttachment)

	// 업로드된 파일 정적 제공
	app.Static("/uploads", "./uploads", fiber.Static{
		Browse: false,
	})

	// 썸네일 생성 미들웨어 - 요청된 썸네일이 없으면 자동 생성
	app.Use("/uploads", func(c *fiber.Ctx) error {
		// 요청 경로가 이미지 파일인지 확인
		path := c.Path()

		// 요청 경로에 '/thumbs/'가 포함되어 있는지 확인
		if strings.Contains(path, "/thumbs/") {
			// 서버 내 실제 파일 경로 계산
			filePath := filepath.Join(".", path)

			// thumbs 부분을 제거한 원본 이미지 경로 생성
			originalPath := strings.Replace(filePath, "/thumbs/", "/", 1)

			// 원본 이미지는 있지만 썸네일이 없는 경우 썸네일 생성
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				if _, err := os.Stat(originalPath); err == nil {
					// 이미지 파일인지 확인
					if utils.IsImageFile(originalPath) {
						thumbDir := filepath.Dir(filePath)
						if err := os.MkdirAll(thumbDir, 0755); err == nil {
							// 썸네일 생성
							if err := utils.EnsureThumbnail(originalPath); err != nil {
								fmt.Printf("자동 썸네일 생성 실패 (%s): %v\n", originalPath, err)
							}
						}
					}
				}
			}
		}

		// 다음 미들웨어로 계속 진행
		return c.Next()
	})

	// 루트 라우트
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/boards")
	})

	// 404 페이지
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).Render("error", fiber.Map{
			"title":   "페이지를 찾을 수 없습니다",
			"message": "요청하신 페이지를 찾을 수 없습니다.",
		})
	})
}

// startServer는 서버를 시작합니다
func startServer(app *fiber.App, address string) {
	fmt.Printf("서버를 시작합니다: http://%s\n", address)
	if err := app.Listen(address); err != nil {
		log.Fatalf("서버 시작 실패: %v", err)
	}
}

// handleShutdown은 서버 종료 시그널을 처리합니다
func handleShutdown(app *fiber.App) {
	// 종료 시그널 처리
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("서버를 종료합니다..")
	shutdownTimeout := 1 * time.Second
	if err := app.ShutdownWithTimeout(shutdownTimeout); err != nil {
		// log.Fatalf("서버 종료 실패: %v", err)
	}
	// fmt.Println("서버가 안전하게 종료되었습니다.")
}
