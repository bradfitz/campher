package perl

import (
	"log"
	"runtime"
	"testing"
)

var _ = log.Printf

func TestPerl(t *testing.T) {
	perl := NewInterpreter()

	if e, g := 13, perl.EvalInt("5 + 8"); e != g {
		t.Errorf("5 + 8 expected %d; got %d", e, g)
	}

	perl.Eval("$foo = 123;")
	if e, g := 123, perl.EvalInt("$foo"); e != g {
		t.Errorf("Int(123) expected %d; got %d", e, g)
	}
	if e, g := "123", perl.EvalString("$foo"); e != g {
		t.Errorf("String(123) expected %q; got %q", e, g)
	}

	perl.Eval("$foo = 0.5;")
	if e, g := 0, perl.EvalInt("$foo"); e != g {
		t.Errorf("Int(0.5) expected %d; got %d", e, g)
	}
	if e, g := 0.5, perl.EvalFloat("$foo"); e != g {
		t.Errorf("Int(0.5) expected %f; got %f", e, g)
	}
}

func TestFinalizer(t *testing.T) {
	for i := 0; i < 10; i++ {
		perl := NewInterpreter()
		for n := 0; n < 500; n++ {
			perl.NewInt(n)
		}
	}
	runtime.GC()
}

func TestVoidNiladicCall(t *testing.T) {
	perl := NewInterpreter()
	perl.Eval("$val = 1;")
	perl.Eval("$code = sub { $val = 42; };")
	sv := perl.Eval("$code")
	cv := sv.CV()
	if cv == nil {
		t.Fatalf("cv is nil")
	}
	cv.CallVoid()
	if e, g := 42, perl.EvalInt("$val"); e != g {
		t.Errorf("Int($val) expected %d; got %d", e, g)
	}
}

func TestVoidCall(t *testing.T) {
	perl := NewInterpreter()
	perl.Eval("$foo = 1;")
	perl.Eval("$bar = 2;")
	perl.Eval("$baz = 3;")
	sv := perl.Eval(`sub { $nargs = @_; ($foo, $bar, $baz) = @_; }`)
	cv := sv.CV()
	if cv == nil {
		t.Fatalf("cv is nil")
	}
	cv.CallVoid(4, "five", perl.NewInt(6))
	if e, g := 3, perl.EvalInt("$nargs"); e != g {
		t.Errorf("Int($nargs) expected %d; got %d", e, g)
	}
	if e, g := 4, perl.EvalInt("$foo"); e != g {
		t.Errorf("Int($foo) expected %d; got %d", e, g)
	}
	if e, g := "five", perl.EvalString("$bar"); e != g {
		t.Errorf("String($bar) expected %q; got %q", e, g)
	}
	if e, g := 6, perl.EvalInt("$baz"); e != g {
		t.Errorf("Int($baz) expected %d; got %d", e, g)
	}
}

func TestScalarCall(t *testing.T) {
	perl := NewInterpreter()
	sv := perl.Eval(`sub { return 42 }`)
	cv := sv.CV()
	if cv == nil {
		t.Fatalf("cv is nil")
	}
	retsv := cv.Call()
	if e, g := 42, retsv.Int(); e != g {
		t.Errorf("Int(retsv) got %d, expected %d", g, e)
	}
	if e, g := "42", retsv.String(); e != g {
		t.Errorf("String(retsv) got %q, expected %q", g, e)
	}
}

