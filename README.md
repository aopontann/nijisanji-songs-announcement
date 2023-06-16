# nijisanji-songs-announcement
- どんなことに挑戦したか（アプリケーションの特徴）
- アプリケーションのURL
- どんなアプリケーションかわかるように、gifや画像を貼る
- 使っている技術
- 参考にしたサイト・動画など

### メモ
```
export PROJECT_ID="my-project"
export REPO="username/name"

gcloud auth login

gcloud config set project ${PROJECT_ID}

./setup.sh
```

シークレットバージョンを追加
```
echo -n $DSN | \
    gcloud secrets versions add DSN --data-file=-

echo -n $YOUTUBE_API_KEY | \
    gcloud secrets versions add YOUTUBE_API_KEY --data-file=-

echo -n $SMTP_PASSWORD | \
    gcloud secrets versions add SMTP_PASSWORD --data-file=-

echo -n $TWITTER_API_KEY | \
    gcloud secrets versions add TWITTER_API_KEY --data-file=-

echo -n $TWITTER_API_SECRET_KEY | \
    gcloud secrets versions add TWITTER_API_SECRET_KEY --data-file=-

echo -n $TWITTER_ACCESS_TOKEN | \
    gcloud secrets versions add TWITTER_ACCESS_TOKEN --data-file=-

echo -n $TWITTER_ACCESS_TOKEN_SECRET | \
    gcloud secrets versions add TWITTER_ACCESS_TOKEN_SECRET --data-file=-

```

Buildpackを使用してアプリケーションのビルド
```
gcloud builds submit --pack image=asia-northeast1-docker.pkg.dev/${PROJECT_ID}/buildpacks-docker-repo/nsa-bot
```

Cloud Run Jobs サービスの作成
```
gcloud run jobs replace job.yaml
```

Cloud Scheduler に Cloud Run Jobsを起動する権限を与える
```
gcloud run jobs add-iam-policy-binding nsa-bot \
    --region='asia-northeast1' \
    --member='serviceAccount:cloud-run-jobs-scheduler@${PROJECT_ID}.iam.gserviceaccount.com' \
    --role='roles/run.invoker'
```

スケジュールに従って実行するよう
```
gcloud scheduler jobs create http job-scheduler \
  --location asia-northeast1 \
  --time-zone="Asia/Tokyo" \
  --schedule="*/5 * * * *" \
  --uri="https://asia-northeast1-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/${PROJECT_ID}/jobs/nsa-bot:run" \
  --http-method POST \
  --oauth-service-account-email cloud-run-jobs-scheduler@${PROJECT_ID}.iam.gserviceaccount.com
```

### 参考記事
- [Buildpack を使用してアプリケーションをビルドする](https://cloud.google.com/docs/buildpacks/build-application?hl=ja)