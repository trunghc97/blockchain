module peer-main-bank

go 1.21

require shared/peer-base v0.0.0

replace shared/peer-base => ../shared/peer-base

require (
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/rs/cors v1.10.1 // indirect
	go.mongodb.org/mongo-driver v1.12.1 // indirect
	google.golang.org/grpc v1.59.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)