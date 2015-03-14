// Package form implements helpers for filling forms.
package form

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/crackcomm/crawl"
)

// Form - Form structure.
type Form struct {
	action string
	values map[string]string

	page *crawl.Response
	form *goquery.Selection
}

// New - Creates new form.
func New() (form *Form) {
	return &Form{values: make(map[string]string)}
}

// GetAction - Gets form action.
// Form action is selected during SetSelector() from form action attribute.
func (form *Form) GetAction() string {
	return form.action
}

// GetData - Gets form values.
func (form *Form) GetData() map[string]string {
	return form.values
}

// HasValue - Checks if value for key in form data.
// Value can be empty and it will also be true.
func (form *Form) HasValue(key string) (ok bool) {
	_, ok = form.values[key]
	return
}

// SetPage - Sets html page containing form.
// It is used in SetSelector() for setting default form values
// and in SetSelect() for finding select node in a form.
func (form *Form) SetPage(page *crawl.Response) {
	form.page = page
}

// SetValue - Sets form value.
// Returns ok when value was set.
func (form *Form) SetValue(key, value string) bool {
	if key != "" {
		form.values[key] = value
		return true
	}
	return false
}

// SetValues - Copies values into form values.
// Current form values are not changed.
func (form *Form) SetValues(values map[string]string) (ok bool) {
	ok = true
	for key, value := range values {
		// Set only if ok is true
		// We want ok = false even if just one failed
		if k := form.SetValue(key, value); ok {
			ok = k
		}
	}
	return
}

// SetSelect - Sets value for input of type select
// Value is chosen by option's text (trimmed of space).
// Returns ok when value was set.
func (form *Form) SetSelect(name, text string) (ok bool) {
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
				ok = form.SetValue(name, value)
			}
		})
	})

	return
}

// SetSelector - Sets form selector and parses default values.
// At this point page has to be set using SetPage() method.
func (form *Form) SetSelector(selector string) {
	// Get form
	form.form = form.page.Query().Find(selector)

	// Set form action
	form.action, _ = form.form.Attr("action")

	// Set default values
	form.setDefaults(selector)
}

// setDefaults - Finds all inputs and selects
// and sets their default values.
func (form *Form) setDefaults(selector string) {
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
		form.SetValue(name, value)
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
		form.SetValue(name, value)
	})

	return
}
