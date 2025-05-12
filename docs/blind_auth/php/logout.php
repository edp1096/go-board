<?php
session_start();

// 로그인된 사용자 정보 저장
$userId = $_SESSION['user_id'] ?? '';
$boardUrl = getenv('BOARD_URL') ?: 'http://localhost:3000';

// 현재 서버에서의 세션 정리
$_SESSION = [];
if (ini_get("session.use_cookies")) {
    $params = session_get_cookie_params();
    setcookie(session_name(), '', time() - 42000, $params["path"], $params["domain"], $params["secure"], $params["httponly"]);
}
session_destroy();

// 로그인 페이지 리다이렉트 URL
$redirect_url = 'login.php';
?>
<!DOCTYPE html>
<html>

<head>
    <title>로그아웃</title>
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

        .message-container {
            background-color: white;
            padding: 30px;
            border-radius: 5px;
            box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
            width: 350px;
            text-align: center;
        }

        h2 {
            color: #333;
            margin-bottom: 20px;
        }

        p {
            margin-bottom: 20px;
        }

        a {
            display: inline-block;
            padding: 10px 20px;
            background-color: #4CAF50;
            color: white;
            text-decoration: none;
            border-radius: 3px;
        }

        a:hover {
            background-color: #45a049;
        }

        #loading {
            display: inline-block;
            margin-right: 10px;
        }
    </style>
</head>

<body>
    <div class="message-container">
        <h2>로그아웃 처리 중...</h2>
        <p id="status">모든 시스템에서 로그아웃 중입니다...</p>
        <div id="loading">⏳</div>
        <a href="<?php echo $redirect_url; ?>" style="display: none;" id="login-link">로그인 페이지로 돌아가기</a>
    </div>

    <script>
        // 게시판 로그아웃
        function logoutFromBoard() {
            // iframe으로 게시판 로그아웃 페이지 접근
            const iframe = document.createElement('iframe');
            iframe.style.width = '0';
            iframe.style.height = '0';
            iframe.style.border = '0';
            iframe.style.display = 'none';
            iframe.src = "<?php echo $boardUrl; ?>/auth/logout";

            // iframe 로드 완료 또는 500ms 후 완료 처리
            iframe.onload = finishLogout;
            setTimeout(finishLogout, 500); // 백업 타이머

            document.body.appendChild(iframe);
        }

        function finishLogout() {
            document.getElementById('status').textContent = '성공적으로 로그아웃되었습니다.';
            document.getElementById('loading').style.display = 'none';
            document.getElementById('login-link').style.display = 'inline-block';
            document.querySelector('h2').textContent = '로그아웃 완료';

            // 3초 후 로그인 페이지로 리다이렉트
            setTimeout(function() {
                window.location.href = "<?php echo $redirect_url; ?>";
            }, 3000);
        }

        // 페이지 로드 시 로그아웃 처리 시작
        window.addEventListener('DOMContentLoaded', logoutFromBoard);
    </script>
</body>

</html>