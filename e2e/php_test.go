// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import "testing"

func TestE2E_PHP(t *testing.T) {
	fixtures := map[string]string{
		"src/Models/Entity.php": `<?php
namespace App\Models;

abstract class Entity {
    protected int $id;
    public string $name;

    abstract public function validate(): bool;

    public function display(): string {
        return $this->name;
    }
}

interface Repository {
    public function find(int $id): mixed;
    public function save(mixed $entity): void;
    public function delete(int $id): void;
}

enum Status {
    case Active;
    case Inactive;
    case Pending;
}
`,
		"src/Models/User.php": `<?php
namespace App\Models;

class User extends Entity implements Repository {
    private string $email;
    public int $age;

    public function __construct(string $name, string $email) {
        $this->name = $name;
        $this->email = $email;
    }

    public function validate(): bool {
        return $this->checkEmail($this->email);
    }

    private function checkEmail(string $addr): bool {
        return str_contains($addr, "@");
    }

    public static function create(string $name, string $email): self {
        return new self($name, $email);
    }

    public function find(int $id): mixed {
        return null;
    }

    public function save(mixed $entity): void {
    }

    public function delete(int $id): void {
    }
}

const MAX_USERS = 100;
`,
		"src/Services/AuthService.php": `<?php
namespace App\Services;

trait Loggable {
    public function log(string $msg): void {
        echo $msg;
    }
}

class AuthService {
    use Loggable;

    private string $secret;

    public function __construct(string $secret) {
        $this->secret = $secret;
    }

    public function authenticate(string $token): bool {
        return $this->verifyToken($token);
    }

    private function verifyToken(string $token): bool {
        return strlen($token) > 0;
    }

    public static function defaultService(): self {
        return new self("default_secret");
    }
}

interface Notifier {
    public function send(string $message): void;
}

interface Logger extends Notifier {
    public function logMessage(string $level, string $message): void;
}
`,
		"src/Utils/helpers.php": `<?php
namespace App\Utils;

function validate(mixed $value): bool {
    return $value !== null;
}

function clamp(int $value, int $low, int $high): int {
    if ($value < $low) return $low;
    if ($value > $high) return $high;
    return $value;
}

class Validator {
    private int $maxRetries;

    public function __construct(int $maxRetries = 3) {
        $this->maxRetries = $maxRetries;
    }

    public function check(mixed $value): bool {
        return validate($value);
    }

    public function run(array $data): array {
        $results = [];
        foreach ($data as $item) {
            $results[] = $this->check($item);
        }
        return $results;
    }
}

const TIMEOUT = 30;
const MAX_RETRIES = 3;
`,
		// Test files that should be excluded by default.
		"src/Models/Entity.test.php": `<?php
namespace App\Models\Tests;

class EntityTest {
    public function testValidate(): void {
    }
}
`,
		"src/Services/AuthService.spec.php": `<?php
namespace App\Services\Tests;

class AuthServiceSpec {
    public function testAuthenticate(): void {
    }
}
`,
		"__tests__/IntegrationTest.php": `<?php
class IntegrationTest {
    public function testEndToEnd(): void {
    }
}
`,
	}

	inputDir := writeFixture(t, fixtures)
	_, outputDir := runcodeknit(t, inputDir)
	assertSnapshot(t, outputDir, inputDir)
}
