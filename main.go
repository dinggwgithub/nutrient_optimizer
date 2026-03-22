package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	_ "nutrient-optimizer-benchmark/docs" // 导入生成的docs文件夹

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title 营养配餐多目标优化算法API
// @version 1.0
// @description 营养配餐多目标优化算法测试套件，包含加权求和算法和MOEA/D算法的实现
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api

// @securityDefinitions.basic BasicAuth

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	r := gin.Default()

	// 添加Swagger UI路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 设置API路由
	api := r.Group("/api")
	{
		api.POST("/optimize-with-bugs", optimizeWithBugsHandler)
		api.POST("/optimize-moead", optimizeMOEADHandler)
		api.POST("/optimize-fixed", optimizeFixedHandler)
		api.GET("/health", healthHandler)
		api.GET("/ingredients", getIngredientsHandler)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("营养配餐优化服务启动在端口 %s", port)
	log.Printf("Swagger文档地址: http://localhost:%s/swagger/index.html", port)
	log.Fatal(r.Run(":" + port))
}

// HealthCheck 健康检查
// @Summary 健康检查
// @Description 检查服务是否正常运行
// @Tags 系统
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "服务状态"
// @Router /health [get]
func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"version": "1.0.0",
		"service": "营养配餐多目标优化算法",
		"swagger": "http://localhost:8080/swagger/index.html",
	})
}

// OptimizeRequest 优化请求参数
// @Description 营养配餐优化请求参数
// @Param ingredients body []Ingredient true "食材列表"
// @Param nutrition_goals body []NutritionGoal true "营养目标"
// @Param constraints body []Constraint true "约束条件"
// @Param weights body []Weight true "权重配置"
// @Param max_iterations body int true "最大迭代次数"
// @Param tolerance body float64 true "收敛阈值"

type OptimizeRequest struct {
	Ingredients    []Ingredient    `json:"ingredients"`
	NutritionGoals []NutritionGoal `json:"nutrition_goals"`
	Constraints    []Constraint    `json:"constraints"`
	Weights        []Weight        `json:"weights"`
	MaxIterations  int             `json:"max_iterations"`
	Tolerance      float64         `json:"tolerance"`
}

// OptimizeWithBugs 含Bug优化
// @Summary 含Bug优化（测试版本）
// @Description 使用含Bug版本的优化器进行测试，用于复现特定类型的科学计算Bug
// @Tags 优化算法
// @Accept json
// @Produce json
// @Param bug_type query string true "Bug类型" Enums(precision_loss, numerical_overflow, constraint_violation, convergence_failure, result_instability)
// @Param request body OptimizeRequest true "优化请求参数"
// @Success 200 {object} map[string]interface{} "优化结果（含Bug信息）"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /optimize-with-bugs [post]
func optimizeWithBugsHandler(c *gin.Context) {
	var req OptimizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 使用含Bug版本的优化器
	bugType := c.Query("bug_type")
	optimizer := NewBuggyOptimizer(bugType)
	result, err := optimizer.Optimize(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bug_type": bugType,
		"result":   result,
		"warnings": optimizer.GetWarnings(),
	})
}

// OptimizeMOEAD MOEA/D优化
// @Summary MOEA/D多目标优化
// @Description 使用MOEA/D（基于分解的多目标进化算法）进行营养配餐优化
// @Tags 优化算法
// @Accept json
// @Produce json
// @Param population_size query int false "种群大小 (默认: 50)" default(50)
// @Param max_iterations query int false "最大迭代次数 (默认: 100)" default(100)
// @Param request body OptimizeRequest true "优化请求参数"
// @Success 200 {object} map[string]interface{} "优化结果"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /optimize-moead [post]
func optimizeMOEADHandler(c *gin.Context) {
	var req OptimizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取参数
	populationSize := 50
	if ps := c.Query("population_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &populationSize)
	}

	maxIterations := 100
	if mi := c.Query("max_iterations"); mi != "" {
		fmt.Sscanf(mi, "%d", &maxIterations)
	}

	// 使用MOEA/D优化器
	optimizer := NewMOEADOptimizer(populationSize, maxIterations)
	defer optimizer.CloseDB()

	result, err := optimizer.Optimize(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"algorithm": "MOEA/D",
		"result":    result,
		"warnings":  optimizer.GetWarnings(),
	})
}

// GetIngredients 获取食材列表
// @Summary 获取食材列表
// @Description 从数据库或JSON文件获取可用的食材列表
// @Tags 食材管理
// @Accept json
// @Produce json
// @Param source query string false "数据源 (db或json, 默认: db)" Enums(db, json) default(db)
// @Param limit query int false "返回数量限制 (默认: 20)" default(20)
// @Param filepath query string false "JSON文件路径 (source=json时使用)" default(ingredients_db_export.json)
// @Success 200 {object} map[string]interface{} "食材列表"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /ingredients [get]
func getIngredientsHandler(c *gin.Context) {
	limit := 20
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	source := c.DefaultQuery("source", "db")
	filepath := c.DefaultQuery("filepath", "ingredients_db_export.json")

	optimizer := NewMOEADOptimizer(10, 10)
	defer optimizer.CloseDB()

	var ingredients []Ingredient
	var err error
	var sourceUsed string

	if source == "json" {
		ingredients, err = optimizer.LoadIngredientsFromJSON(filepath)
		sourceUsed = "JSON文件"
	} else {
		ingredients, err = optimizer.LoadIngredientsFromDB(limit)
		sourceUsed = "数据库"
	}

	if err != nil {
		// 如果数据库失败，尝试从JSON文件加载
		if source == "db" {
			ingredients, err = optimizer.LoadIngredientsFromJSON("ingredients_db_export.json")
			sourceUsed = "JSON文件（数据库连接失败，自动切换）"
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("从所有数据源加载失败: %v", err),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"count":       len(ingredients),
		"source":      sourceUsed,
		"ingredients": ingredients,
	})
}

// OptimizeFixed 修复后的优化
// @Summary 修复后的优化（正确版本）
// @Description 使用修复后的优化器进行营养配餐优化，解决数值计算问题
// @Tags 优化算法
// @Accept json
// @Produce json
// @Param request body OptimizeRequest true "优化请求参数"
// @Success 200 {object} map[string]interface{} "优化结果"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /optimize-fixed [post]
func optimizeFixedHandler(c *gin.Context) {
	var req OptimizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	optimizer := NewFixedOptimizer()
	result, err := optimizer.Optimize(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"algorithm": "FixedOptimizer",
		"result":    result,
		"warnings":  optimizer.GetWarnings(),
	})
}
