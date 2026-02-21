package service

import (
	"encoding/json"
	"tenet-server/database"
	"tenet-server/model"
)

// ToolCallService 工具调用服务
type ToolCallService struct{}

// NewToolCallService 创建工具调用服务
func NewToolCallService() *ToolCallService {
	return &ToolCallService{}
}

// GetToolsByNames 根据工具名称列表查询工具定义
func (s *ToolCallService) GetToolsByNames(toolNames []string) ([]model.Tool, error) {
	if len(toolNames) == 0 {
		return nil, nil
	}

	// 从数据库查询工具定义
	var toolDefs []model.ToolCallDef
	err := database.DB.Where("name IN ? AND checkStatus = 0", toolNames).Find(&toolDefs).Error
	if err != nil {
		return nil, err
	}

	// 转换为 AI 工具格式
	tools := make([]model.Tool, 0, len(toolDefs))
	for _, def := range toolDefs {
		tool := model.Tool{
			Type: "function",
		}
		tool.Function.Name = def.Name
		tool.Function.Description = def.Description

		// 解析 parameters JSON
		if def.Parameters != "" {
			var params interface{}
			if err := json.Unmarshal([]byte(def.Parameters), &params); err == nil {
				tool.Function.Parameters = params
			}
		}

		// 打印日志查看读取的参数
		if def.Name == "add_node" {
			parametersBytes, _ := json.Marshal(tool.Function.Parameters)
			println("[DEBUG] add_node parameters:", string(parametersBytes))
		}

		tools = append(tools, tool)
	}

	return tools, nil
}

// GetAllTools 获取所有启用的工具
func (s *ToolCallService) GetAllTools() ([]model.ToolCallDef, error) {
	var tools []model.ToolCallDef
	err := database.DB.Where("checkStatus = 0").Find(&tools).Error
	return tools, err
}

// GetAllToolsBasicInfo 获取所有工具的轻量级信息（只包含 name 和 description）
func (s *ToolCallService) GetAllToolsBasicInfo() ([]map[string]string, error) {
	var toolDefs []model.ToolCallDef
	result := database.DB.Where("checkStatus = ?", 0).Find(&toolDefs)
	if result.Error != nil {
		return nil, result.Error
	}

	// 只返回 name 和 description
	tools := make([]map[string]string, 0, len(toolDefs))
	for _, def := range toolDefs {
		tools = append(tools, map[string]string{
			"name":        def.Name,
			"description": def.Description,
		})
	}

	return tools, nil
}
