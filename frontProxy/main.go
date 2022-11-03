package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
)

func Proxy(target string) gin.HandlerFunc {

	return func(c *gin.Context) {
		director := func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = target
			req.Host = target
		}
		proxy := &httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func main() {
	router := gin.Default()
	mainService := os.Getenv("MAIN_SERVICE_ADDRESS")
	if mainService == "" {
		log.Fatalln("Main service address is empty")
	} else {
		log.Printf("Main service address is %v", mainService)
	}
	calculationService := os.Getenv("CALCULATOR_SERVICE_ADDRESS")
	if calculationService == "" {
		log.Fatalln("Calculator service address is empty")
	} else {
		log.Printf("Calculator service address is %v", calculationService)
	}

	router.GET("/task/:task_name", Proxy(mainService))
	router.POST("/task", Proxy(mainService))
	router.POST("/task/:task_name", Proxy(mainService))
	router.DELETE("/task/:task_name", Proxy(mainService))

	router.GET("/work/:task_name/:work_name", Proxy(mainService))
	router.POST("/work/:task_name/:work_name", Proxy(mainService))
	router.POST("/work/:task_name", Proxy(mainService))
	router.DELETE("/work/:task_name/:work_name", Proxy(mainService))

	router.GET("/calculate/:task_name", Proxy(calculationService))

	err := router.Run(":8080")
	if err != nil {
		log.Fatalln(err)
	}
}
