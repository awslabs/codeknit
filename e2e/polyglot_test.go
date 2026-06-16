// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestE2E_Polyglot(t *testing.T) {
	fixtures := map[string]string{
		// Go
		"backend/server.go": `package backend

type Server struct {
	Port int
	Host string
}

func NewServer(host string, port int) *Server {
	return &Server{Host: host, Port: port}
}
`,
		// TypeScript
		"frontend/app.ts": `export interface AppProps {
  title: string;
  version: number;
}

export class App {
  constructor(private props: AppProps) {}

  getTitle(): string {
    return this.props.title;
  }
}
`,
		// JavaScript
		"frontend/utils.js": `function formatDate(date) {
  return date.toISOString();
}

class Logger {
  log(msg) {
    console.log(msg);
  }
}

const VERSION = "1.0.0";
`,
		// Python
		"scripts/deploy.py": `class Deployer:
    def __init__(self, target):
        self.target = target

    def run(self):
        return deploy_to(self.target)

def deploy_to(target):
    return "deployed to " + target

MAX_RETRIES = 3
`,
		// Java
		"src/Main.java": `package com.example;

public class Main {
    private String name;

    public Main(String name) {
        this.name = name;
    }

    public String getName() {
        return name;
    }

    public static void main(String[] args) {
        Main m = new Main("polyglot");
    }
}
`,
		// C
		"native/math.c": `#include <stdio.h>

int add(int a, int b) {
    return a + b;
}

int multiply(int a, int b) {
    return a * b;
}

static int square(int x) {
    return multiply(x, x);
}
`,
		// C++
		"native/vector.cpp": `#include <string>

namespace geometry {

class Vector {
public:
    double x, y;

    Vector(double x, double y) : x(x), y(y) {}

    double length() const {
        return x * x + y * y;
    }
};

} // namespace geometry
`,
		// C#
		"dotnet/Service.cs": `namespace App.Services
{
    public interface IService
    {
        string Execute();
    }

    public class Service : IService
    {
        private string name;

        public Service(string name)
        {
            this.name = name;
        }

        public string Execute()
        {
            return "running " + name;
        }
    }
}
`,
		// Ruby
		"scripts/task.rb": `class Task
  def initialize(name)
    @name = name
  end

  def run
    execute(@name)
  end

  private

  def execute(name)
    "executed " + name
  end
end

DEFAULT_TIMEOUT = 30
`,
		// Rust
		"native/config.rs": `pub struct Config {
    pub name: String,
    pub port: u16,
}

impl Config {
    pub fn new(name: String, port: u16) -> Self {
        Config { name, port }
    }
}

pub const DEFAULT_PORT: u16 = 8080;
`,
		// PHP
		"web/index.php": `<?php

class Router {
    private $routes = [];

    public function add($path, $handler) {
        $this->routes[$path] = $handler;
    }

    public function dispatch($path) {
        if (isset($this->routes[$path])) {
            return $this->routes[$path]();
        }
        return null;
    }
}

function create_router() {
    return new Router();
}
`,
		// Scala
		"analytics/Stats.scala": `package analytics

trait Aggregator {
  def aggregate(values: List[Double]): Double
}

class Stats extends Aggregator {
  def aggregate(values: List[Double]): Double = {
    if (values.isEmpty) 0.0 else values.sum / values.size
  }

  def median(values: List[Double]): Double = {
    val sorted = values.sorted
    sorted(sorted.size / 2)
  }
}

object Stats {
  def empty: Stats = new Stats()
}
`,
	}

	// Map of language name -> file extensions used in fixtures above.
	langExtensions := map[string][]string{
		"Go":         {".go"},
		"TypeScript": {".ts"},
		"JavaScript": {".js"},
		"Python":     {".py"},
		"Java":       {".java"},
		"C":          {".c"},
		"C++":        {".cpp"},
		"C#":         {".cs"},
		"Ruby":       {".rb"},
		"Rust":       {".rs"},
		"PHP":        {".php"},
		"Scala":      {".scala"},
	}

	inputDir := writeFixture(t, fixtures)
	sg, outputDir := runcodeknit(t, inputDir)

	// Snapshot test: verify full symbol graph against golden file.
	assertSnapshot(t, outputDir, inputDir)

	// Requirement 14.3: Verify every source file appears as a file header in the output.
	fileSet := make(map[string]bool)
	for _, sym := range sg.Symbols {
		norm := normalizeFilePath(sym.FilePath)
		fileSet[norm] = true
	}
	for relPath := range fixtures {
		found := false
		for normPath := range fileSet {
			if strings.HasSuffix(normPath, relPath) || normPath == relPath {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("source file %q not found in output", relPath)
		}
	}

	// Requirement 14.4: Verify total symbol count >= number of source files.
	numFiles := len(fixtures)
	if len(sg.Symbols) < numFiles {
		t.Errorf("expected at least %d symbols (one per source file), got %d", numFiles, len(sg.Symbols))
	}

	// Requirement 14.5: Verify symbols from each included language are present.
	for lang, exts := range langExtensions {
		found := false
		for _, sym := range sg.Symbols {
			ext := filepath.Ext(sym.FilePath)
			for _, langExt := range exts {
				if ext == langExt {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			t.Errorf("no symbols found for language %s (extensions %v)", lang, exts)
		}
	}
}
