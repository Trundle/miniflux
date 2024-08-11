{ lib, buildGoModule, installShellFiles, nixosTests }:

let
  pname = "miniflux";
  version = "2.1.4-andy";
in
buildGoModule {
  inherit pname version;

  src = ./.;
  vendorHash = "sha256-5ekZca9HF12PQ0CAhqwKpFsNhvQUMSWYHUHfPNETblE=";

  nativeBuildInputs = [ installShellFiles ];

  checkPhase =
    let
      disabledTests = [
        # I use tags differently
        "TestParseEntryWithCategories"
        "TestParseEntryWithMediaCategories"
        "TestParseFeedWithCategories"
        "TestParseFeedWithGooglePlayCategory"
        "TestParseFeedWithItunesCategories"
        "TestParseItemWithCategories"
        "TestParseItemTags"
      ];
    in
    ''
      go test "-skip=^${lib.concatStringsSep "$|^" disabledTests}$" $(go list ./... | grep -v client)
    ''; # skip client tests as they require network access

  ldflags = [
    "-s"
    "-w"
    "-X miniflux.app/version.Version=${version}"
  ];

  postInstall = ''
    mv $out/bin/miniflux.app $out/bin/miniflux
    installManPage miniflux.1
  '';

  passthru.tests = nixosTests.miniflux;

  meta = with lib; {
    description = "Minimalist and opinionated feed reader";
    homepage = "https://miniflux.app/";
    license = licenses.asl20;
    maintainers = with maintainers; [ rvolosatovs benpye ];
  };
}
