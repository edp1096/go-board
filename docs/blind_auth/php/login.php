<?php
session_start();

// 사용자 정보 (실제로는 DB에서 가져와야 함)
$users = [
    'user1' => [
        'password' => 'user123',
        'user_id' => "EXT00001",
        'email' => 'user1@example.com',
        'full_name' => '사용자1'
    ]
];

$boardURL = getenv('BOARD_URL') ?: 'http://localhost:3000';

// 이미 로그인되어 있으면 바로 게시판으로 리다이렉트
if (isset($_SESSION['user_id'])) {
    redirectToBoard($boardURL);
}

// 로그인 처리
$error = '';
if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $username = $_POST['username'] ?? '';
    $password = $_POST['password'] ?? '';

    // 사용자 검증
    if (isset($users[$username]) && $users[$username]['password'] === $password) {
        // 세션에 사용자 정보 저장
        $_SESSION['user_id'] = $users[$username]['user_id'];
        $_SESSION['username'] = $username;
        $_SESSION['email'] = $users[$username]['email'];
        $_SESSION['full_name'] = $users[$username]['full_name'];

        // 게시판으로 리다이렉트
        redirectToBoard($boardURL);
    } else {
        $error = '아이디 또는 비밀번호가 올바르지 않습니다.';
    }
}

// 상위 시스템에서 게시판으로 이동하는 함수
function redirectToBoard($boardURL) {
    // 임시 토큰 생성
    $payload = [
        'external_id' => (string)$_SESSION['user_id'],
        'username' => $_SESSION['username'],
        'email' => $_SESSION['email'] ?? '',
        'full_name' => $_SESSION['full_name'] ?? $_SESSION['username'],
        'system' => 'main_system',
        'iat' => time(),
        'exp' => time() + 60 // 1분 유효 (매우 짧게)
    ];

    $token = generateToken($payload, getenv('EXTERNAL_AUTH_SECRET') ?: 'your_shared_secret_key');

    // 게시판 리다이렉트 URL
    $redirectUrl = $boardURL . '/blind-auth?token=' . urlencode($token);

    // 헤더 리다이렉트 대신 HTML로 로딩 화면 표시 후 js로 리다이렉트
    include 'loading_template.php';
    exit;
}

// JWT 토큰 생성
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

?>

<!DOCTYPE html>
<html>

<head>
    <title>로그인</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: Arial, sans-serif;
            background-color: #f5f5f5;
            margin: 0;
            padding: 0;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
        }

        .login-container {
            background-color: white;
            padding: 30px;
            border-radius: 5px;
            box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
            width: 350px;
        }

        h2 {
            text-align: center;
            margin-bottom: 20px;
            color: #333;
        }

        .form-group {
            margin-bottom: 15px;
        }

        label {
            display: block;
            margin-bottom: 5px;
            font-weight: bold;
        }

        input[type="text"],
        input[type="password"] {
            width: 100%;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 3px;
            box-sizing: border-box;
        }

        button {
            width: 100%;
            padding: 10px;
            background-color: #4CAF50;
            color: white;
            border: none;
            border-radius: 3px;
            cursor: pointer;
            font-size: 16px;
        }

        button:hover {
            background-color: #45a049;
        }

        .error {
            color: red;
            margin-bottom: 15px;
            text-align: center;
        }
    </style>
</head>

<body>
    <div class="login-container">
        <h2>로그인</h2>
        <?php if ($error): ?>
            <div class="error"><?php echo $error; ?></div>
        <?php endif; ?>
        <form method="post" action="">
            <div class="form-group">
                <label for="username">아이디</label>
                <input type="text" id="username" name="username" required>
            </div>
            <div class="form-group">
                <label for="password">비밀번호</label>
                <input type="password" id="password" name="password" required>
            </div>
            <button type="submit">로그인</button>
        </form>
    </div>
</body>

</html>