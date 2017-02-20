package settings

type Settings struct {
	Exchange Exchange
	Services map[string]Service
}

type Exchange struct {
	Headers string
	Topic   string
}

type Service struct {
	Queue      string
	RoutingKey string
}
