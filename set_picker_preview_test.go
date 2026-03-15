package main

import "testing"

type fakePreviewApplier struct {
	applied []*ResolvedScheme
}

func (f *fakePreviewApplier) Apply(scheme *ResolvedScheme) error {
	f.applied = append(f.applied, scheme)
	return nil
}

func TestPreviewSessionApplySelectionDoesNotPersist(t *testing.T) {
	applier := &fakePreviewApplier{}
	original := ResolvedFromScheme(".coltty.toml", "gruvbox", BuiltinSchemes()["gruvbox"])
	session := NewPreviewSession(applier, original)

	if err := session.ApplySelection(ResolvedFromScheme(".coltty.toml", "dracula", BuiltinSchemes()["dracula"])); err != nil {
		t.Fatal(err)
	}

	if len(applier.applied) != 1 {
		t.Fatalf("expected one apply call, got %d", len(applier.applied))
	}
	if session.Current.SchemeName != "dracula" {
		t.Fatalf("expected current preview dracula, got %q", session.Current.SchemeName)
	}
}

func TestPreviewSessionCancelRestoresOriginalScheme(t *testing.T) {
	applier := &fakePreviewApplier{}
	original := ResolvedFromScheme(".coltty.toml", "gruvbox", BuiltinSchemes()["gruvbox"])
	session := NewPreviewSession(applier, original)

	if err := session.ApplySelection(ResolvedFromScheme(".coltty.toml", "dracula", BuiltinSchemes()["dracula"])); err != nil {
		t.Fatal(err)
	}
	if err := session.Cancel(); err != nil {
		t.Fatal(err)
	}

	if got := applier.applied[len(applier.applied)-1].SchemeName; got != "gruvbox" {
		t.Fatalf("expected restore to gruvbox, got %q", got)
	}
}

func TestPreviewSessionConfirmAppliesFinalScheme(t *testing.T) {
	applier := &fakePreviewApplier{}
	original := ResolvedFromScheme(".coltty.toml", "gruvbox", BuiltinSchemes()["gruvbox"])
	final := ResolvedFromScheme(".coltty.toml", "dracula", BuiltinSchemes()["dracula"])
	session := NewPreviewSession(applier, original)

	if err := session.Confirm(final); err != nil {
		t.Fatal(err)
	}

	if got := applier.applied[len(applier.applied)-1].SchemeName; got != "dracula" {
		t.Fatalf("expected confirm apply dracula, got %q", got)
	}
}
