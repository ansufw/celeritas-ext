package render

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
	"github.com/justinas/nosurf"
)

type Render struct {
	Renderer   string
	RootPath   string
	Secure     bool
	Port       string
	ServerName string
	JetViews   *jet.Set
	Session    *scs.SessionManager
}

type TemplateData struct {
	StringMap       map[string]string
	IntMap          map[string]int
	FloatMap        map[string]float64
	BoolMap         map[string]bool
	Data            map[string]any
	Port            string
	CSRFToken       string
	ServerName      string
	IsSecure        bool
	IsAuthenticated bool
	Error           string
	Flash           string
}

func (c *Render) defaultData(td *TemplateData, r *http.Request) *TemplateData {
	td.IsSecure = c.Secure
	td.Port = c.Port
	td.ServerName = c.ServerName
	td.CSRFToken = nosurf.Token(r)
	if c.Session.Exists(r.Context(), "user_id") {
		td.IsAuthenticated = true
	}
	td.Error = c.Session.PopString(r.Context(), "error")
	td.Flash = c.Session.PopString(r.Context(), "flash")
	return td
}

func (c *Render) Page(w http.ResponseWriter, r *http.Request, view string, variables, data any) error {
	switch strings.ToLower(c.Renderer) {
	case "jet":
		return c.jetPage(w, r, view, variables, data)
	case "go":
		return c.goPage(w, r, view, data)
	default:
		return nil
	}
}

// jetPage renders a template using the Jet template engine.
func (c *Render) jetPage(w http.ResponseWriter, r *http.Request, templateName string, variables, data any) error {
	var vars jet.VarMap

	if variables == nil {
		vars = make(jet.VarMap)
	} else {
		vars = variables.(jet.VarMap)
	}

	td := &TemplateData{}
	if data != nil {
		td = data.(*TemplateData)
	}

	td = c.defaultData(td, r)

	t, err := c.JetViews.GetTemplate(templateName + ".jet")
	if err != nil {
		log.Println(err)
		return err
	}

	err = t.Execute(w, vars, td)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (c *Render) goPage(w http.ResponseWriter, r *http.Request, view string, templateData any) error {

	tmpl, err := template.ParseFiles(c.RootPath + "/views/" + view + ".page.tmpl")
	if err != nil {
		return err
	}

	td := &TemplateData{}
	if templateData != nil {
		td = templateData.(*TemplateData)
	}

	return tmpl.Execute(w, td)
}

// jetPage renders a template using the Jet template engine.
func (c *Render) JetPage(w http.ResponseWriter, r *http.Request, templateName string, variables, data any) error {
	var vars jet.VarMap

	if variables == nil {
		vars = make(jet.VarMap)
	} else {
		vars = variables.(jet.VarMap)
	}

	td := &TemplateData{}
	if data != nil {
		td = data.(*TemplateData)
	}

	t, err := c.JetViews.GetTemplate(templateName + ".jet")
	if err != nil {
		log.Println(err)
		return err
	}

	err = t.Execute(w, vars, td)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (c *Render) GoPage(w http.ResponseWriter, r *http.Request, view string, templateData any) error {

	tmpl, err := template.ParseFiles(c.RootPath + "/views/" + view + ".page.tmpl")
	if err != nil {
		return err
	}

	td := &TemplateData{}
	if templateData != nil {
		td = templateData.(*TemplateData)
	}

	return tmpl.Execute(w, td)
}
