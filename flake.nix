{
	description = "A very basic flake";

	inputs = {
		nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
		flake-utils.url = "github:numtide/flake-utils";
	};

	outputs = { self, nixpkgs, flake-utils }:
		flake-utils.lib.eachDefaultSystem (system: let
			pkgs = nixpkgs.legacyPackages.${system};
		in {
			packages = rec {
				default = dht-prometheus-exporter;
				dht-prometheus-exporter = pkgs.buildGoModule {
					pname = "dht-prometheus-exporter";
					version = self.rev or "dirty";
					src = self;

					vendorHash = "sha256-x7gOb/Ze0RyOw94ZplE3id51sIgab4F3hPjF8uutUlg=";

					meta = {
						description = "DHT11/DHT22 Prometheus Exporter in Go.";
						homepage = "https://libdb.so/dht-prometheus-exporter";
						mainProgram = "dht-prometheus-exporter";
					};
				};
			};
			devShells.default = pkgs.mkShell {
				packages = with pkgs; [
					go_1_22
					gopls
					gotools
				];
			};
		});
}
