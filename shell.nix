{ pkgs ? import <nixpkgs> { } }:

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
    libsoup_3
    glib
  ];

  shellHook = ''
    export CGO_ENABLED=1
  '';
}
