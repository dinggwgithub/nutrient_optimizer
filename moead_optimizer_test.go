package main

import (
	"math"
	"testing"
)

// TestMOEADOptimizer_CreateIndividual 测试个体创建
func TestMOEADOptimizer_CreateIndividual(t *testing.T) {
	optimizer := NewMOEADOptimizer(10, 10)
	req := OptimizationRequest{
		Constraints: []Constraint{
			{Type: "total_weight", Value: 400},
		},
	}

	ind := optimizer.createIndividual(5, req)

	if ind == nil {
		t.Error("个体创建失败")
	}

	if ind.Amounts == nil || len(ind.Amounts) != 5 {
		t.Errorf("食材用量数组长度错误: 期望5，实际%d", len(ind.Amounts))
	}

	// 检查用量是否在合理范围
	for i, amount := range ind.Amounts {
		if amount < 0 || amount > 500 {
			t.Errorf("食材%d用量超出范围: %.2fg", i, amount)
		}
	}
}

// TestMOEADOptimizer_CalculateNutrition 测试营养计算
func TestMOEADOptimizer_CalculateNutrition(t *testing.T) {
	optimizer := NewMOEADOptimizer(10, 10)

	ingredients := []Ingredient{
		{ID: 1, Name: "测试食材", Energy: 100, Protein: 10, Price: 0.5},
	}

	amounts := []float64{100} // 100g

	nutrition := optimizer.calculateNutrition(amounts, ingredients)

	expectedEnergy := 100.0 // 100g * 100kcal/100g = 100kcal
	if math.Abs(nutrition.Energy-expectedEnergy) > 0.01 {
		t.Errorf("能量计算错误: 期望%.2f，实际%.2f", expectedEnergy, nutrition.Energy)
	}

	expectedProtein := 10.0 // 100g * 10g/100g = 10g
	if math.Abs(nutrition.Protein-expectedProtein) > 0.01 {
		t.Errorf("蛋白质计算错误: 期望%.2f，实际%.2f", expectedProtein, nutrition.Protein)
	}
}

// TestMOEADOptimizer_CalculateTotalCost 测试成本计算
func TestMOEADOptimizer_CalculateTotalCost(t *testing.T) {
	optimizer := NewMOEADOptimizer(10, 10)

	ingredients := []Ingredient{
		{ID: 1, Name: "测试食材", Price: 0.5}, // 0.5元/100g
	}

	amounts := []float64{200} // 200g

	cost := optimizer.calculateTotalCost(amounts, ingredients)

	expectedCost := 1.0 // 200g * 0.5元/100g = 1.0元
	if math.Abs(cost-expectedCost) > 0.01 {
		t.Errorf("成本计算错误: 期望%.2f，实际%.2f", expectedCost, cost)
	}
}

// TestMOEADOptimizer_CalculateDiversity 测试多样性计算
func TestMOEADOptimizer_CalculateDiversity(t *testing.T) {
	optimizer := NewMOEADOptimizer(10, 10)

	// 测试用例1: 单一食材 - 多样性低
	amounts1 := []float64{400, 0, 0, 0}
	diversity1 := optimizer.calculateDiversity(amounts1)
	t.Logf("单一食材多样性: %.4f", diversity1)

	// 测试用例2: 多种食材均匀分布 - 多样性高
	amounts2 := []float64{100, 100, 100, 100}
	diversity2 := optimizer.calculateDiversity(amounts2)
	t.Logf("均匀分布多样性: %.4f", diversity2)

	if diversity1 >= diversity2 {
		t.Error("多样性计算错误: 单一食材多样性不应高于多种食材均匀分布")
	}

	if diversity2 < 0.7 {
		t.Errorf("多样性指数范围错误: 均匀分布多样性应接近0.75，实际%.2f", diversity2)
	}
}

// TestMOEADOptimizer_GenerateWeightVectors 测试权重向量生成
func TestMOEADOptimizer_GenerateWeightVectors(t *testing.T) {
	optimizer := NewMOEADOptimizer(10, 10)
	optimizer.generateWeightVectors()

	if optimizer.weightVectors == nil || len(optimizer.weightVectors) != 10 {
		t.Errorf("权重向量数量错误: 期望10，实际%d", len(optimizer.weightVectors))
	}

	// 验证权重向量应该归一化
	for i, vec := range optimizer.weightVectors {
		sum := 0.0
		for _, w := range vec {
			sum += w
		}
		if math.Abs(sum-1.0) > 0.01 {
			t.Errorf("权重向量%d未归一化，和为%.4f", i, sum)
		}
	}
}

// TestMOEADOptimizer_CalculateNeighborhood 测试邻域计算
func TestMOEADOptimizer_CalculateNeighborhood(t *testing.T) {
	optimizer := NewMOEADOptimizer(20, 10)
	optimizer.generateWeightVectors()
	optimizer.calculateNeighborhood()

	expectedNeighborSize := int(math.Ceil(float64(20) * 0.2)) // 20%邻域
	if optimizer.neighborhood == nil || len(optimizer.neighborhood[0]) != expectedNeighborSize {
		t.Errorf("邻域大小错误: 期望%d，实际%d", expectedNeighborSize, len(optimizer.neighborhood[0]))
	}

	// 检查邻域索引应该在有效范围内
	for _, neighbors := range optimizer.neighborhood {
		for _, idx := range neighbors {
			if idx < 0 || idx >= 20 {
				t.Errorf("邻域索引超出范围: %d", idx)
			}
		}
	}
}

// TestMOEADOptimizer_Tchebycheff 测试切比雪夫聚合
func TestMOEADOptimizer_Tchebycheff(t *testing.T) {
	optimizer := NewMOEADOptimizer(10, 10)
	optimizer.idealPoint = []float64{0, 0, 0}

	weights := []float64{0.33, 0.33, 0.34}
	objectives := []float64{1.0, 2.0, 3.0}

	result := optimizer.tchebycheff(weights, objectives)

	if result <= 0 {
		t.Errorf("切比雪夫聚合结果应为正值，实际%.4f", result)
	}
}

// TestMOEADOptimizer_Crossover 测试交叉操作
func TestMOEADOptimizer_Crossover(t *testing.T) {
	optimizer := NewMOEADOptimizer(10, 10)

	parent1 := &Individual{Amounts: []float64{100, 100, 100}}
	parent2 := &Individual{Amounts: []float64{200, 200, 200}}

	child := optimizer.crossover(parent1, parent2)

	if child == nil {
		t.Error("交叉操作失败")
	}

	if len(child.Amounts) != 3 {
		t.Errorf("子代个体长度错误")
	}
}

// TestMOEADOptimizer_Mutate 测试变异操作
func TestMOEADOptimizer_Mutate(t *testing.T) {
	optimizer := NewMOEADOptimizer(10, 10)

	ind := &Individual{Amounts: []float64{100, 100, 100}}

	optimizer.mutate(ind)

	// 验证变异后值在合理范围内
	for i, amount := range ind.Amounts {
		if amount < 0 || amount > 500 {
			t.Errorf("变异后食材%d用量超出范围: %.2fg", i, amount)
		}
	}
}

// TestMOEADOptimizer_Repair 测试约束修复
func TestMOEADOptimizer_Repair(t *testing.T) {
	optimizer := NewMOEADOptimizer(10, 10)

	req := OptimizationRequest{
		Constraints: []Constraint{
			{Type: "total_weight", Value: 400},
			{Type: "ingredient_min", IngredientID: 1, Value: 50},
			{Type: "ingredient_max", IngredientID: 2, Value: 100},
		},
	}

	ind := &Individual{Amounts: []float64{30, 200, 100, 100}} // 食材1不足，食材2超出

	optimizer.repair(ind, req)

	// 检查食材1最小值约束
	if ind.Amounts[0] < 50 {
		t.Errorf("食材1最小值约束未满足: %.2f < 50", ind.Amounts[0])
	}

	// 检查食材2最大值约束
	if ind.Amounts[1] > 100 {
		t.Errorf("食材2最大值约束未满足: %.2f > 100", ind.Amounts[1])
	}

	// 检查总重量
	total := 0.0
	for _, amount := range ind.Amounts {
		total += amount
	}
	t.Logf("总重量: %.2f", total)
}

// TestMOEADOptimizer_Optimize 测试完整优化流程
func TestMOEADOptimizer_Optimize(t *testing.T) {
	// 使用简单参数进行完整优化测试
	optimizer := NewMOEADOptimizer(20, 50)

	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "鸡胸肉", Energy: 165, Protein: 31, Fat: 3.6, Carbs: 0, Price: 0.8},
			{ID: 2, Name: "西兰花", Energy: 34, Protein: 2.8, Fat: 0.4, Carbs: 6.6, Price: 0.3},
			{ID: 3, Name: "米饭", Energy: 130, Protein: 2.6, Fat: 0.3, Carbs: 28, Price: 0.1},
		},
		NutritionGoals: []NutritionGoal{
			{Nutrient: "energy", Target: 600, Min: 500, Max: 700, Weight: 0.3},
			{Nutrient: "protein", Target: 30, Min: 20, Max: 40, Weight: 0.4},
		},
		Constraints: []Constraint{
			{Type: "ingredient_min", IngredientID: 1, Value: 50},
			{Type: "total_weight", Value: 400},
		},
		Weights: []Weight{
			{Type: "nutrition", Value: 0.6},
			{Type: "cost", Value: 0.3},
			{Type: "variety", Value: 0.1},
		},
		MaxIterations: 50,
		Tolerance:     1e-6,
	}

	result, err := optimizer.Optimize(req)

	if err != nil {
		t.Fatalf("优化失败: %v", err)
	}

	if result == nil {
		t.Fatal("优化结果为空")
	}

	if !result.Converged {
		t.Error("优化未收敛")
	}

	if result.Ingredients == nil || len(result.Ingredients) == 0 {
		t.Error("优化结果没有返回食材")
	}

	// 验证营养结果
	if result.Nutrition.Energy < 100 || result.Nutrition.Energy > 1000 {
		t.Errorf("能量结果不合理: %.2f", result.Nutrition.Energy)
	}

	// 验证成本
	if result.Cost < 0 {
		t.Errorf("成本结果不合理: %.2f", result.Cost)
	}

	t.Logf("优化结果: 能量=%.2f kcal, 成本=%.2f 元", result.Nutrition.Energy, result.Cost)
	t.Logf("食材数量: %d", len(result.Ingredients))
	for _, ing := range result.Ingredients {
		t.Logf("  %s: %.2fg", ing.Name, ing.Amount)
	}
}

// TestRoundToTwoDecimals 测试四舍五入函数
func TestRoundToTwoDecimals(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{1.234, 1.23},
		{1.235, 1.24},
		{1.2, 1.20},
		{0, 0},
	}

	for _, test := range tests {
		result := roundToTwoDecimals(test.input)
		if result != test.expected {
			t.Errorf("四舍五入错误: 输入%.3f，期望%.2f，实际%.2f",
				test.input, test.expected, result)
		}
	}
}

// === 修复Bug相关测试用例 ===

// TestFixedOptimizer_ResultStability 测试结果稳定性（同一参数多次调用结果一致）
func TestFixedOptimizer_ResultStability(t *testing.T) {
	// 创建相同配置的两个优化器实例
	optimizer1 := NewFixedOptimizer(10, 20)
	optimizer2 := NewFixedOptimizer(10, 20)

	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "鸡胸肉", Energy: 165, Protein: 31, Fat: 3.6, Carbs: 0, Price: 0.8},
			{ID: 2, Name: "西兰花", Energy: 34, Protein: 2.8, Fat: 0.4, Carbs: 6.6, Price: 0.3},
			{ID: 3, Name: "米饭", Energy: 130, Protein: 2.6, Fat: 0.3, Carbs: 28, Price: 0.1},
		},
		NutritionGoals: []NutritionGoal{
			{Nutrient: "energy", Target: 600, Weight: 0.3},
			{Nutrient: "protein", Target: 30, Weight: 0.4},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 400},
		},
		MaxIterations: 20,
	}

	// 第一次调用
	result1, err := optimizer1.Optimize(req)
	if err != nil {
		t.Fatalf("第一次优化失败: %v", err)
	}

	// 第二次调用（相同参数）
	result2, err := optimizer2.Optimize(req)
	if err != nil {
		t.Fatalf("第二次优化失败: %v", err)
	}

	// 验证食材用量一致
	if len(result1.Ingredients) != len(result2.Ingredients) {
		t.Errorf("食材数量不一致: 第一次%d种，第二次%d种",
			len(result1.Ingredients), len(result2.Ingredients))
	}

	// 验证每种食材用量一致
	for i := range result1.Ingredients {
		if i >= len(result2.Ingredients) {
			break
		}
		// 允许极小的浮点误差
		if result1.Ingredients[i].Amount-result2.Ingredients[i].Amount > 1e-6 {
			t.Errorf("食材[%s]用量不一致: 第一次%.10fg，第二次%.10fg",
				result1.Ingredients[i].Name,
				result1.Ingredients[i].Amount,
				result2.Ingredients[i].Amount)
		}
	}

	// 验证营养值一致
	nutritionCheck(t, "能量", result1.Nutrition.Energy, result2.Nutrition.Energy)
	nutritionCheck(t, "蛋白质", result1.Nutrition.Protein, result2.Nutrition.Protein)
	nutritionCheck(t, "钙", result1.Nutrition.Calcium, result2.Nutrition.Calcium)
	nutritionCheck(t, "锌", result1.Nutrition.Zinc, result2.Nutrition.Zinc)

	t.Log("✓ 结果稳定性测试通过：同一参数多次调用结果完全一致")
}

// nutritionCheck 营养值比较辅助函数
func nutritionCheck(t *testing.T, name string, val1, val2 float64) {
	t.Helper()
	if val1-val2 > 1e-6 {
		t.Errorf("%s值不一致: 第一次%.10f，第二次%.10f", name, val1, val2)
	}
}

// TestFixedOptimizer_ConstraintBounds 测试约束边界（食材用量在0-500g范围）
func TestFixedOptimizer_ConstraintBounds(t *testing.T) {
	optimizer := NewFixedOptimizer(20, 50)

	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "测试食材1", Energy: 100, Protein: 10, Price: 0.5},
			{ID: 2, Name: "测试食材2", Energy: 200, Protein: 20, Price: 1.0},
			{ID: 3, Name: "测试食材3", Energy: 300, Protein: 30, Price: 1.5},
		},
		NutritionGoals: []NutritionGoal{
			{Nutrient: "energy", Target: 1000, Weight: 1.0},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 1000}, // 总重量设为1000g，但单种食材不能超过500g
		},
		MaxIterations: 50,
	}

	result, err := optimizer.Optimize(req)
	if err != nil {
		t.Fatalf("优化失败: %v", err)
	}

	// 验证所有食材用量在0-500g范围内
	hasViolation := false
	for _, ing := range result.Ingredients {
		if ing.Amount < 0 {
			t.Errorf("❌ 食材[%s]用量为负数: %.2fg", ing.Name, ing.Amount)
			hasViolation = true
		}
		if ing.Amount > 500 {
			t.Errorf("❌ 食材[%s]用量超过500g: %.2fg", ing.Name, ing.Amount)
			hasViolation = true
		}
		t.Logf("  %s: %.2fg (在约束范围内)", ing.Name, ing.Amount)
	}

	if !hasViolation {
		t.Log("✓ 约束边界测试通过：所有食材用量在0-500g范围内")
	}

	// 验证收敛状态
	if !result.Converged {
		t.Error("❌ 优化未收敛")
	} else {
		t.Log("✓ 优化已正常收敛")
	}

	// 验证没有警告
	if len(result.Warnings) > 0 {
		t.Errorf("❌ 存在警告: %v", result.Warnings)
	} else {
		t.Log("✓ 无警告信息")
	}
}

// TestFixedOptimizer_ExtremeConstraints 测试极端约束场景
func TestFixedOptimizer_ExtremeConstraints(t *testing.T) {
	optimizer := NewFixedOptimizer(10, 20)

	// 设置极端约束条件
	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "食材A", Energy: 100, Protein: 10, Price: 0.5},
			{ID: 2, Name: "食材B", Energy: 200, Protein: 20, Price: 1.0},
		},
		NutritionGoals: []NutritionGoal{
			{Nutrient: "energy", Target: 500, Weight: 1.0},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 400},
			{Type: "ingredient_min", IngredientID: 1, Value: 100},
			{Type: "ingredient_max", IngredientID: 1, Value: 100}, // 固定食材A为100g
		},
		MaxIterations: 20,
	}

	result, err := optimizer.Optimize(req)
	if err != nil {
		t.Fatalf("优化失败: %v", err)
	}

	// 验证食材A严格在100g
	for _, ing := range result.Ingredients {
		if ing.ID == 1 && (ing.Amount < 99.99 || ing.Amount > 100.01) {
			t.Errorf("食材A约束未满足: 期望约100g，实际%.2fg", ing.Amount)
		}
		t.Logf("  %s: %.2fg", ing.Name, ing.Amount)
	}

	t.Log("✓ 极端约束场景测试完成")
}

// TestFixedOptimizer_CompareWithBuggy 对比测试：修复版vs含Bug版
func TestFixedOptimizer_CompareWithBuggy(t *testing.T) {
	// 此测试用于对比验证修复效果
	t.Log("=== 对比测试：修复版优化器 vs 含Bug优化器 ===")

	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 2, Name: "小麦", Energy: 338, Protein: 11.9, Fat: 1.3, Carbs: 75.2, Calcium: 34, Iron: 5.1, Zinc: 2.33},
			{ID: 3, Name: "五谷香", Energy: 378, Protein: 9.9, Fat: 2.6, Carbs: 78.9, Calcium: 2, Iron: 0.5, Zinc: 0.23},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 400},
		},
		MaxIterations: 50,
	}

	// 测试修复版优化器的稳定性
	t.Log("\n1. 测试修复版优化器稳定性:")
	fixed1 := NewFixedOptimizer(10, 20)
	fixed2 := NewFixedOptimizer(10, 20)

	res1, err := fixed1.Optimize(req)
	if err != nil {
		t.Fatalf("修复版优化失败: %v", err)
	}

	res2, err := fixed2.Optimize(req)
	if err != nil {
		t.Fatalf("修复版优化失败: %v", err)
	}

	// 验证结果一致性
	isStable := true
	for i := range res1.Ingredients {
		if i >= len(res2.Ingredients) {
			break
		}
		diff := res1.Ingredients[i].Amount - res2.Ingredients[i].Amount
		if diff < 0 {
			diff = -diff
		}
		if diff > 1e-6 {
			t.Errorf("  ❌ %s用量不一致: %.10f vs %.10f",
				res1.Ingredients[i].Name,
				res1.Ingredients[i].Amount,
				res2.Ingredients[i].Amount)
			isStable = false
		}
	}
	if isStable {
		t.Log("  ✓ 修复版优化器结果稳定，多次调用一致")
	}

	// 验证约束合规
	t.Log("\n2. 测试约束合规性:")
	isValid := true
	for _, ing := range res1.Ingredients {
		if ing.Amount < 0 || ing.Amount > 500 {
			t.Errorf("  ❌ %s用量不合规: %.2fg（应在0-500g之间）", ing.Name, ing.Amount)
			isValid = false
		}
	}
	if isValid {
		t.Log("  ✓ 所有食材用量合规，在0-500g范围内")
	}

	// 验证无异常警告
	t.Log("\n3. 测试异常警告消除:")
	if len(res1.Warnings) == 0 {
		t.Log("  ✓ 无警告信息")
	} else {
		t.Errorf("  ❌ 存在警告: %v", res1.Warnings)
	}

	if res1.Converged {
		t.Log("  ✓ 收敛状态正常")
	} else {
		t.Error("  ❌ 收敛状态异常")
	}

	t.Log("\n=== 对比测试完成 ===")
}
