<!DOCTYPE html>
<html>

<head>
    <title>게시판으로 이동 중...</title>
    <meta charset="UTF-8">
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
            0% {
                transform: rotate(0deg);
            }

            100% {
                transform: rotate(360deg);
            }
        }
    </style>
</head>

<body>
    <div class="loading">
        <div class="spinner"></div>
        <p>게시판으로 이동 중입니다...</p>
    </div>
    <script>
        // 게시판으로 이동
        window.location.href = "<?php echo $redirectUrl; ?>";
    </script>
</body>

</html>