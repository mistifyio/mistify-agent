This is an example sub-agent that posts to a webhook.  You could use
it to post guest creates/deletes for example.  This has been tested
with [Slack](https://api.slack.com/) but could be easily modified to
work with other services.


Example [runit](http://smarden.org/runit/) scripts:

Assuming you built and deployed to /usr/local/bin/mistify-webhook-post

Service run script
```
#!/bin/sh
# place in /etc/services/mistify-webhook-post/run
exec 2>&1
ulimit -n 8192
# this sub-agent does not require any special permissions
exec chpst -u nobody \
/usr/local/bin/mistify-webhook-post -p 31245 --endpoint https:://myendpoint.url/path
```

Log run script
```
#!/bin/sh
# place in /etc/services/mistify-webhook-post/log/run
exec 2>&1
mkdir -p /var/log/mistify-webhook-post
exec svlogd /var/log/mistify-webhook-post
```

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
