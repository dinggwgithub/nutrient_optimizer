package main

// OptimizationRequest 优化请求结构体
type OptimizationRequest struct {
	Ingredients    []Ingredient     `json:"ingredients"`
	NutritionGoals []NutritionGoal  `json:"nutrition_goals"`
	Constraints    []Constraint     `json:"constraints"`
	Weights        []Weight         `json:"weights"`
	MaxIterations  int              `json:"max_iterations"`
	Tolerance      float64          `json:"tolerance"`
}

// Ingredient 食材信息
// @Description 食材的营养成分信息
type Ingredient struct {
	ID       int     `json:"id" example:"1"`
	Name     string  `json:"name" example:"鸡胸肉"`
	Energy   float64 `json:"energy" example:"165.0"`        // 能量 (kcal/100g)
	Protein  float64 `json:"protein" example:"31.0"`       // 蛋白质 (g/100g)
	Fat      float64 `json:"fat" example:"3.6"`           // 脂肪 (g/100g)
	Carbs    float64 `json:"carbs" example:"0.0"`         // 碳水化合物 (g/100g)
	Calcium  float64 `json:"calcium" example:"11.0"`       // 钙 (mg/100g)
	Iron     float64 `json:"iron" example:"1.0"`          // 铁 (mg/100g)
	Zinc     float64 `json:"zinc" example:"0.7"`          // 锌 (mg/100g)
	VitaminC float64 `json:"vitamin_c" example:"0.0"`     // 维生素C (mg/100g)
	Price    float64 `json:"price" example:"0.8"`         // 价格 (元/100g)
}

// NutritionGoal 营养目标
// @Description 营养素的优化目标
type NutritionGoal struct {
	Nutrient string  `json:"nutrient" example:"energy"`     // 营养素名称
	Target   float64 `json:"target" example:"600.0"`       // 目标值
	Min      float64 `json:"min" example:"500.0"`          // 最小值
	Max      float64 `json:"max" example:"700.0"`          // 最大值
	Weight   float64 `json:"weight" example:"0.3"`       // 权重
}

// Constraint 约束条件
// @Description 优化过程中的约束条件
type Constraint struct {
	Type      string  `json:"type" example:"ingredient_min"`        // "ingredient_min", "ingredient_max", "total_weight"
	IngredientID int  `json:"ingredient_id,omitempty" example:"1"`
	Value     float64 `json:"value" example:"50.0"`
}

// Weight 权重配置
// @Description 多目标优化的权重配置
type Weight struct {
	Type   string  `json:"type" example:"nutrition"`           // "nutrition", "cost", "variety"
	Value  float64 `json:"value" example:"0.6"`
}

// OptimizationResult 优化结果
// @Description 营养配餐优化结果
type OptimizationResult struct {
	Ingredients []IngredientAmount `json:"ingredients"`
	Nutrition   NutritionSummary   `json:"nutrition"`
	Cost        float64            `json:"cost" example:"15.5"`
	Iterations  int                `json:"iterations" example:"50"`
	Converged   bool               `json:"converged" example:"true"`
	Error       string             `json:"error,omitempty"`
	Warnings    []string           `json:"warnings,omitempty"`
}

// IngredientAmount 食材用量
// @Description 优化后的食材用量
type IngredientAmount struct {
	Ingredient
	Amount float64 `json:"amount" example:"150.5"`        // 用量 (g)
}

// NutritionSummary 营养汇总
// @Description 优化后的营养汇总信息
type NutritionSummary struct {
	Energy   float64 `json:"energy" example:"600.5"`
	Protein  float64 `json:"protein" example:"30.2"`
	Fat      float64 `json:"fat" example:"15.8"`
	Carbs    float64 `json:"carbs" example:"80.3"`
	Calcium  float64 `json:"calcium" example:"300.5"`
	Iron     float64 `json:"iron" example:"8.2"`
	Zinc     float64 `json:"zinc" example:"5.1"`
	VitaminC float64 `json:"vitamin_c" example:"45.6"`
}

// BugType Bug类型常量
const (
	BugTypePrecisionLoss   = "precision_loss"    // 浮点数精度丢失
	BugTypeNumericalOverflow = "numerical_overflow" // 数值溢出
	BugTypeConstraintViolation = "constraint_violation" // 约束越界
	BugTypeConvergenceFailure = "convergence_failure" // 收敛失败
	BugTypeResultInstability = "result_instability" // 结果不稳定
)