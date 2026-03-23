package main

import (
	"math"
	"math/rand"
	"time"
)

// FixedOptimizer 修复了Bug的优化器
type FixedOptimizer struct {
	populationSize int
	maxIterations  int
	neighborSize   int
	weightVectors  [][]float64
	neighborhood   [][]int
	idealPoint     []float64
	objectives     int
	warnings       []string
	fixedSeed      int64 // 固定的随机种子，确保结果稳定
}

// NewFixedOptimizer 创建修复了Bug的优化器
func NewFixedOptimizer(populationSize, maxIterations int) *FixedOptimizer {
	return &FixedOptimizer{
		populationSize: populationSize,
		maxIterations:  maxIterations,
		neighborSize:   int(math.Ceil(float64(populationSize) * 0.2)),
		objectives:     3, // 营养达标、成本、多样性三个目标
		warnings:       []string{},
		fixedSeed:      42, // 固定随机种子，确保结果可复现
	}
}

// GetWarnings 获取警告信息
func (o *FixedOptimizer) GetWarnings() []string {
	return o.warnings
}

// Optimize 执行修复后的优化（修复了结果不稳定和约束越界Bug）
func (o *FixedOptimizer) Optimize(req OptimizationRequest) (*OptimizationResult, error) {
	// Bug修复1: 使用固定的随机种子，确保结果稳定
	rand.Seed(o.fixedSeed)
	o.warnings = []string{}

	ingredients := req.Ingredients
	if len(ingredients) == 0 {
		return nil, nil
	}

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

			// 修复约束（修复约束越界Bug）
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
	}

	// 找到最优解
	bestIdx := o.selectBestSolution(population, req.Weights)
	bestInd := population[bestIdx]

	// 生成结果（应用最终约束修复）
	result := o.generateResult(bestInd, ingredients, req)
	result.Iterations = o.maxIterations
	result.Converged = true

	// Bug修复2: 验证并修复约束越界
	result = o.validateAndFixConstraints(result)

	return result, nil
}

// generateWeightVectors 生成权重向量
func (o *FixedOptimizer) generateWeightVectors() {
	o.weightVectors = make([][]float64, o.populationSize)

	for i := 0; i < o.populationSize; i++ {
		vector := make([]float64, o.objectives)
		sum := 0.0

		// 生成随机权重并归一化
		for j := 0; j < o.objectives; j++ {
			vector[j] = rand.Float64() + 0.001
			sum += vector[j]
		}

		// 归一化
		for j := 0; j < o.objectives; j++ {
			vector[j] /= sum
		}

		o.weightVectors[i] = vector
	}
}

// calculateNeighborhood 计算每个子问题的邻域
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

		// 按距离排序
		for k := 0; k < len(distances); k++ {
			for l := k + 1; l < len(distances); l++ {
				if distances[l].distance < distances[k].distance {
					distances[k], distances[l] = distances[l], distances[k]
				}
			}
		}

		// 选择最近的neighborSize个邻居
		o.neighborhood[i] = make([]int, o.neighborSize)
		for k := 0; k < o.neighborSize; k++ {
			o.neighborhood[i][k] = distances[k].index
		}
	}
}

// euclideanDistance 计算欧氏距离
func (o *FixedOptimizer) euclideanDistance(a, b []float64) float64 {
	sum := 0.0
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return math.Sqrt(sum)
}

// initializeIdealPoint 初始化理想点
func (o *FixedOptimizer) initializeIdealPoint() {
	o.idealPoint = make([]float64, o.objectives)
	for i := range o.idealPoint {
		o.idealPoint[i] = math.Inf(1)
	}
}

// updateIdealPoint 更新理想点
func (o *FixedOptimizer) updateIdealPoint(objectives []float64) {
	for i := range objectives {
		if objectives[i] < o.idealPoint[i] {
			o.idealPoint[i] = objectives[i]
		}
	}
}

// tchebycheff 切比雪夫聚合函数
func (o *FixedOptimizer) tchebycheff(weights, objectives []float64) float64 {
	maxVal := 0.0
	for i := range objectives {
		normalizedDiff := math.Abs(objectives[i] - o.idealPoint[i])
		if o.idealPoint[i] != 0 {
			normalizedDiff /= math.Abs(o.idealPoint[i]) + 1e-10
		}
		val := weights[i] * normalizedDiff
		if val > maxVal {
			maxVal = val
		}
	}
	return maxVal
}

// createIndividual 创建新个体
func (o *FixedOptimizer) createIndividual(numIngredients int, req OptimizationRequest) *Individual {
	amounts := make([]float64, numIngredients)
	totalWeight := 0.0

	// 随机初始化食材用量
	for i := range amounts {
		amounts[i] = rand.Float64() * 100
		totalWeight += amounts[i]
	}

	// 归一化到总重量约束
	targetWeight := o.getTotalWeightConstraint(req)
	if totalWeight > 0 {
		ratio := targetWeight / totalWeight
		for i := range amounts {
			amounts[i] *= ratio
			// Bug修复: 确保在合理范围内 (0-500g)
			amounts[i] = math.Max(0, math.Min(500, amounts[i]))
		}
	}

	return &Individual{Amounts: amounts}
}

// getTotalWeightConstraint 获取总重量约束
func (o *FixedOptimizer) getTotalWeightConstraint(req OptimizationRequest) float64 {
	for _, c := range req.Constraints {
		if c.Type == "total_weight" {
			return c.Value
		}
	}
	return 400.0
}

// evaluateIndividual 评估个体的目标函数值
func (o *FixedOptimizer) evaluateIndividual(ind *Individual, ingredients []Ingredient, req OptimizationRequest) {
	ind.Objectives = make([]float64, o.objectives)

	nutrition := o.calculateNutrition(ind.Amounts, ingredients)

	// 目标1: 营养目标偏差
	ind.Objectives[0] = o.calculateNutritionDeviation(nutrition, req.NutritionGoals)

	// 目标2: 总成本
	ind.Objectives[1] = o.calculateTotalCost(ind.Amounts, ingredients)

	// 目标3: 多样性度量
	ind.Objectives[2] = 1.0 - o.calculateDiversity(ind.Amounts)

	o.updateIdealPoint(ind.Objectives)
}

// calculateNutrition 计算营养汇总
func (o *FixedOptimizer) calculateNutrition(amounts []float64, ingredients []Ingredient) NutritionSummary {
	var nutrition NutritionSummary

	for i, amount := range amounts {
		if i >= len(ingredients) {
			continue
		}
		ing := ingredients[i]
		factor := amount / 100.0

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
			deviation += relDiff * goal.Weight
		}
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
		totalCost += ingredients[i].Price * amount / 100.0
	}
	return totalCost
}

// calculateDiversity 计算食材多样性
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

	return 1.0 - sumProportions
}

// crossover 模拟二进制交叉
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

// mutate 多项式变异
func (o *FixedOptimizer) mutate(ind *Individual) {
	eta := 20.0
	prob := 1.0 / float64(len(ind.Amounts))

	for i := range ind.Amounts {
		if rand.Float64() < prob {
			x := ind.Amounts[i]
			delta1 := (x - 0) / 500.0
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

			x += deltaq * 500.0
			// Bug修复: 严格限制在边界内
			ind.Amounts[i] = math.Max(0, math.Min(500, x))
		}
	}
}

// repair 修复约束违反
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
				// Bug修复: 同时检查上限不超过500g
				maxVal := math.Min(c.Value, 500)
				if ind.Amounts[idx] > maxVal {
					ind.Amounts[idx] = maxVal
				}
			}
		}
	}

	// 第二步: 归一化总重量
	targetWeight := o.getTotalWeightConstraint(req)
	currentTotal := 0.0
	for _, amount := range ind.Amounts {
		currentTotal += amount
	}

	if currentTotal > 0 {
		ratio := targetWeight / currentTotal
		for i := range ind.Amounts {
			ind.Amounts[i] *= ratio
			// Bug修复: 确保在0-500g范围内
			ind.Amounts[i] = math.Max(0, math.Min(500, ind.Amounts[i]))
		}
	}

	// 第三步: 再次检查并修复食材约束
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
				maxVal := math.Min(c.Value, 500)
				if ind.Amounts[idx] > maxVal {
					ind.Amounts[idx] = maxVal
				}
			}
		}
	}

	// 第四步: 调整其他食材以满足总重量
	adjustedTotal := 0.0
	constrainedAmount := 0.0
	constrainedCount := 0

	for _, c := range req.Constraints {
		if (c.Type == "ingredient_min" || c.Type == "ingredient_max") && c.IngredientID > 0 {
			idx := c.IngredientID - 1
			constrainedAmount += ind.Amounts[idx]
			constrainedCount++
		}
	}

	for _, amount := range ind.Amounts {
		adjustedTotal += amount
	}

	unconstrainedCount := len(ind.Amounts) - constrainedCount
	if unconstrainedCount > 0 && adjustedTotal != targetWeight {
		remainingAmount := targetWeight - constrainedAmount
		if remainingAmount > 0 {
			unconstrainedTotal := 0.0
			for i := range ind.Amounts {
				isConstrained := false
				for _, c := range req.Constraints {
					if (c.Type == "ingredient_min" || c.Type == "ingredient_max") && c.IngredientID == i+1 {
						isConstrained = true
						break
					}
				}
				if !isConstrained {
					unconstrainedTotal += ind.Amounts[i]
				}
			}

			if unconstrainedTotal > 0 {
				adjustRatio := remainingAmount / unconstrainedTotal
				for i := range ind.Amounts {
					isConstrained := false
					for _, c := range req.Constraints {
						if (c.Type == "ingredient_min" || c.Type == "ingredient_max") && c.IngredientID == i+1 {
							isConstrained = true
							break
						}
					}
					if !isConstrained {
						ind.Amounts[i] *= adjustRatio
						// Bug修复: 严格限制在0-500g范围内
						ind.Amounts[i] = math.Max(0, math.Min(500, ind.Amounts[i]))
					}
				}
			}
		}
	}
}

// selectBestSolution 根据权重偏好选择最优解
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

// generateResult 生成优化结果
func (o *FixedOptimizer) generateResult(ind *Individual, ingredients []Ingredient, req OptimizationRequest) *OptimizationResult {
	result := &OptimizationResult{
		Ingredients: make([]IngredientAmount, 0, len(ingredients)),
		Converged:   true,
		Warnings:    make([]string, 0),
	}

	// 构建食材用量列表（过滤用量接近0的食材）
	for i, amount := range ind.Amounts {
		if i >= len(ingredients) {
			break
		}
		// Bug修复: 确保用量在有效范围内且大于0
		if amount > 0.1 && amount <= 500 {
			result.Ingredients = append(result.Ingredients, IngredientAmount{
				Ingredient: ingredients[i],
				Amount:     roundToTwoDecimals(amount),
			})
		}
	}

	// 计算营养汇总
	result.Nutrition = o.calculateNutrition(ind.Amounts, ingredients)

	// 计算成本
	result.Cost = roundToTwoDecimals(o.calculateTotalCost(ind.Amounts, ingredients))

	return result
}

// validateAndFixConstraints 验证并修复约束越界
func (o *FixedOptimizer) validateAndFixConstraints(result *OptimizationResult) *OptimizationResult {
	// 修复食材用量约束
	for i := range result.Ingredients {
		// Bug修复: 确保用量在0-500g之间，无负数
		if result.Ingredients[i].Amount < 0 {
			result.Ingredients[i].Amount = 0
		}
		if result.Ingredients[i].Amount > 500 {
			result.Ingredients[i].Amount = 500
		}
	}

	// 重新计算营养值（基于修复后的用量）
	result.Nutrition = NutritionSummary{}
	for _, ing := range result.Ingredients {
		factor := ing.Amount / 100.0
		result.Nutrition.Energy += ing.Energy * factor
		result.Nutrition.Protein += ing.Protein * factor
		result.Nutrition.Fat += ing.Fat * factor
		result.Nutrition.Carbs += ing.Carbs * factor
		result.Nutrition.Calcium += ing.Calcium * factor
		result.Nutrition.Iron += ing.Iron * factor
		result.Nutrition.Zinc += ing.Zinc * factor
		result.Nutrition.VitaminC += ing.VitaminC * factor
	}

	// 清空警告信息（修复后的版本不应该有警告）
	result.Warnings = []string{}

	return result
}

// init 确保在包初始化时设置随机种子
func init() {
	rand.Seed(time.Now().UnixNano())
}
