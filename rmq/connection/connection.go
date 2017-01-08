package connection
/*
import (
	"github.com/streadway/amqp"
)
*/
type Config struct {
	URL string
}

func Connection(config *Config) {
	println(config.URL)
}
