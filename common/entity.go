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
	HttpAddr       string `json:"http_addr"`
	WebSocketAddr  string `json:"web_socket_addr"`
	GateAddr       string `json:"gate_addr"` // 客户端请求gate地址
	RpcAddr        string `json:"rpc_addr"`
	PprofAddr      string `json:"pprof_addr"`
	NsqAddr        string `json:"nsq_addr"`
	NsqTopic       string `json:"nsq_topic"`
	Log            string `json:"log"` // 日志
	OriginAllow    string `json:"origin_allow"`
	Weight         int32  `json:"weight"`           // gate服务分配权重
	GameNodeName   string `json:"game_node_name"`   // gate连接game名
	CenterNodeName string `json:"center_node_name"` // gate注册center名
	CenterAddr     string `json:"center_addr"`      // test测试center地址
}

type ServiceConf map[string]*Config
