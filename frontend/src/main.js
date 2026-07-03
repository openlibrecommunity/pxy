import './style.css'
import { Install, TestSSH } from '../wailsjs/go/guiservice/App'
import { EventsOn } from '../wailsjs/runtime/runtime'

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
  <section class="grid">
    <form id="form" class="card">
      <h2>1. server</h2>
      <label>ip / host<input id="host" placeholder="2.26.103.97" required></label>
      <label>ssh port<input id="sshPort" value="22"></label>
      <label>ssh user<input id="user" value="root"></label>
      <label>ssh password<input id="password" type="password" required></label>
      <label>domain from pxy<input id="domain" placeholder="u68f32da.ikill.baby" required></label>
      <label>email for acme<input id="email" placeholder="admin@example.com"></label>
      <label>sni / mask host<input id="sni" value="www.microsoft.com"></label>
      <button type="button" id="test">test ssh</button>
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
      <button type="submit" id="install">install selected</button>
    </form>
    <section class="card side">
      <h2>status</h2>
      <div id="state" class="state">idle</div>
      <h2>logs</h2>
      <pre id="logs"></pre>
    </section>
  </section>
  <section class="card result">
    <h2>client data</h2>
    <pre id="result">after install links and configs will be here</pre>
  </section>
</main>`

const ids = (name) => document.getElementById(name)
const logs = ids('logs')
const result = ids('result')
const state = ids('state')

function theme(t) {
  document.documentElement.dataset.theme = t
  localStorage.setItem('theme', t)
  ids('old').className = t === 'old' ? 'active' : ''
  ids('new').className = t === 'new' ? 'active' : ''
}

theme(localStorage.getItem('theme') || 'new')
ids('old').onclick = () => theme('old')
ids('new').onclick = () => theme('new')

EventsOn('install:log', (line) => {
  logs.textContent += line + '\n'
  logs.scrollTop = logs.scrollHeight
})

EventsOn('install:done', (text) => {
  result.textContent = text
})

function req() {
  return {
    host: ids('host').value.trim(),
    sshPort: ids('sshPort').value.trim(),
    user: ids('user').value.trim(),
    password: ids('password').value,
    domain: ids('domain').value.trim(),
    email: ids('email').value.trim(),
    sni: ids('sni').value.trim(),
    protocols: {
      vless: ids('vless').checked,
      hysteria2: ids('hysteria2').checked,
      amneziawg: ids('amneziawg').checked,
      mieru: ids('mieru').checked,
      naive: ids('naive').checked,
      olcrtc: ids('olcrtc').checked
    },
    ports: {
      vless: ids('portVless').value.trim(),
      hysteria2: ids('portHy2').value.trim(),
      amneziawg: ids('portAwg').value.trim(),
      mieru: ids('portMieru').value.trim(),
      naive: ids('portNaive').value.trim()
    },
    olcrtc: {
      provider: ids('olcProvider').value.trim(),
      transport: ids('olcTransport').value.trim(),
      room: ids('olcRoom').value.trim()
    }
  }
}

ids('test').onclick = async () => {
  state.textContent = 'testing ssh...'
  state.textContent = await TestSSH(req())
}

ids('form').onsubmit = async (e) => {
  e.preventDefault()
  logs.textContent = ''
  result.textContent = 'install running...'
  state.textContent = 'installing'
  ids('install').disabled = true
  try {
    result.textContent = await Install(req())
    state.textContent = 'done'
  } catch (err) {
    state.textContent = 'err'
    result.textContent = String(err)
  } finally {
    ids('install').disabled = false
  }
}
