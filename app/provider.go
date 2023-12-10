package app

import (
	"context"
	"xtimer/pkg/log"

	"go.uber.org/dig"

	"xtimer/common/conf"

	"xtimer/pkg/bloom"
	"xtimer/pkg/cron"
	"xtimer/pkg/hash"
	"xtimer/pkg/mysql"
	"xtimer/pkg/promethus"
	"xtimer/pkg/redis"
	"xtimer/pkg/xhttp"

	timerdao "xtimer/dao/timer"

	taskdao "xtimer/dao/task"

	migratorservice "xtimer/service/migrator"

	webservice "xtimer/service/webserver"

	executorservice "xtimer/service/executor"

	triggerservice "xtimer/service/trigger"

	schedulerservice "xtimer/service/scheduler"

	monitorservice "xtimer/service/monitor"

	"xtimer/app/migrator"

	"xtimer/app/webserver"

	"xtimer/app/scheduler"

	"xtimer/app/monitor"
)

var (
	container *dig.Container
)

func init() {
	container = dig.New()

	provideConfig(container)
	providePKG(container)
	provideDAO(container)
	provideService(container)
	provideApp(container)
}

func provideConfig(c *dig.Container) {
	err := c.Provide(conf.DefaultMysqlConfProvider)
	err = c.Provide(conf.DefaultSchedulerAppConfProvider)
	err = c.Provide(conf.DefaultTriggerAppConfProvider)
	err = c.Provide(conf.DefaultWebServerAppConfProvider)
	err = c.Provide(conf.DefaultRedisConfigProvider)
	err = c.Provide(conf.DefaultMigratorAppConfProvider)
	log.InfoContextf(context.Background(), "errors : %v", err)
}

func providePKG(c *dig.Container) {
	err := c.Provide(bloom.NewFilter)
	err = c.Provide(hash.NewMurmur3Encryptor)
	err = c.Provide(hash.NewSHA1Encryptor)
	err = c.Provide(redis.GetClient)
	err = c.Provide(mysql.GetClient)
	err = c.Provide(cron.NewCronParser)
	err = c.Provide(xhttp.NewJSONClient)
	err = c.Provide(promethus.GetReporter)
	log.InfoContextf(context.Background(), "errors : %v", err)
}

func provideDAO(c *dig.Container) {
	err := c.Provide(timerdao.NewTimerDAO)
	err = c.Provide(taskdao.NewTaskDAO)
	err = c.Provide(taskdao.NewTaskCache)
	log.InfoContextf(context.Background(), "errors : %v", err)
}

func provideService(c *dig.Container) {
	err := c.Provide(migratorservice.NewWorker)
	err = c.Provide(migratorservice.NewWorker)
	err = c.Provide(webservice.NewTaskService)
	err = c.Provide(webservice.NewTimerService)
	err = c.Provide(executorservice.NewTimerService)
	err = c.Provide(executorservice.NewWorker)
	err = c.Provide(triggerservice.NewWorker)
	err = c.Provide(triggerservice.NewTaskService)
	err = c.Provide(schedulerservice.NewWorker)
	err = c.Provide(monitorservice.NewWorker)
	log.InfoContextf(context.Background(), "errors : %v", err)
}

func provideApp(c *dig.Container) {
	err := c.Provide(migrator.NewMigratorApp)
	err = c.Provide(webserver.NewTaskApp)
	err = c.Provide(webserver.NewTimerApp)
	err = c.Provide(webserver.NewServer)
	err = c.Provide(scheduler.NewWorkerApp)
	err = c.Provide(monitor.NewMonitorApp)
	log.InfoContextf(context.Background(), "errors : %v", err)
}

func GetSchedulerApp() *scheduler.WorkerApp {
	var schedulerApp *scheduler.WorkerApp
	if err := container.Invoke(func(_s *scheduler.WorkerApp) {
		schedulerApp = _s
	}); err != nil {
		panic(err)
	}
	return schedulerApp
}

func GetWebServer() *webserver.Server {
	var server *webserver.Server
	if err := container.Invoke(func(_s *webserver.Server) {
		server = _s
	}); err != nil {
		panic(err)
	}
	return server
}

func GetMigratorApp() *migrator.MigratorApp {
	var migratorApp *migrator.MigratorApp
	if err := container.Invoke(func(_m *migrator.MigratorApp) {
		migratorApp = _m
	}); err != nil {
		panic(err)
	}
	return migratorApp
}

func GetMonitorApp() *monitor.MonitorApp {
	var monitorApp *monitor.MonitorApp
	if err := container.Invoke(func(_m *monitor.MonitorApp) {
		monitorApp = _m
	}); err != nil {
		panic(err)
	}
	return monitorApp
}
