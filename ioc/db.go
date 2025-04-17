package ioc

import (
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/prometheus"
	"project/internal/repository/dao"
	"project/pkg/gromx"
)

/**
 * @Description
 * @Date 2024/3/12 19:00
 **/

func InitDB() *gorm.DB {
	dsn := "root:123456@tcp(127.0.0.1:3306)/webook?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("数据库驱动错误")
	}

	// 初始化 Prometheus插件 监控 用来监控 MySQL 数据库指标
	prometheusCollector := prometheus.New(prometheus.Config{
		DBName: "webook",
		//指标刷新的时间间隔
		RefreshInterval: 15,
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.MySQL{
				VariableNames: []string{"thread_running"},
			},
		},
	})
	// grom的插件机制 就是实现了gorm.Plugin接口 然后在使用的时候调用Use方法
	err = db.Use(prometheusCollector)
	if err != nil {
		panic("注册 Prometheus 监控失败")
	}

	cb := gromx.NewCallbacks(prometheus2.SummaryOpts{
		Namespace: "geektime_daming",
		Subsystem: "webook",
		Name:      "gorm_db",
		Help:      "统计 GORM 的数据库查询",
		ConstLabels: map[string]string{
			"instance_id": "my_instance",
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	})

	err = db.Use(cb)
	if err != nil {
		panic(err)
	}

	if err := dao.InitTables(db); err != nil {
		panic("创建数据库表错误")
	}
	return db
}
