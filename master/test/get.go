package main

import (
	"fmt"
	"github.com/yddeng/dnet/dhttp"
)

var url = "http://127.0.0.1:23455"

func main() {
	{
		reqUrl := fmt.Sprintf("%s/script/get", url)
		req, _ := dhttp.Get(reqUrl)
		var resp map[string]interface{}
		if err := req.ToJSON(&resp); err != nil {
			fmt.Println("script", err)
		}
		fmt.Println("script", resp)
	}

	{
		reqUrl := fmt.Sprintf("%s/item/get", url)
		req, _ := dhttp.Get(reqUrl)
		var resp map[string]interface{}
		if err := req.ToJSON(&resp); err != nil {
			fmt.Println("item", err)
		}
		fmt.Println("item", resp)
	}

	{
		reqUrl := fmt.Sprintf("%s/node/get?name=list", url)
		req, _ := dhttp.Get(reqUrl)
		var resp map[string]interface{}
		if err := req.ToJSON(&resp); err != nil {
			fmt.Println("node", err)
		}
		fmt.Println("node", resp)
	}
}
