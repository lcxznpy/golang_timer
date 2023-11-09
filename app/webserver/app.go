package webserver

import (
	"fmt"
	"net/http"
	"sync"
	"xtimer/common/conf"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	swaggerFiles "github.com/swaggo/files"

	"github.com/gin-gonic/gin"
	gs "github.com/swaggo/gin-swagger"
)

type Server struct {
	sync.Once
	engine   *gin.Engine
	timerApp *TimerApp
	taskApp  *TaskApp

	timerRouter *gin.RouterGroup
	taskRouter  *gin.RouterGroup
	mockRouter  *gin.RouterGroup

	confProvider *conf.WebServerAppConfProvider
}

// 初始化server app
func NewServer(timer *TimerApp, task *TaskApp, confProvider *conf.WebServerAppConfProvider) *Server {
	s := Server{
		engine:       gin.Default(),
		timerApp:     timer,
		taskApp:      task,
		confProvider: confProvider,
	}

	s.engine.Use(CrosHandler())

	s.timerRouter = s.engine.Group("api/timer/v1")
	s.taskRouter = s.engine.Group("api/task/v1")
	s.mockRouter = s.engine.Group("api/mock/v1")

	s.RegisterBaseRouter()
	s.RegisterMockRouter()
	s.RegisterTimerRouter()
	s.RegisterTaskRouter()
	s.RegisterMonitorRouter()
	return &s
}

func (s *Server) Start() {
	conf := s.confProvider.Get()
	go func() {
		if err := s.engine.Run(fmt.Sprintf(":%d", conf.Port)); err != nil {
			panic(err)
		}
	}()
}

// swagger router
func (s *Server) RegisterBaseRouter() {
	s.engine.GET("/swagger/*any", gs.WrapHandler(swaggerFiles.Handler))
}

// timer router
// todo : timer router handler need initialize
func (s *Server) RegisterTimerRouter() {
	s.timerRouter.GET("/def")
	s.timerRouter.POST("/def", s.timerApp.CreateTimer)
	s.timerRouter.DELETE("/def")
	s.timerRouter.PATCH("/def")

	s.timerRouter.GET("/defs")
	s.timerRouter.GET("/defsByName")

	s.timerRouter.POST("/enable")
	s.timerRouter.POST("/unable")
}

// task router
// todo : task router handler need initialize
func (s *Server) RegisterTaskRouter() {
	s.taskRouter.GET("/records")
}

// mock router
func (s *Server) RegisterMockRouter() {
	s.mockRouter.Any("/mock", func(c *gin.Context) {
		c.JSON(http.StatusOK, struct {
			Word string `json:"word"`
		}{
			Word: "hello world!",
		})
	})
}

// 监管 router
func (s *Server) RegisterMonitorRouter() {
	s.engine.Any("/metrics", func(c *gin.Context) {
		promhttp.Handler().ServeHTTP(c.Writer, c.Request)
	})
}
