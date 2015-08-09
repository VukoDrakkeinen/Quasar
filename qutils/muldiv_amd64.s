// func MultiplyThenDivide(factor1, factor2, divisor int64) int64
TEXT ·MultiplyThenDivide(SB),$0
	MOVQ factor1+0(FP), AX
	IMULQ factor2+8(FP)
	IDIVQ divisor+16(FP)
	MOVQ AX, retval+24(FP)
	RET
	