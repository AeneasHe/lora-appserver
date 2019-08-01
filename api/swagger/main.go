// simple tool to merge different swagger definition into a single file
package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

const apiVersion = "1.0.0"

type model struct {
	Swagger  string `json:"swagger"`
	BasePath string `json:"basePath"`
	Info     struct {
		Title       string `json:"title"`
		Version     string `json:"version"`
		Description string `json:"description"`
	} `json:"info"`
	Schemes     []string               `json:"schemes"`
	Consumes    []string               `json:"consumes"`
	Produces    []string               `json:"produces"`
	Paths       map[string]interface{} `json:"paths"`
	Definitions map[string]interface{} `json:"definitions"`
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: go run main.go inputPath")
	}
	//swagger主配置文件
	swagger := model{
		Swagger:     "2.0",
		Consumes:    []string{"application/json"},
		Produces:    []string{"application/json"},
		Paths:       make(map[string]interface{}),
		Definitions: make(map[string]interface{}),
	}
	swagger.Info.Title = "LoRa App Server REST API"
	swagger.Info.Version = apiVersion
	swagger.Info.Description = `
For more information about the usage of the LoRa App Server (REST) API, see
[https://docs.loraserver.io/lora-app-server/api/](https://docs.loraserver.io/lora-app-server/api/).
`

	fileInfos, err := ioutil.ReadDir(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	//循环读取swagger子配置文件
	for _, fileInfo := range fileInfos {
		if !strings.HasSuffix(fileInfo.Name(), ".swagger.json") {
			continue
		}

		b, err := ioutil.ReadFile(path.Join(os.Args[1], fileInfo.Name()))
		if err != nil {
			log.Fatal(err)
		}

		// replace "title" by "description" for fields
		b = []byte(strings.Replace(string(b), `"title"`, `"description"`, -1))

		var m model
		//解析swagger子配置文件
		err = json.Unmarshal(b, &m)
		if err != nil {
			log.Fatal(err)
		}

		//将swagger子配置文件的路径添加到swagger主配置文件
		for k, v := range m.Paths {
			swagger.Paths[k] = v
		}
		//将swagger子配置文件的定义添加到swagger主配置文件
		for k, v := range m.Definitions {
			swagger.Definitions[k] = v
		}
	}
	//生成swagger主配置，保存到目标路径../static/swagger/api.swagger.json（在gen.sh中定义）
	enc := json.NewEncoder(os.Stdout)
	err = enc.Encode(swagger)
	if err != nil {
		log.Fatal(err)
	}
}
