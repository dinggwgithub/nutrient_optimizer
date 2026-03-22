package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
)

// FixedOptimizer 修复后的优化器
type FixedOptimizer struct {
	warnings []string
}

// NewFixedOptimizer 创建修复后的优化器
func NewFixedOptimizer() *FixedOptimizer {
	return &FixedOptimizer{
		warnings: []string{},
	}
}

// GetWarnings 获取警告信息
func (o *FixedOptimizer) GetWarnings() []string {
	return o.warnings
}

// Optimize 执行优化（修复后版本）
func (o *FixedOptimizer) Optimize(req OptimizationRequest) (*OptimizationResult, error) {
	o.warnings = []string{}

	// 验证输入
	if len(req.Ingredients) == 0 {
		return nil, fmt.Errorf("没有提供食材数据")
	}

	// 设置默认参数
	if req.MaxIterations <= 0 {
		req.MaxIterations = 100
	}
	if req.Tolerance <= 0 {
		req.Tolerance = 0.05 // 5% 收敛阈值
	}

	// 执行加权求和优化
	result, err := o.weightedSumOptimize(req)
	if err != nil {
		return nil, err
	}

	// 验证结果
	if err := o.validateResult(result); err != nil {
		o.warnings = append(o.warnings, fmt.Sprintf("结果验证警告: %v", err))
	}

	return result, nil
}

// weightedSumOptimize 加权求和优化算法（修复后）
func (o *FixedOptimizer) weightedSumOptimize(req OptimizationRequest) (*OptimizationResult, error) {
	rand.Seed(time.Now().UnixNano())

	// 初始化食材用量（平均分配）
	targetWeight := o.getTargetWeight(req)
	numIngredients := len(req.Ingredients)
	baseAmount := targetWeight / float64(numIngredients)

	amounts := make([]float64, numIngredients)
	for i := range amounts {
		amounts[i] = baseAmount
	}

	// 简单的梯度下降优化
	bestScore := math.Inf(1)
	bestAmounts := make([]float64, numIngredients)
	copy(bestAmounts, amounts)

	for iter := 0; iter < req.MaxIterations; iter++ {
		// 计算当前解的得分
		score := o.calculateScore(amounts, req)

		if score < bestScore {
			bestScore = score
			copy(bestAmounts, amounts)
		}

		// 随机扰动
		for i := range amounts {
			perturbation := (rand.Float64() - 0.5) * 10 // ±5g 扰动
			newAmount := amounts[i] + perturbation
			// 应用约束：0-500g
			newAmount = math.Max(0, math.Min(500, newAmount))
			amounts[i] = newAmount
		}

		// 归一化到目标重量
		o.normalizeToTargetWeight(amounts, targetWeight)

		// 检查收敛
		if iter > 10 && math.Abs(score-bestScore) < req.Tolerance*bestScore {
			break
		}
	}

	// 生成结果
	result := o.generateResult(bestAmounts, req)
	result.Iterations = req.MaxIterations
	result.Converged = true

	return result, nil
}

// getTargetWeight 获取目标总重量
func (o *FixedOptimizer) getTargetWeight(req OptimizationRequest) float64 {
	for _, c := range req.Constraints {
		if c.Type == "total_weight" && c.Value > 0 {
			return c.Value
		}
	}
	// 默认300g
	return 300.0
}

// normalizeToTargetWeight 归一化到目标重量
func (o *FixedOptimizer) normalizeToTargetWeight(amounts []float64, targetWeight float64) {
	currentTotal := 0.0
	for _, amount := range amounts {
		currentTotal += amount
	}

	if currentTotal > 0 {
		ratio := targetWeight / currentTotal
		for i := range amounts {
			amounts[i] *= ratio
			// 确保在约束范围内
			amounts[i] = math.Max(0, math.Min(500, amounts[i]))
		}
	}
}

// calculateScore 计算加权得分（越低越好）
func (o *FixedOptimizer) calculateScore(amounts []float64, req OptimizationRequest) float64 {
	// 计算营养
	nutrition := o.calculateNutrition(amounts, req.Ingredients)

	// 计算营养偏差
	nutritionScore := o.calculateNutritionDeviation(nutrition, req.NutritionGoals)

	// 计算成本
	costScore := o.calculateTotalCost(amounts, req.Ingredients)

	// 计算多样性
	diversityScore := 1.0 - o.calculateDiversity(amounts)

	// 获取权重
	nutritionWeight, costWeight, varietyWeight := o.getWeights(req.Weights)

	// 加权求和
	return nutritionScore*nutritionWeight + costScore*costWeight + diversityScore*varietyWeight
}

// calculateNutrition 计算营养汇总（使用float64确保精度）
func (o *FixedOptimizer) calculateNutrition(amounts []float64, ingredients []Ingredient) NutritionSummary {
	var nutrition NutritionSummary

	for i, amount := range amounts {
		if i >= len(ingredients) {
			continue
		}
		ing := ingredients[i]
		// 使用float64进行精确计算
		factor := amount / 100.0 // 转换为每100g单位

		nutrition.Energy += ing.Energy * factor
		nutrition.Protein += ing.Protein * factor
		nutrition.Fat += ing.Fat * factor
		nutrition.Carbs += ing.Carbs * factor
		nutrition.Calcium += ing.Calcium * factor
		nutrition.Iron += ing.Iron * factor
		nutrition.Zinc += ing.Zinc * factor
		nutrition.VitaminC += ing.VitaminC * factor
	}

	return nutrition
}

// calculateNutritionDeviation 计算营养目标偏差
func (o *FixedOptimizer) calculateNutritionDeviation(nutrition NutritionSummary, goals []NutritionGoal) float64 {
	if len(goals) == 0 {
		return 0.0
	}

	deviation := 0.0
	totalWeight := 0.0

	for _, goal := range goals {
		var actual float64
		switch goal.Nutrient {
		case "energy":
			actual = nutrition.Energy
		case "protein":
			actual = nutrition.Protein
		case "fat":
			actual = nutrition.Fat
		case "carbs":
			actual = nutrition.Carbs
		case "calcium":
			actual = nutrition.Calcium
		case "iron":
			actual = nutrition.Iron
		case "zinc":
			actual = nutrition.Zinc
		case "vitamin_c":
			actual = nutrition.VitaminC
		default:
			continue
		}

		// 计算相对偏差
		if goal.Target > 0 {
			relDiff := math.Abs(actual-goal.Target) / goal.Target
			weight := goal.Weight
			if weight <= 0 {
				weight = 1.0
			}
			deviation += relDiff * weight
			totalWeight += weight
		}
	}

	if totalWeight > 0 {
		deviation /= totalWeight
	}

	return deviation
}

// calculateTotalCost 计算总成本
func (o *FixedOptimizer) calculateTotalCost(amounts []float64, ingredients []Ingredient) float64 {
	totalCost := 0.0
	for i, amount := range amounts {
		if i >= len(ingredients) {
			continue
		}
		// 价格是元/100g
		totalCost += ingredients[i].Price * amount / 100.0
	}
	return totalCost
}

// calculateDiversity 计算食材多样性 (Simpson指数)
func (o *FixedOptimizer) calculateDiversity(amounts []float64) float64 {
	total := 0.0
	for _, amount := range amounts {
		total += amount
	}

	if total == 0 {
		return 0.0
	}

	sumProportions := 0.0
	for _, amount := range amounts {
		proportion := amount / total
		sumProportions += proportion * proportion
	}

	// Simpson多样性指数: 1 - Σ(p_i²)
	return 1.0 - sumProportions
}

// getWeights 获取权重配置
func (o *FixedOptimizer) getWeights(weights []Weight) (nutritionWeight, costWeight, varietyWeight float64) {
	nutritionWeight = 0.6
	costWeight = 0.3
	varietyWeight = 0.1

	for _, w := range weights {
		switch w.Type {
		case "nutrition":
			nutritionWeight = w.Value
		case "cost":
			costWeight = w.Value
		case "variety":
			varietyWeight = w.Value
		}
	}

	// 归一化
	total := nutritionWeight + costWeight + varietyWeight
	if total > 0 {
		nutritionWeight /= total
		costWeight /= total
		varietyWeight /= total
	}

	return
}

// generateResult 生成优化结果
func (o *FixedOptimizer) generateResult(amounts []float64, req OptimizationRequest) *OptimizationResult {
	result := &OptimizationResult{
		Ingredients: make([]IngredientAmount, 0, len(req.Ingredients)),
		Converged:   true,
		Warnings:    make([]string, 0),
	}

	// 构建食材用量列表（过滤用量接近0的食材）
	for i, amount := range amounts {
		if i >= len(req.Ingredients) {
			break
		}
		if amount > 0.1 { // 只保留用量大于0.1g的食材
			result.Ingredients = append(result.Ingredients, IngredientAmount{
				Ingredient: req.Ingredients[i],
				Amount:     roundToTwoDecimalsFixed(amount),
			})
		}
	}

	// 按用量排序
	sort.Slice(result.Ingredients, func(i, j int) bool {
		return result.Ingredients[i].Amount > result.Ingredients[j].Amount
	})

	// 计算营养汇总
	result.Nutrition = o.calculateNutrition(amounts, req.Ingredients)

	// 计算成本
	result.Cost = roundToTwoDecimalsFixed(o.calculateTotalCost(amounts, req.Ingredients))

	// 添加多样性警告
	diversity := o.calculateDiversity(amounts)
	if diversity < 0.3 {
		result.Warnings = append(result.Warnings, "食材多样性较低，建议增加食材种类")
	}

	return result
}

// validateResult 验证结果有效性
func (o *FixedOptimizer) validateResult(result *OptimizationResult) error {
	// 检查NaN/Inf
	if math.IsNaN(result.Nutrition.Energy) || math.IsInf(result.Nutrition.Energy, 0) {
		return fmt.Errorf("能量计算结果异常: %v", result.Nutrition.Energy)
	}
	if math.IsNaN(result.Nutrition.Calcium) || math.IsInf(result.Nutrition.Calcium, 0) {
		return fmt.Errorf("钙含量计算结果异常: %v", result.Nutrition.Calcium)
	}

	// 检查负数
	if result.Nutrition.Energy < 0 {
		return fmt.Errorf("能量不能为负数: %.2f", result.Nutrition.Energy)
	}
	if result.Nutrition.Calcium < 0 {
		return fmt.Errorf("钙含量不能为负数: %.2f", result.Nutrition.Calcium)
	}

	// 检查超大值（超过合理范围100倍）
	if result.Nutrition.Energy > 100000 {
		return fmt.Errorf("能量值异常过大: %.2f", result.Nutrition.Energy)
	}
	if result.Nutrition.Calcium > 100000 {
		return fmt.Errorf("钙含量异常过大: %.2f", result.Nutrition.Calcium)
	}

	// 检查食材重量约束
	for _, ing := range result.Ingredients {
		if ing.Amount < 0 {
			return fmt.Errorf("食材重量不能为负数: %s = %.2fg", ing.Name, ing.Amount)
		}
		if ing.Amount > 500 {
			return fmt.Errorf("食材重量超过500g限制: %s = %.2fg", ing.Name, ing.Amount)
		}
	}

	return nil
}

// roundToTwoDecimalsFixed 四舍五入到两位小数
func roundToTwoDecimalsFixed(x float64) float64 {
	return math.Round(x*100) / 100
}
