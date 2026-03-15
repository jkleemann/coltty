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

func TestPreviewStyleRolesDeriveFromScheme(t *testing.T) {
	scheme := BuiltinSchemes()["dracula"]

	roles := newPreviewStyleRoles(scheme)

	if roles.Keyword.GetForeground() == nil {
		t.Fatal("expected keyword foreground")
	}
	if roles.Heading.GetForeground() == nil {
		t.Fatal("expected heading foreground")
	}
}

func TestPreviewStyleRolesFallbackWithoutFullPalette(t *testing.T) {
	roles := newPreviewStyleRoles(Scheme{
		Foreground: "#eeeeee",
		Palette:    []string{"#111111", "#222222"},
	})

	if roles.Base.GetForeground() == nil {
		t.Fatal("expected base foreground fallback")
	}
	if roles.Keyword.GetForeground() == nil {
		t.Fatal("expected keyword foreground fallback")
	}
}
