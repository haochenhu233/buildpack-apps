require "json"

app = ->(env) {
  [200, { "content-type" => "application/json" },
   [JSON.generate({ runtime: "ruby #{RUBY_VERSION}", ok: true })]]
}
run app
