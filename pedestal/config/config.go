package config

var Version string

var ( // When used as a command line tool
	PluginDir string
	//autoFixDir bool // TODO 可以考虑自动修正目录, 因为pluginDir对应的目录下必须有src文件夹, 且所有插件文件夹都应当放在src文件夹下
	RulePath string
)

var ( // When used as a http server
	HttpPort      = 8888
	HttpIP        = "0.0.0.0"
	EnableSwagger bool
)
