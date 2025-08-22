// Thewavess AI Core - 共用 JavaScript 模組

// 全局狀態管理
const AppState = {
    // 系統狀態
    systemOnline: false,
    isAuthenticated: false,
    currentUser: null,
    accessToken: null,
    refreshToken: null,
    
    // 用戶流程狀態
    ageVerified: false,
    selectedCharacter: null,
    currentSession: null,
    
    // 對話狀態
    messages: [],
    emotions: [],
    
    // 監控數據
    apiLogs: [],
    
    // 用戶偏好
    preferences: {
        theme: 'dark',
        auto_scroll: true
    }
};

// API 基礎配置
const API_CONFIG = {
    baseURL: '',
    timeout: 30000,
    headers: {
        'Content-Type': 'application/json'
    }
};

// 工具函數
const Utils = {
    // 時間格式化
    formatTime(date) {
        return moment(date).format('HH:mm:ss');
    },
    
    // 日期時間格式化
    formatDateTime(date) {
        return moment(date).format('YYYY-MM-DD HH:mm:ss');
    },
    
    // 相對時間格式化（例如：2分鐘前）
    formatRelativeTime(date) {
        return moment(date).fromNow();
    },
    
    // 友好的日期時間格式
    formatFriendlyDateTime(date) {
        const now = moment();
        const target = moment(date);
        
        if (target.isSame(now, 'day')) {
            return '今天 ' + target.format('HH:mm');
        } else if (target.isSame(now.clone().subtract(1, 'day'), 'day')) {
            return '昨天 ' + target.format('HH:mm');
        } else if (target.isAfter(now.clone().subtract(7, 'days'))) {
            return target.format('dddd HH:mm');
        } else if (target.isSame(now, 'year')) {
            return target.format('MM-DD HH:mm');
        } else {
            return target.format('YYYY-MM-DD HH:mm');
        }
    },
    
    // 日期格式化
    formatDate(date) {
        return moment(date).format('YYYY-MM-DD');
    },
    
    // 時間格式化（短）
    formatShortTime(date) {
        return moment(date).format('HH:mm');
    },
    
    // 生成隨機 ID
    generateId() {
        return Math.random().toString(36).substr(2, 9);
    },
    
    // 延遲函數
    sleep(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    },
    
    // 本地存儲操作
    storage: {
        set(key, value) {
            try {
                localStorage.setItem(key, JSON.stringify(value));
            } catch (e) {
                console.warn('無法保存到本地存儲:', e);
            }
        },
        
        get(key, defaultValue = null) {
            try {
                const value = localStorage.getItem(key);
                return value ? JSON.parse(value) : defaultValue;
            } catch (e) {
                console.warn('無法從本地存儲讀取:', e);
                return defaultValue;
            }
        },
        
        remove(key) {
            try {
                localStorage.removeItem(key);
            } catch (e) {
                console.warn('無法從本地存儲刪除:', e);
            }
        }
    },
    
    // 表單驗證
    validation: {
        email(email) {
            const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
            return re.test(email);
        },
        
        password(password) {
            // 至少8位，包含大小寫字母和數字
            const re = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d@$!%*?&]{8,}$/;
            return re.test(password);
        },
        
        age(birthDate) {
            const today = new Date();
            const birth = new Date(birthDate);
            const age = today.getFullYear() - birth.getFullYear();
            const monthDiff = today.getMonth() - birth.getMonth();
            
            if (monthDiff < 0 || (monthDiff === 0 && today.getDate() < birth.getDate())) {
                return age - 1;
            }
            return age;
        }
    }
};

// API 請求封裝
const ApiClient = {
    // 通用請求方法
    async request(method, endpoint, data = null, options = {}) {
        const url = API_CONFIG.baseURL + endpoint;
        const config = {
            method,
            headers: { ...API_CONFIG.headers, ...options.headers },
            ...options
        };
        
        // 添加認證頭
        if (AppState.accessToken) {
            config.headers.Authorization = `Bearer ${AppState.accessToken}`;
        }
        
        // 添加請求體
        if (data) {
            config.body = JSON.stringify(data);
        }
        
        try {
            const response = await fetch(url, config);
            const result = await response.json();
            
            // 記錄 API 調用
            this.logApiCall(method, endpoint, response.status, result);
            
            // 處理認證失敗
            const inAuthFlow = endpoint.includes('/auth/login') || endpoint.includes('/auth/register');
            if (response.status === 401 && !endpoint.includes('/auth/refresh') && !inAuthFlow) {
                const refreshed = await this.handleUnauthorized();
                if (refreshed) {
                    // 重新嘗試原始請求
                    return this.request(method, endpoint, data, options);
                } else {
                    throw new Error('認證失敗，請重新登入');
                }
            }
            
            if (!response.ok) {
                throw new Error(result.error?.message || `HTTP ${response.status}`);
            }
            
            return result;
        } catch (error) {
            this.logApiCall(method, endpoint, 0, { error: error.message });
            throw error;
        }
    },
    
    // HTTP 方法快捷方式
    get(endpoint, options = {}) {
        return this.request('GET', endpoint, null, options);
    },
    
    post(endpoint, data, options = {}) {
        return this.request('POST', endpoint, data, options);
    },
    
    put(endpoint, data, options = {}) {
        return this.request('PUT', endpoint, data, options);
    },
    
    delete(endpoint, options = {}) {
        return this.request('DELETE', endpoint, null, options);
    },
    
    // API 調用記錄
    logApiCall(method, endpoint, status, data) {
        const log = {
            id: Utils.generateId(),
            timestamp: new Date().toISOString(),
            method,
            endpoint,
            status,
            data,
            isSuccess: status >= 200 && status < 300
        };
        
        AppState.apiLogs.unshift(log);
        
        // 限制日誌數量
        if (AppState.apiLogs.length > 100) {
            AppState.apiLogs = AppState.apiLogs.slice(0, 100);
        }
        
        // 觸發日誌更新事件
        this.dispatchEvent('apiLogUpdated', log);
    },
    
    // 處理認證失敗
    async handleUnauthorized() {
        if (AppState.refreshToken) {
            try {
                // 直接使用fetch，避免再次觸發401處理
                const response = await fetch(API_CONFIG.baseURL + '/api/v1/auth/refresh', {
                    method: 'POST',
                    headers: API_CONFIG.headers,
                    body: JSON.stringify({
                        refresh_token: AppState.refreshToken
                    })
                });
                
                const result = await response.json();
                
                if (response.ok && result.success) {
                    AppState.accessToken = result.data.token;
                    AppState.refreshToken = result.data.refresh_token;
                    Utils.storage.set('tokens', {
                        access: AppState.accessToken,
                        refresh: AppState.refreshToken
                    });
                    return true; // 成功刷新
                }
            } catch (error) {
                console.warn('Token 刷新失敗:', error);
            }
        }
        
        // 清除認證狀態
        this.logout();
        return false; // 刷新失敗
    },
    
    // 登出
    logout() {
        AppState.isAuthenticated = false;
        AppState.currentUser = null;
        AppState.accessToken = null;
        AppState.refreshToken = null;
        AppState.selectedCharacter = null;
        AppState.currentSession = null;
        
        Utils.storage.remove('tokens');
        Utils.storage.remove('user');
        
        this.dispatchEvent('userLoggedOut');
    },
    
    // 事件系統
    events: {},
    
    addEventListener(event, callback) {
        if (!this.events[event]) {
            this.events[event] = [];
        }
        this.events[event].push(callback);
    },
    
    removeEventListener(event, callback) {
        if (this.events[event]) {
            this.events[event] = this.events[event].filter(cb => cb !== callback);
        }
    },
    
    dispatchEvent(event, data = null) {
        if (this.events[event]) {
            this.events[event].forEach(callback => {
                try {
                    callback(data);
                } catch (error) {
                    console.error('事件處理錯誤:', error);
                }
            });
        }
    }
};

// UI 工具函數
const UI = {
    // Flowbite Toast 通知
    showToast(message, type = 'info', duration = 3000) {
        // 確保 toast 容器存在
        let container = document.getElementById('toast-container');
        if (!container) {
            container = document.createElement('div');
            container.id = 'toast-container';
            container.className = 'fixed top-5 right-5 z-50 space-y-2 md:top-5 md:right-5 sm:top-4 sm:right-4 sm:left-4 sm:right-auto sm:max-w-sm';
            document.body.appendChild(container);
        }
        
        const toastId = 'toast-' + Date.now();
        
        const typeConfig = {
            success: {
                bgColor: 'bg-green-100',
                textColor: 'text-green-500',
                darkBg: 'dark:bg-green-800',
                darkText: 'dark:text-green-200',
                icon: '<svg class="w-5 h-5" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 20 20"><path d="M10 .5a9.5 9.5 0 1 0 9.5 9.5A9.51 9.51 0 0 0 10 .5Zm3.707 8.207-4 4a1 1 0 0 1-1.414 0l-2-2a1 1 0 0 1 1.414-1.414L9 10.586l3.293-3.293a1 1 0 0 1 1.414 1.414Z"/></svg>'
            },
            error: {
                bgColor: 'bg-red-100',
                textColor: 'text-red-500',
                darkBg: 'dark:bg-red-800',
                darkText: 'dark:text-red-200',
                icon: '<svg class="w-5 h-5" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 20 20"><path d="M10 .5a9.5 9.5 0 1 0 9.5 9.5A9.51 9.51 0 0 0 10 .5Zm3.707 11.793a1 1 0 1 1-1.414 1.414L10 11.414l-2.293 2.293a1 1 0 0 1-1.414-1.414L8.586 10 6.293 7.707a1 1 0 0 1 1.414-1.414L10 8.586l2.293-2.293a1 1 0 0 1 1.414 1.414L11.414 10l2.293 2.293Z"/></svg>'
            },
            warning: {
                bgColor: 'bg-orange-100',
                textColor: 'text-orange-500',
                darkBg: 'dark:bg-orange-700',
                darkText: 'dark:text-orange-200',
                icon: '<svg class="w-5 h-5" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 20 20"><path d="M10 .5a9.5 9.5 0 1 0 9.5 9.5A9.51 9.51 0 0 0 10 .5ZM10 15a1 1 0 1 1 0-2 1 1 0 0 1 0 2Zm1-4a1 1 0 0 1-2 0V6a1 1 0 0 1 2 0v5Z"/></svg>'
            },
            info: {
                bgColor: 'bg-blue-100',
                textColor: 'text-blue-500',
                darkBg: 'dark:bg-blue-800',
                darkText: 'dark:text-blue-200',
                icon: '<svg class="w-5 h-5" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 20 20"><path d="M10 .5a9.5 9.5 0 1 0 9.5 9.5A9.51 9.51 0 0 0 10 .5ZM9.5 4a1.5 1.5 0 1 1 0 3 1.5 1.5 0 0 1 0-3ZM12 15H8a1 1 0 0 1 0-2h1v-3H8a1 1 0 0 1 0-2h2a1 1 0 0 1 1 1v4h1a1 1 0 0 1 0 2Z"/></svg>'
            }
        };
        
        const config = typeConfig[type] || typeConfig.info;
        
        const toastHTML = `
            <div id="${toastId}" class="flex items-center w-full max-w-xs p-4 mb-4 text-gray-500 bg-white rounded-lg shadow ${config.darkBg} ${config.darkText}" role="alert">
                <div class="inline-flex items-center justify-center flex-shrink-0 w-8 h-8 ${config.textColor} ${config.bgColor} rounded-lg ${config.darkBg} ${config.darkText}">
                    ${config.icon}
                    <span class="sr-only">${type} icon</span>
                </div>
                <div class="ms-3 text-sm font-normal">${message}</div>
                <button type="button" class="ms-auto -mx-1.5 -my-1.5 bg-white text-gray-400 hover:text-gray-900 rounded-lg focus:ring-2 focus:ring-gray-300 p-1.5 hover:bg-gray-100 inline-flex items-center justify-center h-8 w-8 dark:text-gray-500 dark:hover:text-white dark:bg-gray-800 dark:hover:bg-gray-700" data-dismiss-target="#${toastId}" aria-label="Close">
                    <span class="sr-only">Close</span>
                    <svg class="w-3 h-3" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 14 14">
                        <path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m1 1 6 6m0 0 6 6M7 7l6-6M7 7l-6 6"/>
                    </svg>
                </button>
            </div>
        `;
        
        container.insertAdjacentHTML('beforeend', toastHTML);
        
        // 自動關閉
        setTimeout(() => {
            const toastElement = document.getElementById(toastId);
            if (toastElement) {
                toastElement.remove();
            }
        }, duration);
    },
    
    // 顯示載入狀態
    showLoading(element, show = true) {
        if (!element) return;
        
        if (show) {
            element.disabled = true;
            // 保存原始內容
            if (!element.dataset.originalContent) {
                element.dataset.originalContent = element.innerHTML;
            }
            element.innerHTML = '<svg class="animate-spin h-5 w-5 mr-2 inline" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/></svg>載入中...';
        } else {
            element.disabled = false;
            // 恢復原始內容
            if (element.dataset.originalContent) {
                element.innerHTML = element.dataset.originalContent;
                delete element.dataset.originalContent;
            }
        }
    },
    
    // 全局載入狀態（用於顯示全頁載入）
    showGlobalLoading(show = true) {
        if (show) {
            // 顯示全局載入指示器
            const loading = document.createElement('div');
            loading.id = 'globalLoading';
            loading.className = 'fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50';
            loading.innerHTML = `
                <div class="bg-white rounded-lg p-6 flex items-center space-x-3">
                    <svg class="animate-spin h-6 w-6 text-blue-500" viewBox="0 0 24 24">
                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none"/>
                        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
                    </svg>
                    <span class="text-gray-900">載入中...</span>
                </div>
            `;
            document.body.appendChild(loading);
        } else {
            // 隱藏全局載入指示器
            const loading = document.getElementById('globalLoading');
            if (loading) {
                loading.remove();
            }
        }
    },
    
    // 模態框控制
    modal: {
        show(modalId) {
            const modal = document.getElementById(modalId);
            if (modal) {
                modal.classList.remove('hidden');
                modal.classList.add('flex');
            }
        },
        
        hide(modalId) {
            const modal = document.getElementById(modalId);
            if (modal) {
                modal.classList.add('hidden');
                modal.classList.remove('flex');
            }
        }
    },
    
    // 動態更新元素內容
    updateElement(selector, content) {
        const element = document.querySelector(selector);
        if (element) {
            element.innerHTML = content;
        }
    },
    
    // 切換元素可見性
    toggleElement(selector, show) {
        const element = document.querySelector(selector);
        if (element) {
            if (show) {
                element.classList.remove('hidden');
            } else {
                element.classList.add('hidden');
            }
        }
    },
    
    // 通用錯誤處理
    handleError(error, context = '操作', showToast = true) {
        console.error(`${context}失敗:`, error);
        
        let message = `${context}失敗`;
        
        if (error.message) {
            if (error.message.includes('401')) {
                message = '身份驗證失敗，請重新登入';
            } else if (error.message.includes('403')) {
                message = '權限不足';
            } else if (error.message.includes('404')) {
                message = '請求的資源不存在';
            } else if (error.message.includes('500')) {
                message = '服務器內部錯誤';
            } else if (error.message.includes('Network')) {
                message = '網路連接錯誤';
            } else {
                message += ': ' + error.message;
            }
        }
        
        if (showToast) {
            this.showToast(message, 'error');
        }
        
        return message;
    },
    
    // 顯示連接狀態
    showConnectionStatus(isOnline) {
        const statusElement = document.getElementById('connectionStatus');
        if (statusElement) {
            if (isOnline) {
                statusElement.innerHTML = '<i class="fas fa-wifi text-green-400"></i> 已連接';
                statusElement.className = 'text-green-400 text-sm';
            } else {
                statusElement.innerHTML = '<i class="fas fa-wifi-slash text-red-400"></i> 離線模式';
                statusElement.className = 'text-red-400 text-sm';
            }
        }
    },
    
    // 模態框管理
    modal: {
        show(modalId) {
            const modal = document.getElementById(modalId);
            if (modal) {
                modal.classList.remove('hidden');
                document.body.style.overflow = 'hidden'; // 防止背景滾動
            }
        },
        
        hide(modalId) {
            const modal = document.getElementById(modalId);
            if (modal) {
                modal.classList.add('hidden');
                document.body.style.overflow = ''; // 恢復滾動
            }
        },
        
        toggle(modalId) {
            const modal = document.getElementById(modalId);
            if (modal) {
                if (modal.classList.contains('hidden')) {
                    this.show(modalId);
                } else {
                    this.hide(modalId);
                }
            }
        }
    }
};

// 頁面路由系統
const Router = {
    routes: {
        '/': '/public/',
        '/age-verify': '/public/age-verify.html',
        '/auth': '/public/auth.html',
        '/character': '/public/character.html',
        '/chat': '/public/chat.html',
        '/profile': '/public/profile.html',
        '/emotion': '/public/emotion.html',
        '/memory': '/public/memory.html',
        '/search': '/public/search.html',
        '/novel': '/public/novel.html',
        '/tts': '/public/tts.html',
        '/tags': '/public/tags.html',
        '/chat-test': '/public/chat-test.html',
        '/admin': '/public/admin.html'
    },
    
    navigate(path) {
        if (this.routes[path]) {
            const targetPath = this.routes[path];
            const currentPath = window.location.pathname;
            
            // 防止無限跳轉 - 如果已經在目標路徑，不要再跳轉
            if (currentPath !== targetPath) {
                window.location.href = targetPath;
            }
        }
    },
    
    getCurrentPage() {
        const path = window.location.pathname;
        const page = path.split('/').pop() || 'index.html';
        return page.replace('.html', '');
    }
};

// 初始化應用
const App = {
    async init() {
        try {
            // 載入保存的狀態
            this.loadSavedState();
            
            // 檢查系統狀態
            await this.checkSystemHealth();
            
            // 設置事件監聽器
            this.setupEventListeners();
            
            console.log('應用初始化完成');
        } catch (error) {
            console.error('應用初始化失敗:', error);
        }
    },
    
    loadSavedState() {
        // 載入保存的 tokens
        const tokens = Utils.storage.get('tokens');
        if (tokens) {
            AppState.accessToken = tokens.access;
            AppState.refreshToken = tokens.refresh;
            AppState.isAuthenticated = true;
        }
        
        // 載入用戶資料
        const user = Utils.storage.get('user');
        if (user) {
            AppState.currentUser = user;
        }
        
        // 載入用戶偏好
        const preferences = Utils.storage.get('preferences');
        if (preferences) {
            AppState.preferences = { ...AppState.preferences, ...preferences };
        }
    },
    
    async checkSystemHealth() {
        try {
            // Health endpoint is not under /api/v1, use direct fetch
            const response = await fetch('/health');
            const result = await response.json();
            AppState.systemOnline = result.status === 'ok';
        } catch (error) {
            AppState.systemOnline = false;
            console.warn('系統健康檢查失敗:', error);
        }
    },
    
    setupEventListeners() {
        // 監聽 API 日誌更新
        ApiClient.addEventListener('apiLogUpdated', (log) => {
            this.onApiLogUpdated(log);
        });
        
        // 監聽用戶登出
        ApiClient.addEventListener('userLoggedOut', () => {
            this.onUserLoggedOut();
        });
        
        // 監聽網路狀態變化
        window.addEventListener('online', () => {
            AppState.systemOnline = true;
            UI.showConnectionStatus(true);
            UI.showToast('網路連接已恢復', 'success');
        });
        
        window.addEventListener('offline', () => {
            AppState.systemOnline = false;
            UI.showConnectionStatus(false);
            UI.showToast('網路連接中斷，切換到離線模式', 'warning');
        });
        
        // 監聽頁面卸載，保存狀態
        window.addEventListener('beforeunload', () => {
            this.saveState();
        });
    },
    
    onApiLogUpdated(log) {
        // 可以在這裡更新 UI 或觸發其他操作
        console.log('API 調用記錄:', log);
    },
    
    onUserLoggedOut() {
        // 用戶登出後的處理
        // 只在需要認證的頁面才跳轉
        const currentPage = Router.getCurrentPage();
        if (currentPage !== 'index' && currentPage !== 'age-verify') {
            Router.navigate('/');
        }
    },
    
    saveState() {
        // 保存用戶偏好
        Utils.storage.set('preferences', AppState.preferences);
    }
};

// 字符系統相關 - 移除硬編碼角色，完全依賴 API
const CharacterSystem = {
    // 角色數據應該完全從 API 獲取，不再提供離線數據
};

// NSFW 內容系統
const NSFWSystem = {
    levels: {
        1: { name: '日常對話', color: 'bg-green-500', description: '日常對話內容，安全適宜', engine: 'openai' },
        2: { name: '浪漫內容', color: 'bg-yellow-500', description: '輕度浪漫內容，情感表達', engine: 'openai' },
        3: { name: '親密內容', color: 'bg-orange-500', description: '親密互動，深度情感', engine: 'openai' },
        4: { name: '成人內容', color: 'bg-red-500', description: '成人向內容，身體親密', engine: 'grok' },
        5: { name: '明確內容', color: 'bg-purple-500', description: '明確成人內容，完全開放', engine: 'grok' }
    },
    
    getLevelInfo(level) {
        return this.levels[level] || this.levels[1];
    },
    
    getExpectedEngine(level) {
        const info = this.getLevelInfo(level);
        return info.engine;
    },
    
    checkAge(birthDate) {
        const age = Utils.validation.age(birthDate);
        return age >= 18;
    },
    
    // 檢查引擎是否符合預期
    validateEngine(level, actualEngine) {
        const expectedEngine = this.getExpectedEngine(level);
        return expectedEngine === actualEngine;
    }
};

// API 路徑常量
const API_PATHS = {
    // 基礎路徑
    BASE: '/api/v1',
    HEALTH: '/health',
    
    // 認證相關
    AUTH: {
        LOGIN: '/api/v1/auth/login',
        REGISTER: '/api/v1/auth/register', 
        LOGOUT: '/api/v1/auth/logout',
        REFRESH: '/api/v1/auth/refresh'
    },
    
    // 用戶相關
    USER: {
        PROFILE: '/api/v1/user/profile',
        PREFERENCES: '/api/v1/user/preferences',
        AVATAR: '/api/v1/user/avatar',
        VERIFY: '/api/v1/user/verify'
    },
    
    // 角色相關
    CHARACTER: {
        LIST: '/api/v1/character/list',
        DETAIL: (id) => `/api/v1/character/${id}`,
        STATS: (id) => `/api/v1/character/${id}/stats`,
        SEARCH: '/api/v1/character/search',
        CREATE: '/api/v1/character',
        UPDATE: (id) => `/api/v1/character/${id}`,
        DELETE: (id) => `/api/v1/character/${id}`,
        NSFW_GUIDELINE: (level) => `/api/v1/character/nsfw-guideline/${level}`
    },
    
    // TTS相關
    TTS: {
        GENERATE: '/api/v1/tts/generate',
        VOICES: '/api/v1/tts/voices'
    },
    
    // 標籤相關
    TAGS: {
        ALL: '/api/v1/tags',
        POPULAR: '/api/v1/tags/popular'
    },
    
    // 監控相關
    MONITOR: {
        HEALTH: '/api/v1/monitor/health',
        STATS: '/api/v1/monitor/stats'
    },
    
    // 系統相關
    SYSTEM: {
        VERSION: '/api/v1/version',
        STATUS: '/api/v1/status'
    }
};

// API 響應標準化工具
const ResponseNormalizer = {
    // 標準化聊天消息響應
    normalizeMessageResponse(apiResponse) {
        if (!apiResponse) return null;
        
        // 處理兩種可能的響應格式：直接數據或包裝在data中的數據
        const data = apiResponse.data || apiResponse;
        return {
            id: data.id,
            session_id: data.session_id,
            role: data.role || 'assistant',
            // 使用標準的 content 字段
            content: data.content,
            character_action: data.character_action || '',
            scene_description: data.scene_description || '',
            emotion_state: data.emotion_state || {},
            ai_engine: data.ai_engine || 'unknown',
            nsfw_level: data.nsfw_level || 1,
            response_time: data.response_time_ms || 0,
            special_event: data.special_event || null,
            created_at: data.created_at || new Date().toISOString()
        };
    },
    
    // 標準化角色響應
    normalizeCharacterResponse(apiCharacter) {
        if (!apiCharacter) return null;
        
        return {
            id: apiCharacter.id,
            name: apiCharacter.name,
            avatar_url: apiCharacter.avatar_url,
            type: apiCharacter.type,
            description: apiCharacter.description || `這是 ${apiCharacter.name}`,
            tags: apiCharacter.tags || [],
            popularity: apiCharacter.popularity || 0,
            metadata: apiCharacter.metadata || {}
        };
    },
    
    // 標準化情感狀態響應
    normalizeEmotionResponse(emotionState) {
        if (!emotionState) return {
            affection: 0,
            mood: 'neutral',
            relationship: 'stranger',
            intimacy_level: 'distant'
        };
        
        return {
            affection: emotionState.affection || 0,
            mood: emotionState.mood || 'neutral',
            relationship: emotionState.relationship || 'stranger',
            intimacy_level: emotionState.intimacy_level || 'distant',
            // 可能的額外字段
            trust_level: emotionState.trust_level || 0,
            interaction_count: emotionState.interaction_count || 0
        };
    }
};

// 專用 API 客戶端
const SpecializedApiClients = {
    // TTS API 客戶端
    TTS: {
        async generateSpeech(text, voice, speed = 1.0) {
            return ApiClient.post('/api/v1/tts/generate', {
                text: text,
                voice: voice,
                speed: speed
            });
        },
        
        async getVoices(characterId = null, language = null) {
            let endpoint = '/api/v1/tts/voices';
            const params = new URLSearchParams();
            
            if (characterId) params.append('character_id', characterId);
            if (language) params.append('language', language);
            
            if (params.toString()) {
                endpoint += '?' + params.toString();
            }
            
            return ApiClient.get(endpoint);
        }
    },
    
    // 角色 API 客戶端
    Character: {
        async getList() {
            return ApiClient.get('/api/v1/character/list');
        },
        
        // 別名方法，與 getList 相同
        async getAll() {
            return this.getList();
        },
        
        async getById(id) {
            return ApiClient.get(`/api/v1/character/${id}`);
        },
        
        async getStats(id) {
            return ApiClient.get(`/api/v1/character/${id}/stats`);
        },
        
        async search(query) {
            return ApiClient.get(`/api/v1/character/search?q=${encodeURIComponent(query)}`);
        },
        
        async create(characterData) {
            return ApiClient.post('/api/v1/character', characterData);
        },
        
        async update(id, characterData) {
            return ApiClient.put(`/api/v1/character/${id}`, characterData);
        },
        
        async delete(id) {
            return ApiClient.delete(`/api/v1/character/${id}`);
        },
        
        async getNSFWGuideline(level, locale = null, engine = null) {
            let endpoint = `/api/v1/character/nsfw-guideline/${level}`;
            const params = new URLSearchParams();
            
            if (locale) params.append('locale', locale);
            if (engine) params.append('engine', engine);
            
            if (params.toString()) {
                endpoint += '?' + params.toString();
            }
            
            return ApiClient.get(endpoint);
        }
    },
    
    // 聊天 API 客戶端
    Chat: {
        async createSession(characterId, title = null) {
            return ApiClient.post('/api/v1/chat/session', {
                character_id: characterId,
                title: title
            });
        },
        
        async getSessions(characterId = null, page = 1, limit = 20, status = null) {
            let endpoint = `/api/v1/chat/sessions?page=${page}&limit=${limit}`;
            if (characterId) endpoint += `&character_id=${characterId}`;
            if (status) endpoint += `&status=${status}`;
            return ApiClient.get(endpoint);
        },
        
        async getSession(sessionId) {
            return ApiClient.get(`/api/v1/chat/session/${sessionId}`);
        },
        
        async sendMessage(sessionId, content, nsfwLevel = 1) {
            return ApiClient.post('/api/v1/chat/message', {
                session_id: sessionId,
                message: content
            });
        },
        
        async getHistory(sessionId, page = 1, limit = 50) {
            return ApiClient.get(`/api/v1/chat/session/${sessionId}/history?page=${page}&limit=${limit}`);
        },
        
        async exportSession(sessionId, format = 'json') {
            return ApiClient.get(`/api/v1/chat/session/${sessionId}/export?format=${format}`);
        },
        
        async regenerateResponse(messageId) {
            return ApiClient.post('/api/v1/chat/regenerate', {
                message_id: messageId
            });
        },
        
        async deleteSession(sessionId) {
            return ApiClient.delete(`/api/v1/chat/session/${sessionId}`);
        }
    },
    
    // 記憶 API 客戶端
    Memory: {
        async getTimeline(characterId = 'char_001', page = 1, limit = 20) {
            return ApiClient.get(`/api/v1/memory/timeline?character_id=${characterId}&page=${page}&limit=${limit}`);
        },
        
        async save(sessionId, content, type, importance = 5, characterId) {
            return ApiClient.post('/api/v1/memory/save', {
                session_id: sessionId,
                content: content,
                type: type,
                importance: importance,
                character_id: characterId
            });
        },
        
        async search(query, type = null, characterId = 'char_001') {
            let endpoint = `/api/v1/memory/search?query=${encodeURIComponent(query)}&character_id=${characterId}`;
            if (type) endpoint += `&type=${type}`;
            return ApiClient.get(endpoint);
        },
        
        async getStats(characterId = 'char_001') {
            return ApiClient.get(`/api/v1/memory/stats?character_id=${characterId}`);
        },
        
        async getUserMemory(userId) {
            return ApiClient.get(`/api/v1/memory/user/${userId}`);
        },
        
        async forget(memoryId, forgetType = 'delete', reason = '') {
            return ApiClient.delete('/api/v1/memory/forget', {
                memory_id: memoryId,
                forget_type: forgetType,
                reason: reason
            });
        },
        
        async backup(backupType = 'full', includeTypes = [], compression = false, encryption = false) {
            return ApiClient.post('/api/v1/memory/backup', {
                backup_type: backupType,
                include_types: includeTypes,
                compression: compression,
                encryption: encryption
            });
        },
        
        async restore(backupId, restoreType = 'full', mergeStrategy = 'replace', verifyIntegrity = true) {
            return ApiClient.post('/api/v1/memory/restore', {
                backup_id: backupId,
                restore_type: restoreType,
                merge_strategy: mergeStrategy,
                verify_integrity: verifyIntegrity
            });
        }
    },
    
    
    // 用戶 API 客戶端
    User: {
        async getProfile() {
            return ApiClient.get('/api/v1/user/profile');
        },
        
        async updateProfile(profileData) {
            return ApiClient.put('/api/v1/user/profile', profileData);
        },
        
        async getPreferences() {
            return ApiClient.get('/api/v1/user/preferences');
        },
        
        async updatePreferences(preferences) {
            return ApiClient.put('/api/v1/user/preferences', preferences);
        },
        
        async uploadAvatar(avatarUrl) {
            return ApiClient.post('/api/v1/user/avatar', {
                avatar_url: avatarUrl
            });
        },
        
        async verifyAge(birthDate, documentType = null) {
            return ApiClient.post('/api/v1/user/verify', {
                birth_date: birthDate,
                document_type: documentType
            });
        }
    },
    
    // 認證 API 客戶端
    Auth: {
        async login(username, password) {
            return ApiClient.post('/api/v1/auth/login', {
                username: username,
                password: password
            });
        },
        
        async register(userData) {
            return ApiClient.post('/api/v1/auth/register', userData);
        },
        
        async refresh() {
            return ApiClient.post('/api/v1/auth/refresh', {
                refresh_token: AppState.refreshToken
            });
        },
        
        async logout() {
            return ApiClient.post('/api/v1/auth/logout');
        }
    },
    
    // 管理 API 客戶端
    Admin: {
        async getStats() {
            return ApiClient.get('/api/v1/admin/stats');
        },
        
        async getUsers(page = 1, limit = 20) {
            return ApiClient.get(`/api/v1/admin/users?page=${page}&limit=${limit}`);
        },
        
        async updateUser(userId, userData) {
            return ApiClient.put(`/api/v1/admin/users/${userId}`, userData);
        },
        
        async resetUserPassword(userId, newPassword) {
            return ApiClient.post(`/api/v1/admin/users/${userId}/reset-password`, {
                new_password: newPassword
            });
        },
        
        async getLogs(level = null, limit = 100) {
            let endpoint = `/api/v1/admin/logs?limit=${limit}`;
            if (level) endpoint += `&level=${level}`;
            return ApiClient.get(endpoint);
        }
    },
    
    // 標籤 API 客戶端
    Tags: {
        async getAll() {
            return ApiClient.get('/api/v1/tags');
        },
        
        async getPopular(limit = 10) {
            return ApiClient.get(`/api/v1/tags/popular?limit=${limit}`);
        }
    },
    
    // 搜尋 API 客戶端
    Search: {
        async characters(query, filters = {}) {
            let endpoint = `/api/v1/search/characters?q=${encodeURIComponent(query)}`;
            
            Object.keys(filters).forEach(key => {
                if (filters[key]) {
                    endpoint += `&${key}=${encodeURIComponent(filters[key])}`;
                }
            });
            
            return ApiClient.get(endpoint);
        },
        
        async chats(query, sessionId = null) {
            let endpoint = `/api/v1/search/chats?q=${encodeURIComponent(query)}`;
            if (sessionId) endpoint += `&session_id=${sessionId}`;
            return ApiClient.get(endpoint);
        },
        
        async messages(query, characterId = null) {
            let endpoint = `/api/v1/search/messages?q=${encodeURIComponent(query)}`;
            if (characterId) endpoint += `&character_id=${characterId}`;
            return ApiClient.get(endpoint);
        },
        
        async global(query) {
            return ApiClient.get(`/api/v1/search/global?q=${encodeURIComponent(query)}`);
        }
    },
    
    // 監控 API 客戶端
    Monitor: {
        async getHealth() {
            return ApiClient.get('/api/v1/monitor/health');
        },
        
        async getStats() {
            return ApiClient.get('/api/v1/monitor/stats');
        },
        
        async getMetrics() {
            return ApiClient.get('/api/v1/monitor/metrics');
        },
        
        async getReady() {
            return ApiClient.get('/api/v1/monitor/ready');
        },
        
        async getLive() {
            return ApiClient.get('/api/v1/monitor/live');
        }
    },
    
    // 系統 API 客戶端
    System: {
        async getVersion() {
            return ApiClient.get('/api/v1/version');
        },
        
        async getStatus() {
            return ApiClient.get('/api/v1/status');
        },
        
        async getHealth() {
            return ApiClient.get('/health');
        }
    },
    
    // 小說 API 客戶端
    Novel: {
        async start(characterId, theme) {
            return ApiClient.post('/api/v1/novel/start', {
                character_id: characterId,
                theme: theme
            });
        },
        
        async choice(novelId, choiceId) {
            return ApiClient.post('/api/v1/novel/choice', {
                novel_id: novelId,
                choice_id: choiceId
            });
        },
        
        async getProgress(novelId) {
            return ApiClient.get(`/api/v1/novel/progress/${novelId}`);
        },
        
        async getList() {
            return ApiClient.get('/api/v1/novel/list');
        },
        
        async saveProgress(novelId, progressData) {
            return ApiClient.post('/api/v1/novel/progress/save', {
                novel_id: novelId,
                ...progressData
            });
        },
        
        async getProgressList() {
            return ApiClient.get('/api/v1/novel/progress/list');
        },
        
        async getStats(id) {
            return ApiClient.get(`/api/v1/novel/${id}/stats`);
        },
        
        async deleteProgress(id) {
            return ApiClient.delete(`/api/v1/novel/progress/${id}`);
        }
    },
    
    // 情感 API 客戶端
    Emotion: {
        async getStatus(characterId = 'char_001') {
            const userId = AppState.currentUser?.id || 'test_user_01';
            return ApiClient.get(`/api/v1/emotion/status?character_id=${characterId}&user_id=${userId}`);
        },
        
        async getAffection(characterId = 'char_001') {
            const userId = AppState.currentUser?.id || 'test_user_01';
            return ApiClient.get(`/api/v1/emotion/affection?character_id=${characterId}&user_id=${userId}`);
        },
        
        async triggerEvent(characterId, eventType, intensity = 5) {
            const userId = AppState.currentUser?.id || 'test_user_01';
            return ApiClient.post('/api/v1/emotion/event', {
                character_id: characterId,
                user_id: userId,
                event_type: eventType,
                intensity: intensity
            });
        },
        
        async getAffectionHistory(characterId, days = 30) {
            const userId = AppState.currentUser?.id || 'test_user_01';
            return ApiClient.get(`/api/v1/emotion/affection/history?character_id=${characterId}&user_id=${userId}&days=${days}`);
        },
        
        async getMilestones(characterId) {
            const userId = AppState.currentUser?.id || 'test_user_01';
            return ApiClient.get(`/api/v1/emotion/milestones?character_id=${characterId}&user_id=${userId}`);
        }
    }
};

// 頁面初始化框架
const PageInit = {
    // 通用頁面初始化
    async initPage(options = {}) {
        const {
            requireAuth = true,
            redirectOnNoAuth = '/auth',
            loadUserInfo = false,
            onInitComplete = null
        } = options;
        
        try {
            // 初始化應用
            await App.init();
            
            // 認證檢查
            if (requireAuth && !AppState.isAuthenticated) {
                Router.navigate(redirectOnNoAuth);
                return false;
            }
            
            // 載入用戶信息
            if (loadUserInfo && AppState.isAuthenticated) {
                await this.loadUserInfo();
            }
            
            // 執行自定義初始化邏輯
            if (onInitComplete && typeof onInitComplete === 'function') {
                await onInitComplete();
            }
            
            return true;
        } catch (error) {
            UI.handleError(error, '頁面初始化');
            return false;
        }
    },
    
    // 載入用戶信息的通用方法
    async loadUserInfo() {
        try {
            const user = AppState.currentUser || Utils.storage.get('user');
            if (user) {
                // 更新用戶昵稱顯示
                const nicknameElements = document.querySelectorAll('[data-user-nickname]');
                nicknameElements.forEach(el => {
                    el.textContent = user.nickname || user.username || '用戶';
                });
                
                // 更新用戶名顯示
                const usernameElements = document.querySelectorAll('[data-user-username]');
                usernameElements.forEach(el => {
                    el.textContent = user.username || '用戶';
                });
            }
        } catch (error) {
            console.warn('載入用戶信息失敗:', error);
        }
    },
    
    // 標準化的標籤切換系統
    switchTab(tabName, element) {
        return CommonFunctions.TabSystem.switchTab(tabName, element);
    }
};

// 常用功能函數
const CommonFunctions = {
    // TTS 相關功能
    TTS: {
        async generateAndPlay(text, voice = 'alloy', speed = 1.0) {
            try {
                const result = await SpecializedApiClients.TTS.generateSpeech(text, voice, speed);
                
                if (result.success && result.data && result.data.result) {
                    const audioUrl = result.data.result.audio_url;
                    const audio = new Audio(audioUrl);
                    audio.play();
                    return { audio, result };
                }
                
                throw new Error('TTS 生成失敗');
            } catch (error) {
                UI.handleError(error, 'TTS 生成');
                throw error;
            }
        },
        
        async downloadAudio(base64Data, filename = 'tts_audio.mp3') {
            try {
                const byteCharacters = atob(base64Data);
                const byteNumbers = new Array(byteCharacters.length);
                for (let i = 0; i < byteCharacters.length; i++) {
                    byteNumbers[i] = byteCharacters.charCodeAt(i);
                }
                const byteArray = new Uint8Array(byteNumbers);
                const blob = new Blob([byteArray], { type: 'audio/mp3' });

                const url = URL.createObjectURL(blob);
                const a = document.createElement('a');
                a.href = url;
                a.download = filename;
                document.body.appendChild(a);
                a.click();
                document.body.removeChild(a);
                URL.revokeObjectURL(url);

                UI.showToast('音頻下載成功', 'success');
            } catch (error) {
                UI.handleError(error, '音頻下載');
            }
        }
    },
    
    // 表單處理功能
    Form: {
        // 表單驗證
        validate(formElement, rules) {
            const errors = {};
            
            Object.keys(rules).forEach(fieldName => {
                const field = formElement.querySelector(`[name="${fieldName}"]`);
                const rule = rules[fieldName];
                
                if (!field) return;
                
                const value = field.value.trim();
                
                // 必填驗證
                if (rule.required && !value) {
                    errors[fieldName] = rule.required;
                    return;
                }
                
                // 類型驗證
                if (value && rule.type) {
                    switch (rule.type) {
                        case 'email':
                            if (!Utils.validation.email(value)) {
                                errors[fieldName] = '請輸入有效的電子郵件地址';
                            }
                            break;
                        case 'password':
                            if (!Utils.validation.password(value)) {
                                errors[fieldName] = '密碼至少8位，包含大小寫字母和數字';
                            }
                            break;
                        case 'age':
                            const age = Utils.validation.age(value);
                            if (age < 18) {
                                errors[fieldName] = '必須年滿18歲';
                            }
                            break;
                    }
                }
                
                // 長度驗證
                if (value && rule.minLength && value.length < rule.minLength) {
                    errors[fieldName] = `至少需要 ${rule.minLength} 個字符`;
                }
                
                if (value && rule.maxLength && value.length > rule.maxLength) {
                    errors[fieldName] = `不能超過 ${rule.maxLength} 個字符`;
                }
            });
            
            return errors;
        },
        
        // 顯示錯誤
        showErrors(formElement, errors) {
            // 清除之前的錯誤
            formElement.querySelectorAll('.error-message').forEach(el => el.remove());
            formElement.querySelectorAll('.border-red-500').forEach(el => {
                el.classList.remove('border-red-500');
                el.classList.add('border-gray-300');
            });
            
            // 顯示新錯誤
            Object.keys(errors).forEach(fieldName => {
                const field = formElement.querySelector(`[name="${fieldName}"]`);
                if (field) {
                    field.classList.remove('border-gray-300');
                    field.classList.add('border-red-500');
                    
                    const errorDiv = document.createElement('div');
                    errorDiv.className = 'error-message text-red-500 text-sm mt-1';
                    errorDiv.textContent = errors[fieldName];
                    
                    field.parentNode.appendChild(errorDiv);
                }
            });
        },
        
        // 序列化表單
        serialize(formElement) {
            const formData = new FormData(formElement);
            const data = {};
            
            for (let [key, value] of formData.entries()) {
                data[key] = value;
            }
            
            return data;
        }
    },
    
    // 分頁處理
    Pagination: {
        // 創建分頁元素
        create(currentPage, totalPages, onPageChange) {
            const paginationContainer = document.createElement('div');
            paginationContainer.className = 'flex justify-center items-center space-x-2 mt-4';
            
            // 上一頁按鈕
            if (currentPage > 1) {
                const prevBtn = document.createElement('button');
                prevBtn.className = 'btn-simple';
                prevBtn.textContent = '上一頁';
                prevBtn.onclick = () => onPageChange(currentPage - 1);
                paginationContainer.appendChild(prevBtn);
            }
            
            // 頁碼按鈕
            const startPage = Math.max(1, currentPage - 2);
            const endPage = Math.min(totalPages, currentPage + 2);
            
            for (let i = startPage; i <= endPage; i++) {
                const pageBtn = document.createElement('button');
                pageBtn.className = i === currentPage ? 'btn-primary' : 'btn-simple';
                pageBtn.textContent = i;
                pageBtn.onclick = () => onPageChange(i);
                paginationContainer.appendChild(pageBtn);
            }
            
            // 下一頁按鈕
            if (currentPage < totalPages) {
                const nextBtn = document.createElement('button');
                nextBtn.className = 'btn-simple';
                nextBtn.textContent = '下一頁';
                nextBtn.onclick = () => onPageChange(currentPage + 1);
                paginationContainer.appendChild(nextBtn);
            }
            
            return paginationContainer;
        }
    },
    
    // 標籤切換功能
    TabSystem: {
        // 切換標籤
        switchTab(tabName, buttonElement, contentPrefix = 'tab-content') {
            // 隱藏所有分頁內容
            document.querySelectorAll(`.${contentPrefix}`).forEach(content => {
                content.style.display = 'none';
            });
            
            // 移除所有分頁按鈕的樣式
            document.querySelectorAll('.tab-btn').forEach(tab => {
                tab.classList.remove('active');
            });
            
            // 顯示選中的分頁內容
            const targetContent = document.getElementById(tabName);
            if (targetContent) {
                targetContent.style.display = 'block';
            }
            
            // 設置選中的分頁按鈕樣式
            if (buttonElement) {
                buttonElement.classList.add('active');
            }
        }
    },
    
    // 搜尋功能
    Search: {
        // 防抖搜尋
        debounceSearch(callback, delay = 300) {
            let timeoutId;
            return function(query) {
                clearTimeout(timeoutId);
                timeoutId = setTimeout(() => callback(query), delay);
            };
        },
        
        // 高亮搜尋結果
        highlightText(text, query) {
            if (!query) return text;
            
            const regex = new RegExp(`(${query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi');
            return text.replace(regex, '<span class="highlight">$1</span>');
        }
    }
};

// 導出全局對象
window.AppState = AppState;
window.Utils = Utils;
window.ApiClient = ApiClient;
window.SpecializedApiClients = SpecializedApiClients;
window.CommonFunctions = CommonFunctions;
window.PageInit = PageInit;
window.API_PATHS = API_PATHS;
window.UI = UI;
window.Router = Router;
window.App = App;
window.NSFWSystem = NSFWSystem;
window.ResponseNormalizer = ResponseNormalizer;
