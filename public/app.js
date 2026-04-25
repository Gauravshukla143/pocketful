const API_URL = 'http://localhost:8080';
let currentMode = 'login';
let token = localStorage.getItem('token');
let userRole = localStorage.getItem('role');
let userEmail = localStorage.getItem('email');

// Check if already logged in
if (token) {
    showDashboard();
}

function showToast(msg, isError = false) {
    const container = document.getElementById('toast-container');
    const toast = document.createElement('div');
    toast.className = 'toast';
    toast.innerText = msg;
    if (isError) toast.style.borderLeftColor = '#ff416c';
    container.appendChild(toast);
    setTimeout(() => toast.remove(), 3000);
}

function switchTab(mode) {
    currentMode = mode;
    document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
    event.target.classList.add('active');
    document.getElementById('auth-btn').innerText = mode === 'login' ? 'Login' : 'Register';
}

async function handleAuth(e) {
    e.preventDefault();
    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;

    try {
        const endpoint = currentMode === 'login' ? '/login' : '/register';
        const res = await fetch(API_URL + endpoint, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });
        const data = await res.json();

        if (!res.ok) throw new Error(data.error || 'Authentication failed');

        if (currentMode === 'register') {
            showToast('Registered! Please login now.');
            switchTab('login');
        } else {
            token = data.token;
            userRole = data.user.role;
            userEmail = data.user.email;
            localStorage.setItem('token', token);
            localStorage.setItem('role', userRole);
            localStorage.setItem('email', userEmail);
            showDashboard();
        }
    } catch (err) {
        showToast(err.message, true);
    }
}

function logout() {
    localStorage.clear();
    token = null;
    document.getElementById('dashboard-panel').classList.add('hidden');
    document.getElementById('auth-panel').classList.remove('hidden');
}

async function showDashboard() {
    document.getElementById('auth-panel').classList.add('hidden');
    document.getElementById('dashboard-panel').classList.remove('hidden');
    document.getElementById('user-email-display').innerText = userEmail;

    if (userRole === 'ADMIN') {
        document.getElementById('admin-section').classList.remove('hidden');
        document.getElementById('upload-section').classList.add('hidden');
        document.getElementById('kyc-status-badge').innerText = 'ADMIN MODE';
        document.getElementById('kyc-status-badge').className = 'badge verified';
        return;
    }

    await checkKYCStatus();
}

async function checkKYCStatus() {
    try {
        const res = await fetch(API_URL + '/kyc/status', {
            headers: { 'Authorization': 'Bearer ' + token }
        });
        
        const badge = document.getElementById('kyc-status-badge');
        const initiateBtn = document.getElementById('initiate-btn');
        const uploadSec = document.getElementById('upload-section');

        if (res.status === 404) {
            badge.innerText = 'NOT INITIATED';
            badge.className = 'badge';
            initiateBtn.classList.remove('hidden');
            uploadSec.classList.add('hidden');
            return;
        }

        const data = await res.json();
        const status = data.kyc.status;
        
        badge.innerText = status.replace('_', ' ');
        badge.className = 'badge ' + status.toLowerCase();
        
        initiateBtn.classList.add('hidden');
        
        if (status === 'PENDING') {
            uploadSec.classList.remove('hidden');
            // Check uploaded docs
            if (data.documents) {
                data.documents.forEach(doc => {
                    const box = document.getElementById('box-' + doc.type);
                    if (box) {
                        box.classList.add('uploaded');
                        box.querySelector('button').innerText = 'Re-upload';
                    }
                });
            }
        } else {
            uploadSec.classList.add('hidden');
        }
    } catch (err) {
        showToast('Failed to get KYC status', true);
    }
}

async function initiateKYC() {
    try {
        const res = await fetch(API_URL + '/kyc/initiate', {
            method: 'POST',
            headers: { 'Authorization': 'Bearer ' + token }
        });
        if (!res.ok) throw new Error('Failed to initiate');
        showToast('KYC Initiated!');
        checkKYCStatus();
    } catch (err) {
        showToast(err.message, true);
    }
}

async function uploadDoc(type) {
    const fileInput = document.getElementById('file-' + type);
    if (!fileInput.files[0]) {
        return showToast('Please select a file first', true);
    }

    const formData = new FormData();
    formData.append('doc_type', type);
    formData.append('file', fileInput.files[0]);

    if (type === 'PAN') {
        const val = document.getElementById('pan-number').value;
        if(!val) return showToast('Enter PAN number', true);
        formData.append('pan_number', val);
    } else if (type === 'AADHAAR') {
        const val = document.getElementById('aadhaar-number').value;
        if(!val) return showToast('Enter Aadhaar number', true);
        formData.append('aadhaar_number', val);
    }

    try {
        const res = await fetch(API_URL + '/kyc/upload', {
            method: 'POST',
            headers: { 'Authorization': 'Bearer ' + token },
            body: formData
        });
        const data = await res.json();
        if (!res.ok) throw new Error(data.error);
        
        showToast(type + ' uploaded successfully!');
        checkKYCStatus();
    } catch (err) {
        showToast(err.message, true);
    }
}

async function verifyKYC(action) {
    const kycId = document.getElementById('verify-kyc-id').value;
    const note = document.getElementById('rejection-note').value;

    if (!kycId) return showToast('Enter KYC ID', true);

    try {
        const res = await fetch(API_URL + '/admin/verify', {
            method: 'POST',
            headers: { 
                'Authorization': 'Bearer ' + token,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ kyc_id: kycId, action: action, rejection_note: note })
        });
        const data = await res.json();
        if (!res.ok) throw new Error(data.error);
        
        showToast(data.message);
    } catch (err) {
        showToast(err.message, true);
    }
}
