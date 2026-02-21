package service

import (
	"tenet-server/database"
	"tenet-server/model"
)

// AgentService Agent服务
type AgentService struct{}

// NewAgentService 创建Agent服务
func NewAgentService() *AgentService {
	return &AgentService{}
}

// GetAgentByID 根据agentId查询Agent配置
func (s *AgentService) GetAgentByID(agentID string) (*model.Agent, error) {
	var agent model.Agent
	err := database.DB.Where("agentId = ? AND checkStatus = 0", agentID).First(&agent).Error
	if err != nil {
		return nil, err
	}
	return &agent, nil
}
