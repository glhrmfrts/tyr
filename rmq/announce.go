package rmq

type Announce struct {
	channel  *Channel
	settings *Settings
	future   *ioengine.Future
}

func (a *Announce) DoAnnounce() {
}
