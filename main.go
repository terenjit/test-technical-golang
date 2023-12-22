package main

import (
	"html/template"
	"io"
	"net/http"
	"test-technical-golang/handlers"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	e := echo.New()
	e.Debug = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Renderer = &Template{
		templates: template.Must(template.ParseGlob("html/*.html")),
	}
	e.POST("/login", handlers.Login)
	e.GET("/login", func(c echo.Context) error {
		return c.Render(http.StatusOK, "login.html", nil)
	})
	e.GET("/input", handlers.ShowInputPage)
	e.POST("/input", handlers.ProcessForm)
	e.GET("/output", handlers.ShowOutputPage)

	e.POST("/edit", handlers.EditPhoneNumber)
	e.POST("/update", handlers.UpdatePhoneNumber)
	e.POST("/delete", handlers.DeletePhoneNumber)

	e.Start(":8080")
}
