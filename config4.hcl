debug = false
verbose = true

ip = "0.0.0.0"
port = 9000
secure = false

user = "www-nobody"
group = "www-nobody"

logfile = "foo.log"
nopanic = true
pidfile = "/var/run/foo.pid"

hostname = "foo.br"
tls_certificate = "foo.crt"
tls_certificate_key = "foo.key"
tls_protocols = ["TLSv1.2", "TLSv1.3"]
tls_ciphers = ["foo"]

enable_xrequestid = true
xrequestid_limit = 32

proxy_protocol = true

// Keep?
http_methods = ["POST"]

hook "PREFIX/webhook/{scan_id}" {
  constraints = [ // trigger-rule
    or(
      eq(getenv("FOO"), ""),
      eq(getenv("FOO"), "BAR"),
    ),
    eq(upper(request.method), "POST"),
    eq(lower(request.method), lower("POST")),
    le(since(header("Date")), duration("10m")),
    match("^refs/[^/]+/master", payload("ref")),
    eq(sha256(payload, "secret"), header("X-Signature")),
    any(
      all(
        scontains(header("X-Coral-Signature"), concat("sha1=", sha1(payload, "secret"))),
        debug(concat("sha1=", sha1(payload, "secret"))),
      ),
      all(
        scontains(header("X-Coral-Signature"), concat("sha256=", sha256(payload, "secret"))),
        debug(concat("sha256=", sha256(payload, "secret"))),
      ),
    ),
    eq(url("param1"), "foo"),
    not(match("HTTP.2", request.proto)),
    and(
      cidr("1.2.3.0/24", request.remote_ip),
      cidr("1.2.3.0/24", header(format("%s%s", "X-Forwarded-for", ""))),
    ),
    eq(find("foo.?", "seafood fool"), "food"),

    eq(concat("HMAC ", sha256(payload, "secret")), header("Authorization")), // #512 MS Teams

    eq(readfile("secrets_file"), "foobar\n"), // XXX
    eq("foobar", base64decode("Zm9vYmFy")),
    eq(base64encode("foobar"), "Zm9vYmFy"),

    lt(len("foo"), 3),
    ne(len("foo"), 3),

    eq("food", find("foo.?", "seafood fool")),
    // eq(findAll("foo.?", "seafood fool"), ["food", "fool"]),

    contains([1, 2, 3], 3),
  ]

  request {
    force_content_type = "application/json"
    json_parameters = ["foo", "bar", "baz"]
  }

  task {
    stdin = payload

    // execute-command & pass-arguments-to-command
    cmd = [
      "/issue217/vol-key-${gt(payload("newVolume"), payload("previousVolume")) ? "up" : "down"}.sh",
      "/home/adnon/redeploy-go-webhook.sh",
      payload("a"),
      payload("head_commit.id"),
      payload("pusher.name"),
      payload("pusher.email"),
    ]

    workdir = "/home/adnan/go" // command-working-directory

    env_vars = { // pass-environment-to-command
      EVENT_NAME = payload("a")
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

  response {
    unsatisfied {
      status_code = 444
      headers = { // response-headers
          Strict-Transport-Security = "max-age=63072000; includeSubDomains",
      }
    }

    success {
      status_code = 222
      headers = { // response-headers
          name = result.pid,
          Strict-Transport-Security = "max-age=63072000; includeSubDomains",
      }
      content_type = "application/json"
      body = result.CombinedOutput // include-command-output-in-response
    }

    error {
      status_code = result.exit_code
      headers = { // response-headers
          name = result.pid,
          Strict-Transport-Security = "max-age=63072000; includeSubDomains",
      }
      content_type = "application/json"
      body = result.CombinedOutput // include-command-output-in-response
    }
  }
}
