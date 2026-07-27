const api = 'http://localhost:8086/v2', sessionsEl = document.querySelector('#sessions'), messagesEl = document.querySelector('#messages'),
    titleEl = document.querySelector('#title'), statusEl = document.querySelector('#status'),
    form = document.querySelector('#composer'), input = document.querySelector('#question');
let active = null, items = [];
const setStatus = t => statusEl.textContent = t;
const esc = t => String(t).replace(/[&<>]/g, c => ({'&': '&amp;', '<': '&lt;', '>': '&gt;'}[c]));

function renderMessages(messages = []) {
    messagesEl.innerHTML = messages.length ? '' : '';
    if (!messages.length) {
        messagesEl.innerHTML = '<div class="empty">问我市场、公司或财经新闻。</div>';
        return
    }
    messages.forEach(m => {
        const el = document.createElement('article');
        el.className = 'message ' + (m.role === 'assistant' ? 'assistant' : 'user');
        el.innerHTML = m.role === 'assistant' ? marked.parse(m.content || '') : esc(m.content || '');
        messagesEl.append(el)
    });
    messagesEl.scrollTop = messagesEl.scrollHeight
}

async function select(session) {
    active = session;
    titleEl.textContent = session.title;
    sessionsEl.querySelectorAll('.session').forEach(e => e.classList.toggle('active', e.dataset.id === session.session_id));
    setStatus('加载会话');
    const r = await fetch(`${api}/session/history?session_id=${encodeURIComponent(session.session_id)}`);
    renderMessages(r.ok ? await r.json() : []);
    setStatus('就绪')
}

function renderSessions() {
    sessionsEl.innerHTML = '';
    items.forEach(s => {
        const el = document.createElement('div');
        el.className = 'session';
        el.dataset.id = s.session_id;
        el.innerHTML = `${esc(s.title)}<small>${new Date(s.updated_at * 1000).toLocaleString()}</small>`;
        el.onclick = () => select(s);
        sessionsEl.append(el)
    })
}

async function load() {
    const r = await fetch(`${api}/session/list?offset=0&limit=30`);
    if (!r.ok) return;
    const data = await r.json();
    items = data.sessions || [];
    renderSessions();
    if (items[0]) select(items[0])
}

async function create() {
    const r = await fetch(`${api}/session/new`, {method: 'POST'});
    if (!r.ok) throw Error('创建会话失败');
    const s = await r.json();
    items.unshift(s);
    renderSessions();
    await select(s)
}

document.querySelector('#new').onclick = () => create().catch(e => setStatus(e.message));
form.onsubmit = async e => {
    e.preventDefault();
    const question = input.value.trim();
    if (!question) return;
    try {
        if (!active) await create();
        renderMessages([{role: 'user', content: question}]);
        input.value = '';
        setStatus('正在思考');
        const r = await fetch(`${api}/chat`, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({session_id: active.session_id, question})
        });
        if (!r.ok) throw Error(await r.text());
        const answer = await r.json();
        await select(active);
        setStatus('就绪')
    } catch (e) {
        setStatus(e.message)
    }
};
input.onkeydown = e => {
    if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        form.requestSubmit()
    }
};
load();
