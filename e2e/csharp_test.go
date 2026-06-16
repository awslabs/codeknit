// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import "testing"

func TestE2E_CSharp(t *testing.T) {
	fixtures := map[string]string{
		"Models/Entity.cs": `namespace MyApp.Models {
    public abstract class Entity {
        private int id;
        public string Name { get; set; }

        public abstract void Validate();

        public virtual string Display() {
            return Name;
        }
    }

    public interface IRepository {
        void Save(string name);
        void Delete(int id);
    }

    public enum Status { Active, Inactive, Pending }
}
`,
		"Models/User.cs": `namespace MyApp.Models {
    public class User : Entity {
        private string email;
        public int Age { get; set; }

        public override void Validate() {
            CheckEmail(email);
        }

        private void CheckEmail(string addr) {
        }

        public static void Register(string name, string email) {
            var u = new User();
            u.Validate();
        }
    }

    public delegate void UserChangedHandler(string name);
}
`,
		"Services/UserService.cs": `namespace MyApp.Services {
    public class UserService {
        private int maxUsers;

        public void Process(string name) {
            var user = CreateUser(name);
            NotifyRegistered(user);
        }

        private string CreateUser(string name) {
            return name;
        }

        private void NotifyRegistered(string user) {
        }

        public async void SyncAll() {
        }
    }

    public interface INotifier {
        void Send(string message);
    }

    public interface ILogger : INotifier {
        void Log(string level, string message);
    }
}
`,
		"Models/Point.cs": `namespace MyApp.Models {
    public struct Point : IEquatable {
        public int X;
        public int Y;

        public void Reset() {
            X = 0;
        }
    }
}
`,
		// Test files that should be excluded by default.
		"Models/Entity.test.cs": `namespace MyApp.Models.Tests {
    public class EntityTest {
        public void TestValidate() {
        }
    }
}
`,
		"Services/UserService.spec.cs": `namespace MyApp.Services.Tests {
    public class UserServiceSpec {
        public void TestProcess() {
        }
    }
}
`,
		"__tests__/IntegrationTests.cs": `namespace MyApp.Tests {
    public class IntegrationTests {
        public void TestEndToEnd() {
        }
    }
}
`,
	}

	inputDir := writeFixture(t, fixtures)
	_, outputDir := runcodeknit(t, inputDir)
	assertSnapshot(t, outputDir, inputDir)
}
