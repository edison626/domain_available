package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	fileName        = "domains.txt"     // 文件名
	requestInterval = 300 * time.Second // 监控间隔
	maxResponseTime = 2 * time.Second   // 响应时间阈值
)

func sendTelegramMessage(botToken, chatID, message string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	data := url.Values{
		"chat_id": {chatID},
		"text":    {message},
	}

	resp, err := http.PostForm(apiURL, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message, status: %s", resp.Status)
	}

	return nil
}

type Response struct {
	Status bool   `json:"status"`
	Data   string `json:"data"`
}

func monitorDomain(domain string, ip string) {
	botToken := "6559646803:AAFTKQmzFnx1dzbDT9z3mkqU_RzF2lBw_Fs"
	chatID := "-972839729"

	if !strings.HasPrefix(domain, "http://") && !strings.HasPrefix(domain, "https://") {
		domain = "https://" + domain + "/member/banner/" // 默认添加http前缀，或者您可以选择https /member/banner/
	}

	//client := &http.Client{} // 创建新的 HTTP 客户端

	for {
		client := &http.Client{}
		req, err := http.NewRequest("GET", domain, nil)
		if err != nil {
			fmt.Printf("Error fetching %s: %s\n", domain, err)
			time.Sleep(requestInterval)
			continue
		}
		resp, err := client.Do(req) // 使用 client 发起请求
		if err != nil {
			fmt.Printf("Error fetching %s: %s\n", domain, err)
			time.Sleep(requestInterval)
			continue
		}

		//responseTime := responseDomain.Header.Get("Date")
		//dateTime, _ := time.Parse(time.RFC1123, responseTime)
		//duration := time.Since(dateTime)

		if resp.StatusCode != http.StatusOK {
			log.Fatalf("Error status code: %d", resp.StatusCode)
			message := fmt.Sprintf("Error status code: %d", resp.StatusCode)
			sendTelegramMessage(botToken, chatID, message)
		}

		var response Response
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			log.Fatalf("Error decoding response: %v", err)
		}

		if response.Status {
			fmt.Printf("[服务器: %s] 警告:Status 变 true - 域名 %s , Status= %t\n ", ip, domain, response.Status)
			//纸飞机报警
			message := fmt.Sprintf("[服务器: %s] 警告 : 域名 %s , Status= %t , Data= %s\n ", ip, domain, response.Status, response.Data)
			sendTelegramMessage(botToken, chatID, message)
		} else {
			fmt.Printf("[服务器: %s] 正常 : 域名 %s , Status= %t\n ", ip, domain, response.Status)
		}

		time.Sleep(requestInterval)
	}
}

func serverIp() (string, error) {
	resp, err := http.Get("https://api.ipify.org?format=text")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(ip), nil
}

func main() {

	ip, err := serverIp()
	if err != nil {
		fmt.Println("Error fetching IP:", err)
		return
	}

	//fmt.Println("Your external IP is:", ip)

	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("Error opening file: %s\n", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		domain := scanner.Text()
		go monitorDomain(domain, ip) // 使用 goroutine 并发监控每个域名
	}

	select {} // 阻止主 goroutine 退出
}
