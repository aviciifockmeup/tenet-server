package utils

import (
	"encoding/json"
	"tenet-server/model"

	"github.com/cloudwego/hertz/pkg/app"
)

// SendSSEChunk 发送 SSE 数据块到客户端
func SendSSEChunk(c *app.RequestContext, chunk model.SSEChunk) {
	data, _ := json.Marshal(chunk)
	c.Write([]byte("data: "))
	c.Write(data)
	c.Write([]byte("\n\n"))
	c.Flush()
}
