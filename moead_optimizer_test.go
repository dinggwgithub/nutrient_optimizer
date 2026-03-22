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

// TestFixedOptimizer_IngredientWeightConstraint 测试食材重量约束0-500g
func TestFixedOptimizer_IngredientWeightConstraint(t *testing.T) {
	optimizer := NewFixedOptimizer(30, 100)

	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "小麦", Energy: 338, Protein: 11.9, Fat: 1.3, Carbs: 75.2, Price: 0},
			{ID: 2, Name: "五谷香", Energy: 378, Protein: 9.9, Fat: 2.6, Carbs: 78.9, Price: 0},
		},
		NutritionGoals: []NutritionGoal{
			{Nutrient: "energy", Target: 600, Weight: 0.4},
			{Nutrient: "protein", Target: 30, Weight: 0.3},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 400},
		},
		MaxIterations: 100,
	}

	result, err := optimizer.Optimize(req)
	if err != nil {
		t.Fatalf("优化失败: %v", err)
	}

	// 验证食材用量在0-500g范围内
	for _, ing := range result.Ingredients {
		if ing.Amount < 0 || ing.Amount > 500 {
			t.Errorf("食材用量超出范围: %s = %.2fg (应在0-500g之间)", ing.Name, ing.Amount)
		}
	}

	t.Logf("食材数量: %d", len(result.Ingredients))
	for _, ing := range result.Ingredients {
		t.Logf("  %s: %.2fg", ing.Name, ing.Amount)
	}
}

// TestFixedOptimizer_Extreme1000gWheat 测试1000g小麦超限场景
func TestFixedOptimizer_Extreme1000gWheat(t *testing.T) {
	optimizer := NewFixedOptimizer(30, 100)

	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 2, Name: "小麦", Energy: 338, Protein: 11.9, Fat: 1.3, Carbs: 75.2, Calcium: 34, Iron: 5.1, Zinc: 2.33, VitaminC: 0, Price: 0},
			{ID: 3, Name: "五谷香", Energy: 378, Protein: 9.9, Fat: 2.6, Carbs: 78.9, Calcium: 2, Iron: 0.5, Zinc: 0.23, VitaminC: 0, Price: 0},
		},
		NutritionGoals: []NutritionGoal{
			{Nutrient: "energy", Target: 1500, Weight: 0.4},
			{Nutrient: "protein", Target: 50, Weight: 0.3},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 400},
		},
		MaxIterations: 100,
	}

	result, err := optimizer.Optimize(req)
	if err != nil {
		t.Fatalf("优化失败: %v", err)
	}

	if !result.Converged {
		t.Error("优化未收敛")
	}

	// 验证食材用量约束
	for _, ing := range result.Ingredients {
		if ing.Amount < 0 || ing.Amount > 500 {
			t.Errorf("食材用量超出范围: %s = %.2fg", ing.Name, ing.Amount)
		}
	}

	// 验证没有错误
	if result.Error != "" {
		t.Errorf("优化结果包含错误: %s", result.Error)
	}

	t.Logf("收敛状态: %v", result.Converged)
	t.Logf("食材数量: %d", len(result.Ingredients))
	for _, ing := range result.Ingredients {
		t.Logf("  %s: %.2fg", ing.Name, ing.Amount)
	}
}

// TestFixedOptimizer_BoundaryValues 测试边界值 0g/500g
func TestFixedOptimizer_BoundaryValues(t *testing.T) {
	optimizer := NewFixedOptimizer(30, 100)

	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "测试食材A", Energy: 100, Protein: 10, Price: 0.5},
			{ID: 2, Name: "测试食材B", Energy: 200, Protein: 20, Price: 1.0},
			{ID: 3, Name: "测试食材C", Energy: 300, Protein: 30, Price: 1.5},
		},
		NutritionGoals: []NutritionGoal{
			{Nutrient: "energy", Target: 800, Weight: 0.5},
			{Nutrient: "protein", Target: 60, Weight: 0.5},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 400},
			{Type: "ingredient_min", IngredientID: 1, Value: 0},   // 0g边界
			{Type: "ingredient_max", IngredientID: 2, Value: 500}, // 500g边界
		},
		MaxIterations: 100,
	}

	result, err := optimizer.Optimize(req)
	if err != nil {
		t.Fatalf("优化失败: %v", err)
	}

	if !result.Converged {
		t.Error("优化未收敛")
	}

	// 验证约束
	for _, ing := range result.Ingredients {
		if ing.Amount < 0 || ing.Amount > 500 {
			t.Errorf("食材用量超出范围: %s = %.2fg", ing.Name, ing.Amount)
		}
	}

	t.Logf("收敛状态: %v", result.Converged)
	for _, ing := range result.Ingredients {
		t.Logf("  %s: %.2fg", ing.Name, ing.Amount)
	}
}

// TestFixedOptimizer_CompareWithBuggy 对比测试：修复版vsBug版
func TestFixedOptimizer_CompareWithBuggy(t *testing.T) {
	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 2, Name: "小麦", Energy: 338, Protein: 11.9, Fat: 1.3, Carbs: 75.2, Calcium: 34, Iron: 5.1, Zinc: 2.33, VitaminC: 0, Price: 0},
			{ID: 3, Name: "五谷香", Energy: 378, Protein: 9.9, Fat: 2.6, Carbs: 78.9, Calcium: 2, Iron: 0.5, Zinc: 0.23, VitaminC: 0, Price: 0},
		},
		NutritionGoals: []NutritionGoal{
			{Nutrient: "energy", Target: 1500, Weight: 0.4},
			{Nutrient: "protein", Target: 50, Weight: 0.3},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 400},
		},
		MaxIterations: 100,
	}

	// 测试Buggy版本
	buggyOptimizer := NewBuggyOptimizer(BugTypeConvergenceFailure)
	buggyResult, _ := buggyOptimizer.Optimize(req)

	t.Log("=== Buggy 版本结果 ===")
	t.Logf("收敛状态: %v", buggyResult.Converged)
	t.Logf("错误信息: %s", buggyResult.Error)
	for _, ing := range buggyResult.Ingredients {
		t.Logf("  %s: %.2fg", ing.Name, ing.Amount)
	}
	for _, w := range buggyResult.Warnings {
		t.Logf("  警告: %s", w)
	}

	// 测试修复版本
	fixedOptimizer := NewFixedOptimizer(30, 100)
	fixedResult, err := fixedOptimizer.Optimize(req)
	if err != nil {
		t.Fatalf("修复版优化失败: %v", err)
	}

	t.Log("\n=== 修复版本结果 ===")
	t.Logf("收敛状态: %v", fixedResult.Converged)
	t.Logf("错误信息: %s", fixedResult.Error)
	for _, ing := range fixedResult.Ingredients {
		t.Logf("  %s: %.2fg", ing.Name, ing.Amount)
	}
	for _, w := range fixedResult.Warnings {
		t.Logf("  警告: %s", w)
	}

	// 验证修复效果
	if !fixedResult.Converged {
		t.Error("修复版应该收敛，但实际未收敛")
	}

	if fixedResult.Error != "" {
		t.Errorf("修复版不应该有错误，但得到: %s", fixedResult.Error)
	}

	// 验证食材用量约束
	for _, ing := range fixedResult.Ingredients {
		if ing.Amount < 0 || ing.Amount > 500 {
			t.Errorf("修复版食材用量超出范围: %s = %.2fg", ing.Name, ing.Amount)
		}
	}
}
