// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import "testing"

func TestE2E_C(t *testing.T) {
	fixtures := map[string]string{
		"include/config.h": `#define MAX_BUFFER_SIZE 1024
#define CLAMP(x, lo, hi) ((x) < (lo) ? (lo) : (x) > (hi) ? (hi) : (x))

typedef unsigned long size_type;

struct AppConfig {
    char *app_name;
    int port;
    int max_connections;
};

enum LogLevel {
    LOG_DEBUG,
    LOG_INFO,
    LOG_WARN,
    LOG_ERROR
};

void config_init(struct AppConfig *cfg);
int config_load(struct AppConfig *cfg, const char *path);

extern int verbose_flag;
`,
		"include/utils.h": `#include "config.h"

#define PI 3.14159
#define SQUARE(x) ((x) * (x))

typedef int (*compare_fn)(const void *, const void *);

union Value {
    int i;
    float f;
    char c;
};

int string_length(const char *s);
void string_copy(char *dst, const char *src);
`,
		"src/config.c": `#include "config.h"
#include <stdio.h>
#include <string.h>

static int default_port = 8080;

int verbose_flag = 0;

void config_init(struct AppConfig *cfg) {
    cfg->app_name = "codeknit";
    cfg->port = default_port;
    cfg->max_connections = 100;
}

int config_load(struct AppConfig *cfg, const char *path) {
    config_init(cfg);
    if (string_length(path) == 0) {
        return -1;
    }
    return 0;
}

static void config_reset(struct AppConfig *cfg) {
    config_init(cfg);
    cfg->port = default_port;
}
`,
		"src/utils.c": `#include "utils.h"
#include <stdlib.h>

#define INTERNAL_BUF 256

static char buffer[INTERNAL_BUF];

int string_length(const char *s) {
    int len = 0;
    while (s[len] != '\0') {
        len++;
    }
    return len;
}

void string_copy(char *dst, const char *src) {
    int len = string_length(src);
    for (int i = 0; i <= len; i++) {
        dst[i] = src[i];
    }
}

static int compare_int(const void *a, const void *b) {
    return (*(int *)a - *(int *)b);
}
`,
		"src/main.c": `#include "config.h"
#include "utils.h"
#include <stdio.h>

#define EXIT_OK 0

enum Status {
    STATUS_OK,
    STATUS_ERROR
};

struct Context {
    struct AppConfig config;
    enum Status status;
};

static void print_banner(void) {
    printf("codeknit v1.0\n");
}

int main(int argc, char *argv[]) {
    struct Context ctx;
    config_init(&ctx.config);
    print_banner();
    printf("port: %d\n", ctx.config.port);
    return EXIT_OK;
}
`,
		// Test files that should be excluded by default.
		"src/config.test.c": `#include "config.h"

void test_config_init(void) {
    struct AppConfig cfg;
    config_init(&cfg);
}
`,
		"src/utils.spec.c": `#include "utils.h"

void test_string_length(void) {
    int len = string_length("hello");
}
`,
		"__tests__/integration.c": `#include "config.h"

void test_integration(void) {
    struct AppConfig cfg;
    config_init(&cfg);
}
`,
	}

	inputDir := writeFixture(t, fixtures)
	_, outputDir := runcodeknit(t, inputDir)
	assertSnapshot(t, outputDir, inputDir)
}
