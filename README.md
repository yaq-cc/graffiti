# graffiti

Graffiti for Dialogflow CX decouples Webhook Response messages from the fulfillment server.  By decoupling the server's running code from response message templates we can empower the line of business quickly update response messages from a browser instead of depending on tickets with internal IT or external partners who'd need a couple of days just to process the request.  

The Graffiti server is stateless: it'll happily live in a persistent compute engine instance or you could run it as an auto-scaling serverless service like Cloud Run or Google Kubernetes Engine. 

To achieve statelessness, the response message templates 
