package ports

type Metrics interface {
	RecordDiscarded(term string, reason string, count int)
	RecordValid(term string, count int)
}
