package main

import (
	"fmt"
	"github.com/yddeng/dnet/dhttp"
	"strings"
)

var url = "http://127.0.0.1:23455/script"

func main() {
	fmt.Println("1: create 2:update 3:delete")
	var key int
	fmt.Scan(&key)

	var resp map[string]interface{}
	var err error
	switch key {
	case 1:
		fmt.Println("name args")
		var name, args string
		fmt.Scan(&name, &args)
		reqUrl := fmt.Sprintf("%s/create", url)
		req, _ := dhttp.PostJson(reqUrl, map[string]interface{}{
			"name": name,
			"args": strings.ReplaceAll(args, "&", " "),
		})
		err = req.ToJSON(&resp)
	case 2:
		fmt.Println("id name args")
		var id int
		var name, args string
		fmt.Scan(&id, &name, &args)
		reqUrl := fmt.Sprintf("%s/update", url)
		req, _ := dhttp.PostJson(reqUrl, map[string]interface{}{
			"id":   id,
			"name": name,
			"args": strings.ReplaceAll(args, "&", " "),
		})
		err = req.ToJSON(&resp)
	case 3:
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
