// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package plugin provides Symbol/Edge types and language plugin interfaces for codeknit.
//
// Core types live in the types sub-package; this file re-exports them so that
// existing code using plugin.Symbol, plugin.Edge, etc. continues to compile.
package plugin

import "codeknit/internal/common/types"

// SymbolCategory is a language-agnostic classification for a Symbol.
type SymbolCategory = types.SymbolCategory

// Symbol represents a named element extracted from source code.
type Symbol = types.Symbol

// EdgeKind represents the type of relationship between two Symbols.
type EdgeKind = types.EdgeKind

// Edge represents a typed, directed relationship between two Symbols.
type Edge = types.Edge

// Props is a builder for symbol property maps.
type Props = types.Props

// FPToken represents a language-agnostic semantic operation for fingerprinting.
type FPToken = types.FPToken

// Re-export category constants.
const (
	CategoryCallable = types.CategoryCallable
	CategoryType     = types.CategoryType
	CategoryValue    = types.CategoryValue
	CategoryModule   = types.CategoryModule
	CategoryMeta     = types.CategoryMeta
)

// Re-export edge kind constants.
const (
	EdgeCalls      = types.EdgeCalls
	EdgeInherits   = types.EdgeInherits
	EdgeContains   = types.EdgeContains
	EdgeReferences = types.EdgeReferences
	EdgeImplements = types.EdgeImplements
	EdgeOverrides  = types.EdgeOverrides
	EdgeImports    = types.EdgeImports
	EdgeDecorates  = types.EdgeDecorates
	EdgeAliases    = types.EdgeAliases
	EdgeReturns    = types.EdgeReturns
)

// Re-export FPToken constants.
const (
	FPIf      = types.FPIf
	FPElse    = types.FPElse
	FPFor     = types.FPFor
	FPWhile   = types.FPWhile
	FPReturn  = types.FPReturn
	FPSwitch  = types.FPSwitch
	FPCase    = types.FPCase
	FPBreak   = types.FPBreak
	FPCont    = types.FPCont
	FPTry     = types.FPTry
	FPCatch   = types.FPCatch
	FPThrow   = types.FPThrow
	FPYield   = types.FPYield
	FPAwait   = types.FPAwait
	FPGo      = types.FPGo
	FPSelect  = types.FPSelect
	FPDefer   = types.FPDefer
	FPAssign  = types.FPAssign
	FPCall    = types.FPCall
	FPMember  = types.FPMember
	FPIndex   = types.FPIndex
	FPNew     = types.FPNew
	FPCast    = types.FPCast
	FPLambda  = types.FPLambda
	FPRange   = types.FPRange
	FPMatch   = types.FPMatch
	FPDelete  = types.FPDelete
	FPElseIf  = types.FPElseIf
	FPAdd     = types.FPAdd
	FPSub     = types.FPSub
	FPMul     = types.FPMul
	FPDiv     = types.FPDiv
	FPMod     = types.FPMod
	FPEq      = types.FPEq
	FPNeq     = types.FPNeq
	FPLt      = types.FPLt
	FPGt      = types.FPGt
	FPLte     = types.FPLte
	FPGte     = types.FPGte
	FPAnd     = types.FPAnd
	FPOr      = types.FPOr
	FPNot     = types.FPNot
	FPBitAnd  = types.FPBitAnd
	FPBitOr   = types.FPBitOr
	FPBitXor  = types.FPBitXor
	FPBitNot  = types.FPBitNot
	FPShl     = types.FPShl
	FPShr     = types.FPShr
	FPLitNum  = types.FPLitNum
	FPLitStr  = types.FPLitStr
	FPLitBool = types.FPLitBool
	FPLitNil  = types.FPLitNil
	FPDict    = types.FPDict
	FPArray   = types.FPArray
	FPVar     = types.FPVar
)

// Re-export functions from types.
var (
	NewProps       = types.NewProps
	MakeScopedName = types.MakeScopedName
)
