with import <nixpkgs> {};

stdenv.mkDerivation {
  name = "waymond-tools";
  buildInputs = [
    go_1_19
    just
    nodejs-16_x
  ];
}