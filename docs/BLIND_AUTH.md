# 외부시스템 계정 연동

예시 php.

## 상위 시스템에서 게시판으로 이동 시 자동 인증
```php
<?php
// 상위 시스템에서 게시판으로 이동하는 라우트
function redirectToBoard() {
    session_start();

    // 로그인 확인
    if (!isset($_SESSION['user_id'])) {
        // 로그인 페이지로 리다이렉트
        header('Location: /login?redirect=/board');
        exit;
    }

    // 사용자 정보
    $user_id = $_SESSION['user_id'];
    $username = $_SESSION['username'];
    $email = $_SESSION['email'] ?? '';
    $fullName = $_SESSION['full_name'] ?? $username;

    // 임시 토큰 생성
    $payload = [
        'external_id' => (string)$user_id,
        'username' => $username,
        'email' => $email,
        'full_name' => $fullName,
        'system' => 'main_system',
        'iat' => time(),
        'exp' => time() + 60 // 1분 유효 (매우 짧게)
    ];

    $token = generateToken($payload, getenv('EXTERNAL_AUTH_SECRET'));

    // 게시판으로 리다이렉트 (로딩 화면 표시)
    $boardUrl = getenv('BOARD_URL') . '/blind-auth?token=' . urlencode($token);
?>
<!DOCTYPE html>
<html>
<head>
    <title>게시판으로 이동 중...</title>
    <style>
        .loading {
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            flex-direction: column;
        }
        .spinner {
            border: 4px solid #f3f3f3;
            border-top: 4px solid #3498db;
            border-radius: 50%;
            width: 30px;
            height: 30px;
            animation: spin 1s linear infinite;
            margin-bottom: 20px;
        }
        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
    </style>
</head>
<body>
    <div class="loading">
        <div class="spinner"></div>
        <p>게시판으로 이동 중입니다...</p>
    </div>

    <script>
        // 페이지 로드 즉시 게시판으로 이동
        window.location.href = "<?php echo $boardUrl; ?>";
    </script>
</body>
</html>
<?php
    exit;
}

// JWT 스타일 토큰 생성 함수
function generateToken($payload, $secret) {
    $header = ['alg' => 'HS256', 'typ' => 'JWT'];

    // Base64Url 인코딩
    $base64UrlHeader = rtrim(strtr(base64_encode(json_encode($header)), '+/', '-_'), '=');
    $base64UrlPayload = rtrim(strtr(base64_encode(json_encode($payload)), '+/', '-_'), '=');

    // 서명 생성
    $signature = hash_hmac('sha256', $base64UrlHeader . '.' . $base64UrlPayload, $secret, true);
    $base64UrlSignature = rtrim(strtr(base64_encode($signature), '+/', '-_'), '=');

    // 토큰 조합
    return $base64UrlHeader . '.' . $base64UrlPayload . '.' . $base64UrlSignature;
}
```

## 상위 시스템에서 로그아웃 시 게시판도 함께 로그아웃
```php
<?php
// 상위 시스템의 로그아웃 처리
function logout() {
    session_start();

    // 로그인된 사용자만 처리
    if (isset($_SESSION['user_id'])) {
        $userId = $_SESSION['user_id'];

        // 세션 파기
        session_destroy();

        // 게시판 로그아웃 요청 (백그라운드)
        logoutFromBoard($userId);
    }

    // 로그인 페이지로 리다이렉트
    header('Location: /login');
    exit;
}

// 게시판 로그아웃 요청
function logoutFromBoard($userId) {
    $apiKey = hash_hmac('sha256', $userId . 'main_system', getenv('EXTERNAL_AUTH_SECRET'));

    // 백그라운드 요청 (사용자를 기다리게 하지 않음)
    $url = getenv('BOARD_URL') . '/api/external-logout';
    $data = [
        'external_id' => $userId,
        'system' => 'main_system',
        'api_key' => $apiKey
    ];

    // cURL 요청 설정
    $ch = curl_init($url);
    curl_setopt($ch, CURLOPT_POST, 1);
    curl_setopt($ch, CURLOPT_POSTFIELDS, http_build_query($data));
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_TIMEOUT, 1); // 1초 타임아웃 (비동기)

    // 요청 실행
    curl_exec($ch);
    curl_close($ch);
}
```

## Nginx 설정 예시
```nginx
# 상위 시스템 - domain1.com
server {
    listen 80;
    server_name domain1.com;
    return 301 https://$host$request_uri; # HTTPS로 리다이렉트
}

server {
    listen 443 ssl;
    server_name domain1.com;

    # SSL 설정
    ssl_certificate /path/to/ssl/domain1.com.crt;
    ssl_certificate_key /path/to/ssl/domain1.com.key;

    # 상위 시스템으로 요청 전달
    location / {
        proxy_pass http://localhost:8080; # 상위 시스템 서버
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# 게시판 - domain2.com
server {
    listen 80;
    server_name domain2.com;
    return 301 https://$host$request_uri; # HTTPS로 리다이렉트
}

server {
    listen 443 ssl;
    server_name domain2.com;

    # SSL 설정
    ssl_certificate /path/to/ssl/domain2.com.crt;
    ssl_certificate_key /path/to/ssl/domain2.com.key;

    # 게시판으로 요청 전달
    location / {
        proxy_pass http://localhost:3000; # go-board 서버
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```
