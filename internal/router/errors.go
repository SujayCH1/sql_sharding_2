package router

type RoutingErrorCode int

const (
	ErrInvalid RoutingErrorCode = iota
	ErrNoShardKey
	ErrShardKeyNotInQuery
	ErrUnsupportedPredicate
	ErrPolicyViolation
	ErrFanoutExceeded
)

type RoutingError struct {
	Code    RoutingErrorCode
	Message string
}
