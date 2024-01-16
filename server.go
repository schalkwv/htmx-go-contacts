package main

import (
	"html/template"
	"net/http"

	"github.com/labstack/echo/v4"
)

type contact struct {
	ID    int    `json:"id"`
	First string `json:"first"`
	Last  string `json:"last"`
	Phone string `json:"phone"`
	Email string `json:"email"`
}

var Contacts []contact

func init() {
	Contacts = []contact{
		contact{ID: 1, First: "John", Last: "Doe", Phone: "555-555-5555", Email: "john@mail.com"},
		contact{ID: 2, First: "Jane", Last: "Doe", Phone: "555-555-5555", Email: "jane@mail.com"},
	}
}
func getContacts(c echo.Context) error {
	tmpl := template.Must(template.New("").ParseGlob("templates/*.gohtml"))

	err := tmpl.ExecuteTemplate(c.Response().Writer, "Base", Contacts)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return nil
}

func getNewContactForm(c echo.Context) error {
	tmpl := template.Must(template.New("").ParseGlob("templates/*.gohtml"))

	err := tmpl.ExecuteTemplate(c.Response().Writer, "NewContactPage", nil)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return nil
}

func main() {
	e := echo.New()
	e.GET("/", getContacts)
	e.GET("/contacts", getContacts)
	e.GET("/contacts/new", getNewContactForm)
	e.Logger.Fatal(e.Start(":1323"))
}
