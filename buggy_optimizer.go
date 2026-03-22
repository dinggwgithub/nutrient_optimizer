package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"time"
)

// BuggyOptimizer 含Bug的优化器
type BuggyOptimizer struct {
	bugType  string
	warnings []string
}

// NewBuggyOptimizer 创建含Bug的优化器
func NewBuggyOptimizer(bugType string) *BuggyOptimizer {
	return &BuggyOptimizer{
		bugType:  bugType,
		warnings: []string{},
	}
}

// GetWarnings 获取警告信息
func (o *BuggyOptimizer) GetWarnings() []string {
	return o.warnings
}

// LoadIngredientsFromJSON 从JSON文件加载食材数据
func LoadIngredientsFromJSON(filepath string) ([]Ingredient, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var result struct {
		Ingredients []Ingredient `json:"ingredients"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result.Ingredients, nil
}

// Optimize 执行优化（含Bug版本）
func (o *BuggyOptimizer) Optimize(req OptimizationRequest) (*OptimizationResult, error) {
	switch o.bugType {
	case BugTypePrecisionLoss:
		return o.optimizeWithPrecisionLoss(req)
	case BugTypeNumericalOverflow:
		return o.optimizeWithNumericalOverflow(req)
	case BugTypeConstraintViolation:
		return o.optimizeWithConstraintViolation(req)
	case BugTypeConvergenceFailure:
		return o.optimizeWithConvergenceFailure(req)
	case BugTypeResultInstability:
		return o.optimizeWithResultInstability(req)
	default:
		// 删除无bug_type无效时返回错误，强制测试者必须指定有效的Bug类型
		return nil, fmt.Errorf("无效的bug_type: %s。可选值: precision_loss, numerical_overflow, constraint_violation, convergence_failure, result_instability", o.bugType)
	}
}

// 浮点数精度丢失Bug
func (o *BuggyOptimizer) optimizeWithPrecisionLoss(req OptimizationRequest) (*OptimizationResult, error) {
	o.warnings = append(o.warnings, "启用浮点数精度丢失Bug模式")

	// Bug 1: 使用float32进行浮点运算，导致精度丢失
	var totalEnergy float32
	var totalProtein float32
	var totalFat float32
	var totalCarbs float32

	// 模拟多次迭代累加，放大精度问题
	for i := 0; i < 1000; i++ {
		for _, ing := range req.Ingredients {
			// 使用float32进行运算，故意造成精度丢失
			totalEnergy += float32(ing.Energy) * 0.01
			totalProtein += float32(ing.Protein) * 0.01
			totalFat += float32(ing.Fat) * 0.01
			totalCarbs += float32(ing.Carbs) * 0.01
		}
	}

	// Bug 2: 小数值被大数淹没
	var smallValuesAccumulator float64
	for i := 0; i < 10000; i++ {
		// 添加大量的小数值
		smallValuesAccumulator += 0.0000001
		// 然后添加一个大数
		if i%1000 == 0 {
			smallValuesAccumulator += 1000000.0
		}
	}

	// 生成结果（包含精度丢失）
	result := o.generateBasicResult(req)
	result.Nutrition.Energy = float64(totalEnergy)
	result.Nutrition.Protein = float64(totalProtein)
	result.Nutrition.Fat = float64(totalFat)
	result.Nutrition.Carbs = float64(totalCarbs)
	result.Nutrition.Calcium = smallValuesAccumulator // 这个值会显示精度问题

	result.Warnings = append(result.Warnings,
		"浮点数精度丢失：多次迭代累加导致营养素计算偏差",
		"小数值被大数淹没：钙含量计算异常",
	)

	return result, nil
}

// 数值溢出Bug
func (o *BuggyOptimizer) optimizeWithNumericalOverflow(req OptimizationRequest) (*OptimizationResult, error) {
	o.warnings = append(o.warnings, "启用数值溢出Bug模式")

	result := o.generateBasicResult(req)

	// Bug复现：模拟数值溢出问题（使用极端值表示，确保JSON可序列化）
	// Bug 1: 除零操作导致Inf -> 用float64最大值表示
	// Bug 2: 非法操作导致NaN -> 用float64最小值表示
	// Bug 3: 指数运算溢出 -> 用float64最大值表示
	_ = math.Log(-1.0) // 触发NaN计算（但不使用结果，避免JSON序列化问题）

	// 标记数值溢出Bug
	result.Error = "NUMERICAL_OVERFLOW_BUG_DETECTED"

	// 故意设置极端但可序列化的数值
	result.Nutrition.Energy = 1.7976931348623157e+308   // float64最大值，表示Inf/溢出
	result.Nutrition.Protein = -1.7976931348623157e+308 // float64最小值，表示NaN
	result.Nutrition.Fat = 1.7976931348623157e+308      // float64最大值

	result.Warnings = append(result.Warnings,
		"⚠️ 数值溢出Bug：除零操作导致正无穷结果（表示为float64最大值）",
		"⚠️ 数值异常Bug：非法对数操作导致NaN结果（表示为float64最小值）",
		"⚠️ 指数溢出Bug：计算结果超出浮点数表示范围",
		"💡 提示：Energy=最大值、Protein=最小值、Fat=最大值 是Bug的典型特征",
	)

	return result, nil
}

// 约束越界Bug
func (o *BuggyOptimizer) optimizeWithConstraintViolation(req OptimizationRequest) (*OptimizationResult, error) {
	o.warnings = append(o.warnings, "启用约束越界Bug模式")

	result := o.generateBasicResult(req)

	// Bug 1: 食材重量出现负数
	for i := range result.Ingredients {
		// 随机生成负数或超大值
		rand.Seed(time.Now().UnixNano())
		if i%3 == 0 {
			// 生成负数
			result.Ingredients[i].Amount = -rand.Float64() * 50
		} else if i%5 == 0 {
			// 生成超大值（超过500g）
			result.Ingredients[i].Amount = 500 + rand.Float64()*1000
		} else {
			// 正常值
			result.Ingredients[i].Amount = 50 + rand.Float64()*100
		}
	}

	// Bug 2: 营养素约束范围过窄，导致无可行解
	// 故意设置矛盾的约束条件
	if len(req.NutritionGoals) > 0 {
		// 设置能量目标为极低值，但蛋白质目标为极高值
		result.Nutrition.Energy = 50   // 极低能量
		result.Nutrition.Protein = 200 // 极高蛋白质
	}

	result.Warnings = append(result.Warnings,
		"约束越界：食材重量出现负数或超大值",
		"约束冲突：营养素目标设置矛盾，无可行解空间",
	)

	return result, nil
}

// 收敛失败Bug
func (o *BuggyOptimizer) optimizeWithConvergenceFailure(req OptimizationRequest) (*OptimizationResult, error) {
	o.warnings = append(o.warnings, "启用收敛失败Bug模式")

	// Bug 1: 收敛阈值设置过严
	if req.Tolerance <= 0 {
		req.Tolerance = 1e-15 // 设置极严格的收敛阈值
	}

	// Bug 2: 初始可行解构造不当
	result := &OptimizationResult{
		Ingredients: []IngredientAmount{},
		Converged:   false,
		Iterations:  req.MaxIterations,
		Error:       "求解器无法收敛：陷入局部最优或初始解不当",
	}

	// 故意返回空方案或极端不合理用量
	if len(req.Ingredients) > 0 {
		// 只使用第一个食材，用量极端
		extremeAmount := 1000.0 // 极端用量
		result.Ingredients = append(result.Ingredients, IngredientAmount{
			Ingredient: req.Ingredients[0],
			Amount:     extremeAmount,
		})

		// 计算营养值（极端值）
		result.Nutrition.Energy = req.Ingredients[0].Energy * extremeAmount / 100
		result.Nutrition.Protein = req.Ingredients[0].Protein * extremeAmount / 100
		result.Nutrition.Fat = req.Ingredients[0].Fat * extremeAmount / 100
		result.Nutrition.Carbs = req.Ingredients[0].Carbs * extremeAmount / 100
	}

	result.Warnings = append(result.Warnings,
		"收敛失败：求解器无法收敛，返回空方案或极端不合理用量",
		"参数设置问题：收敛阈值过严或初始解构造不当",
	)

	return result, nil
}

// 结果不稳定Bug
func (o *BuggyOptimizer) optimizeWithResultInstability(req OptimizationRequest) (*OptimizationResult, error) {
	o.warnings = append(o.warnings, "启用结果不稳定Bug模式")

	// Bug 1: 使用随机数影响优化结果
	rand.Seed(time.Now().UnixNano())

	result := o.generateBasicResult(req)

	// 为每个食材添加随机扰动
	for i := range result.Ingredients {
		// 添加±20%的随机扰动
		randomFactor := 0.8 + rand.Float64()*0.4
		result.Ingredients[i].Amount *= randomFactor
	}

	// Bug 2: 目标函数计算不一致
	// 同一参数多次运行结果不一致
	inconsistentCalculation := func() float64 {
		return rand.Float64() * 100
	}

	result.Nutrition.Calcium = inconsistentCalculation()
	result.Nutrition.Iron = inconsistentCalculation()
	result.Nutrition.Zinc = inconsistentCalculation()

	result.Warnings = append(result.Warnings,
		"结果不稳定：同一参数多次运行结果不一致",
		"随机性干扰：优化过程受随机因素影响",
	)

	return result, nil
}

// generateBasicResult 生成基本优化结果
func (o *BuggyOptimizer) generateBasicResult(req OptimizationRequest) *OptimizationResult {
	result := &OptimizationResult{
		Ingredients: make([]IngredientAmount, len(req.Ingredients)),
		Converged:   true,
		Iterations:  50,
	}

	// 生成基本的食材用量（平均分配）
	totalWeight := 300.0 // 假设总重量300g
	averageAmount := totalWeight / float64(len(req.Ingredients))

	for i, ing := range req.Ingredients {
		result.Ingredients[i] = IngredientAmount{
			Ingredient: ing,
			Amount:     averageAmount,
		}

		// 累加营养值
		result.Nutrition.Energy += ing.Energy * averageAmount / 100
		result.Nutrition.Protein += ing.Protein * averageAmount / 100
		result.Nutrition.Fat += ing.Fat * averageAmount / 100
		result.Nutrition.Carbs += ing.Carbs * averageAmount / 100
		result.Nutrition.Calcium += ing.Calcium * averageAmount / 100
		result.Nutrition.Iron += ing.Iron * averageAmount / 100
		result.Nutrition.Zinc += ing.Zinc * averageAmount / 100
		result.Nutrition.VitaminC += ing.VitaminC * averageAmount / 100
		result.Cost += ing.Price * averageAmount / 100
	}

	return result
}
