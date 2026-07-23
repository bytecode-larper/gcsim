package parser

import (
	"errors"
	"strings"
	"testing"

	"github.com/genshinsim/gcsim/pkg/gcs/ast"
)

func parseErr(t *testing.T, input string) error {
	t.Helper()
	file := ast.NewFile()
	p := New(file, input)
	_, _, err := p.Parse()
	return err
}

func requireAstError(t *testing.T, err error, msgSubstr string) ast.Error {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var ae ast.Error
	if !errors.As(err, &ae) {
		t.Fatalf("expected ast.Error, got %T: %v", err, err)
	}
	if msgSubstr != "" && !strings.Contains(ae.Msg, msgSubstr) && !strings.Contains(ae.Error(), msgSubstr) {
		t.Fatalf("error %q does not contain %q", ae.Error(), msgSubstr)
	}
	if !ae.Pos.IsValid() {
		t.Fatalf("expected valid position on error, got %v (%q)", ae.Pos, ae.Error())
	}
	return ae
}

func TestErrorInvalidWeaponName(t *testing.T) {
	err := parseErr(t, `xingqiu add weapon="notarealweapon" refine=1 lvl=1/1;`)
	requireAstError(t, err, "invalid weapon")
}

func TestErrorDuplicateFnParam(t *testing.T) {
	err := parseErr(t, `fn f(a, a) { }`)
	requireAstError(t, err, "duplicated param")
}

func TestErrorNonNumberMapParam(t *testing.T) {
	err := parseErr(t, `xingqiu add weapon="harbingerofdawn" refine=1 lvl=1/1 +params=[x=y];`)
	requireAstError(t, err, "expected number")
}

func TestErrorSyntax(t *testing.T) {
	err := parseErr(t, `let x = ;`)
	if err == nil {
		t.Fatal("expected syntax error")
	}
	// Syntax failures may surface as pigeon expected-set text rather than ast.Error.
	t.Logf("got expected error: %v", err)
}

func TestErrorActionStartLine(t *testing.T) {
	err := parseErr(t, `xingqiu attack; skill`)
	if err == nil {
		t.Fatal("expected error: line starting with action key")
	}
	t.Logf("got expected error: %v", err)
}

func TestErrorTargetNonHPStat(t *testing.T) {
	// Old parser rejected non-HP stats on target rows; atk= must not silently set HP.
	err := parseErr(t, `target atk=1000;`)
	if err == nil {
		t.Fatal("expected error for non-HP target stat")
	}
	t.Logf("got expected error: %v", err)
}

func TestErrorOptionWrongType(t *testing.T) {
	err := parseErr(t, `options iteration=true;`)
	requireAstError(t, err, "expects a number")

	err = parseErr(t, `options defhalt=1;`)
	requireAstError(t, err, "expects a boolean")
}

func TestErrorInvalidSetName(t *testing.T) {
	err := parseErr(t, `xingqiu add set="notarealset" count=4;`)
	requireAstError(t, err, "invalid set")
}
