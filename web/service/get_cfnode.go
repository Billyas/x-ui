package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"x-ui/database/model"
)

type Info struct {
	Ip   string `json:"ip"`
	Line string `json:"line"`
}

type Response struct {
	Code string `json:"code"`
	Info []Info `json:"info"`
}

func (s *GetSubService) GetLatestCFNodeProxy() (string, error) {
	fmt.Println("执行CFNode定时器")
	var cfNodes strings.Builder

	// 1. 数据库查找模版链接
	dynodes, err := s.subService.GetSubsBySubType(model.DynNode)
	if err != nil {
		return "", err
	}

	// 2. 请求CF节点列表
	requestUrl := "https://api.hostmonit.com/get_optimization_ip"
	params := "{\"key\":\"iDetkOys\"}"
	headers := map[string]string{
		"authority":    "api.hostmonit.com",
		"pragma":       "no-cache",
		"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		"content-type": "application/json",
		"Accept":       "*/*",
		"Host":         "api.hostmonit.com",
		"Connection":   "keep-alive",
	}

	lineMap := map[string]string{
		"CM": "移动",
		"CU": "联通",
		"CT": "电信",
	}

	ipLines := make(map[string][]string)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", requestUrl, strings.NewReader(params))
	for key, value := range headers {
		req.Header.Add(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	body, _ := io.ReadAll(resp.Body)
	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	if response.Code == "200" {
		for _, info := range response.Info {
			ip := info.Ip
			line := url.QueryEscape(lineMap[info.Line])
			ipLines[line] = append(ipLines[line], ip)
		}
	}

	// 3. 遍历模版链接，将CF节点与模版链接拼接，在{}这个位置换成ip
	for _, subTemplate := range dynodes {
		templateUrl := subTemplate.Url
		for line, ips := range ipLines {
			for _, ip := range ips {
				cfNode := fmt.Sprintf(templateUrl, ip, line)
				cfNodes.WriteString(cfNode + "\n")
			}
		}
	}

	return cfNodes.String(), nil
}
