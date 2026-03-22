package main

import (
	"fmt"
	"math"
	"sort"
)

const (
	MinIngredientAmount = 0.0
	MaxIngredientAmount = 500.0
	DefaultTolerance    = 0.05
	MaxIterations       = 100
	PopulationSize      = 50
)

type FixedOptimizer struct {
	warnings       []string
	populationSize int
	maxIterations  int
	tolerance      float64
	population     []*FixedIndividual
	idealPoint     []float64
	weightVectors  [][]float64
	neighborhood   [][]int
	neighborSize   int
	objectives     int
}

type FixedIndividual struct {
	Amounts    []float64
	Objectives []float64
	Fitness    float64
}

func NewFixedOptimizer() *FixedOptimizer {
	return &FixedOptimizer{
		warnings:       []string{},
		populationSize: PopulationSize,
		maxIterations:  MaxIterations,
		tolerance:      DefaultTolerance,
		neighborSize:   int(math.Ceil(float64(PopulationSize) * 0.2)),
		objectives:     3,
	}
}

func (o *FixedOptimizer) GetWarnings() []string {
	return o.warnings
}

func (o *FixedOptimizer) Optimize(req OptimizationRequest) (*OptimizationResult, error) {
	o.warnings = []string{}

	if len(req.Ingredients) == 0 {
		return nil, fmt.Errorf("没有可用的食材数据")
	}

	if req.MaxIterations > 0 {
		o.maxIterations = req.MaxIterations
	}
	if req.Tolerance > 0 {
		o.tolerance = req.Tolerance
	}

	o.initializeWeightVectors()
	o.calculateNeighborhood()
	o.initializeIdealPoint()

	o.population = make([]*FixedIndividual, o.populationSize)
	for i := range o.population {
		o.population[i] = o.createSmartIndividual(len(req.Ingredients), req)
		o.evaluateIndividual(o.population[i], req.Ingredients, req)
	}

	bestFitness := math.Inf(1)
	stagnationCount := 0
	convergenceThreshold := 5

	for gen := 0; gen < o.maxIterations; gen++ {
		improved := false

		for i := 0; i < o.populationSize; i++ {
			neighbors := o.neighborhood[i]
			p1 := neighbors[deterministicSelect(gen, i, len(neighbors))]
			p2 := neighbors[deterministicSelect(gen, i+1, len(neighbors))]

			child := o.crossover(o.population[p1], o.population[p2], gen)
			o.mutate(child, gen)
			o.repair(child, req)
			o.evaluateIndividual(child, req.Ingredients, req)

			for _, neighborIdx := range neighbors {
				f1 := o.tchebycheff(o.weightVectors[neighborIdx], o.population[neighborIdx].Objectives)
				f2 := o.tchebycheff(o.weightVectors[neighborIdx], child.Objectives)

				if f2 < f1 {
					o.population[neighborIdx] = o.cloneIndividual(child)
					improved = true
				}
			}
		}

		currentBest := o.getBestFitness()
		if currentBest < bestFitness-o.tolerance {
			bestFitness = currentBest
			stagnationCount = 0
		} else {
			stagnationCount++
		}

		if stagnationCount >= convergenceThreshold {
			o.warnings = append(o.warnings, fmt.Sprintf("在第%d代达到收敛，停止迭代", gen+1))
			break
		}

		if improved {
			stagnationCount = 0
		}
	}

	bestIdx := o.selectBestSolution(req.Weights)
	bestInd := o.population[bestIdx]

	result := o.generateResult(bestInd, req.Ingredients, req)
	result.Converged = true
	result.Iterations = o.maxIterations

	deviation := o.calculateConvergenceDeviation(bestInd, req)
	if deviation > 0.05 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("收敛偏差: %.2f%%", deviation*100))
	}

	return result, nil
}

func (o *FixedOptimizer) createSmartIndividual(numIngredients int, req OptimizationRequest) *FixedIndividual {
	amounts := make([]float64, numIngredients)

	targetWeight := 400.0
	for _, c := range req.Constraints {
		if c.Type == "total_weight" {
			targetWeight = c.Value
			break
		}
	}

	baseAmount := targetWeight / float64(numIngredients)

	for i := range amounts {
		amounts[i] = baseAmount
		amounts[i] = o.clampAmount(amounts[i])
	}

	return &FixedIndividual{Amounts: amounts}
}

func (o *FixedOptimizer) clampAmount(amount float64) float64 {
	return math.Max(MinIngredientAmount, math.Min(MaxIngredientAmount, amount))
}

func (o *FixedOptimizer) initializeWeightVectors() {
	o.weightVectors = make([][]float64, o.populationSize)

	step := 1.0 / float64(o.populationSize-1)
	for i := 0; i < o.populationSize; i++ {
		w1 := step * float64(i)
		w2 := (1.0 - w1) * 0.5
		w3 := 1.0 - w1 - w2
		o.weightVectors[i] = []float64{w1, w2, w3}
	}
}

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

func (o *FixedOptimizer) euclideanDistance(a, b []float64) float64 {
	sum := 0.0
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return math.Sqrt(sum)
}

func (o *FixedOptimizer) initializeIdealPoint() {
	o.idealPoint = []float64{0, 0, 0}
}

func (o *FixedOptimizer) updateIdealPoint(objectives []float64) {
	for i := range objectives {
		if objectives[i] < o.idealPoint[i] {
			o.idealPoint[i] = objectives[i]
		}
	}
}

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

func (o *FixedOptimizer) evaluateIndividual(ind *FixedIndividual, ingredients []Ingredient, req OptimizationRequest) {
	ind.Objectives = make([]float64, o.objectives)

	nutrition := o.calculateNutrition(ind.Amounts, ingredients)
	ind.Objectives[0] = o.calculateNutritionDeviation(nutrition, req.NutritionGoals)
	ind.Objectives[1] = o.calculateTotalCost(ind.Amounts, ingredients)
	ind.Objectives[2] = 1.0 - o.calculateDiversity(ind.Amounts)

	o.updateIdealPoint(ind.Objectives)
}

func (o *FixedOptimizer) calculateNutrition(amounts []float64, ingredients []Ingredient) NutritionSummary {
	var nutrition NutritionSummary

	for i, amount := range amounts {
		if i >= len(ingredients) {
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

func deterministicSelect(gen, idx, size int) int {
	return (gen + idx) % size
}

func (o *FixedOptimizer) crossover(parent1, parent2 *FixedIndividual, gen int) *FixedIndividual {
	numVars := len(parent1.Amounts)
	child := &FixedIndividual{Amounts: make([]float64, numVars)}

	eta := 1.0
	for i := 0; i < numVars; i++ {
		x1 := parent1.Amounts[i]
		x2 := parent2.Amounts[i]

		if math.Abs(x1-x2) > 1e-10 {
			u := deterministicUniform(gen, i)
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

		child.Amounts[i] = o.clampAmount(child.Amounts[i])
	}

	return child
}

func deterministicUniform(gen, idx int) float64 {
	seed := int64(gen*1000 + idx + 1)
	x := float64((seed*1103515245+12345)%2147483648) / 2147483648.0
	return x
}

func (o *FixedOptimizer) mutate(ind *FixedIndividual, gen int) {
	eta := 20.0
	prob := 1.0 / float64(len(ind.Amounts))

	for i := range ind.Amounts {
		mutationKey := gen*100 + i
		if deterministicUniform(gen, mutationKey) < prob {
			x := ind.Amounts[i]
			delta1 := x / MaxIngredientAmount
			delta2 := (MaxIngredientAmount - x) / MaxIngredientAmount

			u := deterministicUniform(gen, mutationKey+1)
			var deltaq float64

			if u <= 0.5 {
				val := 2*u + (1-2*u)*math.Pow(1-delta1, eta+1)
				deltaq = math.Pow(val, 1.0/(eta+1)) - 1
			} else {
				val := 2*(1-u) + 2*(u-0.5)*math.Pow(1-delta2, eta+1)
				deltaq = 1 - math.Pow(val, 1.0/(eta+1))
			}

			x += deltaq * MaxIngredientAmount
			ind.Amounts[i] = o.clampAmount(x)
		}
	}
}

func (o *FixedOptimizer) repair(ind *FixedIndividual, req OptimizationRequest) {
	for i := range ind.Amounts {
		ind.Amounts[i] = o.clampAmount(ind.Amounts[i])
	}

	for _, c := range req.Constraints {
		switch c.Type {
		case "ingredient_min":
			if c.IngredientID > 0 && c.IngredientID <= len(ind.Amounts) {
				idx := c.IngredientID - 1
				minVal := o.clampAmount(c.Value)
				if ind.Amounts[idx] < minVal {
					ind.Amounts[idx] = minVal
				}
			}
		case "ingredient_max":
			if c.IngredientID > 0 && c.IngredientID <= len(ind.Amounts) {
				idx := c.IngredientID - 1
				maxVal := o.clampAmount(c.Value)
				if ind.Amounts[idx] > maxVal {
					ind.Amounts[idx] = maxVal
				}
			}
		}
	}

	targetWeight := 400.0
	for _, c := range req.Constraints {
		if c.Type == "total_weight" {
			targetWeight = c.Value
			break
		}
	}

	currentTotal := 0.0
	for _, amount := range ind.Amounts {
		currentTotal += amount
	}

	if currentTotal > 0 {
		ratio := targetWeight / currentTotal
		for i := range ind.Amounts {
			ind.Amounts[i] *= ratio
			ind.Amounts[i] = o.clampAmount(ind.Amounts[i])
		}
	}

	for _, c := range req.Constraints {
		switch c.Type {
		case "ingredient_min":
			if c.IngredientID > 0 && c.IngredientID <= len(ind.Amounts) {
				idx := c.IngredientID - 1
				minVal := o.clampAmount(c.Value)
				if ind.Amounts[idx] < minVal {
					ind.Amounts[idx] = minVal
				}
			}
		case "ingredient_max":
			if c.IngredientID > 0 && c.IngredientID <= len(ind.Amounts) {
				idx := c.IngredientID - 1
				maxVal := o.clampAmount(c.Value)
				if ind.Amounts[idx] > maxVal {
					ind.Amounts[idx] = maxVal
				}
			}
		}
	}
}

func (o *FixedOptimizer) cloneIndividual(ind *FixedIndividual) *FixedIndividual {
	clone := &FixedIndividual{
		Amounts:    make([]float64, len(ind.Amounts)),
		Objectives: make([]float64, len(ind.Objectives)),
		Fitness:    ind.Fitness,
	}
	copy(clone.Amounts, ind.Amounts)
	copy(clone.Objectives, ind.Objectives)
	return clone
}

func (o *FixedOptimizer) getBestFitness() float64 {
	best := math.Inf(1)
	for _, ind := range o.population {
		fitness := o.tchebycheff([]float64{0.33, 0.33, 0.34}, ind.Objectives)
		if fitness < best {
			best = fitness
		}
	}
	return best
}

func (o *FixedOptimizer) selectBestSolution(weights []Weight) int {
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

	bestIdx := 0
	minScore := math.Inf(1)

	for i, ind := range o.population {
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

func (o *FixedOptimizer) calculateConvergenceDeviation(ind *FixedIndividual, req OptimizationRequest) float64 {
	if len(req.NutritionGoals) == 0 {
		return 0
	}

	nutrition := o.calculateNutrition(ind.Amounts, req.Ingredients)
	return o.calculateNutritionDeviation(nutrition, req.NutritionGoals)
}

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
