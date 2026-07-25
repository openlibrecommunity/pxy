{ pkgs ? import <nixpkgs> { } }:

let
  # nixpkgs only ships webkit2gtk-4.1; Wails (without the webkit2_41 tag) asks
  # pkg-config for webkit2gtk-4.0. Provide a thin alias package so the build
  # works out of the box.
  webkit2gtk-4_0-alias = pkgs.runCommand "webkit2gtk-4.0-alias" { } ''
    mkdir -p $out/lib/pkgconfig
    cat > $out/lib/pkgconfig/webkit2gtk-4.0.pc <<EOF
    Name: webkit2gtk-4.0
    Description: Alias to webkit2gtk-4.1 (nixpkgs ships 4.1 only)
    Version: 4.0.0
    Requires: webkit2gtk-4.1
    EOF
    cat > $out/lib/pkgconfig/webkit2gtk-web-extension-4.0.pc <<EOF
    Name: webkit2gtk-web-extension-4.0
    Description: Alias to webkit2gtk-web-extension-4.1 (nixpkgs ships 4.1 only)
    Version: 4.0.0
    Requires: webkit2gtk-web-extension-4.1
    EOF
  '';
in

pkgs.mkShell {
  nativeBuildInputs = with pkgs; [
    go_1_26
    wails
    nodejs_22
    pkg-config
    gcc
    gobject-introspection
    wrapGAppsHook3
  ];

  buildInputs = with pkgs; [
    gtk3
    webkitgtk_4_1
    webkit2gtk-4_0-alias
    libsoup_3
    glib
  ];

  shellHook = ''
    export CGO_ENABLED=1
  '';
}
