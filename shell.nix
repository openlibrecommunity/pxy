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
    # Wails expects webkit2gtk-4.0; nixpkgs provides 4.1. Alias the pkg-config entry.
    pc_dir="$HOME/.cache/pxy-pkgconfig"
    mkdir -p "$pc_dir"
    cat > "$pc_dir/webkit2gtk-4.0.pc" <<'EOF'
Name: webkit2gtk-4.0
Description: Alias to webkit2gtk-4.1 (nixpkgs ships 4.1 only)
Version: 4.0.0
Requires: webkit2gtk-4.1
EOF
    cat > "$pc_dir/webkit2gtk-web-extension-4.0.pc" <<'EOF'
Name: webkit2gtk-web-extension-4.0
Description: Alias to webkit2gtk-web-extension-4.1 (nixpkgs ships 4.1 only)
Version: 4.0.0
Requires: webkit2gtk-web-extension-4.1
EOF
    export PKG_CONFIG_PATH="$pc_dir''${PKG_CONFIG_PATH:+:$PKG_CONFIG_PATH}"
  '';
}
