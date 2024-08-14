package config

/**
 * @Description
 * @Date 2024/8/14 20:10
 **/

var Config = config{
	DB: DBConfig{
		DSN: "root:root@tcp(localhost:13316)/webook",
	},
	Redis: RedisConfig{
		Addr: "localhost:6379",
	},
}
