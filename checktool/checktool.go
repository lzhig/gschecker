package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type fixStepInfo struct {
	name string
	f    func(r *reportT) error
}

type configStepInfo struct {
	name string
	f    func(c configT, r *reportT) error
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: exe configURL postURL")
		return
	}
	configURL := os.Args[1]
	postURL := os.Args[2]

	var report string
	step := 1

	var r reportT

	fixSteps := []fixStepInfo{
		{"获取系统信息", getSystemInfoToReport},
		{"获取公网ip", getExternalIPToReport},
		{"获取DNS服务器", getDNSServerToReport},
		{"检测 www.gs108.com 域名解析", resolveDNSGS108ToReport},
	}

	for ndx := range fixSteps {
		s := fixSteps[ndx]

		fmt.Printf("%d. 正在%s...\n", step, s.name)

		report += fmt.Sprintf("%d. %s\n", step, s.name)
		step++

		if s.f == nil {
			continue
		}

		if err := s.f(&r); err != nil {
			break
		}
	}

	var c configT
	if err := getConfig(configURL, &c); err != nil {
		fmt.Println(err)
	}

	configSteps := []configStepInfo{
		{"检测指定域名解析", resolveDNSSpecificToReport},
		{"检测指定地址的路由", resolveRoutersToReport},
		{"检测文件下载速度", checkDownloadSpeedToReport},
		{"检测游戏服务器速度", checkGameServerSpeedToReport},
	}

	for ndx := range configSteps {
		s := configSteps[ndx]

		fmt.Printf("%d. 正在%s...\n", step, s.name)

		report += fmt.Sprintf("%d. %s\n", step, s.name)
		step++

		if s.f == nil {
			continue
		}

		if err := s.f(c, &r); err != nil {
			break
		}
	}

	r1, _ := json.Marshal(r)
	//fmt.Println(string(r1))

	body := bytes.NewBuffer(r1)
	res, err := http.Post(postURL, "application/json;charset=utf-8", body)
	if err != nil {
		fmt.Println("提交错误:", err)
	} else {
		ret, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println("error:", err)
		} else {
			res.Body.Close()
			fmt.Println("结果:", string(ret))
		}
	}

	ioutil.WriteFile("report.txt", r1, os.ModePerm)

	fmt.Println("exit.")
}
