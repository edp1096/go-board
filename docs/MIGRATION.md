## Data migrate between databases

`-source-env`와 `-target-env`는 `.env`가 아닌 `.env_sqlite`, `.env_pg`, `.env_my` 같은 별도 env 파일을 사용해야 합니다.

* sqlite to postgres - OK
```sh
./migrate -op purge
./migrate -op data-migrate -source-driver sqlite -target-driver postgres -source-env .env_sqlite -target-env .env_pg
```

* sqlite to mysql - OK
```sh
./migrate -op purge
./migrate -op data-migrate -source-driver sqlite -target-driver mysql -source-env .env_sqlite -target-env .env_my
```

* mysql to postgres - OK
```sh
./migrate -op purge
./migrate -op data-migrate -source-driver mysql -target-driver postgres -source-env .env_my -target-env .env_pg
```

* postgres to mysql - OK
```sh
./migrate -op purge
./migrate -op data-migrate -source-driver postgres -target-driver mysql -source-env .env_pg -target-env .env_my
```

* postgres to sqlite - OK
```sh
./migrate -op purge
./migrate -op data-migrate -source-driver postgres -target-driver sqlite -source-env .env_pg -target-env .env_sqlite

```

* mysql to sqlite - OK
```sh
./migrate -op purge
./migrate -op data-migrate -source-driver mysql -target-driver sqlite -source-env .env_my -target-env .env_sqlite
```
