package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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

	// read contacts from contacts.json
	file, err := os.Open("contacts.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// decode json to contacts
	json.NewDecoder(file).Decode(&Contacts)

	// Contacts = []contact{
	// 	contact{ID: 1, First: "John", Last: "Doe", Phone: "555-555-5555", Email: "john@mail.com"},
	// 	contact{ID: 2, First: "Jane", Last: "Doe", Phone: "555-555-5555", Email: "jane@mail.com"},
	// }
}

func countContacts([]contact) int {
	time.Sleep(2 * time.Second)
	return len(Contacts)
}

func getContacts(c echo.Context) error {
	// time.Sleep(1 * time.Second)
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
		// templateParams.Count = countContacts(filteredContacts)

		tmpl := template.Must(template.New("").ParseGlob("templates/*.gohtml"))
		// tmpl := template.Must(template.New("").ParseGlob("templates/*.gohtml"))

		// check if headers contain HX-Trigger with value "search"
		// if yes, return only the contact list
		// if no, return the whole page
		if c.Request().Header.Get("HX-Trigger") == "search" {
			err := tmpl.ExecuteTemplate(c.Response().Writer, "ContactRows", templateParams)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, err.Error())
			}
			return nil
		}
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
		if !errors.As(err, &validationErrors) {
			return c.JSON(http.StatusBadRequest, err.Error())
		}

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
	if c.Request().Header.Get("HX-Trigger") == "delete-btn" {
		return c.Redirect(http.StatusSeeOther, "/contacts")
	}
	return nil
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

func getContactCount(c echo.Context) error {
	count := countContacts(Contacts)
	writer := c.Response().Writer
	writer.Write([]byte(strconv.Itoa(count)))
	return nil
}

type selectedContactIDs struct {
	SelectedContactIDs []int `form:"selected_contact_ids"`
}

func deleteContacts(c echo.Context) error {

	payload := new(selectedContactIDs)
	if err := c.Bind(payload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid payload")
	}
	// delete contacts with selected_contact_ids
	for _, id := range payload.SelectedContactIDs {
		for i, c := range Contacts {
			if c.ID == id {
				Contacts = append(Contacts[:i], Contacts[i+1:]...)
				break
			}
		}
	}

	// // Parse the form data
	// if err := c.Request().ParseForm(); err != nil {
	// 	return echo.NewHTTPError(http.StatusBadRequest, "Invalid form data")
	// }
	//
	// // Retrieve the selected contact IDs as a slice of strings
	// formValues := c.Request().Form
	// selectedContactIDs := formValues["selected_contact_ids"]
	//
	// // Initialize a slice to hold the converted integers
	// contactIDs := make([]int, 0, len(selectedContactIDs))
	//
	// // Convert each string to an integer
	// for _, idStr := range selectedContactIDs {
	// 	id, err := strconv.Atoi(idStr)
	// 	if err != nil {
	// 		// Handle the error as per your application's requirements
	// 		// For example, log the error and continue, or return a HTTP error
	// 		continue
	// 	}
	// 	contactIDs = append(contactIDs, id)
	// }
	// // // delete contacts with selected_contact_ids
	// for _, id := range selectedContactIDs {
	// 	for i, c := range Contacts {
	// 		if strconv.Itoa(c.ID) == id {
	// 			Contacts = append(Contacts[:i], Contacts[i+1:]...)
	// 			break
	// 		}
	// 	}
	// }
	templateParams := struct {
		Contacts []contact
		Search   string
	}{
		Contacts: Contacts,
		Search:   "",
	}

	tmpl := template.Must(template.New("").ParseGlob("templates/*.gohtml"))
	err := tmpl.ExecuteTemplate(c.Response().Writer, "Base", templateParams)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return nil
}

type examPayload struct {
	IDs []int `form:"ids"`
}

func createExam(c echo.Context) error {
	payload := new(examPayload)
	err := c.Bind(payload)
	if err != nil {
		return echo.ErrBadRequest
	}

	if len(payload.IDs) == 0 {
		return echo.ErrBadRequest
	}

	fmt.Println(payload.IDs)
	return nil
}
func main() {
	e := echo.New()
	e.Static("/static", "static")
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/contacts")
	})
	e.GET("/contacts", getContacts)
	e.POST("/deletecontacts", deleteContacts)
	e.GET("/contacts/count", getContactCount)
	e.GET("/contactlist", getContactList)
	e.GET("/contacts/new", getNewContactForm)
	e.POST("/contacts/new", createContact)
	e.GET("/contacts/:id", getViewContactForm)
	e.GET("/contacts/:id/edit", getEditContactForm)
	e.POST("/contacts/:id/edit", updateContact)
	// e.POST("/contacts/:id/delete", deleteContact)
	e.DELETE("/contacts/:id", deleteContact)
	e.GET("/contacts/:id/email", validateEmail)
	e.POST("/exams", createExam)
	e.DELETE("/exams", createExam)

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
