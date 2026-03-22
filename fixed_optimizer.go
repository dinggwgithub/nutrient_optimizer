package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"sort"
	"time"
)

// FixedOptimizer 修复后的优化器
type FixedOptimizer struct {
	warnings       []string
	useMOEAD       bool
	populationSize int
	maxIterations  int
	neighborSize   int
	weightVectors  [][]float64
	neighborhood   [][]int
	idealPoint     []float64
	objectives     int
}

// NewFixedOptimizer 创建修复后的优化器
func NewFixedOptimizer(useMOEAD bool) *FixedOptimizer {
	return &FixedOptimizer{
		useMOEAD:       useMOEAD,
		populationSize: 50,
		maxIterations:  100,
		neighborSize:   10,
		objectives:     3,
		warnings:       []string{},
	}
}

// GetWarnings 获取警告信息
func (o *FixedOptimizer) GetWarnings() []string {
	return o.warnings
}

// Optimize 执行优化（修复版本）
func (o *FixedOptimizer) Optimize(req OptimizationRequest) (*OptimizationResult, error) {
	o.warnings = []string{}

	// 验证输入
	if err := o.validateRequest(req); err != nil {
		return nil, err
	}

	ingredients := req.Ingredients
	if len(ingredients) == 0 {
		return nil, fmt.Errorf("食材列表不能为空")
	}

	if o.useMOEAD {
		return o.optimizeMOEAD(req, ingredients)
	}
	return o.optimizeWeightedSum(req, ingredients)
}

// validateRequest 验证请求参数
func (o *FixedOptimizer) validateRequest(req OptimizationRequest) error {
	for i, ing := range req.Ingredients {
		if ing.Energy < 0 {
			return fmt.Errorf("食材%d能量不能为负数", i+1)
		}
		if ing.Protein < 0 {
			return fmt.Errorf("食材%d蛋白质不能为负数", i+1)
		}
		if ing.Fat < 0 {
			return fmt.Errorf("食材%d脂肪不能为负数", i+1)
		}
		if ing.Carbs < 0 {
			return fmt.Errorf("食材%d碳水不能为负数", i+1)
		}
		if ing.Calcium < 0 {
			return fmt.Errorf("食材%d钙不能为负数", i+1)
		}
		if ing.Price < 0 {
			return fmt.Errorf("食材%d价格不能为负数", i+1)
		}
	}
	return nil
}

// optimizeWeightedSum 加权求和优化（修复版本 - 无MOEA/D）
func (o *FixedOptimizer) optimizeWeightedSum(req OptimizationRequest, ingredients []Ingredient) (*OptimizationResult, error) {
	result := &OptimizationResult{
		Ingredients: make([]IngredientAmount, len(ingredients)),
		Converged:   true,
		Iterations:  1,
	}

	// 计算总重量约束
	targetWeight := o.getTotalWeightConstraint(req)
	numIngredients := len(ingredients)

	// 平均分配用量作为初始解
	averageAmount := targetWeight / float64(numIngredients)

	for i, ing := range ingredients {
		amount := o.clampAmount(averageAmount, req, i+1)
		result.Ingredients[i] = IngredientAmount{
			Ingredient: ing,
			Amount:     roundToTwoDecimalsFixed(amount),
		}
	}

	// 调整用量以满足总重量约束
	o.adjustAmountsToTargetWeight(result.Ingredients, targetWeight, req)

	// 计算营养值
	result.Nutrition = o.calculateNutrition(result.Ingredients)
	result.Cost = o.calculateCost(result.Ingredients)

	// 验证并修复结果
	o.validateAndFixResult(result)

	return result, nil
}

// optimizeMOEAD MOEA/D多目标优化（修复版本）
func (o *FixedOptimizer) optimizeMOEAD(req OptimizationRequest, ingredients []Ingredient) (*OptimizationResult, error) {
	rand.Seed(time.Now().UnixNano())

	// 初始化MOEA/D组件
	o.generateWeightVectors()
	o.calculateNeighborhood()
	o.initializeIdealPoint()

	// 初始化种群
	population := make([]*Individual, o.populationSize)
	for i := range population {
		population[i] = o.createIndividual(len(ingredients), req)
		o.evaluateIndividual(population[i], ingredients, req)
	}

	// 主迭代循环
	convergenceCount := 0
	bestObjective := math.Inf(1)

	for gen := 0; gen < o.maxIterations; gen++ {
		for i := 0; i < o.populationSize; i++ {
			// 从邻域中选择两个父代
			neighbors := o.neighborhood[i]
			p1 := neighbors[rand.Intn(len(neighbors))]
			p2 := neighbors[rand.Intn(len(neighbors))]

			// 交叉生成子代
			child := o.crossover(population[p1], population[p2])

			// 变异
			o.mutate(child)

			// 修复约束
			o.repair(child, req)

			// 评估子代
			o.evaluateIndividual(child, ingredients, req)

			// 更新邻域解
			for _, neighborIdx := range neighbors {
				f1 := o.tchebycheff(o.weightVectors[neighborIdx], population[neighborIdx].Objectives)
				f2 := o.tchebycheff(o.weightVectors[neighborIdx], child.Objectives)

				if f2 < f1 {
					population[neighborIdx] = child
				}
			}
		}

		// 检查收敛性
		currentBest := o.getCurrentBestObjective(population)
		if math.Abs(currentBest-bestObjective) < 1e-6 {
			convergenceCount++
			if convergenceCount >= 10 {
				break
			}
		} else {
			convergenceCount = 0
			bestObjective = currentBest
		}
	}

	// 找到最优解（根据权重偏好选择）
	bestIdx := o.selectBestSolution(population, req.Weights)
	bestInd := population[bestIdx]

	// 生成结果
	result := o.generateResult(bestInd, ingredients, req)
	result.Iterations = o.maxIterations
	result.Converged = true

	// 验证并修复结果
	o.validateAndFixResult(result)

	return result, nil
}

// getCurrentBestObjective 获取当前最优目标值
func (o *FixedOptimizer) getCurrentBestObjective(population []*Individual) float64 {
	best := math.Inf(1)
	for _, ind := range population {
		if ind.Objectives[0] < best {
			best = ind.Objectives[0]
		}
	}
	return best
}

// calculateNutrition 计算营养汇总（修复版本）
func (o *FixedOptimizer) calculateNutrition(ingredients []IngredientAmount) NutritionSummary {
	var nutrition NutritionSummary

	for _, ia := range ingredients {
		// 使用正确的单位转换：用量(g) / 100 * 每100g含量
		factor := ia.Amount / 100.0

		// 使用float64进行精确计算
		nutrition.Energy = safeAdd(nutrition.Energy, ia.Energy*factor)
		nutrition.Protein = safeAdd(nutrition.Protein, ia.Protein*factor)
		nutrition.Fat = safeAdd(nutrition.Fat, ia.Fat*factor)
		nutrition.Carbs = safeAdd(nutrition.Carbs, ia.Carbs*factor)
		nutrition.Calcium = safeAdd(nutrition.Calcium, ia.Calcium*factor)
		nutrition.Iron = safeAdd(nutrition.Iron, ia.Iron*factor)
		nutrition.Zinc = safeAdd(nutrition.Zinc, ia.Zinc*factor)
		nutrition.VitaminC = safeAdd(nutrition.VitaminC, ia.VitaminC*factor)
	}

	return nutrition
}

// calculateCost 计算总成本
func (o *FixedOptimizer) calculateCost(ingredients []IngredientAmount) float64 {
	total := 0.0
	for _, ia := range ingredients {
		total = safeAdd(total, ia.Price*ia.Amount/100.0)
	}
	return roundToTwoDecimalsFixed(total)
}

// adjustAmountsToTargetWeight 调整用量到目标总重量
func (o *FixedOptimizer) adjustAmountsToTargetWeight(ingredients []IngredientAmount, targetWeight float64, req OptimizationRequest) {
	total := 0.0
	for _, ia := range ingredients {
		total += ia.Amount
	}

	// 只有当总重量显著偏离时才调整
	if math.Abs(total-targetWeight) > 1e-6 && total > 0 {
		ratio := targetWeight / total
		for i := range ingredients {
			newAmount := ingredients[i].Amount * ratio
			ingredients[i].Amount = o.clampAmount(newAmount, req, i+1)
		}

		// 二次调整以精确达到目标
		total = 0.0
		for _, ia := range ingredients {
			total += ia.Amount
		}

		// 对未受约束的食材进行微调
		diff := targetWeight - total
		if math.Abs(diff) > 1e-6 {
			unconstrainedCount := o.countUnconstrainedIngredients(len(ingredients), req)
			if unconstrainedCount > 0 {
				adjustmentPer := diff / float64(unconstrainedCount)
				for i := range ingredients {
					if !o.isIngredientConstrained(i+1, req) {
						newAmount := ingredients[i].Amount + adjustmentPer
						ingredients[i].Amount = o.clampAmount(newAmount, req, i+1)
					}
				}
			}
		}
	}

	// 确保所有值都是合理的
	for i := range ingredients {
		ingredients[i].Amount = roundToTwoDecimals(ingredients[i].Amount)
	}
}

// isIngredientConstrained 检查食材是否有约束
func (o *FixedOptimizer) isIngredientConstrained(ingredientID int, req OptimizationRequest) bool {
	for _, c := range req.Constraints {
		if c.IngredientID == ingredientID {
			return true
		}
	}
	return false
}

// countUnconstrainedIngredients 计算未受约束的食材数量
func (o *FixedOptimizer) countUnconstrainedIngredients(total int, req OptimizationRequest) int {
	constrained := make(map[int]bool)
	for _, c := range req.Constraints {
		if c.IngredientID > 0 {
			constrained[c.IngredientID] = true
		}
	}
	return total - len(constrained)
}

// clampAmount 限制用量在合理范围（0-500g）
func (o *FixedOptimizer) clampAmount(amount float64, req OptimizationRequest, ingredientID int) float64 {
	// 硬约束：0-500g
	minAmount := 0.0
	maxAmount := 500.0

	// 检查用户定义的约束
	for _, c := range req.Constraints {
		if c.IngredientID == ingredientID {
			switch c.Type {
			case "ingredient_min":
				minAmount = math.Max(minAmount, c.Value)
			case "ingredient_max":
				maxAmount = math.Min(maxAmount, c.Value)
			}
		}
	}

	// 确保范围有效
	if minAmount > maxAmount {
		minAmount = maxAmount // 避免矛盾约束
	}

	// 确保不超出硬约束
	minAmount = math.Max(0, minAmount)
	maxAmount = math.Min(500, maxAmount)

	// 应用约束
	result := math.Max(minAmount, math.Min(maxAmount, amount))

	// 避免负值
	return math.Max(0, result)
}

// getTotalWeightConstraint 获取总重量约束
func (o *FixedOptimizer) getTotalWeightConstraint(req OptimizationRequest) float64 {
	for _, c := range req.Constraints {
		if c.Type == "total_weight" {
			return math.Max(10, math.Min(2000, c.Value)) // 合理范围：10g-2000g
		}
	}
	return 300.0 // 默认值
}

// validateAndFixResult 验证并修复结果
func (o *FixedOptimizer) validateAndFixResult(result *OptimizationResult) {
	var warnings []string

	// 验证营养值
	nutrition := &result.Nutrition
	nutrition.Energy = fixValueFixed(nutrition.Energy, &warnings, "能量")
	nutrition.Protein = fixValueFixed(nutrition.Protein, &warnings, "蛋白质")
	nutrition.Fat = fixValueFixed(nutrition.Fat, &warnings, "脂肪")
	nutrition.Carbs = fixValueFixed(nutrition.Carbs, &warnings, "碳水")
	nutrition.Calcium = fixValueFixed(nutrition.Calcium, &warnings, "钙")
	nutrition.Iron = fixValueFixed(nutrition.Iron, &warnings, "铁")
	nutrition.Zinc = fixValueFixed(nutrition.Zinc, &warnings, "锌")
	nutrition.VitaminC = fixValueFixed(nutrition.VitaminC, &warnings, "维生素C")

	// 验证食材用量
	for i := range result.Ingredients {
		amount := &result.Ingredients[i].Amount
		*amount = fixAmountFixed(*amount, &warnings, result.Ingredients[i].Name)
	}

	// 验证成本
	result.Cost = fixValueFixed(result.Cost, &warnings, "成本")

	// 添加警告
	if len(warnings) > 0 {
		result.Warnings = append(result.Warnings, warnings...)
	}

	// 四舍五入到合理精度
	roundNutritionSummaryFixed(nutrition)
}

// MOEA/D相关方法（修复版本）

func (o *FixedOptimizer) generateWeightVectors() {
	o.weightVectors = make([][]float64, o.populationSize)

	for i := 0; i < o.populationSize; i++ {
		vector := make([]float64, o.objectives)
		sum := 0.0

		for j := 0; j < o.objectives; j++ {
			vector[j] = rand.Float64() + 0.001
			sum += vector[j]
		}

		for j := 0; j < o.objectives; j++ {
			vector[j] /= sum
		}

		o.weightVectors[i] = vector
	}
}

func (o *FixedOptimizer) calculateNeighborhood() {
	o.neighborhood = make([][]int, o.populationSize)

	for i := 0; i < o.populationSize; i++ {
		distances := make([]struct {
			index    int
			distance float64
		}, o.populationSize)

		for j := 0; j < o.populationSize; j++ {
			dist := o.euclideanDistance(o.weightVectors[i], o.weightVectors[j])
			distances[j] = struct {
				index    int
				distance float64
			}{j, dist}
		}

		sort.Slice(distances, func(a, b int) bool {
			return distances[a].distance < distances[b].distance
		})

		o.neighborhood[i] = make([]int, o.neighborSize)
		for k := 0; k < o.neighborSize; k++ {
			o.neighborhood[i][k] = distances[k].index
		}
	}
}

func (o *FixedOptimizer) euclideanDistance(a, b []float64) float64 {
	sum := 0.0
	for i := range a {
		diff := a[i] - b[i]
		sum = safeAdd(sum, diff*diff)
	}
	return math.Sqrt(sum)
}

func (o *FixedOptimizer) initializeIdealPoint() {
	o.idealPoint = make([]float64, o.objectives)
	for i := range o.idealPoint {
		o.idealPoint[i] = math.Inf(1)
	}
}

func (o *FixedOptimizer) updateIdealPoint(objectives []float64) {
	for i := range objectives {
		if objectives[i] < o.idealPoint[i] && isValidValue(objectives[i]) {
			o.idealPoint[i] = objectives[i]
		}
	}
}

func (o *FixedOptimizer) tchebycheff(weights, objectives []float64) float64 {
	maxVal := 0.0
	for i := range objectives {
		if !isValidValue(objectives[i]) {
			return math.Inf(1)
		}

		normalizedDiff := math.Abs(objectives[i] - o.idealPoint[i])
		if o.idealPoint[i] != 0 && !math.IsInf(o.idealPoint[i], 1) {
			normalizedDiff /= math.Abs(o.idealPoint[i]) + 1e-10
		}
		val := weights[i] * normalizedDiff
		if val > maxVal {
			maxVal = val
		}
	}
	return maxVal
}

func (o *FixedOptimizer) createIndividual(numIngredients int, req OptimizationRequest) *Individual {
	amounts := make([]float64, numIngredients)
	totalWeight := 0.0

	for i := range amounts {
		amounts[i] = rand.Float64() * 100
		totalWeight = safeAdd(totalWeight, amounts[i])
	}

	targetWeight := o.getTotalWeightConstraint(req)
	if totalWeight > 0 {
		ratio := targetWeight / totalWeight
		for i := range amounts {
			amounts[i] *= ratio
			amounts[i] = o.clampAmount(amounts[i], req, i+1)
		}
	}

	return &Individual{Amounts: amounts}
}

func (o *FixedOptimizer) evaluateIndividual(ind *Individual, ingredients []Ingredient, req OptimizationRequest) {
	ind.Objectives = make([]float64, o.objectives)

	nutrition := o.calculateNutritionFromAmounts(ind.Amounts, ingredients)

	ind.Objectives[0] = o.calculateNutritionDeviation(nutrition, req.NutritionGoals)
	ind.Objectives[1] = o.calculateTotalCostFromAmounts(ind.Amounts, ingredients)
	ind.Objectives[2] = 1.0 - o.calculateDiversity(ind.Amounts)

	o.updateIdealPoint(ind.Objectives)
}

func (o *FixedOptimizer) calculateNutritionFromAmounts(amounts []float64, ingredients []Ingredient) NutritionSummary {
	var nutrition NutritionSummary

	for i, amount := range amounts {
		if i >= len(ingredients) {
			continue
		}
		ing := ingredients[i]
		factor := amount / 100.0

		nutrition.Energy = safeAdd(nutrition.Energy, ing.Energy*factor)
		nutrition.Protein = safeAdd(nutrition.Protein, ing.Protein*factor)
		nutrition.Fat = safeAdd(nutrition.Fat, ing.Fat*factor)
		nutrition.Carbs = safeAdd(nutrition.Carbs, ing.Carbs*factor)
		nutrition.Calcium = safeAdd(nutrition.Calcium, ing.Calcium*factor)
		nutrition.Iron = safeAdd(nutrition.Iron, ing.Iron*factor)
		nutrition.Zinc = safeAdd(nutrition.Zinc, ing.Zinc*factor)
		nutrition.VitaminC = safeAdd(nutrition.VitaminC, ing.VitaminC*factor)
	}

	return nutrition
}

func (o *FixedOptimizer) calculateNutritionDeviation(nutrition NutritionSummary, goals []NutritionGoal) float64 {
	deviation := 0.0

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

		if goal.Target > 0 {
			relDiff := math.Abs(actual-goal.Target) / goal.Target
			deviation = safeAdd(deviation, relDiff*goal.Weight)
		}
	}

	return deviation
}

func (o *FixedOptimizer) calculateTotalCostFromAmounts(amounts []float64, ingredients []Ingredient) float64 {
	totalCost := 0.0
	for i, amount := range amounts {
		if i >= len(ingredients) {
			continue
		}
		totalCost = safeAdd(totalCost, ingredients[i].Price*amount/100.0)
	}
	return totalCost
}

func (o *FixedOptimizer) calculateDiversity(amounts []float64) float64 {
	total := 0.0
	for _, amount := range amounts {
		total = safeAdd(total, amount)
	}

	if total == 0 {
		return 0.0
	}

	sumProportions := 0.0
	for _, amount := range amounts {
		proportion := amount / total
		sumProportions = safeAdd(sumProportions, proportion*proportion)
	}

	return 1.0 - sumProportions
}

func (o *FixedOptimizer) crossover(parent1, parent2 *Individual) *Individual {
	numVars := len(parent1.Amounts)
	child := &Individual{Amounts: make([]float64, numVars)}

	eta := 1.0
	for i := 0; i < numVars; i++ {
		if rand.Float64() < 0.5 {
			x1 := parent1.Amounts[i]
			x2 := parent2.Amounts[i]

			if math.Abs(x1-x2) > 1e-10 {
				u := rand.Float64()
				var beta float64

				if u <= 0.5 {
					beta = math.Pow(2*u, 1.0/(eta+1))
				} else {
					beta = math.Pow(1.0/(2*(1-u)), 1.0/(eta+1))
				}

				child.Amounts[i] = 0.5 * ((1+beta)*x1 + (1-beta)*x2)
			} else {
				child.Amounts[i] = x1
			}
		} else {
			child.Amounts[i] = parent1.Amounts[i]
		}
	}

	return child
}

func (o *FixedOptimizer) mutate(ind *Individual) {
	eta := 20.0
	prob := 1.0 / float64(len(ind.Amounts))

	for i := range ind.Amounts {
		if rand.Float64() < prob {
			x := ind.Amounts[i]
			delta1 := x / 500.0
			delta2 := (500 - x) / 500.0

			u := rand.Float64()
			var deltaq float64

			if u <= 0.5 {
				val := 2*u + (1-2*u)*math.Pow(1-delta1, eta+1)
				deltaq = math.Pow(val, 1.0/(eta+1)) - 1
			} else {
				val := 2*(1-u) + 2*(u-0.5)*math.Pow(1-delta2, eta+1)
				deltaq = 1 - math.Pow(val, 1.0/(eta+1))
			}

			x = safeAdd(x, deltaq*500.0)
			ind.Amounts[i] = math.Max(0, math.Min(500, x))
		}
	}
}

func (o *FixedOptimizer) repair(ind *Individual, req OptimizationRequest) {
	// 第一步: 应用食材约束
	for _, c := range req.Constraints {
		switch c.Type {
		case "ingredient_min":
			if c.IngredientID > 0 && c.IngredientID <= len(ind.Amounts) {
				idx := c.IngredientID - 1
				if ind.Amounts[idx] < c.Value {
					ind.Amounts[idx] = c.Value
				}
			}
		case "ingredient_max":
			if c.IngredientID > 0 && c.IngredientID <= len(ind.Amounts) {
				idx := c.IngredientID - 1
				if ind.Amounts[idx] > c.Value {
					ind.Amounts[idx] = c.Value
				}
			}
		}
	}

	// 确保所有值在0-500范围内
	for i := range ind.Amounts {
		ind.Amounts[i] = math.Max(0, math.Min(500, ind.Amounts[i]))
	}

	// 第二步: 归一化总重量
	targetWeight := o.getTotalWeightConstraint(req)
	currentTotal := 0.0
	for _, amount := range ind.Amounts {
		currentTotal = safeAdd(currentTotal, amount)
	}

	if currentTotal > 0 {
		ratio := targetWeight / currentTotal
		for i := range ind.Amounts {
			ind.Amounts[i] *= ratio
			ind.Amounts[i] = math.Max(0, math.Min(500, ind.Amounts[i]))
		}
	}

	// 再次检查约束
	for _, c := range req.Constraints {
		switch c.Type {
		case "ingredient_min":
			if c.IngredientID > 0 && c.IngredientID <= len(ind.Amounts) {
				idx := c.IngredientID - 1
				ind.Amounts[idx] = math.Max(c.Value, ind.Amounts[idx])
			}
		case "ingredient_max":
			if c.IngredientID > 0 && c.IngredientID <= len(ind.Amounts) {
				idx := c.IngredientID - 1
				ind.Amounts[idx] = math.Min(c.Value, ind.Amounts[idx])
			}
		}
	}
}

func (o *FixedOptimizer) selectBestSolution(population []*Individual, weights []Weight) int {
	nutritionWeight := 0.6
	costWeight := 0.3
	varietyWeight := 0.1

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

	totalWeight := nutritionWeight + costWeight + varietyWeight
	if totalWeight > 0 {
		nutritionWeight /= totalWeight
		costWeight /= totalWeight
		varietyWeight /= totalWeight
	}

	bestIdx := 0
	minScore := math.Inf(1)

	for i, ind := range population {
		score := ind.Objectives[0]*nutritionWeight +
			ind.Objectives[1]*costWeight +
			ind.Objectives[2]*varietyWeight

		if score < minScore {
			minScore = score
			bestIdx = i
		}
	}

	return bestIdx
}

func (o *FixedOptimizer) generateResult(ind *Individual, ingredients []Ingredient, req OptimizationRequest) *OptimizationResult {
	result := &OptimizationResult{
		Ingredients: make([]IngredientAmount, 0, len(ingredients)),
		Converged:   true,
		Warnings:    make([]string, 0),
	}

	for i, amount := range ind.Amounts {
		if i >= len(ingredients) {
			break
		}
		// 只保留用量大于0.1g的食材
		if amount > 0.1 {
			result.Ingredients = append(result.Ingredients, IngredientAmount{
				Ingredient: ingredients[i],
				Amount:     roundToTwoDecimalsFixed(amount),
			})
		}
	}

	// 计算营养汇总（使用修复后的计算）
	var iaList []IngredientAmount
	for i, amount := range ind.Amounts {
		if i >= len(ingredients) {
			break
		}
		iaList = append(iaList, IngredientAmount{
			Ingredient: ingredients[i],
			Amount:     amount,
		})
	}
	result.Nutrition = o.calculateNutrition(iaList)
	result.Cost = roundToTwoDecimals(o.calculateCost(iaList))

	diversity := 1.0 - ind.Objectives[2]
	if diversity < 0.3 {
		result.Warnings = append(result.Warnings, "食材多样性较低，建议增加食材种类")
	}

	return result
}

// 工具函数

// isValidValue 检查值是否有效（非NaN、非Inf、非负数）
func isValidValue(x float64) bool {
	return !math.IsNaN(x) && !math.IsInf(x, 0) && x >= 0
}

// fixValueFixed 修复数值异常
func fixValueFixed(x float64, warnings *[]string, name string) float64 {
	if math.IsNaN(x) {
		*warnings = append(*warnings, fmt.Sprintf("%s计算结果为NaN，已修正为0", name))
		return 0
	}
	if math.IsInf(x, 1) {
		*warnings = append(*warnings, fmt.Sprintf("%s计算结果为正无穷，已修正为最大值", name))
		return 1e6 // 设置一个合理的最大值
	}
	if math.IsInf(x, -1) {
		*warnings = append(*warnings, fmt.Sprintf("%s计算结果为负无穷，已修正为0", name))
		return 0
	}
	if x < 0 {
		*warnings = append(*warnings, fmt.Sprintf("%s计算结果为负数(%.2f)，已修正为0", name, x))
		return 0
	}
	if x > 1e6 {
		*warnings = append(*warnings, fmt.Sprintf("%s计算结果过大(%.2f)，已修正为最大值", name, x))
		return 1e6
	}
	return x
}

// fixAmountFixed 修复用量异常
func fixAmountFixed(x float64, warnings *[]string, name string) float64 {
	fixed := fixValueFixed(x, warnings, name+"用量")
	if fixed < 0 {
		fixed = 0
	}
	if fixed > 500 {
		*warnings = append(*warnings, fmt.Sprintf("%s用量(%.2fg)超过500g限制，已修正", name, fixed))
		fixed = 500
	}
	return roundToTwoDecimalsFixed(fixed)
}

// safeAdd 安全加法（避免溢出）
func safeAdd(a, b float64) float64 {
	result := a + b
	if math.IsInf(result, 1) {
		return 1e6
	}
	if math.IsInf(result, -1) {
		return 0
	}
	if math.IsNaN(result) {
		return a
	}
	return result
}

// roundNutritionSummaryFixed 四舍五入营养汇总值
func roundNutritionSummaryFixed(n *NutritionSummary) {
	n.Energy = roundToTwoDecimalsFixed(n.Energy)
	n.Protein = roundToTwoDecimalsFixed(n.Protein)
	n.Fat = roundToTwoDecimalsFixed(n.Fat)
	n.Carbs = roundToTwoDecimalsFixed(n.Carbs)
	n.Calcium = roundToTwoDecimalsFixed(n.Calcium)
	n.Iron = roundToTwoDecimalsFixed(n.Iron)
	n.Zinc = roundToTwoDecimalsFixed(n.Zinc)
	n.VitaminC = roundToTwoDecimalsFixed(n.VitaminC)
}

// LoadIngredientsFromJSON 从JSON文件加载食材数据
func (o *FixedOptimizer) LoadIngredientsFromJSON(filepath string) ([]Ingredient, error) {
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

// roundToTwoDecimalsFixed 四舍五入到两位小数
func roundToTwoDecimalsFixed(x float64) float64 {
	return math.Round(x*100) / 100
}
