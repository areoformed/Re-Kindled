package main

import "testing"

func TestParseDocumentTitleAndPages(t *testing.T) {
	document, err := parseDocument("@title: A Small Letter\r\nfirst\r\n---PAGE---\r\nsecond\n", "test")
	if err != nil {
		t.Fatal(err)
	}
	if document.Title != "A Small Letter" {
		t.Fatalf("title = %q", document.Title)
	}
	if len(document.Pages) != 2 || document.Pages[0] != "first" || document.Pages[1] != "second" {
		t.Fatalf("pages = %#v", document.Pages)
	}
}

func TestParseDocumentWithoutTitle(t *testing.T) {
	document, err := parseDocument("ordinary first line\n", "test")
	if err != nil {
		t.Fatal(err)
	}
	if document.Title != "" || len(document.Pages) != 1 || document.Pages[0] != "ordinary first line" {
		t.Fatalf("document = %#v", document)
	}
}

func TestParseDocumentRejectsEmptyTitle(t *testing.T) {
	if _, err := parseDocument("@title:   \nbody", "test"); err == nil {
		t.Fatal("expected empty title to be rejected")
	}
}
