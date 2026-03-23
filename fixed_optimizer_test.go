package main

import (
	"math"
	"testing"
)

func TestFixedOptimizer_ResultStability(t *testing.T) {
	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 2, Name: "小麦", Energy: 338, Protein: 11.9, Fat: 1.3, Carbs: 75.2, Calcium: 34, Iron: 5.1, Zinc: 2.33, VitaminC: 0, Price: 0},
			{ID: 3, Name: "五谷香", Energy: 378, Protein: 9.9, Fat: 2.6, Carbs: 78.9, Calcium: 2, Iron: 0.5, Zinc: 0.23, VitaminC: 0, Price: 0},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 300},
		},
	}

	optimizer := NewFixedOptimizer()

	result1, err1 := optimizer.Optimize(req)
	if err1 != nil {
		t.Fatalf("第一次优化失败: %v", err1)
	}

	result2, err2 := optimizer.Optimize(req)
	if err2 != nil {
		t.Fatalf("第二次优化失败: %v", err2)
	}

	result3, err3 := optimizer.Optimize(req)
	if err3 != nil {
		t.Fatalf("第三次优化失败: %v", err3)
	}

	for i := range result1.Ingredients {
		if math.Abs(result1.Ingredients[i].Amount-result2.Ingredients[i].Amount) > 0.001 {
			t.Errorf("结果不稳定: 食材%s第一次用量%.6f，第二次用量%.6f",
				result1.Ingredients[i].Name, result1.Ingredients[i].Amount, result2.Ingredients[i].Amount)
		}
		if math.Abs(result2.Ingredients[i].Amount-result3.Ingredients[i].Amount) > 0.001 {
			t.Errorf("结果不稳定: 食材%s第二次用量%.6f，第三次用量%.6f",
				result2.Ingredients[i].Name, result2.Ingredients[i].Amount, result3.Ingredients[i].Amount)
		}
	}

	if math.Abs(result1.Nutrition.Zinc-result2.Nutrition.Zinc) > 0.001 {
		t.Errorf("营养值不稳定: 锌第一次%.6f，第二次%.6f",
			result1.Nutrition.Zinc, result2.Nutrition.Zinc)
	}

	t.Logf("稳定性测试通过: 三次调用结果完全一致")
	t.Logf("  小麦用量: %.2fg", result1.Ingredients[0].Amount)
	t.Logf("  五谷香用量: %.2fg", result1.Ingredients[1].Amount)
	t.Logf("  锌含量: %.2fmg", result1.Nutrition.Zinc)
}

func TestFixedOptimizer_ConstraintBounds(t *testing.T) {
	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "鸡胸肉", Energy: 165, Protein: 31, Fat: 3.6, Carbs: 0, Calcium: 11, Iron: 1, Zinc: 0.7, VitaminC: 0, Price: 0.8},
			{ID: 2, Name: "西兰花", Energy: 34, Protein: 2.8, Fat: 0.4, Carbs: 6.6, Calcium: 67, Iron: 1, Zinc: 0.4, VitaminC: 51, Price: 0.3},
			{ID: 3, Name: "米饭", Energy: 130, Protein: 2.6, Fat: 0.3, Carbs: 28, Calcium: 7, Iron: 0.2, Zinc: 0.5, VitaminC: 0, Price: 0.1},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 400},
		},
	}

	optimizer := NewFixedOptimizer()
	result, err := optimizer.Optimize(req)
	if err != nil {
		t.Fatalf("优化失败: %v", err)
	}

	for _, ing := range result.Ingredients {
		if ing.Amount < MinIngredientAmount {
			t.Errorf("约束越界: 食材%s用量%.2fg为负数", ing.Name, ing.Amount)
		}
		if ing.Amount > MaxIngredientAmount {
			t.Errorf("约束越界: 食材%s用量%.2fg超过500g上限", ing.Name, ing.Amount)
		}
	}

	t.Logf("约束边界测试通过: 所有食材用量在0-500g区间内")
	for _, ing := range result.Ingredients {
		t.Logf("  %s: %.2fg (范围: 0-500g)", ing.Name, ing.Amount)
	}
}

func TestFixedOptimizer_NoNegativeAmounts(t *testing.T) {
	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "测试食材1", Energy: 100, Protein: 10, Fat: 5, Carbs: 20, Calcium: 50, Iron: 2, Zinc: 1, VitaminC: 10, Price: 1.0},
			{ID: 2, Name: "测试食材2", Energy: 200, Protein: 15, Fat: 8, Carbs: 25, Calcium: 60, Iron: 3, Zinc: 2, VitaminC: 20, Price: 1.5},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 200},
		},
	}

	optimizer := NewFixedOptimizer()

	for i := 0; i < 10; i++ {
		result, err := optimizer.Optimize(req)
		if err != nil {
			t.Fatalf("第%d次优化失败: %v", i+1, err)
		}

		for _, ing := range result.Ingredients {
			if ing.Amount < 0 {
				t.Errorf("第%d次优化出现负数用量: 食材%s用量%.2fg", i+1, ing.Name, ing.Amount)
			}
		}
	}

	t.Logf("负数用量测试通过: 10次优化均无负数用量")
}

func TestFixedOptimizer_NoSuperLargeAmounts(t *testing.T) {
	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "食材A", Energy: 100, Protein: 10, Fat: 5, Carbs: 20, Calcium: 50, Iron: 2, Zinc: 1, VitaminC: 10, Price: 1.0},
			{ID: 2, Name: "食材B", Energy: 200, Protein: 15, Fat: 8, Carbs: 25, Calcium: 60, Iron: 3, Zinc: 2, VitaminC: 20, Price: 1.5},
			{ID: 3, Name: "食材C", Energy: 150, Protein: 12, Fat: 6, Carbs: 22, Calcium: 55, Iron: 2.5, Zinc: 1.5, VitaminC: 15, Price: 1.2},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 500},
		},
	}

	optimizer := NewFixedOptimizer()

	for i := 0; i < 10; i++ {
		result, err := optimizer.Optimize(req)
		if err != nil {
			t.Fatalf("第%d次优化失败: %v", i+1, err)
		}

		for _, ing := range result.Ingredients {
			if ing.Amount > MaxIngredientAmount {
				t.Errorf("第%d次优化出现超大用量: 食材%s用量%.2fg > 500g", i+1, ing.Name, ing.Amount)
			}
		}
	}

	t.Logf("超大用量测试通过: 10次优化均无超过500g的用量")
}

func TestFixedOptimizer_Converged(t *testing.T) {
	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "鸡胸肉", Energy: 165, Protein: 31, Fat: 3.6, Carbs: 0, Calcium: 11, Iron: 1, Zinc: 0.7, VitaminC: 0, Price: 0.8},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 200},
		},
	}

	optimizer := NewFixedOptimizer()
	result, err := optimizer.Optimize(req)
	if err != nil {
		t.Fatalf("优化失败: %v", err)
	}

	if !result.Converged {
		t.Error("优化未收敛")
	}

	if len(result.Warnings) > 0 {
		t.Logf("警告信息: %v", result.Warnings)
	}

	t.Logf("收敛状态测试通过: converged=%v", result.Converged)
}

func TestFixedOptimizer_NutritionCalculation(t *testing.T) {
	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "测试食材", Energy: 100, Protein: 10, Fat: 5, Carbs: 20, Calcium: 50, Iron: 2, Zinc: 1, VitaminC: 10, Price: 1.0},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 100},
		},
	}

	optimizer := NewFixedOptimizer()
	result, err := optimizer.Optimize(req)
	if err != nil {
		t.Fatalf("优化失败: %v", err)
	}

	expectedEnergy := 100.0
	if math.Abs(result.Nutrition.Energy-expectedEnergy) > 0.01 {
		t.Errorf("能量计算错误: 期望%.2f，实际%.2f", expectedEnergy, result.Nutrition.Energy)
	}

	expectedProtein := 10.0
	if math.Abs(result.Nutrition.Protein-expectedProtein) > 0.01 {
		t.Errorf("蛋白质计算错误: 期望%.2f，实际%.2f", expectedProtein, result.Nutrition.Protein)
	}

	t.Logf("营养计算测试通过: 能量=%.2f, 蛋白质=%.2f", result.Nutrition.Energy, result.Nutrition.Protein)
}

func TestFixedOptimizer_MultipleIngredients(t *testing.T) {
	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "鸡胸肉", Energy: 165, Protein: 31, Fat: 3.6, Carbs: 0, Calcium: 11, Iron: 1, Zinc: 0.7, VitaminC: 0, Price: 0.8},
			{ID: 2, Name: "西兰花", Energy: 34, Protein: 2.8, Fat: 0.4, Carbs: 6.6, Calcium: 67, Iron: 1, Zinc: 0.4, VitaminC: 51, Price: 0.3},
			{ID: 3, Name: "米饭", Energy: 130, Protein: 2.6, Fat: 0.3, Carbs: 28, Calcium: 7, Iron: 0.2, Zinc: 0.5, VitaminC: 0, Price: 0.1},
			{ID: 4, Name: "鸡蛋", Energy: 144, Protein: 13, Fat: 9.5, Carbs: 1.1, Calcium: 56, Iron: 1.8, Zinc: 1.1, VitaminC: 0, Price: 0.5},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 400},
		},
	}

	optimizer := NewFixedOptimizer()
	result, err := optimizer.Optimize(req)
	if err != nil {
		t.Fatalf("优化失败: %v", err)
	}

	totalAmount := 0.0
	for _, ing := range result.Ingredients {
		totalAmount += ing.Amount
	}

	if math.Abs(totalAmount-400) > 0.01 {
		t.Errorf("总重量约束未满足: 期望400g，实际%.2fg", totalAmount)
	}

	t.Logf("多食材测试通过: 总重量=%.2fg", totalAmount)
	for _, ing := range result.Ingredients {
		t.Logf("  %s: %.2fg", ing.Name, ing.Amount)
	}
}
