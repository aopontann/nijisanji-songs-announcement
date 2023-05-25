# nijisanji-songs-announcement
- どんなことに挑戦したか（アプリケーションの特徴）
- アプリケーションのURL
- どんなアプリケーションかわかるように、gifや画像を貼る
- 使っている技術
- 参考にしたサイト・動画など

### メモ
```
gcloud auth login

gcloud config set project <PROJECT_NAME>

gcloud builds submit --pack image=gcr.io/<PROJECT_NAME>

gcloud run jobs create nsa-bot --image gcr.io/<PROJECT_NAME>/nsa-bot:latest --tasks 2
```