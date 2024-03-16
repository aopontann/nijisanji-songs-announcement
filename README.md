# nijisanji-songs-announcement
- どんなことに挑戦したか（アプリケーションの特徴）
- アプリケーションのURL
- どんなアプリケーションかわかるように、gifや画像を貼る
- 使っている技術
- 参考にしたサイト・動画など

### コマンド
##### マイグレーション
```
atlas schema apply \
  --url "postgres://postgres:example@/test_db?&sslmode=disable" \
  --to "file://schema.sql" \
  --dev-url "docker://postgres/16"
```
##### デプロイ
バッチ処理のAPI
```
gcloud builds submit --pack image=asia-northeast1-docker.pkg.dev/${PROJECT_ID}/buildpacks-docker-repo/nsa-bot,env=GOOGLE_BUILDABLE="cmd/api/main.go"
```
公開用API
```
gcloud builds submit --pack image=asia-northeast1-docker.pkg.dev/${PROJECT_ID}/buildpacks-docker-repo/nsa-bot-web,env=GOOGLE_BUILDABLE="cmd/web/main.go"
```

### メモ
export GOOGLE_APPLICATION_CREDENTIALS="token.json"