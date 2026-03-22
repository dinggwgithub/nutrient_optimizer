package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"sort"

	_ "github.com/go-sql-driver/mysql"
)

// MOEADOptimizer MOEA/D多目标优化算法实现
type MOEADOptimizer struct {
	db             *sql.DB     // 数据库连接
	populationSize int         // 种群大小
	maxIterations  int         // 最大迭代次数
	neighborSize   int         // 邻域大小
	weightVectors  [][]float64 // 权重向量
	neighborhood   [][]int     // 邻域索引
	idealPoint     []float64   // 理想点
	objectives     int         // 目标函数数量
	warnings       []string    // 警告信息
	fixedSeed      int64       // 固定随机种子
}

// NewMOEADOptimizer 创建MOEA/D优化器
func NewMOEADOptimizer(populationSize, maxIterations int) *MOEADOptimizer {
	return &MOEADOptimizer{
		populationSize: populationSize,
		maxIterations:  maxIterations,
		neighborSize:   int(math.Ceil(float64(populationSize) * 0.2)), // 邻域大小为种群的20%
		objectives:     3,                                             // 营养达标、成本、多样性三个目标
		warnings:       []string{},
		fixedSeed:      42, // 使用固定种子确保结果可重复
	}
}

// NewMOEADOptimizerWithSeed 创建带自定义种子的MOEA/D优化器
func NewMOEADOptimizerWithSeed(populationSize, maxIterations int, seed int64) *MOEADOptimizer {
	return &MOEADOptimizer{
		populationSize: populationSize,
		maxIterations:  maxIterations,
		neighborSize:   int(math.Ceil(float64(populationSize) * 0.2)),
		objectives:     3,
		warnings:       []string{},
		fixedSeed:      seed,
	}
}

// ConnectDB 连接数据库
func (o *MOEADOptimizer) ConnectDB() error {
	dsn := "root:123456@tcp(127.0.0.1:3306)/recipe_system?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	}
	o.db = db
	return nil
}

// CloseDB 关闭数据库连接
func (o *MOEADOptimizer) CloseDB() {
	if o.db != nil {
		o.db.Close()
	}
}

// LoadIngredientsFromDB 从数据库加载食材数据
func (o *MOEADOptimizer) LoadIngredientsFromDB(limit int) ([]Ingredient, error) {
	if o.db == nil {
		if err := o.ConnectDB(); err != nil {
			return nil, err
		}
	}

	query := `
	SELECT id, name, energy, protein, fat, carbs, calcium, iron, zinc, vitamin_c, price
	FROM ingredients 
	WHERE status = 'ACTIVE'
	LIMIT ?`

	rows, err := o.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ingredients []Ingredient
	for rows.Next() {
		var ing Ingredient
		err := rows.Scan(
			&ing.ID, &ing.Name, &ing.Energy, &ing.Protein, &ing.Fat, &ing.Carbs,
			&ing.Calcium, &ing.Iron, &ing.Zinc, &ing.VitaminC, &ing.Price,
		)
		if err != nil {
			return nil, err
		}
		ingredients = append(ingredients, ing)
	}

	return ingredients, nil
}

// LoadIngredientsFromJSON 从JSON文件加载食材数据
func (o *MOEADOptimizer) LoadIngredientsFromJSON(filepath string) ([]Ingredient, error) {
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
func (o *MOEADOptimizer) GetWarnings() []string {
	return o.warnings
}

// generateWeightVectors 生成权重向量
func (o *MOEADOptimizer) generateWeightVectors() {
	o.weightVectors = make([][]float64, o.populationSize)

	// 使用均匀分布生成权重向量
	for i := 0; i < o.populationSize; i++ {
		vector := make([]float64, o.objectives)
		sum := 0.0

		// 生成随机权重并归一化
		for j := 0; j < o.objectives; j++ {
			vector[j] = rand.Float64() + 0.001 // 避免零权重
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
func (o *MOEADOptimizer) calculateNeighborhood() {
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
func (o *MOEADOptimizer) euclideanDistance(a, b []float64) float64 {
	sum := 0.0
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return math.Sqrt(sum)
}

// initializeIdealPoint 初始化理想点
func (o *MOEADOptimizer) initializeIdealPoint() {
	o.idealPoint = make([]float64, o.objectives)
	for i := range o.idealPoint {
		o.idealPoint[i] = math.Inf(1)
	}
}

// updateIdealPoint 更新理想点
func (o *MOEADOptimizer) updateIdealPoint(objectives []float64) {
	for i := range objectives {
		if objectives[i] < o.idealPoint[i] {
			o.idealPoint[i] = objectives[i]
		}
	}
}

// tchebycheff 切比雪夫聚合函数
func (o *MOEADOptimizer) tchebycheff(weights, objectives []float64) float64 {
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

// Individual 表示种群中的个体
type Individual struct {
	Amounts    []float64 // 食材用量 (g)
	Objectives []float64 // 目标函数值 [营养偏差, 成本, 多样性]
}

// createIndividual 创建新个体
func (o *MOEADOptimizer) createIndividual(numIngredients int, req OptimizationRequest) *Individual {
	amounts := make([]float64, numIngredients)
	totalWeight := 0.0

	// 随机初始化食材用量
	for i := range amounts {
		// 每个食材初始用量在0-100g之间
		amounts[i] = rand.Float64() * 100
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

	return &Individual{Amounts: amounts}
}

// getTotalWeightConstraint 获取总重量约束
func (o *MOEADOptimizer) getTotalWeightConstraint(req OptimizationRequest) float64 {
	for _, c := range req.Constraints {
		if c.Type == "total_weight" {
			return c.Value
		}
	}
	return 400.0 // 默认值
}

// evaluateIndividual 评估个体的目标函数值
func (o *MOEADOptimizer) evaluateIndividual(ind *Individual, ingredients []Ingredient, req OptimizationRequest) {
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
func (o *MOEADOptimizer) calculateNutrition(amounts []float64, ingredients []Ingredient) NutritionSummary {
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
func (o *MOEADOptimizer) calculateNutritionDeviation(nutrition NutritionSummary, goals []NutritionGoal) float64 {
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
func (o *MOEADOptimizer) calculateTotalCost(amounts []float64, ingredients []Ingredient) float64 {
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
func (o *MOEADOptimizer) calculateDiversity(amounts []float64) float64 {
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

// crossover 模拟二进制交叉 (SBX)
func (o *MOEADOptimizer) crossover(parent1, parent2 *Individual) *Individual {
	numVars := len(parent1.Amounts)
	child := &Individual{Amounts: make([]float64, numVars)}

	eta := 1.0 // 分布指数
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
func (o *MOEADOptimizer) mutate(ind *Individual) {
	eta := 20.0                             // 分布指数
	prob := 1.0 / float64(len(ind.Amounts)) // 每个变量的变异概率

	for i := range ind.Amounts {
		if rand.Float64() < prob {
			x := ind.Amounts[i]
			delta1 := (x - 0) / 500.0   // 下界0
			delta2 := (500 - x) / 500.0 // 上界500

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
			// 确保在边界内
			ind.Amounts[i] = math.Max(0, math.Min(500, x))
		}
	}
}

// repair 修复约束违反
func (o *MOEADOptimizer) repair(ind *Individual, req OptimizationRequest) {
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
			ind.Amounts[i] = math.Max(0, math.Min(500, ind.Amounts[i]))
		}
	}

	// 第三步: 再次检查并修复食材约束（归一化后可能违反约束）
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

	// 第四步: 调整其他食材以满足总重量（在满足食材约束后）
	adjustedTotal := 0.0
	constrainedAmount := 0.0
	constrainedCount := 0

	// 计算已被约束固定的食材总量
	for _, c := range req.Constraints {
		if (c.Type == "ingredient_min" || c.Type == "ingredient_max") && c.IngredientID > 0 {
			idx := c.IngredientID - 1
			constrainedAmount += ind.Amounts[idx]
			constrainedCount++
		}
	}

	// 计算当前总量
	for _, amount := range ind.Amounts {
		adjustedTotal += amount
	}

	// 如果有未被约束的食材，调整它们以达到目标总重量
	unconstrainedCount := len(ind.Amounts) - constrainedCount
	if unconstrainedCount > 0 && adjustedTotal != targetWeight {
		remainingAmount := targetWeight - constrainedAmount
		if remainingAmount > 0 {
			// 计算未被约束食材的当前总量
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

			// 按比例调整未被约束的食材
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
						ind.Amounts[i] = math.Max(0, math.Min(500, ind.Amounts[i]))
					}
				}
			}
		}
	}
}

// Optimize 执行MOEA/D优化
func (o *MOEADOptimizer) Optimize(req OptimizationRequest) (*OptimizationResult, error) {
	// 使用固定种子确保结果可重复
	rand.Seed(o.fixedSeed)
	o.warnings = []string{}

	// 加载食材数据（优先级: 请求 > 数据库 > JSON文件）
	ingredients := req.Ingredients

	// 如果请求中没有食材数据，尝试从数据库加载
	if len(ingredients) == 0 {
		dbIngredients, err := o.LoadIngredientsFromDB(20)
		if err != nil {
			o.warnings = append(o.warnings, fmt.Sprintf("从数据库加载食材失败: %v，尝试从JSON文件加载", err))
			// 数据库失败，尝试从JSON文件加载
			jsonIngredients, err := o.LoadIngredientsFromJSON("ingredients_db_export.json")
			if err != nil {
				o.warnings = append(o.warnings, fmt.Sprintf("从JSON文件加载食材失败: %v", err))
			} else if len(jsonIngredients) > 0 {
				ingredients = jsonIngredients
			}
		} else if len(dbIngredients) > 0 {
			ingredients = dbIngredients
		}
	}

	if len(ingredients) == 0 {
		return nil, fmt.Errorf("没有可用的食材数据（数据库和JSON文件均不可用，请提供食材数据）")
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

	// 收敛跟踪
	prevBestScore := math.Inf(1)
	convergenceCount := 0
	tolerance := 0.05 // 5%偏差

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

		// 检查收敛
		bestIdx := o.selectBestSolution(population, req.Weights)
		currentScore := o.calculateWeightedScore(population[bestIdx], req.Weights)

		if math.Abs(currentScore-prevBestScore) < tolerance*math.Abs(prevBestScore) {
			convergenceCount++
			if convergenceCount >= 10 {
				o.warnings = append(o.warnings, fmt.Sprintf("算法在第%d代收敛", gen))
				break
			}
		} else {
			convergenceCount = 0
		}
		prevBestScore = currentScore
	}

	// 找到最优解（根据权重偏好选择）
	bestIdx := o.selectBestSolution(population, req.Weights)
	bestInd := population[bestIdx]

	// 生成结果
	result := o.generateResult(bestInd, ingredients, req)
	result.Iterations = o.maxIterations
	result.Converged = convergenceCount >= 10 || o.checkConvergence(result, req.NutritionGoals, tolerance)

	return result, nil
}

// calculateWeightedScore 计算加权得分
func (o *MOEADOptimizer) calculateWeightedScore(ind *Individual, weights []Weight) float64 {
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

	return ind.Objectives[0]*nutritionWeight +
		ind.Objectives[1]*costWeight +
		ind.Objectives[2]*varietyWeight
}

// checkConvergence 检查结果是否收敛到目标
func (o *MOEADOptimizer) checkConvergence(result *OptimizationResult, goals []NutritionGoal, tolerance float64) bool {
	if len(goals) == 0 {
		return true
	}

	for _, goal := range goals {
		var actual float64
		switch goal.Nutrient {
		case "energy":
			actual = result.Nutrition.Energy
		case "protein":
			actual = result.Nutrition.Protein
		case "fat":
			actual = result.Nutrition.Fat
		case "carbs":
			actual = result.Nutrition.Carbs
		case "calcium":
			actual = result.Nutrition.Calcium
		case "iron":
			actual = result.Nutrition.Iron
		case "zinc":
			actual = result.Nutrition.Zinc
		case "vitamin_c":
			actual = result.Nutrition.VitaminC
		default:
			continue
		}

		if goal.Target > 0 {
			relDiff := math.Abs(actual-goal.Target) / goal.Target
			if relDiff > tolerance {
				return false
			}
		}
	}

	return true
}

// selectBestSolution 根据权重偏好选择最优解
func (o *MOEADOptimizer) selectBestSolution(population []*Individual, weights []Weight) int {
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
func (o *MOEADOptimizer) generateResult(ind *Individual, ingredients []Ingredient, req OptimizationRequest) *OptimizationResult {
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
		// 确保食材重量在合理范围内 (0-500g)
		if amount < 0 {
			amount = 0
		}
		if amount > 500 {
			amount = 500
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

	// 验证并修复营养值
	result.Nutrition.Energy = validateNutritionValue(result.Nutrition.Energy, "能量", 0, 5000, &result.Warnings)
	result.Nutrition.Protein = validateNutritionValue(result.Nutrition.Protein, "蛋白质", 0, 200, &result.Warnings)
	result.Nutrition.Fat = validateNutritionValue(result.Nutrition.Fat, "脂肪", 0, 200, &result.Warnings)
	result.Nutrition.Carbs = validateNutritionValue(result.Nutrition.Carbs, "碳水化合物", 0, 500, &result.Warnings)
	result.Nutrition.Calcium = validateNutritionValue(result.Nutrition.Calcium, "钙", 0, 3000, &result.Warnings)
	result.Nutrition.Iron = validateNutritionValue(result.Nutrition.Iron, "铁", 0, 100, &result.Warnings)
	result.Nutrition.Zinc = validateNutritionValue(result.Nutrition.Zinc, "锌", 0, 50, &result.Warnings)
	result.Nutrition.VitaminC = validateNutritionValue(result.Nutrition.VitaminC, "维生素C", 0, 2000, &result.Warnings)

	// 计算成本
	result.Cost = roundToTwoDecimals(o.calculateTotalCost(ind.Amounts, ingredients))

	// 添加多样性警告
	diversity := 1.0 - ind.Objectives[2]
	if diversity < 0.3 {
		result.Warnings = append(result.Warnings, "食材多样性较低，建议增加食材种类")
	}

	return result
}

// validateNutritionValue 验证营养值是否在合理范围内
func validateNutritionValue(value float64, name string, minVal, maxVal float64, warnings *[]string) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		*warnings = append(*warnings, fmt.Sprintf("%s计算结果异常(NaN/Inf)，已修正为0", name))
		return 0
	}
	if value < minVal {
		*warnings = append(*warnings, fmt.Sprintf("%s值%.2f低于最小值%.2f，已修正", name, value, minVal))
		return minVal
	}
	if value > maxVal {
		*warnings = append(*warnings, fmt.Sprintf("%s值%.2f超过最大值%.2f，已修正", name, value, maxVal))
		return maxVal
	}
	return roundToTwoDecimals(value)
}

// roundToTwoDecimals 四舍五入到两位小数
func roundToTwoDecimals(x float64) float64 {
	return math.Round(x*100) / 100
}
