set GOOS=linux
set GOARCH=amd64
#go build -o proxy-server main.go
# 注入动态信息（结合 Git 和日期）
GIT_COMMIT=$(git rev-parse --short HEAD)
BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)
[ -e agdb ] && rm -f agdb
go build -ldflags "-X main.version=1.0.0 -X main.commit=$GIT_COMMIT -X main.date=$BUILD_TIME" -o agdb .