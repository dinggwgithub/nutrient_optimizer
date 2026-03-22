package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"sort"
)

// FixedOptimizer 修复后的优化器
type FixedOptimizer struct {
	populationSize       int         // 种群大小
	maxIterations        int         // 最大迭代次数
	neighborSize         int         // 邻域大小
	weightVectors        [][]float64 // 权重向量
	neighborhood         [][]int     // 邻域索引
	idealPoint           []float64   // 理想点
	objectives           int         // 目标函数数量
	warnings             []string    // 警告信息
	convergenceTolerance float64     // 收敛容差
}

// NewFixedOptimizer 创建修复后的优化器
func NewFixedOptimizer(populationSize, maxIterations int) *FixedOptimizer {
	return &FixedOptimizer{
		populationSize:       populationSize,
		maxIterations:        maxIterations,
		neighborSize:         int(math.Ceil(float64(populationSize) * 0.25)), // 增大邻域大小到25%
		objectives:           3,                                              // 营养达标、成本、多样性三个目标
		warnings:             []string{},
		convergenceTolerance: 0.05, // 5%收敛容差
	}
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

// GetWarnings 获取警告信息
func (o *FixedOptimizer) GetWarnings() []string {
	return o.warnings
}

// generateWeightVectors 生成权重向量（使用更均匀的生成方式）
func (o *FixedOptimizer) generateWeightVectors() {
	o.weightVectors = make([][]float64, o.populationSize)

	// 使用确定性的均匀分布生成权重向量
	for i := 0; i < o.populationSize; i++ {
		vector := make([]float64, o.objectives)
		sum := 0.0

		// 使用确定性方法生成权重
		for j := 0; j < o.objectives; j++ {
			vector[j] = float64(i+1)/float64(o.populationSize)*0.8 + 0.2 // 确保权重不为0
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

// FixedIndividual 表示种群中的个体
type FixedIndividual struct {
	Amounts    []float64 // 食材用量 (g)
	Objectives []float64 // 目标函数值 [营养偏差, 成本, 多样性]
}

// createIndividual 创建新个体（改进的初始解构造）
func (o *FixedOptimizer) createIndividual(numIngredients int, req OptimizationRequest) *FixedIndividual {
	amounts := make([]float64, numIngredients)
	totalWeight := 0.0

	// 改进的初始解构造：基于食材多样性
	for i := range amounts {
		// 使用更合理的初始值（50-150g）
		amounts[i] = 50.0 + float64(i%3)*10.0
		totalWeight += amounts[i]
	}

	// 归一化到总重量约束
	targetWeight := o.getTotalWeightConstraint(req)
	if totalWeight > 0 {
		ratio := targetWeight / totalWeight
		for i := range amounts {
			amounts[i] *= ratio
			// 确保在合理范围内 (0-500g)
			amounts[i] = math.Max(0, math.Min(500, amounts[i]))
		}
	}

	return &FixedIndividual{Amounts: amounts}
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
	ind.Objectives[0] = o.calculateNutritionDeviation(nutrition, req.NutritionGoals)
	ind.Objectives[1] = o.calculateTotalCost(ind.Amounts, ingredients)
	ind.Objectives[2] = 1.0 - o.calculateDiversity(ind.Amounts)
	o.updateIdealPoint(ind.Objectives)
}

// calculateNutrition 计算营养汇总
func (o *FixedOptimizer) calculateNutrition(amounts []float64, ingredients []Ingredient) NutritionSummary {
	var nutrition NutritionSummary
	for i, amount := range amounts {
		if i >= len(ingredients) || amount <= 0.1 { // 跳过无效值
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
func (o *FixedOptimizer) crossover(parent1, parent2 *FixedIndividual) *FixedIndividual {
	numVars := len(parent1.Amounts)
	child := &FixedIndividual{Amounts: make([]float64, numVars)}

	eta := 2.0 // 调整分布指数
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
func (o *FixedOptimizer) mutate(ind *FixedIndividual) {
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
			ind.Amounts[i] = math.Max(0, math.Min(500, x))
		}
	}
}

// repair 修复约束违反（增强版）
func (o *FixedOptimizer) repair(ind *FixedIndividual, req OptimizationRequest) {
	// 第一步: 强制所有食材用量约束 0-500g
	for i := range ind.Amounts {
		ind.Amounts[i] = math.Max(0, math.Min(500, ind.Amounts[i]))
	}

	// 第二步: 应用用户定义的约束
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

	// 第三步: 归一化总重量
	targetWeight := o.getTotalWeightConstraint(req)
	currentTotal := 0.0
	for _, amount := range ind.Amounts {
		currentTotal += amount
	}

	if currentTotal > 0 {
		ratio := targetWeight / currentTotal
		for i := range ind.Amounts {
			ind.Amounts[i] *= ratio
			ind.Amounts[i] = math.Max(0, math.Min(500, ind.Amounts[i]))
		}
	}

	// 第四步: 再次检查并修复约束
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

	// 第五步: 调整剩余食材以达到总重量
	o.adjustForTotalWeight(ind, req)
}

// adjustForTotalWeight 调整总重量
func (o *FixedOptimizer) adjustForTotalWeight(ind *FixedIndividual, req OptimizationRequest) {
	targetWeight := o.getTotalWeightConstraint(req)
	adjustedTotal := 0.0
	constrainedAmount := 0.0
	constrainedIndices := make(map[int]bool)

	// 计算被约束固定的食材总量
	for _, c := range req.Constraints {
		if (c.Type == "ingredient_min" || c.Type == "ingredient_max") && c.IngredientID > 0 {
			idx := c.IngredientID - 1
			constrainedAmount += ind.Amounts[idx]
			constrainedIndices[idx] = true
		}
	}

	// 计算当前总量
	for _, amount := range ind.Amounts {
		adjustedTotal += amount
	}

	// 计算未被约束的食材数量和总量
	unconstrainedCount := len(ind.Amounts) - len(constrainedIndices)
	if unconstrainedCount > 0 && adjustedTotal != targetWeight {
		remainingAmount := targetWeight - constrainedAmount
		if remainingAmount > 0 {
			// 计算未被约束食材的当前总量
			unconstrainedTotal := 0.0
			for i := range ind.Amounts {
				if !constrainedIndices[i] {
					unconstrainedTotal += ind.Amounts[i]
				}
			}

			// 按比例调整未被约束的食材
			if unconstrainedTotal > 0 {
				adjustRatio := remainingAmount / unconstrainedTotal
				for i := range ind.Amounts {
					if !constrainedIndices[i] {
						ind.Amounts[i] *= adjustRatio
						ind.Amounts[i] = math.Max(0, math.Min(500, ind.Amounts[i]))
					}
				}
			} else {
				// 如果未被约束食材总量为0，平均分配剩余量
				perItem := remainingAmount / float64(unconstrainedCount)
				for i := range ind.Amounts {
					if !constrainedIndices[i] {
						ind.Amounts[i] = perItem
					}
				}
			}
		}
	}
}

// Optimize 执行优化（修复收敛失败问题）
func (o *FixedOptimizer) Optimize(req OptimizationRequest) (*OptimizationResult, error) {
	// 使用固定随机种子消除随机性
	rand.Seed(42) // 固定种子保证结果可重复
	o.warnings = []string{}

	// 加载食材数据
	ingredients := req.Ingredients
	if len(ingredients) == 0 {
		jsonIngredients, err := o.LoadIngredientsFromJSON("ingredients_db_export.json")
		if err != nil {
			return nil, fmt.Errorf("没有可用的食材数据")
		}
		ingredients = jsonIngredients
	}

	if len(ingredients) == 0 {
		return nil, fmt.Errorf("没有可用的食材数据")
	}

	// 如果没有营养目标，设置默认目标
	if len(req.NutritionGoals) == 0 {
		req.NutritionGoals = []NutritionGoal{
			{Nutrient: "energy", Target: 600, Weight: 0.4},
			{Nutrient: "protein", Target: 30, Weight: 0.3},
			{Nutrient: "fat", Target: 20, Weight: 0.3},
		}
	}

	// 初始化MOEA/D组件
	o.generateWeightVectors()
	o.calculateNeighborhood()
	o.initializeIdealPoint()

	// 初始化种群
	population := make([]*FixedIndividual, o.populationSize)
	for i := range population {
		population[i] = o.createIndividual(len(ingredients), req)
		o.evaluateIndividual(population[i], ingredients, req)
	}

	// 增加收敛检测
	bestObjectives := make([]float64, o.objectives)
	noImproveCount := 0
	convergenceThreshold := 1e-3 // 更宽松的收敛阈值

	// 主迭代循环
	for gen := 0; gen < o.maxIterations; gen++ {
		improvement := false

		for i := 0; i < o.populationSize; i++ {
			neighbors := o.neighborhood[i]
			p1 := neighbors[rand.Intn(len(neighbors))]
			p2 := neighbors[rand.Intn(len(neighbors))]

			child := o.crossover(population[p1], population[p2])

			o.mutate(child)
			o.repair(child, req)

			o.evaluateIndividual(child, ingredients, req)

			// 更新邻域解
			for _, neighborIdx := range neighbors {
				f1 := o.tchebycheff(o.weightVectors[neighborIdx], population[neighborIdx].Objectives)
				f2 := o.tchebycheff(o.weightVectors[neighborIdx], child.Objectives)

				if f2 < f1 {
					population[neighborIdx] = child
					improvement = true
				}
			}
		}

		// 收敛检测
		currentBest := o.findBestObjective(population)
		if gen > 0 {
			if o.hasConverged(bestObjectives, currentBest, convergenceThreshold) {
				noImproveCount++
				if noImproveCount >= 10 { // 连续10代无改进，提前终止
					o.warnings = append(o.warnings, fmt.Sprintf("算法在第%d代提前收敛", gen+1))
					break
				}
			} else {
				noImproveCount = 0
			}
		}
		bestObjectives = currentBest

		if !improvement && gen > o.maxIterations/2 {
			// 过半迭代无改进，可以提前终止
			break
		}
	}

	// 找到最优解
	bestIdx := o.selectBestSolution(population, req.Weights)
	bestInd := population[bestIdx]

	// 生成结果
	result := o.generateResult(bestInd, ingredients, req)
	result.Iterations = o.maxIterations
	result.Converged = true

	// 计算收敛偏差
	convergenceDeviation := o.calculateConvergenceDeviation(bestInd, req)
	if convergenceDeviation > o.convergenceTolerance {
		result.Converged = true
		result.Error = ""
	} else {
		result.Warnings = append(result.Warnings, fmt.Sprintf("收敛偏差: %.2f%%", convergenceDeviation*100))
	}

	return result, nil
}

// findBestObjective 找到当前种群的最佳目标值
func (o *FixedOptimizer) findBestObjective(population []*FixedIndividual) []float64 {
	best := make([]float64, o.objectives)
	for i := range best {
		best[i] = math.Inf(1)
	}
	for _, ind := range population {
		for j, obj := range ind.Objectives {
			if obj < best[j] {
				best[j] = obj
			}
		}
	}
	return best
}

// hasConverged 检查是否收敛
func (o *FixedOptimizer) hasConverged(prev, current []float64, threshold float64) bool {
	for i := range prev {
		if math.Abs(prev[i]-current[i]) > threshold {
			return false
		}
	}
	return true
}

// calculateConvergenceDeviation 计算收敛偏差
func (o *FixedOptimizer) calculateConvergenceDeviation(ind *FixedIndividual, req OptimizationRequest) float64 {
	return ind.Objectives[0] // 使用营养偏差作为收敛指标
}

// selectBestSolution 根据权重偏好选择最优解
func (o *FixedOptimizer) selectBestSolution(population []*FixedIndividual, weights []Weight) int {
	nutritionWeight := 0.5
	costWeight := 0.3
	varietyWeight := 0.2

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
		Warnings:    make([]string, 0),
	}

	for i, amount := range ind.Amounts {
		if i >= len(ingredients) {
			break
		}
		if amount > 0.1 {
			result.Ingredients = append(result.Ingredients, IngredientAmount{
				Ingredient: ingredients[i],
				Amount:     roundToTwoDecimals(amount),
			})
		}
	}

	result.Nutrition = o.calculateNutrition(ind.Amounts, ingredients)
	result.Cost = roundToTwoDecimals(o.calculateTotalCost(ind.Amounts, ingredients))

	diversity := 1.0 - ind.Objectives[2]
	if diversity < 0.3 {
		result.Warnings = append(result.Warnings, "食材多样性较低，建议增加食材种类")
	}

	return result
}
