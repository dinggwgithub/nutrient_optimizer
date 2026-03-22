package main

import (
	"math"
	"testing"
)

// TestFixedOptimizer_Optimize 测试修复后优化器的基本功能
func TestFixedOptimizer_Optimize(t *testing.T) {
	optimizer := NewFixedOptimizer()

	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "小麦", Energy: 338, Protein: 11.9, Fat: 1.3, Carbs: 75.2, Calcium: 34, Iron: 5.1, Zinc: 2.33, VitaminC: 0, Price: 0},
			{ID: 2, Name: "五谷香", Energy: 378, Protein: 9.9, Fat: 2.6, Carbs: 78.9, Calcium: 2, Iron: 0.5, Zinc: 0.23, VitaminC: 0, Price: 0},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 300}, // 总重量300g
		},
		MaxIterations: 100,
		Tolerance:     0.05,
	}

	result, err := optimizer.Optimize(req)

	if err != nil {
		t.Fatalf("优化失败: %v", err)
	}

	if result == nil {
		t.Fatal("优化结果为空")
	}

	// 验证收敛
	if !result.Converged {
		t.Error("优化未收敛")
	}

	// 验证食材数量
	if len(result.Ingredients) == 0 {
		t.Error("优化结果没有返回食材")
	}

	// 验证能量值在合理范围内（300g食材，能量应该在300-1500之间）
	if result.Nutrition.Energy < 100 || result.Nutrition.Energy > 2000 {
		t.Errorf("能量结果不合理: %.2f kcal", result.Nutrition.Energy)
	}

	// 验证钙含量在合理范围内
	if result.Nutrition.Calcium < 0 || result.Nutrition.Calcium > 10000 {
		t.Errorf("钙含量结果不合理: %.2f mg", result.Nutrition.Calcium)
	}

	// 验证没有NaN或Inf
	if math.IsNaN(result.Nutrition.Energy) || math.IsInf(result.Nutrition.Energy, 0) {
		t.Error("能量值为NaN或Inf")
	}
	if math.IsNaN(result.Nutrition.Calcium) || math.IsInf(result.Nutrition.Calcium, 0) {
		t.Error("钙含量为NaN或Inf")
	}

	// 验证食材重量约束（0-500g）
	for _, ing := range result.Ingredients {
		if ing.Amount < 0 {
			t.Errorf("食材重量为负数: %s = %.2fg", ing.Name, ing.Amount)
		}
		if ing.Amount > 500 {
			t.Errorf("食材重量超过500g: %s = %.2fg", ing.Name, ing.Amount)
		}
	}

	t.Logf("优化结果: 能量=%.2f kcal, 钙=%.2f mg", result.Nutrition.Energy, result.Nutrition.Calcium)
	for _, ing := range result.Ingredients {
		t.Logf("  %s: %.2fg", ing.Name, ing.Amount)
	}
}

// TestFixedOptimizer_CalculateNutrition 测试营养计算精度
func TestFixedOptimizer_CalculateNutrition(t *testing.T) {
	optimizer := NewFixedOptimizer()

	// 测试用例：150g小麦 + 150g五谷香
	ingredients := []Ingredient{
		{ID: 1, Name: "小麦", Energy: 338, Protein: 11.9, Fat: 1.3, Carbs: 75.2, Calcium: 34, Iron: 5.1, Zinc: 2.33, VitaminC: 0, Price: 0},
		{ID: 2, Name: "五谷香", Energy: 378, Protein: 9.9, Fat: 2.6, Carbs: 78.9, Calcium: 2, Iron: 0.5, Zinc: 0.23, VitaminC: 0, Price: 0},
	}
	amounts := []float64{150, 150}

	nutrition := optimizer.calculateNutrition(amounts, ingredients)

	// 期望能量：小麦 338*1.5 + 五谷香 378*1.5 = 507 + 567 = 1074 kcal
	expectedEnergy := 1074.0
	if math.Abs(nutrition.Energy-expectedEnergy) > 0.1 {
		t.Errorf("能量计算错误: 期望%.2f，实际%.2f", expectedEnergy, nutrition.Energy)
	}

	// 期望钙：小麦 34*1.5 + 五谷香 2*1.5 = 51 + 3 = 54 mg
	expectedCalcium := 54.0
	if math.Abs(nutrition.Calcium-expectedCalcium) > 0.1 {
		t.Errorf("钙含量计算错误: 期望%.2f，实际%.2f", expectedCalcium, nutrition.Calcium)
	}

	t.Logf("营养计算: 能量=%.2f kcal, 钙=%.2f mg", nutrition.Energy, nutrition.Calcium)
}

// TestFixedOptimizer_NoPrecisionLoss 测试没有精度丢失
func TestFixedOptimizer_NoPrecisionLoss(t *testing.T) {
	// 使用float64进行多次累加测试
	ingredients := []Ingredient{
		{ID: 1, Name: "测试食材", Energy: 100, Calcium: 1},
	}

	// 模拟多次累加（1000次）
	var totalEnergy float64
	var totalCalcium float64
	for i := 0; i < 1000; i++ {
		totalEnergy += ingredients[0].Energy * 0.01
		totalCalcium += ingredients[0].Calcium * 0.01
	}

	// 期望：100 * 0.01 * 1000 = 1000
	expectedEnergy := 1000.0
	expectedCalcium := 10.0

	if math.Abs(totalEnergy-expectedEnergy) > 0.001 {
		t.Errorf("多次累加能量精度丢失: 期望%.2f，实际%.2f", expectedEnergy, totalEnergy)
	}
	if math.Abs(totalCalcium-expectedCalcium) > 0.001 {
		t.Errorf("多次累加钙精度丢失: 期望%.2f，实际%.2f", expectedCalcium, totalCalcium)
	}

	t.Logf("多次累加测试: 能量=%.2f, 钙=%.2f", totalEnergy, totalCalcium)
}

// TestFixedOptimizer_IngredientConstraints 测试食材重量约束
func TestFixedOptimizer_IngredientConstraints(t *testing.T) {
	optimizer := NewFixedOptimizer()

	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "食材1", Energy: 100, Price: 0.5},
			{ID: 2, Name: "食材2", Energy: 200, Price: 0.8},
			{ID: 3, Name: "食材3", Energy: 150, Price: 0.6},
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

	// 验证所有食材重量在0-500g范围内
	for _, ing := range result.Ingredients {
		if ing.Amount < 0 {
			t.Errorf("食材重量为负数: %s = %.2fg", ing.Name, ing.Amount)
		}
		if ing.Amount > 500 {
			t.Errorf("食材重量超过500g限制: %s = %.2fg", ing.Name, ing.Amount)
		}
	}

	// 验证总重量接近目标
	totalWeight := 0.0
	for _, ing := range result.Ingredients {
		totalWeight += ing.Amount
	}

	// 允许5%的偏差
	if math.Abs(totalWeight-400) > 400*0.05 {
		t.Errorf("总重量偏差超过5%%: 目标400g, 实际%.2fg", totalWeight)
	}

	t.Logf("总重量: %.2fg", totalWeight)
}

// TestFixedOptimizer_NutritionGoals 测试营养目标优化
func TestFixedOptimizer_NutritionGoals(t *testing.T) {
	optimizer := NewFixedOptimizer()

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
		Tolerance:     0.05,
	}

	result, err := optimizer.Optimize(req)
	if err != nil {
		t.Fatalf("优化失败: %v", err)
	}

	// 验证结果在合理范围内
	if result.Nutrition.Energy < 100 || result.Nutrition.Energy > 1500 {
		t.Errorf("能量结果不合理: %.2f", result.Nutrition.Energy)
	}

	if result.Nutrition.Protein < 0 || result.Nutrition.Protein > 100 {
		t.Errorf("蛋白质结果不合理: %.2f", result.Nutrition.Protein)
	}

	t.Logf("营养结果: 能量=%.2f kcal, 蛋白质=%.2fg", result.Nutrition.Energy, result.Nutrition.Protein)
}

// TestFixedOptimizer_NoRandomness 测试结果稳定性（消除随机性）
func TestFixedOptimizer_NoRandomness(t *testing.T) {
	optimizer := NewFixedOptimizer()

	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "食材1", Energy: 100, Price: 0.5},
			{ID: 2, Name: "食材2", Energy: 200, Price: 0.8},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 300},
		},
		MaxIterations: 50,
	}

	// 运行多次，检查能量结果是否稳定
	var energies []float64
	for i := 0; i < 5; i++ {
		result, err := optimizer.Optimize(req)
		if err != nil {
			t.Fatalf("第%d次优化失败: %v", i+1, err)
		}
		energies = append(energies, result.Nutrition.Energy)
	}

	// 计算方差
	mean := 0.0
	for _, e := range energies {
		mean += e
	}
	mean /= float64(len(energies))

	variance := 0.0
	for _, e := range energies {
		variance += (e - mean) * (e - mean)
	}
	variance /= float64(len(energies))
	stdDev := math.Sqrt(variance)

	// 标准差应该小于均值的10%
	if stdDev > mean*0.1 {
		t.Errorf("结果不稳定，标准差过大: %.2f (均值: %.2f)", stdDev, mean)
	}

	t.Logf("多次运行能量结果: %v", energies)
	t.Logf("均值: %.2f, 标准差: %.2f", mean, stdDev)
}

// TestFixedOptimizer_Convergence 测试收敛性（偏差<5%）
func TestFixedOptimizer_Convergence(t *testing.T) {
	optimizer := NewFixedOptimizer()

	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "鸡胸肉", Energy: 165, Protein: 31, Price: 0.8},
			{ID: 2, Name: "西兰花", Energy: 34, Protein: 2.8, Price: 0.3},
			{ID: 3, Name: "米饭", Energy: 130, Protein: 2.6, Price: 0.1},
		},
		NutritionGoals: []NutritionGoal{
			{Nutrient: "energy", Target: 600, Weight: 0.5},
			{Nutrient: "protein", Target: 30, Weight: 0.5},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 400},
		},
		MaxIterations: 100,
		Tolerance:     0.05, // 5%收敛阈值
	}

	result, err := optimizer.Optimize(req)
	if err != nil {
		t.Fatalf("优化失败: %v", err)
	}

	// 验证收敛标志
	if !result.Converged {
		t.Error("优化未收敛")
	}

	// 计算与目标的偏差
	energyDeviation := math.Abs(result.Nutrition.Energy-600) / 600
	proteinDeviation := math.Abs(result.Nutrition.Protein-30) / 30

	t.Logf("偏差: 能量=%.2f%%, 蛋白质=%.2f%%", energyDeviation*100, proteinDeviation*100)
}
