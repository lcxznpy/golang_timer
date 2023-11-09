package conf

// RedisConfig 缓存配置
type RedisConfig struct {
	Network            string `mapstructure:"network"`
	Address            string `mapstructure:"address"`
	Password           string `mapstructure:"password"`
	MaxIdle            int    `mapstructure:"maxIdle"`
	IdleTimeoutSeconds int    `mapstructure:"idleTimeout"`
	// 连接池最大存活的连接数.
	MaxActive int `mapstructure:"maxActive"`
	// 当连接数达到上限时，新的请求是等待还是立即报错.
	Wait bool `mapstructure:"wait"`
}

type RedisConfigProvider struct {
	conf *RedisConfig
}

func NewRedisConfigProvider(conf *RedisConfig) *RedisConfigProvider {
	return &RedisConfigProvider{
		conf: conf,
	}
}

func (r *RedisConfigProvider) Get() *RedisConfig {
	return r.conf
}

var defaultRedisConfProvider *RedisConfigProvider

func DefaultRedisConfigProvider() *RedisConfigProvider {
	return defaultRedisConfProvider
}
