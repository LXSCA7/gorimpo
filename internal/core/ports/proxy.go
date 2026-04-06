package ports

type ProxyProvider interface {
	GetProxy() (string, error)
	MarkInvalid(proxy string)
}
