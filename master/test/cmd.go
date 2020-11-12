package main

import (
	"fmt"
	"github.com/yddeng/dnet/dhttp"
)

var url = "http://127.0.0.1:23455/itemCmd"

func main() {
	fmt.Println("1: start 2:stop 3:kill")
	var key int
	fmt.Scan(&key)

	var resp map[string]interface{}
	var err error
	switch key {
	case 1:
		fmt.Println("itemID")
		var id int
		fmt.Scan(&id)
		reqUrl := fmt.Sprintf("%s/start", url)
		req, _ := dhttp.PostJson(reqUrl, map[string]interface{}{
			"id": id,
		})
		err = req.ToJSON(&resp)
	case 2:
		fmt.Println("itemID")
		var id int
		fmt.Scan(&id)
		reqUrl := fmt.Sprintf("%s/signal", url)
		req, _ := dhttp.PostJson(reqUrl, map[string]interface{}{
			"id":     id,
			"signal": "term",
		})
		err = req.ToJSON(&resp)
	case 3:
		fmt.Println("itemID")
		var id int
		fmt.Scan(&id)
		reqUrl := fmt.Sprintf("%s/signal", url)
		req, _ := dhttp.PostJson(reqUrl, map[string]interface{}{
			"id":     id,
			"signal": "kill",
		})
		err = req.ToJSON(&resp)
	}

	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(resp)
}
