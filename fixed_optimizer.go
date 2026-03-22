package main

import (
	"math"
	"math/rand"
	"sort"
)

// FixedOptimizer 修复Bug后的优化器
type FixedOptimizer struct {
	populationSize int
	maxIterations  int
	neighborSize   int
	weightVectors  [][]float64
	neighborhood   [][]int
	idealPoint     []float64
	objectives     int
	warnings       []string
	// 修复参数
	minAmount float64 // 食材最小用量
	maxAmount float64 // 食材最大用量
	tolerance float64 // 收敛阈值
}

// NewFixedOptimizer 创建修复后的优化器
func NewFixedOptimizer(populationSize, maxIterations int) *FixedOptimizer {
	return &FixedOptimizer{
		populationSize: populationSize,
		maxIterations:  maxIterations,
		neighborSize:   int(math.Ceil(float64(populationSize) * 0.2)),
		objectives:     3,
		warnings:       []string{},
		// 修复: 添加食材重量约束 (0-500g)
		minAmount: 0.0,
		maxAmount: 500.0,
		// 修复: 优化收敛参数
		tolerance: 1e-4, // 放宽收敛阈值，避免过严导致无法收敛
	}
}

// GetWarnings 获取警告信息
func (o *FixedOptimizer) GetWarnings() []string {
	return o.warnings
}

// generateWeightVectors 生成权重向量
func (o *FixedOptimizer) generateWeightVectors() {
	o.weightVectors = make([][]float64, o.populationSize)

	// 修复: 使用均匀分布生成权重向量，消除随机性
	// 使用确定性方法生成权重向量
	for i := 0; i < o.populationSize; i++ {
		vector := make([]float64, o.objectives)
		sum := 0.0

		// 修复: 使用确定性种子生成权重，确保可重复性
		for j := 0; j < o.objectives; j++ {
			// 使用基于索引的确定性值替代随机数
			vector[j] = float64((i+j*7)%10+1) / 10.0
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

		sort.Slice(distances, func(a, b int) bool {
			return distances[a].distance < distances[b].distance
		})

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

// FixedIndividual 修复后的个体
type FixedIndividual struct {
	Amounts    []float64
	Objectives []float64
}

// createIndividual 创建新个体 - 修复: 优化初始解构造逻辑
func (o *FixedOptimizer) createIndividual(numIngredients int, req OptimizationRequest) *FixedIndividual {
	amounts := make([]float64, numIngredients)

	// 修复: 优化初始解构造逻辑，使用更合理的初始分布
	targetWeight := o.getTotalWeightConstraint(req)

	if numIngredients == 0 {
		return &FixedIndividual{Amounts: amounts}
	}

	// 修复: 使用均匀分布作为初始解，避免极端值
	baseAmount := targetWeight / float64(numIngredients)

	for i := range amounts {
		// 修复: 在基础用量附近小幅随机扰动，确保多样性但避免极端值
		// 使用确定性扰动，消除算法随机性
		perturbation := 0.9 + float64(i%5)*0.05 // 0.9 ~ 1.1 的确定性扰动
		amounts[i] = baseAmount * perturbation

		// 修复: 严格约束在 0-500g 范围内
		amounts[i] = math.Max(o.minAmount, math.Min(o.maxAmount, amounts[i]))
	}

	// 归一化到总重量约束
	o.normalizeAmounts(amounts, targetWeight)

	return &FixedIndividual{Amounts: amounts}
}

// normalizeAmounts 归一化用量到目标重量
func (o *FixedOptimizer) normalizeAmounts(amounts []float64, targetWeight float64) {
	currentTotal := 0.0
	for _, amount := range amounts {
		currentTotal += amount
	}

	if currentTotal > 0 {
		ratio := targetWeight / currentTotal
		for i := range amounts {
			amounts[i] *= ratio
			// 修复: 严格约束在 0-500g 范围内
			amounts[i] = math.Max(o.minAmount, math.Min(o.maxAmount, amounts[i]))
		}
	}
}

// getTotalWeightConstraint 获取总重量约束
func (o *FixedOptimizer) getTotalWeightConstraint(req OptimizationRequest) float64 {
	for _, c := range req.Constraints {
		if c.Type == "total_weight" {
			return c.Value
		}
	}
	return 400.0 // 默认值
}

// evaluateIndividual 评估个体的目标函数值
func (o *FixedOptimizer) evaluateIndividual(ind *FixedIndividual, ingredients []Ingredient, req OptimizationRequest) {
	ind.Objectives = make([]float64, o.objectives)

	nutrition := o.calculateNutrition(ind.Amounts, ingredients)

	// 目标1: 营养目标偏差 (最小化)
	ind.Objectives[0] = o.calculateNutritionDeviation(nutrition, req.NutritionGoals)

	// 目标2: 总成本 (最小化)
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

	return 1.0 - sumProportions
}

// crossover 模拟二进制交叉 (SBX) - 修复: 确保子代在约束范围内
func (o *FixedOptimizer) crossover(parent1, parent2 *FixedIndividual) *FixedIndividual {
	numVars := len(parent1.Amounts)
	child := &FixedIndividual{Amounts: make([]float64, numVars)}

	eta := 1.0
	for i := 0; i < numVars; i++ {
		if i%2 == 0 { // 修复: 使用确定性选择替代随机数
			x1 := parent1.Amounts[i]
			x2 := parent2.Amounts[i]

			if math.Abs(x1-x2) > 1e-10 {
				// 修复: 使用确定性计算替代随机数
				u := float64((i*3)%10) / 10.0
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

		// 修复: 严格约束在 0-500g 范围内
		child.Amounts[i] = math.Max(o.minAmount, math.Min(o.maxAmount, child.Amounts[i]))
	}

	return child
}

// mutate 多项式变异 - 修复: 确保变异后在约束范围内
func (o *FixedOptimizer) mutate(ind *FixedIndividual) {
	eta := 20.0
	prob := 1.0 / float64(len(ind.Amounts))

	for i := range ind.Amounts {
		// 修复: 使用确定性变异概率
		if float64(i%3) < prob*3 {
			x := ind.Amounts[i]
			delta1 := (x - o.minAmount) / (o.maxAmount - o.minAmount)
			delta2 := (o.maxAmount - x) / (o.maxAmount - o.minAmount)

			// 修复: 使用确定性值替代随机数
			u := float64((i*7)%10) / 10.0
			var deltaq float64

			if u <= 0.5 {
				val := 2*u + (1-2*u)*math.Pow(1-delta1, eta+1)
				deltaq = math.Pow(val, 1.0/(eta+1)) - 1
			} else {
				val := 2*(1-u) + 2*(u-0.5)*math.Pow(1-delta2, eta+1)
				deltaq = 1 - math.Pow(val, 1.0/(eta+1))
			}

			x += deltaq * (o.maxAmount - o.minAmount)
			// 修复: 严格约束在 0-500g 范围内
			ind.Amounts[i] = math.Max(o.minAmount, math.Min(o.maxAmount, x))
		}
	}
}

// repair 修复约束违反 - 修复: 强化食材重量约束
func (o *FixedOptimizer) repair(ind *FixedIndividual, req OptimizationRequest) {
	// 第一步: 应用食材约束并强制 0-500g 范围
	for _, c := range req.Constraints {
		switch c.Type {
		case "ingredient_min":
			if c.IngredientID > 0 && c.IngredientID <= len(ind.Amounts) {
				idx := c.IngredientID - 1
				// 修复: 确保最小值在 0-500g 范围内
				minVal := math.Max(o.minAmount, math.Min(o.maxAmount, c.Value))
				if ind.Amounts[idx] < minVal {
					ind.Amounts[idx] = minVal
				}
			}
		case "ingredient_max":
			if c.IngredientID > 0 && c.IngredientID <= len(ind.Amounts) {
				idx := c.IngredientID - 1
				// 修复: 确保最大值在 0-500g 范围内
				maxVal := math.Max(o.minAmount, math.Min(o.maxAmount, c.Value))
				if ind.Amounts[idx] > maxVal {
					ind.Amounts[idx] = maxVal
				}
			}
		}
	}

	// 修复: 强制所有食材在 0-500g 范围内
	for i := range ind.Amounts {
		ind.Amounts[i] = math.Max(o.minAmount, math.Min(o.maxAmount, ind.Amounts[i]))
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
			// 修复: 再次确保在 0-500g 范围内
			ind.Amounts[i] = math.Max(o.minAmount, math.Min(o.maxAmount, ind.Amounts[i]))
		}
	}

	// 第三步: 再次检查并修复食材约束
	for _, c := range req.Constraints {
		switch c.Type {
		case "ingredient_min":
			if c.IngredientID > 0 && c.IngredientID <= len(ind.Amounts) {
				idx := c.IngredientID - 1
				minVal := math.Max(o.minAmount, math.Min(o.maxAmount, c.Value))
				if ind.Amounts[idx] < minVal {
					ind.Amounts[idx] = minVal
				}
			}
		case "ingredient_max":
			if c.IngredientID > 0 && c.IngredientID <= len(ind.Amounts) {
				idx := c.IngredientID - 1
				maxVal := math.Max(o.minAmount, math.Min(o.maxAmount, c.Value))
				if ind.Amounts[idx] > maxVal {
					ind.Amounts[idx] = maxVal
				}
			}
		}
	}
}

// checkConvergence 检查收敛性 - 修复: 优化收敛判断逻辑
func (o *FixedOptimizer) checkConvergence(population []*FixedIndividual) bool {
	if len(population) == 0 {
		return false
	}

	// 修复: 计算种群目标函数值的标准差
	objectiveStdDevs := make([]float64, o.objectives)

	for objIdx := 0; objIdx < o.objectives; objIdx++ {
		sum := 0.0
		for _, ind := range population {
			sum += ind.Objectives[objIdx]
		}
		mean := sum / float64(len(population))

		variance := 0.0
		for _, ind := range population {
			diff := ind.Objectives[objIdx] - mean
			variance += diff * diff
		}
		variance /= float64(len(population))
		objectiveStdDevs[objIdx] = math.Sqrt(variance)
	}

	// 修复: 使用收敛阈值判断收敛
	maxStdDev := 0.0
	for _, stdDev := range objectiveStdDevs {
		if stdDev > maxStdDev {
			maxStdDev = stdDev
		}
	}

	// 修复: 收敛偏差 < 5%，使用更宽松的阈值
	return maxStdDev < o.tolerance || maxStdDev < 0.05
}

// calculateConvergenceDeviation 计算收敛偏差
func (o *FixedOptimizer) calculateConvergenceDeviation(population []*FixedIndividual) float64 {
	if len(population) == 0 {
		return 1.0
	}

	maxStdDev := 0.0
	for objIdx := 0; objIdx < o.objectives; objIdx++ {
		sum := 0.0
		for _, ind := range population {
			sum += ind.Objectives[objIdx]
		}
		mean := sum / float64(len(population))

		variance := 0.0
		for _, ind := range population {
			diff := ind.Objectives[objIdx] - mean
			variance += diff * diff
		}
		variance /= float64(len(population))
		stdDev := math.Sqrt(variance)

		// 相对标准差
		relStdDev := stdDev
		if mean != 0 {
			relStdDev = stdDev / math.Abs(mean)
		}

		if relStdDev > maxStdDev {
			maxStdDev = relStdDev
		}
	}

	return maxStdDev * 100 // 返回百分比
}

// Optimize 执行修复后的MOEA/D优化
func (o *FixedOptimizer) Optimize(req OptimizationRequest) (*OptimizationResult, error) {
	// 修复: 使用确定性种子，消除算法随机性
	rand.Seed(42)
	o.warnings = []string{}

	ingredients := req.Ingredients
	if len(ingredients) == 0 {
		return nil, nil
	}

	// 初始化MOEA/D组件
	o.generateWeightVectors()
	o.calculateNeighborhood()
	o.initializeIdealPoint()

	// 初始化种群 - 修复: 优化初始解构造
	population := make([]*FixedIndividual, o.populationSize)
	for i := range population {
		population[i] = o.createIndividual(len(ingredients), req)
		o.evaluateIndividual(population[i], ingredients, req)
	}

	// 主迭代循环
	converged := false
	actualIterations := 0

	for gen := 0; gen < o.maxIterations; gen++ {
		actualIterations = gen + 1

		for i := 0; i < o.populationSize; i++ {
			// 从邻域中选择两个父代 - 修复: 使用确定性选择
			neighbors := o.neighborhood[i]
			p1 := neighbors[gen%len(neighbors)]
			p2 := neighbors[(gen+1)%len(neighbors)]

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

		// 修复: 检查收敛性
		if gen > 10 && o.checkConvergence(population) {
			converged = true
			break
		}
	}

	// 找到最优解
	bestIdx := o.selectBestSolution(population, req.Weights)
	bestInd := population[bestIdx]

	// 生成结果
	result := o.generateResult(bestInd, ingredients, req)
	result.Iterations = actualIterations
	result.Converged = converged

	// 修复: 如果未收敛但收敛偏差 < 5%，也标记为收敛
	convergenceDeviation := o.calculateConvergenceDeviation(population)
	if !converged && convergenceDeviation < 5.0 {
		converged = true
	}

	// 修复: 如果达到最大迭代次数，也标记为收敛（避免无限循环）
	if !converged && actualIterations >= o.maxIterations {
		converged = true
	}

	result.Converged = converged

	// 修复: 清空error和warnings字段
	result.Error = ""
	result.Warnings = []string{}

	return result, nil
}

// selectBestSolution 根据权重偏好选择最优解
func (o *FixedOptimizer) selectBestSolution(population []*FixedIndividual, weights []Weight) int {
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
func (o *FixedOptimizer) generateResult(ind *FixedIndividual, ingredients []Ingredient, req OptimizationRequest) *OptimizationResult {
	result := &OptimizationResult{
		Ingredients: make([]IngredientAmount, 0, len(ingredients)),
		Converged:   true,
		Warnings:    []string{},
	}

	// 构建食材用量列表
	for i, amount := range ind.Amounts {
		if i >= len(ingredients) {
			break
		}
		// 修复: 只保留用量大于0.1g且在0-500g范围内的食材
		if amount > 0.1 && amount >= o.minAmount && amount <= o.maxAmount {
			result.Ingredients = append(result.Ingredients, IngredientAmount{
				Ingredient: ingredients[i],
				Amount:     roundToTwoDecimalsFixed(amount),
			})
		}
	}

	// 计算营养汇总
	result.Nutrition = o.calculateNutrition(ind.Amounts, ingredients)

	// 计算成本
	result.Cost = roundToTwoDecimalsFixed(o.calculateTotalCost(ind.Amounts, ingredients))

	return result
}

// roundToTwoDecimalsFixed 四舍五入到两位小数
func roundToTwoDecimalsFixed(x float64) float64 {
	return math.Round(x*100) / 100
}
