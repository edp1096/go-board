## 사용법

사용 방법은 다음과 같습니다:

1. `xemigration` 디렉토리를 생성하고 위 파일들을 저장합니다.
2. Go 모듈 초기화 및 의존성 설치:
   ```bash
   go mod init github.com/edp1096/go-board/xemigration
   go mod tidy
   ```
3. 다음 명령어로 마이그레이션을 실행합니다:
   ```bash
   ./xemigration --source-driver=mysql --source-host=localhost --source-port=3306 --source-db=xedb --source-user=user_pass --source-pass=user_pass --prefix=xe_ --upload-path=./files --target-upload-path=./uploads --env=.env --batch-size=500 --migrate-files=true --verbose=true
   ```
