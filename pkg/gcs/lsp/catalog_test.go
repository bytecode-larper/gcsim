package lsp

import (
	"strings"
	"testing"

	"github.com/genshinsim/gcsim/pkg/gcs/ast"
	"github.com/genshinsim/gcsim/pkg/gcs/eval"
)

func TestCatalogUsesLanguageTables(t *testing.T) {
	cat := getCatalog()
	labels := map[string]bool{}
	for _, it := range cat.items {
		labels[it.Label] = true
	}

	for _, k := range ast.Keywords() {
		if !labels[k] {
			t.Errorf("catalog missing keyword %q from ast.Keywords", k)
		}
	}
	for k := range ast.ActionKeys {
		if !labels[k] {
			t.Errorf("catalog missing action %q from ast.ActionKeys", k)
		}
	}
	for _, k := range eval.BuiltinSysFuncs {
		if !labels[k] {
			t.Errorf("catalog missing sysfunc %q from eval.BuiltinSysFuncs", k)
		}
	}
	for k := range ast.StatKeys {
		if !labels[k] {
			t.Errorf("catalog missing stat %q", k)
		}
	}
	for k := range ast.EleKeys {
		if !labels[k] {
			t.Errorf("catalog missing element %q", k)
		}
	}
}

func TestDescribeKeywordWithoutDocStillWorks(t *testing.T) {
	for _, k := range ast.Keywords() {
		if md := describeCatalog(k, false); md == "" {
			t.Errorf("describeCatalog(%q) empty", k)
		}
	}
}

func TestDescribeSysFuncFromBuiltinList(t *testing.T) {
	if md := describeCatalog("f", false); md == "" || !strings.Contains(md, "system function") {
		t.Fatalf("expected system function hover for f, got %q", md)
	}
	if md := describeCatalog("f", true); md == "" || !strings.Contains(md, "param") {
		t.Fatalf("expected param hover for f= context, got %q", md)
	}
}
