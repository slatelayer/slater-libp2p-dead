package core

type view struct {
	layout []string
	slates map[string]slate
	pages  map[string]page
}

type page struct {
	start int
	end   int
}

func newView() view {
	setup := newEphemeralSlate("setup")

	return view{
		layout: []string{"setup"},
		slates: map[string]slate{
			"setup": setup,
		},
		pages: map[string]page{
			"setup": {0, 0},
		},
	}
}
