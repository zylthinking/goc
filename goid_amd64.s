// +build amd64
#include "textflag.h"

TEXT ·Goid(SB), NOSPLIT, $0-8
	MOVQ (TLS), AX
	MOVQ ·offset(SB), BX
	MOVQ (AX)(BX*1), AX
	MOVQ AX, ret(FP)
	RET
