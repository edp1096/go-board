XpressEngine 1.x (1.8.3 ~ 1.11.6) 버전에서 마이그레이션하는 방법입니다.

기본적인 요소인 게시판과 댓글과 첨부파일만 변환되며, 게시물 내 이미지 등은 직접 수정해야합니다.

첨부파일의 폴더구조도 같이 변경됩니다.

연관키 문제가 있으니 가능하면 sqlite로 변환 후, mysql 또는 postgresql로 변환하는게 좋습니다.

이기종 DB간 변환은 MIGRATION.md를 참고하세요.


## 사용법

1. Go 모듈 초기화 및 의존성 설치:
   ```bash
   go mod init github.com/edp1096/go-board/xe_convert
   go mod tidy
   go build cmd/xe_convert
   ```
2. 다음 명령어로 마이그레이션을 실행합니다:
   ```bash
   ./migrate -op purge
   ./xe_convert --source-driver=mysql --source-host=localhost --source-port=3306 --source-db=xedb --source-user=user_pass --source-pass=user_pass --prefix=xe_ --upload-path=./files --target-upload-path=./uploads --env=.env --batch-size=500 --migrate-files=true --verbose=true
   ```
