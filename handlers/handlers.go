package handlers

import (
	"bytes"
	"log"
	"net/http"

	"github.com/yaq-cc/graffiti/cache"
	cx "github.com/yaq-cc/graffiti/godfcx"
)

type WebhookManager struct {
	cx.WebhookRequest
	cx.WebhookResponse
	Cache        *cache.TemplateCache
	Template     cache.Template
	VariablesMap map[string]string
}

func (wm *WebhookManager) Initialize(c *cache.TemplateCache, ep string, r *http.Request) {
	wm.Cache = c
	t, ok := c.Load(ep)
	if !ok {
		log.Fatal("No template found.")
	}
	wm.Template = t
	err := wm.FromRequest(r)
	if err != nil {
		log.Fatal(err)
	}
	// Why isn't MapSession Variables being promoted?
	wm.VariablesMap = wm.Template.MapSessionVariables(&wm.WebhookRequest)
}

func (wm *WebhookManager) ExecuteTemplate() bytes.Buffer {
	buf := wm.Template.Execute(wm.VariablesMap)
	return buf
}

func (wm *WebhookManager) MapCalculated(name, value string) {
	wm.VariablesMap[name] = value
}

func RegisterHandlers(c *cache.TemplateCache) {
	// for each template, register a handler.
	// 1 - create a handler (returns func(w, r))
	// 2 - execute http.HandlFunc(route, handler)
	// 3 - handle accepts a map of funcs which return string
	// 4 - funcs are called to generate calculated vars
	// 5 - only funcs with inferredargs are called

}

func TestEndpoint2Handler(c *cache.TemplateCache) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var wm WebhookManager
		defer r.Body.Close()

		// Calculated Value
		calcVar := "42"

		wm.Initialize(c, "/test_endpoint_2", r)
		wm.MapCalculated("UniversalAnswer", calcVar) // Use MapCalculated to include calculated variables.
		resp := wm.ExecuteTemplate()
		wm.TextResponse(w, resp.String())

	}
}

func TestEndpoint1Handler(c *cache.TemplateCache) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Release the request body resource.
		defer r.Body.Close()

		// Define WebhookRequest / WebhookResponse variables.
		var whreq cx.WebhookRequest
		var whresp cx.WebhookResponse

		// Load /test_endpoint_1 template.
		t, ok := c.Load("/test_endpoint_1")
		if !ok {
			log.Fatal("Could not find template.")
		}

		// Populate WebhookRequest from the http.Request body.
		err := whreq.FromRequest(r)
		if err != nil {
			log.Fatal(err)
		}
		// Map session variables to values provided by the request.
		v := t.MapSessionVariables(&whreq)

		// Add in calculated variables manually.
		v["UniversalAnswer"] = "42"

		// Execute the template (returns bytes.Buffer)
		m := t.Execute(v)

		// Write the response out.
		whresp.TextResponse(w, m.String())
	}
}

func GetAllHandler(c *cache.TemplateCache) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		c.CacheCopier(w)
	}
}
