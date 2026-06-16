// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import "testing"

func TestE2E_JavaScript(t *testing.T) {
	fixtures := map[string]string{
		"src/models/animal.js": `export class Animal {
  constructor(name) {
    this.name = name;
    this.sound = "...";
  }

  speak() {
    return this.name + " says " + this.sound;
  }

  describe() {
    return "Animal: " + this.speak();
  }
}

export class Dog extends Animal {
  constructor(name) {
    super(name);
    this.sound = "Woof";
  }

  speak() {
    return this.name + " barks";
  }

  fetch(item) {
    return this.name + " fetches " + item;
  }
}

export class Cat extends Animal {
  speak() {
    return this.name + " meows";
  }
}
`,
		"src/services/math.js": `export function add(a, b) {
  return a + b;
}

export function multiply(a, b) {
  return a * b;
}

export function subtract(a, b) {
  return add(a, multiply(-1, b));
}

export function compute(x, y) {
  const product = multiply(x, y);
  const sum = add(product, x);
  return subtract(sum, y);
}

let accumulator = 0;

export function accumulate(value) {
  accumulator = add(accumulator, value);
  return accumulator;
}

export function resetAccumulator() {
  accumulator = 0;
}
`,
		"src/utils/helpers.js": `export const format = (value) => {
  return String(value).trim();
};

export const compose = (f, g) => {
  return (x) => f(g(x));
};

let count = 0;

export function increment() {
  count++;
  return count;
}

export function decrement() {
  count--;
  return count;
}

export function getCount() {
  return count;
}
`,
		"src/config.jsx": `export const DEFAULT_OPTIONS = { timeout: 30, retries: 3 };

export const FEATURE_FLAGS = { darkMode: true, beta: false };

export function createApp(config) {
  const merged = Object.assign({}, DEFAULT_OPTIONS, config);
  return initializeApp(merged);
}

function initializeApp(config) {
  return { started: true, config };
}

let appInstance = null;

export function getApp() {
  return appInstance;
}
`,
		// Test files that should be excluded by default.
		"src/services/math.test.js": `import { add, multiply } from './math';

test('add returns sum', () => {
  expect(add(1, 2)).toBe(3);
});
`,
		"src/models/animal.spec.js": `import { Animal, Dog } from './animal';

describe('Dog', () => {
  it('speaks', () => {
    const d = new Dog('Rex');
    expect(d.speak()).toBe('Rex barks');
  });
});
`,
		"src/__tests__/integration.js": `import { createApp } from '../config';

test('createApp works', () => {
  const app = createApp({ timeout: 60 });
  expect(app.started).toBe(true);
});
`,
	}

	inputDir := writeFixture(t, fixtures)
	_, outputDir := runcodeknit(t, inputDir)
	assertSnapshot(t, outputDir, inputDir)
}
