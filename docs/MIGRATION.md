## Data migrate between databases

* sqlite to postgres - OK
```sh
.\migrate_windows_amd64.exe -op purge
.\migrate_windows_amd64.exe -op data-migrate -source-driver sqlite -target-driver postgres -source-env .env_sqlite -target-env .env_pg
```

* sqlite to mysql - OK
```sh
.\migrate_windows_amd64.exe -op purge
.\migrate_windows_amd64.exe -op data-migrate -source-driver sqlite -target-driver mysql -source-env .env_sqlite -target-env .env_my
```

* mysql to postgres - OK
```sh
.\migrate_windows_amd64.exe -op purge
.\migrate_windows_amd64.exe -op data-migrate -source-driver mysql -target-driver postgres -source-env .env_my -target-env .env_pg
```

* postgres to mysql - OK
```sh
.\migrate_windows_amd64.exe -op purge
.\migrate_windows_amd64.exe -op data-migrate -source-driver postgres -target-driver mysql -source-env .env_pg -target-env .env_my
```

* postgres to sqlite - OK
```sh
.\migrate_windows_amd64.exe -op purge
.\migrate_windows_amd64.exe -op data-migrate -source-driver postgres -target-driver sqlite -source-env .env_pg -target-env .env_sqlite

```

* mysql to sqlite - OK
```sh
.\migrate_windows_amd64.exe -op purge
.\migrate_windows_amd64.exe -op data-migrate -source-driver mysql -target-driver sqlite -source-env .env_my -target-env .env_sqlite
```
