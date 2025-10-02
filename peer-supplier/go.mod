module peer-supplier

go 1.23.0

require (
	github.com/gorilla/mux v1.8.0
	go.mongodb.org/mongo-driver v1.12.1
	google.golang.org/grpc v1.75.1
	google.golang.org/protobuf v1.36.9
	share v0.0.0
)

replace share => ../share

require (
	github.com/golang/snappy v0.0.4 // indirect
	github.com/klauspost/compress v1.15.14 // indirect
	github.com/montanaflynn/stats v0.0.0-20171201202039-1bf9dbcd8cbe // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sync v0.15.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
)
