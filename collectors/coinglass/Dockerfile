# 🐳 Coinglass 收集器
# 📝 需在 repo 根目录建置: docker build -f collectors/coinglass/Dockerfile -t coinglass-collector .

FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY platforms/coinglass ./platforms/coinglass
COPY collectors/coinglass ./collectors/coinglass
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /collector ./collectors/coinglass

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /collector /collector
USER nobody
ENTRYPOINT ["/collector"]
