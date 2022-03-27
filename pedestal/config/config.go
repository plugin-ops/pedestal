package config

var ( // public
	Version   string
	PluginDir string
	RuleDir   string
)

var ( // interact with the server
	CenterIP       string
	CenterHttpPort int
	CenterGRCPPort int
	Token          string
)

var ( // When used as a http server
	HttpPort         = 8888
	HttpIP           = "0.0.0.0"
	EnableHttpServer bool
)
