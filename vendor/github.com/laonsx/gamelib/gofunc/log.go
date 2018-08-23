package gofunc

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
)

func PrintPanic() {

	if err := recover(); err != nil {

		log.Println("[panic]", err)

		i := 0
		for {

			funcName, file, line, ok := runtime.Caller(i)
			if !ok {

				break
			}

			name := runtime.FuncForPC(funcName).Name()

			log.Printf("[panic] %v func:%v file:%v line:%v", i, name, file, line)

			i++
		}
	}
}

func Pprof(addr string) {

	go func() {

		log.Println(http.ListenAndServe(addr, nil))
	}()
}
