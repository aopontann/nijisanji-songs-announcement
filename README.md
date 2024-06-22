# にじ通
- にじさんじの動画公開時などに通知をするWEBアプリ

#### 使用技術
- Go
- FCM
- PostgreSQL
- Bun
- atlas
- Astro
- Bulma
- ko

##### マイグレーション
```
atlas schema apply \
  --url "postgres://postgres:example@/test_db?&sslmode=disable" \
  --to "file://schema.sql" \
  --dev-url "docker://postgres/16"
```

##### デプロイ
バッチ
```
ko build ./cmd/batch
```
WEBページ
```
ko build ./frontend
```

### メモ
'''
export GOOGLE_APPLICATION_CREDENTIALS="token.json"
export KO_DOCKER_REPO=asia-northeast1-docker.pkg.dev/niji-tuu/buildpacks
gcloud auth configure-docker asia-northeast1-docker.pkg.dev
'''

### DBコンテナ
'''
podman run --name some-postgres -e POSTGRES_PASSWORD=mysecretpassword -p 5432:5432 -d docker.io/library/postgres:16
podman start some-postgres
podman stop some-postgres
'''
