package v1

import (
	"github.com/aosanya/CodeValdCortex/internal/agency/arangodb"
	"github.com/aosanya/CodeValdCortex/internal/handlers"
	driver "github.com/arangodb/go-driver"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RegisterWorkItemsRoutes registers work items API endpoints
func RegisterWorkItemsRoutes(rg *gin.RouterGroup, db driver.Database, logger *logrus.Logger) error {
	workItemRepo, err := arangodb.NewWorkItemRepository(db)
	if err != nil {
		return err
	}

	workItemHandler := handlers.NewWorkItemsHandler(workItemRepo, logger)

	workItems := rg.Group("/agencies/:agencyID/work-items")
	{
		workItems.GET("", workItemHandler.ListWorkItems)
		workItems.POST("", workItemHandler.CreateWorkItem)
		workItems.GET("/:workItemID", workItemHandler.GetWorkItem)
		workItems.PUT("/:workItemID", workItemHandler.UpdateWorkItem)
		workItems.DELETE("/:workItemID", workItemHandler.DeleteWorkItem)
	}

	logger.Info("Work items REST API endpoints registered")
	return nil
}
