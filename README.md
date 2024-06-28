# にじ通
- にじさんじの動画公開時などに通知をするWEBアプリ

#### 使用技術
- [Go](https://go.dev/)
- [FCM](https://firebase.google.com/docs/cloud-messaging?hl=ja)
- [PostgreSQL](https://www.postgresql.org/)
- [Bun](https://bun.uptrace.dev/)
- [atlas](https://atlasgo.io/)
- [Astro](https://astro.build/)
- [Bulma](https://bulma.io/)
- [ko](https://github.com/ko-build/ko)

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
podman run --name niji-tuu-postgres -e POSTGRES_PASSWORD=example -p 5432:5432 -d docker.io/library/postgres:16
podman start niji-tuu-postgres
podman stop niji-tuu-postgres
'''
