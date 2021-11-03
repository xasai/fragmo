package config

var (
	//Http server
	HttpServerAddr = "0.0.0.0:8000"
	TemplatesPath  = "templates/"

	//Storage server
	StorageServerAddr = "storage:4242"
	StoragePath       = "storage/"
	RecordsFilename   = "file_records.csv"

	FragmentSize = int64(1 << 20) // 1 MB
)
