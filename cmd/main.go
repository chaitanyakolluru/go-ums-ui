package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/carlmjohnson/requests"
	"github.com/chaitanyakolluru/go-ums-backend/pkg/model"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Data struct {
	Users []model.UserData
}

type Form struct {
	Values map[string]string
	Errors map[string]string
}

type Page struct {
	Data Data
	Form Form
}

type Templates struct {
	templates *template.Template
}

func NewForm() Form {
	return Form{
		Values: make(map[string]string),
		Errors: make(map[string]string),
	}
}

func NewPage() Page {
	ctx := context.Background()
	result := []model.UserData{}
	err := requests.
		URL(fmt.Sprintf("http://localhost:8080/users")).
		CheckStatus(http.StatusOK).
		ToJSON(&result).
		Fetch(ctx)

	if err != nil {
		return Page{
			Data: Data{
				Users: []model.UserData{},
			},
			Form: NewForm(),
		}
	}

	return Page{
		Data: Data{
			Users: result,
		},
		Form: NewForm(),
	}
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func NewTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Renderer = NewTemplate()

	e.GET("/", func(c echo.Context) error {
		page := NewPage()
		return c.Render(http.StatusOK, "index", page)
	})

	e.POST("/users", func(c echo.Context) error {
		name := c.FormValue("name")
		email := c.FormValue("email")

		ctx := context.Background()

		user := model.User{
			Name:  name,
			Email: email,
		}

		err := requests.
			URL(fmt.Sprintf("http://localhost:8080/users")).
			Accept("application/json").
			BodyJSON(&user).
			CheckStatus(http.StatusCreated).
			Fetch(ctx)

		if err != nil {
			formData := NewForm()
			formData.Values["name"] = name
			formData.Values["email"] = email
			formData.Errors["name"] = err.Error()

			return c.Render(422, "form", formData)
		}

		return c.Redirect(http.StatusFound, "/")
	})

	e.Logger.Fatal(e.Start(":1323"))
}
