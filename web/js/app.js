// Глобальные переменные
let authToken = localStorage.getItem('token');
let currentUser = null;

// URL API
const API_URL = 'http://localhost:8080';

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', () => {
    if (window.location.pathname.includes('dashboard.html')) {
        if (!authToken) {
            window.location.href = '/';
        } else {
            loadDashboard();
        }
    } else {
        setupAuthForms();
    }
});

// Настройка форм авторизации
function setupAuthForms() {
    const loginForm = document.getElementById('login-form');
    const registerForm = document.getElementById('register-form');
    
    if (loginForm) {
        loginForm.onsubmit = async (e) => {
            e.preventDefault();
            await login();
        };
    }
    
    if (registerForm) {
        registerForm.onsubmit = async (e) => {
            e.preventDefault();
            await register();
        };
    }
}

// Регистрация
async function register() {
    const username = document.getElementById('reg-username').value;
    const email = document.getElementById('reg-email').value;
    const password = document.getElementById('reg-password').value;
    
    try {
        const response = await fetch(`${API_URL}/register`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, email, password })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            showMessage('Регистрация успешна! Теперь войдите в систему', 'success');
            document.getElementById('reg-username').value = '';
            document.getElementById('reg-email').value = '';
            document.getElementById('reg-password').value = '';
            showTab('login');
        } else {
            showMessage(data.error || 'Ошибка регистрации', 'error');
        }
    } catch (error) {
        showMessage('Ошибка соединения с сервером', 'error');
    }
}

// Вход в систему
async function login() {
    const email = document.getElementById('login-email').value;
    const password = document.getElementById('login-password').value;
    
    try {
        const response = await fetch(`${API_URL}/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            authToken = data.token;
            currentUser = { id: data.user_id, username: data.username };
            localStorage.setItem('token', authToken);
            localStorage.setItem('user', JSON.stringify(currentUser));
            window.location.href = '/dashboard.html';
        } else {
            showMessage(data.error || 'Ошибка входа', 'error');
        }
    } catch (error) {
        showMessage('Ошибка соединения с сервером', 'error');
    }
}

// Выход из системы
function logout() {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    window.location.href = '/';
}

// Загрузка дашборда
async function loadDashboard() {
    const user = JSON.parse(localStorage.getItem('user'));
    document.getElementById('user-name').innerHTML = `👤 ${user.username}`;
    
    await loadAccounts();
    await loadCards();
    await loadSelects();
}

// Загрузка счетов
async function loadAccounts() {
    try {
        const response = await fetch(`${API_URL}/api/accounts`, {
            headers: { 'Authorization': `Bearer ${authToken}` }
        });
        
        const accounts = await response.json();
        const container = document.getElementById('accounts-list');
        
        if (accounts.length === 0) {
            container.innerHTML = '<p>У вас пока нет счетов. Создайте первый счет!</p>';
        } else {
            container.innerHTML = accounts.map(acc => `
                <div class="account-card">
                    <h3>💰 Счет ${acc.account_number}</h3>
                    <div class="amount">${acc.balance.toFixed(2)} ${acc.currency}</div>
                    <small>Создан: ${new Date(acc.created_at).toLocaleDateString()}</small>
                </div>
            `).join('');
        }
    } catch (error) {
        console.error('Ошибка загрузки счетов:', error);
    }
}

// Создание счета
async function createAccount() {
    try {
        const response = await fetch(`${API_URL}/api/accounts`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authToken}`
            },
            body: JSON.stringify({ currency: 'RUB' })
        });
        
        if (response.ok) {
            alert('Счет успешно создан!');
            await loadAccounts();
            await loadSelects();
        } else {
            const error = await response.json();
            alert('Ошибка: ' + (error.error || 'Не удалось создать счет'));
        }
    } catch (error) {
        alert('Ошибка соединения с сервером');
    }
}

// Пополнение счета
async function deposit() {
    const accountId = document.getElementById('deposit-account').value;
    const amount = parseFloat(document.getElementById('deposit-amount').value);
    
    if (!accountId || !amount || amount <= 0) {
        alert('Выберите счет и введите корректную сумму');
        return;
    }
    
    try {
        const response = await fetch(`${API_URL}/api/accounts/${accountId}/deposit`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authToken}`
            },
            body: JSON.stringify({ amount })
        });
        
        if (response.ok) {
            alert('Счет успешно пополнен!');
            document.getElementById('deposit-amount').value = '';
            await loadAccounts();
        } else {
            const error = await response.json();
            alert('Ошибка: ' + (error.error || 'Не удалось пополнить счет'));
        }
    } catch (error) {
        alert('Ошибка соединения с сервером');
    }
}

// Загрузка карт
async function loadCards() {
    // Загружаем счета для выбора в модальном окне
    const accountsResponse = await fetch(`${API_URL}/api/accounts`, {
        headers: { 'Authorization': `Bearer ${authToken}` }
    });
    const accounts = await accountsResponse.json();
    
    const cardAccountSelect = document.getElementById('card-account');
    if (cardAccountSelect) {
        cardAccountSelect.innerHTML = accounts.map(acc => 
            `<option value="${acc.id}">${acc.account_number} (${acc.balance.toFixed(2)} RUB)</option>`
        ).join('');
    }
}

// Создание карты
async function createCard() {
    const accountId = document.getElementById('card-account').value;
    
    try {
        const response = await fetch(`${API_URL}/api/cards`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authToken}`
            },
            body: JSON.stringify({ account_id: accountId })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            alert(`Карта успешно выпущена!\nНомер: ${data.card_number}\nCVV: ${data.cvv || '***'}\nСохраните данные карты!`);
            closeModal();
            await loadCardsList();
        } else {
            alert('Ошибка: ' + (data.error || 'Не удалось выпустить карту'));
        }
    } catch (error) {
        alert('Ошибка соединения с сервером');
    }
}

// Загрузка списка карт
async function loadCardsList() {
    const accounts = await fetch(`${API_URL}/api/accounts`, {
        headers: { 'Authorization': `Bearer ${authToken}` }
    }).then(r => r.json());
    
    let allCards = [];
    for (const account of accounts) {
        const cards = await fetch(`${API_URL}/api/cards/${account.id}`, {
            headers: { 'Authorization': `Bearer ${authToken}` }
        }).then(r => r.json()).catch(() => []);
        allCards.push(...cards);
    }
    
    const container = document.getElementById('cards-list');
    if (container) {
        if (allCards.length === 0) {
            container.innerHTML = '<p>У вас пока нет карт. Выпустите первую карту!</p>';
        } else {
            container.innerHTML = allCards.map(card => `
                <div class="card-item">
                    <h3>💳 ${card.card_number}</h3>
                    <p>Срок: ${card.expiry}</p>
                    <p>Статус: ${card.is_active ? 'Активна' : 'Заблокирована'}</p>
                </div>
            `).join('');
        }
    }
}

// Перевод средств
async function transfer() {
    const fromAccountId = document.getElementById('transfer-from').value;
    const toAccountNumber = document.getElementById('transfer-to').value;
    const amount = parseFloat(document.getElementById('transfer-amount').value);
    const description = document.getElementById('transfer-desc').value;
    
    if (!fromAccountId || !toAccountNumber || !amount || amount <= 0) {
        alert('Заполните все поля корректно');
        return;
    }
    
    try {
        const response = await fetch(`${API_URL}/api/transfer`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authToken}`
            },
            body: JSON.stringify({ 
                from_account_id: fromAccountId,
                to_account_number: toAccountNumber, 
                amount, 
                description 
            })
        });
        
        if (response.ok) {
            alert('Перевод успешно выполнен!');
            document.getElementById('transfer-to').value = '';
            document.getElementById('transfer-amount').value = '';
            document.getElementById('transfer-desc').value = '';
            await loadAccounts();
        } else {
            const error = await response.json();
            alert('Ошибка: ' + (error.error || 'Не удалось выполнить перевод'));
        }
    } catch (error) {
        alert('Ошибка соединения с сервером');
    }
}

// Оформление кредита
async function applyCredit() {
    const accountId = document.getElementById('credit-account').value;
    const amount = parseFloat(document.getElementById('credit-amount').value);
    const termMonths = parseInt(document.getElementById('credit-term').value);
    
    if (!accountId || !amount || amount < 10000 || amount > 1000000) {
        alert('Сумма кредита должна быть от 10 000 до 1 000 000 RUB');
        return;
    }
    
    try {
        const response = await fetch(`${API_URL}/api/credits`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authToken}`
            },
            body: JSON.stringify({ account_id: accountId, amount, term_months: termMonths })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            alert(`Кредит одобрен!\nСумма: ${data.amount} RUB\nЕжемесячный платеж: ${data.monthly_payment.toFixed(2)} RUB`);
            await loadAccounts();
            await loadCreditInfo(data.id);
        } else {
            alert('Ошибка: ' + (data.error || 'Не удалось оформить кредит'));
        }
    } catch (error) {
        alert('Ошибка соединения с сервером');
    }
}

// Загрузка аналитики
async function loadAnalytics() {
    try {
        const response = await fetch(`${API_URL}/api/analytics`, {
            headers: { 'Authorization': `Bearer ${authToken}` }
        });
        
        const data = await response.json();
        
        const container = document.getElementById('analytics-stats');
        container.innerHTML = `
            <div class="stat-card">
                <h3>💰 Доходы</h3>
                <div class="stat-value">${data.total_income.toFixed(2)} RUB</div>
            </div>
            <div class="stat-card">
                <h3>📉 Расходы</h3>
                <div class="stat-value">${data.total_expense.toFixed(2)} RUB</div>
            </div>
            <div class="stat-card">
                <h3>💵 Баланс</h3>
                <div class="stat-value">${data.balance.toFixed(2)} RUB</div>
            </div>
            <div class="stat-card">
                <h3>🏦 Ключевая ставка</h3>
                <div class="stat-value">${data.key_rate}%</div>
            </div>
        `;
    } catch (error) {
        console.error('Ошибка загрузки аналитики:', error);
    }
}

// Прогноз баланса
async function predictBalance() {
    const accountId = document.getElementById('predict-account').value;
    const days = parseInt(document.getElementById('predict-days').value);
    
    if (!accountId || days < 1 || days > 365) {
        alert('Введите корректное количество дней (1-365)');
        return;
    }
    
    try {
        const response = await fetch(`${API_URL}/api/accounts/${accountId}/predict?days=${days}`, {
            headers: { 'Authorization': `Bearer ${authToken}` }
        });
        
        const predictions = await response.json();
        
        const container = document.getElementById('predict-result');
        container.innerHTML = `
            <h3>📈 Прогноз баланса на ${days} дней</h3>
            <div class="predict-table">
                ${predictions.map(p => `
                    <div class="predict-row">
                        <span>День ${p.day} (${p.date})</span>
                        <span class="${p.balance >= 0 ? 'positive' : 'negative'}">${p.balance.toFixed(2)} RUB</span>
                    </div>
                `).join('')}
            </div>
        `;
    } catch (error) {
        alert('Ошибка получения прогноза');
    }
}

// Загрузка списков для select
async function loadSelects() {
    const accounts = await fetch(`${API_URL}/api/accounts`, {
        headers: { 'Authorization': `Bearer ${authToken}` }
    }).then(r => r.json());
    
    const selects = ['deposit-account', 'transfer-from', 'credit-account', 'predict-account'];
    selects.forEach(id => {
        const select = document.getElementById(id);
        if (select) {
            select.innerHTML = accounts.map(acc => 
                `<option value="${acc.id}">${acc.account_number} (${acc.balance.toFixed(2)} RUB)</option>`
            ).join('');
        }
    });
}

// Загрузка информации о кредите
async function loadCreditInfo(creditId) {
    try {
        const response = await fetch(`${API_URL}/api/credits/${creditId}/schedule`, {
            headers: { 'Authorization': `Bearer ${authToken}` }
        });
        
        const schedules = await response.json();
        const container = document.getElementById('credit-info');
        
        container.innerHTML = `
            <h3>📅 График платежей</h3>
            <div class="schedule-list">
                ${schedules.map(s => `
                    <div class="schedule-item">
                        <span>Платеж №${s.payment_number}</span>
                        <span>Срок: ${new Date(s.due_date).toLocaleDateString()}</span>
                        <span>Сумма: ${s.amount.toFixed(2)} RUB</span>
                        <span class="status-${s.status}">${s.status === 'pending' ? 'Ожидает' : 'Оплачен'}</span>
                    </div>
                `).join('')}
            </div>
        `;
    } catch (error) {
        console.error('Ошибка загрузки графика платежей:', error);
    }
}

// Вспомогательные функции
function showMessage(msg, type) {
    const msgDiv = document.getElementById('message');
    msgDiv.textContent = msg;
    msgDiv.className = `message ${type}`;
    setTimeout(() => {
        msgDiv.className = 'message';
        msgDiv.textContent = '';
    }, 3000);
}

function showDashboardTab(tab) {
    document.querySelectorAll('.dash-content').forEach(t => t.classList.remove('active'));
    document.querySelectorAll('.dash-btn').forEach(b => b.classList.remove('active'));
    
    document.getElementById(`${tab}-tab`).classList.add('active');
    event.target.classList.add('active');
    
    if (tab === 'cards') loadCardsList();
    if (tab === 'analytics') loadAnalytics();
}

function showCreateCardModal() {
    document.getElementById('card-modal').style.display = 'block';
}

function closeModal() {
    document.getElementById('card-modal').style.display = 'none';
}