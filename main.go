package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"net/http"
	"os/exec"
	"strings"
	"github.com/spf13/viper"
)

type IPInfo struct {
	IP       string   `json:"ip"`
	Location []string `json:"location"`
}

type IPResponse struct {
	Ret  string `json:"ret"`
	Data IPInfo `json:"data"`
}

func checkInitialize() (machineType, serid string, err error) {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config/")
	err = viper.ReadInConfig()
	if err != nil {
		fmt.Println(fmt.Errorf("\033[91mFatal error config file: %s\033[0m", err))
		os.WriteFile(
			"./example.yaml",
			[]byte(
				`domain:
  main: example.com # 主服域名
  current: example.example.com # 本服对应域名

secret:
  current: none # 保留
				`),
			0644,
		)
        os.Exit(1)
	}
    var config Config
    viper.Unmarshal(&config)
    required := []string{
        "domain.main",
        "domain.current",
        "secret.current",
    }
    var flag bool = false
    for _, field := range required {
        if ! viper.IsSet(field) {
            flag = true
            fmt.Printf("\033[91mParam %s does not exist.\033[0m\n", field)
        }
    }
    if flag {os.Exit(1)}

	fmt.Print("机器类型:")
	if viper.Get("domain.main") == viper.Get("domain.current") {
		fmt.Println("主服务器")
	} else {
		fmt.Println("从服务器")
	}
	fmt.Printf("机器域名:%s ", viper.Get("domain.current"))

	fmt.Print("\033[93m测试中...\033[0m")

	resp, err := http.Get("https://myip.ipip.net/json")
	if err != nil {
		fmt.Printf("\r机器域名:%s \033[91m网络不通: %v\033[0m\n", viper.Get("domain.current"), err)
		return "", "", fmt.Errorf("网络不通: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("\r机器域名:%s \033[91m读取内容失败: %v\033[0m\n", viper.Get("domain.current"), err)
		return "", "", fmt.Errorf("读取内容失败: %v", err)
	}

	var result IPResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("\r机器域名:%s \033[91m解析JSON失败: %v\033[0m\n", viper.Get("domain.current"), err)
		return "", "", fmt.Errorf("解析JSON失败: %v", err)
	}
	if result.Ret != "ok" {
		fmt.Printf("\r机器域名:%s \033[91m接口返回错误: %v\033[0m\n", viper.Get("domain.current"), err)
		return "", "", fmt.Errorf("接口返回错误: %s", result.Ret)
	}

	fmt.Printf("\r机器域名:%s \033[92m本机公网 IP: %s\033[0m\n", viper.Get("domain.current"), result.Data.IP)

	cmd := exec.Command("dig", viper.GetString("domain.current"), "+short")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("\033[91mdig 执行失败: %v\033[0m", err)
		return "", "", fmt.Errorf("dig 执行失败: %v", err)
	}

	ips := strings.Split(strings.TrimSpace(string(output)), "\n")
	matched := false
	for _, ip := range ips {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}
		if ip == result.Data.IP {
			matched = true
			break
		}
	}

	if !matched {
		fmt.Printf("\033[91mIP校验失败\033[0m，本机IP: \033[92m%s\033[0m，解析IP: \033[91m%s\033[0m", result.Data.IP, strings.Join(ips, ", "))
		return "", "", fmt.Errorf("IP校验失败，本机IP: %s，解析IP: %s", result.Data.IP, strings.Join(ips, ", "))
	}

	fmt.Println("\033[92m测试完成\033[0m")
	var tmp string
	if viper.Get("domain.main") == viper.Get("domain.current") {
		tmp = "主服务器"
	} else {tmp = "从服务器"}
	return tmp, viper.GetString("domain.current"), nil
}

func main() {
	
	machineType, serid, err := checkInitialize()
	if err != nil {
		fmt.Printf("\n初始化失败。\n")
		return
	}
	fmt.Printf("配置完成: 类型=%s, 域名=%s\n", machineType, serid)
}