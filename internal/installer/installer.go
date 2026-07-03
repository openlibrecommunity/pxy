package installer

import (
	"fmt"
	"strings"
)

// Request describes one server install run.
type Request struct {
	Host      string    `json:"host"`
	SSHPort   string    `json:"sshPort"`
	User      string    `json:"user"`
	Password  string    `json:"password"`
	Domain    string    `json:"domain"`
	Email     string    `json:"email"`
	SNI       string    `json:"sni"`
	Protocols Protocols `json:"protocols"`
	Ports     Ports     `json:"ports"`
	OLCRTC    OLCRTC    `json:"olcrtc"`
}

// Protocols toggles install modules.
type Protocols struct {
	VLESS     bool `json:"vless"`
	Hysteria2 bool `json:"hysteria2"`
	Mieru     bool `json:"mieru"`
	AmneziaWG bool `json:"amneziawg"`
	Naive     bool `json:"naive"`
	OLCRTC    bool `json:"olcrtc"`
}

// Ports has protocol listen ports.
type Ports struct {
	VLESS     string `json:"vless"`
	Hysteria2 string `json:"hysteria2"`
	Mieru     string `json:"mieru"`
	AmneziaWG string `json:"amneziawg"`
	Naive     string `json:"naive"`
}

// OLCRTC configures olcrtc carrier and transport.
type OLCRTC struct {
	Provider  string `json:"provider"`
	Transport string `json:"transport"`
	Room      string `json:"room"`
}

// Defaults fills empty input fields.
func (r *Request) Defaults() {
	if r.SSHPort == "" {
		r.SSHPort = "22"
	}
	if r.User == "" {
		r.User = "root"
	}
	if r.Email == "" {
		r.Email = "admin@" + r.Domain
	}
	if r.SNI == "" {
		r.SNI = "www.microsoft.com"
	}
	if r.Ports.VLESS == "" {
		r.Ports.VLESS = "443"
	}
	if r.Ports.Hysteria2 == "" {
		r.Ports.Hysteria2 = "30000"
	}
	if r.Ports.Mieru == "" {
		r.Ports.Mieru = "444-448"
	}
	if r.Ports.AmneziaWG == "" {
		r.Ports.AmneziaWG = "39743"
	}
	if r.Ports.Naive == "" {
		r.Ports.Naive = "8443"
	}
	if r.OLCRTC.Provider == "" {
		r.OLCRTC.Provider = "jitsi"
	}
	if r.OLCRTC.Transport == "" {
		r.OLCRTC.Transport = "datachannel"
	}
	if r.OLCRTC.Room == "" {
		r.OLCRTC.Room = "https://meet.egovm.ru/pxy-" + r.Domain
	}
}

// Script returns a full remote bash script.
func Script(req Request) string {
	req.Defaults()
	var b strings.Builder
	writeHeader(&b, req)
	if req.Protocols.VLESS {
		b.WriteString(vlessBlock())
	}
	if req.Protocols.Hysteria2 {
		b.WriteString(hy2Block())
	}
	if req.Protocols.Mieru {
		b.WriteString(mieruBlock())
	}
	if req.Protocols.AmneziaWG {
		b.WriteString(awgBlock())
	}
	if req.Protocols.Naive {
		b.WriteString(naiveBlock())
	}
	if req.Protocols.OLCRTC {
		b.WriteString(olcrtcBlock())
	}
	b.WriteString("echo PXYSTARTRESULT\ncat /root/pxy/result.txt\n")
	return b.String()
}

func writeHeader(b *strings.Builder, req Request) {
	fmt.Fprintf(b, `#!/usr/bin/env bash
set -euo pipefail
export DEBIAN_FRONTEND=noninteractive
mkdir -p /root/pxy
: > /root/pxy/result.txt
DOMAIN=%s
SERVER=%s
EMAIL=%s
SNI=%s
VLESS_PORT=%s
HY2_PORT=%s
MIERU_PORT=%s
AWG_PORT=%s
NAIVE_PORT=%s
OLC_PROVIDER=%s
OLC_TRANSPORT=%s
OLC_ROOM=%s
log(){ printf '%%s\n' "$*"; }
res(){ printf '%%s\n' "$*" >> /root/pxy/result.txt; }
randhex(){ openssl rand -hex "$1"; }
pkg(){ apt-get update -y; apt-get install -y "$@"; }
rm -f /etc/apt/sources.list.d/amnezia.list /etc/apt/sources.list.d/testing.list /usr/share/keyrings/amnezia.gpg
log 'pxy: base packages'
pkg curl wget unzip openssl git ca-certificates python3
`, shq(req.Domain), shq(req.Host), shq(req.Email), shq(req.SNI), shq(req.Ports.VLESS), shq(req.Ports.Hysteria2), shq(req.Ports.Mieru), shq(req.Ports.AmneziaWG), shq(req.Ports.Naive), shq(req.OLCRTC.Provider), shq(req.OLCRTC.Transport), shq(req.OLCRTC.Room))
}

func vlessBlock() string {
	return `
log 'pxy: install vless xray'
bash -c "$(curl -L https://github.com/XTLS/Xray-install/raw/main/install-release.sh)" @ install
UUID=$(xray uuid)
XKEYS=$(xray x25519)
PRIV=$(printf '%s\n' "$XKEYS" | awk '/PrivateKey:/ {print $2}')
PUB=$(printf '%s\n' "$XKEYS" | awk '/Password \(PublicKey\):/ {print $3}')
SID=$(randhex 8)
cat > /usr/local/etc/xray/config.json <<EOF
{"log":{"loglevel":"warning"},"inbounds":[{"listen":"0.0.0.0","port":$VLESS_PORT,"protocol":"vless","settings":{"clients":[{"id":"$UUID","flow":""}],"decryption":"none"},"streamSettings":{"network":"xhttp","security":"reality","realitySettings":{"show":false,"dest":"$SNI:443","xver":0,"serverNames":["$SNI"],"privateKey":"$PRIV","shortIds":["$SID"]},"xhttpSettings":{"mode":"auto","host":"$SNI","path":"/"}}}],"outbounds":[{"protocol":"freedom","tag":"direct"},{"protocol":"blackhole","tag":"block"}]}
EOF
systemctl enable --now xray
systemctl restart xray
res "vless://$UUID@$DOMAIN:$VLESS_PORT?encryption=none&type=xhttp&security=reality&sni=$SNI&fp=chrome&pbk=$PUB&sid=$SID&host=$SNI&path=/&mode=auto#pxy-vless"
`
}

func hy2Block() string {
	return `
log 'pxy: install hysteria2'
bash <(curl -fsSL https://get.hy2.sh/)
HY_PASS=$(randhex 12)
HY_OBFS=$(randhex 16)
cat > /etc/hysteria/config.yaml <<EOF
listen: :$HY2_PORT
acme:
  domains:
    - $DOMAIN
  email: $EMAIL
auth:
  type: password
  password: $HY_PASS
obfs:
  type: gecko
  gecko:
    password: $HY_OBFS
masquerade:
  type: proxy
  proxy:
    url: https://zarazaex.xyz/
    rewriteHost: true
EOF
systemctl enable --now hysteria-server
systemctl restart hysteria-server
res "hysteria2://$HY_PASS@$DOMAIN:$HY2_PORT/?obfs=gecko&obfs-password=$HY_OBFS#pxy-hy2"
`
}

func mieruBlock() string {
	return `
log 'pxy: install mieru'
MIERU_VER=v3.34.0
curl -LSO https://github.com/enfein/mieru/releases/download/$MIERU_VER/mita_3.34.0_amd64.deb
dpkg -i mita_3.34.0_amd64.deb || apt-get -f install -y
rm -f mita_3.34.0_amd64.deb
MIERU_USER=pxy$(randhex 2)
MIERU_PASS=$(randhex 5)
cat > /tmp/mita.json <<EOF
{"portBindings":[{"portRange":"$MIERU_PORT","protocol":"TCP"}],"users":[{"name":"$MIERU_USER","password":"$MIERU_PASS"}],"loggingLevel":"INFO","mtu":1400}
EOF
mita apply config /tmp/mita.json
mita stop || true
mita start
systemctl enable --now mita
res "mieru tcp $DOMAIN ports $MIERU_PORT user $MIERU_USER pass $MIERU_PASS"
res "mieru client json: {\"profiles\":[{\"profileName\":\"pxy\",\"user\":{\"name\":\"$MIERU_USER\",\"password\":\"$MIERU_PASS\"},\"servers\":[{\"ipAddress\":\"$SERVER\",\"domainName\":\"$DOMAIN\",\"portBindings\":[{\"portRange\":\"$MIERU_PORT\",\"protocol\":\"TCP\"}]}],\"handshakeMode\":\"HANDSHAKE_STANDARD\"}],\"activeProfile\":\"pxy\",\"socks5Port\":1080,\"httpProxyPort\":8080,\"loggingLevel\":\"INFO\"}"
`
}

func awgBlock() string {
	return `
log 'pxy: install amneziawg'
if ! command -v awg >/dev/null; then
  pkg gpg sudo ethtool build-essential dkms dpkg-dev qrencode wireguard-tools linux-headers-amd64
  curl -fsSL 'https://keyserver.ubuntu.com/pks/lookup?op=get&search=0x75C9DD72C799870E310542E24166F2C257290828' | gpg --dearmor >/usr/share/keyrings/amnezia.gpg
  echo 'deb [signed-by=/usr/share/keyrings/amnezia.gpg] https://ppa.launchpadcontent.net/amnezia/ppa/ubuntu noble main' >/etc/apt/sources.list.d/amnezia.list
  apt-get update -y
  apt-get install -y amneziawg-dkms amneziawg-tools
  modprobe amneziawg || true
fi
mkdir -p /etc/amnezia/amneziawg /root/pxy/awg
SRV_PRIV=$(awg genkey)
SRV_PUB=$(printf '%s' "$SRV_PRIV" | awg pubkey)
CLI_PRIV=$(awg genkey)
CLI_PUB=$(printf '%s' "$CLI_PRIV" | awg pubkey)
Jc=3; Jmin=62; Jmax=157; S1=49; S2=54; S3=9; S4=12
H1=168771320-311865390; H2=404210777-749860699; H3=974164843-1203785257; H4=1253151579-2031452500
cat > /etc/amnezia/amneziawg/awg0.conf <<EOF
[Interface]
PrivateKey = $SRV_PRIV
Address = 10.9.9.1/24
ListenPort = $AWG_PORT
Jc = $Jc
Jmin = $Jmin
Jmax = $Jmax
S1 = $S1
S2 = $S2
S3 = $S3
S4 = $S4
H1 = $H1
H2 = $H2
H3 = $H3
H4 = $H4
I1 = <r 150>
PostUp = sysctl -w net.ipv4.ip_forward=1; iptables -t nat -A POSTROUTING -o $(ip route show default | awk '{print $5; exit}') -j MASQUERADE
PostDown = iptables -t nat -D POSTROUTING -o $(ip route show default | awk '{print $5; exit}') -j MASQUERADE

[Peer]
PublicKey = $CLI_PUB
AllowedIPs = 10.9.9.2/32
EOF
cat > /root/pxy/awg/pxy_phone.conf <<EOF
[Interface]
PrivateKey = $CLI_PRIV
Address = 10.9.9.2/32
DNS = 1.1.1.1, 1.0.0.1
MTU = 1280
Jc = $Jc
Jmin = $Jmin
Jmax = $Jmax
S1 = $S1
S2 = $S2
S3 = $S3
S4 = $S4
H1 = $H1
H2 = $H2
H3 = $H3
H4 = $H4
I1 = <r 150>

[Peer]
PublicKey = $SRV_PUB
Endpoint = $DOMAIN:$AWG_PORT
AllowedIPs = 0.0.0.0/0, ::/0
PersistentKeepalive = 33
EOF
sysctl -w net.ipv4.ip_forward=1
systemctl enable awg-quick@awg0
systemctl restart awg-quick@awg0 || log 'pxy: awg needs reboot - module built for newer kernel'
res 'amneziawg config:'
cat /root/pxy/awg/pxy_phone.conf >> /root/pxy/result.txt
`
}

func naiveBlock() string {
	return `
log 'pxy: install naiveproxy caddy'
if ! command -v go >/dev/null; then
  echo 'deb http://deb.debian.org/debian/ testing main non-free-firmware' >/etc/apt/sources.list.d/testing.list
  printf 'Package: *\nPin: release a=testing\nPin-Priority: 100\n' >/etc/apt/preferences.d/testing-pin
  apt-get update -y
  apt-get install -y -t testing golang-go
fi
export PATH="$HOME/go/bin:$PATH"
go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest
/root/go/bin/xcaddy build --with github.com/caddyserver/forwardproxy@caddy2=github.com/klzgrad/forwardproxy@naive
install -m 0755 caddy /usr/bin/caddy
setcap cap_net_bind_service=+ep /usr/bin/caddy 2>/dev/null || true
NAIVE_USER=pxy$(randhex 2)
NAIVE_PASS=$(randhex 5)
mkdir -p /etc/caddy
cat > /etc/caddy/Caddyfile <<EOF
:$NAIVE_PORT, $DOMAIN:$NAIVE_PORT
tls $EMAIL
route {
 forward_proxy {
   basic_auth $NAIVE_USER $NAIVE_PASS
   hide_ip
   hide_via
   probe_resistance
 }
 reverse_proxy https://zarazaex.xyz {
   header_up Host {upstream_hostport}
   header_up X-Forwarded-Host {host}
 }
}
EOF
cat > /etc/systemd/system/caddy.service <<EOF
[Unit]
Description=Caddy NaiveProxy
After=network-online.target
[Service]
Type=notify
User=root
ExecStart=/usr/bin/caddy run --environ --config /etc/caddy/Caddyfile
ExecReload=/usr/bin/caddy reload --config /etc/caddy/Caddyfile --force
Restart=always
RestartSec=5s
LimitNOFILE=1048576
AmbientCapabilities=CAP_NET_BIND_SERVICE
[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload
systemctl enable --now caddy
res "naive+https://$NAIVE_USER:$NAIVE_PASS@$DOMAIN:$NAIVE_PORT#pxy-naive"
`
}

func olcrtcBlock() string {
	return `
log 'pxy: install olcrtc'
MEM_MB=$(free -m | awk '/^Mem:/ {print $2}')
if [ "$MEM_MB" -lt 4096 ]; then
  fallocate -l 4G /swapfile 2>/dev/null && chmod 600 /swapfile && mkswap /swapfile && swapon /swapfile
fi
if ! command -v go >/dev/null; then
  echo 'deb http://deb.debian.org/debian/ testing main non-free-firmware' >/etc/apt/sources.list.d/testing.list
  printf 'Package: *\nPin: release a=testing\nPin-Priority: 100\n' >/etc/apt/preferences.d/testing-pin
  apt-get update -y
  apt-get install -y -t testing golang-go
fi
export PATH="$HOME/go/bin:$PATH"
go install github.com/magefile/mage@latest
mkdir -p /root/pj /root/.config/olcrtc
if [ ! -d /root/pj/olcrtc/.git ]; then
  git clone https://github.com/openlibrecommunity/olcrtc /root/pj/olcrtc
  cd /root/pj/olcrtc
else
  cd /root/pj/olcrtc && git pull --recurse-submodules 2>/dev/null || true
fi
/root/go/bin/mage build
OLC_KEY=$(randhex 32)
cat > /root/.config/olcrtc/server.yaml <<EOF
mode: srv
auth:
  provider: $OLC_PROVIDER
room:
  id: "$OLC_ROOM"
crypto:
  key: "$OLC_KEY"
net:
  transport: $OLC_TRANSPORT
  dns: "8.8.8.8:53"
liveness:
  interval: 10s
  timeout: 5s
  failures: 3
data: data
debug: false
EOF
cat > /etc/systemd/system/olcrtc.service <<EOF
[Unit]
Description=olcrtc server
After=network-online.target
[Service]
Type=simple
WorkingDirectory=/root/pj/olcrtc
ExecStart=/root/pj/olcrtc/build/olcrtc-linux-amd64 /root/.config/olcrtc/server.yaml
Restart=always
RestartSec=5s
LimitNOFILE=1048576
[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload
systemctl enable --now olcrtc
res "olcrtc://$OLC_PROVIDER?$OLC_TRANSPORT@$OLC_ROOM#$OLC_KEY\$pxy-olcrtc"
`
}

func shq(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'" }
