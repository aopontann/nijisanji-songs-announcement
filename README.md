# nijisanji-songs-announcement
- どんなことに挑戦したか（アプリケーションの特徴）
- アプリケーションのURL
- どんなアプリケーションかわかるように、gifや画像を貼る
- 使っている技術
- 参考にしたサイト・動画など

### デプロイ
```
gcloud functions deploy nsa-bot \
--gen2 \
--region=asia-northeast1 \
--runtime=go121 \
--source=./ \
--entry-point=MyHTTPFunction \
--trigger-http
```