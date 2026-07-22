// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package types

// FPToken represents a language-agnostic semantic operation extracted from
// a function body's AST.
type FPToken = byte

// Universal fingerprint token alphabet.
const (
	FPIf      FPToken = 0x01
	FPElse    FPToken = 0x02
	FPFor     FPToken = 0x03
	FPWhile   FPToken = 0x04
	FPReturn  FPToken = 0x05
	FPSwitch  FPToken = 0x06
	FPCase    FPToken = 0x07
	FPBreak   FPToken = 0x08
	FPCont    FPToken = 0x09
	FPTry     FPToken = 0x0A
	FPCatch   FPToken = 0x0B
	FPThrow   FPToken = 0x0C
	FPYield   FPToken = 0x0D
	FPAwait   FPToken = 0x0E
	FPGo      FPToken = 0x0F
	FPSelect  FPToken = 0x10
	FPDefer   FPToken = 0x11
	FPAssign  FPToken = 0x12
	FPCall    FPToken = 0x13
	FPMember  FPToken = 0x14
	FPIndex   FPToken = 0x15
	FPNew     FPToken = 0x16
	FPCast    FPToken = 0x17
	FPLambda  FPToken = 0x18
	FPRange   FPToken = 0x19
	FPMatch   FPToken = 0x1A
	FPDelete  FPToken = 0x1B
	FPElseIf  FPToken = 0x1C
	FPAdd     FPToken = 0x20
	FPSub     FPToken = 0x21
	FPMul     FPToken = 0x22
	FPDiv     FPToken = 0x23
	FPMod     FPToken = 0x24
	FPEq      FPToken = 0x28
	FPNeq     FPToken = 0x29
	FPLt      FPToken = 0x2A
	FPGt      FPToken = 0x2B
	FPLte     FPToken = 0x2C
	FPGte     FPToken = 0x2D
	FPAnd     FPToken = 0x30
	FPOr      FPToken = 0x31
	FPNot     FPToken = 0x32
	FPBitAnd  FPToken = 0x38
	FPBitOr   FPToken = 0x39
	FPBitXor  FPToken = 0x3A
	FPBitNot  FPToken = 0x3B
	FPShl     FPToken = 0x3C
	FPShr     FPToken = 0x3D
	FPLitNum  FPToken = 0x40
	FPLitStr  FPToken = 0x41
	FPLitBool FPToken = 0x42
	FPLitNil  FPToken = 0x43
	FPDict    FPToken = 0x44
	FPArray   FPToken = 0x45
	FPVar     FPToken = 0x46
)
