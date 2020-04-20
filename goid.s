
TEXT ·Goid(SB), $0-8
    MOVD -8(FS), AX
    MOVD 0x98(AX), AX
    MOVQ AX, ret(FP)
    RET
