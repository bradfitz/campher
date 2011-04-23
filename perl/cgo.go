package perl

/*
#cgo CFLAGS: -D_REENTRANT -D_GNU_SOURCE -DDEBIAN -fno-strict-aliasing -pipe -fstack-protector -I/usr/local/include -D_LARGEFILE_SOURCE -D_FILE_OFFSET_BITS=64  -I/usr/lib/perl/5.10/CORE
#cgo LDFLAGS: -Wl,-E  -fstack-protector -L/usr/local/lib  -L/usr/lib/perl/5.10/CORE -lperl -ldl -lm -lpthread -lc -lcrypt
#include <EXTERN.h>
#include <perl.h>

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

*/
import "C"

func init() {
	C.campher_init()
}

type Interpreter struct {
	perl *_Ctypedef_PerlInterpreter
}

func (in *Interpreter) be_context() {
	C.campher_set_context(in.perl)
}

func NewInterpreter() *Interpreter {
	int := new(Interpreter)
	int.perl = C.perl_alloc()
	// TODO: set finalizer and stuff

	int.be_context()
	

	return int
}