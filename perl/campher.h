#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"

void campher_call_sv_scalar(PerlInterpreter* my_perl, SV* sv, SV** arg, SV** ret);
void campher_call_sv_void(PerlInterpreter* my_perl, SV* sv, SV** arg);
SV* campher_eval_pv(PerlInterpreter* my_perl, char* code);
int campher_get_sv_bool(PerlInterpreter* my_perl, SV* sv);
SV* campher_get_sv_cv(PerlInterpreter* my_perl, SV* sv);
NV campher_get_sv_float(PerlInterpreter* my_perl, SV* sv);
int campher_get_sv_int(PerlInterpreter* my_perl, SV* sv);
void campher_get_sv_string(PerlInterpreter* my_perl, SV* sv, char** out_char, int* out_len);
void campher_init();
SV* campher_mortal_sv_string(PerlInterpreter* my_perl, char* c, int len);
SV* campher_new_mortal_sv_int(PerlInterpreter* my_perl, int val);
PerlInterpreter* campher_new_perl();
SV* campher_new_sv_int(PerlInterpreter* my_perl, int val);
SV* campher_new_sv_string(PerlInterpreter* my_perl, char* c, int len);
void campher_set_context(PerlInterpreter* perl);
void campher_sv_decref(PerlInterpreter* my_perl, SV* sv);
SV* campher_undef_sv(PerlInterpreter* my_perl);