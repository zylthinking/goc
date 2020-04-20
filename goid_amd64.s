// +build amd64

TEXT ·Goid(SB), $0-8
    MOVQ -8(FS), AX
    MOVQ 0x98(AX), AX
    MOVQ AX, ret(FP)
    RET
