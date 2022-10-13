wasm:
	GOOS=js GOARCH=wasm go build -o ./public/server.wasm ./cmd/linkport
run-wasm:
	caddy file-server -listen :7072 -browse -root ./public
