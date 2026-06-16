// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package fingerprint provides fuzzy hashing of normalized code structure
// for duplicate and near-duplicate detection across languages.
package fingerprint

import "codeknit/internal/plugin"

// Token constants are defined in plugin.FP* and re-exported here for
// convenience. The canonical definitions live in plugin/fptokens.go to
// avoid circular imports (language plugins need the constants but can't
// import the fingerprint package).

// Token is an alias for the universal fingerprint token type.
type Token = plugin.FPToken
