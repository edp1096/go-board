# Claude Conversation

## Prompt:
네, 가능합니다. 여러 방법으로 연동할 수 있는데, 몇 가지 대표적인 방법은 다음과 같습니다:
1. **API 기반 연동:** Golang 게시판에서 RESTful API를 제공하고, PHP나 다른 시스템에서는 해당 API를 호출하여 사용자 정보를 검증하거나 동기화하는 방식입니다.
2. **토큰 기반 인증:** JWT(JSON Web Token)나 OAuth2와 같은 토큰 방식을 사용해 사용자가 한 번 로그인하면 토큰을 발급받아 여러 시스템에서 인증을 공유할 수 있습니다. 이를 통해 SSO(Single Sign-On)와 유사한 환경을 구성할 수 있습니다.
3. **SSO(Single Sign-On) 솔루션 사용:** OpenID Connect, CAS, SAML 등의 프로토콜을 활용하면 여러 시스템 간에 중앙 집중식 인증 시스템을 구축하여 회원 정보를 공유할 수 있습니다.
4. **공유 데이터베이스 사용:** 만약 보안 요구사항을 만족할 수 있다면, Golang 게시판과 다른 시스템이 동일한 사용자 데이터베이스를 참조하여 회원 정보를 통합 관리할 수도 있습니다.
각 방법마다 구현 난이도와 보안 고려사항이 다르므로, 시스템의 요구사항과 환경에 맞는 방식을 선택하는 것이 중요합니다.

프로젝트 규모와 기존 시스템 구조에 따라 다르지만, 일반적으로 **REST API를 통한 JWT 기반 인증 방식**이 가장 쉽고 간편한 방법으로 꼽힙니다.
이유:
* **독립성 및 확장성:** Golang 게시판에서 인증 토큰(JWT)을 발급하고, PHP나 다른 시스템에서는 이를 검증하는 방식은 각 시스템이 서로 독립적으로 동작할 수 있어 확장하기 쉽습니다.
* **구현 난이도:** 이미 HTTP 기반 통신을 하고 있다면, REST API 엔드포인트를 추가하는 것만으로 연동이 가능해 복잡한 설정 없이 구현할 수 있습니다.
* **보안 측면:** 토큰 기반 인증은 중앙 집중형 인증 서버 없이도 안전하게 인증 정보를 주고받을 수 있어, 공유 데이터베이스를 사용하는 방식보다 보안 위험을 줄일 수 있습니다.
다른 대안:
* **공유 데이터베이스:** 두 시스템이 같은 데이터베이스를 참조하게 하는 방법은 구현이 간단할 수 있으나, 보안 및 유지보수 측면에서 단점이 있습니다.
* **SSO 솔루션:** OAuth2, OpenID Connect 등의 SSO 시스템은 보다 중앙집중적인 인증 관리를 가능하게 하지만, 초기 설정과 인프라 구성이 복잡해질 수 있습니다.
따라서, 상대적으로 구현이 쉽고 유지보수 및 보안 측면에서도 효율적인 **REST API와 JWT를 활용한 방법**을 추천합니다.


go-board를 php나 쇼핑몰과 같은 여타 시스템에 활용하고 싶은데, 그 여타 시스템에는 회원관리가 이미 있기 때문에 연동해서 활용하려 해. chatGPT는 이렇게 대답하네. chatGPT 말대로 jwt가 최선인가? 만약 맞다면 어떤식으로 구현하는게 좋을거 같아?

## Response:

이 질문은 "go-board"라는 Go로 작성된 게시판 시스템과 기존의 PHP 쇼핑몰이나 다른 시스템(이미 회원관리 기능이 있는)을 연동하는 방법에 대한 것입니다. 사용자는 ChatGPT가 JWT 방식을 추천했다고 언급하면서, 이것이 최선인지, 그리고 만약 그렇다면 어떻게 구현해야 하는지 물어보고 있습니다.

제공된 코드를 분석해 보면, go-board는 Fiber 웹 프레임워크를 사용하는 Go 기반 게시판 시스템으로, JWT 인증을 이미 구현하고 있습니다. 사용자, 게시판, 댓글 등의 기능을 갖추고 있으며, MySQL과 PostgreSQL을 지원합니다.

다른 시스템과의 회원 연동 방법에 대해 고려해야 할 사항:

1. JWT 인증 방식의 적절성:
   - JWT는 토큰 기반 인증 방식으로, 분산 시스템 간 인증 정보를 안전하게 전달하는 좋은 방법입니다.
   - 서버 간 통신에 적합하며 상태를 유지하지 않아 확장성이 좋습니다.
   - 그러나 구현 방식과 상황에 따라 장단점이 있습니다.

2. 가능한 구현 방식:
   - JWT 활용 방식
   - 공유 데이터베이스 방식
   - API 기반 인증 검증
   - SSO(Single Sign-On) 구현

3. go-board의 현재 인증 시스템:
   - 이미 JWT를 사용하고 있습니다.
   - authService.go에서 토큰 생성과 검증을 처리합니다.
   - auth_middleware.go에서 토큰 검증을 미들웨어로 구현하고 있습니다.

사용자의 질문에 대한 응답으로, JWT가 최선인지, 그리고 어떻게 구현하는 것이 좋을지에 대해 분석적인 답변이 필요합니다.

JWT(JSON Web Token)는 여러 시스템 간 인증 정보를 공유하는 데 좋은 방법 중 하나입니다. go-board 코드를 살펴보니 이미 JWT 인증 시스템을 구현하고 있어서 이를 활용하는 것이 가장 자연스러울 것 같습니다.

JWT가 최선인지는 상황에 따라 달라질 수 있지만, 다음과 같은 이유로 많은 경우에 좋은 선택입니다:

1. **무상태(Stateless)**: 서버에 세션 정보를 저장할 필요가 없어 확장성이 좋습니다.
2. **자체 완결성**: 토큰 자체에 필요한 정보가 포함되어 있어 추가 DB 조회가 줄어듭니다.
3. **크로스 도메인**: 다른 도메인 간에도 인증 정보를 쉽게 공유할 수 있습니다.

## 구현 방법

### 1. go-board를 인증 서버로 활용하는 방식

```
[사용자] → 로그인 → [go-board] → JWT 발급 → [PHP 시스템에서 검증]
```

1. **go-board의 인증 API 확장**:
   ```go
   // API 엔드포인트 추가 (cmd/main.go의 setupRoutes 함수에 추가)
   api := app.Group("/api")
   api.Post("/login", authHandler.APILogin) // 새 함수 추가
   ```

2. **새 로그인 핸들러 구현**:
   ```go
   // internal/handlers/auth_handler.go에 추가
   func (h *AuthHandler) APILogin(c *fiber.Ctx) error {
       var req struct {
           Username string `json:"username"`
           Password string `json:"password"`
       }
       
       if err := c.BodyParser(&req); err != nil {
           return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
               "success": false,
               "message": "잘못된 요청 형식",
           })
       }
       
       user, token, err := h.authService.Login(c.Context(), req.Username, req.Password)
       if err != nil {
           return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
               "success": false,
               "message": "로그인 실패",
           })
       }
       
       return c.JSON(fiber.Map{
           "success": true,
           "token": token,
           "user": map[string]interface{}{
               "id": user.ID,
               "username": user.Username,
               "email": user.Email,
               "role": user.Role,
           },
       })
   }
   ```

3. **PHP 시스템에서 JWT 검증**:
   ```php
   <?php
   require 'vendor/autoload.php';
   use \Firebase\JWT\JWT;
   
   function verifyToken($token) {
       $key = "your_jwt_secret"; // go-board의 JWT_SECRET과 동일해야 함
       
       try {
           $decoded = JWT::decode($token, $key, array('HS256'));
           return $decoded;
       } catch (Exception $e) {
           return false;
       }
   }
   
   // 요청 헤더에서 토큰 가져오기
   $headers = getallheaders();
   $token = $headers['Authorization'] ?? '';
   $token = str_replace('Bearer ', '', $token);
   
   if ($user = verifyToken($token)) {
       // 인증 성공, $user 객체에 사용자 정보가 있음
       echo "환영합니다, " . $user->username;
   } else {
       // 인증 실패
       header('HTTP/1.0 401 Unauthorized');
       echo "인증 필요";
   }
   ?>
   ```

### 2. 공유 JWT 시크릿 설정

JWT를 사용하려면 두 시스템이 같은 시크릿 키를 사용해야 합니다:

1. go-board의 `.env` 파일에 설정된 `JWT_SECRET` 값을 PHP 시스템에서도 사용합니다.
2. 보안을 위해 이 시크릿은 충분히 복잡하고 강력해야 합니다.

### 3. 로그인 흐름 통합

1. **단일 로그인 페이지**: PHP 시스템에서 로그인하면 go-board API를 호출해 JWT 받기
2. **토큰 저장**: 브라우저 쿠키나 localStorage에 JWT 저장
3. **API 요청 시**: 저장된 토큰을 Authorization 헤더에 포함시켜 전송

### 4. 추가 보안 고려사항

1. **토큰 만료 시간**: JWT에 적절한 만료 시간 설정 (go-board는 현재 24시간으로 설정됨)
2. **HTTPS 사용**: 모든 통신은 HTTPS로 암호화
3. **CORS 설정**: API 호출이 가능하도록 go-board의 CORS 설정 확인

이 방식은 기존 go-board 코드에 최소한의 변경만으로 구현할 수 있으며, PHP 시스템에서도 JWT 라이브러리만 추가하면 쉽게 연동할 수 있습니다.

## Prompt:
저렇게 하면 go-board에 user 테이블에 사용자 정보가 없어도 되는거야?

## Response:

아, 이 질문의 핵심은 JWT 인증 방식을 사용할 때 go-board의 사용자 정보 저장 여부에 관한 것입니다. 분석해보겠습니다.

현재 go-board 시스템은:
1. 자체 데이터베이스의 users 테이블에 사용자 정보를 저장함
2. 로그인 시 사용자 이름과 비밀번호를 확인하고 JWT 토큰을 발급함
3. 이 토큰에는 사용자 ID, 사용자 이름, 역할 등의 정보가 포함됨

만약 JWT를 사용하여 외부 PHP 시스템과 연동한다면:

### 1. 외부 시스템(PHP)에만 사용자 정보 저장
이 경우 go-board에는 사용자 테이블이 없거나 비어있고, 인증 과정은 모두 외부 PHP 시스템에 의존하게 됩니다. 이 방식은:

- PHP 시스템에서 인증 후 JWT 토큰을 발급하고, go-board는 이 토큰만 검증
- go-board는 users 테이블에 접근할 필요가 없음
- 하지만 이 경우 go-board의 내부 로직(게시글 작성자 표시, 권한 확인 등)이 사용자 정보에 의존한다면 문제가 생길 수 있음

### 2. 두 시스템 모두 사용자 정보 저장 (데이터 동기화)
이 경우 PHP 시스템과 go-board 둘 다 사용자 정보를 가지고 있지만 동기화가 필요합니다:

- 사용자 등록, 정보 변경 등이 발생할 때마다 데이터 동기화 필요
- 동기화 메커니즘(DB 복제, API 호출 등)이 필요
- 데이터 일관성 유지가 어려울 수 있음

### 3. 하이브리드 접근 (제한된 사용자 정보 저장)
go-board가 필수 사용자 정보만 저장하는 방식:

- 처음 외부 토큰 검증 시 필요한 최소한의 사용자 정보만 go-board DB에 저장
- 이후 요청에서는 저장된 정보 활용
- 정기적으로 데이터 동기화 또는 토큰 갱신 시 정보 업데이트

현재 go-board의 코드를 보면, 사용자 정보와 연결된 많은 기능(게시물 작성, 댓글, 권한 제어 등)이 있어 사용자 테이블이 완전히 없으면 기능이 제대로 작동하지 않을 가능성이 높습니다.

정확한 지적입니다. 제가 제안한 방식에서는 여전히 go-board의 users 테이블에 사용자 정보가 필요합니다. 기존 go-board는 사용자 정보를 내부 DB에 저장하고, 이를 기반으로 JWT를 발급하는 구조입니다.

외부 PHP 시스템과 완전히 통합하려면 몇 가지 추가 작업이 필요합니다:

## 사용자 정보 없이 go-board 사용하는 방법

### 1. 사용자 동기화 방식

```
[PHP 시스템] → 사용자 생성/변경 → [go-board API] → users 테이블에 동기화
```

1. **사용자 동기화 API 추가**:
   ```go
   // API 엔드포인트 추가
   api.Post("/sync-user", authHandler.SyncUser) // 새 함수 추가
   ```

2. **동기화 핸들러 구현**:
   ```go
   func (h *AuthHandler) SyncUser(c *fiber.Ctx) error {
       // 보안 키 검증 (PHP 시스템에서 알고 있는 비밀 키)
       apiKey := c.Get("X-API-Key")
       if apiKey != os.Getenv("API_SYNC_KEY") {
           return c.Status(fiber.StatusUnauthorized).JSON(...)
       }
       
       var userData struct {
           ID       int64       `json:"id"`
           Username string      `json:"username"`
           Email    string      `json:"email"`
           Role     models.Role `json:"role"`
           // 필요한 다른 필드들
       }
       
       if err := c.BodyParser(&userData); err != nil {
           return c.Status(fiber.StatusBadRequest).JSON(...)
       }
       
       // 사용자 존재 여부 확인 후 업데이트 또는 생성
       user, err := h.authService.GetUserByUsername(c.Context(), userData.Username)
       if err != nil {
           // 사용자 생성 (비밀번호는 더미값, 실제 인증은 PHP에서 처리)
           dummyPassword := utils.GenerateRandomString(32)
           user, err = h.authService.Register(c.Context(), userData.Username, userData.Email, dummyPassword, "")
           if err != nil {
               return c.Status(fiber.StatusInternalServerError).JSON(...)
           }
       }
       
       // 사용자 정보 업데이트
       user.Role = userData.Role
       user.Email = userData.Email
       // 필요한 다른 필드 업데이트
       
       if err := h.authService.UpdateUser(c.Context(), user); err != nil {
           return c.Status(fiber.StatusInternalServerError).JSON(...)
       }
       
       return c.JSON(fiber.Map{"success": true, "id": user.ID})
   }
   ```

### 2. 외부 인증 방식 (JWT 검증만)

```
[PHP 시스템] → JWT 생성 → [go-board] → JWT 검증만 하고 DB 조회 안함
```

이 방식을 위해서는 `internal/service/auth_service.go`의 `ValidateToken` 메서드를 수정해야 합니다:

```go
func (s *authService) ValidateToken(ctx context.Context, tokenString string) (*models.User, error) {
    // 토큰 파싱
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        // ... 현재 코드 유지 ...
    })

    // 토큰 클레임 검증
    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        userID := int64(claims["user_id"].(float64))
        username := claims["username"].(string)
        role := models.Role(claims["role"].(string))

        // 데이터베이스 조회 생략하고 토큰 정보만으로 User 객체 생성
        user := &models.User{
            ID:       userID,
            Username: username,
            Role:     role,
            Active:   true,
        }

        return user, nil
    }

    return nil, errors.New("유효하지 않은 토큰")
}
```

### 3. 실제 적용 시 고려사항

1. **데이터 일관성**: 
   - PHP 시스템에서 사용자 정보 변경 시 go-board에도 동기화 필요
   - 사용자 삭제, 권한 변경 등의 이벤트도 처리해야 함

2. **보안 고려사항**:
   - 사용자 동기화 API는 적절한 인증 필요
   - JWT 시크릿 키 노출 방지

3. **코드 수정 범위**:
   - go-board의 많은 부분이 users 테이블에 의존하므로 대규모 수정 필요
   - 외부 인증만 사용할 경우 기존 회원가입/로그인 기능 비활성화

이 두 가지 접근법 중에서는 **사용자 동기화 방식**이 go-board의 기존 기능을 더 적은 변경으로 유지할 수 있습니다. PHP 시스템에서 사용자 정보가 변경될 때마다 go-board API를 호출하여 사용자 정보를 동기화하는 방식이 실용적일 것 같습니다.

## Prompt:
앞으로 나도 php 시스템이라고 부를게.
php 시스템에서 동기화를 하는건 전체 사용자정보 동기화를 뜻하는거지?

## Response:

네, "PHP 시스템"으로 부르는 것이 좋겠습니다. 그리고 질문하신 내용은 동기화의 범위에 관한 것인데, 제가 설명한 것이 전체 사용자 동기화만을 의미하는지 묻고 계신 것 같습니다.

제가 설명한 사용자 동기화 방식은 다양한 방법으로 구현할 수 있습니다:

1. **전체 사용자 동기화**: PHP 시스템의 모든 사용자 정보를 go-board로 한 번에 동기화하는 방식
   - 장점: 간단하게 일괄 처리 가능
   - 단점: 사용자가 많을 경우 부하가 크고, 모든 사용자가 게시판을 사용하지 않을 수 있음

2. **개별 사용자 동기화**: 사용자가 필요할 때마다(예: 로그인시) 또는 변경이 있을 때(예: 정보 수정, 계정 생성)마다 동기화
   - 장점: 필요한 사용자만 동기화하여 효율적
   - 단점: 동기화 이벤트 관리 필요

3. **점진적 동기화**: 처음 접속한 사용자를 go-board에 등록하고, 이후 변경 사항만 업데이트
   - 장점: 필요한 사용자만 go-board 데이터베이스에 저장되어 공간 효율적
   - 단점: 구현이 조금 더 복잡함

제 설명에서는 주로 개별 사용자 동기화 방식(옵션 2)을 염두에 두고 있었습니다. 이 방식이 대부분의 경우 가장 효율적이고 실용적입니다.

네, PHP 시스템이라고 부르겠습니다.

제가 설명한 동기화 방식은 반드시 전체 사용자 동기화를 의미하는 것은 아닙니다. 사실 동기화 방식은 크게 세 가지로 구현할 수 있습니다:

1. **전체 사용자 동기화**:
   - PHP 시스템의 모든 사용자 데이터를 주기적으로 go-board로 동기화
   - 배치 작업으로 처리 (예: 매일 밤 또는 주 단위)
   - 장점: 간단한 구현
   - 단점: 사용자가 많을 경우 부하가 크고 비효율적

2. **이벤트 기반 동기화** (개별 사용자):
   - 사용자 생성, 수정, 삭제 등의 이벤트가 발생할 때만 동기화
   - PHP 시스템에서 사용자 관련 액션이 있을 때 go-board API 호출
   - 장점: 실시간 데이터 일관성, 필요한 작업만 수행
   - 단점: 이벤트 처리 로직 구현 필요

3. **지연 동기화** (Just-in-time):
   - 사용자가 go-board를 처음 사용할 때만 데이터 동기화
   - 이후 필요시에만 업데이트
   - 장점: 게시판을 사용하는 사용자만 동기화하여 효율적
   - 단점: 동기화 타이밍 관리가 복잡할 수 있음

실무에서는 두 번째 방식인 **이벤트 기반 동기화**가 가장 효과적인 경우가 많습니다. PHP 시스템에서 다음과 같은 이벤트가 발생할 때 go-board API를 호출하면 됩니다:

- 회원가입
- 회원정보 수정
- 계정 비활성화/활성화
- 권한 변경
- 계정 삭제

이렇게 하면 불필요한 동기화를 최소화하면서도 데이터 일관성을 유지할 수 있습니다.

## Prompt:
그럼 php 시스템에서 회원가입이나 탈퇴가 일어날때 go-board에서도 api로 이걸 받아서 처리하게 하면 되는거야?

## Response:

예, 정확합니다. PHP 시스템에서 사용자 관련 이벤트(회원가입, 탈퇴, 정보 수정 등)가 발생할 때마다 go-board API를 호출하여 사용자 정보를 동기화하는 방식을 구현하면 됩니다.

이벤트 기반 동기화의 기본적인 구현 흐름은 다음과 같습니다:

1. PHP 시스템에서 사용자 이벤트 발생:
   - 회원가입 완료
   - 회원 정보 수정
   - 회원 탈퇴/삭제
   - 계정 상태 변경(활성화/비활성화)

2. PHP 시스템에서 go-board API 호출:
   - 해당 이벤트에 맞는 API 엔드포인트 호출
   - 보안 키나 토큰으로 API 요청 인증
   - 필요한 사용자 정보 전송

3. go-board API에서 처리:
   - 요청 인증 검증
   - 사용자 정보 저장/수정/삭제
   - 결과 응답

go-board 측에서는 API 엔드포인트를 구현해야 합니다. 주요 엔드포인트는:

1. 사용자 생성/수정 API: `/api/users/sync`
2. 사용자 삭제 API: `/api/users/delete`
3. 사용자 상태 변경 API: `/api/users/status`

그리고 PHP 시스템에서는 적절한 시점에 이 API를 호출하는 코드를 추가하면 됩니다.

또한, 양쪽 시스템의 사용자 ID 관리 방식이 다를 수 있으므로, 매핑 테이블을 두거나 외부 ID 필드를 추가하여 연동할 수도 있습니다. 이렇게 하면 PHP 시스템의 사용자 ID와 go-board의 사용자 ID를 매핑하여 관리할 수 있습니다.

정확합니다! 그런 방식이 가장 효과적입니다. PHP 시스템에서 사용자 관련 이벤트가 발생할 때마다 go-board API를 호출하여 동기화하는 방식을 구현하면 됩니다.

### 구현 개요

1. **go-board에 API 엔드포인트 추가**:
   - 사용자 생성/업데이트 API
   - 사용자 삭제 API
   - 사용자 상태 변경 API

2. **PHP 시스템에서 이벤트 발생 시 API 호출**:
   - 회원가입 완료 후
   - 회원정보 수정 후
   - 회원 탈퇴 처리 후
   - 계정 상태 변경 후 (활성화/비활성화)

### go-board에 필요한 API 예시

```go
// cmd/main.go의 setupRoutes 함수에 추가
api := app.Group("/api")
userSync := api.Group("/users", apiKeyMiddleware) // API 키로 보호
userSync.Post("/sync", authHandler.SyncUser)      // 생성/업데이트
userSync.Delete("/delete/:id", authHandler.DeleteUser)
userSync.Put("/status/:id", authHandler.UpdateUserStatus)

// API 키 미들웨어 구현
func apiKeyMiddleware(c *fiber.Ctx) error {
    apiKey := c.Get("X-API-Key")
    if apiKey != os.Getenv("API_SYNC_KEY") {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "success": false,
            "message": "인증되지 않은 요청입니다",
        })
    }
    return c.Next()
}
```

### PHP 시스템에서의 호출 예시

```php
<?php
// 회원가입 처리 후 go-board에 동기화
function syncUserToBoard($userId, $username, $email, $role = 'user') {
    $apiKey = getenv('GO_BOARD_API_KEY');
    $apiUrl = getenv('GO_BOARD_API_URL') . '/api/users/sync';
    
    $userData = [
        'external_id' => $userId,  // PHP 시스템의 사용자 ID
        'username' => $username,
        'email' => $email,
        'role' => $role,
        'active' => true
    ];
    
    $ch = curl_init($apiUrl);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_POST, true);
    curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($userData));
    curl_setopt($ch, CURLOPT_HTTPHEADER, [
        'Content-Type: application/json',
        'X-API-Key: ' . $apiKey
    ]);
    
    $response = curl_exec($ch);
    $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    curl_close($ch);
    
    if ($httpCode >= 200 && $httpCode < 300) {
        // 성공
        return json_decode($response, true);
    } else {
        // 실패 처리
        logError('Failed to sync user to board: ' . $response);
        return false;
    }
}

// 회원 탈퇴 시 호출
function deleteUserFromBoard($userId) {
    $apiKey = getenv('GO_BOARD_API_KEY');
    $apiUrl = getenv('GO_BOARD_API_URL') . '/api/users/delete/' . $userId;
    
    // DELETE 요청 보내기
    // ... 위와 유사한 curl 코드 ...
}
```

### 추가 고려사항

1. **아이디 매핑**:
   - PHP 시스템과 go-board의 사용자 ID가 다를 수 있으므로, go-board 사용자 테이블에 `external_id` 필드를 추가하여 PHP 시스템의 ID를 저장하는 것이 좋습니다.

2. **오류 처리**:
   - API 호출 실패 시 재시도 메커니즘 구현
   - 로깅 및 알림 시스템 구축

3. **보안**:
   - 강력한 API 키 사용
   - HTTPS 통신 필수
   - IP 제한 고려

이 방식을 구현하면 PHP 시스템의 사용자 관리 기능을 그대로 사용하면서도 go-board와 완벽하게 통합할 수 있습니다.



---
Powered by [Claude Exporter](https://www.claudexporter.com)