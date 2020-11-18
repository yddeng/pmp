package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/yddeng/dnet/dhttp"
	"io"
	"io/ioutil"
	"os"
)

var url = "http://127.0.0.1:23455/file"

func main() {
	fmt.Println("1: get 2:update 3:delete")
	var key int
	fmt.Scan(&key)

	var resp map[string]interface{}
	var err error
	switch key {
	case 1:
		fmt.Println("path")
		var name string
		fmt.Scan(&name)
		reqUrl := fmt.Sprintf("%s/get?path=%s", url, name)
		req, _ := dhttp.Get(reqUrl)
		err = req.ToJSON(&resp)
	case 2:
		fmt.Println("dir file filename")
		var dir, file, filename string
		fmt.Scan(&dir, &file, &filename)

		data, err := ioutil.ReadFile(file)
		if err != nil {
			panic(err)
		}

		h := md5.New()
		f, err := os.Open(file)
		if err != nil {
			panic(err)
		}

		_, err = io.Copy(h, f)
		if err != nil {
			panic(err)
		}

		md5 := hex.EncodeToString(h.Sum(nil))

		fmt.Println("md5", md5)

		size := 100
		length := len(data)
		total := length / size
		if length%size > 0 {
			total++
		}

		crt := 1
		start := 0
		for start < length {
			reqUrl := fmt.Sprintf("%s/update", url)
			reqData := map[string]interface{}{
				"dir":      dir,
				"filename": filename,
				"current":  crt,
				"total":    total,
				"md5":      md5,
			}
			if length-start > size {
				reqData["data"] = data[start : start+size]
			} else {
				reqData["data"] = data[start:]
			}

			req, _ := dhttp.PostJson(reqUrl, reqData)
			start += size
			crt += 1

			err = req.ToJSON(&resp)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(resp)
			}
		}
	case 3:
		fmt.Println("path file")
		var name, file string
		fmt.Scan(&name, &file)
		reqUrl := fmt.Sprintf("%s/delete", url)
		req, _ := dhttp.PostJson(reqUrl, map[string]interface{}{
			"dir":      name,
			"filename": file,
		})
		err = req.ToJSON(&resp)
	}

	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(resp)

}
