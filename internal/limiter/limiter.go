package limiter

type Limiter interface {
	Allow() bool
	Remaining() int
	Requests() int
}
