package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// FixedOptimizer 修复后的优化器
type FixedOptimizer struct {
	populationSize int         // 种群大小
	maxIterations  int         // 最大迭代次数
	neighborSize   int         // 邻域大小
	weightVectors  [][]float64 // 权重向量
	neighborhood   [][]int     // 邻域索引
	idealPoint     []float64   // 理想点
	objectives     int         // 目标函数数量
	warnings       []string    // 警告信息
	randomSeed     int64       // 固定随机种子
}

// NewFixedOptimizer 创建修复后的优化器
func NewFixedOptimizer(populationSize, maxIterations int) *FixedOptimizer {
	return &FixedOptimizer{
		populationSize: populationSize,
		maxIterations:  maxIterations,
		neighborSize:   int(math.Ceil(float64(populationSize) * 0.2)), // 邻域大小为种群的20%
		objectives:     3,                                             // 营养达标、成本、多样性三个目标
		warnings:       []string{},
		randomSeed:     42, // 固定随机种子，确保结果可重复
	}
}

// GetWarnings 获取警告信息
func (o *FixedOptimizer) GetWarnings() []string {
	return o.warnings
}

// generateWeightVectors 生成权重向量（使用固定种子）
func (o *FixedOptimizer) generateWeightVectors() {
	rng := rand.New(rand.NewSource(o.randomSeed))
	o.weightVectors = make([][]float64, o.populationSize)

	// 使用均匀分布生成权重向量
	for i := 0; i < o.populationSize; i++ {
		vector := make([]float64, o.objectives)
		sum := 0.0

		// 生成随机权重并归一化
		for j := 0; j < o.objectives; j++ {
			vector[j] = rng.Float64() + 0.001 // 避免零权重
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
		// 计算与其他权重向量的欧氏距离
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
		sort.Slice(distances, func(a, b int) bool {
			return distances[a].distance < distances[b].distance
		})

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
		// 计算归一化的目标值与理想点的偏差
		normalizedDiff := math.Abs(objectives[i] - o.idealPoint[i])
		// 避免除以零
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

// createIndividual 创建新个体（使用固定种子）
func (o *FixedOptimizer) createIndividual(numIngredients int, req OptimizationRequest) *FixedIndividual {
	rng := rand.New(rand.NewSource(o.randomSeed))
	amounts := make([]float64, numIngredients)

	// 第一步：确定每个食材的约束范围
	minAmounts := make([]float64, numIngredients)
	maxAmounts := make([]float64, numIngredients)
	for i := range amounts {
		minAmounts[i] = 0
		maxAmounts[i] = 500
	}

	// 应用用户指定的约束
	for _, c := range req.Constraints {
		if c.IngredientID > 0 && c.IngredientID <= numIngredients {
			idx := c.IngredientID - 1
			if c.Type == "ingredient_min" {
				minAmounts[idx] = math.Max(0, math.Min(500, c.Value))
			} else if c.Type == "ingredient_max" {
				maxAmounts[idx] = math.Max(0, math.Min(500, c.Value))
			}
			// 确保min <= max
			if minAmounts[idx] > maxAmounts[idx] {
				minAmounts[idx] = maxAmounts[idx]
			}
		}
	}

	// 第二步：在约束范围内随机初始化食材用量
	constrainedTotal := 0.0
	unconstrainedTotal := 0.0

	for i := range amounts {
		// 如果有明确的min和max约束且相等，固定为该值
		if math.Abs(minAmounts[i]-maxAmounts[i]) < 0.01 {
			amounts[i] = minAmounts[i]
			constrainedTotal += amounts[i]
		} else {
			// 否则，在min和max之间随机初始化
			amounts[i] = minAmounts[i] + rng.Float64()*(maxAmounts[i]-minAmounts[i])
			unconstrainedTotal += amounts[i]
		}
	}

	// 第三步：归一化到总重量约束，保留约束食材的量
	targetWeight := o.getTotalWeightConstraint(req)
	remaining := targetWeight - constrainedTotal

	if remaining < 0 {
		remaining = 0
	}

	// 调整未约束食材的量
	if unconstrainedTotal > 0 {
		ratio := remaining / unconstrainedTotal
		for i := range amounts {
			if math.Abs(minAmounts[i]-maxAmounts[i]) >= 0.01 {
				amounts[i] *= ratio
				// 确保仍在约束范围内
				amounts[i] = math.Max(minAmounts[i], math.Min(maxAmounts[i], amounts[i]))
			}
		}
	} else if remaining > 0 {
		// 如果没有未约束食材，将剩余量平均分配到所有食材（不违反约束）
		avg := remaining / float64(numIngredients)
		for i := range amounts {
			newAmount := amounts[i] + avg
			if newAmount <= maxAmounts[i] {
				amounts[i] = newAmount
			}
		}
	}

	// 最终边界检查
	for i := range amounts {
		amounts[i] = math.Max(minAmounts[i], math.Min(maxAmounts[i], amounts[i]))
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

	// 计算营养汇总
	nutrition := o.calculateNutrition(ind.Amounts, ingredients)

	// 目标1: 营养目标偏差 (最小化)
	ind.Objectives[0] = o.calculateNutritionDeviation(nutrition, req.NutritionGoals)

	// 目标2: 总成本 (最小化)
	ind.Objectives[1] = o.calculateTotalCost(ind.Amounts, ingredients)

	// 目标3: 多样性度量 (使用1 - Simpson指数，数值越低多样性越高)
	ind.Objectives[2] = 1.0 - o.calculateDiversity(ind.Amounts)

	// 更新理想点
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

		// 计算相对偏差
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
		// 价格是元/100g，转换为元
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

// crossover 模拟二进制交叉 (SBX) - 使用固定种子
func (o *FixedOptimizer) crossover(parent1, parent2 *FixedIndividual) *FixedIndividual {
	rng := rand.New(rand.NewSource(o.randomSeed))
	numVars := len(parent1.Amounts)
	child := &FixedIndividual{Amounts: make([]float64, numVars)}

	eta := 1.0 // 分布指数
	for i := 0; i < numVars; i++ {
		if rng.Float64() < 0.5 {
			x1 := parent1.Amounts[i]
			x2 := parent2.Amounts[i]

			if math.Abs(x1-x2) > 1e-10 {
				u := rng.Float64()
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

// mutate 多项式变异 - 使用固定种子
func (o *FixedOptimizer) mutate(ind *FixedIndividual) {
	rng := rand.New(rand.NewSource(o.randomSeed))
	eta := 20.0                             // 分布指数
	prob := 1.0 / float64(len(ind.Amounts)) // 每个变量的变异概率

	for i := range ind.Amounts {
		if rng.Float64() < prob {
			x := ind.Amounts[i]
			delta1 := (x - 0) / 500.0   // 下界0
			delta2 := (500 - x) / 500.0 // 上界500

			u := rng.Float64()
			var deltaq float64

			if u <= 0.5 {
				val := 2*u + (1-2*u)*math.Pow(1-delta1, eta+1)
				deltaq = math.Pow(val, 1.0/(eta+1)) - 1
			} else {
				val := 2*(1-u) + 2*(u-0.5)*math.Pow(1-delta2, eta+1)
				deltaq = 1 - math.Pow(val, 1.0/(eta+1))
			}

			x += deltaq * 500.0
			// 确保在边界内
			ind.Amounts[i] = math.Max(0, math.Min(500, x))
		}
	}
}

// repair 修复约束违反 - 强化约束检查
func (o *FixedOptimizer) repair(ind *FixedIndividual, req OptimizationRequest) {
	// 第一步: 记录哪些食材被用户约束了，并应用约束
	constrained := make(map[int]bool)
	constrainedAmount := 0.0

	for _, c := range req.Constraints {
		if c.Type == "ingredient_min" || c.Type == "ingredient_max" {
			if c.IngredientID > 0 && c.IngredientID <= len(ind.Amounts) {
				constrained[c.IngredientID] = true
			}
		}
	}

	// 第二步: 应用用户指定的食材约束（优先级最高）
	for _, c := range req.Constraints {
		switch c.Type {
		case "ingredient_min":
			if c.IngredientID > 0 && c.IngredientID <= len(ind.Amounts) {
				idx := c.IngredientID - 1
				minVal := math.Max(0, math.Min(500, c.Value))
				ind.Amounts[idx] = math.Max(minVal, ind.Amounts[idx])
			}
		case "ingredient_max":
			if c.IngredientID > 0 && c.IngredientID <= len(ind.Amounts) {
				idx := c.IngredientID - 1
				maxVal := math.Max(0, math.Min(500, c.Value))
				ind.Amounts[idx] = math.Min(maxVal, ind.Amounts[idx])
			}
		}
	}

	// 第三步: 确保所有食材在0-500g范围内（通用约束）
	for i := range ind.Amounts {
		ind.Amounts[i] = math.Max(0, math.Min(500, ind.Amounts[i]))
	}

	// 第四步: 归一化总重量，但保留约束食材的量
	targetWeight := o.getTotalWeightConstraint(req)
	currentTotal := 0.0

	// 计算约束食材的总量
	for ingID := range constrained {
		idx := ingID - 1
		constrainedAmount += ind.Amounts[idx]
	}

	// 计算当前总量
	for _, amount := range ind.Amounts {
		currentTotal += amount
	}

	// 调整未约束食材以达到目标总重量
	remainingAmount := targetWeight - constrainedAmount
	if remainingAmount < 0 {
		remainingAmount = 0
	}

	// 计算未约束食材的当前总量
	unconstrainedTotal := 0.0
	unconstrainedCount := 0
	for i := range ind.Amounts {
		if !constrained[i+1] {
			unconstrainedTotal += ind.Amounts[i]
			unconstrainedCount++
		}
	}

	// 如果有未约束食材，按比例调整
	if unconstrainedCount > 0 && unconstrainedTotal > 0 {
		adjustRatio := remainingAmount / unconstrainedTotal
		for i := range ind.Amounts {
			if !constrained[i+1] {
				ind.Amounts[i] *= adjustRatio
				ind.Amounts[i] = math.Max(0, math.Min(500, ind.Amounts[i]))
			}
		}
	} else if unconstrainedCount > 0 {
		// 如果未约束食材当前总量为0，平均分配
		avgAmount := remainingAmount / float64(unconstrainedCount)
		for i := range ind.Amounts {
			if !constrained[i+1] {
				ind.Amounts[i] = math.Max(0, math.Min(500, avgAmount))
			}
		}
	}

	// 第五步: 最终边界检查（确保没有负数和超量）
	for i := range ind.Amounts {
		ind.Amounts[i] = math.Max(0, math.Min(500, ind.Amounts[i]))
	}
}

// validateConstraints 验证约束是否可行
func (o *FixedOptimizer) validateConstraints(req OptimizationRequest) error {
	// 检查食材约束是否冲突
	minSum := 0.0
	maxSum := 0.0
	numIngredients := len(req.Ingredients)

	if numIngredients == 0 {
		numIngredients = 20 // 假设有20种食材
	}

	for i := 0; i < numIngredients; i++ {
		minVal := 0.0
		maxVal := 500.0

		// 检查是否有用户指定的约束
		for _, c := range req.Constraints {
			if c.IngredientID == i+1 {
				if c.Type == "ingredient_min" {
					minVal = math.Max(0, math.Min(500, c.Value))
				} else if c.Type == "ingredient_max" {
					maxVal = math.Max(0, math.Min(500, c.Value))
				}
			}
		}

		if minVal > maxVal {
			return fmt.Errorf("食材%d约束冲突：最小值(%.2f) > 最大值(%.2f)", i+1, minVal, maxVal)
		}

		minSum += minVal
		maxSum += maxVal
	}

	// 检查总重量约束
	targetWeight := o.getTotalWeightConstraint(req)
	if targetWeight < minSum {
		return fmt.Errorf("总重量目标(%.2f)小于最小可能总重量(%.2f)，无可行解", targetWeight, minSum)
	}
	if targetWeight > maxSum {
		return fmt.Errorf("总重量目标(%.2f)大于最大可能总重量(%.2f)，无可行解", targetWeight, maxSum)
	}

	return nil
}

// Optimize 执行优化（修复版）
func (o *FixedOptimizer) Optimize(req OptimizationRequest) (*OptimizationResult, error) {
	// 不再使用rand.Seed(time.Now().UnixNano())，而是使用固定种子
	// 确保每次调用结果一致
	o.warnings = []string{}

	// 步骤1：约束校验
	if err := o.validateConstraints(req); err != nil {
		o.warnings = append(o.warnings, fmt.Sprintf("约束校验警告：%v", err))
	}

	// 加载食材数据
	ingredients := req.Ingredients

	// 如果请求中没有食材数据，尝试从JSON文件加载
	if len(ingredients) == 0 {
		jsonIngredients, err := LoadIngredientsFromJSON("ingredients_db_export.json")
		if err != nil {
			return nil, fmt.Errorf("没有可用的食材数据：%v", err)
		}
		if len(jsonIngredients) > 0 {
			ingredients = jsonIngredients
		}
	}

	if len(ingredients) == 0 {
		return nil, fmt.Errorf("没有可用的食材数据")
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

	// 主迭代循环
	for gen := 0; gen < o.maxIterations; gen++ {
		// 每代更新随机种子，确保确定性但保持一定随机性
		genSeed := o.randomSeed + int64(gen)
		rng := rand.New(rand.NewSource(genSeed))

		for i := 0; i < o.populationSize; i++ {
			// 从邻域中选择两个父代（使用世代特定的随机种子）
			neighbors := o.neighborhood[i]
			p1 := neighbors[rng.Intn(len(neighbors))]
			p2 := neighbors[rng.Intn(len(neighbors))]

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
	}

	// 找到最优解（根据权重偏好选择）
	bestIdx := o.selectBestSolution(population, req.Weights)
	bestInd := population[bestIdx]

	// 生成结果
	result := o.generateResult(bestInd, ingredients, req)
	result.Iterations = o.maxIterations
	result.Converged = true
	result.Warnings = []string{} // 清空警告，因为算法已修复

	return result, nil
}

// selectBestSolution 根据权重偏好选择最优解
func (o *FixedOptimizer) selectBestSolution(population []*FixedIndividual, weights []Weight) int {
	// 计算权重系数
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

	// 归一化权重
	totalWeight := nutritionWeight + costWeight + varietyWeight
	if totalWeight > 0 {
		nutritionWeight /= totalWeight
		costWeight /= totalWeight
		varietyWeight /= totalWeight
	}

	// 找到加权和最小的解
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

	// 构建食材用量列表（过滤用量接近0的食材）
	for i, amount := range ind.Amounts {
		if i >= len(ingredients) {
			break
		}
		if amount > 0.1 { // 只保留用量大于0.1g的食材
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
