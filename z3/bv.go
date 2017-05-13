// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package z3

import (
	"math"
	"math/big"
)

/*
#cgo LDFLAGS: -lz3
#include <z3.h>
#include <stdlib.h>
*/
import "C"

// BV is an expression with bit-vector sort.
//
// Bit vectors correspond to machine words. They have finite domains
// of size 2^n and implement modular arithmetic in both unsigned and
// two's complement signed forms.
//
// BV implements Expr.
type BV expr

func init() {
	sortWrappers[SortBV] = func(x expr) Expr {
		return BV(x)
	}
}

// BVSort returns a bit-vector sort of the given width in bits.
func (ctx *Context) BVSort(bits int) *Sort {
	var csort C.Z3_sort
	ctx.do(func() {
		csort = C.Z3_mk_bv_sort(ctx.c, C.unsigned(bits))
	})
	return wrapSort(ctx, csort, SortBV)
}

// BVConst returns a bit-vector constant named "name" with the given
// width in bits.
func (ctx *Context) BVConst(name string, bits int) BV {
	return ctx.Const(name, ctx.BVSort(bits)).(BV)
}

// AsBigSigned returns the value of expr as a math/big.Int,
// interpreting expr as a signed two's complement number. If expr is
// not a literal, it returns nil, false.
func (expr BV) AsBigSigned() (val *big.Int, isLiteral bool) {
	v, isLiteral := expr.AsBigUnsigned()
	if v == nil {
		return v, isLiteral
	}
	size := expr.Sort().BVSize()
	if v.Bit(size-1) != 0 {
		shift := big.NewInt(1)
		shift.Lsh(shift, uint(size))
		v.Sub(v, shift)
	}
	return v, true
}

// AsBigUnsigned is like AsBigSigned, but interprets expr as unsigned.
func (expr BV) AsBigUnsigned() (val *big.Int, isLiteral bool) {
	return expr.asBigInt()
}

// AsInt64 returns the value of expr as an int64, interpreting expr as
// a two's complement signed number. If expr is not a literal, it
// returns 0, false, false. If expr is a literal, but its value cannot
// be represented as an int64, it returns 0, true, false.
func (expr BV) AsInt64() (val int64, isLiteral, ok bool) {
	// Z3_get_numeral_int64 (expr.asInt64) interprets the number
	// as unsigned because it's general-purpose. However, since
	// this method is specific to BV, we make this instead mirror
	// Z3_mk_int64. So, use Z3_get_numeral_uint64 and sign extend
	// it ourselves.
	uval, isLiteral, ok := expr.asUint64()
	if !isLiteral {
		return 0, isLiteral, ok
	}
	size := expr.Sort().BVSize()
	if ok && size < 64 {
		// Fits in an int64 regardless of sign. Sign-extend it.
		return int64(uval) << uint(64-size) >> uint(64-size), true, true
	}
	// size is >= 64, so we have to tread carefully.
	if ok && uval < 1<<63 {
		// Positive and fits in an int64.
		return int64(uval), true, true
	}
	// It may have overflowed uint64 just because of sign bits.
	// Take the slow path.
	bigVal, _ := expr.AsBigSigned()
	if bigVal.Cmp(big.NewInt(math.MaxInt64)) > 0 {
		return 0, true, false
	}
	if bigVal.Cmp(big.NewInt(math.MinInt64)) < 0 {
		return 0, true, false
	}
	return bigVal.Int64(), true, true
}

// AsUint64 is like AsInt64, but interprets expr as unsigned and fails
// if expr cannot be represented as a uint64.
func (expr BV) AsUint64() (val uint64, isLiteral, ok bool) {
	return expr.asUint64()
}

//go:generate go run genwrap.go -t BV $GOFILE

// Not returns the bit-wise negation of l.
//
//wrap:expr Not Z3_mk_bvnot l

// And returns the bit-wise and of l and r.
//
// l and r must have the same size.
//
//wrap:expr And Z3_mk_bvand l r

// Or returns the bit-wise or of l and r.
//
// l and r must have the same size.
//
//wrap:expr Or Z3_mk_bvor l r

// Xor returns the bit-wise xor of l and r.
//
// l and r must have the same size.
//
//wrap:expr Xor Z3_mk_bvxor l r

// Nand returns the bit-wise nand of l and r.
//
// l and r must have the same size.
//
//wrap:expr Nand Z3_mk_bvnand l r

// Nor returns the bit-wise nor of l and r.
//
// l and r must have the same size.
//
//wrap:expr Nor Z3_mk_bvnor l r

// Xnor returns the bit-wise xnor of l and r.
//
// l and r must have the same size.
//
//wrap:expr Xnor Z3_mk_bvxnor l r

// Neg returns the two's complement negation of l.
//
//wrap:expr Neg Z3_mk_bvneg l

// Add returns the two's complement sum of l and r.
//
// l and r must have the same size.
//
//wrap:expr Add Z3_mk_bvadd l r

// Sub returns the two's complement subtraction l minus r.
//
// l and r must have the same size.
//
//wrap:expr Sub Z3_mk_bvsub l r

// Mul returns the two's complement product of l and r.
//
// l and r must have the same size.
//
//wrap:expr Mul Z3_mk_bvmul l r

// UDiv returns the unsigned quotient of l divided by r.
//
// l and r must have the same size.
//
//wrap:expr UDiv Z3_mk_bvudiv l r

// SDiv returns the two's complement signed quotient of l divided by r.
//
// l and r must have the same size.
//
//wrap:expr SDiv Z3_mk_bvsdiv l r

// URem returns the unsigned remainder of l divided by r.
//
// l and r must have the same size.
//
//wrap:expr URem Z3_mk_bvurem l r

// SRem returns the two's complement signed remainder of l divided by r.
//
// The sign of the result follows the sign of l.
//
// l and r must have the same size.
//
//wrap:expr SRem Z3_mk_bvsrem l r

// SMod returns the two's complement signed modulus of l divided by r.
//
// The sign of the result follows the sign of r.
//
// l and r must have the same size.
//
//wrap:expr SMod Z3_mk_bvsmod l r

// ULT returns the l < r, where l and r are unsigned.
//
// l and r must have the same size.
//
//wrap:expr ULT:Bool Z3_mk_bvult l r

// SLT returns the l < r, where l and r are signed.
//
// l and r must have the same size.
//
//wrap:expr SLT:Bool Z3_mk_bvslt l r

// ULE returns the l <= r, where l and r are unsigned.
//
// l and r must have the same size.
//
//wrap:expr ULE:Bool Z3_mk_bvule l r

// SLE returns the l <= r, where l and r are signed.
//
// l and r must have the same size.
//
//wrap:expr SLE:Bool Z3_mk_bvsle l r

// UGE returns the l >= r, where l and r are unsigned.
//
// l and r must have the same size.
//
//wrap:expr UGE:Bool Z3_mk_bvuge l r

// SGE returns the l >= r, where l and r are signed.
//
// l and r must have the same size.
//
//wrap:expr SGE:Bool Z3_mk_bvsge l r

// UGT returns the l > r, where l and r are unsigned.
//
// l and r must have the same size.
//
//wrap:expr UGT:Bool Z3_mk_bvugt l r

// SGT returns the l > r, where l and r are signed.
//
// l and r must have the same size.
//
//wrap:expr SGT:Bool Z3_mk_bvsgt l r

// Concat returns concatenation of l and r.
//
// The result is a bit-vector whose length is the sum of the lengths
// of l and r.
//
//wrap:expr Concat Z3_mk_concat l r

// Extract returns bits [high, low] (inclusive) of l, where bit 0 is
// the least significant bit.
//
//wrap:expr Extract l high:int low:int : Z3_mk_extract high:unsigned low:unsigned l

// SignExtend returns l sign-extended to a bit-vector of length m+i,
// where m is the length of l.
//
//wrap:expr SignExtend l i:int : Z3_mk_sign_ext i:unsigned l

// ZeroExtend returns l zero-extended to a bit-vector of length m+i,
// where m is the length of l.
//
//wrap:expr ZeroExtend l i:int : Z3_mk_zero_ext i:unsigned l

// Repeat returns l repeated up to length i.
//
//wrap:expr Repeat l i:int : Z3_mk_repeat i:unsigned l

// Lsh returns l shifted left by i bits.
//
// This is equivalent to l * 2^i.
//
// l and i must have the same size. The result has the same sort.
//
//wrap:expr Lsh Z3_mk_bvshl l i

// URsh returns l logically shifted right by i bits.
//
// This is equivalent to l / 2^i, where l and i are unsigned.
//
// l and i must have the same size. The result has the same sort.
//
//wrap:expr URsh Z3_mk_bvlshr l i

// SRsh returns l arithmetically shifted right by i bits.
//
// This is like URsh, but the sign of the result is the sign of l.
//
// l and i must have the same size. The result has the same sort.
//
//wrap:expr SRsh Z3_mk_bvashr l i

// RotateLeft returns l rotated left by i bits.
//
// l and i must have the same size.
//
//wrap:expr RotateLeft Z3_mk_ext_rotate_left l i

// RotateRight returns l rotated right by i bits.
//
// l and i must have the same size.
//
//wrap:expr RotateRight Z3_mk_ext_rotate_right l i

// SToInt converts signed bit-vector l to an integer.
//
//wrap:expr SToInt:Int l : Z3_mk_bv2int l "C.Z3_TRUE"

// UToInt converts unsigned bit-vector l to an integer.
//
//wrap:expr UToInt:Int l : Z3_mk_bv2int l "C.Z3_FALSE"

// TODO: Z3_mk_bv*_no_{over,under}flow