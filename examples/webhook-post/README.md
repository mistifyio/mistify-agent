# webhook-post

[![webhook-post](https://godoc.org/github.com/mistifyio/mistify-agent/examples/webhook-post?status.png)](https://godoc.org/github.com/mistifyio/mistify-agent/examples/webhook-post)

webhook-post is an example sub-agent that posts to a webhook. This has been
tested with Slack but could be easily modified to work with other services.


### Usage

The following arguments are understood:

    $ webhook-post -h
    Usage of webhook-post:
    -e, --endpoint="": webhook endpoint
    -p, --port=31245: listen port


### Adding To The Agent

Adding the `webhook` service and, as an example, adding the `Webhook.Post`
method as a stage of the create action:

    {
    "services": {
    	"someotherservice": {
    	"port": 8000
    	},
    	"webhook": {
    	"port": 31245
    	}
    },
    "actions": {
    	"create": {
    	"stages": [
    		{
    		"service": "someotherservice",
    		"method": "Some.Method"
    		},
    		{
    		"service": "webhook",
    		"method": "Webhook.Post",
    		"args": {
    			"emoji": "rocket",
    			"name": "creator"
    		}
    		}
    	]
    	}
    }
    }

The emoji and username can be overwritten via the `Webhook.Post` stage args.


--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
