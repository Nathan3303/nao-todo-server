// TASK-26 / T130 路由契约：普通清单偏好复用既有 REST（复数 /projects），移动端共用语义不变。
//
// 覆盖 qa 用例 CT-02（复数路径）/ CT-09（REST 语义不变）。
package routers

import (
	"strings"
	"testing"

	"naotodoserver/interfaces/controllers"

	"github.com/gin-gonic/gin"
)

// TestProjectPreferenceRoutes_PluralAndUnchanged CT-02 / CT-09：
// GET/POST /projects/:projectId/preference 必须存在；不得出现单数 /project/ 路径。
func TestProjectPreferenceRoutes_PluralAndUnchanged(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	// 路由注册不触发控制器与中间件逻辑（app / auth 可为 nil）
	UseProjectRouter(engine.Group("/api/v1"), controllers.NewProjectController(nil), nil)

	routes := make(map[string]bool, len(engine.Routes()))
	for _, route := range engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}

	for _, want := range []string{
		"GET /api/v1/projects/:projectId/preference",
		"POST /api/v1/projects/:projectId/preference",
	} {
		if !routes[want] {
			t.Fatalf("缺少普通清单偏好路由 %q（REST 语义变更）", want)
		}
	}

	for route := range routes {
		if strings.Contains(route, "/project/") {
			t.Fatalf("出现单数清单路径 %q（应为复数 /projects）", route)
		}
	}
}
