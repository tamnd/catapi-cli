package catapi

import (
	"testing"
)

// These tests are offline: they exercise the URI driver's pure string functions,
// which need no network. The client's HTTP behaviour is covered in catapi_test.go.

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "catapi" {
		t.Errorf("Scheme = %q, want catapi", info.Scheme)
	}
	if len(info.Hosts) == 0 || info.Hosts[0] != Host {
		t.Errorf("Hosts = %v, want [%s]", info.Hosts, Host)
	}
	if info.Identity.Binary != "catapi" {
		t.Errorf("Identity.Binary = %q, want catapi", info.Identity.Binary)
	}
}

func TestClassify(t *testing.T) {
	typ, id, err := Domain{}.Classify("abys")
	if err != nil {
		t.Fatalf("Classify error: %v", err)
	}
	if typ != "breed" {
		t.Errorf("Classify type = %q, want breed", typ)
	}
	if id != "abys" {
		t.Errorf("Classify id = %q, want abys", id)
	}
}

func TestLocate(t *testing.T) {
	got, err := Domain{}.Locate("breed", "abys")
	want := "https://api.thecatapi.com/v1/breeds/abys"
	if err != nil || got != want {
		t.Errorf("Locate = (%q, %v), want (%q, nil)", got, err, want)
	}
}

func TestClassifyEmpty(t *testing.T) {
	_, _, err := Domain{}.Classify("")
	if err == nil {
		t.Error("Classify(\"\") should return error")
	}
}

func TestLocateUnknownType(t *testing.T) {
	_, err := Domain{}.Locate("image", "3v7")
	if err == nil {
		t.Error("Locate with unknown type should return error")
	}
}
