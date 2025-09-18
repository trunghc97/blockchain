package config

const (
	MongoURI          = "mongodb://root:example@mongo:27017/blockchain?authSource=admin"
	DatabaseName      = "blockchain"
	ApprovalThreshold = 2 // Số lượng approval cần thiết
	SupplierURL       = "http://supplier-mock:8082/supplier/execute"
)
