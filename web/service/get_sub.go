package service

import (
	"crypto/tls"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
	"x-ui/database/model"
	"x-ui/logger"
)

type GetSubService struct {
	subService SubService
}

func (s *GetSubService) GetLatestUrlSub() (string, error) {
	var stringBuilder strings.Builder
	// 0. 如果存在首选节点先添加
	firstNode, err := s.subService.GetSubsBySubType(model.FistNode)
	// 如果存在首选节点则添加,不存在执行下面的操作
	if err == nil && len(firstNode) > 0 {
		for _, node := range firstNode {
			stringBuilder.WriteString(node.Url)
			stringBuilder.WriteString("\n")
		}
	} else if err != nil {
		logger.Errorf("步骤0：" + err.Error())
	}

	// 1. 对所有订阅进行HTTP请求并解码base64响应
	    subs, err := s.subService.GetSubsBySubType(model.SubURL)
	    if err == nil && len(subs) > 0 {
	        for _, sub := range subs {
	            client := &http.Client{
	                Transport: &http.Transport{
	                    TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	                },
	            }
	
	            req, err := http.NewRequest("GET", sub.Url, nil)
	            if err != nil {
	                logger.Errorf(err.Error())
	                continue
	            }
	
	            req.Header.Set("User-Agent", "V2rayN/5.0.1 Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	
	            response, err := client.Do(req)
	            if err != nil {
	                logger.Errorf(err.Error())
	                continue
	            }
	
	            defer response.Body.Close()
	
	            body, err := io.ReadAll(response.Body)
	            if err != nil {
	                logger.Errorf(err.Error())
	                continue
	            }
	
	            subStr, err := base64.StdEncoding.DecodeString(string(body))
	            if err != nil {
	                logger.Errorf(err.Error())
	                continue
	            }
	
	            stringBuilder.WriteString(string(subStr))
		    stringBuilder.WriteString("\n")
	        }
	    } else if err != nil {
	        logger.Errorf("步骤1：" + err.Error())
	    }

	// 2. 获取自定义节点并将其添加到订阅内容中
	nodes, err := s.subService.GetSubsBySubType(model.OwnNode)

	if err == nil && len(nodes) > 0 {
		for _, node := range nodes {
			stringBuilder.WriteString(node.Url)
			stringBuilder.WriteString("\n")
		}
	} else if err != nil {
		logger.Errorf("步骤2：" + err.Error())
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
