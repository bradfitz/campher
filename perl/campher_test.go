package perl

import (
	"log"
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
