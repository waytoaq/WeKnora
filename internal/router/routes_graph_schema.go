package router

import (
	"net/http"

	"github.com/Tencent/WeKnora/internal/handler"
	"github.com/gin-gonic/gin"
)

// RegisterGraphSchemaRoutes registers KB-scoped graph schema CRUD and ingest tasks.
// Write routes mirror knowledge-base edit permissions (OwnedKB + KBAccessWrite).
func RegisterGraphSchemaRoutes(r *gin.RouterGroup, h *handler.GraphSchemaHandler, g *rbacGuards) {
	if h == nil {
		return
	}
	base := "/knowledge-bases/:id/graph-schema"

	// Attributes
	g.apiKeyRoute(r, http.MethodGet, base+"/attributes",
		apiKeyRetrieve(apiKeyFullAccess()), g.Viewer(), g.KBAccessRead("id"), h.ListAttributes)
	g.apiKeyRoute(r, http.MethodPost, base+"/attributes",
		apiKeyManageKnowledgeBases(apiKeyFullAccess()), g.OwnedKBOrAdmin(), g.KBAccessWrite("id"), h.CreateAttribute)
	g.apiKeyRoute(r, http.MethodPut, base+"/attributes/:attrId",
		apiKeyManageKnowledgeBases(apiKeyFullAccess()), g.OwnedKBOrAdmin(), g.KBAccessWrite("id"), h.UpdateAttribute)
	g.apiKeyRoute(r, http.MethodDelete, base+"/attributes/:attrId",
		apiKeyManageKnowledgeBases(apiKeyFullAccess()), g.OwnedKBOrAdmin(), g.KBAccessWrite("id"), h.DeleteAttribute)

	// Entities
	g.apiKeyRoute(r, http.MethodGet, base+"/entities",
		apiKeyRetrieve(apiKeyFullAccess()), g.Viewer(), g.KBAccessRead("id"), h.ListEntities)
	g.apiKeyRoute(r, http.MethodPost, base+"/entities",
		apiKeyManageKnowledgeBases(apiKeyFullAccess()), g.OwnedKBOrAdmin(), g.KBAccessWrite("id"), h.CreateEntity)
	g.apiKeyRoute(r, http.MethodPut, base+"/entities/:entityId",
		apiKeyManageKnowledgeBases(apiKeyFullAccess()), g.OwnedKBOrAdmin(), g.KBAccessWrite("id"), h.UpdateEntity)
	g.apiKeyRoute(r, http.MethodDelete, base+"/entities/:entityId",
		apiKeyManageKnowledgeBases(apiKeyFullAccess()), g.OwnedKBOrAdmin(), g.KBAccessWrite("id"), h.DeleteEntity)

	// Relations
	g.apiKeyRoute(r, http.MethodGet, base+"/relations",
		apiKeyRetrieve(apiKeyFullAccess()), g.Viewer(), g.KBAccessRead("id"), h.ListRelations)
	g.apiKeyRoute(r, http.MethodPost, base+"/relations",
		apiKeyManageKnowledgeBases(apiKeyFullAccess()), g.OwnedKBOrAdmin(), g.KBAccessWrite("id"), h.CreateRelation)
	g.apiKeyRoute(r, http.MethodPut, base+"/relations/:relationId",
		apiKeyManageKnowledgeBases(apiKeyFullAccess()), g.OwnedKBOrAdmin(), g.KBAccessWrite("id"), h.UpdateRelation)
	g.apiKeyRoute(r, http.MethodDelete, base+"/relations/:relationId",
		apiKeyManageKnowledgeBases(apiKeyFullAccess()), g.OwnedKBOrAdmin(), g.KBAccessWrite("id"), h.DeleteRelation)

	// Ingest tasks
	g.apiKeyRoute(r, http.MethodGet, base+"/tasks",
		apiKeyRetrieve(apiKeyFullAccess()), g.Viewer(), g.KBAccessRead("id"), h.ListTasks)
	g.apiKeyRoute(r, http.MethodPost, base+"/tasks",
		apiKeyManageKnowledgeBases(apiKeyFullAccess()), g.OwnedKBOrAdmin(), g.KBAccessWrite("id"), h.CreateTask)
	g.apiKeyRoute(r, http.MethodGet, base+"/tasks/:taskId",
		apiKeyRetrieve(apiKeyFullAccess()), g.Viewer(), g.KBAccessRead("id"), h.GetTask)
	g.apiKeyRoute(r, http.MethodPost, base+"/tasks/:taskId/cancel",
		apiKeyManageKnowledgeBases(apiKeyFullAccess()), g.OwnedKBOrAdmin(), g.KBAccessWrite("id"), h.CancelTask)
	g.apiKeyRoute(r, http.MethodPost, base+"/tasks/:taskId/retry",
		apiKeyManageKnowledgeBases(apiKeyFullAccess()), g.OwnedKBOrAdmin(), g.KBAccessWrite("id"), h.RetryTask)
}
