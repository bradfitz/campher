static int dummy_argc = 3;
static char** dummy_argv;
static char** dummy_env;
 
static void campher_init() {
  dummy_argv = malloc(sizeof(char*) * 3);
  dummy_env = malloc(sizeof(char*) * 2);
  dummy_argv[0] = "campher";
  dummy_argv[1] = "-e";
  dummy_argv[2] = "0";
  dummy_env[0] = "FOO=bar";
  dummy_env[1] = NULL;
  PERL_SYS_INIT3(&dummy_argc,&dummy_argv,&dummy_env);
}

static void campher_set_context(PerlInterpreter* perl) {
  PERL_SET_CONTEXT(perl);
}

static char *campher_embedding[] = { "", "-e", "0" };

static PerlInterpreter* campher_new_perl() {
  PerlInterpreter* my_perl = perl_alloc();
  PERL_SET_CONTEXT(my_perl);
  perl_construct(my_perl);
  perl_parse(my_perl, NULL, 3, campher_embedding, NULL);
  PL_exit_flags |= PERL_EXIT_DESTRUCT_END;
  perl_run(my_perl);
  return my_perl;
}

static SV* campher_eval_pv(PerlInterpreter* my_perl, char* code) {
  SV* ret = eval_pv(code, TRUE);
  // TODO: this might already be done and thus wrong + leaky:
  SvREFCNT_inc(ret);
  return ret;
}

static SV* campher_new_mortal_sv_int(PerlInterpreter* my_perl, int val) {
  return sv_2mortal(newSViv(val));
}

static SV* campher_new_sv_int(PerlInterpreter* my_perl, int val) {
  return newSViv(val);
}

static void campher_sv_decref(PerlInterpreter* my_perl, SV* sv) {
  SvREFCNT_dec(sv);
}

static SV* campher_mortal_sv_string(PerlInterpreter* my_perl, char* c, int len) {
  return sv_2mortal(newSVpvn(c, len));
}

static int campher_get_sv_int(PerlInterpreter* my_perl, SV* sv) {
  return SvIVx(sv);
}

static void campher_get_sv_string(PerlInterpreter* my_perl, SV* sv, char** out_char, int* out_len) {
  STRLEN len;
  char* c = SvPVutf8x(sv, len);
  *out_char = c;
  *out_len = len;
}

static NV campher_get_sv_float(PerlInterpreter* my_perl, SV* sv) {
  return SvNVx(sv);
}

static svtype campher_get_sv_type(PerlInterpreter* my_perl, SV* sv) {
  return SvTYPE(sv);
}

// arg is NULL-terminated and caller must free.
static void campher_call_sv_void(PerlInterpreter* my_perl, SV* sv, SV** arg) {
  PERL_SET_CONTEXT(my_perl); // TODO: is this needed?

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
static void campher_call_sv_scalar(PerlInterpreter* my_perl, SV* sv, SV** arg, SV** ret) {
  PERL_SET_CONTEXT(my_perl); // TODO: is this needed?

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
  // TODO: deal with error flag. will just crash process for now.

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
