package main

import "strings"

type PickerViewMode string

const (
	ViewAll       PickerViewMode = "all"
	ViewFavorites PickerViewMode = "favorites"
)

type PickerItem struct {
	Name       string
	Scheme     Scheme
	Tag        string
	Favorite   bool
	UsageCount int
}

type PickerState struct {
	Query    string
	ViewMode PickerViewMode
	Items    []PickerItem
	Filtered []int
	Selected int
}

func NewPickerState(items []PickerItem, initialName string) *PickerState {
	state := &PickerState{
		ViewMode: ViewAll,
		Items:    append([]PickerItem(nil), items...),
	}
	state.refreshFiltered(initialName)
	return state
}

func (s *PickerState) SelectedItem() PickerItem {
	if len(s.Filtered) == 0 {
		return PickerItem{}
	}
	return s.Items[s.Filtered[s.Selected]]
}

func (s *PickerState) SetQuery(query string) {
	s.Query = query
	currentName := s.SelectedItem().Name
	s.refreshFiltered(currentName)
}

func (s *PickerState) ToggleFavorite() {
	if len(s.Filtered) == 0 {
		return
	}
	index := s.Filtered[s.Selected]
	s.Items[index].Favorite = !s.Items[index].Favorite
	currentName := s.Items[index].Name
	s.refreshFiltered(currentName)
}

func (s *PickerState) ToggleViewMode() {
	if s.ViewMode == ViewAll {
		s.ViewMode = ViewFavorites
	} else {
		s.ViewMode = ViewAll
	}
	s.refreshFiltered(s.SelectedItem().Name)
}

func (s *PickerState) refreshFiltered(preferredName string) {
	type match struct {
		index int
		score int
	}

	matches := make([]match, 0, len(s.Items))
	for i, item := range s.Items {
		if s.ViewMode == ViewFavorites && !item.Favorite {
			continue
		}
		score, ok := fuzzyScore(item.Name, s.Query)
		if !ok {
			continue
		}
		matches = append(matches, match{index: i, score: score})
	}

	// items are already stable in insertion order; preserve that inside same score bucket
	bestFiltered := make([]int, 0, len(matches))
	for score := 0; score <= 2; score++ {
		for _, match := range matches {
			if match.score == score {
				bestFiltered = append(bestFiltered, match.index)
			}
		}
	}

	s.Filtered = bestFiltered
	s.Selected = 0
	for i, index := range s.Filtered {
		if s.Items[index].Name == preferredName {
			s.Selected = i
			return
		}
	}
}

func fuzzyScore(name, query string) (int, bool) {
	if query == "" {
		return 0, true
	}

	name = strings.ToLower(name)
	query = strings.ToLower(query)

	if strings.HasPrefix(name, query) {
		return 0, true
	}
	if strings.Contains(name, query) {
		return 1, true
	}

	pos := 0
	for _, q := range query {
		found := false
		for pos < len(name) {
			if rune(name[pos]) == q {
				found = true
				pos++
				break
			}
			pos++
		}
		if !found {
			return 0, false
		}
	}
	return 2, true
}
