package handler

import (
	"net/http"

	"github.com/Tencent/WeKnora/internal/errors"
	"github.com/Tencent/WeKnora/internal/logger"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	secutils "github.com/Tencent/WeKnora/internal/utils"
	"github.com/gin-gonic/gin"
)

// GraphSchemaHandler serves KB-scoped graph schema CRUD and ingest tasks.
type GraphSchemaHandler struct {
	svc interfaces.GraphSchemaService
}

// NewGraphSchemaHandler creates a GraphSchemaHandler.
func NewGraphSchemaHandler(svc interfaces.GraphSchemaService) *GraphSchemaHandler {
	return &GraphSchemaHandler{svc: svc}
}

func (h *GraphSchemaHandler) kbID(c *gin.Context) string {
	return secutils.SanitizeForLog(c.Param("id"))
}

// ListAttributes GET /knowledge-bases/:id/graph-schema/attributes
func (h *GraphSchemaHandler) ListAttributes(c *gin.Context) {
	ctx := c.Request.Context()
	list, err := h.svc.ListAttributes(ctx, h.kbID(c))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": list})
}

// CreateAttribute POST /knowledge-bases/:id/graph-schema/attributes
func (h *GraphSchemaHandler) CreateAttribute(c *gin.Context) {
	ctx := c.Request.Context()
	var body types.KBGraphAttribute
	if err := c.ShouldBindJSON(&body); err != nil {
		c.Error(errors.NewBadRequestError("请求参数不合法").WithDetails(err.Error()))
		return
	}
	attr, err := h.svc.CreateAttribute(ctx, h.kbID(c), &body)
	if err != nil {
		logger.ErrorWithFields(ctx, err, nil)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": attr})
}

// UpdateAttribute PUT /knowledge-bases/:id/graph-schema/attributes/:attrId
func (h *GraphSchemaHandler) UpdateAttribute(c *gin.Context) {
	ctx := c.Request.Context()
	var body types.KBGraphAttribute
	if err := c.ShouldBindJSON(&body); err != nil {
		c.Error(errors.NewBadRequestError("请求参数不合法").WithDetails(err.Error()))
		return
	}
	attr, err := h.svc.UpdateAttribute(ctx, h.kbID(c), secutils.SanitizeForLog(c.Param("attrId")), &body)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": attr})
}

// DeleteAttribute DELETE /knowledge-bases/:id/graph-schema/attributes/:attrId
func (h *GraphSchemaHandler) DeleteAttribute(c *gin.Context) {
	ctx := c.Request.Context()
	if err := h.svc.DeleteAttribute(ctx, h.kbID(c), secutils.SanitizeForLog(c.Param("attrId"))); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ListEntities GET /knowledge-bases/:id/graph-schema/entities
func (h *GraphSchemaHandler) ListEntities(c *gin.Context) {
	ctx := c.Request.Context()
	list, err := h.svc.ListEntities(ctx, h.kbID(c))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": list})
}

// CreateEntity POST /knowledge-bases/:id/graph-schema/entities
func (h *GraphSchemaHandler) CreateEntity(c *gin.Context) {
	ctx := c.Request.Context()
	var body types.KBGraphEntity
	if err := c.ShouldBindJSON(&body); err != nil {
		c.Error(errors.NewBadRequestError("请求参数不合法").WithDetails(err.Error()))
		return
	}
	ent, err := h.svc.CreateEntity(ctx, h.kbID(c), &body)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": ent})
}

// UpdateEntity PUT /knowledge-bases/:id/graph-schema/entities/:entityId
func (h *GraphSchemaHandler) UpdateEntity(c *gin.Context) {
	ctx := c.Request.Context()
	var body types.KBGraphEntity
	if err := c.ShouldBindJSON(&body); err != nil {
		c.Error(errors.NewBadRequestError("请求参数不合法").WithDetails(err.Error()))
		return
	}
	ent, err := h.svc.UpdateEntity(ctx, h.kbID(c), secutils.SanitizeForLog(c.Param("entityId")), &body)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": ent})
}

// DeleteEntity DELETE /knowledge-bases/:id/graph-schema/entities/:entityId
func (h *GraphSchemaHandler) DeleteEntity(c *gin.Context) {
	ctx := c.Request.Context()
	if err := h.svc.DeleteEntity(ctx, h.kbID(c), secutils.SanitizeForLog(c.Param("entityId"))); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ListRelations GET /knowledge-bases/:id/graph-schema/relations
func (h *GraphSchemaHandler) ListRelations(c *gin.Context) {
	ctx := c.Request.Context()
	list, err := h.svc.ListRelations(ctx, h.kbID(c))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": list})
}

// CreateRelation POST /knowledge-bases/:id/graph-schema/relations
func (h *GraphSchemaHandler) CreateRelation(c *gin.Context) {
	ctx := c.Request.Context()
	var body types.KBGraphRelation
	if err := c.ShouldBindJSON(&body); err != nil {
		c.Error(errors.NewBadRequestError("请求参数不合法").WithDetails(err.Error()))
		return
	}
	rel, err := h.svc.CreateRelation(ctx, h.kbID(c), &body)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": rel})
}

// UpdateRelation PUT /knowledge-bases/:id/graph-schema/relations/:relationId
func (h *GraphSchemaHandler) UpdateRelation(c *gin.Context) {
	ctx := c.Request.Context()
	var body types.KBGraphRelation
	if err := c.ShouldBindJSON(&body); err != nil {
		c.Error(errors.NewBadRequestError("请求参数不合法").WithDetails(err.Error()))
		return
	}
	rel, err := h.svc.UpdateRelation(ctx, h.kbID(c), secutils.SanitizeForLog(c.Param("relationId")), &body)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": rel})
}

// DeleteRelation DELETE /knowledge-bases/:id/graph-schema/relations/:relationId
func (h *GraphSchemaHandler) DeleteRelation(c *gin.Context) {
	ctx := c.Request.Context()
	if err := h.svc.DeleteRelation(ctx, h.kbID(c), secutils.SanitizeForLog(c.Param("relationId"))); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ListTasks GET /knowledge-bases/:id/graph-schema/tasks
func (h *GraphSchemaHandler) ListTasks(c *gin.Context) {
	ctx := c.Request.Context()
	list, err := h.svc.ListIngestTasks(ctx, h.kbID(c))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": list})
}

// GetTask GET /knowledge-bases/:id/graph-schema/tasks/:taskId
func (h *GraphSchemaHandler) GetTask(c *gin.Context) {
	ctx := c.Request.Context()
	task, err := h.svc.GetIngestTask(ctx, h.kbID(c), secutils.SanitizeForLog(c.Param("taskId")))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": task})
}

// CreateTask POST /knowledge-bases/:id/graph-schema/tasks
func (h *GraphSchemaHandler) CreateTask(c *gin.Context) {
	ctx := c.Request.Context()
	var body types.KBGraphIngestTask
	if err := c.ShouldBindJSON(&body); err != nil {
		c.Error(errors.NewBadRequestError("请求参数不合法").WithDetails(err.Error()))
		return
	}
	task, err := h.svc.CreateIngestTask(ctx, h.kbID(c), &body)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": task})
}

// CancelTask POST /knowledge-bases/:id/graph-schema/tasks/:taskId/cancel
func (h *GraphSchemaHandler) CancelTask(c *gin.Context) {
	ctx := c.Request.Context()
	task, err := h.svc.CancelIngestTask(ctx, h.kbID(c), secutils.SanitizeForLog(c.Param("taskId")))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": task})
}

// RetryTask POST /knowledge-bases/:id/graph-schema/tasks/:taskId/retry
func (h *GraphSchemaHandler) RetryTask(c *gin.Context) {
	ctx := c.Request.Context()
	task, err := h.svc.RetryIngestTask(ctx, h.kbID(c), secutils.SanitizeForLog(c.Param("taskId")))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": task})
}
