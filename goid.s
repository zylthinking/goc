
#include "textflag.h"
#include "funcdata.h"

#ifdef GOARCH_amd64
TEXT ·goid2(SB), NOSPLIT, $0-8
    MOVQ (TLS), AX
    MOVQ ·offset(SB), BX
    MOVQ (AX)(BX*1), AX
    MOVQ AX, ret(FP)
    RET
#define ONE 8
#define TWO 16
#define THREE 24
#define FOUR 32

TEXT ·getg(SB), NOSPLIT, $0-ONE
    MOVQ (TLS), AX
    MOVQ AX, ret(FP)
    RET

TEXT ·getg_it(SB), NOSPLIT, $FOUR-TWO
    NO_LOCAL_POINTERS
    MOVQ $0, ret_type+0(FP)
    MOVQ $0, ret_data+ONE(FP)
    GO_RESULTS_INITIALIZED

    MOVQ (TLS), BX
    MOVQ $type·runtime·g(SB), AX

    MOVQ BX, ONE(SP)
    MOVQ AX, 0(SP)
    CALL runtime·convT2E(SB)

    CMPB ·newabi(SB), $0
    JNE LABEL

    MOVQ TWO(SP), AX
    MOVQ THREE(SP), BX

LABEL:
    MOVQ AX, ret+0(FP)
    MOVQ BX, ret+ONE(FP)
    RET
#endif

#ifdef GOARCH_386
#define MOVQ MOVL
#define ONE 4
#define TWO 8
#define THREE 12
#define FOUR 16

TEXT ·getg(SB), NOSPLIT, $0-ONE
    MOVQ (TLS), AX
    MOVQ AX, ret(FP)
    RET

TEXT ·getg_it(SB), NOSPLIT, $FOUR-TWO
    NO_LOCAL_POINTERS
    MOVQ $0, ret_type+0(FP)
    MOVQ $0, ret_data+ONE(FP)
    GO_RESULTS_INITIALIZED

    MOVQ (TLS), BX
    MOVQ $type·runtime·g(SB), AX

    MOVQ BX, ONE(SP)
    MOVQ AX, 0(SP)
    CALL runtime·convT2E(SB)

    CMPB ·newabi(SB), $0
    JNE LABEL

    MOVQ TWO(SP), AX
    MOVQ THREE(SP), BX

LABEL:
    MOVQ AX, ret+0(FP)
    MOVQ BX, ret+ONE(FP)
    RET
#endif
