# syntax=docker/dockerfile:1

FROM golang:1.17-alpine

WORKDIR /workspaces/nijisanji-songs-announcement

# COPY go.mod ./
# COPY go.sum ./

# RUN go mod download

# DBマイグレーションツールのインストール
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Golang ホットリロードするためインストールする
RUN go get -u github.com/cosmtrek/air

# gitをインストール(vscodeの拡張機能「remote - Containers」を使って開発する場合)
RUN apk add git

# Goの拡張機能のインストール
RUN go install -v golang.org/x/tools/gopls@latest
RUN go install -v github.com/ramya-rao-a/go-outline@latest

COPY . ./

# ホットリロード機能が付いた開発環境を起動
# CMD [ "air -c .air.toml" ]

EXPOSE 8000
