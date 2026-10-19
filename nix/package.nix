{
  lib,
  buildGoModule,
  rev ? "dev",
}:

buildGoModule {
  pname = "ping";
  version = rev;

  src = lib.fileset.toSource {
    root = ./..;
    fileset = lib.fileset.difference ./.. (
      lib.fileset.unions [
        ./.
        (./.. + "/.github")
        (./.. + "/flake.nix")
        (lib.fileset.maybeMissing (./.. + "/flake.lock"))
      ]
    );
  };

  # Update with script/update-vendor-hash after changing go.mod / go.sum.
  vendorHash = "sha256-QXxZfwz+1UXqnpMIz9R0RcpN9dmKEdB6f3X+wmTdHFY=";

  # cgo is required by go-sqlite3.
  env.CGO_ENABLED = "1";

  subPackages = [
    "cmd/ping"
    "cmd/ping-healthcheck"
    "cmd/ping-initialize-db"
  ];

  # Tests need a running MySQL (see script/test-env); CI runs them separately.
  doCheck = false;

  meta = {
    description = "Small analytics ping server";
    homepage = "https://github.com/parkr/ping";
    license = lib.licenses.mit;
    mainProgram = "ping";
  };
}
