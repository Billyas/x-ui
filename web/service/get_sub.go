package service

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"x-ui/database/model"
	"x-ui/logger"
)

type GetSubService struct {
	subService SubService
}

func (s *GetSubService) GetLatestUrlSub() (string, error) {
	var stringBuilder strings.Builder
	// 获取排除节点规则
	excludeNodes, _ := s.subService.GetSubsBySubType(model.ExcludeNode)
	var excludePatterns []string
	for _, node := range excludeNodes {
		// 支持多种分隔符：|、，、,、、
		keywords := s.splitByMultipleDelimiters(node.Url)
		for _, keyword := range keywords {
			keyword = strings.TrimSpace(keyword)
			if keyword != "" {
				excludePatterns = append(excludePatterns, keyword)
			}
		}
	}

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

			req.Header.Set("User-Agent", "V2rayN/7.16.6")

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

			// 过滤订阅内容中的节点
			filteredContent := s.filterNodesByRemark(string(subStr), excludePatterns)
			stringBuilder.WriteString(filteredContent)
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

// filterNodesByRemark 根据节点备注过滤节点
func (s *GetSubService) filterNodesByRemark(content string, excludePatterns []string) string {
	if len(excludePatterns) == 0 {
		return content
	}

	var filteredBuilder strings.Builder
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 检查是否需要排除该节点
		if s.shouldExcludeNode(line, excludePatterns) {

		} else {
			filteredBuilder.WriteString(line)
			filteredBuilder.WriteString("\n")
		}
	}

	return filteredBuilder.String()
}

// shouldExcludeNode 检查节点是否应该被排除
func (s *GetSubService) shouldExcludeNode(nodeUrl string, excludePatterns []string) bool {
	// 提取节点备注：所有节点格式都是 协议://其他信息#名称，直接提取#后面的内容
	remark := s.extractNodeRemark(nodeUrl)
	if remark == "" {
		return false
	}

	// 检查是否匹配任何排除模式
	for _, pattern := range excludePatterns {
		if strings.Contains(remark, pattern) {
			return true
		}
	}

	return false
}

// extractNodeRemark 从节点URL中提取备注
func (s *GetSubService) extractNodeRemark(nodeUrl string) string {
	// 特殊处理VMess协议
	if strings.HasPrefix(nodeUrl, "vmess://") {
		return s.extractVMessRemark(nodeUrl)
	}

	// 其他协议：协议://其他信息#名称，直接提取#后面的内容并进行URL解码
	remarkIndex := strings.Index(nodeUrl, "#")
	if remarkIndex == -1 {
		// 没有#，无法提取备注
		return ""
	}

	// 提取#后面的内容并进行URL解码
	encodedRemark := nodeUrl[remarkIndex+1:]
	decodedRemark, err := url.QueryUnescape(encodedRemark)
	if err != nil {
		// 解码失败，返回原始编码内容
		return encodedRemark
	}

	return decodedRemark
}

// extractVMessRemark 从VMess链接中提取备注
func (s *GetSubService) extractVMessRemark(vmessUrl string) string {
	// 去除vmess://前缀
	vmessData := strings.TrimPrefix(vmessUrl, "vmess://")

	// 尝试使用标准base64解码
	decoded, err := base64.StdEncoding.DecodeString(vmessData)
	if err != nil {
		// 尝试使用URL安全的base64解码
		vmessData = strings.ReplaceAll(vmessData, "-", "+")
		vmessData = strings.ReplaceAll(vmessData, "_", "/")
		// 补全padding
		padding := len(vmessData) % 4
		if padding > 0 {
			vmessData += strings.Repeat("=", 4-padding)
		}
		decoded, err = base64.StdEncoding.DecodeString(vmessData)
		if err != nil {
			// 解码失败，无法提取备注
			return ""
		}
	}

	// 解析JSON，提取ps字段
	var vmessInfo map[string]interface{}
	err = json.Unmarshal(decoded, &vmessInfo)
	if err != nil {
		return ""
	}

	// 获取ps字段（节点名称）
	if ps, ok := vmessInfo["ps"].(string); ok {
		return ps
	}

	return ""
}

// splitByMultipleDelimiters 按多种分隔符分割字符串
func (s *GetSubService) splitByMultipleDelimiters(input string) []string {
	// 将所有分隔符替换为同一分隔符，然后分割
	str := input
	str = strings.ReplaceAll(str, "|", ",")
	str = strings.ReplaceAll(str, "，", ",")
	str = strings.ReplaceAll(str, "、", ",")
	return strings.Split(str, ",")
}
