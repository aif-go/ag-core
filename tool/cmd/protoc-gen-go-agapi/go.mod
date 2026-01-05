module ag-core/tool/cmd/protoc-gen-go-agapi

go 1.24.8

replace ag-core => ../../../

require (
	ag-core v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/cloudwego/kitex v0.14.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20251222181119-0a764e51fe1b // indirect
)
