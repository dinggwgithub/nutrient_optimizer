package main

import (
	"math"
	"testing"
)

// TestFixedOptimizer_Stability 测试修复后的结果稳定性
// 验证：相同输入多次调用，食材用量、营养值完全一致
func TestFixedOptimizer_Stability(t *testing.T) {
	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 2, Name: "小麦", Energy: 338, Protein: 11.9, Fat: 1.3, Carbs: 75.2, Calcium: 34, Iron: 5.1, Zinc: 2.33, VitaminC: 0, Price: 0},
			{ID: 3, Name: "五谷香", Energy: 378, Protein: 9.9, Fat: 2.6, Carbs: 78.9, Calcium: 2, Iron: 0.5, Zinc: 0.23, VitaminC: 0, Price: 0},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 400},
		},
	}

	optimizer := NewFixedOptimizer(50, 100)

	// 第一次调用
	result1, err := optimizer.Optimize(req)
	if err != nil {
		t.Fatalf("第一次优化失败: %v", err)
	}

	// 第二次调用
	result2, err := optimizer.Optimize(req)
	if err != nil {
		t.Fatalf("第二次优化失败: %v", err)
	}

	// 第三次调用
	result3, err := optimizer.Optimize(req)
	if err != nil {
		t.Fatalf("第三次优化失败: %v", err)
	}

	// 验证食材用量一致性
	if len(result1.Ingredients) != len(result2.Ingredients) || len(result1.Ingredients) != len(result3.Ingredients) {
		t.Error("三次调用返回的食材数量不一致")
	}

	for i := range result1.Ingredients {
		if i >= len(result2.Ingredients) || i >= len(result3.Ingredients) {
			break
		}

		// 验证食材用量完全一致
		if math.Abs(result1.Ingredients[i].Amount-result2.Ingredients[i].Amount) > 1e-10 {
			t.Errorf("食材%d用量不一致: 第一次=%.2f, 第二次=%.2f",
				i, result1.Ingredients[i].Amount, result2.Ingredients[i].Amount)
		}

		if math.Abs(result1.Ingredients[i].Amount-result3.Ingredients[i].Amount) > 1e-10 {
			t.Errorf("食材%d用量不一致: 第一次=%.2f, 第三次=%.2f",
				i, result1.Ingredients[i].Amount, result3.Ingredients[i].Amount)
		}
	}

	// 验证营养值一致性
	if math.Abs(result1.Nutrition.Energy-result2.Nutrition.Energy) > 1e-10 {
		t.Errorf("能量不一致: 第一次=%.2f, 第二次=%.2f", result1.Nutrition.Energy, result2.Nutrition.Energy)
	}

	if math.Abs(result1.Nutrition.Protein-result2.Nutrition.Protein) > 1e-10 {
		t.Errorf("蛋白质不一致: 第一次=%.2f, 第二次=%.2f", result1.Nutrition.Protein, result2.Nutrition.Protein)
	}

	if math.Abs(result1.Nutrition.Zinc-result2.Nutrition.Zinc) > 1e-10 {
		t.Errorf("锌不一致: 第一次=%.2f, 第二次=%.2f", result1.Nutrition.Zinc, result2.Nutrition.Zinc)
	}

	t.Logf("✅ 结果稳定性验证通过：三次调用结果完全一致")
	t.Logf("   小麦用量: %.2fg", result1.Ingredients[0].Amount)
	t.Logf("   五谷香用量: %.2fg", result1.Ingredients[1].Amount)
	t.Logf("   能量: %.2f kcal", result1.Nutrition.Energy)
	t.Logf("   锌: %.2f mg", result1.Nutrition.Zinc)
}

// TestFixedOptimizer_ConstraintBounds 测试约束边界
// 验证：食材用量在0-500g区间（无负数/超大值）
func TestFixedOptimizer_ConstraintBounds(t *testing.T) {
	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "鸡胸肉", Energy: 165, Protein: 31, Fat: 3.6, Carbs: 0, Price: 0.8},
			{ID: 2, Name: "西兰花", Energy: 34, Protein: 2.8, Fat: 0.4, Carbs: 6.6, Price: 0.3},
			{ID: 3, Name: "米饭", Energy: 130, Protein: 2.6, Fat: 0.3, Carbs: 28, Price: 0.1},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 400},
		},
	}

	optimizer := NewFixedOptimizer(50, 100)
	result, err := optimizer.Optimize(req)
	if err != nil {
		t.Fatalf("优化失败: %v", err)
	}

	// 验证所有食材用量在0-500g范围内
	for _, ing := range result.Ingredients {
		if ing.Amount < 0 {
			t.Errorf("食材%s用量为负数: %.2fg", ing.Name, ing.Amount)
		}
		if ing.Amount > 500 {
			t.Errorf("食材%s用量超过500g: %.2fg", ing.Name, ing.Amount)
		}
	}

	t.Logf("✅ 约束边界验证通过：所有食材用量在0-500g范围内")
	for _, ing := range result.Ingredients {
		t.Logf("   %s: %.2fg", ing.Name, ing.Amount)
	}
}

// TestFixedOptimizer_NoWarnings 测试修复后无警告
// 验证：warnings字段清空，收敛状态（converged=true）稳定
func TestFixedOptimizer_NoWarnings(t *testing.T) {
	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "鸡胸肉", Energy: 165, Protein: 31, Fat: 3.6, Carbs: 0, Price: 0.8},
			{ID: 2, Name: "西兰花", Energy: 34, Protein: 2.8, Fat: 0.4, Carbs: 6.6, Price: 0.3},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 400},
		},
	}

	optimizer := NewFixedOptimizer(50, 100)
	result, err := optimizer.Optimize(req)
	if err != nil {
		t.Fatalf("优化失败: %v", err)
	}

	// 验证收敛状态
	if !result.Converged {
		t.Error("优化未收敛")
	}

	// 验证warnings为空
	if len(result.Warnings) > 0 {
		t.Errorf("修复后不应有警告，但得到: %v", result.Warnings)
	}

	// 验证优化器警告为空
	if len(optimizer.GetWarnings()) > 0 {
		t.Errorf("优化器不应有警告，但得到: %v", optimizer.GetWarnings())
	}

	t.Logf("✅ 异常消除验证通过：warnings字段清空，converged=true")
}

// TestFixedOptimizer_RepairNegativeAmount 测试修复负数用量
func TestFixedOptimizer_RepairNegativeAmount(t *testing.T) {
	optimizer := NewFixedOptimizer(50, 100)

	req := OptimizationRequest{
		Constraints: []Constraint{
			{Type: "total_weight", Value: 400},
		},
	}

	ind := &Individual{Amounts: []float64{-50, 200, 300}} // 包含负数用量

	optimizer.repair(ind, req)

	// 验证负数被修复为0
	if ind.Amounts[0] < 0 {
		t.Errorf("负数用量未被修复: %.2f", ind.Amounts[0])
	}

	// 验证所有用量在0-500g范围内
	for i, amount := range ind.Amounts {
		if amount < 0 || amount > 500 {
			t.Errorf("食材%d用量超出范围: %.2fg", i, amount)
		}
	}

	t.Logf("✅ 负数用量修复验证通过")
	t.Logf("   修复后用量: %.2f, %.2f, %.2f", ind.Amounts[0], ind.Amounts[1], ind.Amounts[2])
}

// TestFixedOptimizer_RepairExcessAmount 测试修复超大用量
func TestFixedOptimizer_RepairExcessAmount(t *testing.T) {
	optimizer := NewFixedOptimizer(50, 100)

	req := OptimizationRequest{
		Constraints: []Constraint{
			{Type: "total_weight", Value: 400},
		},
	}

	ind := &Individual{Amounts: []float64{100, 600, 100}} // 包含超过500g的用量

	optimizer.repair(ind, req)

	// 验证超大值被修复到500g以内
	if ind.Amounts[1] > 500 {
		t.Errorf("超大用量未被修复: %.2f", ind.Amounts[1])
	}

	// 验证所有用量在0-500g范围内
	for i, amount := range ind.Amounts {
		if amount < 0 || amount > 500 {
			t.Errorf("食材%d用量超出范围: %.2fg", i, amount)
		}
	}

	t.Logf("✅ 超大用量修复验证通过")
	t.Logf("   修复后用量: %.2f, %.2f, %.2f", ind.Amounts[0], ind.Amounts[1], ind.Amounts[2])
}

// TestFixedOptimizer_FixedSeed 测试固定随机种子
func TestFixedOptimizer_FixedSeed(t *testing.T) {
	optimizer1 := NewFixedOptimizer(50, 100)
	optimizer2 := NewFixedOptimizer(50, 100)

	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "鸡胸肉", Energy: 165, Protein: 31, Fat: 3.6, Carbs: 0, Price: 0.8},
			{ID: 2, Name: "西兰花", Energy: 34, Protein: 2.8, Fat: 0.4, Carbs: 6.6, Price: 0.3},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 400},
		},
	}

	result1, err := optimizer1.Optimize(req)
	if err != nil {
		t.Fatalf("优化1失败: %v", err)
	}

	result2, err := optimizer2.Optimize(req)
	if err != nil {
		t.Fatalf("优化2失败: %v", err)
	}

	// 验证两个优化器结果一致（因为使用了相同的固定种子）
	if len(result1.Ingredients) != len(result2.Ingredients) {
		t.Error("两个优化器返回的食材数量不一致")
	}

	for i := range result1.Ingredients {
		if i >= len(result2.Ingredients) {
			break
		}
		if math.Abs(result1.Ingredients[i].Amount-result2.Ingredients[i].Amount) > 1e-10 {
			t.Errorf("两个优化器食材%d用量不一致: %.2f vs %.2f",
				i, result1.Ingredients[i].Amount, result2.Ingredients[i].Amount)
		}
	}

	t.Logf("✅ 固定随机种子验证通过：不同优化器实例结果一致")
}

// TestFixedOptimizer_CompleteFlow 测试完整优化流程
func TestFixedOptimizer_CompleteFlow(t *testing.T) {
	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "鸡胸肉", Energy: 165, Protein: 31, Fat: 3.6, Carbs: 0, Calcium: 11, Iron: 1, Zinc: 0.7, VitaminC: 0, Price: 0.8},
			{ID: 2, Name: "西兰花", Energy: 34, Protein: 2.8, Fat: 0.4, Carbs: 6.6, Calcium: 47, Iron: 0.6, Zinc: 0.4, VitaminC: 89, Price: 0.3},
			{ID: 3, Name: "米饭", Energy: 130, Protein: 2.6, Fat: 0.3, Carbs: 28, Calcium: 10, Iron: 0.2, Zinc: 0.1, VitaminC: 0, Price: 0.1},
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
	}

	optimizer := NewFixedOptimizer(50, 100)
	result, err := optimizer.Optimize(req)
	if err != nil {
		t.Fatalf("优化失败: %v", err)
	}

	// 验证结果有效性
	if result == nil {
		t.Fatal("优化结果为空")
	}

	if !result.Converged {
		t.Error("优化未收敛")
	}

	if len(result.Ingredients) == 0 {
		t.Error("优化结果没有返回食材")
	}

	// 验证营养结果合理性
	if result.Nutrition.Energy < 100 || result.Nutrition.Energy > 1000 {
		t.Errorf("能量结果不合理: %.2f", result.Nutrition.Energy)
	}

	// 验证成本非负
	if result.Cost < 0 {
		t.Errorf("成本为负数: %.2f", result.Cost)
	}

	t.Logf("✅ 完整优化流程验证通过")
	t.Logf("   优化结果: 能量=%.2f kcal, 成本=%.2f 元", result.Nutrition.Energy, result.Cost)
	t.Logf("   食材数量: %d", len(result.Ingredients))
	for _, ing := range result.Ingredients {
		t.Logf("   %s: %.2fg", ing.Name, ing.Amount)
	}
}
