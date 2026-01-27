package v1

import (
	giteaWork "github.com/aosanya/CodeValdCortex/internal/infrastructure/gitea"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RegisterWebhookRoutes registers webhook endpoints for work item integration
func RegisterWebhookRoutes(rg *gin.RouterGroup, webhookHandler *giteaWork.Handler, logger *logrus.Logger) {
	if webhookHandler == nil {
		return
	}

	work := rg.Group("/work")
	{
		work.POST("/issues", webhookHandler.HandleIssueWebhook)
		work.POST("/pull-requests", webhookHandler.HandlePullRequestWebhook)
	}

	logger.Info("Work item webhook endpoints registered")
}
