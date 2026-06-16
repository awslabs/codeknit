// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import "testing"

func TestE2E_Java(t *testing.T) {
	fixtures := map[string]string{
		"src/models/Animal.java": `package com.example.models;

public abstract class Animal {
    private String name;
    protected String sound;

    public Animal(String name, String sound) {
        this.name = name;
        this.sound = sound;
    }

    public String getName() {
        return name;
    }

    public abstract String speak();

    public String describe() {
        return getName() + " says " + speak();
    }
}
`,
		"src/models/Dog.java": `package com.example.models;

public class Dog extends Animal implements Comparable<Dog> {
    private int age;

    public Dog(String name, int age) {
        super(name, "Woof");
        this.age = age;
    }

    @Override
    public String speak() {
        return getName() + " barks";
    }

    public String fetch(String item) {
        return getName() + " fetches " + item;
    }

    @Override
    public int compareTo(Dog other) {
        return Integer.compare(this.age, other.age);
    }
}
`,
		"src/services/AnimalService.java": `package com.example.services;

public interface AnimalService {
    void register(String name);
    String find(String name);
    void remove(String name);
}
`,
		"src/services/AnimalServiceImpl.java": `package com.example.services;

import java.util.ArrayList;
import java.util.List;

public class AnimalServiceImpl implements AnimalService {
    private List<String> registry = new ArrayList<>();

    public AnimalServiceImpl() {
    }

    @Override
    public void register(String name) {
        registry.add(name);
    }

    @Override
    public String find(String name) {
        for (String entry : registry) {
            if (entry.equals(name)) {
                return entry;
            }
        }
        return null;
    }

    @Override
    public void remove(String name) {
        registry.remove(name);
    }

    public int count() {
        return registry.size();
    }
}
`,
		"src/models/Status.java": `package com.example.models;

public enum Status {
    ACTIVE,
    INACTIVE,
    PENDING
}
`,
		// Test files that should be excluded by default.
		"src/models/Animal.test.java": `package com.example.models;

public class AnimalTest {
    public void testSpeak() {
        Dog d = new Dog("Rex", 3);
        assert d.speak().contains("barks");
    }
}
`,
		"src/services/AnimalService.spec.java": `package com.example.services;

public class AnimalServiceSpec {
    public void testRegister() {
        AnimalServiceImpl svc = new AnimalServiceImpl();
        svc.register("Rex");
    }
}
`,
		"src/__tests__/IntegrationTest.java": `package com.example;

public class IntegrationTest {
    public void testEndToEnd() {
        AnimalServiceImpl svc = new AnimalServiceImpl();
        svc.register("Rex");
        assert svc.count() == 1;
    }
}
`,
	}

	inputDir := writeFixture(t, fixtures)
	_, outputDir := runcodeknit(t, inputDir)
	assertSnapshot(t, outputDir, inputDir)
}
