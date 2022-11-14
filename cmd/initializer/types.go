package main

type StorageConfig struct {
	Type    string `toml:"type"`
	DataDir string `toml:"data_dir"`
}

type MetaClientOptions struct {
	MetaSrvAddr    string `toml:"metasrv_addr"`
	Timeout        int32  `toml:"timeout_millis"`
	ConnectTimeout int32  `toml:"connect_timeout_millis"`
	TCPNoDelay     bool   `toml:"tcp_nodelay"`
}

type DatanodeConfig struct {
	NodeID              uint64 `toml:"node_id"`
	HTTPAddr            string `toml:"http_addr"`
	RPCAddr             string `toml:"rpc_addr"`
	WALDir              string `toml:"wal_dir"`
	RPCRuntimeSize      int32  `toml:"rpc_runtime_size"`
	Mode                string `toml:"mode"`
	MySQLAddr           string `toml:"mysql_addr"`
	MySQLRuntimeSize    int32  `toml:"mysql_runtime_size"`
	PostgresAddr        string `toml:"postgres_addr"`
	PostgresRuntimeSize int32  `toml:"postgres_runtime_size"`

	StorageConfig     `toml:"storage"`
	MetaClientOptions `toml:"meta_client_opts"`
}
