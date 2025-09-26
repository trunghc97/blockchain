package config

import "os"

var (
	MongoURI     = getMongoURI()
	DatabaseName = "blockchain"
)

func getMongoURI() string {
	if uri := os.Getenv("MONGO_URI"); uri != "" {
		return uri
	}
	return "mongodb://root:example@localhost:27017/blockchain?authSource=admin"
}
