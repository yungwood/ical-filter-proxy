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
  }: let
    # NixOS module that works across all systems
    nixosModule = args: import ./service.nix (args // {inherit self;});

    packageVersion = self.shortRev or "dev";
    packageRevision = self.shortRev or "unknown";

    mkIcalFilterProxy = pkgs:
      pkgs.buildGoModule {
        pname = "ical-filter-proxy";
        version = packageVersion;

        src = ./.;

        vendorHash = "sha256-OKOfPkssluK9uWE8BLZ9iyVpC1aLvQqzm6NyRyUYQY0=";

        ldflags = [
          "-s"
          "-w"
          "-X main.version=${packageVersion}"
          "-X main.commit=${packageRevision}"
        ];

        meta = with pkgs.lib; {
          description = "iCal proxy with support for user-defined filtering rules";
          homepage = "https://github.com/yungwood/ical-filter-proxy";
          license = licenses.mit;
          maintainers = [];
          platforms = platforms.unix;
        };
      };

    # Overlay to make ical-filter-proxy available in nixpkgs
    overlay = final: prev: {
      ical-filter-proxy = mkIcalFilterProxy prev;
    };
  in
    {
      # Export the NixOS module
      nixosModules.default = nixosModule;
      nixosModules.ical-filter-proxy = nixosModule;

      # Export the overlay
      overlays.default = overlay;
      overlays.ical-filter-proxy = overlay;
    }
    // flake-utils.lib.eachDefaultSystem (system: let
      pkgs = nixpkgs.legacyPackages.${system};

      ical-filter-proxy = mkIcalFilterProxy pkgs;
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

      # Basic check to ensure the package builds
      checks = {
        default = ical-filter-proxy;
      };
    });
}
