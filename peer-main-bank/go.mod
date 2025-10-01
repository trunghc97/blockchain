module peer-main-bank

go 1.21

require (
	github.com/gorilla/mux v1.8.0
	go.mongodb.org/mongo-driver v1.12.1
	share v0.0.0
)

replace share => ../share
