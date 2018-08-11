package gofunc

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

func init() {

	appPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	appPath = appPath + string(os.PathSeparator)
	appPath = filepath.Join(appPath + "..")

}

var appPath string

const (
	CONFIGS = "configs"
	DATAS   = "datas"
)

//LoadJsonConf 加载json静态配置数据
//model：configs or data
//name: 文件名
func LoadJsonConf(model, name string, v interface{}) {

	fileName := filepath.Join(appPath, model, name+".json")

	err := loadJsonFile(fileName, v)
	if err != nil {

		panic(err)
	}
}

//GetAppPath 获取应用路径
func GetAppPath() string {

	return appPath
}

func loadJsonFile(file string, v interface{}) error {

	buf, err := ioutil.ReadFile(file)
	if err != nil {

		return err
	}

	err = json.Unmarshal(buf, v)

	return err
}

// GameConf 游戏策划配置
type GameConf struct {
	mux  sync.RWMutex
	data map[string]interface{}
}

// NewGameConf 初始化
func NewGameConf() *GameConf {
	gameConf := new(GameConf)
	gameConf.data = make(map[string]interface{})
	return gameConf
}

// GetConf 数据
func (gameConf *GameConf) GetConf(key string) interface{} {

	gameConf.mux.RLock()
	defer gameConf.mux.RUnlock()

	return gameConf.data[key]
}

// SetConf 数据
func (gameConf *GameConf) SetConf(key string, v interface{}) {

	gameConf.mux.Lock()
	defer gameConf.mux.Unlock()

	gameConf.data[key] = v
}
