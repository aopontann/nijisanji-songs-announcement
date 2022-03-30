FROM golang:1.17-alpine AS dev
WORKDIR /workspaces/nijisanji-songs-announcement
COPY go.mod go.sum ./
RUN go mod download
RUN go get -u github.com/cosmtrek/air
RUN apk add git
RUN go install -v golang.org/x/tools/gopls@latest
RUN go install -v github.com/ramya-rao-a/go-outline@latest
COPY . ./
EXPOSE 8081

# http://docs.docker.jp/v19.03/develop/develop-images/multistage-build.htmlのコピー
FROM golang:1.17-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest AS prod
# コンテナでSSL接続するためインストール
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=build /app .
CMD ["./app"]
