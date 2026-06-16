// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import "testing"

func TestE2E_Cpp(t *testing.T) {
	fixtures := map[string]string{
		"include/shapes.hpp": `#pragma once

namespace geometry {

class Shape {
public:
    virtual double area() const = 0;
    virtual void draw() const = 0;
    virtual ~Shape() {}
};

struct Point {
    double x;
    double y;
};

enum Color { RED, GREEN, BLUE };

template<typename T>
class BoundingBox {
public:
    T minVal;
    T maxVal;
    bool contains(T value) const {
        return value >= minVal && value <= maxVal;
    }
};

} // namespace geometry
`,
		"include/utils.hpp": `#pragma once
#include "shapes.hpp"

namespace utils {

template<typename T>
T clamp(T value, T lo, T hi) {
    if (value < lo) return lo;
    if (value > hi) return hi;
    return value;
}

double distance(geometry::Point a, geometry::Point b);

} // namespace utils

extern int debugLevel;
`,
		"src/circle.cpp": `#include "shapes.hpp"
#include <cmath>

namespace geometry {

class Circle : public Shape {
public:
    Circle(double r) : radius(r) {}

    double area() const override {
        return 3.14159 * radius * radius;
    }

    void draw() const override {}

    double getRadius() const { return radius; }

private:
    double radius;
};

Circle createCircle(double r) {
    return Circle(r);
}

} // namespace geometry
`,
		"src/rectangle.cpp": `#include "shapes.hpp"

namespace geometry {

class Rectangle : public Shape {
public:
    Rectangle(double w, double h) : width(w), height(h) {}

    double area() const override {
        return width * height;
    }

    void draw() const override {}

protected:
    double width;
    double height;
};

class Square : public Rectangle {
public:
    Square(double side) : Rectangle(side, side) {}
};

} // namespace geometry
`,
		"src/main.cpp": `#include "shapes.hpp"
#include "utils.hpp"
#include <iostream>

int debugLevel = 0;

enum AppMode {
    MODE_NORMAL,
    MODE_DEBUG,
    MODE_VERBOSE
};

struct AppConfig {
    int maxShapes;
    AppMode mode;
};

void printArea(geometry::Shape* shape) {
    std::cout << shape->area() << std::endl;
}

int main() {
    AppConfig config;
    config.maxShapes = 100;
    config.mode = MODE_NORMAL;
    printArea(nullptr);
    return 0;
}
`,
		// Test files that should be excluded by default.
		"src/circle.test.cpp": `#include "shapes.hpp"

void test_circle_area() {
    // test circle area calculation
}
`,
		"src/rectangle.spec.cpp": `#include "shapes.hpp"

void test_rectangle_area() {
    // test rectangle area calculation
}
`,
		"__tests__/integration.cpp": `#include "shapes.hpp"

void test_integration() {
    // integration test
}
`,
	}

	inputDir := writeFixture(t, fixtures)
	_, outputDir := runcodeknit(t, inputDir)
	assertSnapshot(t, outputDir, inputDir)
}
