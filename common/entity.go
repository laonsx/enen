package common

type MySQLConf struct {
	Name   string `json:"name"`
	User   string `json:"user"`
	Dbname string `json:"dbname"`
	Addr   string `json:"addr"`
}

//rpc config entity
type (
	RpcConf struct {
		Node   []*NodeConf   `json:"node"`
		Method []*MethodConf `json:"method"`
	}

	NodeConf struct {
		Name string `json:"name"`
		Addr string `json:"addr"`
	}

	MethodConf struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}
)

type Config struct {
	HttpAddr      string `json:"http_addr"`
	WebSocketAddr string `json:"web_socket_addr"`
	RpcAddr       string `json:"rpc_addr"`
	PprofAddr     string `json:"pprof_addr"`
	NsqAddr       string `json:"nsq_addr"`
	NsqTopic      string `json:"nsq_topic"`
	LogDir        string `json:"log_dir"`
	OriginAllow   string `json:"origin_allow"`
}

type ServiceConf map[string]*Config
