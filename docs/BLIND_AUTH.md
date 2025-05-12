# 외부시스템 계정 연동

## 상위 시스템에서 게시판으로 로그인 예시

php
* [인증/리디렉션 예시](/docs/blind_auth/php/login.php)
    * [loading_template.php](/docs/blind_auth/php/loading_template.php)
* [로그아웃](/docs/blind_auth/php/logout.php)


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
