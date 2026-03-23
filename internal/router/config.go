package router

type RouterConfig struct {
	AllowBroadcast    bool
	AllowRangeQueries bool
	MaxShardFanout    int
	MaxRangeShardSpan int
}

func DefaultRouterConfig() RouterConfig {
	return RouterConfig{
		AllowBroadcast:    true,
		AllowRangeQueries: false,
		MaxShardFanout:    4,
		MaxRangeShardSpan: 4,
	}
}
