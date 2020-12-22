server {
    ip = "0.0.0.0"
    port = 9000
    secure = false
    http_methods = ["POST"]

    hook "PREFIX/webhook" {
        constraints { // trigger-rule
          all {
            expressions = [
              "${since(header("Date")) <= duration("10m")}",
              "${match("^refs/[^/]+/master", payload("ref")) && sha256(payload, "secret") == header("X-Signature")}",
            ]
            any {
              expressions = [
                // https://github.com/adnanh/webhook/pull/355
                "${contains(header("X-Coral-Signature"), concat("sha1=", sha1(payload, "secret")))}",
                "${contains(header("X-Coral-Signature"), concat("sha256=", sha256(payload, "secret")))}",

                "${debug(concat("sha1=", sha1(payload, "secret")))}",
                "${debug(concat("sha256=", sha256(payload, "secret")))}",
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
            headers = { // response-headers
                Strict-Transport-Security = "max-age=63072000; includeSubDomains",
            }
          }

          success {
            status_code = 222
            headers = { // response-headers
                name = "${result.pid}",
                Strict-Transport-Security = "max-age=63072000; includeSubDomains",
            }
            content_type = "application/json"
            body = "${result.CombinedOutput}" // include-command-output-in-response
          }

          error {
            status_code = 555
            headers = { // response-headers
                name = "${result.pid}",
                Strict-Transport-Security = "max-age=63072000; includeSubDomains",
            }
            content_type = "application/json"
            body = "${result.CombinedOutput}" // include-command-output-in-response
          }
        }
    }
}
