// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import "testing"

func TestE2E_TypeScript(t *testing.T) {
	fixtures := map[string]string{
		"src/models/user.ts": `export interface IUser {
  getName(): string;
  getEmail(): string;
}

export interface IAdmin extends IUser {
  getRole(): string;
}

export class User implements IUser {
  constructor(private name: string, private email: string) {}

  getName(): string {
    return this.name;
  }

  getEmail(): string {
    return this.email;
  }

  toJSON(): string {
    return JSON.stringify({ name: this.getName(), email: this.getEmail() });
  }
}

export class AdminUser extends User {
  constructor(name: string, email: string, private role: string) {
    super(name, email);
  }

  getRole(): string {
    return this.role;
  }
}

export class SuperAdmin extends AdminUser {
  override getRole(): string {
    return "super_" + this.role;
  }
}
`,
		"src/models/config.ts": `export type AppConfig = {
  host: string;
  port: number;
  debug: boolean;
};

export type DatabaseConfig = {
  url: string;
  pool: number;
};

export enum Environment {
  Development = "dev",
  Staging = "staging",
  Production = "prod",
}

export enum LogLevel {
  Debug,
  Info,
  Warn,
  Error,
}

export const DEFAULT_CONFIG: AppConfig = {
  host: "localhost",
  port: 3000,
  debug: true,
};

export const DB_CONFIG: DatabaseConfig = {
  url: "postgres://localhost:5432/app",
  pool: 10,
};

let currentEnv: Environment = Environment.Development;

export function setEnvironment(env: Environment): void {
  currentEnv = env;
}

export function getEnvironment(): Environment {
  return currentEnv;
}
`,
		"src/services/greeter.ts": `export function greet(name: string): string {
  return formatMessage(name);
}

function formatMessage(name: string): string {
  const prefix = getPrefix();
  return prefix + ", " + name;
}

function getPrefix(): string {
  return "Hello";
}

export const createGreeter = (prefix: string) => {
  return (name: string) => prefix + " " + name;
};

export async function fetchGreeting(userId: number): Promise<string> {
  const name = await fetchName(userId);
  return greet(name);
}

async function fetchName(userId: number): Promise<string> {
  return "User_" + userId;
}
`,
		"src/utils/helpers.ts": `let counter = 0;

export function increment(): number {
  counter++;
  return counter;
}

export function decrement(): number {
  counter--;
  return counter;
}

export function reset(): void {
  counter = 0;
}

export function identity<T>(value: T): T {
  return value;
}

export function map<T, U>(arr: T[], fn: (item: T) => U): U[] {
  return arr.map(fn);
}

export const clamp = (value: number, low: number, high: number): number => {
  if (value < low) return low;
  if (value > high) return high;
  return value;
};
`,
		// Test files that should be excluded by default.
		"src/services/greeter.test.ts": `import { greet } from './greeter';

describe('greet', () => {
  it('returns greeting', () => {
    expect(greet('World')).toBe('Hello, World');
  });
});
`,
		"src/models/user.spec.ts": `import { User } from './user';

describe('User', () => {
  it('returns name', () => {
    const u = new User('Alice', 'alice@example.com');
    expect(u.getName()).toBe('Alice');
  });
});
`,
		"src/__tests__/integration.ts": `import { createGreeter } from '../services/greeter';

test('createGreeter works', () => {
  const g = createGreeter('Hi');
  expect(g('World')).toBe('Hi World');
});
`,
	}

	inputDir := writeFixture(t, fixtures)
	_, outputDir := runcodeknit(t, inputDir)
	assertSnapshot(t, outputDir, inputDir)
}
