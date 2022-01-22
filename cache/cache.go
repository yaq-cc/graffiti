package cache

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"sync"
	"text/template"

	cx "graffiti-2/godfcx"

	"cloud.google.com/go/firestore"
)

type Template struct {
	Handler             string            `firestore:"handler-name,omitempty" json:"handler-name,omitempty"`
	CalculatedVariables []string          `firestore:"calculated-variables,omitempty" json:"calculated-variables,omitempty"`
	SessionVariables    map[string]string `firestore:"session-variables,omitempty" json:"mapped-variables,omitempty"`
	Prototype           string            `firestore:"prototype,omitempty" json:"prototype,omitempty"`
	template            *template.Template
	templateArgs        []string
}

func (t *Template) Compile(ep string) {
	t.template = template.Must(template.New(ep).Parse(t.Prototype))
}

func (t *Template) inferArgs() {
	re := regexp.MustCompile(`({{.[a-zA-Z]+}})`)
	e := re.FindAllString(t.Prototype, -1)
	t.templateArgs = make([]string, len(e))
	for i, exp := range e {
		t.templateArgs[i] = exp[3 : len(exp)-2]
	}
}

func (t *Template) Equals(o *Template) bool {
	h := t.Handler == o.Handler
	c := reflect.DeepEqual(t.CalculatedVariables, o.CalculatedVariables)
	s := reflect.DeepEqual(t.SessionVariables, o.SessionVariables)
	p := t.Prototype == o.Prototype

	// All must key fields must match.
	return h && c && s && p
}

func (t *Template) FromMap(ep string, m map[string]interface{}) {
	b, e := json.Marshal(m)
	if e != nil {
		log.Fatal(e)
	}
	e = json.Unmarshal(b, t)
	if e != nil {
		log.Fatal(e)
	}
	t.Compile(ep) // compile the template.
}

// Extracts Session Variables
func (t *Template) MapSessionVariables(wr *cx.WebhookRequest) map[string]string {
	vm := make(map[string]string)
	// sv = session variable name, tv = template variable name
	for sv, tv := range t.SessionVariables {
		val, ok := wr.SessionInfo.Parameters[sv]
		if ok {
			vm[tv] = val
		}
	}
	return vm
}

// Execute infers arguments from the provided prototype,
// checks the provided map for required arguments, and then
// Executes the template.
func (t *Template) Execute(d map[string]string) bytes.Buffer {
	t.inferArgs()
	for _, arg := range t.templateArgs {
		_, ok := d[arg]
		if !ok {
			log.Fatal("Missing argument:", arg)
		}
	}
	var b bytes.Buffer
	t.template.Execute(&b, d)
	return b
}

// Separated type for testing purposes.
type TemplateDefinitions map[string]Template

type TemplateCache struct {
	AgentName string
	Cache     TemplateDefinitions
	mu        sync.RWMutex
}

func (c *TemplateCache) Store(ep string, tmp Template) {
	c.mu.Lock()
	defer c.mu.Unlock()
	old, ok := c.Cache[ep]
	// Check if the template exists
	if !ok {
		fmt.Println("Adding template:", ep)
		c.Cache[ep] = tmp
	} else {
		// Check if the template is new.
		same := old.Equals(&tmp)
		if !same {
			fmt.Println("New template for:", ep)
			c.Cache[ep] = tmp
		}
	}
}

func (c *TemplateCache) Load(name string) (tmp Template, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	tmp, ok = c.Cache[name]
	return tmp, ok
}

// Listen listen's for updates to string template definitions in Firestore
// and updates the template cache in a concurrent safe manner.  It returns a
// channel which is used to sync completion of the initial data load.
func (c *TemplateCache) Listen(ctx context.Context, client *firestore.Client) {
	// Setup inter-goroutine communication
	snaps := make(chan *firestore.DocumentSnapshot)
	snapIt := client.Collection("go-testing").Doc("test-1").Snapshots(ctx)

	// Signal ready via channel
	var once sync.Once
	ready := make(chan struct{})
	readyFunc := func() {
		defer close(ready)
		ready <- struct{}{}

	}

	// Listen for new snapshots
	go func() {
		for {
			snap, err := snapIt.Next()
			if err != nil {
				log.Fatal(err)
				return
			} else {
				snaps <- snap
			}
		}
	}()

	// Process snapshots and signals.  This is the controller for TemplateCache.Listen.
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println(ctx.Err())
				return
			case snap := <-snaps:
				data := snap.Data()
				// Initialize Cache if not already initialized.
				if c.Cache == nil {
					c.Cache = make(map[string]Template, len(data))
				}
				for ep, val := range data {
					var t Template
					tm, ok := val.(map[string]interface{})
					if ok {
						t.FromMap(ep, tm)
						c.Store(ep, t)
					}
				}
				once.Do(readyFunc)
			}
		}
	}()
	<-ready
}
