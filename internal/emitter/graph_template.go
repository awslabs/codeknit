// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package emitter

import _ "embed"

//go:embed graph_template.html
var graphTemplateHTML string

//go:embed d3.v7.min.js
var d3JS string
