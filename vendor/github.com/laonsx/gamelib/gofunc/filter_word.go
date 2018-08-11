package gofunc

import (
	"bytes"
	"io/ioutil"
	"strings"
)

type groupWords struct {
	key  string
	list []string
}

var filterWordMap map[string]*groupWords

//FilterWord 简单的敏感词检查 过滤
//参数 str 需要检查的字符串 replace 是否把关键词替换为**
//返回 （是否有关键词 过滤后的字符串）
func FilterWord(str string, replace bool) (bool, string) {

	var iskw bool
	rs := []rune(str)
	for i, v := range rs {

		key := string(v)
		if gw, ok := filterWordMap[key]; ok {

			for _, kw := range gw.list {

				a := rs[i:]
				iskw = strings.Contains(string(a), kw)
				if iskw {

					if replace {

						str = strings.Replace(str, kw, "**", 2)
					} else {

						return iskw, str
					}
				}
			}
		}
	}

	return iskw, str
}

//BuildDict 初始化敏感词数据
func BuildDict(file string) {

	filterWordMap = make(map[string]*groupWords)

	strs, err := readFile(file)
	if err != nil {

		panic("build dict error")
	}

	for _, v := range strs {

		key := SubStr(v, 0, 1)
		if _, ok := filterWordMap[key]; ok {

			filterWordMap[key].list = append(filterWordMap[key].list, v)
		} else {

			gw := new(groupWords)
			gw.key = key
			gw.list = make([]string, 0)
			gw.list = append(gw.list, v)
			filterWordMap[key] = gw
		}
	}
}

func readFile(filename string) ([]string, error) {

	var buf []byte
	buf, err := ioutil.ReadFile(filename)
	if err != nil {

		return nil, err
	}

	strs := make([]string, 0)

	lines := bytes.Split(buf, []byte("\n"))
	for _, v := range lines {

		if string(v) == "" {

			continue
		}

		strs = append(strs, strings.TrimSpace(string(v)))
	}

	return strs, err
}
