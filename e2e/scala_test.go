// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import "testing"

func TestE2E_Scala(t *testing.T) {
	fixtures := map[string]string{
		"src/models/Animal.scala": `package com.example.models

trait Animal {
  def speak(): String
  def name: String
}

class Dog(val name: String, val age: Int) extends Animal {
  override def speak(): String = name + " barks"

  def fetch(item: String): String = name + " fetches " + item
}

class Cat(val name: String) extends Animal {
  override def speak(): String = name + " meows"
}

case class Bird(name: String, canFly: Boolean) extends Animal {
  override def speak(): String = name + " chirps"
}

val MAX_ANIMALS: Int = 100
`,
		"src/models/User.scala": `package com.example.models

abstract class Entity {
  val id: Int
  def validate(): Boolean
}

class User(val id: Int, val name: String, val email: String) extends Entity with Serializable {
  override def validate(): Boolean = checkEmail(email)

  private def checkEmail(addr: String): Boolean = addr.contains("@")

  def display(): String = formatName(name) + " <" + email + ">"

  private def formatName(n: String): String = n.trim
}

object User {
  def create(name: String, email: String): User = new User(0, name, email)

  val DEFAULT_NAME: String = "guest"
}

var userCount: Int = 0
`,
		"src/services/AuthService.scala": `package com.example.services

trait Authenticator {
  def authenticate(token: String): Boolean
}

class AuthService(secret: String, ttl: Int) extends Authenticator {
  override def authenticate(token: String): Boolean = verifyToken(token)

  private def verifyToken(token: String): Boolean = token.nonEmpty

  def refresh(token: String): Option[String] = {
    if (authenticate(token)) Some(generateToken())
    else None
  }

  private def generateToken(): String = secret + ":" + ttl.toString
}

object AuthService {
  def default(): AuthService = new AuthService("default", 3600)

  val MAX_TOKEN_LEN: Int = 4096
}
`,
		"src/utils/Helpers.scala": `package com.example.utils

object Helpers {
  def validate(value: Any): Boolean = value != null

  def clamp(value: Int, low: Int, high: Int): Int = {
    if (value < low) low
    else if (value > high) high
    else value
  }

  val TIMEOUT: Int = 30
  var retryCount: Int = 0
}

type StringAlias = String

sealed trait Result
case class Success(value: String) extends Result
case class Failure(error: String) extends Result
`,
		// Test files that should be excluded by default.
		"src/models/Animal.test.scala": `package com.example.models

class AnimalTest {
  def testSpeak(): Unit = {
    val dog = new Dog("Rex", 3)
    assert(dog.speak().contains("barks"))
  }
}
`,
		"src/services/AuthService.spec.scala": `package com.example.services

class AuthServiceSpec {
  def testAuthenticate(): Unit = {
    val svc = AuthService.default()
    assert(svc.authenticate("token"))
  }
}
`,
		"__tests__/IntegrationTest.scala": `package com.example

class IntegrationTest {
  def testEndToEnd(): Unit = {
    assert(true)
  }
}
`,
	}

	inputDir := writeFixture(t, fixtures)
	_, outputDir := runcodeknit(t, inputDir)
	assertSnapshot(t, outputDir, inputDir)
}
