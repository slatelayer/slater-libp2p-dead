package core

type session struct {
	id   string
	view view
}

func newSession(id string) session {
	return session{
		id:   id,
		view: newView(),
	}
}
