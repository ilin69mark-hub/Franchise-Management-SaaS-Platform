package docs

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type SwaggerInfo struct {
	Swagger string `json:"swagger"`
	Info   struct {
		Description string `json:"description"`
		Title       string `json:"title"`
		Version     string `json:"version"`
	} `json:"info"`
	Host    string   `json:"host"`
	BasePath string `json:"basePath"`
	Paths   map[string]map[string]interface{} `json:"paths"`
}

func GetSwaggerSpec() SwaggerInfo {
	return SwaggerInfo{
		Swagger: "2.0",
		Info: struct {
			Description string `json:"description"`
			Title       string `json:"title"`
			Version     string `json:"version"`
		}{
			Description: "Multi-tenant SaaS platform for franchise network management",
			Title:       "Franchise SaaS API",
			Version:     "1.0.0",
		},
		Host:     "localhost:8080",
		BasePath: "/api/v1",
		Paths: map[string]map[string]interface{}{
			"/auth/register": {"post": map[string]interface{}{
				"tags":    []string{"Auth"},
				"summary": "Register user",
			}},
			"/auth/login": {"post": map[string]interface{}{
				"tags":    []string{"Auth"},
				"summary": "Login",
			}},
			"/auth/refresh": {"post": map[string]interface{}{
				"tags":    []string{"Auth"},
				"summary": "Refresh token",
			}},
			"/auth/me": {"get": map[string]interface{}{
				"tags":    []string{"Auth"},
				"summary": "Get profile",
			}},
			"/users": {
				"get": map[string]interface{}{"tags": []string{"Users"}, "summary": "List employees"},
				"post": map[string]interface{}{"tags": []string{"Users"}, "summary": "Create employee"},
			},
			"/leads": {
				"get": map[string]interface{}{"tags": []string{"CRM"}, "summary": "List leads"},
				"post": map[string]interface{}{"tags": []string{"CRM"}, "summary": "Create lead"},
			},
			"/checklists": {"get": map[string]interface{}{
				"tags":    []string{"Checklists"},
				"summary": "List checklists",
			}},
			"/goals": {
				"get": map[string]interface{}{"tags": []string{"Goals"}, "summary": "List goals"},
				"post": map[string]interface{}{"tags": []string{"Goals"}, "summary": "Set goal"},
			},
			"/schedule": {"get": map[string]interface{}{
				"tags":    []string{"Schedule"},
				"summary": "List events",
			}},
			"/notifications": {"get": map[string]interface{}{
				"tags":    []string{"Notifications"},
				"summary": "List notifications",
			}},
			"/salons": {"get": map[string]interface{}{
				"tags":    []string{"Salons"},
				"summary": "List salons",
			}},
			"/plans": {"get": map[string]interface{}{
				"tags":    []string{"Plans"},
				"summary": "List plans",
			}},
			"/stats/my": {"get": map[string]interface{}{
				"tags":    []string{"Stats"},
				"summary": "My stats",
			}},
			"/admin/stats": {"get": map[string]interface{}{
				"tags":    []string{"Admin"},
				"summary": "Dashboard stats",
			}},
			"/admin/tenants": {"get": map[string]interface{}{
				"tags":    []string{"Admin"},
				"summary": "List tenants",
			}},
		},
	}
}

func SwaggerHandler(c *gin.Context) {
	c.JSON(http.StatusOK, GetSwaggerSpec())
}