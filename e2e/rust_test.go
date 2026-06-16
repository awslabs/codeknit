// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import "testing"

func TestE2E_Rust(t *testing.T) {
	fixtures := map[string]string{
		"src/models/animal.rs": `pub trait Animal {
    fn speak(&self) -> String;
    fn name(&self) -> &str;
}

pub struct Dog {
    pub name: String,
    age: u32,
}

impl Dog {
    pub fn new(name: String, age: u32) -> Self {
        Dog { name, age }
    }

    fn validate(&self) -> bool {
        self.name.len() > 0
    }
}

impl Animal for Dog {
    fn speak(&self) -> String {
        format!("{} barks", self.name)
    }

    fn name(&self) -> &str {
        &self.name
    }
}

pub struct Cat {
    pub name: String,
}

impl Animal for Cat {
    fn speak(&self) -> String {
        format!("{} meows", self.name)
    }

    fn name(&self) -> &str {
        &self.name
    }
}

pub const MAX_ANIMALS: usize = 100;
`,
		"src/models/config.rs": `pub enum LogLevel {
    Debug,
    Info,
    Warn,
    Error,
}

pub struct Config {
    pub name: String,
    pub port: u16,
    pub log_level: LogLevel,
}

impl Config {
    pub fn default_config() -> Self {
        Config {
            name: String::from("app"),
            port: 8080,
            log_level: LogLevel::Info,
        }
    }

    pub fn with_port(mut self, port: u16) -> Self {
        self.port = port;
        self
    }
}

pub type PortNumber = u16;

pub const DEFAULT_PORT: u16 = 8080;
static mut REQUEST_COUNT: u64 = 0;
`,
		"src/services/auth.rs": `use crate::models::config::Config;

pub struct AuthService {
    secret: String,
    ttl: u64,
}

impl AuthService {
    pub fn new(secret: String, ttl: u64) -> Self {
        AuthService { secret, ttl }
    }

    pub fn authenticate(&self, token: &str) -> bool {
        verify_token(token, &self.secret)
    }

    pub fn refresh(&self, token: &str) -> Option<String> {
        if self.authenticate(token) {
            Some(generate_token(&self.secret, self.ttl))
        } else {
            None
        }
    }
}

fn verify_token(token: &str, secret: &str) -> bool {
    !token.is_empty() && !secret.is_empty()
}

fn generate_token(secret: &str, ttl: u64) -> String {
    format!("{}:{}", secret, ttl)
}

pub const MAX_TOKEN_LEN: usize = 4096;
`,
		"src/lib.rs": `pub mod models;
pub mod services;

pub trait Describable {
    fn describe(&self) -> String;
}

pub fn greet(name: &str) -> String {
    format!("Hello, {}", name)
}

macro_rules! log_msg {
    ($msg:expr) => {
        println!("[LOG] {}", $msg);
    };
}

pub const VERSION: &str = "1.0.0";
`,
		// Test files that should be excluded by default.
		"src/models/animal.test.rs": `#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_dog_speak() {
        let dog = Dog::new("Rex".to_string(), 3);
        assert!(dog.speak().contains("barks"));
    }
}
`,
		"src/services/auth.spec.rs": `#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_authenticate() {
        let svc = AuthService::new("secret".to_string(), 3600);
        assert!(svc.authenticate("valid_token"));
    }
}
`,
		"__tests__/integration.rs": `fn test_integration() {
    // integration test placeholder
}
`,
	}

	inputDir := writeFixture(t, fixtures)
	_, outputDir := runcodeknit(t, inputDir)
	assertSnapshot(t, outputDir, inputDir)
}
