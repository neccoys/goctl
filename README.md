# goctl使用方式

### 安裝
```go
go install github.com/neccoys/goctl@latest

goctl -v
// goctl version 1.3.2 darwin/amd64
```

### API 初次使用
```go

goctl api go -api {$name}.api -dir {$name} -remote https://github.com/neccohuang/go-zero-template -common ./


go mod tidy


// 啟動
make run

// 新建api
make api // make api t={$serviceName}

// i18n
make lang

```

### RPC
```go
goctl rpc proto -src {$name}.proto -dir . -consul ttl --remote https://github.com/neccohuang/go-zero-template

go mod tidy

```



