# TODO

- [ ] Fuller example with webserver and mux
- [ ] How do we step through the contraints to show which rule failed?
- [ ] Reloading config on signal


## Configuration File Feature Parity with webhook v1

### Hook Properties

Most assume a `service.hook` prefix:

- [x] id = .id as hook block label
- [x] execute-command = .task.cmd
- [x] command-working-directory = .task.workdir
- [x] response-message = .response.success.body
- [x] response-headers = .response.success.headers
- [x] success-http-response-code = .response.success.status_code
- [x] incoming-payload-content-type = .request.content_type
- [x] http-methods = n/a; solve with contraints
- [x] include-command-output-in-response = .response.success.body = "${result.CombinedOutput}"
- [x] include-command-output-in-response-on-error = .response.error.body = "${result.CombinedOutput}"
- [x] parse-parameters-as-json = .request.json_parameters
- [x] pass-arguments-to-command = .task.cmd
- [x] pass-environment-to-command = .task.cmd
- [ ] pass-file-to-command =
- [x] trigger-rule = .contraints
- [x] trigger-rule-mismatch-http-response-code = .response.unsatisfied.status_code
- [x] trigger-signature-soft-failures = n/a; solve with contraints


### CLI Parameters

Most assume a `service` prefix:

- [x] -cert = .tls_certificate
- [x] -cipher-suites = .tls_ciphers
- [x] -debug = .debug
- [x] -header = *deprecate*
- [ ] -hotreload = n/a for config, but we need to support config reloading
- [x] -ip = .ip
- [x] -key = .tls_certificate_key
- [x] -logfile = .logfile
- [x] -nopanic = .nopanic
- [x] -pidfile = .pidfile
- [x] -port = .port
- [x] -secure = .secure
- [x] -setgid = .user
- [x] -setuid = .group
- [ ] -template = .deprecate; use "${env("foo")}"
- [x] -tls-min-version = .tls_protocols
- [x] -urlprefix = hook.id
- [x] -verbose = .verbose
- [x] -version = n/a
- [x] -x-request-id = .enable_xrequestid
- [x] -x-request-id-limit = .xrequestid_limit


### Rules

- [x] And = all() or and(); constraints[] evals as and()
- [x] Or = or() or any()
- [x] Not = not()
- [x] Multi-level = yep
- [x] Match value = eq(), ne()
- [x] Match regex = match(), find()
- [x] Match payload-hmac-sha1 = eq(sha1(payload, "secret"), header("X-Signature"))
- [x] Match payload-hmac-sha256 = eq(sha256(payload, "secret"), header("X-Signature"))
- [x] Match payload-hmac-sha512 = eq(sha512(payload, "secret"), header("X-Signature"))
- [x] Match ip-whitelist = cidr("10/8", "10.0.0.1")
- [x] Match scalr-signature = and(
        le(since(header("Date")), duration("5m")),
        eq(sha256(payload, "secret"), header("X-Signature")),
      )

## Enhancement Requests

- [x] #505 = X-forwarded-for in whitelist
      Use header() and cidr()
- [x] #406 = string formatting of cmd arguments
      Add format() with printf libc syntax
- [x] #336 = concat params in cmd
      Add concat()
- [x] #442 = dynamic URL paths
      Can use {variable} substitution in the hook ID
- [x] #358 = pass temp file name to cmd
      Should be trivial for config to support it
- [x] #349 = response-message-failed
      See hook.response sub-blocks
- [x] #267 = time-based match rule
      Use since() and duration()
- [x] #263 = use cmd exit code as response code
      Use result.exit_code
- [x] #152 = PROXY protocol support
      Add service.proxy_protocol on the config side
- [x] #148 = allow limiting hook concurrency
      Add service[.hook].max_concurrency on the config side
- [x] #190 = pass stdin to cmd
      Add hook.stdin = payload
- [x] #468 = read value from file
      Add readfile() function; security implications?

- [ ] #504 = Reference to any array element with match
      Have payload("foo.*.bar") return an array?
      May need a contains for collections (stdlib)
        Can we have a cty func that handles both strings and collections?

- [ ] #326 = Support setting flags from config
      Surely we can figure this out; see hashicorp projects
