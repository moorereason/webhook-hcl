server {
    ip = "0.0.0.0"
    port = 9000
    secure = false

    hook "PREFIX/webhook" {
        constraints { // trigger-rule
          all {
            expressions = [
              "${since(header("Date")) <= duration("10m")}",
              "${match("^refs/[^/]+/master", payload("ref")) && sha256(payload, "secret") == header("X-Signature")}",
            ]
            any {
              expressions = [
                "${since(header("Date")) <= duration("10m")}",
                "${match("^refs/[^/]+/master", payload("ref")) && sha256(payload, "secret") == header("X-Signature")}",
              ]
            }
          }
        }

        task {
            // execute-command & pass-arguments-to-command
            cmd = [
              // "/issue217/vol-key-${payload("newVolume") > payload("previousVolume") ? "up" : "down"}.sh",
              "/home/adnon/redeploy-go-webhook.sh",
              "${payload("a")}",
              "${payload("head_commit.id")}",
              "${payload("pusher.name")}",
              " ${payload("pusher.email")}",
            ]

            workdir = "/home/adnan/go" // command-working-directory

            env_vars = { // pass-environment-to-command
              EVENT_NAME = "${payload("foo")}"
            }

            // pass-file-to-command - FIXME not sure how to define this yet...
            pass_file {
              source = "payload"
              name = "zippedBinary"
              filename = "binaryFile.zip"
              base64decode = true
              envname = "ENV_VAR"
              keep = false
            }
            create_file {
              // FIXME - HCL doesn't handle binary data very well.
              // This may not be workable.
              // content = base64decode(payload("zippedBinary"))
              content = [0]
              filename = "binaryFile.zip"
              keep = false
              envname = "ENV_VAR"
            }
        }

        request {
          content_type = "application/json"
          json_parameters = ["foo", "bar", "baz"]
        }

        response {
          unsatisfied_constraints {
            status_code = 444
          }

          success {
            headers = { // response-headers
                name = "${result.pid}",
            }
            content_type = "application/json"
            body = "${result.CombinedOutput}" // include-command-output-in-response
          }
        }
    }
}
