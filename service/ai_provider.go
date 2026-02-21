package service

import (
    "context"
    "tenet-server/model"

    "github.com/cloudwego/hertz/pkg/app"
)

// AIProvider AI提供商接口（通用）
type AIProvider interface {
    // StreamChat 流式对话
    StreamChat(ctx context.Context, request *model.AIStreamRequest, c *app.RequestContext) error
}

// AIService AI服务（根据provider分发）
type AIService struct {
    providers map[string]AIProvider
}

// NewAIService 创建AI服务
func NewAIService() *AIService {
    return &AIService{
        providers: make(map[string]AIProvider),
    }
}

// RegisterProvider 注册provider
func (s *AIService) RegisterProvider(name string, provider AIProvider) {
    s.providers[name] = provider
}

// StreamChat 统一入口
func (s *AIService) StreamChat(ctx context.Context, request *model.AIStreamRequest, c *app.RequestContext) error {
    provider, ok := s.providers[request.Config.Provider]
    if !ok {
        return sendSSEError(c, "不支持的provider: "+request.Config.Provider)
    }
    
    return provider.StreamChat(ctx, request, c)
}