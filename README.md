간단한 회원관리와 게시판

예시: https://bbs.enjoytools.net


## 처음 실행

```sh
go-board export-env
mv .env.example .env
migrate
go-board
```


## 실행

```sh
go-board
```


## 설정

* `.env` 파일 수정


## 템플릿

* 아래 명령 실행 후, `web` 폴더 수정
```sh
go-board export-web
```


## 도움말

```sh
go-board -h
migrate -h
```


----

ai 써서 만들었습니다.

`클로드 프로`를 사용했고, 참고용으로 대화내역을 `claude`폴더에 올려두었습니다.

`goose` 마이그레이션 모듈은 사전지식 없이 ai의 제안을 그대로 반영했습니다.

ai에게 갑질은 최대한 지양하려 노력했지만 보면서 불편을 끼치는 어구가 있다면 양해 부탁드립니다.

누군가에게는 도움이 되길 바랍니다.
