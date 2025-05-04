XpressEngine 1.x 버전에서 마이그레이션하는 방법입니다.

기본적인 요소인 게시판과 댓글과 첨부파일만 변환되며, 게시물 내 이미지 등은 직접 수정해야합니다.

첨부파일의 폴더구조도 같이 변경됩니다.


## 사용법

1. Go 모듈 초기화 및 의존성 설치:
   ```bash
   go mod init github.com/edp1096/go-board/xe_convert
   go mod tidy
   ```
2. 다음 명령어로 마이그레이션을 실행합니다:
   ```bash
   ./xe_convert --source-driver=mysql --source-host=localhost --source-port=3306 --source-db=xedb --source-user=user_pass --source-pass=user_pass --prefix=xe_ --upload-path=./files --target-upload-path=./uploads --env=.env --batch-size=500 --migrate-files=true --verbose=true
   ```
