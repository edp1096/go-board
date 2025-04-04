# Go-Board 프로젝트 구조

```
.
├── cmd/                    # 애플리케이션 진입점
│   ├── help.go             # 명령행 도움말 처리
│   └── main.go             # 메인 애플리케이션 시작점
│
├── config/                 # 설정 관련 코드
│   ├── config.go           # 환경 설정 로드 및 관리
│   └── database.go         # 데이터베이스 연결 설정
│
├── internal/               # 내부 구현 코드
│   ├── handlers/           # HTTP 요청 핸들러
│   │   ├── admin_handler.go # 관리자 기능 핸들러
│   │   ├── auth_handler.go  # 인증 관련 핸들러
│   │   └── board_handler.go # 게시판 핸들러
│   │
│   ├── middleware/         # HTTP 미들웨어
│   │   ├── admin_middleware.go # 관리자 권한 검증
│   │   ├── auth_middleware.go  # 인증 검증
│   │   ├── csrf_middleware.go  # CSRF 보호
│   │   └── global_auth.go      # 전역 인증 상태 확인
│   │
│   ├── models/             # 데이터 모델
│   │   ├── board.go        # 게시판 모델
│   │   ├── post.go         # 게시물 모델
│   │   └── user.go         # 사용자 모델
│   │
│   ├── repository/         # 데이터 액세스 레이어
│   │   ├── board_repository.go # 게시판 데이터 액세스
│   │   └── user_repository.go  # 사용자 데이터 액세스
│   │
│   ├── service/            # 비즈니스 로직
│   │   ├── auth_service.go       # 인증 관련 서비스
│   │   ├── board_service.go      # 게시판 서비스
│   │   └── dynamic_board_service.go # 게시판 서비스
│   │
│   └── utils/              # 유틸리티 함수
│       ├── convert_utils.go  # 데이터 변환 유틸리티
│       ├── render_utils.go   # 렌더링 유틸리티
│       └── template_utils.go # 템플릿 관련 유틸리티
│
├── migrations/             # 데이터베이스 마이그레이션 파일
│   ├── backup/             # 보관용 - 이전 마이그레이션
│   ├── mysql/              # MySQL용 마이그레이션
│   └── postgres/           # PostgreSQL용 마이그레이션
│
├── web/                    # 프론트엔드 파일
│   ├── static/             # 정적 파일
│   │   ├── css/            # CSS 스타일시트
│   │   └── js/             # JavaScript 파일
│   │       └── pages/      # 페이지별 스크립트
│   │
│   └── templates/          # HTML 템플릿
│       ├── admin/          # 관리자 페이지 템플릿
│       ├── auth/           # 인증 관련 템플릿
│       ├── board/          # 게시판 템플릿
│       ├── layouts/        # 레이아웃 템플릿
│       └── partials/       # 부분 템플릿
│
├── .env                    # 환경 변수 설정 (gitignore)
├── Makefile                # 빌드 및 실행 스크립트
├── go.mod                  # Go 모듈 정의
├── go.sum                  # Go 모듈 체크섬
└── README.md               # 프로젝트 설명서
```

## 주요 디렉토리 설명

1. **cmd**: 애플리케이션의 진입점으로, main 함수와 명령행 인터페이스 관련 코드가 위치함
2. **config**: 애플리케이션 설정 로드 및 관리 코드가 포함됨
3. **internal**: 패키지 외부에서 접근하지 않는 내부 구현 코드
   - **handlers**: HTTP 요청을 처리하는 핸들러 함수들
   - **middleware**: 요청 전/후에 처리를 수행하는 미들웨어
   - **models**: 데이터 구조를 정의하는 모델
   - **repository**: 데이터베이스 액세스 로직
   - **service**: 비즈니스 로직을 구현하는 서비스
   - **utils**: 공통 유틸리티 함수들
4. **migrations**: 데이터베이스 스키마 변경을 관리하는 마이그레이션 파일들
5. **web**: 프론트엔드 관련 파일들
   - **static**: 정적 자원 (CSS, JavaScript 등)
   - **templates**: HTML 템플릿 파일들