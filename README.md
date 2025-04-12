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
