import './style.css'
import { Install, TestSSH, Update } from '../wailsjs/go/guiservice/App'
import { EventsOn, BrowserOpenURL, ClipboardSetText } from '../wailsjs/runtime/runtime'

const app = document.querySelector('#app')

app.innerHTML = `
<main>
  <header class="top">
    <section>
      <h1>pxy</h1>
      <p>one click proxy server installer</p>
    </section>
    <nav>
      <button id="old">old</button>
      <button id="new">new</button>
      <a href="https://pxy.zarazaex.xyz">dns</a>
      <a href="https://zarazaex.xyz">zarazaex</a>
    </nav>
  </header>
  <section class="panel panel-server">
    <h2>1. server</h2>
    <label>ip / host<input id="host" placeholder="2.26.103.97"></label>
    <label>ssh port<input id="sshPort" value="22"></label>
    <label>ssh user<input id="user" value="root"></label>
    <label>ssh password<input id="password" type="password"></label>
    <label>domain from pxy<input id="domain" placeholder="u68f32da.ikill.baby"></label>
    <label>email for acme<input id="email" placeholder="admin@example.com"></label>
    <label>sni / mask host<input id="sni" value="www.microsoft.com"></label>
    <div class="btn-row">
      <button type="button" id="test" disabled>test ssh</button>
      <button type="button" id="update" disabled>update (need restart)</button>
    </div>
  </section>
  <section class="panel panel-proto">
    <h2>2. protocols</h2>
    <div class="checks">
      <label><input type="checkbox" id="vless" checked> vless reality xhttp</label>
      <label><input type="checkbox" id="hysteria2" checked> hysteria2 gecko</label>
      <label><input type="checkbox" id="amneziawg" checked> amneziawg</label>
      <label><input type="checkbox" id="mieru"> mieru</label>
      <label><input type="checkbox" id="naive"> naiveproxy</label>
      <label><input type="checkbox" id="olcrtc"> olcrtc</label>
    </div>
    <h2>3. ports</h2>
    <div class="ports">
      <label>vless<input id="portVless" value="443"></label>
      <label>hy2<input id="portHy2" value="30000"></label>
      <label>awg<input id="portAwg" value="39743"></label>
      <label>mieru<input id="portMieru" value="444-448"></label>
      <label>naive<input id="portNaive" value="8443"></label>
    </div>
    <details>
      <summary>olcrtc cfg</summary>
      <label>provider<input id="olcProvider" value="jitsi"></label>
      <label>transport<input id="olcTransport" value="datachannel"></label>
      <label>room<input id="olcRoom" placeholder="https://meet.egovm.ru/room"></label>
    </details>
    <div class="btn-row">
      <button type="button" id="install" disabled>install selected</button>
    </div>
  </section>
  <section class="panel panel-logs">
    <div class="result-header">
      <h2>status <span class="state" id="state">idle</span></h2>
    </div>
    <pre id="logs"></pre>
  </section>
  <section class="panel panel-result">
    <div class="result-header">
      <h2>client data</h2>
      <button class="copy-btn" id="copyBtn">copy all</button>
    </div>
    <div id="configs"><p class="muted">after install links and configs will be here</p></div>
    <div class="links">
      <span>clients:</span>
      <a href="#" id="link-exclave">vless, mieru, naive, hy</a>
      <a href="#" id="link-olcbox">olcrtc</a>
      <a href="#" id="link-wt">olcrtc(alt)</a>
      <a href="#" id="link-awg">awg</a>
    </div>
  </section>
</main>`

const $ = (id) => document.getElementById(id)
const logs = $('logs')
const configs = $('configs')
const state = $('state')
let phase = 'init'
let lastResult = ''

$('link-exclave').onclick = (e) => { e.preventDefault(); BrowserOpenURL('https://github.com/ExclaveNetwork/Exclave/releases/') }
$('link-olcbox').onclick = (e) => { e.preventDefault(); BrowserOpenURL('https://github.com/alananisimov/olcbox') }
$('link-wt').onclick = (e) => { e.preventDefault(); BrowserOpenURL('https://github.com/spkprsnts/WireTurn') }
$('link-awg').onclick = (e) => { e.preventDefault(); BrowserOpenURL('https://github.com/amnezia-vpn/amnezia-client/releases/tag/4.8.19.0') }

function theme(t) {
  document.documentElement.dataset.theme = t
  localStorage.setItem('theme', t)
  $('old').className = t === 'old' ? 'active' : ''
  $('new').className = t === 'new' ? 'active' : ''
}

theme(localStorage.getItem('theme') || 'new')
$('old').onclick = () => theme('old')
$('new').onclick = () => theme('new')

EventsOn('install:log', (line) => {
  logs.textContent += line + '\n'
  logs.scrollTop = logs.scrollHeight
})

EventsOn('install:done', (text) => {
  showConfigs(text)
})

function showConfigs(text) {
  lastResult = text
  const parsed = parseConfigs(text)
  configs.innerHTML = ''
  if (parsed.length === 0) {
    configs.innerHTML = '<pre class="c-fallback">' + esc(text) + '</pre>'
    return
  }
  for (const c of parsed) {
    const row = document.createElement('div')
    row.className = 'c-row'
    const badge = document.createElement('span')
    badge.className = 'c-badge'
    badge.textContent = c.type
    const code = document.createElement('code')
    code.className = 'c-text'
    code.textContent = c.label.length > 80 ? c.label.slice(0, 77) + '...' : c.label
    const btn = document.createElement('button')
    btn.className = 'c-copy'
    btn.textContent = 'copy'
    btn.onclick = () => {
      ClipboardSetText(c.copy)
      btn.textContent = 'copied!'
      setTimeout(() => { btn.textContent = 'copy' }, 1200)
    }
    row.appendChild(badge)
    row.appendChild(code)
    row.appendChild(btn)
    configs.appendChild(row)
  }
}

function parseConfigs(text) {
  const lines = text.split('\n')
  const res = []
  let i = 0
  while (i < lines.length) {
    const l = lines[i].trim()
    if (!l) { i++; continue }
    if (l.startsWith('vless://')) {
      res.push({ type: 'vless', label: l, copy: l })
    } else if (l.startsWith('hysteria2://')) {
      res.push({ type: 'hy2', label: l, copy: l })
    } else if (l.startsWith('mieru tcp')) {
      res.push({ type: 'mieru', label: l, copy: l })
    } else if (l.startsWith('mieru client json:')) {
      res.push({ type: 'mieru-json', label: 'mieru client config', copy: l.replace('mieru client json: ', '') })
    } else if (l === 'amneziawg config:' || l.startsWith('amneziawg config:')) {
      i++; let awg = []
      while (i < lines.length && lines[i].trim() && !lines[i].includes('://') && !lines[i].startsWith('mieru')) {
        awg.push(lines[i]); i++
      }
      res.push({ type: 'awg', label: 'amneziawg config (' + awg.length + ' lines)', copy: awg.join('\n') })
      continue
    } else if (l.startsWith('naive+https://')) {
      res.push({ type: 'naive', label: l, copy: l })
    } else if (l.startsWith('olcrtc://')) {
      res.push({ type: 'olcrtc', label: l, copy: l })
    }
    i++
  }
  return res
}

function esc(s) {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}

function req() {
  return {
    host: $('host').value.trim(),
    sshPort: $('sshPort').value.trim(),
    user: $('user').value.trim(),
    password: $('password').value,
    domain: $('domain').value.trim(),
    email: $('email').value.trim(),
    sni: $('sni').value.trim(),
    protocols: {
      vless: $('vless').checked,
      hysteria2: $('hysteria2').checked,
      amneziawg: $('amneziawg').checked,
      mieru: $('mieru').checked,
      naive: $('naive').checked,
      olcrtc: $('olcrtc').checked
    },
    ports: {
      vless: $('portVless').value.trim(),
      hysteria2: $('portHy2').value.trim(),
      amneziawg: $('portAwg').value.trim(),
      mieru: $('portMieru').value.trim(),
      naive: $('portNaive').value.trim()
    },
    olcrtc: {
      provider: $('olcProvider').value.trim(),
      transport: $('olcTransport').value.trim(),
      room: $('olcRoom').value.trim()
    }
  }
}

function fieldsOk() {
  return $('host').value.trim() &&
    $('password').value &&
    $('domain').value.trim() &&
    ($('vless').checked || $('hysteria2').checked || $('amneziawg').checked ||
     $('mieru').checked || $('naive').checked || $('olcrtc').checked)
}

function refresh() {
  const ok = fieldsOk()
  $('test').disabled = !ok
  if (!ok || state.textContent !== 'ok') {
    $('update').disabled = true
    $('install').disabled = true
    return
  }
  $('update').disabled = phase === 'updated'
  $('install').disabled = phase !== 'updated'
}

function onCriticalChange() {
  phase = 'init'
  state.textContent = 'idle'
  refresh()
}

function onMinorChange() {
  refresh()
}

;['host','password','domain'].forEach(id => $(id).addEventListener('input', onCriticalChange))
;['vless','hysteria2','amneziawg','mieru','naive','olcrtc'].forEach(id => $(id).addEventListener('change', onMinorChange))

$('test').onclick = async () => {
  $('update').disabled = true
  $('install').disabled = true
  state.textContent = 'testing ssh...'
  state.textContent = await TestSSH(req())
  if (state.textContent === 'ok' && phase === 'updated') {
    $('install').disabled = false
  } else if (state.textContent === 'ok') {
    $('update').disabled = false
  }
}

$('update').onclick = async () => {
  $('update').disabled = true
  logs.textContent = ''
  state.textContent = 'updating & rebooting...'
  const r = await Update(req())
  logs.textContent += r + '\n'
  phase = 'updated'
  state.textContent = 'rebooted, test ssh to continue'
}

$('install').onclick = async () => {
  logs.textContent = ''
  configs.innerHTML = '<p class="muted">install running...</p>'
  state.textContent = 'installing'
  $('install').disabled = true
  try {
    const text = await Install(req())
    showConfigs(text)
    state.textContent = 'done'
  } catch (err) {
    state.textContent = 'err'
    configs.innerHTML = '<pre class="c-fallback">' + esc(String(err)) + '</pre>'
  } finally {
    $('install').disabled = false
  }
}

$('copyBtn').onclick = async () => {
  if (!lastResult) return
  ClipboardSetText(lastResult)
  $('copyBtn').textContent = 'copied!'
  setTimeout(() => { $('copyBtn').textContent = 'copy all' }, 1200)
}
