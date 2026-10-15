package picker

type Item interface {
	Render() string
	Search() string
}

type ItemSource[I Item] struct {
	items []I
}

func (s ItemSource[I]) Len() int {
	return len(s.items)
}

func (s ItemSource[I]) string(i int) string {
	return s.items[i].Search()
}

func getTitles[I Item](items []I) []string {
	strs := []string{}

	for _, i := range items {
		strs = append(strs, i.Render())
	}

	return strs

}
