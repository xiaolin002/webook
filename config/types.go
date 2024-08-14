package config

/**
 * @Description
 * @Date 2024/8/14 20:09
 **/
type config struct {
	DB    DBConfig
	Redis RedisConfig
}

type DBConfig struct {
	DSN string
}

type RedisConfig struct {
	Addr string
}
