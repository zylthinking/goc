
#include "textflag.h"
#include "funcdata.h"

#ifdef GOARCH_amd64
TEXT ·goid(SB), NOSPLIT, $0-8
    MOVQ (TLS), AX
    MOVQ ·offset(SB), BX
    MOVQ (AX)(BX*1), AX
    MOVQ AX, ret(FP)
    RET
#endif

TEXT ·getg(SB), NOSPLIT, $0-8
    MOVQ (TLS), AX
    MOVQ AX, ret(FP)
    RET

TEXT ·getg_it(SB), NOSPLIT, $32-16
    NO_LOCAL_POINTERS
    MOVQ $0, ret_type+0(FP)
    MOVQ $0, ret_data+8(FP)
    GO_RESULTS_INITIALIZED

    CALL runtime·Goid(SB)

    MOVQ (TLS), AX
    MOVQ $type·runtime·g(SB), BX

    MOVQ AX, 8(SP)
    MOVQ BX, 0(SP)
    CALL runtime·convT2E(SB)

    MOVQ 16(SP), AX
    MOVQ 24(SP), BX

    MOVQ AX, ret+0(FP)
    MOVQ BX, ret+8(FP)
    RET
