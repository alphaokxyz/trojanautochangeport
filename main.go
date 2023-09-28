package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

type TrojanConfig struct {
	RunType    string      `json:"run_type"`
	LocalAddr  string      `json:"local_addr"`
	LocalPort  int         `json:"local_port"`
	RemoteAddr string      `json:"remote_addr"`
	RemotePort int         `json:"remote_port"`
	Password   []string    `json:"password"`
	LogLevel   int         `json:"log_level"`
	SSL        SSLConfig   `json:"ssl"`
	TCP        TCPConfig   `json:"tcp"`
	MySQL      MySQLConfig `json:"mysql"`
}

type SSLConfig struct {
	Cert               string         `json:"cert"`
	Key                string         `json:"key"`
	KeyPassword        string         `json:"key_password"`
	Cipher             string         `json:"cipher"`
	CipherTLS13        string         `json:"cipher_tls13"`
	PreferServerCipher bool           `json:"prefer_server_cipher"`
	ALPN               []string       `json:"alpn"`
	ALPNPortOverride   map[string]int `json:"alpn_port_override"`
	ReuseSession       bool           `json:"reuse_session"`
	SessionTicket      bool           `json:"session_ticket"`
	SessionTimeout     int            `json:"session_timeout"`
	PlainHTTPResponse  string         `json:"plain_http_response"`
	Curves             string         `json:"curves"`
	DHParam            string         `json:"dhparam"`
}

type TCPConfig struct {
	PreferIPv4   bool `json:"prefer_ipv4"`
	NoDelay      bool `json:"no_delay"`
	KeepAlive    bool `json:"keep_alive"`
	ReusePort    bool `json:"reuse_port"`
	FastOpen     bool `json:"fast_open"`
	FastOpenQLen int  `json:"fast_open_qlen"`
}

type MySQLConfig struct {
	Enabled    bool   `json:"enabled"`
	ServerAddr string `json:"server_addr"`
	ServerPort int    `json:"server_port"`
	Database   string `json:"database"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Key        string `json:"key"`
	Cert       string `json:"cert"`
	CA         string `json:"ca"`
}

func main() {
	configFile := "/usr/local/etc/trojan/config.json"
	logFile := "/file/log/log.txt"
	scheduleTime := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 5, 35, 0, 0, time.Local)

	portIncrement := 1

	for {
		now := time.Now()
		if now.After(scheduleTime) {
			scheduleTime = scheduleTime.Add(24 * time.Hour)
		}
		timeUntilNextRun := scheduleTime.Sub(now)

		fmt.Printf("等待 %s 直到下一个执行时间...\n", timeUntilNextRun)
		time.Sleep(timeUntilNextRun)

		configData, err := os.ReadFile(configFile)
		if err != nil {
			fmt.Printf("读取文件失败%v\n", err)
			continue
		}

		var config TrojanConfig
		err = json.Unmarshal(configData, &config)
		if err != nil {
			fmt.Printf("解析JSON失败%v\n", err)
			continue
		}

		currentTime := time.Now().Format(time.RFC3339)
		logMessage := fmt.Sprintf("[%s] 修改前 local_port 值%d\n", currentTime, config.LocalPort)
		err = appendLog(logFile, logMessage)
		if err != nil {
			fmt.Printf("写入日志失败%v\n", err)
		}

		config.LocalPort += portIncrement

		modifiedConfigData, err := json.MarshalIndent(config, "", "    ")
		if err != nil {
			fmt.Printf("重新编码JSON失败%v\n", err)
			continue
		}

		err = os.WriteFile(configFile, modifiedConfigData, 0644)
		if err != nil {
			fmt.Printf("写入文件失败%v\n", err)
			continue
		}

		logMessage = fmt.Sprintf("[%s] 修改后 local_port 值%d\n", currentTime, config.LocalPort)
		err = appendLog(logFile, logMessage)
		if err != nil {
			fmt.Printf("写入日志失败%v\n", err)
		}

		cmd := exec.Command("systemctl", "restart", "trojan")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			fmt.Printf("执行命令失败%v\n", err)
			continue
		}
	}
}

func appendLog(logFile string, message string) error {
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.WriteString(file, message)
	if err != nil {
		return err
	}

	return nil
}
