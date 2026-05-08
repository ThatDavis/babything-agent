import './style.css';
import {GetConfig, SaveConfig, StartAgent, StopAgent, GetStatus} from '../wailsjs/go/main/App';

const statusEl = document.getElementById('status');
const form = document.getElementById('settings-form');
const saveBtn = document.getElementById('save-btn');
const stopBtn = document.getElementById('stop-btn');

const fields = {
    cloudUrl: document.getElementById('cloud-url'),
    rtspUrl: document.getElementById('rtsp-url'),
    agentToken: document.getElementById('agent-token'),
    turnUrl: document.getElementById('turn-url'),
    turnUsername: document.getElementById('turn-username'),
    turnPassword: document.getElementById('turn-password'),
};

function setStatus(text) {
    statusEl.textContent = text;
    statusEl.className = 'status ' + text.toLowerCase();
}

function loadConfig() {
    GetConfig()
        .then(cfg => {
            fields.cloudUrl.value = cfg.cloud_url || '';
            fields.rtspUrl.value = cfg.rtsp_url || '';
            fields.agentToken.value = cfg.agent_token || '';
            fields.turnUrl.value = cfg.turn_url || '';
            fields.turnUsername.value = cfg.turn_username || '';
            fields.turnPassword.value = cfg.turn_password || '';
        })
        .catch(err => console.error('Failed to load config:', err));
}

function checkStatus() {
    GetStatus()
        .then(status => setStatus(status))
        .catch(err => {
            console.error('Failed to get status:', err);
            setStatus('stopped');
        });
}

form.addEventListener('submit', (e) => {
    e.preventDefault();
    saveBtn.disabled = true;
    saveBtn.textContent = 'Saving...';

    const cfg = {
        cloud_url: fields.cloudUrl.value,
        rtsp_url: fields.rtspUrl.value,
        agent_token: fields.agentToken.value,
        turn_url: fields.turnUrl.value,
        turn_username: fields.turnUsername.value,
        turn_password: fields.turnPassword.value,
    };

    SaveConfig(cfg)
        .then(() => {
            setStatus('running');
            saveBtn.textContent = 'Save & Start';
            saveBtn.disabled = false;
        })
        .catch(err => {
            console.error('Failed to save config:', err);
            alert('Failed to save: ' + err);
            saveBtn.textContent = 'Save & Start';
            saveBtn.disabled = false;
        });
});

stopBtn.addEventListener('click', () => {
    StopAgent()
        .then(() => setStatus('stopped'))
        .catch(err => console.error('Failed to stop:', err));
});

// Initial load
setStatus('checking');
loadConfig();
checkStatus();
