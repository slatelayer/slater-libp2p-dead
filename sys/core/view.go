package core

import "slater/core/slate"

type view struct {
	layout []string
	slates map[string]slate.Slate
	pages  map[string]page
}

type page struct {
	start int
	end   int
}

func newView() view {
	setup := slate.NewEphemeralSlate("setup")

	return view{
		layout: []string{"setup"},
		slates: map[string]slate.Slate{
			"setup": setup,
		},
		pages: map[string]page{
			"setup": {0, 0},
		},
	}
}
