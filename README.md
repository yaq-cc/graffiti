# Graffiti: No Code updates for Webhook Fulfillment responses in realtime.  

Graffiti for Dialogflow CX decouples Webhook Response messages from the fulfillment server.  By decoupling the server's running code from response message templates we can empower the line of business to quickly update response messages from a browser instead of depending on tickets with internal IT or external partners who'd need (at least) a couple days just to process the request.  

The Graffiti server is stateless: it'll happily live in a persistent compute engine instance or you could run it as an auto-scaling serverless service like Cloud Run or Google Kubernetes Engine. 

To achieve statelessness, the response message templates live in Cloud Firestore, a NoSQL document database.  Why Firestore you ask? Because: (a) it really is easy to use and (b) it has built-in realtime updates, which Graffiti listens for.  Under the hood, the Graffiti server generates a cache of templates from Firestore and then listens for realtime updates and changes.  This concurrent-safe cache lets us respond to request without introducing extra round-trip latency (the templates are sitting in memory now instead of making calls Firestore for each request) while also managing any changes templates while Graffiti is serving up Webhook responses.  

Pretty cool, right?

# Graffiti Designer

Graffiti Designer is a command line tool used by the fulfillment server developer to provide template definitions for users to interact with.  Template definitions are defined in yaml.  They look something like this:

```yaml
agent-collection: dialogflow-agents
agent-name: graffiti-cx-demo-1
templates:
  /start: 
    handler-name: StartHandler
    calculated-variables:
      - Date
    prototype: "Hello there, it's {{.Date}}.  How are you doing?"
  /state: 
    handler-name: StateHandler
    session-variables:
      state: State
    prototype: "{{.State}} ... interesting.  Thanks for sharing.  How old are you?"
```

Once the definitions are staged into Firestore, the Graffiti GUI (a web based application) can be used by users to make updaates to the template's prototype - the string "template" that the fulfillment server will reply to a request with.  