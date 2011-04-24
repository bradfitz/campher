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
	"fmt"
	"log"
	"runtime"
	"unsafe"
)

var _ = log.Printf

func init() {
	C.campher_init()
}

type Interpreter struct {
	perl *_Ctypedef_PerlInterpreter
}

type SV struct {
	ip *Interpreter
	sv *C.SV
}

// A code value CV that's callable.
type CV SV

func NewInterpreter() *Interpreter {
	ip := &Interpreter{
		perl: C.campher_new_perl(),
	}
	runtime.SetFinalizer(ip, func(ip *Interpreter) {
		C.perl_destruct(ip.perl)
		C.perl_free(ip.perl)
	})
	return ip
}

func (ip *Interpreter) be_context() {
	C.campher_set_context(ip.perl)
}

// newSvDecLater returns a new SV from a C.SV that has a reference
// count we need to decrement later.
func (ip *Interpreter) newSvDecLater(csv *C.SV) *SV {
	sv := &SV{ip, csv}
	sv.setFinalizer()
	return sv
}

func (ip *Interpreter) NewInt(val int) *SV {
	return ip.newSvDecLater(C.campher_new_sv_int(ip.perl, C.int(val)))
}

func (sv *SV) setFinalizer() {
	runtime.SetFinalizer(sv, func(sv *SV) {
		C.campher_sv_decref(sv.ip.perl, sv.sv)
	})
}

func (sv *SV) String() string {
	var cstr *C.char
	var length C.int
	C.campher_get_sv_string(sv.ip.perl, sv.sv, &cstr, &length)
	return C.GoStringN(cstr, length)
}

func (sv *SV) Int() int {
	return int(C.campher_get_sv_int(sv.ip.perl, sv.sv))
}

func (sv *SV) Bool() bool {
	return C.campher_get_sv_bool(sv.ip.perl, sv.sv) != 0
}

var dummySVPtr *C.SV
var svPtrSize = unsafe.Sizeof(dummySVPtr)

func (ip Interpreter) rawSvForFuncCall(arg interface{}) *C.SV {
	switch val := arg.(type) {
	case int:
		return C.campher_new_mortal_sv_int(ip.perl, C.int(val))
	case string:
		cstr := C.CString(val)
		defer C.free(unsafe.Pointer(cstr))
		return C.campher_mortal_sv_string(ip.perl, cstr, C.int(len(val)))
	case *SV:
		return val.sv
	}
	panic(fmt.Sprintf("TODO: can't use type %T in call", arg))
}

func (cv *CV) buildCallArgs(goargs ...interface{}) (**C.SV, bool) {
	if len(goargs) == 0 {
		return (**C.SV)(unsafe.Pointer(uintptr(0))), false
	}
	var args **C.SV
	var mallocSize int = svPtrSize * (len(goargs) + 1)
	var memory unsafe.Pointer = C.malloc(C.size_t(mallocSize))
	args = (**C.SV)(memory)
	for idx, goarg := range goargs {
		var thisArg **C.SV = (**C.SV)(unsafe.Pointer(uintptr(memory) + uintptr(idx*svPtrSize)))
		*thisArg = cv.ip.rawSvForFuncCall(goarg)
	}
	nullArg := (**C.SV)(unsafe.Pointer(uintptr(memory) + uintptr(len(goargs)*svPtrSize)))
	*nullArg = (*C.SV)(unsafe.Pointer(uintptr(0)))
	return args, true
}

// Call calls cv with any provided args in scalar context.
func (cv *CV) Call(args ...interface{}) *SV {
	perlargs, needFree := cv.buildCallArgs(args...)
	if needFree {
		defer C.free(unsafe.Pointer(perlargs))
	}
	var ret *C.SV
	C.campher_call_sv_scalar(cv.ip.perl, cv.sv, perlargs, &ret)
	return cv.ip.newSvDecLater(ret)
}

// Call calls cv  any provided args in void context.
func (cv *CV) CallVoid(args ...interface{}) {
	perlargs, needFree := cv.buildCallArgs(args...)
	if needFree {
		defer C.free(unsafe.Pointer(perlargs))
	}
	C.campher_call_sv_void(cv.ip.perl, cv.sv, perlargs)
}

// CV returns an SV's code value or nil if the SV is not of that type.
func (sv SV) CV() *CV {
	sv.ip.be_context()
	t := C.campher_get_sv_type(sv.ip.perl, sv.sv)
	if t&C.SVt_PVCV == 0 {
		log.Printf("t = %d; wanted = %d", t, C.SVt_PVCV)
		return nil
	}
	// inc the ref?
	cv := CV(sv)
	return &cv
}

func (ip *Interpreter) Eval(str string) *SV {
	ip.be_context()
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))
	return ip.newSvDecLater(C.campher_eval_pv(ip.perl, cstr))
}

func (ip *Interpreter) EvalInt(str string) int {
	return ip.Eval(str).Int()
}

func (ip *Interpreter) EvalString(str string) string {
	return ip.Eval(str).String()
}

func (ip *Interpreter) EvalFloat(str string) float64 {
	sv := ip.Eval(str)
	return float64(C.campher_get_sv_float(ip.perl, sv.sv))
}
