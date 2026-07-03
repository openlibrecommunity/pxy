pxy - one click server bypass installer + free domains for bypass

two cmds:
  cmd/pxy      - dns web service (for pxy.zarazaex.xyz)
  main.go      - wails gui installer (desktop)

install:
  go build -o pxy . && ./pxy                     # linux
  GOOS=windows GOARCH=amd64 go build -o pxy.exe . # windows cross

deps:
  wails v2.12   - gui
  golang.org/x/crypto/ssh - remote install

protocols: vless reality xhttp | hysteria2 | amneziawg | mieru | naive | olcrtc
