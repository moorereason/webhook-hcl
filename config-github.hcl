ip = "0.0.0.0"
port = 9000
secure = false
http_methods = ["POST"]

hook "PREFIX/webhook" {
  constraints = [
    eq(upper(request.method), "POST"),
    eq(sha256(payload, "mysecret"), header("X-Hub-Signature")),
    eq("refs/heads/master", payload("ref")),

    cidr("1.2.3.0/24", request.remote_ip),
  ]

  task {
    cmd = [
      "/home/adnon/redeploy-go-webhook.sh",
      payload("head_commit.id"),
      payload("pusher.name"),
      payload("pusher.email"),
    ]

    workdir = "/home/adnan/go" // command-working-directory
  }
}
