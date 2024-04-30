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
'''
export GOOGLE_APPLICATION_CREDENTIALS="token.json"

npx wrangler pages dev ./public

npx wrangler pages deploy ./public

nvm install 20

wrangler d1 execute niji-tuu \
  --local --command "CREATE TABLE IF NOT EXISTS users ( token TEXT PRIMARY KEY, song INTEGER, word TEXT, time TEXT);"
'''