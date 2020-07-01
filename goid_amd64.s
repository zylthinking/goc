// +build amd64
#include "textflag.h"

TEXT ·Goid(SB), NOSPLIT, $0-8
    MOVQ -8(FS), AX
    MOVQ 0x98(AX), AX
    MOVQ AX, ret(FP)
    RET
