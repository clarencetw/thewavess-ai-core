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
            if (response.status === 401 && !endpoint.includes('/auth/refresh')) {
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
            container.className = 'fixed top-5 right-5 z-50 space-y-2';
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
        1: { name: '日常對話', color: 'bg-green-500', description: '工作、興趣、天氣' },
        2: { name: '浪漫內容', color: 'bg-yellow-500', description: '愛你、想你、約會' },
        3: { name: '親密內容', color: 'bg-orange-500', description: '擁抱、親吻、愛撫' },
        4: { name: '成人內容', color: 'bg-red-500', description: '身體接觸、情慾表達' },
        5: { name: '明確內容', color: 'bg-purple-500', description: '性器官描述、明確性行為' }
    },
    
    getLevelInfo(level) {
        return this.levels[level] || this.levels[1];
    },
    
    checkAge(birthDate) {
        const age = Utils.validation.age(birthDate);
        return age >= 18;
    }
};

// 導出全局對象
window.AppState = AppState;
window.Utils = Utils;
window.ApiClient = ApiClient;
window.UI = UI;
window.Router = Router;
window.App = App;
window.NSFWSystem = NSFWSystem;