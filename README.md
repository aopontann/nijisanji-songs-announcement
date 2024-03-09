# nijisanji-songs-announcement
- どんなことに挑戦したか（アプリケーションの特徴）
- アプリケーションのURL
- どんなアプリケーションかわかるように、gifや画像を貼る
- 使っている技術
- 参考にしたサイト・動画など

### コマンド
##### マイグレーション
```
atlas migrate diff create_vtubers_videos_table \
  --dir "file://db/migrations" \
  --to "file://db/schema.sql" \
  --dev-url "docker://mysql/8/nsa"

atlas schema apply \
  --url "mysql://root:password@localhost:3306/nsa" \
  --to "file://db/migrations" \
  --dev-url "docker://mysql/8/example"
```
##### デプロイ
```
gcloud builds submit --pack image=asia-northeast1-docker.pkg.dev/${PROJECT_ID}/buildpacks-docker-repo/nsa-bot,env=GOOGLE_BUILDABLE="cmd/api/main.go"
```

export GOOGLE_APPLICATION_CREDENTIALS="token.json"