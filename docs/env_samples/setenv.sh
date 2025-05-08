#!/bin/bash

export APP_ENV=production

export SITE_NAME="Go Board System"
export LOGO_PATH=/static/images/logo.png
export LOGO_DISPLAY_MODE=image

export SERVER_ADDRESS=127.0.0.1:3000

export DB_DRIVER=mysql

export DB_HOST=localhost
export DB_PORT=13306
export DB_USER=root
export DB_PASSWORD=
export DB_NAME=go_board

export DB_PATH=./data/go_board.db

export JWT_SECRET=your-jwt-secret-key
export SESSION_SECRET=your-session-secret
export COOKIE_SECURE=false
export COOKIE_HTTP_ONLY=true

export TEMPLATE_DIR=./web/templates
export STATIC_DIR=./web/static

export UPLOAD_DIR=./uploads
export MAX_UPLOAD_SIZE=20480
export MAX_MEDIA_UPLOAD_SIZE=20480
export MAX_BODY_LIMIT=800
