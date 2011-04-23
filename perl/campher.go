package perl

/*
#cgo CFLAGS: -D_REENTRANT -D_GNU_SOURCE -DDEBIAN -fno-strict-aliasing -pipe -fstack-protector -I/usr/local/include -D_LARGEFILE_SOURCE -D_FILE_OFFSET_BITS=64  -I/usr/lib/perl/5.10/CORE
#cgo LDFLAGS: -Wl,-E  -fstack-protector -L/usr/local/lib  -L/usr/lib/perl/5.10/CORE -lperl -ldl -lm -lpthread -lc -lcrypt
#include <EXTERN.h>
#include <perl.h>
#include "campher.c"
*/
import "C"

import (
	"unsafe"
)

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
	int.perl = C.campher_new_perl()
	// TODO: set finalizer and stuff
	return int
}

type SV *C.SV

func (ip *Interpreter) Eval(str string) SV {
	ip.be_context()
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))
	return SV(C.campher_eval_pv(ip.perl, cstr))
}

func (ip *Interpreter) EvalInt(str string) int {
	ip.be_context()
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))
	sv := C.campher_eval_pv(ip.perl, cstr)
	return int(C.campher_sv_int(ip.perl, sv))
}

func (ip *Interpreter) EvalString(str string) string {
	ip.be_context()
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))
	//sv := C.campher_eval_pv(ip.perl, cstr)
	return "foo" // int(C.campher_sv_int(ip.perl, sv))
}
