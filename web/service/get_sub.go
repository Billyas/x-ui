package service

import (
	"encoding/base64"
	"io"
	"net/http"
	"strings"
	"x-ui/database/model"
)

type GetSubService struct {
	subService SubService
}

func (s *GetSubService) GetLatestUrlSub() (string, error) {
	var stringBuilder strings.Builder

	// 1. 对所有订阅进行HTTP请求并解码base64响应
	subs, err := s.subService.GetSubsBySubType(model.SubURL)
	if err != nil {
		return "", err
	}
	for _, sub := range subs {
		response, err := http.Get(sub.Url)
		if err != nil {
			println(err.Error())
			return "", err
		}
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		if err != nil {
			return "", err
		}

		subStr, err := base64.StdEncoding.DecodeString(string(body))
		if err != nil {
			return "", err
		}

		stringBuilder.WriteString(string(subStr))
	}

	// 2. 获取自定义节点并将其添加到订阅内容中
	nodes, err := s.subService.GetSubsBySubType(model.OwnNode)
	if err != nil {
		return "", err
	}
	for _, node := range nodes {
		stringBuilder.WriteString(node.Url)
		stringBuilder.WriteString("\n")
	}

	// 3. 添加CFNode内容
	stringBuilder.WriteString("\n")
	cfNodeContent, _ := s.GetLatestCFNode()
	stringBuilder.WriteString(cfNodeContent)

	// 4. 将订阅内容重新编码为base64并将其写回数据库
	subContent := base64.StdEncoding.EncodeToString([]byte(stringBuilder.String()))
	finNode := model.Sub{
		Id:   10,
		Type: model.FinData,
		Name: "订阅缓存",
		Url:  subContent,
	}

	err = s.subService.UpdateSub(&finNode)
	if err != nil {
		println(err.Error())
		return "", err
	}

	return subContent, nil
}
