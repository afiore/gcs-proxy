package server

import "testing"

func TestReplaceEmptyBase(t *testing.T) {
	got := replaceEmptyBase("", "index.html")
	if got != "" {
		t.Errorf("replaceEmptyBase('') = %s; want ''", got)
	}

	got = replaceEmptyBase("/foo/bar/baz", "index.html")
	if got != "/foo/bar/baz" {
		t.Errorf("replaceEmptyBase('/foo/bar/baz') = %s; want /foo/bar/baz", got)
	}

	got = replaceEmptyBase("/foo/bar", "index.html")
	if got != "/foo/bar" {
		t.Errorf("replaceEmptyBase('/foo/bar') = %s; want /foo/index.html", got)
	}

}
