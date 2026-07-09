package pxyapp

const indexHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>pxy</title>
<style>
:root{color-scheme:light dark;--bg-color:Canvas;--text-color:CanvasText;--surface-color:color-mix(in oklab,Canvas 96%,CanvasText 6%);--surface-border:color-mix(in oklab,CanvasText 25%,transparent);--accent-color:inherit;--switcher-active:CanvasText}[data-theme=new]{color-scheme:dark;--bg-color:#090a0d;--text-color:#f2dec4;--surface-color:#191526;--surface-border:#a63232;--accent-color:#a63232;--switcher-active:#a63232}html,body{margin:0;padding:0;background:var(--bg-color);color:var(--text-color);font-family:ui-sans-serif,system-ui,-apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;line-height:1.6}main{max-width:780px;margin:0 auto;padding:20px;position:relative}.top-controls{position:absolute;top:20px;right:20px;display:flex;flex-direction:column;align-items:flex-end;gap:4px;font-size:.9rem}.top-controls a{color:var(--text-color);text-decoration:none;padding:2px 4px;opacity:.6}.top-controls a.active{font-weight:700;text-decoration:underline;opacity:1;color:var(--switcher-active)}a{color:var(--accent-color)}h1,h2{font-weight:600;margin:0 0 10px}h1{font-size:1.25rem}h2{font-size:1.05rem}section{margin:22px 0}form{background:var(--surface-color);border:1px solid var(--surface-border);padding:16px}label{display:block;margin:12px 0 6px}input,select,button{font:inherit;padding:8px;background:var(--bg-color);color:var(--text-color);border:1px solid var(--surface-border)}input,select{width:100%;box-sizing:border-box}button{cursor:pointer;margin-top:12px}.row{display:grid;grid-template-columns:1fr auto;gap:8px}.muted{opacity:.7;font-size:.85rem}.out{white-space:pre-wrap;background:var(--surface-color);border-left:3px solid var(--surface-border);padding:12px}@media(max-width:768px){main{padding:60px 20px 20px}.top-controls{top:10px;right:10px}}
</style>
</head>
<body>
<main>
<div class="top-controls"><div><a href="#" id="theme-old">old</a> <a href="#" id="theme-new">new</a></div><div><a href="#" id="back-link">back</a> <a href="https://zarazaex.xyz">zarazaex</a> <a href="https://github.com/openlibrecommunity/pxy">source</a></div></div>
<section>
<h1>pxy</h1>
<p>one click dns for proxy servers</p>
</section>
<form id="f">
<h2>1. enter api</h2>
<label for="ip">server ipv4</label>
<input id="ip" name="ip" inputmode="decimal" placeholder="2.2.2.2" required>
<h2>2. enter domain</h2>
<label for="sub">subdomain</label>
<div class="row"><input id="sub" name="sub" placeholder="x" required><button type="button" id="rnd">random</button></div>
<label for="zone">zone</label>
<select id="zone" name="zone">{{range .Domains}}<option value="{{.Name}}">{{.Name}}</option>{{end}}</select>
<h2>3. proof of work</h2>
<p class="muted">solve sha-256 challenge, difficulty: {{.Bits}} zero bits. keep this tab open.</p>
<button type="submit" id="go">create domain</button>
</form>
<p class="out" id="out">idle</p>
</main>
<script>
const $=id=>document.getElementById(id);
const enc=new TextEncoder();
function theme(t){document.documentElement.dataset.theme=t;localStorage.setItem('theme',t);$('theme-old').className=t==='old'?'active':'';$('theme-new').className=t==='new'?'active':''}
theme(localStorage.getItem('theme')||'old')
$('theme-old').onclick=e=>{e.preventDefault();theme('old')}
$('theme-new').onclick=e=>{e.preventDefault();theme('new')}
function clean(s){return s.toLowerCase().replace(/[^a-z0-9-]/g,'-').replace(/^-+|-+$/g,'')}
function rnd(){let a=new Uint8Array(8);crypto.getRandomValues(a);return [...a].map(x=>(x%36).toString(36)).join('')}
function hex(buf){return [...new Uint8Array(buf)].map(x=>x.toString(16).padStart(2,'0')).join('')}
function zeros(h){let n=0;for(const c of h){let v=parseInt(c,16);if(v===0){n+=4;continue}for(let m=8;m>0&&(v&m)===0;m>>=1)n++;break}return n}
async function sha(s){return hex(await crypto.subtle.digest('SHA-256',enc.encode(s)))}
async function solve(ch,ip,fqdn,bits){for(let i=0;;i++){if(i%5000===0)out.textContent='pow: '+i+' tries';let h=await sha(ch+':'+ip+':'+fqdn+':'+i);if(zeros(h)>=bits)return String(i)}}
$('rnd').onclick=()=>{$('sub').value=rnd()}
$('back-link').onclick=e=>{e.preventDefault();history.back()}
$('f').onsubmit=async e=>{e.preventDefault();go.disabled=true;try{let ip=$('ip').value.trim();let sub=clean($('sub').value.trim());let fqdn=sub+'.'+$('zone').value;out.textContent='challenge for '+fqdn;let r=await fetch('/challenge?ip='+encodeURIComponent(ip)+'&fqdn='+encodeURIComponent(fqdn));if(!r.ok)throw new Error(await r.text());let c=await r.json();let sol=await solve(c.challenge,ip,fqdn,c.bits);out.textContent='submitting solution '+sol;let fd=new FormData();fd.set('ip',ip);fd.set('fqdn',fqdn);fd.set('challenge',c.challenge);fd.set('solution',sol);let cr=await fetch('/create',{method:'POST',body:fd});out.textContent=await cr.text();if(!cr.ok)throw new Error(out.textContent)}catch(err){let m=err.message.replace(/<[^>]*>/g,'').replace(/\s+/g,' ').trim();out.textContent=m.length<100?m:'error: server unreachable (502)';}finally{go.disabled=false}}
</script>
</body>
</html>`
