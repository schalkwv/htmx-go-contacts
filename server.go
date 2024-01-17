package main

import (
	"errors"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type contact struct {
	ID     int    `json:"id" form:"ID"`
	First  string `json:"first" form:"first_name" validate:"required"`
	Last   string `json:"last" form:"last_name" validate:"required"`
	Phone  string `json:"phone" form:"phone" validate:"required"`
	Email  string `json:"email" form:"email" validate:"required,email"`
	Errors map[string]string
}

var Contacts []contact

func init() {
	Contacts = []contact{
		contact{ID: 1, First: "John", Last: "Doe", Phone: "555-555-5555", Email: "john@mail.com"},
		contact{ID: 2, First: "Jane", Last: "Doe", Phone: "555-555-5555", Email: "jane@mail.com"},
	}
}
func getContacts(c echo.Context) error {
	search := c.QueryParam("q")
	templateParams := struct {
		Contacts []contact
		Search   string
	}{
		Contacts: Contacts,
		Search:   search,
	}

	if search != "" {
		var filteredContacts []contact
		for _, c := range Contacts {

			if strings.Contains(c.First, search) || strings.Contains(c.Last, search) || strings.Contains(c.Phone, search) || strings.Contains(c.Email, search) {
				filteredContacts = append(filteredContacts, c)
			}
		}
		templateParams.Contacts = filteredContacts
		tmpl := template.Must(template.New("").ParseGlob("templates/*.gohtml"))
		// tmpl := template.Must(template.New("").ParseGlob("templates/*.gohtml"))

		err := tmpl.ExecuteTemplate(c.Response().Writer, "Base", templateParams)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		return nil
	}
	tmpl := template.Must(template.New("").ParseGlob("templates/*.gohtml"))

	err := tmpl.ExecuteTemplate(c.Response().Writer, "Base", templateParams)
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
	err := BindAndValidate(c, &newContact)
	if err != nil {
		var validationErrors validator.ValidationErrors
		errors.As(err, &validationErrors)

		newContact.Errors = make(map[string]string)
		for _, e := range validationErrors {
			newContact.Errors[e.Field()] = e.Tag()
		}

		tmpl := template.Must(template.New("").ParseGlob("templates/*.gohtml"))

		err := tmpl.ExecuteTemplate(c.Response().Writer, "NewContactPage", newContact)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
	} else {
		Contacts = append(Contacts, newContact)
		return c.Redirect(http.StatusMovedPermanently, "/contacts")
	}
	return nil
}

func getViewContactForm(c echo.Context) error {
	id := c.Param("id")
	var contact contact
	for _, c := range Contacts {
		if strconv.Itoa(c.ID) == id {
			contact = c
			break
		}
	}

	tmpl := template.Must(template.New("").ParseGlob("templates/*.gohtml"))

	err := tmpl.ExecuteTemplate(c.Response().Writer, "ViewContactPage", contact)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return nil
}

func getEditContactForm(c echo.Context) error {
	id := c.Param("id")
	var contact contact
	for _, c := range Contacts {
		if strconv.Itoa(c.ID) == id {
			contact = c
			break
		}
	}

	tmpl := template.Must(template.New("").ParseGlob("templates/*.gohtml"))

	err := tmpl.ExecuteTemplate(c.Response().Writer, "EditContactPage", contact)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return nil
}

func updateContact(c echo.Context) error {
	id := c.Param("id")
	var contact contact
	for _, c := range Contacts {
		if strconv.Itoa(c.ID) == id {
			contact = c
			break
		}
	}

	err := BindAndValidate(c, &contact)
	if err != nil {
		var validationErrors validator.ValidationErrors
		errors.As(err, &validationErrors)

		contact.Errors = make(map[string]string)
		for _, e := range validationErrors {
			contact.Errors[e.Field()] = e.Tag()
		}
		tmpl := template.Must(template.New("").ParseGlob("templates/*.gohtml"))

		err := tmpl.ExecuteTemplate(c.Response().Writer, "EditContactPage", contact)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
	} else {
		for i, c := range Contacts {
			if strconv.Itoa(c.ID) == id {
				Contacts[i] = contact
				break
			}
		}
		// redirect to contact view page
		return c.Redirect(http.StatusMovedPermanently, "/contacts/"+id)
		// return c.Redirect(http.StatusMovedPermanently, "/contacts")
	}
	return nil
}

func deleteContact(c echo.Context) error {
	id := c.Param("id")
	for i, c := range Contacts {
		if strconv.Itoa(c.ID) == id {
			Contacts = append(Contacts[:i], Contacts[i+1:]...)
			break
		}
	}
	return c.Redirect(http.StatusSeeOther, "/contacts")
}

func getContactList(c echo.Context) error {
	tmpl := template.Must(template.New("").ParseGlob("templates/*.gohtml"))

	err := tmpl.ExecuteTemplate(c.Response().Writer, "ContactList", Contacts)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return nil
}

func validateEmail(c echo.Context) error {
	id := c.Param("id")
	email := c.QueryParam("email")
	// check if email is unique
	for _, c1 := range Contacts {
		if c1.Email == email && strconv.Itoa(c1.ID) != id {
			return c.HTML(http.StatusOK, "email already exists")
		}
	}

	return nil
}

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/contacts")
	})
	e.GET("/contacts", getContacts)
	e.GET("/contactlist", getContactList)
	e.GET("/contacts/new", getNewContactForm)
	e.POST("/contacts/new", createContact)
	e.GET("/contacts/:id", getViewContactForm)
	e.GET("/contacts/:id/edit", getEditContactForm)
	e.POST("/contacts/:id/edit", updateContact)
	// e.POST("/contacts/:id/delete", deleteContact)
	e.DELETE("/contacts/:id", deleteContact)
	e.GET("/contacts/:id/email", validateEmail)

	e.Validator = &Validator{validator: validator.New(validator.WithRequiredStructEnabled())}

	e.Logger.Fatal(e.Start(":1323"))
}

func BindAndValidate(c echo.Context, req interface{}) error {
	if err := c.Bind(req); err != nil {
		return err
	}
	if err := c.Validate(req); err != nil {
		return err
	}
	return nil
}

// Validator is a custom validator for Echo.
type Validator struct {
	validator *validator.Validate
}

// Validate validates the request according to the required tags.
// Returns HTTPError if the required parameter is missing in the request.
func (v *Validator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}
