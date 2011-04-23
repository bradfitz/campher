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
  return eval_pv(code, TRUE);
}

static int campher_sv_int(PerlInterpreter* my_perl, SV* sv) {
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

static void campher_call_sv_void(PerlInterpreter* my_perl, SV* sv) {
  I32 flags = 0;
  I32 ret;

  PERL_SET_CONTEXT(my_perl); // TODO: is this needed?
  flags |= G_VOID | G_NOARGS;
  ret = call_sv(sv, flags);
}
