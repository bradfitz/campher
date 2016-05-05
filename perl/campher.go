/*
Copyright 2011 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

//
// Author: Brad Fitzpatrick <brad@danga.com>
//
// Disclaimer: this is more of a proof of concept than anything. I wouldn't
// use this in production yet.
//
// Patches welcome!  (or you can take over maintainership too... :))
//

// Package perl embeds a Perl5 interpreter in Go.
package perl

/*
#cgo CFLAGS: -D_REENTRANT -D_GNU_SOURCE -DDEBIAN -fno-strict-aliasing -pipe -fstack-protector -I/usr/local/include -D_LARGEFILE_SOURCE -D_FILE_OFFSET_BITS=64
#cgo LDFLAGS:  -fstack-protector -L/usr/local/lib -lperl -ldl -lm -lpthread -lc
#cgo !darwin CFLAGS: -I/usr/lib/perl/5.10/CORE
#cgo !darwin LDFLAGS: -Wl,-E -lcrypt -L/usr/lib/perl/5.10/CORE
#cgo darwin CFLAGS: -I/System/Library/Perl/5.18/darwin-thread-multi-2level/CORE
#cgo darwin LDFLAGS: -L/System/Library/Perl/5.18/darwin-thread-multi-2level/CORE
#include <EXTERN.h>
#include <perl.h>
#include "campher.h"
*/
import "C"

import (
	"fmt"
	"log"
	"reflect"
	"runtime"
	"sync"
	"unsafe"
)

var _ = log.Printf

func init() {
	C.campher_init()
}

type Interpreter struct {
	perl  *C.struct_interpreter
	undef *SV // lazily initialized
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

func (ip *Interpreter) NewString(val string) *SV {
	cstr := C.CString(val)
	defer C.free(unsafe.Pointer(cstr))
	return ip.newSvDecLater(C.campher_new_sv_string(ip.perl, cstr, C.int(len(val))))
}

func (ip *Interpreter) Undef() *SV {
	if ip.undef == nil {
		ip.undef = &SV{ip, C.campher_undef_sv(ip.perl)}
	}
	return ip.undef
}

type Callback func(args ...*SV) interface{}

type callbackState struct {
	cv       *CV
	callback Callback
}

var callbackLock sync.Mutex
var callbackMap = make(map[uintptr]*callbackState)

func (ip *Interpreter) NewCV(fn Callback) *CV {
	addr := uintptr(unsafe.Pointer(&fn))
	sv := ip.Eval(fmt.Sprintf("sub { Campher::callback(%d, @_); }", addr))
	callbackLock.Lock()
	defer callbackLock.Unlock()
	cv := (*CV)(sv)
	callbackMap[addr] = &callbackState{cv, fn}
	return cv
}

//export callCampherGoFunc
func callCampherGoFunc(fnAddr unsafe.Pointer, narg C.int, svArgsPtr unsafe.Pointer, svOutResult unsafe.Pointer) {
	// svArgsPtr is **C.SV (input array)
	// svOutResult is **C.SV (optional output for scalar result. value of 0 gets mapped to undef)
	callbackLock.Lock()
	cb := callbackMap[uintptr(fnAddr)]
	callbackLock.Unlock()

	if cb == nil {
		log.Printf("callback but cv not in map")
		return
	}

	ip := cb.cv.ip

	cbargs := make([]*SV, narg)
	for i := 0; i < int(narg); i++ {
		csv := *((**C.SV)(unsafe.Pointer(uintptr(svArgsPtr) + uintptr(i*svPtrSize))))
		cbargs[i] = ip.newSvDecLater(csv)
	}
	ei := cb.callback(cbargs...)
	var svOut **C.SV = (**C.SV)(svOutResult)
	switch val := ei.(type) {
	case int:
		*svOut = ip.NewInt(val).sv
	case string:
		*svOut = ip.NewString(val).sv
	case bool:
		if val {
			*svOut = ip.NewInt(1).sv
		} else {
			*svOut = ip.NewInt(0).sv
		}
	case *SV:
		*svOut = val.sv
	default:
		panic(fmt.Sprintf("can't yet deal with func return values of type %T: %#v", val, val))
	}
}

func (sv *SV) setFinalizer() {
	runtime.SetFinalizer(sv, func(sv *SV) {
		// TODO: does this get lost when things are converted to CV*?
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
var svPtrSize = int(unsafe.Sizeof(dummySVPtr))

func (ip *Interpreter) rawSvForFuncCall(arg interface{}) *C.SV {
	switch val := arg.(type) {
	case int:
		return C.campher_new_mortal_sv_int(ip.perl, C.int(val))
	case string:
		cstr := C.CString(val)
		defer C.free(unsafe.Pointer(cstr))
		return C.campher_mortal_sv_string(ip.perl, cstr, C.int(len(val)))
	case *SV:
		return val.sv
	case *CV:
		return val.sv
	}
	ftype := reflect.TypeOf(arg)
	if ftype.Kind() == reflect.Func {
		cv := ip.NewCV(func(args ...*SV) interface{} {
			callArg := make([]reflect.Value, ftype.NumIn())
			// extend incoming args, if necessary, to be long enough for the ftype
			// number of arguments
			for len(args) < ftype.NumIn() {
				args = append(args, ip.Undef())
			}
			for i := 0; i < ftype.NumIn(); i++ {
				kind := ftype.In(i).Kind()
				switch kind {
				case reflect.Bool:
					callArg[i] = reflect.ValueOf(args[i].Bool())
				case reflect.Int:
					callArg[i] = reflect.ValueOf(args[i].Int())
				case reflect.String:
					callArg[i] = reflect.ValueOf(args[i].String())
				default:
					panic(fmt.Sprintf("unsupported func callback arg type of kind: %d", kind))
				}
			}
			if ftype.NumOut() != 1 {
				panic(fmt.Sprintf("unsupported func callback returning %d arguments (only 1 supported now)", ftype.NumOut()))
			}
			fval := reflect.ValueOf(arg)
			results := fval.Call(callArg)
			switch ftype.Out(0).Kind() {
			case reflect.Bool:
				return results[0].Bool()
			case reflect.Int:
				return int(results[0].Int())
			case reflect.String:
				return results[0].String()
			}
			panic(fmt.Sprintf("unsupported func callback result kind of %d", ftype.Out(0).Kind()))
		})
		return cv.sv
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
func (sv *SV) CV() *CV {
	cv := C.campher_get_sv_cv(sv.ip.perl, sv.sv)
	if cv == nil {
		log.Printf("SV is not CV")
		return nil
	}
	// inc the ref?
	return &CV{sv.ip, cv}
}

func (ip *Interpreter) Eval(str string) *SV {
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
