# Hook examples
This page is still work in progress. Feel free to contribute!

## Incoming Github webhook
```hcl
server {
  hook "webhook" {
    constraints {
      all {
        expressions = [
          "${"refs/heads/master" == payload("ref")}",
          "${sha256(payload, "mysecret") == header("X-Signature")}",
        ]
      }
    }
    task {
      workdir = "/home/adnan/go"

      cmd = [
        "/home/adnon/redeploy-go-webhook.sh",
        "${payload("head_commit.id")}",
        "${payload("pusher.name")}",
        "${payload("pusher.email")}",
      ]
    }
  }
}
```

## Incoming Bitbucket webhook

Bitbucket does not pass any secrets back to the webhook.  [Per their documentation](https://confluence.atlassian.com/bitbucket/manage-webhooks-735643732.html#Managewebhooks-trigger_webhookTriggeringwebhooks), in order to verify that the webhook came from Bitbucket you must whitelist the IP range `104.192.143.0/24`:

```hcl
server {
  hook "webhook" {
    constraints {
      all {
        expressions = [
          "${request.RemoteAddr.WithinCIDR("104.192.143.0/24")}",
        ]
      }
    }
    task {
      workdir = "/home/adnan/go"

      cmd = [
        "/home/adnon/redeploy-go-webhook.sh",
        "${payload("actor.username")}",
      ]
    }
  }
}
```

## Incoming Gitlab Webhook

Gitlab provides webhooks for many kinds of events. 
Refer to this URL for example request body content: [gitlab-ce/integrations/webhooks](https://gitlab.com/gitlab-org/gitlab-ce/blob/master/doc/user/project/integrations/webhooks.md)
Values in the request body can be accessed in the command or to the match rule by referencing 'payload' as the source:

```hcl
server {
  hook "redeploy-webhook" {
    constraints {
      all {
        expressions = [
          "${header("X-Gitlab-Token") == "<YOUR-GENERATED-TOKEN>"}",
        ]
      }
    }
    task {
      workdir = "/home/adnan/go"

      cmd = [
        "/home/adnon/redeploy-go-webhook.sh",
        "${payload("user_name")}",
      ]
    }
  }
}
```

## Incoming Gogs webhook

```hcl
server {
  hook "webhook" {
    constraints {
      all {
        expressions = [
          "${"refs/heads/master" == payload("ref")}",
          "${sha256(payload, "mysecret") == header("X-Gogs-Signature")}",
        ]
      }
    }
    task {
      workdir = "/home/adnan/go"

      cmd = [
        "/home/adnon/redeploy-go-webhook.sh",
        "${payload("head_commit.id")}",
        "${payload("pusher.name")}",
        "${payload("pusher.email")}",
      ]
    }
  }
}
```

## Incoming Gitea webhook

```hcl
server {
  hook "webhook" {
    constraints {
      all {
        expressions = [
          "${"refs/heads/master" == payload("ref")}",
          "${"mysecret" == payload("secret")}",
        ]
      }
    }
    task {
      workdir = "/home/adnan/go"

      cmd = [
        "/home/adnon/redeploy-go-webhook.sh",
        "${payload("head_commit.id")}",
        "${payload("pusher.name")}",
        "${payload("pusher.email")}",
      ]
    }
  }
}
```

## Slack slash command

```hcl
server {
  hook "redeploy-webhook" {
    constraints {
      all {
        expressions = [
          "${payload("token") == "<YOUR-GENERATED-TOKEN>"}",
        ]
      }
    }
    task {
      workdir = "/home/adnan/go"

      cmd = [
        "/home/adnon/redeploy-go-webhook.sh",
        "${payload("user_name")}",
      ]
    }
    response {
      success {
        body = "Executing redeploy script"
      }
    }
  }
}
```

## A simple webhook with a secret key in GET query

__Not recommended in production due to low security__

`example.com:9000/hooks/simple-one` - won't work  
`example.com:9000/hooks/simple-one?token=42` - will work

```hcl
server {
  hook "simple-one" {
    constraints {
      all {
        expressions = [
          "${parameter("token") == "42"}",
        ]
      }
    }
    task {
      cmd = ["/path/to/command.sh"]
    }
    response {
      success {
        body = "Executing simple webhook..."
      }
    }
  }
}
```

# JIRA Webhooks

[Guide by @perfecto25](https://sites.google.com/site/mrxpalmeiras/notes/jira-webhooks)

# Pass File-to-command sample

## Webhook configuration

```hcl
server {
  hook "test-file-webhook" {
    task {
      workdir = "/tmp"
      cmd = ["/bin/ls"]

      pass_file {
        source = "payload"
        name = "binary"
        envname = "ENV_VARIABLE" // to use $ENV_VARIABLE in execute-command
                                 // if not defined, $HOOK_BINARY will be provided
        base64decode = true
      }
    }
    response {
      success {
        body = "${result.CombinedOutput}"
      }
    }
  }
}
```

## Sample client usage 

Store the following file as `testRequest.json`. 

<pre>
{"binary":"iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJbWFnZVJlYWR5ccllPAAAA2lpVFh0WE1MOmNvbS5hZG9iZS54bXAAAAAAADw/eHBhY2tldCBiZWdpbj0i77u/IiBpZD0iVzVNME1wQ2VoaUh6cmVTek5UY3prYzlkIj8+IDx4OnhtcG1ldGEgeG1sbnM6eD0iYWRvYmU6bnM6bWV0YS8iIHg6eG1wdGs9IkFkb2JlIFhNUCBDb3JlIDUuMC1jMDYwIDYxLjEzNDc3NywgMjAxMC8wMi8xMi0xNzozMjowMCAgICAgICAgIj4gPHJkZjpSREYgeG1sbnM6cmRmPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5LzAyLzIyLXJkZi1zeW50YXgtbnMjIj4gPHJkZjpEZXNjcmlwdGlvbiByZGY6YWJvdXQ9IiIgeG1sbnM6eG1wUmlnaHRzPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvcmlnaHRzLyIgeG1sbnM6eG1wTU09Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC9tbS8iIHhtbG5zOnN0UmVmPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvc1R5cGUvUmVzb3VyY2VSZWYjIiB4bWxuczp4bXA9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC8iIHhtcFJpZ2h0czpNYXJrZWQ9IkZhbHNlIiB4bXBNTTpEb2N1bWVudElEPSJ4bXAuZGlkOjEzMTA4RDI0QzMxQjExRTBCMzYzRjY1QUQ1Njc4QzFBIiB4bXBNTTpJbnN0YW5jZUlEPSJ4bXAuaWlkOjEzMTA4RDIzQzMxQjExRTBCMzYzRjY1QUQ1Njc4QzFBIiB4bXA6Q3JlYXRvclRvb2w9IkFkb2JlIFBob3Rvc2hvcCBDUzMgV2luZG93cyI+IDx4bXBNTTpEZXJpdmVkRnJvbSBzdFJlZjppbnN0YW5jZUlEPSJ1dWlkOkFDMUYyRTgzMzI0QURGMTFBQUI4QzUzOTBEODVCNUIzIiBzdFJlZjpkb2N1bWVudElEPSJ1dWlkOkM5RDM0OTY2NEEzQ0REMTFCMDhBQkJCQ0ZGMTcyMTU2Ii8+IDwvcmRmOkRlc2NyaXB0aW9uPiA8L3JkZjpSREY+IDwveDp4bXBtZXRhPiA8P3hwYWNrZXQgZW5kPSJyIj8+IBFgEwAAAmJJREFUeNqkk89rE1EQx2d/NNq0xcYYayPYJDWC9ODBsKIgAREjBmvEg2cvHnr05KHQ9iB49SL+/BMEfxBQKHgwCEbTNNIYaqgaoanFJi+rcXezye4689jYkIMIDnx47837zrx583YFx3Hgf0xA6/dJyAkkgUy4vgryAnmNWH9L4EVmotFoKplMHgoGg6PkrFarjXQ6/bFcLj/G5W1E+3NaX4KZeDx+dX5+7kg4HBlmrC6JoiDFYrGhROLM/mp1Y6JSqdCd3/SW0GUqEAjkl5ZyHTSHKBQKnO6a9khD2m5cr91IJBJ1VVWdiM/n6LruNJtNDs3JR3ukIW03SHTHi8iVsbG9I51OG1bW16HVasHQZopDc/JZVgdIQ1o3BmTkEnJXURS/KIpgGAYPkCQJPi0u8uzDKQN0XQPbtgE1MmrHs9nsfSqAEjxCNtHxZHLy4G4smUQgyzL4LzOegDGGp1ucVqsNqKVrpJCM7F4hg6iaZvhqtZrg8XjA4xnAU3XeKLqWaRImoIZeQXVjQO5pYp4xNVirsR1erxer2O4yfa227WCwhtWoJmn7m0h270NxmemFW4706zMm8GCgxBGEASCfhnukIW03iFdQnOPz0LNKp3362JqQzSw4u2LXBe+Bs3xD+/oc1NxN55RiC9fOme0LEQiRf2rBzaKEeJJ37ZWTVunBeGN2WmQjg/DeLTVP89nzAive2dMwlo9bpFVC2xWMZr+A720FVn88fAUb3wDMOjyN7YNc6TvUSHQ4AH6TOUdLL7em68UtWPsJqxgTpgeiLu1EBt1R+Me/mF7CQPTfAgwAGxY2vOTrR3oAAAAASUVORK5CYII="}
</pre>

use then the curl tool to execute a request to the webhook.

<pre>
#!/bin/bash
curl -H "Content-Type:application/json" -X POST -d @testRequest.json \
http://localhost:9000/hooks/test-file-webhook
</pre>

or in a single line, using https://github.com/jpmens/jo to generate the JSON code
<pre>
jo binary=%filename.zip | curl -H "Content-Type:application/json" -X POST -d @- \
http://localhost:9000/hooks/test-file-webhook
</pre>


## Incoming Scalr Webhook

[Guide by @hassanbabaie]
Scalr makes webhook calls based on an event to a configured webhook endpoint (for example Host Down, Host Up). Webhook endpoints are URLs where Scalr will deliver Webhook notifications.  
Scalr assigns a unique signing key for every configured webhook endpoint.
Refer to this URL for information on how to setup the webhook call on the Scalr side: [Scalr Wiki Webhooks](https://scalr-wiki.atlassian.net/wiki/spaces/docs/pages/6193173/Webhooks)
In order to leverage the Signing Key for addtional authentication/security you must configure the trigger rule with a match type of "scalr-signature".

```hcl
server {
  hook "redeploy-webhook" {
    constraints {
      all {
        expressions = [
          "${since(header("Date")) <= duration("300s")}",
          "${sha1(payload, "Scalr-provided signing key") == header("X-Signature")}",
        ]
      }
    }
    task {
      workdir = "/home/adnan/go"

      cmd = ["/home/adnon/redeploy-go-webhook.sh"]

      env_vars = {
        EVENT_NAME = "${payload("eventName")}"
        SERVER_HOSTNAME = "${payload("data.SCALR_SERVER_HOSTNAME")}"
      }
    }
    response {
      success {
        body = "${result.CombinedOutput}"
      }
    }
  }
}
```

## Travis CI webhook
Travis sends webhooks as `payload=<JSON_STRING>`, so the payload needs to be parsed as JSON. Here is an example to run on successful builds of the master branch.

```hcl
server {
  hook "deploy" {
    constraints {
      all {
        expressions = [
          "${payload("payload.state") == "passed"}",
          "${payload("payload.branch") == "master"}",
        ]
      }
    }
    request {
      json_parameters = ["payload"]
    }
    task {
      workdir = "/root/my-server"
      cmd = ["/root/my-server/deployment.sh"]
    }
  }
}
```
