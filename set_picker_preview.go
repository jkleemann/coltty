package main

type PreviewApplier interface {
	Apply(*ResolvedScheme) error
}

type PreviewSession struct {
	applier  PreviewApplier
	Original *ResolvedScheme
	Current  *ResolvedScheme
}

func NewPreviewSession(applier PreviewApplier, original *ResolvedScheme) *PreviewSession {
	return &PreviewSession{
		applier:  applier,
		Original: original,
		Current:  original,
	}
}

func (s *PreviewSession) ApplySelection(scheme *ResolvedScheme) error {
	if err := s.applier.Apply(scheme); err != nil {
		return err
	}
	s.Current = scheme
	return nil
}

func (s *PreviewSession) Cancel() error {
	return s.ApplySelection(s.Original)
}

func (s *PreviewSession) Confirm(scheme *ResolvedScheme) error {
	return s.ApplySelection(scheme)
}
