// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import "testing"

func TestE2E_Python(t *testing.T) {
	fixtures := map[string]string{
		"models/animal.py": `class Animal:
    def __init__(self, name, sound="..."):
        self.name = name
        self.sound = sound

    def speak(self):
        return self.name + " says " + self.sound

    def describe(self):
        return "Animal: " + self.speak()

class Dog(Animal):
    def __init__(self, name):
        super().__init__(name, "Woof")

    def speak(self):
        return self.name + " barks"

    def fetch(self, item):
        return self.name + " fetches " + item

class Cat(Animal):
    def speak(self):
        return self.name + " meows"
`,
		"models/user.py": `class User:
    def __init__(self, name, email):
        self.name = name
        self.email = email

    def display(self):
        return format_name(self.name) + " <" + self.email + ">"

    def validate(self):
        return validate_email(self.email)

def format_name(name):
    return name.strip().title()

def validate_email(email):
    return "@" in email and "." in email

MAX_USERS = 100
ADMIN_EMAIL = "admin@example.com"
`,
		"services/greeter.py": `def greet(name):
    msg = format_greeting(name)
    return msg

def format_greeting(name):
    prefix = get_prefix()
    return prefix + ", " + name

def get_prefix():
    return "Hello"

async def fetch_user(user_id):
    return {"id": user_id, "name": "user_" + str(user_id)}

async def fetch_users(ids):
    results = []
    for uid in ids:
        u = await fetch_user(uid)
        results.append(u)
    return results
`,
		"utils/helpers.py": `TIMEOUT = 30
MAX_RETRIES = 3

def validate(value):
    return value is not None and len(str(value)) > 0

def clamp(value, low, high):
    if value < low:
        return low
    if value > high:
        return high
    return value

class Validator:
    @staticmethod
    def check(value):
        return validate(value)

    @classmethod
    def from_config(cls, config):
        return cls()

    def run(self, data):
        results = []
        for item in data:
            results.append(self.check(item))
        return results
`,
		// Test files that should be excluded by default.
		"models/test_user.py": `from models.user import User

def test_user_display():
    u = User("Alice", "alice@example.com")
    assert "Alice" in u.display()
`,
		"models/user.test.py": `from models.user import User

def test_validate():
    u = User("Bob", "bob@example.com")
    assert u.validate()
`,
		"__tests__/test_greeter.py": `from services.greeter import greet

def test_greet():
    assert greet("World") == "Hello, World"
`,
	}

	inputDir := writeFixture(t, fixtures)
	_, outputDir := runcodeknit(t, inputDir)
	assertSnapshot(t, outputDir, inputDir)
}
