package controller

import (
	"chatroom-socket/internal/service"
	"github.com/gin-gonic/gin"
)

type AssistantController struct {
	Engine           *gin.Engine
	AssistantService *service.AssistantService
}

func NewAssistantController(engine *gin.Engine, assistantService *service.AssistantService) *AssistantController {
	return &AssistantController{
		Engine:           engine,
		AssistantService: assistantService,
	}
}

func (controller *AssistantController) RegisterRoutes() {

}
