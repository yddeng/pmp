package main

import (
	"fmt"
	"github.com/yddeng/dnet/dhttp"
)

var url = "http://127.0.0.1:23455/item"

func main() {
	fmt.Println("1: create 2:delete")
	var key int
	fmt.Scan(&key)

	var resp map[string]interface{}
	var err error
	switch key {
	case 1:
		fmt.Println("name scriptId nodeName")
		var name, node string
		var id int
		fmt.Scan(&name, &id, &node)
		reqUrl := fmt.Sprintf("%s/create", url)
		req, _ := dhttp.PostJson(reqUrl, map[string]interface{}{
			"script": id,
			"name":   name,
			"slave":  node,
		})
		err = req.ToJSON(&resp)
	case 2:
		fmt.Println("id")
		var id int
		fmt.Scan(&id)
		reqUrl := fmt.Sprintf("%s/delete", url)
		req, _ := dhttp.PostJson(reqUrl, map[string]interface{}{
			"id": id,
		})
		err = req.ToJSON(&resp)
	}

	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(resp)
}
