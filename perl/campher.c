#include <stdlib.h>
#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"

static int dummy_argc = 3;
static char** dummy_argv;
static char** dummy_env;
 
void campher_init() {
  dummy_argv = malloc(sizeof(char*) * 3);
  dummy_env = malloc(sizeof(char*) * 2);
  dummy_argv[0] = "campher";
  dummy_argv[1] = "-e";
  dummy_argv[2] = "0";
  dummy_env[0] = "FOO=bar";
  dummy_env[1] = NULL;
  PERL_SYS_INIT3(&dummy_argc,&dummy_argv,&dummy_env);
}

void campher_set_context(PerlInterpreter* perl) {
  PERL_SET_CONTEXT(perl);
}

static char *campher_embedding[] = { "", "-e", "0" };

static void xs_init (pTHX);

EXTERN_C void boot_DynaLoader (pTHX_ CV* cv);

// SV** args, SV** out_ret
extern void callCampherGoFunc(void* fnAddr, int narg, void* args, void* out_ret);

XS(XS_Campher_callback);
XS(XS_Campher_callback) {
  dXSARGS;
  if (items < 2) {
    croak("expected at least 2 arguments");
  }      
  int n_perl_args = items - 1;
  SV** perl_args = malloc(sizeof(SV*) * n_perl_args);
  int i;
  for (i = 0; i < n_perl_args; i++) {
    perl_args[i] = ST(i+1);
    SvREFCNT_inc(perl_args[i]);
  }
  SV* scalar_return = 0;
  callCampherGoFunc((void*)(SvIVx(ST(0))), n_perl_args, perl_args, &scalar_return);
  free(perl_args);
  if (!scalar_return) {
    ST(0) = &PL_sv_undef;
  } else {
    ST(0) = sv_2mortal(scalar_return);
  }
  XSRETURN(1);
}

EXTERN_C void
xs_init(pTHX)
{
  char *file = __FILE__;
  /* DynaLoader is a special case */
  newXS("DynaLoader::boot_DynaLoader", boot_DynaLoader, file);
  newXS("Campher::callback", XS_Campher_callback, file);
}

PerlInterpreter* campher_new_perl() {
  PerlInterpreter* my_perl = perl_alloc();
  PERL_SET_CONTEXT(my_perl);
  perl_construct(my_perl);
  perl_parse(my_perl, xs_init, 3, campher_embedding, NULL);
  PL_exit_flags |= PERL_EXIT_DESTRUCT_END;
  perl_run(my_perl);
  return my_perl;
}

SV* campher_eval_pv(PerlInterpreter* my_perl, char* code) {
  PERL_SET_CONTEXT(my_perl);
  SV* ret = eval_pv(code, TRUE);
  // TODO: this might already be done and thus wrong + leaky:
  SvREFCNT_inc(ret);
  return ret;
}

SV* campher_new_mortal_sv_int(PerlInterpreter* my_perl, int val) {
  PERL_SET_CONTEXT(my_perl);
  return sv_2mortal(newSViv(val));
}

SV* campher_new_sv_int(PerlInterpreter* my_perl, int val) {
  PERL_SET_CONTEXT(my_perl);
  return newSViv(val);
}

SV* campher_undef_sv(PerlInterpreter* my_perl) {
  PERL_SET_CONTEXT(my_perl);
  return &PL_sv_undef;
}

void campher_sv_decref(PerlInterpreter* my_perl, SV* sv) {
  PERL_SET_CONTEXT(my_perl);
  SvREFCNT_dec(sv);
}

SV* campher_new_sv_string(PerlInterpreter* my_perl, char* c, int len) {
  PERL_SET_CONTEXT(my_perl);
  return newSVpvn(c, len);
}

SV* campher_mortal_sv_string(PerlInterpreter* my_perl, char* c, int len) {
  PERL_SET_CONTEXT(my_perl);
  return sv_2mortal(newSVpvn(c, len));
}

int campher_get_sv_int(PerlInterpreter* my_perl, SV* sv) {
  PERL_SET_CONTEXT(my_perl);
  return SvIVx(sv);
}

int campher_get_sv_bool(PerlInterpreter* my_perl, SV* sv) {
  PERL_SET_CONTEXT(my_perl);
  return SvTRUE(sv);
}

void campher_get_sv_string(PerlInterpreter* my_perl, SV* sv, char** out_char, int* out_len) {
  PERL_SET_CONTEXT(my_perl);
  STRLEN len;
  char* c = SvPVutf8x(sv, len);
  *out_char = c;
  *out_len = len;
}

NV campher_get_sv_float(PerlInterpreter* my_perl, SV* sv) {
  PERL_SET_CONTEXT(my_perl);
  return SvNVx(sv);
}

svtype campher_get_sv_type(PerlInterpreter* my_perl, SV* sv) {
  PERL_SET_CONTEXT(my_perl);
  return SvTYPE(sv);
}

// arg is NULL-terminated and caller must free.
void campher_call_sv_void(PerlInterpreter* my_perl, SV* sv, SV** arg) {
  PERL_SET_CONTEXT(my_perl);

  dSP;

  ENTER;
  SAVETMPS;

  PUSHMARK(SP);
  if (arg != NULL) {
    while (*arg != NULL) {
      XPUSHs(*arg);
      arg++;
    }
  }
  PUTBACK;

  I32 ret = call_sv(sv, G_VOID);
  if (ret != 0) {
    assert(false);
  }

  FREETMPS;
  LEAVE;
}

// arg is NULL-terminated and caller must free.
void campher_call_sv_scalar(PerlInterpreter* my_perl, SV* sv, SV** arg, SV** ret) {
  PERL_SET_CONTEXT(my_perl);

  dSP;

  ENTER;
  SAVETMPS;

  PUSHMARK(SP);
  if (arg != NULL) {
    while (*arg != NULL) {
      XPUSHs(*arg);
      arg++;
    }
  }
  PUTBACK;

  I32 count = call_sv(sv, G_SCALAR);
  // TOD: deal with error flag. will just crash process for now.

  SPAGAIN;

  if (count != 1) {
    croak("expected 1 in campher_call_sv_scalar");
  }
  SV* result = POPs;
  SvREFCNT_inc(result);
  *ret = result;

  PUTBACK;
  FREETMPS;
  LEAVE;
}
