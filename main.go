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
		api.POST("/optimize-with-bugs-fixed", optimizeWithBugsFixedHandler)
		api.POST("/ab-test-numerical-overflow", abTestNumericalOverflowHandler)
		api.POST("/optimize-moead", optimizeMOEADHandler)
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

// OptimizeWithBugsFixed 含Bug优化（修复版本）
// @Summary 含Bug优化修复版本
// @Description 使用修复后的优化器进行数值溢出Bug修复，专门处理NaN/Inf/负数/超大值等异常输出
// @Tags 优化算法
// @Accept json
// @Produce json
// @Param bug_type query string true "Bug类型" Enums(numerical_overflow)
// @Param request body OptimizeRequest true "优化请求参数"
// @Success 200 {object} map[string]interface{} "修复后的优化结果"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /optimize-with-bugs-fixed [post]
func optimizeWithBugsFixedHandler(c *gin.Context) {
	var req OptimizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 使用修复后的优化器
	bugType := c.Query("bug_type")
	optimizer := NewFixedOptimizer(bugType)
	result, err := optimizer.Optimize(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bug_type":    bugType,
		"result":      result,
		"warnings":    optimizer.GetWarnings(),
		"fixes":       optimizer.GetFixes(),
		"fixed":       true,
		"description": "数值溢出Bug已修复：NaN/Inf/负数/超大值已被处理",
	})
}

// ABTestNumericalOverflow 数值溢出A/B测试
// @Summary 数值溢出Bug A/B测试
// @Description 对比原始Bug版本和修复版本的输出差异，输出A/B测试对比数据
// @Tags 优化算法
// @Accept json
// @Produce json
// @Param request body OptimizeRequest true "优化请求参数"
// @Success 200 {object} map[string]interface{} "A/B测试对比结果"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /ab-test-numerical-overflow [post]
func abTestNumericalOverflowHandler(c *gin.Context) {
	var req OptimizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// A组：原始Bug版本
	buggyOptimizer := NewBuggyOptimizer(BugTypeNumericalOverflow)
	buggyResult, err := buggyOptimizer.Optimize(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// B组：修复版本
	fixedOptimizer := NewFixedOptimizer(BugTypeNumericalOverflow)
	fixedResult, err := fixedOptimizer.Optimize(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 生成对比报告
	comparison := generateABTestComparison(buggyResult, fixedResult, fixedOptimizer.GetFixes())

	c.JSON(http.StatusOK, gin.H{
		"test_type":        "numerical_overflow_ab_test",
		"test_description": "数值溢出Bug修复A/B测试：对比原始Bug版本和修复版本",
		"group_a_buggy": map[string]interface{}{
			"version":    "原始Bug版本",
			"result":     buggyResult,
			"warnings":   buggyOptimizer.GetWarnings(),
			"has_errors": hasNumericalErrors(buggyResult),
		},
		"group_b_fixed": map[string]interface{}{
			"version":    "修复版本",
			"result":     fixedResult,
			"warnings":   fixedOptimizer.GetWarnings(),
			"fixes":      fixedOptimizer.GetFixes(),
			"has_errors": hasNumericalErrors(fixedResult),
		},
		"comparison":   comparison,
		"improvements": calculateImprovements(buggyResult, fixedResult),
	})
}

// generateABTestComparison 生成A/B测试对比报告
func generateABTestComparison(buggy, fixed *OptimizationResult, fixes []string) map[string]interface{} {
	return map[string]interface{}{
		"nutrition_comparison": map[string]interface{}{
			"energy": map[string]interface{}{
				"buggy":  buggy.Nutrition.Energy,
				"fixed":  fixed.Nutrition.Energy,
				"status": compareValue(buggy.Nutrition.Energy, fixed.Nutrition.Energy),
			},
			"protein": map[string]interface{}{
				"buggy":  buggy.Nutrition.Protein,
				"fixed":  fixed.Nutrition.Protein,
				"status": compareValue(buggy.Nutrition.Protein, fixed.Nutrition.Protein),
			},
			"fat": map[string]interface{}{
				"buggy":  buggy.Nutrition.Fat,
				"fixed":  fixed.Nutrition.Fat,
				"status": compareValue(buggy.Nutrition.Fat, fixed.Nutrition.Fat),
			},
			"carbs": map[string]interface{}{
				"buggy":  buggy.Nutrition.Carbs,
				"fixed":  fixed.Nutrition.Carbs,
				"status": compareValue(buggy.Nutrition.Carbs, fixed.Nutrition.Carbs),
			},
		},
		"fixes_applied":    fixes,
		"total_fixes":      len(fixes),
		"buggy_has_errors": hasNumericalErrors(buggy),
		"fixed_has_errors": hasNumericalErrors(fixed),
		"fix_successful":   !hasNumericalErrors(fixed),
	}
}

// compareValue 比较两个值
func compareValue(buggy, fixed float64) string {
	if isInvalidValue(buggy) && !isInvalidValue(fixed) {
		return "已修复"
	}
	if isInvalidValue(buggy) && isInvalidValue(fixed) {
		return "修复失败"
	}
	if !isInvalidValue(buggy) && !isInvalidValue(fixed) {
		return "正常"
	}
	return "未知"
}

// isInvalidValue 检查是否为无效值
func isInvalidValue(v float64) bool {
	return v < 0 || v > 100000 || v == 1.7976931348623157e+308 || v == -1.7976931348623157e+308
}

// hasNumericalErrors 检查结果是否有数值错误
func hasNumericalErrors(result *OptimizationResult) bool {
	n := result.Nutrition
	if isInvalidValue(n.Energy) || isInvalidValue(n.Protein) || isInvalidValue(n.Fat) ||
		isInvalidValue(n.Carbs) || isInvalidValue(n.Calcium) || isInvalidValue(n.Iron) ||
		isInvalidValue(n.Zinc) || isInvalidValue(n.VitaminC) {
		return true
	}
	return false
}

// calculateImprovements 计算改进指标
func calculateImprovements(buggy, fixed *OptimizationResult) map[string]interface{} {
	return map[string]interface{}{
		"energy_normalized":   isInvalidValue(buggy.Nutrition.Energy) && !isInvalidValue(fixed.Nutrition.Energy),
		"protein_normalized":  isInvalidValue(buggy.Nutrition.Protein) && !isInvalidValue(fixed.Nutrition.Protein),
		"fat_normalized":      isInvalidValue(buggy.Nutrition.Fat) && !isInvalidValue(fixed.Nutrition.Fat),
		"all_nutrients_valid": !hasNumericalErrors(fixed),
		"cost_calculated":     fixed.Cost >= 0,
	}
}
