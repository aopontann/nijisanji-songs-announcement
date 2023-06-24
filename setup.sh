# APIの有効化
gcloud services enable artifactregistry.googleapis.com cloudbuild.googleapis.com secretmanager.googleapis.com run.googleapis.com youtube.googleapis.com cloudscheduler.googleapis.com

# -------------------- secret manager --------------------

# サービスアカウント作成
gcloud iam service-accounts create cloud-run-jobs \
    --description="Cloud Run Jobs を起動し、Secret Manager にアクセスする" \
    --display-name="cloud-run-jobs"

# シークレットを作成
gcloud secrets create DSN \
    --replication-policy="automatic"

gcloud secrets create YOUTUBE_API_KEY \
    --replication-policy="automatic"

gcloud secrets create SMTP_PASSWORD \
    --replication-policy="automatic"

gcloud secrets create TWITTER_API_KEY \
    --replication-policy="automatic"

gcloud secrets create TWITTER_API_SECRET_KEY \
    --replication-policy="automatic"

gcloud secrets create TWITTER_ACCESS_TOKEN \
    --replication-policy="automatic"

gcloud secrets create TWITTER_ACCESS_TOKEN_SECRET \
    --replication-policy="automatic"

gcloud secrets create MISSKEY_TOKEN \
    --replication-policy="automatic"

# シークレットへのアクセスを許可
gcloud secrets add-iam-policy-binding DSN \
    --member="serviceAccount:cloud-run-jobs@${PROJECT_ID}.iam.gserviceaccount.com" \
    --role="roles/secretmanager.secretAccessor"

gcloud secrets add-iam-policy-binding YOUTUBE_API_KEY \
    --member="serviceAccount:cloud-run-jobs@${PROJECT_ID}.iam.gserviceaccount.com" \
    --role="roles/secretmanager.secretAccessor"

gcloud secrets add-iam-policy-binding SMTP_PASSWORD \
    --member="serviceAccount:cloud-run-jobs@${PROJECT_ID}.iam.gserviceaccount.com" \
    --role="roles/secretmanager.secretAccessor"

gcloud secrets add-iam-policy-binding TWITTER_API_KEY \
    --member="serviceAccount:cloud-run-jobs@${PROJECT_ID}.iam.gserviceaccount.com" \
    --role="roles/secretmanager.secretAccessor"

gcloud secrets add-iam-policy-binding TWITTER_API_SECRET_KEY \
    --member="serviceAccount:cloud-run-jobs@${PROJECT_ID}.iam.gserviceaccount.com" \
    --role="roles/secretmanager.secretAccessor"

gcloud secrets add-iam-policy-binding TWITTER_ACCESS_TOKEN \
    --member="serviceAccount:cloud-run-jobs@${PROJECT_ID}.iam.gserviceaccount.com" \
    --role="roles/secretmanager.secretAccessor"

gcloud secrets add-iam-policy-binding TWITTER_ACCESS_TOKEN_SECRET \
    --member="serviceAccount:cloud-run-jobs@${PROJECT_ID}.iam.gserviceaccount.com" \
    --role="roles/secretmanager.secretAccessor"

gcloud secrets add-iam-policy-binding MISSKEY_TOKEN \
    --member="serviceAccount:cloud-run-jobs@${PROJECT_ID}.iam.gserviceaccount.com" \
    --role="roles/secretmanager.secretAccessor"

# --------------------------------------------------------

# -------------------- cloud build --------------------
# Artifacts Registry のリポジトリを作成
gcloud artifacts repositories create buildpacks-docker-repo \
    --repository-format="docker" \
    --location="asia-northeast1" \
    --description="Docker repository"

gcloud iam service-accounts create github-actions-builder \
    --description="GitHub Actions でアプリケーションのビルドをする" \
    --display-name="github-actions-builder"

# Workload Identityプールを作成
gcloud iam workload-identity-pools create "build-pool" \
  --location="global" \
  --display-name="GitHub Actions で Workload Identity 連携を使う"

# Workload Identityプールの中にWorkload Identityプロバイダを作成
gcloud iam workload-identity-pools providers create-oidc "my-provider" \
  --location="global" \
  --workload-identity-pool="build-pool" \
  --display-name="Demo provider" \
  --attribute-mapping="google.subject=assertion.repository" \
  --attribute_condition="assertion.repository_owner == ${local.github_repo_owner}" \
  --issuer-uri="https://token.actions.githubusercontent.com"

gcloud iam service-accounts add-iam-policy-binding "${SERVICE_ACCOUNT_NAME}@${PROJECT_ID}.iam.gserviceaccount.com" \
  --project="${PROJECT_ID}" \
  --role="roles/iam.workloadIdentityUser" \
  --member="principalSet://iam.googleapis.com/${WORKLOAD_IDENTITY_POOL_ID}/attribute.repository/${GITHUB_REPO}"