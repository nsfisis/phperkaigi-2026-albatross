package gen

//go:generate go tool sqlc generate
//go:generate go tool oapi-codegen -config oapi-codegen.api-server.yaml ../../openapi/api-server.yaml
//go:generate go tool oapi-codegen -config oapi-codegen.fortee.yaml ../../openapi/fortee.yaml
//go:generate go run ./api/handler_wrapper_gen.go -i ../api/generated.go -o ../api/handler_wrapper.go
//go:generate go run ./taskqueue/processor_wrapper_gen.go -i ../taskqueue/tasks.go -o ../taskqueue/processor_wrapper.go
