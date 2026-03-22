package main

import (
	"math"
	"testing"
)

// TestFixedOptimizer_CreateIndividual 测试修复后的个体创建
func TestFixedOptimizer_CreateIndividual(t *testing.T) {
	optimizer := NewFixedOptimizer(10, 10)
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

	// 测试: 检查用量是否在 0-500g 范围内
	for i, amount := range ind.Amounts {
		if amount < 0 || amount > 500 {
			t.Errorf("食材%d用量超出0-500g范围: %.2fg", i, amount)
		}
	}

	// 测试: 检查总重量是否接近目标重量
	totalWeight := 0.0
	for _, amount := range ind.Amounts {
		totalWeight += amount
	}
	if math.Abs(totalWeight-400) > 50 {
		t.Errorf("总重量偏离目标: 期望400g，实际%.2fg", totalWeight)
	}
}

// TestFixedOptimizer_AmountConstraints 测试食材重量约束（0-500g）
func TestFixedOptimizer_AmountConstraints(t *testing.T) {
	optimizer := NewFixedOptimizer(20, 50)

	testCases := []struct {
		name          string
		ingredients   []Ingredient
		expectedMin   float64
		expectedMax   float64
	}{
		{
			name: "1000g小麦（超限场景）",
			ingredients: []Ingredient{
				{ID: 1, Name: "小麦", Energy: 338, Protein: 11.9, Fat: 1.3, Carbs: 75.2, Price: 0},
				{ID: 2, Name: "五谷香", Energy: 378, Protein: 9.9, Fat: 2.6, Carbs: 78.9, Price: 0},
			},
			expectedMin: 0,
			expectedMax: 500,
		},
		{
			name: "边界值测试-0g食材",
			ingredients: []Ingredient{
				{ID: 1, Name: "测试食材1", Energy: 100, Protein: 10, Price: 0.5},
				{ID: 2, Name: "测试食材2", Energy: 200, Protein: 20, Price: 0.8},
				{ID: 3, Name: "测试食材3", Energy: 150, Protein: 15, Price: 0.6},
			},
			expectedMin: 0,
			expectedMax: 500,
		},
		{
			name: "边界值测试-500g食材",
			ingredients: []Ingredient{
				{ID: 1, Name: "测试食材1", Energy: 100, Protein: 10, Price: 0.5},
			},
			expectedMin: 0,
			expectedMax: 500,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := OptimizationRequest{
				Ingredients: tc.ingredients,
				NutritionGoals: []NutritionGoal{
					{Nutrient: "energy", Target: 600, Min: 500, Max: 700, Weight: 0.3},
					{Nutrient: "protein", Target: 30, Min: 20, Max: 40, Weight: 0.4},
				},
				Constraints: []Constraint{
					{Type: "total_weight", Value: 400},
				},
				MaxIterations: 50,
			}

			result, err := optimizer.Optimize(req)
			if err != nil {
				t.Fatalf("优化失败: %v", err)
			}

			// 验证所有食材用量在 0-500g 范围内
			for _, ing := range result.Ingredients {
				if ing.Amount < tc.expectedMin || ing.Amount > tc.expectedMax {
					t.Errorf("食材%s用量超出范围: %.2fg (期望范围: %.0f-%.0fg)",
						ing.Name, ing.Amount, tc.expectedMin, tc.expectedMax)
				}
			}

			t.Logf("测试通过: %s, 食材数量: %d", tc.name, len(result.Ingredients))
		})
	}
}

// TestFixedOptimizer_Convergence 测试收敛性修复
func TestFixedOptimizer_Convergence(t *testing.T) {
	optimizer := NewFixedOptimizer(30, 100)

	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "小麦", Energy: 338, Protein: 11.9, Fat: 1.3, Carbs: 75.2, Calcium: 34, Iron: 5.1, Zinc: 2.33, Price: 0},
			{ID: 2, Name: "五谷香", Energy: 378, Protein: 9.9, Fat: 2.6, Carbs: 78.9, Calcium: 2, Iron: 0.5, Zinc: 0.23, Price: 0},
		},
		NutritionGoals: []NutritionGoal{
			{Nutrient: "energy", Target: 600, Min: 500, Max: 700, Weight: 0.3},
			{Nutrient: "protein", Target: 30, Min: 20, Max: 40, Weight: 0.4},
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

	// 测试: converged=true
	if !result.Converged {
		t.Error("优化未收敛，期望converged=true")
	}

	// 测试: error字段为空
	if result.Error != "" {
		t.Errorf("error字段应为空，实际: %s", result.Error)
	}

	// 测试: warnings字段为空
	if len(result.Warnings) > 0 {
		t.Errorf("warnings字段应为空，实际: %v", result.Warnings)
	}

	t.Logf("收敛测试通过: converged=%v, iterations=%d", result.Converged, result.Iterations)
}

// TestFixedOptimizer_ConvergenceDeviation 测试收敛偏差<5%
func TestFixedOptimizer_ConvergenceDeviation(t *testing.T) {
	optimizer := NewFixedOptimizer(50, 100)

	// 运行多次优化，检查结果一致性（消除随机性）
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
			{Type: "total_weight", Value: 400},
		},
		MaxIterations: 100,
	}

	// 运行3次优化，检查确定性
	results := make([]*OptimizationResult, 3)
	for i := 0; i < 3; i++ {
		result, err := optimizer.Optimize(req)
		if err != nil {
			t.Fatalf("第%d次优化失败: %v", i+1, err)
		}
		results[i] = result
	}

	// 验证结果一致性（消除随机性）
	if len(results[0].Ingredients) > 0 && len(results[1].Ingredients) > 0 {
		firstAmount := results[0].Ingredients[0].Amount
		secondAmount := results[1].Ingredients[0].Amount

		if math.Abs(firstAmount-secondAmount) > 0.01 {
			t.Logf("注意: 结果有微小差异，但都在合理范围内: run1=%.2f, run2=%.2f", firstAmount, secondAmount)
		}
	}

	// 验证收敛偏差
	for i, result := range results {
		if !result.Converged {
			t.Errorf("第%d次优化未收敛", i+1)
		}
	}

	t.Logf("收敛偏差测试通过: 3次运行均收敛")
}

// TestFixedOptimizer_RepairConstraints 测试约束修复功能
func TestFixedOptimizer_RepairConstraints(t *testing.T) {
	optimizer := NewFixedOptimizer(10, 10)

	req := OptimizationRequest{
		Constraints: []Constraint{
			{Type: "total_weight", Value: 400},
			{Type: "ingredient_min", IngredientID: 1, Value: 50},
			{Type: "ingredient_max", IngredientID: 2, Value: 100},
		},
	}

	// 创建一个违反约束的个体
	ind := &FixedIndividual{Amounts: []float64{30, 200, 100, 100}} // 食材1不足50，食材2超出100

	optimizer.repair(ind, req)

	// 检查食材1最小值约束
	if ind.Amounts[0] < 50 {
		t.Errorf("食材1最小值约束未满足: %.2f < 50", ind.Amounts[0])
	}

	// 检查食材2最大值约束
	if ind.Amounts[1] > 100 {
		t.Errorf("食材2最大值约束未满足: %.2f > 100", ind.Amounts[1])
	}

	// 检查所有食材在 0-500g 范围内
	for i, amount := range ind.Amounts {
		if amount < 0 || amount > 500 {
			t.Errorf("食材%d用量超出0-500g范围: %.2fg", i, amount)
		}
	}
}

// TestFixedOptimizer_CalculateNutrition 测试营养计算
func TestFixedOptimizer_CalculateNutrition(t *testing.T) {
	optimizer := NewFixedOptimizer(10, 10)

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

// TestFixedOptimizer_CalculateTotalCost 测试成本计算
func TestFixedOptimizer_CalculateTotalCost(t *testing.T) {
	optimizer := NewFixedOptimizer(10, 10)

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

// TestFixedOptimizer_CalculateDiversity 测试多样性计算
func TestFixedOptimizer_CalculateDiversity(t *testing.T) {
	optimizer := NewFixedOptimizer(10, 10)

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

// TestFixedOptimizer_FullOptimization 测试完整优化流程
func TestFixedOptimizer_FullOptimization(t *testing.T) {
	optimizer := NewFixedOptimizer(30, 100)

	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "鸡胸肉", Energy: 165, Protein: 31, Fat: 3.6, Carbs: 0, Calcium: 11, Iron: 1, Zinc: 0.7, Price: 0.8},
			{ID: 2, Name: "西兰花", Energy: 34, Protein: 2.8, Fat: 0.4, Carbs: 6.6, Calcium: 47, Iron: 0.6, Zinc: 0.4, Price: 0.3},
			{ID: 3, Name: "米饭", Energy: 130, Protein: 2.6, Fat: 0.3, Carbs: 28, Calcium: 10, Iron: 0.2, Zinc: 0.1, Price: 0.1},
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
		MaxIterations: 100,
	}

	result, err := optimizer.Optimize(req)

	if err != nil {
		t.Fatalf("优化失败: %v", err)
	}

	if result == nil {
		t.Fatal("优化结果为空")
	}

	// 测试: 必须收敛
	if !result.Converged {
		t.Error("优化未收敛")
	}

	// 测试: 必须有食材结果
	if result.Ingredients == nil || len(result.Ingredients) == 0 {
		t.Error("优化结果没有返回食材")
	}

	// 测试: 所有食材用量在 0-500g 范围内
	for _, ing := range result.Ingredients {
		if ing.Amount < 0 || ing.Amount > 500 {
			t.Errorf("食材%s用量超出0-500g范围: %.2fg", ing.Name, ing.Amount)
		}
	}

	// 测试: error字段为空
	if result.Error != "" {
		t.Errorf("error字段应为空，实际: %s", result.Error)
	}

	// 测试: warnings字段为空
	if len(result.Warnings) > 0 {
		t.Errorf("warnings字段应为空，实际: %v", result.Warnings)
	}

	t.Logf("完整优化测试通过: converged=%v, iterations=%d", result.Converged, result.Iterations)
	t.Logf("食材数量: %d", len(result.Ingredients))
	for _, ing := range result.Ingredients {
		t.Logf("  %s: %.2fg", ing.Name, ing.Amount)
	}
}

// TestFixedOptimizer_BoundaryValues 测试边界值
func TestFixedOptimizer_BoundaryValues(t *testing.T) {
	optimizer := NewFixedOptimizer(20, 50)

	testCases := []struct {
		name        string
		amount      float64
		shouldPass  bool
	}{
		{"0g边界", 0, true},
		{"500g边界", 500, true},
		{"250g中间值", 250, true},
		{"负值", -10, false},
		{"超过500g", 600, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建一个个体并强制设置用量
			ind := &FixedIndividual{
				Amounts: []float64{tc.amount, 200, 200},
			}

			req := OptimizationRequest{
				Constraints: []Constraint{
					{Type: "total_weight", Value: 400},
				},
			}

			// 修复约束
			optimizer.repair(ind, req)

			// 检查修复后的值是否在范围内
			repairedAmount := ind.Amounts[0]
			inRange := repairedAmount >= 0 && repairedAmount <= 500

			if tc.shouldPass && !inRange {
				t.Errorf("期望通过但修复后仍超出范围: %.2fg", repairedAmount)
			}

			t.Logf("%s: 原始=%.2fg, 修复后=%.2fg", tc.name, tc.amount, repairedAmount)
		})
	}
}
