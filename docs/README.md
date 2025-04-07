# 초기 관리자 계정 설정 기능 통합 가이드

이 기능은 시스템이 처음 시작될 때 관리자 계정이 없는 경우 초기 설정 페이지로 리다이렉트하여 관리자 계정을 생성할 수 있게 합니다.

## 1. 파일 추가

다음 파일들이 새로 생성되었습니다:

- `web/templates/admin/setup.html` - 초기 관리자 설정 페이지
- `internal/handlers/setup_handler.go` - 설정 페이지 핸들러
- `internal/service/setup_service.go` - 설정 관련 서비스
- `internal/middleware/setup_middleware.go` - 설정 관련 미들웨어
- `web/static/js/pages/admin-setup.js` - 설정 페이지 JavaScript

## 2. `config/config.go` 수정

`Config` 구조체에 다음 필드를 추가하세요:

```go
type Config struct {
    // 기존 필드들...
    RequireSetup   bool   // 초기 설정이 필요한지 여부
}
```

`Load()` 함수에 다음 코드를 반환 직전에 추가하세요:

```go
// 초기 설정 필요 여부 설정
requireSetup := os.Getenv("REQUIRE_SETUP") == "true"
// 환경 변수가 설정되어 있지 않으면 기본값은 true (자동 감지)
if os.Getenv("REQUIRE_SETUP") == "" {
    requireSetup = true
}

// 기존 반환 코드...
return &Config{
    // 기존 필드들...
    RequireSetup:   requireSetup,
}, nil
```

## 3. `.env.example` 수정

`.env.example` 파일에 다음 라인을 추가하세요:

```
# 초기 설정 관련
REQUIRE_SETUP=true    # true이면 최초 실행 시 관리자 설정 페이지로 리다이렉트
```

## 4. `cmd/main.go` 수정

`main.go` 파일의 서비스 및 핸들러 초기화 부분에 다음 코드를 추가하세요:

```go
// 저장소 초기화 다음, 서비스 초기화 부분에 추가
setupService := service.NewSetupService(userRepo)
```

핸들러 초기화 부분에 다음 코드를 추가하세요:

```go
// 핸들러 초기화 부분에 추가
setupHandler := handlers.NewSetupHandler(authService, setupService)
```

`setupRoutes` 함수에 다음 라우트를 추가하세요:

```go
// 초기 설정 관련 라우트
app.Get("/admin/setup", setupHandler.SetupPage)
app.Post("/admin/setup", setupHandler.SetupAdmin)
```

미들웨어 설정 부분에 다음 코드를 추가하세요 (예: `setupMiddleware` 함수 마지막에):

```go
// 초기 설정 미들웨어 - 관리자가 없으면 설정 페이지로 리다이렉트
if cfg.RequireSetup {
    app.Use(middleware.SetupMiddleware(setupService))
}
```

## 5. 빌드 및 실행

이제 시스템을 빌드하고 실행하세요:

```bash
go build -o go-board ./cmd
./go-board
```

처음 실행 시, 관리자 계정이 없다면 `/admin/setup` 페이지로 리다이렉트되어 초기 관리자 계정을 설정할 수 있습니다.