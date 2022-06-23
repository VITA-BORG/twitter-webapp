package main

import (
	"html/template"
	"path/filepath"
)

//newTemplateCache is a helper function that loads all HTML templates into a template cache, and returns a map of template names to template.
//This will make it easy to render templates in the future, since the templates will be in the cache already and you will not have to parse them for every request.
func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./ui/html/pages/*.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {

		name := filepath.Base(page)

		//Parse base template file and add to set
		tmpl, err := template.ParseFiles("./ui/html/base.html")
		if err != nil {
			return nil, err
		}

		//Parse partials and add to set
		tmpl, err = tmpl.ParseGlob("./ui/html/partials/*.html")
		if err != nil {
			return nil, err
		}

		//Parse page and add to set
		tmpl, err = tmpl.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		cache[name] = tmpl

	}
	return cache, nil
}
