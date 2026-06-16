// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import "testing"

func TestE2E_Ruby(t *testing.T) {
	fixtures := map[string]string{
		"lib/models/animal.rb": `class Animal
  def initialize(name, sound)
    @name = name
    @sound = sound
  end

  def speak
    @name + " says " + @sound
  end

  def describe
    speak()
  end

  private

  def validate
    @name.length > 0
  end
end

class Dog < Animal
  def initialize(name)
    super(name, "Woof")
  end

  def speak
    @name + " barks"
  end

  def fetch(item)
    @name + " fetches " + item
  end
end

class Cat < Animal
  def speak
    @name + " meows"
  end
end
`,
		"lib/models/user.rb": `class User
  def initialize(name, email)
    @name = name
    @email = email
  end

  def display
    format_name(@name) + " <" + @email + ">"
  end

  def self.from_hash(data)
    new(data[:name], data[:email])
  end

  protected

  def format_name(name)
    name.strip
  end
end

MAX_USERS = 100
ADMIN_EMAIL = "admin@example.com"
`,
		"lib/services/greeter.rb": `module Greeter
  def self.greet(name)
    format_greeting(name)
  end

  def self.format_greeting(name)
    "Hello, " + name
  end
end

module Formatter
  def format(value)
    value.to_s.strip
  end
end

def standalone_helper(x)
  x.to_s
end

GREETING_PREFIX = "Hello"
`,
		"lib/services/auth_service.rb": `class AuthService
  include Formatter

  def initialize(secret)
    @secret = secret
  end

  def authenticate(token)
    verify_token(token)
  end

  def self.default_service
    new("default_secret")
  end

  private

  def verify_token(token)
    token.length > 0
  end
end

TOKEN_TTL = 3600
`,
		// Test files that should be excluded by default.
		"test/test_animal.rb": `require "minitest/autorun"
require_relative "../lib/models/animal"

class TestAnimal < Minitest::Test
  def test_speak
    dog = Dog.new("Rex")
    assert_includes dog.speak, "barks"
  end
end
`,
		"lib/models/user.test.rb": `require "minitest/autorun"
require_relative "user"

class TestUser < Minitest::Test
  def test_display
    u = User.new("Alice", "alice@example.com")
    assert_includes u.display, "Alice"
  end
end
`,
		"__tests__/integration_test.rb": `require "minitest/autorun"

class IntegrationTest < Minitest::Test
  def test_end_to_end
    assert true
  end
end
`,
	}

	inputDir := writeFixture(t, fixtures)
	_, outputDir := runcodeknit(t, inputDir)
	assertSnapshot(t, outputDir, inputDir)
}
