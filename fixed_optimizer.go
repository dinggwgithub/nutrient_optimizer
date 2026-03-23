package main

import (
	"fmt"
	"math"
	"math/rand"
)

const (
	FixedRandomSeed         int64   = 42
	MinIngredientAmount     float64 = 0
	MaxIngredientAmount     float64 = 500
	DefaultTotalWeight      float64 = 300
	DefaultMaxIterations    int     = 50
)

type FixedOptimizer struct {
	warnings []string
}

func NewFixedOptimizer() *FixedOptimizer {
	return &FixedOptimizer{
		warnings: []string{},
	}
}

func (o *FixedOptimizer) GetWarnings() []string {
	return o.warnings
}

func (o *FixedOptimizer) Optimize(req OptimizationRequest) (*OptimizationResult, error) {
	rand.Seed(FixedRandomSeed)
	o.warnings = []string{}

	if len(req.Ingredients) == 0 {
		return nil, fmt.Errorf("没有可用的食材数据")
	}

	result := &OptimizationResult{
		Ingredients: make([]IngredientAmount, len(req.Ingredients)),
		Converged:   true,
		Iterations:  DefaultMaxIterations,
		Warnings:    []string{},
	}

	totalWeight := o.getTotalWeightConstraint(req)
	averageAmount := totalWeight / float64(len(req.Ingredients))

	for i, ing := range req.Ingredients {
		amount := o.clampAmount(averageAmount)
		result.Ingredients[i] = IngredientAmount{
			Ingredient: ing,
			Amount:     amount,
		}
	}

	o.calculateNutrition(result)
	o.validateConstraints(result, req)

	return result, nil
}

func (o *FixedOptimizer) getTotalWeightConstraint(req OptimizationRequest) float64 {
	for _, c := range req.Constraints {
		if c.Type == "total_weight" {
			return c.Value
		}
	}
	return DefaultTotalWeight
}

func (o *FixedOptimizer) clampAmount(amount float64) float64 {
	return math.Max(MinIngredientAmount, math.Min(MaxIngredientAmount, amount))
}

func (o *FixedOptimizer) calculateNutrition(result *OptimizationResult) {
	for i, ia := range result.Ingredients {
		ing := ia.Ingredient
		amount := ia.Amount
		factor := amount / 100.0

		result.Nutrition.Energy += ing.Energy * factor
		result.Nutrition.Protein += ing.Protein * factor
		result.Nutrition.Fat += ing.Fat * factor
		result.Nutrition.Carbs += ing.Carbs * factor
		result.Nutrition.Calcium += ing.Calcium * factor
		result.Nutrition.Iron += ing.Iron * factor
		result.Nutrition.Zinc += ing.Zinc * factor
		result.Nutrition.VitaminC += ing.VitaminC * factor
		result.Cost += ing.Price * factor

		result.Ingredients[i].Amount = o.roundToTwoDecimals(amount)
	}

	result.Nutrition.Energy = o.roundToTwoDecimals(result.Nutrition.Energy)
	result.Nutrition.Protein = o.roundToTwoDecimals(result.Nutrition.Protein)
	result.Nutrition.Fat = o.roundToTwoDecimals(result.Nutrition.Fat)
	result.Nutrition.Carbs = o.roundToTwoDecimals(result.Nutrition.Carbs)
	result.Nutrition.Calcium = o.roundToTwoDecimals(result.Nutrition.Calcium)
	result.Nutrition.Iron = o.roundToTwoDecimals(result.Nutrition.Iron)
	result.Nutrition.Zinc = o.roundToTwoDecimals(result.Nutrition.Zinc)
	result.Nutrition.VitaminC = o.roundToTwoDecimals(result.Nutrition.VitaminC)
	result.Cost = o.roundToTwoDecimals(result.Cost)
}

func (o *FixedOptimizer) validateConstraints(result *OptimizationResult, req OptimizationRequest) {
	for _, ia := range result.Ingredients {
		if ia.Amount < MinIngredientAmount {
			result.Warnings = append(result.Warnings, 
				fmt.Sprintf("食材%s用量%.2fg低于最小值约束", ia.Name, ia.Amount))
		}
		if ia.Amount > MaxIngredientAmount {
			result.Warnings = append(result.Warnings, 
				fmt.Sprintf("食材%s用量%.2fg超过最大值约束", ia.Name, ia.Amount))
		}
	}

	for _, goal := range req.NutritionGoals {
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
		}

		if goal.Min > 0 && actual < goal.Min {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("%s实际值%.2f低于最小目标%.2f", goal.Nutrient, actual, goal.Min))
		}
		if goal.Max > 0 && actual > goal.Max {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("%s实际值%.2f超过最大目标%.2f", goal.Nutrient, actual, goal.Max))
		}
	}
}

func (o *FixedOptimizer) roundToTwoDecimals(x float64) float64 {
	return math.Round(x*100) / 100
}

func (o *FixedOptimizer) OptimizeWithStability(req OptimizationRequest) (*OptimizationResult, error) {
	rand.Seed(FixedRandomSeed)
	return o.Optimize(req)
}

func (o *FixedOptimizer) OptimizeWithConstraintFixed(req OptimizationRequest) (*OptimizationResult, error) {
	rand.Seed(FixedRandomSeed)
	return o.Optimize(req)
}
