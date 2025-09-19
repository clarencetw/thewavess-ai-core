/* Thewavess AI Core - API Client Library */

// =============================================================================
// API Client - 專注於 API 請求和響應處理
// =============================================================================

const API = {
    baseUrl: '/api/v1',
    
    // 創建 axios 實例
    client: axios.create({
        baseURL: '/api/v1',
        headers: {
            'Content-Type': 'application/json'
        }
    }),
    
    // 初始化攔截器
    init() {
        // 請求攔截器 - 自動添加認證 token
        this.client.interceptors.request.use(
            (config) => {
                const token = localStorage.getItem('adminToken');
                if (token) {
                    config.headers.Authorization = `Bearer ${token}`;
                }
                console.log(`🚀 API 請求: ${config.method?.toUpperCase()} ${config.url}`, config.data || '');
                return config;
            },
            (error) => {
                console.error('❌ 請求攔截器錯誤:', error);
                return Promise.reject(error);
            }
        );
        
        // 響應攔截器 - 統一處理響應和錯誤
        this.client.interceptors.response.use(
            (response) => {
                console.log(`✅ API 響應: ${response.config.method?.toUpperCase()} ${response.config.url}`, response.data);
                return response;
            },
            async (error) => {
                console.error(`❌ API 錯誤: ${error.config?.method?.toUpperCase()} ${error.config?.url}`, {
                    status: error.response?.status,
                    statusText: error.response?.statusText,
                    data: error.response?.data
                });

                // 處理認證錯誤
                if (error.response?.status === 401) {
                    console.warn('🔒 認證失敗，嘗試刷新 Token 或重新登入');

                    // 如果不是登入請求和刷新請求，嘗試刷新 token
                    const originalRequest = error.config;
                    if (!originalRequest._retry &&
                        !originalRequest.url.includes('/admin/auth/login') &&
                        !originalRequest.url.includes('/auth/refresh')) {

                        originalRequest._retry = true;

                        const refreshResult = await API.auth.refreshToken();
                        if (refreshResult.success) {
                            // 更新原請求的 Authorization header
                            const newToken = localStorage.getItem('adminToken');
                            originalRequest.headers.Authorization = `Bearer ${newToken}`;

                            // 重試原請求
                            return API.client(originalRequest);
                        }
                    }

                    // 刷新失敗或其他情況，清除認證並跳轉到登入頁
                    API.clearAuth();
                    if (window.location.pathname !== '/admin/login') {
                        window.location.href = '/admin/login';
                    }
                }

                return Promise.reject(error);
            }
        );
    },
    
    // HTTP 方法 - 統一請求處理器
    async _makeRequest(method, url, data = null, config = {}) {
        try {
            let response;
            switch (method.toLowerCase()) {
                case 'get':
                    response = await this.client.get(url, config);
                    break;
                case 'post':
                    response = await this.client.post(url, data, config);
                    break;
                case 'put':
                    response = await this.client.put(url, data, config);
                    break;
                case 'delete':
                    response = await this.client.delete(url, config);
                    break;
                default:
                    throw new Error(`不支持的 HTTP 方法: ${method}`);
            }
            return response.data;
        } catch (error) {
            return this._handleError(error);
        }
    },
    
    async get(url, config = {}) {
        return this._makeRequest('get', url, null, config);
    },
    
    async post(url, data, config = {}) {
        return this._makeRequest('post', url, data, config);
    },
    
    async put(url, data, config = {}) {
        return this._makeRequest('put', url, data, config);
    },
    
    async delete(url, config = {}) {
        return this._makeRequest('delete', url, null, config);
    },
    
    // 錯誤處理
    _handleError(error) {
        if (error.response) {
            // 服務器響應了錯誤狀態
            return {
                success: false,
                error: error.response.status,
                message: error.response.data?.message || '服務器錯誤'
            };
        } else if (error.request) {
            // 請求發送但沒有收到響應
            return {
                success: false,
                error: 'NETWORK_ERROR',
                message: '網路錯誤，請檢查連接'
            };
        } else {
            // 其他錯誤
            return {
                success: false,
                error: 'UNKNOWN_ERROR',
                message: error.message || '未知錯誤'
            };
        }
    },
    
    // 認證相關 API
    auth: {
        async login(username, password) {
            console.log('🚀 管理員登入:', { username });
            const result = await API.post('/admin/auth/login', { username, password });

            if (result.success) {
                console.log('🔐 登入成功，儲存認證資訊');
                API.setAuth(result.data.access_token, result.data.admin);
            }

            return result;
        },

        async refreshToken() {
            console.log('🔄 嘗試刷新 Token');
            const refreshToken = localStorage.getItem('adminRefreshToken');

            if (!refreshToken) {
                console.warn('⚠️ 沒有 Refresh Token，需要重新登入');
                this.logout();
                return { success: false, message: '沒有刷新令牌' };
            }

            try {
                const result = await API.post('/auth/refresh', { refresh_token: refreshToken });

                if (result.success) {
                    console.log('✅ Token 刷新成功');
                    API.setAuth(result.data.access_token, null, result.data.refresh_token);
                    return result;
                } else {
                    console.warn('⚠️ Token 刷新失敗，需要重新登入');
                    this.logout();
                    return result;
                }
            } catch (error) {
                console.error('❌ Token 刷新錯誤:', error);
                this.logout();
                return { success: false, message: '刷新令牌失敗' };
            }
        },
        
        logout() {
            console.log('📤 登出，清除認證資訊');
            API.clearAuth();
            return { success: true };
        }
    },
    
    // 管理員 API
    admin: {
        getStats: () => API.get('/admin/stats'),
        
        getUsers: (params = {}) => {
            return API.get('/admin/users', { params });
        },
        
        getUserById: (userId) => API.get(`/admin/users/${userId}`),
        
        updateUser: (userId, data) => API.put(`/admin/users/${userId}`, data),
        
        updateUserStatus: (userId, status) => API.put(`/admin/users/${userId}/status`, { status }),
        
        deleteUser: (userId) => API.delete(`/admin/users/${userId}`),
        
        getChats: (params = {}) => {
            return API.get('/admin/chats', { params });
        },
        
        getChatHistory: (chatId) => API.get(`/admin/chats/${chatId}/history`),
        
        exportChat: (chatId) => API.get(`/admin/chats/${chatId}/export`),
        
        // 使用公開角色API（管理員也可以訪問）
        getCharacters: (params = {}) => API.get('/character/list', { params })
    },

    // 聊天記錄 API
    chats: {
        getHistory: (chatId) => API.get(`/chats/${chatId}/history`),
        export: (chatId) => API.get(`/chats/${chatId}/export`),
        search: (params = {}) => API.get('/search/chats', { params })
    },
    
    // 監控 API
    monitor: {
        getHealth: () => API.get('/monitor/health'),
        getStats: () => API.get('/monitor/stats'),
        getMetrics: () => API.get('/monitor/metrics'),
        getReady: () => API.get('/monitor/ready'),
        getLive: () => API.get('/monitor/live')
    },
    
    // 認證狀態管理
    setAuth(token, adminInfo, refreshToken = null) {
        console.log('💾 儲存認證資料:', {
            token: token.substring(0, 20) + '...',
            admin: adminInfo ? adminInfo.username + ' (' + adminInfo.role + ')' : '(保持現有資料)',
            hasRefreshToken: !!refreshToken
        });
        localStorage.setItem('adminToken', token);
        if (adminInfo) {
            localStorage.setItem('adminInfo', JSON.stringify(adminInfo));
        }
        if (refreshToken) {
            localStorage.setItem('adminRefreshToken', refreshToken);
        }
    },
    
    clearAuth() {
        console.log('🗑️ 清除認證資料');
        localStorage.removeItem('adminToken');
        localStorage.removeItem('adminInfo');
        localStorage.removeItem('adminRefreshToken');
    },
    
    getAuth() {
        const token = localStorage.getItem('adminToken');
        const adminInfoStr = localStorage.getItem('adminInfo');
        
        if (!token || !adminInfoStr) {
            return null;
        }
        
        try {
            return {
                token,
                adminInfo: JSON.parse(adminInfoStr)
            };
        } catch (e) {
            console.error('認證資料解析失敗:', e);
            API.clearAuth();
            return null;
        }
    },
    
    isAuthenticated() {
        const auth = API.getAuth();
        const isAuth = !!auth;
        console.log('🔐 認證狀態檢查:', {
            status: isAuth ? '已認證' : '未認證',
            hasToken: !!auth?.token,
            hasAdminInfo: !!auth?.adminInfo,
            tokenPreview: auth?.token ? auth.token.substring(0, 30) + '...' : null,
            adminUser: auth?.adminInfo?.username
        });
        return isAuth;
    },
    
    // 監控 API
    monitor: {
        async getStats() {
            return await API.get('/monitor/stats');
        },
        
        async getHealth() {
            return await API.get('/monitor/health');
        },
        
        async getMetrics() {
            return await API.get('/monitor/metrics');
        }
    }
};

// =============================================================================
// 工具函數
// =============================================================================

const Utils = {
    // 格式化日期
    formatDate(dateString, options = {}) {
        if (!dateString) return null;
        const date = new Date(dateString);
        return date.toLocaleDateString('zh-TW', options) + ' ' + 
               date.toLocaleTimeString('zh-TW', { hour12: false });
    },
    
    // 防抖函數
    debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    },
    
    // 顯示/隱藏元素
    show: (element) => element?.classList.remove('hidden'),
    hide: (element) => element?.classList.add('hidden'),
    toggle: (element) => element?.classList.toggle('hidden'),
    
    // 通過 ID 操作元素
    showById: (id) => Utils.show(document.getElementById(id)),
    hideById: (id) => Utils.hide(document.getElementById(id)),
    toggleById: (id) => Utils.toggle(document.getElementById(id)),
    
    // 狀態處理 - 統一狀態顯示函數
    renderStatus(status) {
        const statusConfig = {
            'active': { class: 'bg-green-100 text-green-800', text: '活躍' },
            'inactive': { class: 'bg-yellow-100 text-yellow-800', text: '未活躍' },
            'blocked': { class: 'bg-red-100 text-red-800', text: '已封鎖' },
            'running': { class: 'bg-green-100 text-green-800', text: '正常' },
            'error': { class: 'bg-red-100 text-red-800', text: '異常' },
            'warning': { class: 'bg-yellow-100 text-yellow-800', text: '警告' }
        };
        
        const config = statusConfig[status] || { class: 'bg-gray-100 text-gray-800', text: '未知' };
        return `<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${config.class}">${config.text}</span>`;
    },
    
    // 向後兼容的函數
    getStatusClass(status) {
        const statusConfig = {
            'active': 'bg-green-100 text-green-800',
            'inactive': 'bg-yellow-100 text-yellow-800', 
            'blocked': 'bg-red-100 text-red-800'
        };
        return statusConfig[status] || 'bg-gray-100 text-gray-800';
    },
    
    getStatusText(status) {
        const statusConfig = {
            'active': '活躍',
            'inactive': '未活躍',
            'blocked': '已封鎖'
        };
        return statusConfig[status] || '未知';
    },

    // 統計卡片渲染函數 - 統一所有統計卡片的樣式
    renderStatsCards(cards) {
        return cards.map(card => {
            // 格式化數值顯示
            let valueDisplay = card.value;
            if (typeof card.value === 'number' && !card.isUptime && !card.isMemory && !card.isResponseTime && !card.isAIEngine) {
                valueDisplay = card.value.toLocaleString();
            }

            // 格式化變化信息
            let changeDisplay = '';
            if (card.change !== undefined && card.changeText) {
                if (card.isResponseTime) {
                    changeDisplay = `<p class="text-sm text-gray-500 mt-1">${card.changeText}: ${card.change}%</p>`;
                } else if (card.isMemory) {
                    changeDisplay = `<p class="text-sm text-gray-500 mt-1">${card.change} ${card.changeText}</p>`;
                } else if (card.isAIEngine) {
                    changeDisplay = `<p class="text-sm text-blue-600 mt-1">${card.change} ${card.changeText}</p>`;
                    // 添加引擎詳細信息
                    if (card.aiEngineData) {
                        const openai = card.aiEngineData.openai_requests || 0;
                        const grok = card.aiEngineData.grok_requests || 0;
                        changeDisplay += `<p class="text-xs text-gray-500 mt-1">OpenAI: ${openai} | Grok: ${grok}</p>`;
                    }
                } else if (card.change > 0) {
                    changeDisplay = `<p class="text-sm text-green-600 mt-1">+${card.change} ${card.changeText}</p>`;
                } else if (card.changeText) {
                    changeDisplay = `<p class="text-sm text-gray-500 mt-1">${card.changeText}</p>`;
                }
            }

            return `
                <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
                    <div class="flex items-center justify-between">
                        <div>
                            <p class="text-sm text-gray-600 font-medium">${card.title}</p>
                            <p class="text-2xl font-bold text-gray-900">${valueDisplay}</p>
                            ${changeDisplay}
                        </div>
                        <div class="text-${card.color}-600">
                            <i class="${card.icon} text-2xl"></i>
                        </div>
                    </div>
                </div>
            `;
        }).join('');
    },

    // 分頁渲染函數 - 統一分頁組件樣式
    renderPagination(pagination, onPageChange) {
        if (!pagination || pagination.totalPages <= 1) return '';
        
        const { currentPage, totalPages, hasNext, hasPrev } = pagination;
        const pages = [];
        
        // 計算顯示的頁碼範圍
        const startPage = Math.max(1, currentPage - 2);
        const endPage = Math.min(totalPages, currentPage + 2);
        
        // 上一頁
        pages.push(`
            <button ${!hasPrev ? 'disabled' : ''} 
                onclick="${onPageChange}(${currentPage - 1})"
                class="px-3 py-1 text-sm border border-gray-300 rounded-md ${!hasPrev ? 'text-gray-400 cursor-not-allowed' : 'text-gray-700 hover:bg-gray-50'}">
                上一頁
            </button>
        `);
        
        // 頁碼
        for (let i = startPage; i <= endPage; i++) {
            pages.push(`
                <button onclick="${onPageChange}(${i})"
                    class="px-3 py-1 text-sm border border-gray-300 rounded-md ${i === currentPage ? 'bg-blue-600 text-white' : 'text-gray-700 hover:bg-gray-50'}">
                    ${i}
                </button>
            `);
        }
        
        // 下一頁
        pages.push(`
            <button ${!hasNext ? 'disabled' : ''} 
                onclick="${onPageChange}(${currentPage + 1})"
                class="px-3 py-1 text-sm border border-gray-300 rounded-md ${!hasNext ? 'text-gray-400 cursor-not-allowed' : 'text-gray-700 hover:bg-gray-50'}">
                下一頁
            </button>
        `);
        
        return `
            <div class="flex items-center justify-between px-4 py-3 bg-white border-t border-gray-200 sm:px-6">
                <div class="flex justify-between sm:hidden">
                    ${pages[0]}
                    ${pages[pages.length - 1]}
                </div>
                <div class="hidden sm:flex sm:flex-1 sm:items-center sm:justify-between">
                    <div>
                        <p class="text-sm text-gray-700">
                            顯示 <span class="font-medium">${(currentPage - 1) * pagination.pageSize + 1}</span> 至 
                            <span class="font-medium">${Math.min(currentPage * pagination.pageSize, pagination.total)}</span> 
                            共 <span class="font-medium">${pagination.total}</span> 筆結果
                        </p>
                    </div>
                    <div class="flex space-x-1">
                        ${pages.join('')}
                    </div>
                </div>
            </div>
        `;
    },
    
    // HTML 轉義函數
    escapeHtml(text) {
        if (!text) return '';
        return text
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#x27;');
    }
};

// =============================================================================
// 管理頁面功能 - AdminPages
// =============================================================================

const AdminPages = {
    // 通用功能
    common: {
        showLoading(containerId) {
            const container = document.getElementById(containerId);
            if (container) {
                container.innerHTML = `
                    <div class="p-8 text-center">
                        <i class="fas fa-spinner animate-spin text-2xl text-blue-600 mb-2"></i>
                        <p class="text-gray-600">載入中...</p>
                    </div>
                `;
            }
        },

        showError(containerId, message = '載入失敗') {
            const container = document.getElementById(containerId);
            if (container) {
                container.innerHTML = `
                    <div class="p-8 text-center">
                        <i class="fas fa-exclamation-triangle text-2xl text-red-600 mb-2"></i>
                        <p class="text-red-600">${message}</p>
                    </div>
                `;
            }
        },

        showEmpty(containerId, message = '沒有資料') {
            const container = document.getElementById(containerId);
            if (container) {
                container.innerHTML = `
                    <div class="p-8 text-center">
                        <i class="fas fa-folder-open text-4xl text-gray-300 mb-4"></i>
                        <p class="text-gray-600">${message}</p>
                    </div>
                `;
            }
        },

        // 統一搜索初始化函數
        initSearchInput(inputId, searchCallback, debounceMs = 500) {
            const searchInput = document.getElementById(inputId);
            if (searchInput) {
                const debouncedSearch = Utils.debounce((query) => {
                    searchCallback(query);
                }, debounceMs);
                
                searchInput.addEventListener('input', (e) => {
                    debouncedSearch(e.target.value);
                });
            }
        },

        // 統一載入狀態管理
        setLoadingState(tableLoadingId, tableEmptyId, isLoading = true) {
            if (isLoading) {
                Utils.showById(tableLoadingId);
                Utils.hideById(tableEmptyId);
            } else {
                Utils.hideById(tableLoadingId);
            }
        },

        // 統一錯誤處理
        handleTableError(errorId, message = '載入失敗') {
            this.showError(errorId, message);
            console.error('表格載入失敗:', message);
        },

        showAlert(message, type = 'success') {
            const alertId = type === 'success' ? 'successAlert' : 'errorAlert';
            const messageId = type === 'success' ? 'successMessage' : 'errorMessage';
            
            document.getElementById(messageId).textContent = message;
            Utils.showById(alertId);
            
            // 自動隱藏
            setTimeout(() => Utils.hideById(alertId), 3000);
        },

        createPagination(currentPage, totalPages, onPageChange) {
            if (totalPages <= 1) return '';
            
            let pagination = '<div class="flex items-center justify-between"><div class="flex items-center space-x-2">';
            
            // 上一頁
            if (currentPage > 1) {
                pagination += `<button onclick="${onPageChange}(${currentPage - 1})" class="px-3 py-2 text-sm text-gray-600 hover:text-blue-600 border border-gray-300 rounded-md hover:border-blue-300 transition-colors">上一頁</button>`;
            }
            
            // 頁碼
            const startPage = Math.max(1, currentPage - 2);
            const endPage = Math.min(totalPages, currentPage + 2);
            
            for (let i = startPage; i <= endPage; i++) {
                const isActive = i === currentPage;
                const btnClass = isActive 
                    ? 'px-3 py-2 text-sm bg-blue-600 text-white border border-blue-600 rounded-md'
                    : 'px-3 py-2 text-sm text-gray-600 hover:text-blue-600 border border-gray-300 rounded-md hover:border-blue-300 transition-colors';
                pagination += `<button onclick="${onPageChange}(${i})" class="${btnClass}">${i}</button>`;
            }
            
            // 下一頁
            if (currentPage < totalPages) {
                pagination += `<button onclick="${onPageChange}(${currentPage + 1})" class="px-3 py-2 text-sm text-gray-600 hover:text-blue-600 border border-gray-300 rounded-md hover:border-blue-300 transition-colors">下一頁</button>`;
            }
            
            pagination += '</div><div class="text-sm text-gray-600">';
            pagination += `第 ${currentPage} 頁，共 ${totalPages} 頁`;
            pagination += '</div></div>';
            
            return pagination;
        }
    },

    // 儀表板頁面
    dashboard: {
        currentData: null,
        autoRefreshInterval: null,
        isAutoRefreshEnabled: false,
        metricsHistory: {
            memory: [],
            goroutines: [],
            gc: [],
            timestamps: []
        },
        alerts: [],
        alertsVisible: false,
        alertThresholds: {
            goroutines: 100,  // 超過100個goroutines發出警告
            memoryMB: 500,    // 超過500MB發出警告
            gcCount: 50,      // GC次數過多
            dbLatencyMs: 1000 // 資料庫延遲超過1秒
        },

        async init() {
            console.log('📊 初始化儀表板');
            await this.loadStats();
            await this.loadSystemStatus();
            await this.loadPerformanceMetrics();
            await this.loadExtendedSystemInfo();
            await this.loadRecentActivity();
        },

        async reload() {
            console.log('🔄 重新載入儀表板');
            await this.init();
        },

        toggleAutoRefresh() {
            const btn = document.getElementById('autoRefreshBtn');
            if (!btn) return;

            if (this.isAutoRefreshEnabled) {
                // 停止自動更新
                if (this.autoRefreshInterval) {
                    clearInterval(this.autoRefreshInterval);
                    this.autoRefreshInterval = null;
                }
                this.isAutoRefreshEnabled = false;
                btn.innerHTML = '<i class="fas fa-play mr-2"></i>自動更新';
                btn.className = btn.className.replace('bg-red-600', 'bg-green-600').replace('hover:bg-red-700', 'hover:bg-green-700');
                console.log('⏸️ 自動更新已停止');
            } else {
                // 開始自動更新
                this.isAutoRefreshEnabled = true;
                this.autoRefreshInterval = setInterval(() => {
                    console.log('🔄 自動更新監控數據...');
                    this.loadSystemStatus();
                    this.loadPerformanceMetrics();
                    this.loadExtendedSystemInfo();
                    this.checkSystemAlerts();
                }, 30000); // 每30秒更新一次
                btn.innerHTML = '<i class="fas fa-pause mr-2"></i>停止更新';
                btn.className = btn.className.replace('bg-green-600', 'bg-red-600').replace('hover:bg-green-700', 'hover:bg-red-700');
                console.log('▶️ 自動更新已啟動 (30秒間隔)');
            }
        },

        async loadStats() {
            AdminPages.common.showLoading('statsGrid');
            
            try {
                const response = await API.admin.getStats();
                if (response.success) {
                    this.renderStats(response.data);
                } else {
                    AdminPages.common.showError('statsGrid', '統計資料載入失敗');
                }
            } catch (error) {
                console.error('載入統計資料失敗:', error);
                AdminPages.common.showError('statsGrid', '統計資料載入失敗');
            }
        },

        renderStats(stats) {
            const container = document.getElementById('statsGrid');
            if (!container) return;
            
            console.log('Admin stats received:', stats); // Debug log
            
            const cards = [
                {
                    title: '總用戶數',
                    value: stats.users?.total || 0,
                    icon: 'fas fa-users',
                    color: 'blue',
                    change: stats.users?.today_new || 0,
                    changeText: '今日新增'
                },
                {
                    title: '活躍用戶',
                    value: stats.users?.active_7d || 0,
                    icon: 'fas fa-user-check',
                    color: 'green',
                    change: stats.users?.week_new || 0,
                    changeText: '本週新增'
                },
                {
                    title: '聊天會話',
                    value: stats.chats?.total_sessions || 0,
                    icon: 'fas fa-comments',
                    color: 'purple',
                    change: stats.chats?.today_sessions || 0,
                    changeText: '今日新增'
                },
                {
                    title: '總訊息數',
                    value: stats.chats?.total_messages || 0,
                    icon: 'fas fa-envelope-open-text',
                    color: 'indigo',
                    change: stats.chats?.today_messages || 0,
                    changeText: '今日新增'
                },
                {
                    title: '角色數量',
                    value: stats.characters?.total || 0,
                    icon: 'fas fa-user-friends',
                    color: 'pink',
                    change: 0,
                    changeText: '活躍角色'
                },
                {
                    title: '系統運行',
                    value: stats.uptime || '0天',
                    icon: 'fas fa-server',
                    color: 'teal',
                    change: 0,
                    changeText: '持續運行',
                    isUptime: true
                },
                {
                    title: '記憶體使用',
                    value: stats.memory_usage || '0MB',
                    icon: 'fas fa-memory',
                    color: 'yellow',
                    change: parseInt(stats.go_routines) || 0,
                    changeText: 'Goroutines',
                    isMemory: true
                },
                {
                    title: 'AI 引擎使用',
                    value: this.calculateAIEngineUsage(stats),
                    icon: 'fas fa-brain',
                    color: 'cyan',
                    change: stats.ai_engines?.total_requests || 0,
                    changeText: '總請求數',
                    isAIEngine: true,
                    aiEngineData: stats.ai_engines
                },
                {
                    title: '回應時間',
                    value: stats.avg_response_time || '0ms',
                    icon: 'fas fa-tachometer-alt',
                    color: 'red',
                    change: parseFloat(stats.error_rate) || 0,
                    changeText: '錯誤率',
                    isResponseTime: true
                }
            ];
            
            // 使用統一的統計卡片渲染函數
            container.innerHTML = Utils.renderStatsCards(cards);
        },

        calculateAIEngineUsage(stats) {
            const engines = stats.ai_engines;
            if (!engines) return '未知';

            const openai = engines.openai_requests || 0;
            const grok = engines.grok_requests || 0;
            const total = openai + grok;

            if (total === 0) return '無使用';

            // 顯示主要使用的引擎
            if (openai > grok) {
                const percentage = Math.round((openai / total) * 100);
                return `OpenAI ${percentage}%`;
            } else if (grok > openai) {
                const percentage = Math.round((grok / total) * 100);
                return `Grok ${percentage}%`;
            } else {
                return '平均使用';
            }
        },

        async loadSystemStatus() {
            const container = document.getElementById('systemStatus');
            if (!container) return;
            
            try {
                // 直接調用監控 API 獲取實時系統狀態
                const response = await API.monitor.getStats();
                if (response.success) {
                    this.renderSystemStatus(response.data);
                } else {
                    container.innerHTML = '<p class="text-red-600">系統狀態載入失敗</p>';
                }
            } catch (error) {
                console.error('載入系統狀態失敗:', error);
                container.innerHTML = '<p class="text-red-600">系統狀態載入失敗</p>';
            }
        },

        renderSystemStatus(data) {
            const container = document.getElementById('systemStatus');
            if (!container) return;
            
            // 解析監控 API 的真實數據格式
            const isHealthy = data.status === 'healthy';
            const isDatabaseConnected = data.database?.connected || false;
            const isOpenAIConfigured = data.services?.openai === 'configured';
            const isGrokConfigured = data.services?.grok === 'configured';
            
            // 更新系統狀態指示器
            const statusIndicator = document.getElementById('systemStatusIndicator');
            if (statusIndicator) {
                statusIndicator.className = `px-2 py-1 text-xs rounded-full ${
                    isHealthy ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                }`;
                statusIndicator.textContent = isHealthy ? '正常運行' : '異常狀態';
            }
            
            container.innerHTML = `
                <div class="space-y-4">
                    <div class="flex justify-between items-center">
                        <span class="text-gray-600">資料庫連接</span>
                        <span class="px-3 py-1 text-xs rounded-full ${isDatabaseConnected ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}">
                            ${isDatabaseConnected ? '正常' : '異常'}
                        </span>
                    </div>
                    <div class="flex justify-between items-center">
                        <span class="text-gray-600">資料庫延遲</span>
                        <span class="text-gray-900">${data.database?.ping_latency || 'N/A'}</span>
                    </div>
                    <div class="flex justify-between items-center">
                        <span class="text-gray-600">OpenAI API</span>
                        <span class="px-2 py-1 text-xs rounded-full ${isOpenAIConfigured ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'}">
                            ${isOpenAIConfigured ? '已配置' : '未配置'}
                        </span>
                    </div>
                    <div class="flex justify-between items-center">
                        <span class="text-gray-600">Grok API</span>
                        <span class="px-2 py-1 text-xs rounded-full ${isGrokConfigured ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'}">
                            ${isGrokConfigured ? '已配置' : '未配置'}
                        </span>
                    </div>
                </div>
            `;
        },

        async loadPerformanceMetrics() {
            const container = document.getElementById('performanceMetrics');
            if (!container) return;
            
            try {
                const response = await API.monitor.getStats();
                if (response.success) {
                    this.renderPerformanceMetrics(response.data);
                } else {
                    container.innerHTML = '<p class="text-red-600">性能指標載入失敗</p>';
                }
            } catch (error) {
                console.error('載入性能指標失敗:', error);
                container.innerHTML = '<p class="text-red-600">性能指標載入失敗</p>';
            }
        },

        renderPerformanceMetrics(data) {
            const container = document.getElementById('performanceMetrics');
            if (!container) return;
            
            // 計算記憶體使用率百分比（假設系統有合理的記憶體）
            const memoryBytes = data.runtime?.memory_usage || '0 B';
            const gcCount = data.runtime?.gc_count || 0;
            const goroutines = data.runtime?.goroutines || 0;
            
            container.innerHTML = `
                <div class="space-y-4">
                    <div class="flex justify-between items-center">
                        <span class="text-gray-600">執行緒數</span>
                        <span class="text-lg font-semibold ${goroutines > 50 ? 'text-red-600' : 'text-green-600'}">${goroutines}</span>
                    </div>
                    <div class="flex justify-between items-center">
                        <span class="text-gray-600">GC 執行次數</span>
                        <span class="text-lg font-semibold text-purple-600">${gcCount}</span>
                    </div>
                    <div class="flex justify-between items-center">
                        <span class="text-gray-600">下次 GC</span>
                        <span class="text-gray-900 text-sm">${data.runtime?.next_gc || 'N/A'}</span>
                    </div>
                    <div class="flex justify-between items-center">
                        <span class="text-gray-600">記憶體堆疊</span>
                        <span class="text-gray-900 text-sm">${data.runtime?.heap_objects || 'N/A'} 物件</span>
                    </div>
                </div>
            `;
        },

        async loadExtendedSystemInfo() {
            const container = document.getElementById('extendedSystemInfo');
            if (!container) return;
            
            try {
                // 獲取更多系統信息
                const [statsResponse, healthResponse] = await Promise.all([
                    API.monitor.getStats(),
                    API.monitor.getHealth()
                ]);
                
                if (statsResponse.success && healthResponse.success) {
                    this.renderExtendedSystemInfo(statsResponse.data, healthResponse.data);
                } else {
                    container.innerHTML = '<p class="text-red-600">擴展系統資訊載入失敗</p>';
                }
            } catch (error) {
                console.error('載入擴展系統資訊失敗:', error);
                container.innerHTML = '<p class="text-red-600">擴展系統資訊載入失敗</p>';
            }
        },

        renderExtendedSystemInfo(statsData, healthData) {
            const container = document.getElementById('extendedSystemInfo');
            if (!container) return;
            
            const uptime = healthData.uptime || 'N/A';
            const version = healthData.version || 'N/A';
            const os = statsData.system?.os || 'N/A';
            const arch = statsData.system?.architecture || 'N/A';
            const cpuCores = statsData.system?.num_cpu || 'N/A';
            const goVersion = statsData.system?.go_version || 'N/A';
            
            container.innerHTML = `
                <div class="bg-gray-50 rounded-lg p-4">
                    <h4 class="font-medium text-gray-900 mb-2">運行信息</h4>
                    <div class="space-y-2 text-sm">
                        <div class="flex justify-between">
                            <span class="text-gray-600">系統運行時間</span>
                            <span class="font-mono text-blue-600">${uptime}</span>
                        </div>
                        <div class="flex justify-between">
                            <span class="text-gray-600">應用程序版本</span>
                            <span class="font-mono text-gray-900">${version}</span>
                        </div>
                    </div>
                </div>
                <div class="bg-gray-50 rounded-lg p-4">
                    <h4 class="font-medium text-gray-900 mb-2">系統架構</h4>
                    <div class="space-y-2 text-sm">
                        <div class="flex justify-between">
                            <span class="text-gray-600">操作系統</span>
                            <span class="font-mono text-gray-900">${os}</span>
                        </div>
                        <div class="flex justify-between">
                            <span class="text-gray-600">架構</span>
                            <span class="font-mono text-gray-900">${arch}</span>
                        </div>
                    </div>
                </div>
                <div class="bg-gray-50 rounded-lg p-4">
                    <h4 class="font-medium text-gray-900 mb-2">硬體規格</h4>
                    <div class="space-y-2 text-sm">
                        <div class="flex justify-between">
                            <span class="text-gray-600">CPU 核心數</span>
                            <span class="font-mono text-green-600">${cpuCores}</span>
                        </div>
                        <div class="flex justify-between">
                            <span class="text-gray-600">Go 版本</span>
                            <span class="font-mono text-gray-900">${goVersion}</span>
                        </div>
                    </div>
                </div>
                <div class="bg-gray-50 rounded-lg p-4">
                    <h4 class="font-medium text-gray-900 mb-2">服務狀態</h4>
                    <div class="space-y-2 text-sm">
                        <div class="flex justify-between">
                            <span class="text-gray-600">資料庫類型</span>
                            <span class="font-mono text-gray-900">${statsData.database?.type || 'N/A'}</span>
                        </div>
                        <div class="flex justify-between">
                            <span class="text-gray-600">TTS 服務</span>
                            <span class="font-mono text-green-600">${statsData.services?.tts || 'N/A'}</span>
                        </div>
                    </div>
                </div>
            `;
        },

        async loadRecentActivity() {
            const container = document.getElementById('recentActivity');
            if (!container) return;
            
            try {
                // 使用系統日誌作為最近活動
                const response = await API.get('/admin/logs?limit=10');
                if (response.success) {
                    this.renderRecentActivity(response.data.logs);
                } else {
                    container.innerHTML = '<p class="text-gray-600">最近活動載入失敗</p>';
                }
            } catch (error) {
                console.error('載入最近活動失敗:', error);
                container.innerHTML = '<p class="text-gray-600">最近活動載入失敗</p>';
            }
        },

        renderRecentActivity(logs) {
            const container = document.getElementById('recentActivity');
            if (!container) return;
            
            if (!logs || logs.length === 0) {
                container.innerHTML = '<p class="text-gray-600">暫無最近活動</p>';
                return;
            }
            
            container.innerHTML = `
                <div class="space-y-3">
                    ${logs.map(log => `
                        <div class="flex items-start space-x-3">
                            <div class="flex-shrink-0">
                                <i class="fas ${this.getLogIcon(log.level)} ${this.getLogColor(log.level)}"></i>
                            </div>
                            <div class="min-w-0 flex-1">
                                <p class="text-sm text-gray-900">${log.message}</p>
                                <p class="text-xs text-gray-500">${Utils.formatDate(log.timestamp)}</p>
                            </div>
                        </div>
                    `).join('')}
                </div>
            `;
        },
        
        getLogIcon(level) {
            switch(level) {
                case 'error': return 'fa-exclamation-triangle';
                case 'warning': return 'fa-exclamation-circle';
                case 'info': return 'fa-info-circle';
                default: return 'fa-circle';
            }
        },
        
        getLogColor(level) {
            switch(level) {
                case 'error': return 'text-red-500';
                case 'warning': return 'text-yellow-500';
                case 'info': return 'text-blue-500';
                default: return 'text-gray-400';
            }
        },




        // 系統警報功能
        toggleAlerts() {
            const panel = document.getElementById('alertsPanel');
            if (!panel) return;
            
            this.alertsVisible = !this.alertsVisible;
            
            if (this.alertsVisible) {
                panel.classList.remove('hidden');
                this.loadAlerts();
            } else {
                panel.classList.add('hidden');
            }
        },

        async checkSystemAlerts() {
            try {
                const response = await API.monitor.getStats();
                if (response.success) {
                    this.analyzeSystemMetrics(response.data);
                }
            } catch (error) {
                console.error('檢查系統警報失敗:', error);
            }
        },

        analyzeSystemMetrics(data) {
            const now = new Date();
            let newAlerts = [];

            // 檢查 Goroutines 數量
            const goroutines = data.runtime?.goroutines || 0;
            if (goroutines > this.alertThresholds.goroutines) {
                newAlerts.push({
                    id: `goroutines_${now.getTime()}`,
                    type: 'warning',
                    title: 'Goroutines 數量過高',
                    message: `當前 Goroutines 數量為 ${goroutines}，超過警告閾值 ${this.alertThresholds.goroutines}`,
                    timestamp: now,
                    metric: 'goroutines',
                    value: goroutines,
                    threshold: this.alertThresholds.goroutines
                });
            }

            // 檢查記憶體使用
            const memoryStr = data.runtime?.memory_usage || '0 MB';
            const memoryMB = this.parseMemoryToMB(memoryStr);
            if (memoryMB > this.alertThresholds.memoryMB) {
                newAlerts.push({
                    id: `memory_${now.getTime()}`,
                    type: 'warning',
                    title: '記憶體使用量過高',
                    message: `當前記憶體使用量為 ${memoryStr}，超過警告閾值 ${this.alertThresholds.memoryMB} MB`,
                    timestamp: now,
                    metric: 'memory',
                    value: memoryMB,
                    threshold: this.alertThresholds.memoryMB
                });
            }

            // 檢查 GC 次數
            const gcCount = data.runtime?.gc_count || 0;
            if (gcCount > this.alertThresholds.gcCount) {
                newAlerts.push({
                    id: `gc_${now.getTime()}`,
                    type: 'info',
                    title: 'GC 執行次數較高',
                    message: `當前 GC 執行次數為 ${gcCount}，可能需要關注記憶體使用模式`,
                    timestamp: now,
                    metric: 'gc',
                    value: gcCount,
                    threshold: this.alertThresholds.gcCount
                });
            }

            // 檢查資料庫連接狀態
            if (!data.database?.connected) {
                newAlerts.push({
                    id: `db_disconnected_${now.getTime()}`,
                    type: 'error',
                    title: '資料庫連接中斷',
                    message: '資料庫連接已中斷，請立即檢查資料庫服務狀態',
                    timestamp: now,
                    metric: 'database',
                    value: 'disconnected',
                    threshold: 'connected'
                });
            }

            // 添加新警報到列表
            newAlerts.forEach(alert => {
                // 檢查是否已存在相同類型的警報（避免重複）
                const exists = this.alerts.find(a => a.metric === alert.metric && a.type === alert.type);
                if (!exists) {
                    this.alerts.unshift(alert); // 新警報放在前面
                }
            });

            // 限制警報數量（最多保留20個）
            if (this.alerts.length > 20) {
                this.alerts = this.alerts.slice(0, 20);
            }

            // 更新警報計數
            this.updateAlertsCount();

            // 如果有新的嚴重警報，可以考慮自動顯示
            const criticalAlerts = newAlerts.filter(a => a.type === 'error');
            if (criticalAlerts.length > 0 && !this.alertsVisible) {
                this.showCriticalAlertNotification(criticalAlerts[0]);
            }
        },

        parseMemoryToMB(memoryStr) {
            if (!memoryStr) return 0;
            
            const match = memoryStr.match(/^(\d+\.?\d*)\s*(B|KB|MB|GB)$/i);
            if (!match) return 0;
            
            const value = parseFloat(match[1]);
            const unit = match[2].toUpperCase();
            
            switch (unit) {
                case 'B': return value / (1024 * 1024);
                case 'KB': return value / 1024;
                case 'MB': return value;
                case 'GB': return value * 1024;
                default: return 0;
            }
        },

        updateAlertsCount() {
            const countEl = document.getElementById('alertsCount');
            if (countEl) {
                countEl.textContent = this.alerts.length;
                
                const btn = document.getElementById('alertsBtn');
                if (btn) {
                    if (this.alerts.length > 0) {
                        btn.className = btn.className.replace('bg-yellow-600', 'bg-red-600').replace('hover:bg-yellow-700', 'hover:bg-red-700');
                    } else {
                        btn.className = btn.className.replace('bg-red-600', 'bg-yellow-600').replace('hover:bg-red-700', 'hover:bg-yellow-700');
                    }
                }
            }
        },

        showCriticalAlertNotification(alert) {
            // 顯示系統通知（如果瀏覽器支持）
            if ('Notification' in window && Notification.permission === 'granted') {
                new Notification('系統警報', {
                    body: alert.message,
                    icon: '/public/favicon.ico',
                    tag: alert.id
                });
            }
            
            // 顯示頁面內通知
            Utils.showAlert('error', alert.message);
        },

        loadAlerts() {
            const container = document.getElementById('alertsList');
            if (!container) return;
            
            if (this.alerts.length === 0) {
                container.innerHTML = `
                    <div class="text-center py-8 text-gray-500">
                        <i class="fas fa-check-circle text-3xl mb-2"></i>
                        <p>目前沒有系統警報</p>
                        <p class="text-sm">系統運行正常</p>
                    </div>
                `;
                return;
            }
            
            container.innerHTML = this.alerts.map(alert => {
                const typeColors = {
                    error: 'bg-red-50 border-red-200 text-red-800',
                    warning: 'bg-yellow-50 border-yellow-200 text-yellow-800',
                    info: 'bg-blue-50 border-blue-200 text-blue-800'
                };
                
                const typeIcons = {
                    error: 'fas fa-exclamation-circle text-red-500',
                    warning: 'fas fa-exclamation-triangle text-yellow-500',
                    info: 'fas fa-info-circle text-blue-500'
                };
                
                return `
                    <div class="flex items-start space-x-3 p-3 rounded-lg border ${typeColors[alert.type] || typeColors.info}">
                        <div class="flex-shrink-0 mt-1">
                            <i class="${typeIcons[alert.type] || typeIcons.info}"></i>
                        </div>
                        <div class="flex-1 min-w-0">
                            <h4 class="text-sm font-medium">${alert.title}</h4>
                            <p class="text-sm opacity-90 mt-1">${alert.message}</p>
                            <p class="text-xs opacity-75 mt-2">${Utils.formatDate(alert.timestamp)}</p>
                        </div>
                        <button onclick="AdminPages.dashboard.dismissAlert('${alert.id}')" class="flex-shrink-0 text-gray-400 hover:text-gray-600">
                            <i class="fas fa-times"></i>
                        </button>
                    </div>
                `;
            }).join('');
        },

        dismissAlert(alertId) {
            this.alerts = this.alerts.filter(alert => alert.id !== alertId);
            this.updateAlertsCount();
            this.loadAlerts();
        },

        clearAlerts() {
            this.alerts = [];
            this.updateAlertsCount();
            this.loadAlerts();
        }
    },

    // 用戶管理頁面
    users: {
        currentData: null,
        currentPage: 1,
        pageSize: 20,
        searchQuery: '',
        sortBy: 'created_at',
        sortOrder: 'desc',

        async init() {
            console.log('👥 初始化用戶管理');
            await this.loadStats();
            await this.loadUsers();
            this.initSearch();
            this.initSorting();
        },

        async reload() {
            console.log('🔄 重新載入用戶管理');
            this.currentPage = 1;
            await this.loadUsers();
        },

        initSearch() {
            AdminPages.common.initSearchInput('userSearchInput', (query) => {
                this.searchQuery = query;
                this.currentPage = 1;
                this.loadUsers();
            });
        },

        async loadStats() {
            try {
                const response = await API.admin.getStats();
                if (response.success) {
                    this.renderUserStats(response.data);
                }
            } catch (error) {
                console.error('載入用戶統計失敗:', error);
            }
        },

        renderUserStats(stats) {
            const container = document.getElementById('userStatsGrid');
            if (!container) return;
            
            console.log('User stats received:', stats); // Debug log
            
            const cards = [
                { title: '總用戶數', value: stats.users?.total || 0, icon: 'fas fa-users', color: 'blue' },
                { title: '活躍用戶', value: stats.users?.active_7d || 0, icon: 'fas fa-user-check', color: 'green' },
                { title: '已封鎖', value: stats.users?.blocked || 0, icon: 'fas fa-user-lock', color: 'red' },
                { title: '今日註冊', value: stats.users?.today_new || 0, icon: 'fas fa-user-plus', color: 'purple' }
            ];
            
            // 使用統一的統計卡片渲染函數
            container.innerHTML = Utils.renderStatsCards(cards);
        },

        async loadUsers() {
            AdminPages.common.setLoadingState('usersTableLoading', 'usersTableEmpty', true);

            try {
                const params = {
                    page: this.currentPage,
                    limit: this.pageSize,
                    sort_by: this.sortBy,
                    sort_order: this.sortOrder
                };

                if (this.searchQuery) {
                    params.search = this.searchQuery;
                }
                
                const response = await API.admin.getUsers(params);
                if (response.success) {
                    const data = response.data || {};
                    this.currentData = data;
                    this.renderUsersTable(data);
                    this.renderPagination(data.pagination);
                } else {
                    this.showUsersError(response.message || '用戶載入失敗');
                }
            } catch (error) {
                console.error('載入用戶失敗:', error);
                this.showUsersError('用戶載入失敗');
            } finally {
                AdminPages.common.setLoadingState('usersTableLoading', 'usersTableEmpty', false);
            }
        },

        renderUsersTable(data) {
            const tbody = document.getElementById('usersTableBody');
            if (!tbody || !data.users) return;

            if (data.users.length === 0) {
                Utils.showById('usersTableEmpty');
                tbody.innerHTML = '';
                return;
            }

            Utils.hideById('usersTableEmpty');
            tbody.innerHTML = data.users.map(user => `
                <tr class="hover:bg-gray-50">
                    <td class="px-6 py-4">
                        <div class="flex items-center">
                            <div class="flex-shrink-0 h-10 w-10">
                                ${user.avatar_url ? 
                                    `<img class="h-10 w-10 rounded-full object-cover" src="${user.avatar_url}" alt="${user.username}" onerror="this.style.display='none'; this.nextElementSibling.style.display='flex';">
                                     <div class="h-10 w-10 rounded-full bg-gray-300 items-center justify-center" style="display:none;">
                                        <i class="fas fa-user text-gray-600"></i>
                                     </div>` :
                                    `<div class="h-10 w-10 rounded-full bg-gray-300 flex items-center justify-center">
                                        <i class="fas fa-user text-gray-600"></i>
                                    </div>`
                                }
                            </div>
                            <div class="ml-4">
                                <div class="text-sm font-medium text-gray-900">${user.username}</div>
                                <div class="text-sm text-gray-500">${user.display_name || user.email || 'N/A'}</div>
                            </div>
                        </div>
                    </td>
                    <td class="px-6 py-4">
                        <span class="px-2 py-1 text-xs font-semibold rounded-full ${Utils.getStatusClass(user.status)}">
                            ${Utils.getStatusText(user.status)}
                        </span>
                    </td>
                    <td class="px-6 py-4 text-sm text-gray-900">
                        ${Utils.formatDate(user.created_at)}
                    </td>
                    <td class="px-6 py-4 text-sm space-x-2">
                        <button onclick="AdminPages.users.viewUser('${user.id}')" class="text-blue-600 hover:text-blue-800">
                            <i class="fas fa-eye"></i> 查看
                        </button>
                        <button onclick="AdminPages.users.editUser('${user.id}')" class="text-green-600 hover:text-green-800">
                            <i class="fas fa-edit"></i> 編輯
                        </button>
                        <button onclick="AdminPages.users.showPasswordModal('${user.id}', '${user.username}')" class="text-orange-600 hover:text-orange-800">
                            <i class="fas fa-key"></i> 密碼
                        </button>
                        <button onclick="AdminPages.users.toggleUserStatus('${user.id}', '${user.status}')"
                                class="text-${user.status === 'active' ? 'red' : 'green'}-600 hover:text-${user.status === 'active' ? 'red' : 'green'}-800">
                            <i class="fas fa-${user.status === 'active' ? 'ban' : 'check'}"></i>
                            ${user.status === 'active' ? '暫停' : '啟用'}
                        </button>
                    </td>
                </tr>
            `).join('');
        },

        renderPagination(pagination) {
            const container = document.getElementById('usersPagination');
            if (!container || !pagination) return;
            
            container.innerHTML = AdminPages.common.createPagination(
                pagination.current_page,
                pagination.total_pages,
                'AdminPages.users.goToPage'
            );
        },

        showUsersError(message) {
            const tbody = document.getElementById('usersTableBody');
            if (tbody) {
                tbody.innerHTML = `
                    <tr>
                        <td colspan="4" class="px-6 py-8 text-center">
                            <i class="fas fa-exclamation-triangle text-2xl text-red-600 mb-2"></i>
                            <p class="text-red-600">${message}</p>
                        </td>
                    </tr>
                `;
            }
        },

        async goToPage(page) {
            this.currentPage = page;
            await this.loadUsers();
        },

        async changePageSize(newPageSize) {
            this.pageSize = parseInt(newPageSize);
            this.currentPage = 1;
            await this.loadUsers();
        },

        async viewUser(userId) {
            try {
                const response = await API.admin.getUserById(userId);
                if (response.success) {
                    this.showUserModal(response.data, 'view');
                } else {
                    AdminPages.common.showAlert('用戶資料載入失敗', 'error');
                }
            } catch (error) {
                console.error('載入用戶資料失敗:', error);
                AdminPages.common.showAlert('用戶資料載入失敗', 'error');
            }
        },

        async editUser(userId) {
            try {
                const response = await API.get(`/admin/users/${userId}`);
                if (response.success) {
                    this.showUserModal(response.data, 'edit');
                } else {
                    AdminPages.common.showAlert('用戶資料載入失敗', 'error');
                }
            } catch (error) {
                console.error('載入用戶資料失敗:', error);
                AdminPages.common.showAlert('用戶資料載入失敗', 'error');
            }
        },

        showUserModal(user, mode = 'view') {
            const modalId = mode === 'edit' ? 'userEditModal' : 'userDetailModal';
            const contentId = mode === 'edit' ? 'userEditContent' : 'userDetailContent';
            
            if (mode === 'view') {
                document.getElementById(contentId).innerHTML = `
                    <div class="space-y-6">
                        <!-- 用戶基本信息卡片 -->
                        <div class="bg-gray-50 rounded-lg p-4">
                            <h4 class="text-sm font-semibold text-gray-900 mb-3 flex items-center">
                                <i class="fas fa-user-circle text-blue-600 mr-2"></i>
                                基本信息
                            </h4>
                            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div class="flex items-center">
                                    <div class="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center mr-3">
                                        <i class="fas fa-user text-blue-600 text-xs"></i>
                                    </div>
                                    <div>
                                        <p class="text-xs text-gray-500">用戶名稱</p>
                                        <p class="font-medium text-gray-900">${user.username}</p>
                                    </div>
                                </div>
                                <div class="flex items-center">
                                    <div class="w-8 h-8 bg-green-100 rounded-full flex items-center justify-center mr-3">
                                        <i class="fas fa-envelope text-green-600 text-xs"></i>
                                    </div>
                                    <div>
                                        <p class="text-xs text-gray-500">電子郵件</p>
                                        <p class="font-medium text-gray-900">${user.email || 'N/A'}</p>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <!-- 狀態信息卡片 -->
                        <div class="bg-gray-50 rounded-lg p-4">
                            <h4 class="text-sm font-semibold text-gray-900 mb-3 flex items-center">
                                <i class="fas fa-info-circle text-purple-600 mr-2"></i>
                                狀態信息
                            </h4>
                            <div class="flex items-center">
                                <div class="w-8 h-8 bg-purple-100 rounded-full flex items-center justify-center mr-3">
                                    <i class="fas fa-toggle-on text-purple-600 text-xs"></i>
                                </div>
                                <div>
                                    <p class="text-xs text-gray-500">當前狀態</p>
                                    <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${Utils.getStatusClass(user.status)}">
                                        ${Utils.getStatusText(user.status)}
                                    </span>
                                </div>
                            </div>
                        </div>

                        <!-- 時間信息卡片 -->
                        <div class="bg-gray-50 rounded-lg p-4">
                            <h4 class="text-sm font-semibold text-gray-900 mb-3 flex items-center">
                                <i class="fas fa-clock text-orange-600 mr-2"></i>
                                時間記錄
                            </h4>
                            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div class="flex items-center">
                                    <div class="w-8 h-8 bg-orange-100 rounded-full flex items-center justify-center mr-3">
                                        <i class="fas fa-calendar-plus text-orange-600 text-xs"></i>
                                    </div>
                                    <div>
                                        <p class="text-xs text-gray-500">註冊時間</p>
                                        <p class="font-medium text-gray-900">${Utils.formatDate(user.created_at)}</p>
                                    </div>
                                </div>
                                <div class="flex items-center">
                                    <div class="w-8 h-8 bg-red-100 rounded-full flex items-center justify-center mr-3">
                                        <i class="fas fa-sign-in-alt text-red-600 text-xs"></i>
                                    </div>
                                    <div>
                                        <p class="text-xs text-gray-500">最後登入</p>
                                        <p class="font-medium text-gray-900">${Utils.formatDate(user.last_login_at) || 'N/A'}</p>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                `;
            } else {
                document.getElementById(contentId).innerHTML = `
                    <div class="space-y-4">
                        <div>
                            <label class="block text-sm font-medium text-gray-700">用戶名稱</label>
                            <input type="text" name="username" value="${user.username}" class="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2">
                        </div>
                        <div>
                            <label class="block text-sm font-medium text-gray-700">電子郵件</label>
                            <input type="email" name="email" value="${user.email || ''}" class="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2">
                        </div>
                        <div>
                            <label class="block text-sm font-medium text-gray-700">狀態</label>
                            <select name="status" class="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2">
                                <option value="active" ${user.status === 'active' ? 'selected' : ''}>活躍</option>
                                <option value="inactive" ${user.status === 'inactive' ? 'selected' : ''}>未活躍</option>
                                <option value="suspended" ${user.status === 'suspended' ? 'selected' : ''}>已暫停</option>
                            </select>
                        </div>
                        <input type="hidden" name="user_id" value="${user.id}">
                    </div>
                `;
                
                // 設置表單提交處理
                const form = document.getElementById('userEditForm');
                form.onsubmit = (e) => {
                    e.preventDefault();
                    this.saveUserChanges(new FormData(form));
                };
            }
            
            Utils.showById(modalId);
        },

        async saveUserChanges(formData) {
            try {
                const userId = formData.get('user_id');
                const userData = {
                    username: formData.get('username'),
                    email: formData.get('email'),
                    status: formData.get('status')
                };
                
                const response = await API.admin.updateUser(userId, userData);
                if (response.success) {
                    AdminPages.common.showAlert('用戶更新成功');
                    Utils.hideById('userEditModal');
                    await this.loadUsers();
                } else {
                    AdminPages.common.showAlert(response.message || '更新失敗', 'error');
                }
            } catch (error) {
                console.error('更新用戶失敗:', error);
                AdminPages.common.showAlert('更新失敗', 'error');
            }
        },

        async toggleUserStatus(userId, currentStatus) {
            const newStatus = currentStatus === 'active' ? 'suspended' : 'active';
            const action = newStatus === 'suspended' ? '暫停' : '啟用';

            if (!confirm(`確定要${action}此用戶嗎？`)) return;

            try {
                const response = await API.admin.updateUserStatus(userId, newStatus);
                if (response.success) {
                    AdminPages.common.showAlert(`用戶${action}成功`);
                    await this.loadUsers();
                } else {
                    AdminPages.common.showAlert(response.message || `${action}失敗`, 'error');
                }
            } catch (error) {
                console.error(`${action}用戶失敗:`, error);
                AdminPages.common.showAlert(`${action}失敗`, 'error');
            }
        },

        showPasswordModal(userId, username) {
            const modal = document.getElementById('userPasswordModal');
            const form = document.getElementById('userPasswordForm');
            const usernameSpan = document.getElementById('passwordModalUsername');

            if (!modal || !form || !usernameSpan) {
                AdminPages.common.showAlert('密碼修改模態框初始化失敗', 'error');
                return;
            }

            // 設置用戶名顯示
            usernameSpan.textContent = username;

            // 清空表單
            form.reset();

            // 設置表單提交處理
            form.onsubmit = (e) => {
                e.preventDefault();
                this.updateUserPassword(userId);
            };

            // 顯示模態框
            modal.classList.remove('hidden');
        },

        async updateUserPassword(userId) {
            const newPassword = document.getElementById('newPassword').value;
            const confirmPassword = document.getElementById('confirmPassword').value;

            // 驗證密碼
            if (!newPassword || newPassword.length < 8) {
                AdminPages.common.showAlert('密碼長度至少需要8個字符', 'error');
                return;
            }

            if (newPassword !== confirmPassword) {
                AdminPages.common.showAlert('兩次輸入的密碼不一致', 'error');
                return;
            }

            try {
                const response = await API.client.put(`/admin/users/${userId}/password`, {
                    new_password: newPassword
                });

                if (response.data.success) {
                    AdminPages.common.showAlert('密碼修改成功');
                    this.hidePasswordModal();
                } else {
                    AdminPages.common.showAlert(response.data.message || '密碼修改失敗', 'error');
                }
            } catch (error) {
                console.error('密碼修改失敗:', error);
                const errorMessage = error.response?.data?.message || '密碼修改時發生錯誤';
                AdminPages.common.showAlert(errorMessage, 'error');
            }
        },

        hidePasswordModal() {
            const modal = document.getElementById('userPasswordModal');
            if (modal) {
                modal.classList.add('hidden');
            }
        },

        initSorting() {
            // 為每個排序按鈕綁定點擊事件
            const sortButtons = document.querySelectorAll('#usersTable th button');
            sortButtons.forEach((button, index) => {
                const fieldMap = ['username', 'status', 'created_at']; // 對應表格列的字段
                const field = fieldMap[index];
                if (field) {
                    button.addEventListener('click', () => {
                        this.sortByField(field);
                    });
                }
            });
        },

        async sortByField(field) {
            // 如果點擊同一欄位，切換排序方向
            if (this.sortBy === field) {
                this.sortOrder = this.sortOrder === 'asc' ? 'desc' : 'asc';
            } else {
                this.sortBy = field;
                this.sortOrder = 'desc'; // 預設降序
            }

            this.currentPage = 1; // 重新排序時回到第一頁
            await this.loadUsers();
        }
    },

    // 聊天記錄頁面
    chats: {
        currentData: null,
        currentPage: 1,
        pageSize: 10,
        sortBy: 'created_at',
        sortOrder: 'desc',
        filters: {
            search: '',
            user: '',
            character: '',
            dateFrom: '',
            dateTo: ''
        },

        async init() {
            console.log('💬 初始化聊天管理');
            await this.loadStats();
            await this.loadChats();
            await this.loadFilterOptions();
            this.initSearch();
            this.initSorting();
        },

        async reload() {
            console.log('🔄 重新載入聊天管理');
            this.currentPage = 1;
            await this.loadChats();
        },

        initSearch() {
            const searchInput = document.getElementById('chatSearchInput');
            if (searchInput) {
                const debouncedSearch = Utils.debounce((query) => {
                    this.filters.search = query;
                    this.currentPage = 1;
                    this.loadChats();
                }, 500);
                
                searchInput.addEventListener('input', (e) => {
                    debouncedSearch(e.target.value);
                });
            }
        },

        async loadStats() {
            try {
                const response = await API.admin.getStats();
                if (response.success) {
                    this.renderChatStats(response.data);
                }
            } catch (error) {
                console.error('載入聊天統計失敗:', error);
            }
        },

        renderChatStats(stats) {
            const container = document.getElementById('chatStatsGrid');
            if (!container) return;
            
            console.log('Chat stats received:', stats); // Debug log
            
            const cards = [
                { title: '總會話數', value: stats.chats?.total_sessions || 0, icon: 'fas fa-comments', color: 'blue' },
                { title: '今日會話', value: stats.chats?.today_sessions || 0, icon: 'fas fa-comment-dots', color: 'green' },
                { title: '總訊息數', value: stats.chats?.total_messages || 0, icon: 'fas fa-envelope', color: 'purple' },
                { title: '今日訊息', value: stats.chats?.today_messages || 0, icon: 'fas fa-paper-plane', color: 'orange' }
            ];
            
            container.innerHTML = cards.map(card => `
                <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
                    <div class="flex items-center justify-between">
                        <div>
                            <p class="text-sm text-gray-600 font-medium">${card.title}</p>
                            <p class="text-2xl font-bold text-gray-900">${card.value.toLocaleString()}</p>
                        </div>
                        <div class="text-${card.color}-600">
                            <i class="${card.icon} text-2xl"></i>
                        </div>
                    </div>
                </div>
            `).join('');
        },

        async loadFilterOptions() {
            try {
                const [usersResponse, charactersResponse] = await Promise.all([
                    API.admin.getUsers({ page: 1, limit: 100 }),
                    API.get('/character/list')
                ]);
                
                if (usersResponse.success) {
                    this.renderUserFilter(usersResponse.data.users);
                }
                
                if (charactersResponse.success) {
                    this.renderCharacterFilter(charactersResponse.data);
                }
            } catch (error) {
                console.error('載入篩選選項失敗:', error);
            }
        },

        renderUserFilter(users) {
            const select = document.getElementById('userFilter');
            if (!select || !users) return;
            
            const options = users.map(user => 
                `<option value="${user.id}">${user.display_name || user.username || user.id}</option>`
            ).join('');
            
            select.innerHTML = '<option value="">所有用戶</option>' + options;
        },

        renderCharacterFilter(charactersData) {
            const select = document.getElementById('characterFilter');
            if (!select || !charactersData) return;
            
            // 處理不同的數據結構 - characters 可能在 data 或 characters 字段中，或直接是陣列
            let characters = charactersData;
            if (charactersData.characters && Array.isArray(charactersData.characters)) {
                characters = charactersData.characters;
            } else if (charactersData.data && Array.isArray(charactersData.data)) {
                characters = charactersData.data;
            } else if (!Array.isArray(charactersData)) {
                console.warn('Characters data is not an array:', charactersData);
                return;
            }
            
            const options = characters.map(character => 
                `<option value="${character.id}">${character.name}</option>`
            ).join('');
            
            select.innerHTML = '<option value="">所有角色</option>' + options;
        },

        async loadChats() {
            Utils.showById('chatsTableLoading');
            Utils.hideById('chatsTableEmpty');

            try {
                const params = {
                    page: this.currentPage,
                    limit: this.pageSize,
                    sort_by: this.sortBy,
                    sort_order: this.sortOrder
                };

                // 只添加非空的篩選參數
                if (this.filters.search) params.query = this.filters.search;
                if (this.filters.user) params.user_id = this.filters.user;
                if (this.filters.character) params.character_id = this.filters.character;
                if (this.filters.dateFrom) params.date_from = this.filters.dateFrom;
                if (this.filters.dateTo) params.date_to = this.filters.dateTo;

                console.log('🔍 聊天搜尋參數:', params);
                
                const response = await API.admin.getChats(params);
                if (response.success) {
                    this.currentData = response.data;
                    this.renderChatsTable(response.data);
                    this.renderPagination(response.pagination);
                } else {
                    this.showChatsError(response.message || '聊天記錄載入失敗');
                }
            } catch (error) {
                console.error('載入聊天記錄失敗:', error);
                this.showChatsError('聊天記錄載入失敗');
            } finally {
                Utils.hideById('chatsTableLoading');
            }
        },

        renderChatsTable(data) {
            const tbody = document.getElementById('chatsTableBody');
            if (!tbody || !data.chats) return;
            
            if (data.chats.length === 0) {
                Utils.showById('chatsTableEmpty');
                tbody.innerHTML = '';
                return;
            }
            
            tbody.innerHTML = data.chats.map(chat => {
                // Debug logging to check data structure
                console.log('Chat data:', chat);
                
                const relationship = chat.relationship || {};
                const trustLevel = relationship.trust_level || 0;
                const affectionLevel = relationship.affection_level || 0;
                const relationshipStage = relationship.relationship_stage || '初次見面';
                
                // 關係狀態顏色和顯示文字 - 與 AI prompt 中定義的關係狀態一致
                // AI 定義: stranger, friend, close_friend, lover, soulmate
                const getRelationshipDisplay = (stage) => {
                    switch(stage) {
                        case 'soulmate':
                        case '靈魂伴侶':
                            return { text: '靈魂伴侶', color: 'bg-purple-100 text-purple-800' };
                        case 'lover':
                        case '戀人':
                            return { text: '戀人', color: 'bg-pink-100 text-pink-800' };
                        case 'close_friend':  
                        case '親密朋友':
                            return { text: '親密朋友', color: 'bg-blue-100 text-blue-800' };
                        case 'friend':
                        case '朋友':
                            return { text: '朋友', color: 'bg-green-100 text-green-800' };
                        case 'stranger':
                        case '陌生人':
                        case '初次見面':
                        default:
                            return { text: '初次見面', color: 'bg-gray-100 text-gray-800' };
                    }
                };
                
                return `
                <tr class="hover:bg-gray-50">
                    <td class="px-6 py-4">
                        <div class="text-sm font-medium text-gray-900">#${chat.id}</div>
                        <div class="text-sm text-gray-500">${chat.title || '未命名會話'}</div>
                    </td>
                    <td class="px-6 py-4">
                        <div class="text-sm text-gray-900">${chat.user ? (chat.user.display_name || chat.user.username || '未知用戶') : '未知用戶'}</div>
                    </td>
                    <td class="px-6 py-4">
                        <div class="text-sm text-gray-900">${chat.character_name || '未知角色'}</div>
                    </td>
                    <td class="px-6 py-4">
                        <div class="space-y-1">
                            <span class="px-2 py-1 text-xs font-semibold rounded-full ${getRelationshipDisplay(relationshipStage).color}">
                                ${getRelationshipDisplay(relationshipStage).text}
                            </span>
                            <div class="text-xs text-gray-500">
                                信任: ${trustLevel}/100 | 好感: ${affectionLevel}/100
                            </div>
                        </div>
                    </td>
                    <td class="px-6 py-4">
                        <span class="px-2 py-1 text-xs font-semibold rounded-full ${chat.status === 'active' ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'}">
                            ${chat.status === 'active' ? '進行中' : '已結束'}
                        </span>
                    </td>
                    <td class="px-6 py-4 text-sm text-gray-900">
                        ${Utils.formatDate(chat.created_at)}
                    </td>
                    <td class="px-6 py-4 text-sm space-x-2">
                        <button onclick="AdminPages.chats.viewChatHistory('${chat.id}')" class="text-blue-600 hover:text-blue-800">
                            <i class="fas fa-history"></i> 記錄
                        </button>
                        <button onclick="AdminPages.chats.exportChat('${chat.id}')" class="text-green-600 hover:text-green-800">
                            <i class="fas fa-download"></i> 匯出
                        </button>
                    </td>
                </tr>
                `;
            }).join('');
        },

        renderPagination(pagination) {
            const container = document.getElementById('chatsPagination');
            if (!container || !pagination) return;
            
            container.innerHTML = AdminPages.common.createPagination(
                pagination.current_page,
                pagination.total_pages,
                'AdminPages.chats.goToPage'
            );
        },

        showChatsError(message) {
            const tbody = document.getElementById('chatsTableBody');
            if (tbody) {
                tbody.innerHTML = `
                    <tr>
                        <td colspan="6" class="px-6 py-8 text-center">
                            <i class="fas fa-exclamation-triangle text-2xl text-red-600 mb-2"></i>
                            <p class="text-red-600">${message}</p>
                        </td>
                    </tr>
                `;
            }
        },

        async goToPage(page) {
            this.currentPage = page;
            await this.loadChats();
        },

        applyFilters() {
            this.filters.user = document.getElementById('userFilter')?.value || '';
            this.filters.character = document.getElementById('characterFilter')?.value || '';
            this.filters.dateFrom = document.getElementById('dateFromFilter')?.value || '';
            this.filters.dateTo = document.getElementById('dateToFilter')?.value || '';
            
            this.currentPage = 1;
            this.loadChats();
        },

        // 聊天記錄分頁狀態
        chatHistoryState: {
            currentChatId: null,
            currentPage: 1,
            pageSize: 10,
            totalPages: 1,
            totalMessages: 0,
            allMessages: [],
            sessionInfo: {}
        },

        async viewChatHistory(chatId) {
            try {
                const response = await API.admin.getChatHistory(chatId);
                if (response.success) {
                    // 初始化分頁狀態
                    this.chatHistoryState = {
                        currentChatId: chatId,
                        currentPage: 1,
                        pageSize: 10,
                        totalPages: Math.ceil((response.data.messages || []).length / 10),
                        totalMessages: (response.data.messages || []).length,
                        allMessages: response.data.messages || [],
                        sessionInfo: response.data.session_info || {}
                    };
                    
                    this.showChatHistoryModal();
                } else {
                    AdminPages.common.showAlert('聊天記錄載入失敗', 'error');
                }
            } catch (error) {
                console.error('載入聊天記錄失敗:', error);
                AdminPages.common.showAlert('聊天記錄載入失敗', 'error');
            }
        },

        // 切換聊天記錄頁面
        goToChatHistoryPage(page) {
            if (page >= 1 && page <= this.chatHistoryState.totalPages) {
                this.chatHistoryState.currentPage = page;
                this.showChatHistoryModal();
            }
        },

        showChatHistoryModal() {
            const content = document.getElementById('chatHistoryContent');
            if (!content || !this.chatHistoryState.allMessages.length) return;
            
            const state = this.chatHistoryState;
            const sessionInfo = state.sessionInfo;
            const user = sessionInfo.user || {};
            const character = sessionInfo.character || {};
            const sessionId = sessionInfo.id || state.currentChatId || '未知';
            
            // 計算分頁
            const startIndex = (state.currentPage - 1) * state.pageSize;
            const endIndex = startIndex + state.pageSize;
            const currentMessages = state.allMessages.slice(startIndex, endIndex);
            
            // 計算聊天統計信息
            const totalMessages = state.allMessages.length;
            const userMessages = state.allMessages.filter(m => m.role === 'user').length;
            const assistantMessages = state.allMessages.filter(m => m.role === 'assistant').length;
            const totalWords = state.allMessages.reduce((sum, msg) => {
                const text = msg.dialogue || msg.content || '';
                return sum + text.length;
            }, 0);
            
            content.innerHTML = `
                <!-- 聊天會話信息 -->
                <div class="bg-gradient-to-r from-blue-50 to-purple-50 rounded-lg p-4 mb-4 border">
                    <h4 class="text-sm font-medium text-gray-900 mb-3 flex items-center">
                        <i class="fas fa-comments mr-2 text-blue-600"></i>聊天會話詳情
                    </h4>
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div class="bg-white rounded-lg p-3 border">
                            <div class="flex items-center mb-2">
                                <i class="fas fa-user mr-2 text-blue-600"></i>
                                <span class="font-medium text-gray-900">用戶信息</span>
                            </div>
                            <div class="text-sm space-y-1">
                                <div><strong>用戶名:</strong> ${user.username || '未知'}</div>
                                <div><strong>顯示名:</strong> ${user.display_name || '未設定'}</div>
                                <div><strong>ID:</strong> <code class="bg-gray-100 px-1 rounded text-xs">${user.id || '未知'}</code></div>
                            </div>
                        </div>
                        <div class="bg-white rounded-lg p-3 border">
                            <div class="flex items-center mb-2">
                                <i class="fas fa-robot mr-2 text-purple-600"></i>
                                <span class="font-medium text-gray-900">角色信息</span>
                            </div>
                            <div class="text-sm space-y-1">
                                <div class="flex items-center">
                                    ${character.avatar_url ? `<img src="${character.avatar_url}" class="w-6 h-6 rounded-full mr-2" alt="${character.name}">` : '<i class="fas fa-user-circle text-gray-400 mr-2"></i>'}
                                    <strong>${character.name || '未知角色'}</strong>
                                </div>
                                <div><strong>ID:</strong> <code class="bg-gray-100 px-1 rounded text-xs">${character.id || '未知'}</code></div>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- 聊天統計信息 -->
                <div class="bg-gray-50 rounded-lg p-4 mb-4 border">
                    <h4 class="text-sm font-medium text-gray-900 mb-2 flex items-center">
                        <i class="fas fa-chart-bar mr-2 text-orange-600"></i>聊天統計
                    </h4>
                    <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-xs">
                        <div class="text-center bg-white rounded-lg p-3 border">
                            <div class="font-semibold text-blue-600 text-lg">${totalMessages}</div>
                            <div class="text-gray-500">總消息數</div>
                        </div>
                        <div class="text-center bg-white rounded-lg p-3 border">
                            <div class="font-semibold text-green-600 text-lg">${userMessages}</div>
                            <div class="text-gray-500">用戶消息</div>
                        </div>
                        <div class="text-center bg-white rounded-lg p-3 border">
                            <div class="font-semibold text-purple-600 text-lg">${assistantMessages}</div>
                            <div class="text-gray-500">AI回覆</div>
                        </div>
                        <div class="text-center bg-white rounded-lg p-3 border">
                            <div class="font-semibold text-orange-600 text-lg">${totalWords}</div>
                            <div class="text-gray-500">總字數</div>
                        </div>
                    </div>
                </div>
                
                <!-- 分頁信息和控件 -->
                <div class="flex justify-between items-center mb-4 p-3 bg-gray-50 rounded-lg">
                    <div class="text-sm text-gray-600 flex items-center">
                        <i class="fas fa-list mr-2 text-green-600"></i>
                        消息記錄 (第 ${state.currentPage} 頁，共 ${state.totalPages} 頁) 
                        <span class="ml-2 text-xs bg-blue-100 text-blue-800 px-2 py-1 rounded">
                            顯示 ${startIndex + 1}-${Math.min(endIndex, totalMessages)} 條，共 ${totalMessages} 條消息
                        </span>
                    </div>
                    <div class="flex items-center space-x-2">
                        <button onclick="AdminPages.chats.goToChatHistoryPage(${state.currentPage - 1})" 
                                class="px-2 py-1 text-xs border rounded hover:bg-gray-100 ${state.currentPage <= 1 ? 'opacity-50 cursor-not-allowed' : ''}"
                                ${state.currentPage <= 1 ? 'disabled' : ''}>
                            <i class="fas fa-chevron-left"></i> 上一頁
                        </button>
                        <span class="text-xs text-gray-500">${state.currentPage}/${state.totalPages}</span>
                        <button onclick="AdminPages.chats.goToChatHistoryPage(${state.currentPage + 1})" 
                                class="px-2 py-1 text-xs border rounded hover:bg-gray-100 ${state.currentPage >= state.totalPages ? 'opacity-50 cursor-not-allowed' : ''}"
                                ${state.currentPage >= state.totalPages ? 'disabled' : ''}>
                            下一頁 <i class="fas fa-chevron-right"></i>
                        </button>
                    </div>
                </div>

                <!-- 消息列表 -->
                <div class="space-y-4">
                    ${currentMessages.map((msg, index) => {
                        const actualIndex = startIndex + index; // 實際消息索引
                        const messageText = msg.dialogue || msg.content || '無內容';
                        const wordCount = messageText.length;
                        const isUser = msg.role === 'user';
                        
                        return `
                        <div class="flex ${isUser ? 'justify-end' : 'justify-start'} mb-4">
                            <div class="max-w-4xl ${isUser ? 'ml-8' : 'mr-8'}">
                                <!-- 發送者標籤 -->
                                <div class="flex ${isUser ? 'justify-end' : 'justify-start'} items-center mb-1">
                                    <div class="flex items-center space-x-2">
                                        <span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                                            isUser 
                                                ? 'bg-blue-100 text-blue-800' 
                                                : 'bg-purple-100 text-purple-800'
                                        }">
                                            <i class="fas ${isUser ? 'fa-user' : 'fa-robot'} mr-1"></i>
                                            ${isUser ? (user.display_name || user.username || '用戶') : (character.name || 'AI助手')}
                                        </span>
                                        <span class="text-xs text-gray-500">#${actualIndex + 1}</span>
                                        <span class="text-xs text-gray-400">
                                            <i class="fas fa-clock mr-1"></i>
                                            ${Utils.formatDate(msg.created_at)}
                                        </span>
                                        ${msg.nsfw_level ? `
                                            <span class="px-2 py-1 text-xs rounded-full ${
                                                msg.nsfw_level >= 4 ? 'bg-red-100 text-red-800' : 
                                                msg.nsfw_level >= 3 ? 'bg-yellow-100 text-yellow-800' : 
                                                'bg-green-100 text-green-800'
                                            }">
                                                <i class="fas fa-shield-alt mr-1"></i>L${msg.nsfw_level}
                                            </span>
                                        ` : ''}
                                    </div>
                                </div>
                                
                                <!-- 消息氣泡 -->
                                <div class="relative">
                                    <div class="rounded-2xl px-4 py-3 shadow-sm ${
                                        isUser 
                                            ? 'bg-blue-600 text-white' 
                                            : 'bg-white border border-gray-200 text-gray-900'
                                    }">
                                        <div class="text-sm leading-relaxed whitespace-pre-wrap">${messageText}</div>
                                    </div>
                                    
                                    <!-- 氣泡箭頭 -->
                                    <div class="absolute top-3 ${isUser ? 'right-0 translate-x-2' : 'left-0 -translate-x-2'}">
                                        <div class="w-3 h-3 rotate-45 ${
                                            isUser 
                                                ? 'bg-blue-600' 
                                                : 'bg-white border-l border-t border-gray-200'
                                        }"></div>
                                    </div>
                                </div>
                                
                                <!-- 技術資訊 -->
                                <div class="flex ${isUser ? 'justify-end' : 'justify-start'} mt-2">
                                    <div class="flex flex-wrap items-center gap-2 text-xs text-gray-400 bg-gray-50 px-3 py-1 rounded-full">
                                        <span><i class="fas fa-font"></i> ${wordCount}字</span>
                                        ${msg.ai_engine ? `<span><i class="fas fa-brain"></i> ${msg.ai_engine}</span>` : ''}
                                        ${msg.response_time_ms ? `<span><i class="fas fa-stopwatch"></i> ${msg.response_time_ms}ms</span>` : ''}
                                        ${msg.token_count ? `<span><i class="fas fa-coins"></i> ${msg.token_count}t</span>` : ''}
                                        ${msg.id ? `<span title="消息ID: ${msg.id}"><i class="fas fa-tag"></i> ${msg.id.substring(0, 8)}</span>` : ''}
                                    </div>
                                </div>
                            </div>
                        </div>
                        `;
                    }).join('')}
                </div>
                
                <!-- 底部翻頁控件 -->
                <div class="flex justify-center items-center mt-6 py-4 bg-gray-50 rounded-lg">
                    <div class="flex items-center space-x-3">
                        <button onclick="AdminPages.chats.goToChatHistoryPage(1)" 
                                class="px-3 py-2 text-xs border rounded hover:bg-gray-100 ${state.currentPage <= 1 ? 'opacity-50 cursor-not-allowed' : ''}"
                                ${state.currentPage <= 1 ? 'disabled' : ''}>
                            <i class="fas fa-angle-double-left"></i> 首頁
                        </button>
                        <button onclick="AdminPages.chats.goToChatHistoryPage(${state.currentPage - 1})" 
                                class="px-3 py-2 text-xs border rounded hover:bg-gray-100 ${state.currentPage <= 1 ? 'opacity-50 cursor-not-allowed' : ''}"
                                ${state.currentPage <= 1 ? 'disabled' : ''}>
                            <i class="fas fa-chevron-left"></i> 上一頁
                        </button>
                        
                        <!-- 頁碼顯示 -->
                        <div class="flex items-center space-x-1">
                            ${Array.from({ length: Math.min(5, state.totalPages) }, (_, i) => {
                                let pageNum;
                                if (state.totalPages <= 5) {
                                    pageNum = i + 1;
                                } else if (state.currentPage <= 3) {
                                    pageNum = i + 1;
                                } else if (state.currentPage > state.totalPages - 3) {
                                    pageNum = state.totalPages - 4 + i;
                                } else {
                                    pageNum = state.currentPage - 2 + i;
                                }
                                
                                return `
                                    <button onclick="AdminPages.chats.goToChatHistoryPage(${pageNum})" 
                                            class="px-3 py-2 text-xs rounded ${pageNum === state.currentPage ? 'bg-blue-600 text-white' : 'border hover:bg-gray-100'}">
                                        ${pageNum}
                                    </button>
                                `;
                            }).join('')}
                        </div>
                        
                        <button onclick="AdminPages.chats.goToChatHistoryPage(${state.currentPage + 1})" 
                                class="px-3 py-2 text-xs border rounded hover:bg-gray-100 ${state.currentPage >= state.totalPages ? 'opacity-50 cursor-not-allowed' : ''}"
                                ${state.currentPage >= state.totalPages ? 'disabled' : ''}>
                            下一頁 <i class="fas fa-chevron-right"></i>
                        </button>
                        <button onclick="AdminPages.chats.goToChatHistoryPage(${state.totalPages})" 
                                class="px-3 py-2 text-xs border rounded hover:bg-gray-100 ${state.currentPage >= state.totalPages ? 'opacity-50 cursor-not-allowed' : ''}"
                                ${state.currentPage >= state.totalPages ? 'disabled' : ''}>
                            末頁 <i class="fas fa-angle-double-right"></i>
                        </button>
                    </div>
                </div>
                
                <!-- 會話詳情與導出 -->
                <div class="mt-6 pt-4 border-t border-gray-200">
                    <div class="flex justify-between items-center">
                        <div class="text-xs text-gray-500 space-y-1">
                            <div><i class="fas fa-info-circle mr-1"></i>會話ID: <code class="bg-gray-100 px-1 rounded">${sessionId}</code></div>
                            ${sessionInfo.title ? `<div><i class="fas fa-heading mr-1"></i>會話標題: ${sessionInfo.title}</div>` : ''}
                            ${sessionInfo.created_at ? `<div><i class="fas fa-calendar mr-1"></i>創建時間: ${Utils.formatDate(sessionInfo.created_at)}</div>` : ''}
                        </div>
                        <button onclick="AdminPages.chats.exportChat('${sessionId}')" 
                                class="px-3 py-2 bg-green-600 text-white text-xs rounded hover:bg-green-700 transition-colors flex items-center">
                            <i class="fas fa-download mr-2"></i>導出聊天記錄
                        </button>
                    </div>
                </div>
            `;
            
            Utils.showById('chatHistoryModal');
        },

        async exportChat(chatId) {
            try {
                const response = await API.admin.exportChat(chatId);
                if (response.success) {
                    // 創建下載連結
                    const blob = new Blob([JSON.stringify(response.data, null, 2)], {
                        type: 'application/json'
                    });
                    const url = URL.createObjectURL(blob);
                    const a = document.createElement('a');
                    a.href = url;
                    a.download = `chat_${chatId}_${new Date().toISOString().split('T')[0]}.json`;
                    document.body.appendChild(a);
                    a.click();
                    document.body.removeChild(a);
                    URL.revokeObjectURL(url);
                    
                    AdminPages.common.showAlert('聊天記錄匯出成功');
                } else {
                    AdminPages.common.showAlert('匯出失敗', 'error');
                }
            } catch (error) {
                console.error('匯出聊天記錄失敗:', error);
                AdminPages.common.showAlert('匯出失敗', 'error');
            }
        },

        initSorting() {
            const sortButtons = document.querySelectorAll('#chatsTable th button');
            sortButtons.forEach((button, index) => {
                const fieldMap = ['title', 'username', 'character_name', 'created_at']; // 對應表格列的字段
                const field = fieldMap[index];
                if (field) {
                    button.addEventListener('click', () => {
                        this.sortByField(field);
                    });
                }
            });
        },

        async sortByField(field) {
            if (this.sortBy === field) {
                this.sortOrder = this.sortOrder === 'asc' ? 'desc' : 'asc';
            } else {
                this.sortBy = field;
                this.sortOrder = 'desc';
            }

            this.currentPage = 1;
            await this.loadChats();
        }
    },

    // 角色管理頁面模組
    characters: {
        currentPage: 1,
        limit: 20,
        sortBy: 'updated_at',
        sortOrder: 'desc',
        filters: {
            query: '',
            type: 'all',
            status: 'all',
            includeDeleted: false
        },

        init() {
            console.log('🎭 初始化角色管理頁面');
            this.initEventListeners();
            this.initSorting();
            this.loadCharacters();
        },

        initEventListeners() {
            // 搜索輸入框
            const searchInput = document.getElementById('characterSearchInput');
            if (searchInput) {
                searchInput.addEventListener('input', Utils.debounce((e) => {
                    this.filters.query = e.target.value;
                    this.currentPage = 1;
                    this.loadCharacters();
                }, 500));
            }

            // 類型篩選
            const typeFilter = document.getElementById('characterTypeFilter');
            if (typeFilter) {
                typeFilter.addEventListener('change', (e) => {
                    this.filters.type = e.target.value;
                    this.currentPage = 1;
                    this.loadCharacters();
                });
            }

            // 狀態篩選
            const activeFilter = document.getElementById('characterActiveFilter');
            if (activeFilter) {
                activeFilter.addEventListener('change', (e) => {
                    this.filters.status = e.target.value;
                    this.currentPage = 1;
                    this.loadCharacters();
                });
            }

            // 已刪除角色篩選
            const deletedFilter = document.getElementById('characterDeletedFilter');
            if (deletedFilter) {
                deletedFilter.addEventListener('change', (e) => {
                    this.filters.includeDeleted = e.target.value === 'true';
                    this.currentPage = 1;
                    this.loadCharacters();
                });
            }
        },

        async loadCharacters() {
            try {
                AdminPages.common.showLoading('charactersTableBody');
                
                const params = {
                    page: this.currentPage,
                    limit: this.limit,
                    sort_by: this.sortBy,
                    sort_order: this.sortOrder,
                    status: this.filters.status,
                    type: this.filters.type,
                    include_deleted: this.filters.includeDeleted
                };

                // 移除空值參數
                Object.keys(params).forEach(key => {
                    if (params[key] === '' || params[key] === null || params[key] === undefined) {
                        delete params[key];
                    }
                });

                const response = await API.get('/admin/characters', { params });
                
                if (response.success) {
                    this.renderCharacters(response.data.characters);
                    this.renderPagination(response.data.pagination);
                    this.renderStats(response.data.stats);
                } else {
                    this.showCharactersError('載入角色列表失敗');
                }
            } catch (error) {
                console.error('載入角色列表失敗:', error);
                this.showCharactersError('載入角色列表時發生錯誤');
            }
        },

        renderStats(stats) {
            const statsGrid = document.getElementById('characterStatsGrid');
            if (!statsGrid || !stats) return;

            const cards = [
                { title: '總角色數', value: stats.total || 0, icon: 'fas fa-user-friends', color: 'blue' },
                { title: '活躍角色', value: stats.active || 0, icon: 'fas fa-user-check', color: 'green' },
                { title: '已停用', value: stats.inactive || 0, icon: 'fas fa-user-times', color: 'red' },
                { title: '今日新增', value: stats.today || 0, icon: 'fas fa-plus-circle', color: 'purple' }
            ];

            statsGrid.innerHTML = cards.map(card => `
                <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
                    <div class="flex items-center">
                        <div class="p-2 bg-${card.color}-100 rounded-md">
                            <i class="${card.icon} text-${card.color}-600"></i>
                        </div>
                        <div class="ml-4">
                            <p class="text-sm font-medium text-gray-600">${card.title}</p>
                            <p class="text-2xl font-bold text-gray-900">${card.value}</p>
                        </div>
                    </div>
                </div>
            `).join('');
        },

        renderCharacters(characters) {
            const tbody = document.getElementById('charactersTableBody');
            if (!tbody) return;

            if (!characters || characters.length === 0) {
                tbody.innerHTML = `
                    <tr>
                        <td colspan="7" class="px-6 py-8 text-center">
                            <i class="fas fa-user-friends text-4xl text-gray-300 mb-4"></i>
                            <p class="text-gray-600">沒有找到角色資料</p>
                        </td>
                    </tr>
                `;
                return;
            }

            tbody.innerHTML = characters.map(char => `
                <tr class="hover:bg-gray-50 ${char.deleted_at ? 'bg-red-50 opacity-75' : ''}">
                    <td class="px-6 py-4">
                        <div class="flex items-center">
                            <div class="h-10 w-10 flex-shrink-0">
                                <img class="h-10 w-10 rounded-full object-cover" 
                                     src="${char.avatar_url || '/public/default-avatar.png'}" 
                                     alt="${char.name}">
                            </div>
                            <div class="ml-4">
                                <div class="text-sm font-medium text-gray-900">
                                    ${char.name}
                                    ${char.is_system ? '<span class="ml-2 inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-blue-100 text-blue-800">系統</span>' : ''}
                                    ${char.deleted_at ? '<span class="ml-2 inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-red-100 text-red-800">已刪除</span>' : ''}
                                </div>
                                <div class="text-sm text-gray-500">ID: ${char.id}</div>
                            </div>
                        </div>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap">
                        <span class="inline-flex px-2 py-1 text-xs font-semibold rounded-full ${char.is_system ? 'bg-blue-100 text-blue-800' : 'bg-gray-100 text-gray-800'}">
                            ${char.is_system ? '系統角色' : '用戶角色'}
                        </span>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap">
                        <span class="inline-flex px-2 py-1 text-xs font-semibold rounded-full ${char.is_active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}">
                            ${char.is_active ? '活躍' : '已停用'}
                        </span>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        <div>
                            ${char.created_by_name || '系統'}
                            ${char.created_by_display_name ? `<div class="text-xs text-gray-500">${char.created_by_display_name}</div>` : ''}
                        </div>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        ${char.popularity || 0}
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        <div class="text-xs">
                            ${new Date(char.updated_at).toLocaleDateString('zh-TW')}
                            <div>${new Date(char.updated_at).toLocaleTimeString('zh-TW')}</div>
                        </div>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                        <div class="flex space-x-2">
                            <button onclick="AdminPages.characters.viewCharacter('${char.id}')" class="text-blue-600 hover:text-blue-800">
                                <i class="fas fa-eye"></i>
                            </button>
                            ${!char.deleted_at ? `
                                <button onclick="AdminPages.characters.editCharacter('${char.id}')" class="text-green-600 hover:text-green-800">
                                    <i class="fas fa-edit"></i>
                                </button>
                                <button onclick="AdminPages.characters.toggleCharacterStatus('${char.id}', ${char.is_active})" 
                                        class="text-${char.is_active ? 'red' : 'green'}-600 hover:text-${char.is_active ? 'red' : 'green'}-800">
                                    <i class="fas fa-${char.is_active ? 'pause' : 'play'}"></i>
                                </button>
                            ` : `
                                <button onclick="AdminPages.characters.restoreCharacter('${char.id}')" class="text-green-600 hover:text-green-800">
                                    <i class="fas fa-undo"></i>
                                </button>
                                <button onclick="AdminPages.characters.permanentDeleteCharacter('${char.id}')" class="text-red-600 hover:text-red-800">
                                    <i class="fas fa-trash-alt"></i>
                                </button>
                            `}
                        </div>
                    </td>
                </tr>
            `).join('');
        },

        renderPagination(pagination) {
            const container = document.getElementById('charactersPagination');
            if (!container || !pagination) return;
            
            container.innerHTML = AdminPages.common.createPagination(
                pagination.current_page,
                pagination.total_pages,
                'AdminPages.characters.goToPage'
            );
        },

        showCharactersError(message) {
            const tbody = document.getElementById('charactersTableBody');
            if (tbody) {
                tbody.innerHTML = `
                    <tr>
                        <td colspan="6" class="px-6 py-8 text-center">
                            <i class="fas fa-exclamation-triangle text-2xl text-red-600 mb-2"></i>
                            <p class="text-red-600">${message}</p>
                        </td>
                    </tr>
                `;
            }
        },

        reload() {
            console.log('🔄 重新載入角色列表');
            this.loadCharacters();
        },

        goToPage(page) {
            this.currentPage = page;
            this.loadCharacters();
        },

        async viewCharacter(characterId) {
            try {
                const response = await API.get(`/admin/characters/${characterId}`);
                if (response.success) {
                    this.showCharacterDetail(response.data);
                } else {
                    AdminPages.common.showAlert('載入角色詳情失敗', 'error');
                }
            } catch (error) {
                console.error('載入角色詳情失敗:', error);
                AdminPages.common.showAlert('載入角色詳情時發生錯誤', 'error');
            }
        },

        showCharacterDetail(character) {
            const content = document.getElementById('characterDetailContent');
            if (!content) return;

            content.innerHTML = `
                <div class="space-y-4">
                    <div class="flex items-start space-x-4">
                        <img src="${character.avatar_url || '/public/default-avatar.png'}" 
                             alt="${character.name}" 
                             class="h-20 w-20 rounded-full object-cover">
                        <div class="flex-1">
                            <h4 class="text-lg font-medium text-gray-900">${character.name}</h4>
                            <p class="text-sm text-gray-600">ID: ${character.id}</p>
                            <div class="mt-2 space-x-2">
                                <span class="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-blue-100 text-blue-800">
                                    ${character.type}
                                </span>
                                <span class="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-gray-100 text-gray-800">
                                    ${character.locale || 'zh-TW'}
                                </span>
                                <span class="inline-flex px-2 py-1 text-xs font-semibold rounded-full ${character.is_active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}">
                                    ${character.is_active ? '活躍' : '已停用'}
                                </span>
                            </div>
                        </div>
                    </div>
                    
                    <div class="border-t pt-4">
                        <dl class="grid grid-cols-1 gap-4 sm:grid-cols-2">
                            <div>
                                <dt class="text-sm font-medium text-gray-500">人氣度</dt>
                                <dd class="mt-1 text-sm text-gray-900">${character.popularity || 0}</dd>
                            </div>
                            <div>
                                <dt class="text-sm font-medium text-gray-500">創建時間</dt>
                                <dd class="mt-1 text-sm text-gray-900">${Utils.formatDate(character.created_at)}</dd>
                            </div>
                            <div>
                                <dt class="text-sm font-medium text-gray-500">更新時間</dt>
                                <dd class="mt-1 text-sm text-gray-900">${Utils.formatDate(character.updated_at)}</dd>
                            </div>
                            <div>
                                <dt class="text-sm font-medium text-gray-500">創建者</dt>
                                <dd class="mt-1 text-sm text-gray-900">${character.created_by_name || '系統'}</dd>
                            </div>
                            <div>
                                <dt class="text-sm font-medium text-gray-500">最後編輯者</dt>
                                <dd class="mt-1 text-sm text-gray-900">${character.updated_by_name || '系統'}</dd>
                            </div>
                            ${character.deleted_at ? `
                            <div>
                                <dt class="text-sm font-medium text-gray-500">刪除時間</dt>
                                <dd class="mt-1 text-sm text-red-600">${Utils.formatDate(character.deleted_at)}</dd>
                            </div>
                            ` : ''}
                        </dl>
                    </div>
                    
                    ${character.user_description ? `
                        <div class="border-t pt-4">
                            <dt class="text-sm font-medium text-gray-500">用戶描述</dt>
                            <dd class="mt-1 text-sm text-gray-900">${character.user_description.substring(0, 500)}${character.user_description.length > 500 ? '...' : ''}</dd>
                        </div>
                    ` : ''}
                    
                    ${character.tags && character.tags.length > 0 ? `
                        <div class="border-t pt-4">
                            <dt class="text-sm font-medium text-gray-500">標籤</dt>
                            <dd class="mt-1">
                                <div class="flex flex-wrap gap-1">
                                    ${character.tags.map(tag => `
                                        <span class="inline-flex px-2 py-1 text-xs font-medium rounded bg-gray-100 text-gray-800">
                                            ${tag}
                                        </span>
                                    `).join('')}
                                </div>
                            </dd>
                        </div>
                    ` : ''}
                </div>
            `;
            
            Utils.showById('characterDetailModal');
        },

        async editCharacter(characterId) {
            try {
                // 獲取角色詳情
                const response = await API.get(`/admin/characters/${characterId}`);
                if (!response.success) {
                    throw new Error(response.error?.message || '無法獲取角色詳情');
                }

                const character = response.data;
                
                // 構建編輯表單
                const formContent = `
                    <div class="space-y-4">
                        <!-- 基本資訊 -->
                        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                            <div>
                                <label class="block text-sm font-medium text-gray-700 mb-1">角色名稱 <span class="text-red-500">*</span></label>
                                <input type="text" id="editCharacterName" value="${Utils.escapeHtml(character.name)}" 
                                       class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                       required maxlength="50">
                            </div>
                            <div>
                                <label class="block text-sm font-medium text-gray-700 mb-1">角色類型</label>
                                <select id="editCharacterType" class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500">
                                    <option value="dominant" ${character.type === 'dominant' ? 'selected' : ''}>霸道型 (Dominant)</option>
                                    <option value="gentle" ${character.type === 'gentle' ? 'selected' : ''}>溫柔型 (Gentle)</option>
                                    <option value="playful" ${character.type === 'playful' ? 'selected' : ''}>活潑型 (Playful)</option>
                                    <option value="mystery" ${character.type === 'mystery' ? 'selected' : ''}>神秘型 (Mystery)</option>
                                    <option value="reliable" ${character.type === 'reliable' ? 'selected' : ''}>可靠型 (Reliable)</option>
                                </select>
                            </div>
                        </div>

                        <!-- 頭像和狀態 -->
                        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                            <div>
                                <label class="block text-sm font-medium text-gray-700 mb-1">頭像 URL</label>
                                <input type="url" id="editCharacterAvatar" value="${character.avatar_url || ''}"
                                       class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                      placeholder="https://www.gravatar.com/avatar/?d=mp">
                            </div>
                            <div>
                                <label class="block text-sm font-medium text-gray-700 mb-1">人氣度</label>
                                <input type="number" id="editCharacterPopularity" value="${character.popularity || 0}" 
                                       class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                       min="0" max="100">
                            </div>
                        </div>

                        <!-- 標籤 -->
                        <div>
                            <label class="block text-sm font-medium text-gray-700 mb-1">標籤</label>
                            <div class="flex flex-wrap gap-2 mb-2" id="editCharacterTagsDisplay">
                                ${(character.tags || []).map(tag => 
                                    `<span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                                        ${Utils.escapeHtml(tag)}
                                        <button type="button" onclick="this.parentElement.remove()" class="ml-1 text-blue-600 hover:text-blue-800">
                                            <i class="fas fa-times text-xs"></i>
                                        </button>
                                    </span>`
                                ).join('')}
                            </div>
                            <div class="flex">
                                <input type="text" id="editCharacterNewTag" placeholder="新增標籤" 
                                       class="flex-1 px-3 py-2 border border-gray-300 rounded-l-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                       onkeypress="if(event.key==='Enter'){event.preventDefault();AdminPages.characters.addTag();}">
                                <button type="button" onclick="AdminPages.characters.addTag()" 
                                        class="px-4 py-2 bg-blue-600 text-white rounded-r-md hover:bg-blue-700">
                                    <i class="fas fa-plus"></i>
                                </button>
                            </div>
                        </div>

                        <!-- 描述 -->
                        <div>
                            <label class="block text-sm font-medium text-gray-700 mb-1">角色描述</label>
                            <textarea id="editCharacterDescription" rows="4" 
                                      class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                      placeholder="描述這個角色的特點...">${character.user_description || ''}</textarea>
                        </div>

                        <!-- 狀態設定 -->
                        <div class="flex items-center space-x-4">
                            <label class="flex items-center">
                                <input type="checkbox" id="editCharacterActive" ${character.is_active ? 'checked' : ''}
                                       class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded">
                                <span class="ml-2 text-sm text-gray-700">啟用角色</span>
                            </label>
                            <label class="flex items-center">
                                <input type="checkbox" id="editCharacterPublic" ${character.is_public ? 'checked' : ''}
                                       class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded">
                                <span class="ml-2 text-sm text-gray-700">公開角色</span>
                            </label>
                        </div>
                    </div>
                `;

                // 顯示編輯表單
                document.getElementById('characterEditContent').innerHTML = formContent;
                
                // 設定表單提交處理器
                const form = document.getElementById('characterEditForm');
                form.onsubmit = (e) => {
                    e.preventDefault();
                    this.saveCharacterEdits(characterId);
                };
                
                Utils.showById('characterEditModal');
                
            } catch (error) {
                console.error('編輯角色失敗:', error);
                AdminPages.common.showAlert('無法載入角色編輯表單: ' + error.message, 'error');
            }
        },

        // 添加標籤功能
        addTag() {
            const newTagInput = document.getElementById('editCharacterNewTag');
            const tagValue = newTagInput.value.trim();
            
            if (!tagValue) {
                AdminPages.common.showAlert('請輸入標籤內容', 'warning');
                return;
            }
            
            // 檢查是否重複
            const tagsDisplay = document.getElementById('editCharacterTagsDisplay');
            const existingTags = Array.from(tagsDisplay.querySelectorAll('span')).map(span => 
                span.textContent.trim().replace('×', '').trim()
            );
            
            if (existingTags.includes(tagValue)) {
                AdminPages.common.showAlert('此標籤已存在', 'warning');
                return;
            }
            
            // 添加新標籤
            const tagElement = document.createElement('span');
            tagElement.className = 'inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800';
            tagElement.innerHTML = `
                ${Utils.escapeHtml(tagValue)}
                <button type="button" onclick="this.parentElement.remove()" class="ml-1 text-blue-600 hover:text-blue-800">
                    <i class="fas fa-times text-xs"></i>
                </button>
            `;
            
            tagsDisplay.appendChild(tagElement);
            newTagInput.value = '';
        },

        // 保存角色編輯
        async saveCharacterEdits(characterId) {
            try {
                // 收集表單資料
                const name = document.getElementById('editCharacterName').value.trim();
                const type = document.getElementById('editCharacterType').value;
                const avatarUrl = document.getElementById('editCharacterAvatar').value.trim();
                const popularity = parseInt(document.getElementById('editCharacterPopularity').value) || 0;
                const description = document.getElementById('editCharacterDescription').value.trim();
                const isActive = document.getElementById('editCharacterActive').checked;
                const isPublic = document.getElementById('editCharacterPublic').checked;
                
                // 收集標籤
                const tagsDisplay = document.getElementById('editCharacterTagsDisplay');
                const tags = Array.from(tagsDisplay.querySelectorAll('span')).map(span => 
                    span.textContent.trim().replace('×', '').trim()
                ).filter(tag => tag.length > 0);

                // 驗證必填欄位
                if (!name) {
                    AdminPages.common.showAlert('角色名稱不能為空', 'error');
                    return;
                }

                // 構建更新資料
                const updateData = {
                    name: name,
                    type: type,
                    is_active: isActive,
                    is_public: isPublic,
                    user_description: description || null,
                    avatar_url: avatarUrl || null,
                    tags: tags,
                    popularity: popularity
                };

                // 發送更新請求
                AdminPages.common.showAlert('正在保存角色資料...', 'info');
                
                const response = await API.put(`/admin/characters/${characterId}`, updateData);
                
                if (!response.success) {
                    throw new Error(response.error?.message || '保存失敗');
                }

                AdminPages.common.showAlert('角色資料已成功更新！', 'success');
                Utils.hideById('characterEditModal');
                
                // 重新載入角色列表
                this.reload();
                
            } catch (error) {
                console.error('保存角色編輯失敗:', error);
                AdminPages.common.showAlert('保存失敗: ' + error.message, 'error');
            }
        },

        async toggleCharacterStatus(characterId, currentStatus) {
            const action = currentStatus ? '停用' : '啟用';
            
            if (!confirm(`確定要${action}這個角色嗎？`)) {
                return;
            }

            try {
                const response = await API.put(`/admin/characters/${characterId}`, {
                    is_active: !currentStatus
                });
                
                if (response.success) {
                    AdminPages.common.showAlert(`角色${action}成功`);
                    this.loadCharacters(); // 重新載入列表
                } else {
                    AdminPages.common.showAlert(`角色${action}失敗`, 'error');
                }
            } catch (error) {
                console.error(`角色${action}失敗:`, error);
                AdminPages.common.showAlert(`角色${action}時發生錯誤`, 'error');
            }
        },

        async restoreCharacter(characterId) {
            if (!confirm('確定要恢復這個角色嗎？')) {
                return;
            }

            try {
                const response = await API.post(`/admin/characters/${characterId}/restore`);
                
                if (response.success) {
                    AdminPages.common.showAlert('角色恢復成功');
                    this.loadCharacters(); // 重新載入列表
                } else {
                    AdminPages.common.showAlert('角色恢復失敗', 'error');
                }
            } catch (error) {
                console.error('角色恢復失敗:', error);
                AdminPages.common.showAlert('角色恢復時發生錯誤', 'error');
            }
        },

        async permanentDeleteCharacter(characterId) {
            if (!confirm('⚠️ 警告：這將永久刪除此角色，無法恢復！\n\n確定要繼續嗎？')) {
                return;
            }

            try {
                const response = await API.delete(`/admin/characters/${characterId}/permanent`);
                
                if (response.success) {
                    AdminPages.common.showAlert('角色已永久刪除');
                    this.loadCharacters(); // 重新載入列表
                } else {
                    AdminPages.common.showAlert('永久刪除失敗', 'error');
                }
            } catch (error) {
                console.error('永久刪除失敗:', error);
                AdminPages.common.showAlert('永久刪除時發生錯誤', 'error');
            }
        },

        initSorting() {
            const sortButtons = document.querySelectorAll('#charactersTable th button');
            sortButtons.forEach((button, index) => {
                const fieldMap = ['name', 'type', 'status', 'creator', 'popularity', 'updated_at']; // 對應表格列的字段
                const field = fieldMap[index];
                if (field) {
                    button.addEventListener('click', () => {
                        this.sortByField(field);
                    });
                }
            });
        },

        async sortByField(field) {
            if (this.sortBy === field) {
                this.sortOrder = this.sortOrder === 'asc' ? 'desc' : 'asc';
            } else {
                this.sortBy = field;
                this.sortOrder = 'desc';
            }

            this.currentPage = 1;
            await this.loadCharacters();
        }
    }
};

// 初始化 API 客戶端
API.init();

// 導出到全局作用域
window.API = API;
window.Utils = Utils;
window.AdminPages = AdminPages;
