{
  description = "iCal Filter Proxy - A simple service for proxying multiple iCal feeds while applying user-defined filtering rules";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = nixpkgs.legacyPackages.${system};

      ical-filter-proxy = pkgs.buildGoModule {
        pname = "ical-filter-proxy";
        version = "0.1.0";

        src = ./.;

        vendorHash = "sha256-tdIHHUN9/Qg07wUKvwGw0Lsz6uNFTUR6CpBrxx3jNQg=";

        ldflags = [
          "-s"
          "-w"
          "-X main.version=${self.rev or "dev"}"
        ];

        meta = with pkgs.lib; {
          description = "iCal proxy with support for user-defined filtering rules";
          homepage = "https://github.com/yungwood/ical-filter-proxy";
          license = licenses.mit;
          maintainers = [];
          platforms = platforms.unix;
        };
      };
    in {
      packages = {
        default = ical-filter-proxy;
        ical-filter-proxy = ical-filter-proxy;
      };

      apps = {
        default = flake-utils.lib.mkApp {
          drv = ical-filter-proxy;
        };
        ical-filter-proxy = flake-utils.lib.mkApp {
          drv = ical-filter-proxy;
        };
      };

      devShells.default = pkgs.mkShell {
        buildInputs = with pkgs; [
          go
          gopls
          gotools
          go-tools
        ];
      };
    });
}
