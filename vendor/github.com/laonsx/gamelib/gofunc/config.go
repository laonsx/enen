package gofunc

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

func init() {

	AppPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	AppPath = AppPath + string(os.PathSeparator)
	AppPath = filepath.Join(AppPath + "..")
}

var AppPath string

const (
	CONFIGS = "configs"
	DATAS   = "datas"
)

//LoadJsonConf 加载json静态配置数据
//model：configs or data
//name: 文件名
func LoadJsonConf(model, name string, v interface{}) {

	fileName := filepath.Join(AppPath, model, name+".json")

	err := loadJsonFile(fileName, v)
	if err != nil {

		panic(err)
	}
}

func loadJsonFile(file string, v interface{}) error {

	buf, err := ioutil.ReadFile(file)
	if err != nil {

		return err
	}

	err = json.Unmarshal(buf, v)

	return err
}
