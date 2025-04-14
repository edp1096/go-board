# 리버스프록시 설정 예시

## Nginx

```nginx
server {
    listen 80;
    listen [::]:80;

    server_name bbs.enjoytools.net;

    return 301 https://$host$request_uri;

#    client_max_body_size 50M;

#    location / {
#        proxy_set_header Host $host;
#        proxy_set_header Accept-Encoding "";
#       proxy_set_header Real-Ip $remote_addr;
#        proxy_pass http://127.0.0.1:3000;
#    }

#    error_page 404 /404.html;
#    error_page 500 502 503 504 /50x.html;

}

server {
    listen 443 ssl;
    listen [::]:443 ssl;

    server_name bbs.enjoytools.net;

    client_max_body_size 50M;

    location / {
        proxy_pass http://127.0.0.1:3000;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-Host $host;
    }

    error_page 404 /404.html;
    error_page 500 502 503 504 /50x.html;

    ssl_certificate /etc/letsencrypt/live/my_website/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/my_website/privkey.pem;

    include /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;
}
```