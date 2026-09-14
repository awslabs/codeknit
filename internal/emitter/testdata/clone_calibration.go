// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package calibration

func computeA(x, y int) int {
	result := x * y
	if result > 1000 {
		result /= 2
	}
	if result > 500 {
		result -= 100
	}
	if result > 200 {
		result -= 50
	}
	if result < 0 {
		result = 0
	}
	tax := computeTax(result)
	result += tax
	if result > 900 {
		result = applyCap(result)
	}
	return result + 1
}

// Same pipeline with one bounds-clamp branch removed: a near clone.
func computeB(x, y int) int {
	result := x * y
	if result > 1000 {
		result /= 2
	}
	if result > 500 {
		result -= 100
	}
	if result > 200 {
		result -= 50
	}
	tax := computeTax(result)
	result += tax
	if result > 900 {
		result = applyCap(result)
	}
	return result + 1
}

// A renamed exact clone of computeA.
func computeCopy(a, b int) int {
	total := a * b
	if total > 1000 {
		total /= 2
	}
	if total > 500 {
		total -= 100
	}
	if total > 200 {
		total -= 50
	}
	if total < 0 {
		total = 0
	}
	tax := computeTax(total)
	total += tax
	if total > 900 {
		total = applyCap(total)
	}
	return total + 1
}

// An unrelated lookup with a loop and different data flow.
func lookup(items []string, target string) int {
	for i, item := range items {
		if item == target {
			return i
		}
	}
	return -1
}
