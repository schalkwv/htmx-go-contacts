package main

import (
	"html/template"
	"net/http"

	"github.com/labstack/echo/v4"
)

type contact struct {
	ID    int    `json:"id" form:"ID"`
	First string `json:"first" form:"first_name"`
	Last  string `json:"last" form:"last_name"`
	Phone string `json:"phone" form:"phone"`
	Email string `json:"email" form:"email"`
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

func createContact(c echo.Context) error {
	var newContact contact
	err := c.Bind(&newContact)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	Contacts = append(Contacts, newContact)
	return c.Redirect(http.StatusMovedPermanently, "/contacts")
}

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/contacts")
	})
	e.GET("/contacts", getContacts)
	e.GET("/contacts/new", getNewContactForm)
	e.POST("/contacts/new", createContact)
	e.Logger.Fatal(e.Start(":1323"))
}
