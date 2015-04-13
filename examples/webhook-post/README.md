This is an example sub-agent that posts to a webhook.  You could use
it to post guest creates/deletes for example.  This has been tested
with [Slack](https://api.slack.com/) but could be easily modified to
work with other services.

## Adding to Agent ##

We can add this to the create pipeline something like:

```json
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
```

You can override the emoji and username if desired by using the args.
