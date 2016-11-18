// Package forms implements helpers for filling forms.
package forms

import (
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/crackcomm/crawl"
)

// Form - Form structure.
type Form struct {
	Action string
	Values url.Values

	page *crawl.Response
	form *goquery.Selection
}

// New - Creates new form.
// Page argument is a list because it's optional - only first argument is used.
func New(page ...*crawl.Response) (form *Form) {
	form = &Form{Values: make(url.Values)}
	if len(page) >= 1 {
		form.Page(page[0])
	}
	return
}

// NewSelector - Creates new form with selector.
func NewSelector(page *crawl.Response, selector string) (form *Form) {
	form = New(page)
	form.Selector(selector)
	return
}

// Page - Sets html page containing form.
// It is used in Selector() for setting default form values
// and in Select() for finding select node in a form.
func (form *Form) Page(page *crawl.Response) {
	form.page = page
}

// Select - Sets value for input of type select
// Value is chosen by option's text (trimmed of space).
// Returns ok when value was set.
func (form *Form) Select(name, text string) (ok bool) {
	// Get all select fields
	// Iterate and add value by option text
	form.form.Find("select").Each(func(_ int, s *goquery.Selection) {
		// Get select field name
		if n, _ := s.Attr("name"); n != name {
			return
		}

		// Select option by text
		// And set it's value
		s.Find("option").Each(func(_ int, o *goquery.Selection) {
			if strings.TrimSpace(o.Text()) == text {
				value, _ := o.Attr("value")
				form.Values.Set(name, value)
			}
		})
	})

	return
}

// Selector - Sets form selector and parses default values.
// At this point page has to be set using Page() method.
func (form *Form) Selector(selector string) {
	form.form = form.page.Query().Find(selector)
	form.Action, _ = form.form.Attr("action")
	form.selector(selector)
}

// selector - Finds all inputs and selects
// and sets their default values.
func (form *Form) selector(selector string) {
	// Iterate over all inputs and set their default value
	// If input is of kind radio or checkbox
	// Set value only if it's selected
	form.form.Find("input").Each(func(_ int, s *goquery.Selection) {
		ftype, _ := s.Attr("type")
		switch ftype {
		case "submit", "reset":
			return
		case "radio", "checkbox":
			if c, _ := s.Attr("checked"); c != "checked" {
				return
			}
		default:
		}

		name, _ := s.Attr("name")
		value, _ := s.Attr("value")
		form.Values.Set(name, value)
	})

	// Iterate over all select nodes
	// And set values of selected option
	// Or empty if none was found
	form.form.Find("select").Each(func(_ int, s *goquery.Selection) {
		// Get select
		name, _ := s.Attr("name")
		if name == "" {
			return
		}

		// Iterate over all options and get the selected one
		var value string
		s.Find("option").Each(func(_ int, o *goquery.Selection) {
			// Pass if value is already set
			if len(value) > 0 {
				return
			}

			// If option is selected
			if s, _ := o.Attr("selected"); s == "selected" {
				value, _ = o.Attr("value")
			}
		})

		// Set form value
		form.Values.Set(name, value)
	})

	return
}
