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

func (ip *Interpreter) NewInt(val int) *SV {
	sv := &SV{ip, C.campher_new_sv_int(ip.perl, C.int(val))}
	sv.setFinalizer()
	return sv
}

type SV struct {
	ip *Interpreter
	sv *C.SV
}

func (sv *SV) setFinalizer() {
	runtime.SetFinalizer(sv, func(sv *SV) {
		C.campher_sv_decref(sv.ip.perl, sv.sv)
	})
}

type CV SV

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

func (cv CV) CallVoid(goargs ...interface{}) {
	var args **C.SV
	if len(goargs) > 0 {
		var mallocSize int = svPtrSize * (len(goargs) + 1)
		var memory unsafe.Pointer = C.malloc(C.size_t(mallocSize))
		defer C.free(memory)
		args = (**C.SV)(memory)
		for idx, goarg := range goargs {
			var thisArg **C.SV = (**C.SV)(unsafe.Pointer(uintptr(memory) + uintptr(idx*svPtrSize)))
			*thisArg = cv.ip.rawSvForFuncCall(goarg)
		}
		nullArg := (**C.SV)(unsafe.Pointer(uintptr(memory) + uintptr(len(goargs)*svPtrSize)))
		*nullArg = (*C.SV)(unsafe.Pointer(uintptr(0)))
	}
	C.campher_call_sv_void(cv.ip.perl, cv.sv, args)
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

func (ip *Interpreter) Eval(str string) SV {
	ip.be_context()
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))
	// TODO: figure out refcounting here
	return SV{ip, C.campher_eval_pv(ip.perl, cstr)}
}

func (ip *Interpreter) EvalInt(str string) int {
	sv := ip.Eval(str)
	return int(C.campher_get_sv_int(ip.perl, sv.sv))
}

func (ip *Interpreter) EvalString(str string) string {
	sv := ip.Eval(str)
	var cstr *C.char
	var length C.int
	C.campher_get_sv_string(ip.perl, sv.sv, &cstr, &length)
	return C.GoStringN(cstr, length)
}

func (ip *Interpreter) EvalFloat(str string) float64 {
	sv := ip.Eval(str)
	return float64(C.campher_get_sv_float(ip.perl, sv.sv))
}
