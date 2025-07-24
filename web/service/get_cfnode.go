package service

import (
	"bytes"
	"encoding/json"
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
	Code int    `json:"code"`
	Info []Info `json:"info"`
}

func (s *GetSubService) GetLatestCFNode() (string, error) {
	var cfNodes strings.Builder

	// 1. 数据库查找模版链接
	dynodes, err := s.subService.GetSubsBySubType(model.DynNode)
	if err != nil {
		return "", err
	}
	// 如果没有找到动态节点，直接返回空结果
	if len(dynodes) == 0 {
		return "", nil
	}

	// 2. 请求CF节点列表
	requestUrl := "https://api.hostmonit.com/get_optimization_ip"
	params := map[string]string{
		"key": "iDetkOys",
	}
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
	// 将 params 转换为 JSON
	jsonParams, _ := json.Marshal(params)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", requestUrl, bytes.NewBuffer(jsonParams))
	for key, value := range headers {
		req.Header.Add(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		println(err.Error())
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
		println(err.Error())
		return "", err
	}

	if response.Code == 200 {
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
				cfNode := strings.Replace(templateUrl, "{0}", ip, -1)
				cfNode = strings.Replace(cfNode, "{1}", line, -1)
				cfNodes.WriteString(cfNode + "\n")
			}
		}
	}

	return cfNodes.String(), nil
}
