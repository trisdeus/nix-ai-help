/**
 * NixAI Enhanced Web Interface JavaScript
 * Provides real-time collaboration, WebSocket communication, and interactive features
 */

class NixAIApp {
    constructor() {
        this.ws = null;
        this.currentUser = null;
        this.currentTeam = null;
        this.notifications = [];
        this.init();
    }

    async init() {
        console.log('🚀 Initializing NixAI Enhanced Web Interface');
        
        // Initialize WebSocket connection
        this.initWebSocket();
        
        // Setup event listeners
        this.setupEventListeners();
        
        // Initialize theme
        this.initTheme();
        
        // Load initial data
        await this.loadInitialData();
        
        // Setup auto-refresh for dashboard
        this.setupAutoRefresh();
        
        console.log('✅ NixAI App initialized successfully');
    }

    initWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/api/ws`;
        
        try {
            this.ws = new WebSocket(wsUrl);
            
            this.ws.onopen = () => {
                console.log('🔗 WebSocket connected');
                this.showNotification('Connected to real-time collaboration', 'success');
            };
            
            this.ws.onmessage = (event) => {
                this.handleWebSocketMessage(event);
            };
            
            this.ws.onclose = () => {
                console.log('📵 WebSocket disconnected');
                this.showNotification('Disconnected from real-time collaboration', 'warning');
                // Attempt to reconnect after 3 seconds
                setTimeout(() => this.initWebSocket(), 3000);
            };
            
            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error);
                this.showNotification('Real-time collaboration error', 'danger');
            };
        } catch (error) {
            console.error('Failed to initialize WebSocket:', error);
        }
    }

    handleWebSocketMessage(event) {
        try {
            const message = JSON.parse(event.data);
            console.log('📨 WebSocket message received:', message);
            
            switch (message.type) {
                case 'config_update':
                    this.handleConfigUpdate(message.data);
                    break;
                case 'team_update':
                    this.handleTeamUpdate(message.data);
                    break;
                case 'fleet_update':
                    this.handleFleetUpdate(message.data);
                    break;
                case 'notification':
                    this.showNotification(message.data.message, message.data.type);
                    break;
                case 'user_activity':
                    this.handleUserActivity(message.data);
                    break;
                default:
                    console.log('Unknown message type:', message.type);
            }
        } catch (error) {
            console.error('Error parsing WebSocket message:', error);
        }
    }

    sendWebSocketMessage(type, data) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({ type, data }));
        }
    }

    setupEventListeners() {
        // Navigation
        document.addEventListener('click', (e) => {
            if (e.target.matches('[data-nav]')) {
                e.preventDefault();
                this.navigate(e.target.getAttribute('data-nav'));
            }
        });

        // Theme toggle
        const themeToggle = document.getElementById('theme-toggle');
        if (themeToggle) {
            themeToggle.addEventListener('click', () => this.toggleTheme());
        }

        // Form submissions
        document.addEventListener('submit', (e) => {
            if (e.target.matches('.nixai-form')) {
                e.preventDefault();
                this.handleFormSubmission(e.target);
            }
        });

        // Modal controls
        document.addEventListener('click', (e) => {
            if (e.target.matches('[data-modal-open]')) {
                e.preventDefault();
                this.openModal(e.target.getAttribute('data-modal-open'));
            }
            if (e.target.matches('[data-modal-close]')) {
                e.preventDefault();
                this.closeModal();
            }
        });

        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            this.handleKeyboardShortcuts(e);
        });
    }

    initTheme() {
        const savedTheme = localStorage.getItem('nixai-theme') || 'light';
        document.documentElement.setAttribute('data-theme', savedTheme);
    }

    toggleTheme() {
        const currentTheme = document.documentElement.getAttribute('data-theme');
        const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
        document.documentElement.setAttribute('data-theme', newTheme);
        localStorage.setItem('nixai-theme', newTheme);
        this.showNotification(`Switched to ${newTheme} theme`, 'info');
    }

    async loadInitialData() {
        try {
            // Load dashboard data
            await this.loadDashboardData();
            
            // Load user info
            await this.loadUserInfo();
            
            // Load active page data
            await this.loadPageData();
        } catch (error) {
            console.error('Error loading initial data:', error);
            this.showNotification('Failed to load some data', 'warning');
        }
    }

    async loadDashboardData() {
        try {
            const response = await fetch('/api/dashboard');
            if (response.ok) {
                const text = await response.text();
                console.log('Dashboard API response (first 100 chars):', text.substring(0, 100));
                try {
                    const data = JSON.parse(text);
                    this.updateDashboard(data.data);
                } catch (jsonError) {
                    console.error('JSON parsing error in dashboard data:', jsonError);
                    console.error('Response text:', text);
                }
            } else {
                console.error('Dashboard API error:', response.status, response.statusText);
                const text = await response.text();
                console.error('Error response:', text);
            }
        } catch (error) {
            console.error('Error loading dashboard data:', error);
        }
    }

    async loadUserInfo() {
        try {
            const response = await fetch('/api/auth/status');
            if (response.ok) {
                const text = await response.text();
                console.log('Auth API response (first 100 chars):', text.substring(0, 100));
                try {
                    const data = JSON.parse(text);
                    this.currentUser = data.data;
                } catch (jsonError) {
                    console.error('JSON parsing error in user info:', jsonError);
                    console.error('Response text:', text);
                }
            } else {
                console.error('Auth API error:', response.status, response.statusText);
                const text = await response.text();
                console.error('Error response:', text);
            }
        } catch (error) {
            console.error('Error loading user info:', error);
        }
    }

    async loadPageData() {
        const path = window.location.pathname;
        
        switch (path) {
            case '/dashboard':
                await this.loadDashboardDetails();
                break;
            case '/fleet':
                await this.loadFleetData();
                break;
            case '/teams':
                await this.loadTeamsData();
                break;
            case '/versions':
                await this.loadVersionsData();
                break;
            case '/builder':
                await this.loadBuilderData();
                break;
        }
    }

    updateDashboard(data) {
        // Update overview stats
        this.updateElement('#total-machines', data.overview?.total_machines || 0);
        this.updateElement('#healthy-machines', data.overview?.healthy_machines || 0);
        this.updateElement('#total-configs', data.overview?.total_configs || 0);
        this.updateElement('#active-teams', data.overview?.active_teams || 0);

        // Update activity feed
        if (data.activities) {
            this.updateActivityFeed(data.activities);
        }

        // Update alerts
        if (data.alerts) {
            this.updateAlerts(data.alerts);
        }
    }

    updateElement(selector, content) {
        const element = document.querySelector(selector);
        if (element) {
            element.textContent = content;
        }
    }

    updateActivityFeed(activities) {
        const feedElement = document.querySelector('#activity-feed');
        if (!feedElement) return;

        feedElement.innerHTML = activities.map(activity => `
            <div class="activity-item fade-in">
                <div class="activity-icon">
                    ${this.getActivityIcon(activity.type)}
                </div>
                <div class="activity-content">
                    <div class="activity-message">${activity.message}</div>
                    <div class="activity-time">${this.formatTime(activity.timestamp)}</div>
                </div>
            </div>
        `).join('');
    }

    updateAlerts(alerts) {
        const alertsElement = document.querySelector('#alerts-list');
        if (!alertsElement) return;

        alertsElement.innerHTML = alerts.map(alert => `
            <div class="nixai-alert nixai-alert-${alert.level} fade-in">
                <strong>${alert.title}</strong>
                <p>${alert.message}</p>
                <small>${this.formatTime(alert.timestamp)}</small>
            </div>
        `).join('');
    }

    getActivityIcon(type) {
        const icons = {
            'config_update': '⚙️',
            'deployment': '🚀',
            'team_join': '👥',
            'build_success': '✅',
            'build_failure': '❌',
            'health_check': '💚'
        };
        return icons[type] || '📝';
    }

    formatTime(timestamp) {
        return new Date(timestamp).toLocaleString();
    }

    showNotification(message, type = 'info', duration = 5000) {
        const notification = {
            id: Date.now(),
            message,
            type,
            timestamp: new Date()
        };

        this.notifications.push(notification);
        this.renderNotification(notification);

        // Auto-remove notification
        setTimeout(() => {
            this.removeNotification(notification.id);
        }, duration);
    }

    renderNotification(notification) {
        const container = this.getNotificationContainer();
        const element = document.createElement('div');
        element.className = `nixai-notification nixai-notification-${notification.type} fade-in`;
        element.setAttribute('data-id', notification.id);
        element.innerHTML = `
            <div class="notification-content">
                <strong>${notification.message}</strong>
                <small>${this.formatTime(notification.timestamp)}</small>
            </div>
            <button class="notification-close" onclick="nixaiApp.removeNotification(${notification.id})">&times;</button>
        `;
        
        container.appendChild(element);
    }

    removeNotification(id) {
        this.notifications = this.notifications.filter(n => n.id !== id);
        const element = document.querySelector(`[data-id="${id}"]`);
        if (element) {
            element.remove();
        }
    }

    getNotificationContainer() {
        let container = document.getElementById('notification-container');
        if (!container) {
            container = document.createElement('div');
            container.id = 'notification-container';
            container.className = 'notification-container';
            container.style.cssText = `
                position: fixed;
                top: 20px;
                right: 20px;
                z-index: 1000;
                display: flex;
                flex-direction: column;
                gap: 10px;
                max-width: 400px;
            `;
            document.body.appendChild(container);
        }
        return container;
    }

    navigate(path) {
        window.history.pushState({}, '', path);
        this.loadPageData();
        this.updateActiveNavigation(path);
    }

    updateActiveNavigation(path) {
        document.querySelectorAll('.nixai-nav-links a').forEach(link => {
            link.classList.remove('active');
            if (link.getAttribute('data-nav') === path) {
                link.classList.add('active');
            }
        });
    }

    async handleFormSubmission(form) {
        const formData = new FormData(form);
        const action = form.getAttribute('action') || form.getAttribute('data-action');
        const method = form.getAttribute('method') || 'POST';

        try {
            this.showLoading(form);
            
            const response = await fetch(action, {
                method: method,
                body: formData
            });

            const responseText = await response.text();
            console.log(`Form ${action} response (first 100 chars):`, responseText.substring(0, 100));
            
            let result;
            try {
                result = JSON.parse(responseText);
            } catch (jsonError) {
                console.error(`JSON parsing error for form ${action}:`, jsonError);
                console.error('Full response text:', responseText);
                this.showNotification(`Invalid response from server: ${jsonError.message}`, 'danger');
                return;
            }

            if (result.success) {
                this.showNotification(result.message || 'Operation completed successfully', 'success');
                
                // Handle successful form submission
                if (form.hasAttribute('data-reload')) {
                    window.location.reload();
                } else if (form.hasAttribute('data-redirect')) {
                    window.location.href = form.getAttribute('data-redirect');
                }
            } else {
                this.showNotification(result.message || 'Operation failed', 'danger');
            }
        } catch (error) {
            console.error('Form submission error:', error);
            this.showNotification('Network error occurred', 'danger');
        } finally {
            this.hideLoading(form);
        }
    }

    showLoading(element) {
        const loadingEl = element.querySelector('.loading') || 
                         element.querySelector('button[type="submit"]');
        if (loadingEl) {
            loadingEl.disabled = true;
            loadingEl.innerHTML = '<span class="nixai-loading"></span> Processing...';
        }
    }

    hideLoading(element) {
        const loadingEl = element.querySelector('.loading') || 
                         element.querySelector('button[type="submit"]');
        if (loadingEl) {
            loadingEl.disabled = false;
            loadingEl.innerHTML = loadingEl.getAttribute('data-original-text') || 'Submit';
        }
    }

    openModal(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.style.display = 'flex';
            modal.classList.add('fade-in');
            document.body.style.overflow = 'hidden';
        }
    }

    closeModal() {
        const modals = document.querySelectorAll('.nixai-modal');
        modals.forEach(modal => {
            modal.style.display = 'none';
            modal.classList.remove('fade-in');
        });
        document.body.style.overflow = '';
    }

    handleKeyboardShortcuts(e) {
        // Ctrl/Cmd + K for search
        if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
            e.preventDefault();
            this.openSearch();
        }
        
        // Escape to close modals
        if (e.key === 'Escape') {
            this.closeModal();
        }
        
        // Ctrl/Cmd + / for help
        if ((e.ctrlKey || e.metaKey) && e.key === '/') {
            e.preventDefault();
            this.openHelp();
        }
    }

    openSearch() {
        // Implementation for global search functionality
        const searchModal = document.getElementById('search-modal');
        if (searchModal) {
            this.openModal('search-modal');
            const searchInput = searchModal.querySelector('input[type="search"]');
            if (searchInput) {
                searchInput.focus();
            }
        }
    }

    openHelp() {
        this.openModal('help-modal');
    }

    setupAutoRefresh() {
        // Auto-refresh dashboard every 30 seconds
        if (window.location.pathname === '/dashboard') {
            setInterval(() => {
                this.loadDashboardData();
            }, 30000);
        }
    }

    // Real-time collaboration handlers
    handleConfigUpdate(data) {
        console.log('Config update received:', data);
        this.showNotification(`Configuration "${data.filename}" was updated by ${data.user}`, 'info');
        
        // Update UI if we're viewing the same config
        if (window.location.pathname.includes('/builder') && 
            document.querySelector(`[data-config="${data.filename}"]`)) {
            this.refreshConfigView();
        }
    }

    handleTeamUpdate(data) {
        console.log('Team update received:', data);
        this.showNotification(`Team "${data.team_name}" was updated`, 'info');
        
        // Refresh teams page if active
        if (window.location.pathname === '/teams') {
            this.loadTeamsData();
        }
    }

    handleFleetUpdate(data) {
        console.log('Fleet update received:', data);
        this.showNotification(`Machine "${data.machine_name}" status changed to ${data.status}`, 'info');
        
        // Refresh fleet page if active
        if (window.location.pathname === '/fleet') {
            this.loadFleetData();
        }
    }

    handleUserActivity(data) {
        console.log('User activity received:', data);
        // Update activity indicators or user presence
        this.updateUserPresence(data);
    }

    updateUserPresence(data) {
        // Implementation for showing active users in collaboration
        const presenceContainer = document.getElementById('user-presence');
        if (presenceContainer) {
            // Update presence indicators
        }
    }

    async refreshConfigView() {
        // Refresh the current configuration view
        try {
            const response = await fetch(window.location.pathname + '?refresh=1');
            if (response.ok) {
                // Update the view without full page reload
                const data = await response.json();
                this.updateConfigView(data);
            }
        } catch (error) {
            console.error('Error refreshing config view:', error);
        }
    }

    // API helper methods
    async apiRequest(url, options = {}) {
        const defaultOptions = {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            }
        };

        try {
            const response = await fetch(url, { ...defaultOptions, ...options });
            
            // Get response as text first to check for parsing issues
            const responseText = await response.text();
            console.log(`API ${url} response (first 100 chars):`, responseText.substring(0, 100));
            
            let data;
            try {
                data = JSON.parse(responseText);
            } catch (jsonError) {
                console.error(`JSON parsing error for ${url}:`, jsonError);
                console.error('Full response text:', responseText);
                throw new Error(`Invalid JSON response from ${url}: ${jsonError.message}`);
            }
            
            if (!response.ok) {
                throw new Error(data.message || `Request failed with status ${response.status}`);
            }
            
            return data;
        } catch (error) {
            console.error('API request failed:', error);
            this.showNotification(`API Error: ${error.message}`, 'danger');
            throw error;
        }
    }

    async loadFleetData() {
        try {
            const data = await this.apiRequest('/api/fleet');
            this.updateFleetView(data.data);
        } catch (error) {
            console.error('Error loading fleet data:', error);
        }
    }

    async loadTeamsData() {
        try {
            const data = await this.apiRequest('/api/teams');
            this.updateTeamsView(data.data);
        } catch (error) {
            console.error('Error loading teams data:', error);
        }
    }

    async loadVersionsData() {
        try {
            const data = await this.apiRequest('/api/config/branches');
            this.updateVersionsView(data.data);
        } catch (error) {
            console.error('Error loading versions data:', error);
        }
    }

    async loadBuilderData() {
        try {
            const data = await this.apiRequest('/api/config/files');
            this.updateBuilderView(data.data);
        } catch (error) {
            console.error('Error loading builder data:', error);
        }
    }

    // View update methods
    updateFleetView(data) {
        // Implementation for updating fleet management view
        console.log('Updating fleet view with data:', data);
    }

    updateTeamsView(data) {
        // Implementation for updating teams view
        console.log('Updating teams view with data:', data);
    }

    updateVersionsView(data) {
        // Implementation for updating version control view
        console.log('Updating versions view with data:', data);
    }

    updateBuilderView(data) {
        // Implementation for updating configuration builder view
        console.log('Updating builder view with data:', data);
    }

    updateConfigView(data) {
        // Implementation for updating configuration view
        console.log('Updating config view with data:', data);
    }

    async loadDashboardDetails() {
        // Load additional dashboard details
        try {
            const data = await this.apiRequest('/api/dashboard/details');
            this.updateDashboardDetails(data.data);
        } catch (error) {
            console.error('Error loading dashboard details:', error);
        }
    }

    updateDashboardDetails(data) {
        // Implementation for updating detailed dashboard view
        console.log('Updating dashboard details with data:', data);
    }
}

// Initialize the app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.nixaiApp = new NixAIApp();
});

// CSS for notifications (injected via JavaScript)
const notificationStyles = `
.nixai-notification {
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    padding: var(--spacing-md);
    box-shadow: var(--shadow-lg);
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    min-width: 300px;
    max-width: 400px;
}

.nixai-notification-success {
    border-color: var(--success-color);
    background: rgb(34 197 94 / 0.1);
}

.nixai-notification-warning {
    border-color: var(--warning-color);
    background: rgb(245 158 11 / 0.1);
}

.nixai-notification-danger {
    border-color: var(--danger-color);
    background: rgb(239 68 68 / 0.1);
}

.nixai-notification-info {
    border-color: var(--primary-color);
    background: rgb(59 130 246 / 0.1);
}

.notification-close {
    background: none;
    border: none;
    font-size: 1.2rem;
    cursor: pointer;
    color: var(--text-secondary);
    margin-left: var(--spacing-sm);
}

.notification-close:hover {
    color: var(--text-primary);
}

.notification-content strong {
    display: block;
    margin-bottom: var(--spacing-xs);
}

.notification-content small {
    color: var(--text-secondary);
    font-size: 0.75rem;
}
`;

// Inject notification styles
const styleSheet = document.createElement('style');
styleSheet.textContent = notificationStyles;
document.head.appendChild(styleSheet);

// Global function definitions for HTML onclick handlers
// These functions need to be globally accessible for the HTML templates

// Configuration Builder Functions
window.createNewConfig = function() {
    console.log('Creating new configuration...');
    
    // Show modal for configuration creation
    const modal = `
        <div class="nixai-modal" id="configModal">
            <div class="nixai-modal-content">
                <div class="nixai-modal-header">
                    <h3>Create New Configuration</h3>
                    <button class="nixai-modal-close" onclick="closeModal('configModal')">&times;</button>
                </div>
                <div class="nixai-modal-body">
                    <form id="createConfigForm" class="nixai-form">
                        <div class="nixai-field">
                            <label for="configName">Configuration Name</label>
                            <input type="text" id="configName" name="name" required class="nixai-input" placeholder="my-nixos-config">
                        </div>
                        <div class="nixai-field">
                            <label for="configType">Configuration Type</label>
                            <select id="configType" name="type" class="nixai-select">
                                <option value="desktop">Desktop Environment</option>
                                <option value="server">Server</option>
                                <option value="development">Development Machine</option>
                                <option value="minimal">Minimal System</option>
                                <option value="custom">Custom</option>
                            </select>
                        </div>
                        <div class="nixai-field">
                            <label for="configDescription">Description</label>
                            <textarea id="configDescription" name="description" class="nixai-textarea" placeholder="Describe your configuration..."></textarea>
                        </div>
                        <div class="nixai-actions">
                            <button type="submit" class="nixai-button nixai-button-primary">
                                <i class="fas fa-plus"></i> Create Configuration
                            </button>
                            <button type="button" class="nixai-button nixai-button-secondary" onclick="closeModal('configModal')">
                                Cancel
                            </button>
                        </div>
                    </form>
                </div>
            </div>
        </div>
    `;
    
    // Add modal to page
    document.body.insertAdjacentHTML('beforeend', modal);
    
    // Handle form submission
    document.getElementById('createConfigForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        const configData = Object.fromEntries(formData);
        
        try {
            const response = await fetch('/api/configurations', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${localStorage.getItem('nixai-token')}`
                },
                body: JSON.stringify(configData)
            });
            
            if (response.ok) {
                const result = await response.json();
                console.log('Configuration created:', result);
                closeModal('configModal');
                window.nixaiApp?.showNotification('Configuration created successfully!', 'success');
                // Refresh the configuration list
                if (typeof refreshConfigurations === 'function') {
                    refreshConfigurations();
                }
            } else {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
        } catch (error) {
            console.error('Error creating configuration:', error);
            window.nixaiApp?.showNotification('Failed to create configuration: ' + error.message, 'error');
        }
    });
};

window.loadTemplate = function() {
    console.log('Loading configuration template...');
    window.nixaiApp?.showNotification('Template loading feature coming soon!', 'info');
};

window.importConfig = function() {
    console.log('Importing existing configuration...');
    window.nixaiApp?.showNotification('Import feature coming soon!', 'info');
};

// Fleet Management Functions
window.addMachine = function() {
    console.log('Adding new machine...');
    
    const modal = `
        <div class="nixai-modal" id="machineModal">
            <div class="nixai-modal-content">
                <div class="nixai-modal-header">
                    <h3>Add New Machine</h3>
                    <button class="nixai-modal-close" onclick="closeModal('machineModal')">&times;</button>
                </div>
                <div class="nixai-modal-body">
                    <form id="addMachineForm" class="nixai-form">
                        <div class="nixai-field">
                            <label for="machineId">Machine ID</label>
                            <input type="text" id="machineId" name="id" required class="nixai-input" placeholder="server-01">
                        </div>
                        <div class="nixai-field">
                            <label for="machineName">Machine Name</label>
                            <input type="text" id="machineName" name="name" required class="nixai-input" placeholder="Production Server 1">
                        </div>
                        <div class="nixai-field">
                            <label for="machineAddress">IP Address/Hostname</label>
                            <input type="text" id="machineAddress" name="address" required class="nixai-input" placeholder="192.168.1.100">
                        </div>
                        <div class="nixai-field">
                            <label for="machineEnvironment">Environment</label>
                            <select id="machineEnvironment" name="environment" class="nixai-select">
                                <option value="production">Production</option>
                                <option value="staging">Staging</option>
                                <option value="development">Development</option>
                                <option value="testing">Testing</option>
                            </select>
                        </div>
                        <div class="nixai-field">
                            <label for="machineTags">Tags (comma-separated)</label>
                            <input type="text" id="machineTags" name="tags" class="nixai-input" placeholder="web,database,critical">
                        </div>
                        <div class="nixai-field">
                            <label for="sshUser">SSH User</label>
                            <input type="text" id="sshUser" name="ssh_user" class="nixai-input" placeholder="root" value="root">
                        </div>
                        <div class="nixai-field">
                            <label for="sshPort">SSH Port</label>
                            <input type="number" id="sshPort" name="ssh_port" class="nixai-input" placeholder="22" value="22">
                        </div>
                        <div class="nixai-actions">
                            <button type="submit" class="nixai-button nixai-button-primary">
                                <i class="fas fa-plus"></i> Add Machine
                            </button>
                            <button type="button" class="nixai-button nixai-button-secondary" onclick="closeModal('machineModal')">
                                Cancel
                            </button>
                        </div>
                    </form>
                </div>
            </div>
        </div>
    `;
    
    document.body.insertAdjacentHTML('beforeend', modal);
    
    document.getElementById('addMachineForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        const machineData = Object.fromEntries(formData);
        
        // Convert tags to array
        if (machineData.tags) {
            machineData.tags = machineData.tags.split(',').map(tag => tag.trim()).filter(tag => tag);
        }
        
        try {
            const response = await fetch('/api/fleet/machines', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${localStorage.getItem('nixai-token')}`
                },
                body: JSON.stringify(machineData)
            });
            
            if (response.ok) {
                const result = await response.json();
                console.log('Machine added:', result);
                closeModal('machineModal');
                window.nixaiApp?.showNotification('Machine added successfully!', 'success');
                // Refresh the fleet list
                if (typeof refreshFleet === 'function') {
                    refreshFleet();
                }
            } else {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
        } catch (error) {
            console.error('Error adding machine:', error);
            window.nixaiApp?.showNotification('Failed to add machine: ' + error.message, 'error');
        }
    });
};

window.bulkDeploy = function() {
    console.log('Starting bulk deployment...');
    window.nixaiApp?.showNotification('Bulk deployment feature coming soon!', 'info');
};

window.refreshFleet = async function() {
    console.log('Refreshing fleet...');
    try {
        const response = await fetch('/api/fleet', {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('nixai-token')}`
            }
        });
        
        if (response.ok) {
            const data = await response.json();
            console.log('Fleet data refreshed:', data);
            window.nixaiApp?.showNotification('Fleet data refreshed!', 'success');
            // Update UI with new data
            updateFleetUI(data);
        } else {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
    } catch (error) {
        console.error('Error refreshing fleet:', error);
        window.nixaiApp?.showNotification('Failed to refresh fleet: ' + error.message, 'error');
    }
};

// Team Management Functions
window.createTeam = function() {
    console.log('Creating new team...');
    
    const modal = `
        <div class="nixai-modal" id="teamModal">
            <div class="nixai-modal-content">
                <div class="nixai-modal-header">
                    <h3>Create New Team</h3>
                    <button class="nixai-modal-close" onclick="closeModal('teamModal')">&times;</button>
                </div>
                <div class="nixai-modal-body">
                    <form id="createTeamForm" class="nixai-form">
                        <div class="nixai-field">
                            <label for="teamName">Team Name</label>
                            <input type="text" id="teamName" name="name" required class="nixai-input" placeholder="Development Team">
                        </div>
                        <div class="nixai-field">
                            <label for="teamDescription">Description</label>
                            <textarea id="teamDescription" name="description" class="nixai-textarea" placeholder="Describe your team's purpose and goals..."></textarea>
                        </div>
                        <div class="nixai-field">
                            <label>
                                <input type="checkbox" id="teamPublic" name="public" class="nixai-checkbox">
                                Make team publicly discoverable
                            </label>
                        </div>
                        <div class="nixai-actions">
                            <button type="submit" class="nixai-button nixai-button-primary">
                                <i class="fas fa-plus"></i> Create Team
                            </button>
                            <button type="button" class="nixai-button nixai-button-secondary" onclick="closeModal('teamModal')">
                                Cancel
                            </button>
                        </div>
                    </form>
                </div>
            </div>
        </div>
    `;
    
    document.body.insertAdjacentHTML('beforeend', modal);
    
    document.getElementById('createTeamForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        const teamData = Object.fromEntries(formData);
        teamData.public = !!teamData.public; // Convert to boolean
        
        try {
            const response = await fetch('/api/teams', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${localStorage.getItem('nixai-token')}`
                },
                body: JSON.stringify(teamData)
            });
            
            if (response.ok) {
                const result = await response.json();
                console.log('Team created:', result);
                closeModal('teamModal');
                window.nixaiApp?.showNotification('Team created successfully!', 'success');
                // Refresh the teams list
                if (typeof refreshTeams === 'function') {
                    refreshTeams();
                }
            } else {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
        } catch (error) {
            console.error('Error creating team:', error);
            window.nixaiApp?.showNotification('Failed to create team: ' + error.message, 'error');
        }
    });
};

window.joinTeam = function() {
    console.log('Joining team...');
    window.nixaiApp?.showNotification('Join team feature coming soon!', 'info');
};

window.inviteMembers = function() {
    console.log('Inviting members...');
    window.nixaiApp?.showNotification('Invite members feature coming soon!', 'info');
};

window.refreshTeams = async function() {
    console.log('Refreshing teams...');
    try {
        const response = await fetch('/api/teams', {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('nixai-token')}`
            }
        });
        
        if (response.ok) {
            const data = await response.json();
            console.log('Teams data refreshed:', data);
            window.nixaiApp?.showNotification('Teams data refreshed!', 'success');
            // Update UI with new data
            updateTeamsUI(data);
        } else {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
    } catch (error) {
        console.error('Error refreshing teams:', error);
        window.nixaiApp?.showNotification('Failed to refresh teams: ' + error.message, 'error');
    }
};

// Modal Management Functions
window.closeModal = function(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
        modal.remove();
    }
};

// UI Update Functions
function updateFleetUI(data) {
    // Update fleet statistics
    if (data.data && data.data.machines) {
        const machines = data.data.machines;
        document.getElementById('totalMachines').textContent = machines.length;
        
        const healthy = machines.filter(m => m.status === 'healthy').length;
        const warnings = machines.filter(m => m.status === 'warning').length;
        const errors = machines.filter(m => m.status === 'error').length;
        
        document.getElementById('healthyMachines').textContent = healthy;
        document.getElementById('warningMachines').textContent = warnings;
        document.getElementById('errorMachines').textContent = errors;
    }
}

function updateTeamsUI(data) {
    // Update teams list and statistics
    if (data.data) {
        const teams = Array.isArray(data.data) ? data.data : [];
        document.getElementById('totalTeams').textContent = teams.length;
        
        // Update teams list in sidebar
        const teamsList = document.getElementById('teamsList');
        if (teamsList) {
            teamsList.innerHTML = teams.map(team => `
                <div class="team-item" data-team-id="${team.id}">
                    <div class="team-info">
                        <h4>${team.name}</h4>
                        <p>${team.description || 'No description'}</p>
                        <span class="team-members">${Object.keys(team.members || {}).length} members</span>
                    </div>
                </div>
            `).join('');
        }
    }
}

// Initialize global functions when document loads
document.addEventListener('DOMContentLoaded', () => {
    console.log('✅ Global interactive functions loaded');
    
    // Store reference to the main app instance
    if (window.nixaiApp) {
        console.log('📱 NixAI App instance available globally');
    }
});
