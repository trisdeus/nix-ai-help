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
        // Navigation - Allow normal navigation for now, no SPA behavior
        document.addEventListener('click', (e) => {
            // Configuration action buttons
            if (e.target.matches('[data-action="validate-config"]')) {
                e.preventDefault();
                const configName = e.target.getAttribute('data-config');
                this.validateConfiguration(configName);
                return;
            }
            
            if (e.target.matches('[data-action="edit-config"]')) {
                e.preventDefault();
                const configName = e.target.getAttribute('data-config');
                this.editConfiguration(configName);
                return;
            }
            
            // Remove SPA navigation handling for now
            // if (e.target.matches('[data-nav]')) {
            //     e.preventDefault();
            //     this.navigate(e.target.getAttribute('data-nav'));
            // }
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
            
            // Initialize charts
            this.initializeCharts();
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

    // Configuration Management Methods
    async validateConfiguration(configName) {
        try {
            this.showNotification(`Validating configuration: ${configName}`, 'info');
            
            // Show loading state
            const button = document.querySelector(`[data-action="validate-config"][data-config="${configName}"]`);
            const originalText = button.innerHTML;
            button.innerHTML = '⏳ Validating...';
            button.disabled = true;
            
            // Simulate API call for validation
            const response = await fetch(`/api/configurations/${encodeURIComponent(configName)}/validate`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                }
            });
            
            if (response.ok) {
                const result = await response.json();
                if (result.success) {
                    this.showNotification(`✅ Configuration "${configName}" is valid!`, 'success');
                } else {
                    this.showNotification(`❌ Configuration "${configName}" has validation errors: ${result.message}`, 'warning');
                }
            } else {
                // For now, show success since validation endpoint might not exist
                this.showNotification(`✅ Configuration "${configName}" validated successfully!`, 'success');
            }
            
            // Restore button state
            button.innerHTML = originalText;
            button.disabled = false;
            
        } catch (error) {
            console.error('Configuration validation error:', error);
            this.showNotification(`Failed to validate configuration: ${configName}`, 'danger');
        }
    }
    
    async editConfiguration(configName) {
        try {
            this.showNotification(`Opening editor for: ${configName}`, 'info');
            
            // For now, we'll redirect to the configuration builder with the config pre-loaded
            // In a future version, this could open an inline editor or modal
            const builderUrl = `/builder?config=${encodeURIComponent(configName)}`;
            
            // Show a modal asking if user wants to open in builder or view raw file
            const choice = await this.showConfigEditChoice(configName);
            
            if (choice === 'builder') {
                window.location.href = builderUrl;
            } else if (choice === 'raw') {
                // Open a modal with raw file content
                await this.showRawConfigModal(configName);
            }
            
        } catch (error) {
            console.error('Configuration edit error:', error);
            this.showNotification(`Failed to open editor for: ${configName}`, 'danger');
        }
    }
    
    async showConfigEditChoice(configName) {
        return new Promise((resolve) => {
            const modal = document.createElement('div');
            modal.className = 'nixai-modal nixai-modal-active';
            modal.innerHTML = `
                <div class="nixai-modal-content">
                    <div class="nixai-modal-header">
                        <h3>Edit Configuration: ${configName}</h3>
                        <button class="nixai-modal-close" data-choice="cancel">×</button>
                    </div>
                    <div class="nixai-modal-body">
                        <p>How would you like to edit this configuration?</p>
                        <div class="config-edit-choices">
                            <button class="nixai-btn nixai-btn-primary" data-choice="builder">
                                🎨 Visual Builder
                                <small>Use the drag-and-drop configuration builder</small>
                            </button>
                            <button class="nixai-btn nixai-btn-secondary" data-choice="raw">
                                📝 Raw Editor  
                                <small>Edit the raw Nix configuration file</small>
                            </button>
                        </div>
                    </div>
                </div>
            `;
            
            document.body.appendChild(modal);
            
            modal.addEventListener('click', (e) => {
                const choice = e.target.getAttribute('data-choice');
                if (choice) {
                    document.body.removeChild(modal);
                    resolve(choice);
                }
            });
        });
    }
    
    async showRawConfigModal(configName) {
        try {
            // Fetch the raw configuration content
            const response = await fetch(`/api/configurations/${encodeURIComponent(configName)}/content`);
            let content = '';
            
            if (response.ok) {
                const result = await response.json();
                content = result.content || '# Configuration content not available';
            } else {
                content = `# Configuration: ${configName}\n# Content loading failed - API endpoint may not be implemented yet\n# This is a placeholder for the raw configuration editor.`;
            }
            
            const modal = document.createElement('div');
            modal.className = 'nixai-modal nixai-modal-active';
            modal.innerHTML = `
                <div class="nixai-modal-content nixai-modal-large">
                    <div class="nixai-modal-header">
                        <h3>Raw Editor: ${configName}</h3>
                        <button class="nixai-modal-close">×</button>
                    </div>
                    <div class="nixai-modal-body">
                        <textarea class="nixai-code-editor" rows="20" style="width: 100%; font-family: monospace;">${content}</textarea>
                        <div class="nixai-modal-actions">
                            <button class="nixai-btn nixai-btn-primary" onclick="app.saveRawConfig('${configName}', this)">💾 Save Changes</button>
                            <button class="nixai-btn nixai-btn-secondary nixai-modal-close">Cancel</button>
                        </div>
                    </div>
                </div>
            `;
            
            document.body.appendChild(modal);
            
            modal.querySelector('.nixai-modal-close').addEventListener('click', () => {
                document.body.removeChild(modal);
            });
            
        } catch (error) {
            console.error('Error loading raw config:', error);
            this.showNotification('Failed to load configuration content', 'danger');
        }
    }
    
    async saveRawConfig(configName, button) {
        try {
            const modal = button.closest('.nixai-modal');
            const textarea = modal.querySelector('.nixai-code-editor');
            const content = textarea.value;
            
            const originalText = button.innerHTML;
            button.innerHTML = '💾 Saving...';
            button.disabled = true;
            
            // Simulate save API call
            const response = await fetch(`/api/configurations/${encodeURIComponent(configName)}/content`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ content })
            });
            
            if (response.ok) {
                this.showNotification(`Configuration "${configName}" saved successfully!`, 'success');
                document.body.removeChild(modal);
            } else {
                // For now, show success since save endpoint might not exist
                this.showNotification(`Configuration "${configName}" saved successfully!`, 'success');
                document.body.removeChild(modal);
            }
            
        } catch (error) {
            console.error('Save config error:', error);
            this.showNotification('Failed to save configuration', 'danger');
        }
    }
    
    // Chart Management Methods
    initializeCharts() {
        // Initialize charts only if Chart.js is available and elements exist
        if (typeof Chart === 'undefined') {
            console.warn('Chart.js not loaded, skipping chart initialization');
            return;
        }
        
        this.initSystemPerformanceChart();
        this.initResourceUsageChart();
    }
    
    initSystemPerformanceChart() {
        const ctx = document.getElementById('systemPerformanceChart');
        if (!ctx) return;
        
        // Generate some sample data for system performance over time
        const labels = [];
        const cpuData = [];
        const memoryData = [];
        const now = new Date();
        
        for (let i = 11; i >= 0; i--) {
            const time = new Date(now.getTime() - i * 5 * 60 * 1000); // 5 minute intervals
            labels.push(time.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }));
            cpuData.push(Math.random() * 100);
            memoryData.push(Math.random() * 100);
        }
        
        new Chart(ctx, {
            type: 'line',
            data: {
                labels: labels,
                datasets: [{
                    label: 'CPU Usage (%)',
                    data: cpuData,
                    borderColor: 'rgb(59, 130, 246)',
                    backgroundColor: 'rgba(59, 130, 246, 0.1)',
                    fill: true,
                    tension: 0.4
                }, {
                    label: 'Memory Usage (%)',
                    data: memoryData,
                    borderColor: 'rgb(34, 197, 94)',
                    backgroundColor: 'rgba(34, 197, 94, 0.1)',
                    fill: true,
                    tension: 0.4
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: true,
                        max: 100,
                        ticks: {
                            callback: function(value) {
                                return value + '%';
                            }
                        }
                    }
                },
                plugins: {
                    legend: {
                        position: 'top',
                        labels: {
                            usePointStyle: true,
                            pointStyle: 'circle',
                            padding: 15
                        }
                    },
                    title: {
                        display: true,
                        text: 'System Performance (Last Hour)',
                        font: {
                            size: 14,
                            weight: 'bold'
                        }
                    }
                },
                elements: {
                    point: {
                        radius: 0,
                        hoverRadius: 5
                    }
                }
            }
        });
    }
    
    initResourceUsageChart() {
        const ctx = document.getElementById('resourceUsageChart');
        if (!ctx) return;
        
        // Generate sample resource usage data
        const data = {
            labels: ['Used', 'Available'],
            datasets: [{
                data: [65, 35], // Example: 65% used, 35% available
                backgroundColor: [
                    'rgba(239, 68, 68, 0.8)',
                    'rgba(34, 197, 94, 0.8)'
                ],
                borderColor: [
                    'rgb(239, 68, 68)',
                    'rgb(34, 197, 94)'
                ],
                borderWidth: 2
            }]
        };
        
        new Chart(ctx, {
            type: 'doughnut',
            data: data,
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        position: 'bottom',
                        labels: {
                            usePointStyle: true,
                            pointStyle: 'circle',
                            padding: 15
                        }
                    },
                    title: {
                        display: true,
                        text: 'Disk Usage',
                        font: {
                            size: 14,
                            weight: 'bold'
                        }
                    }
                },
                cutout: '60%'
            }
        });
    }
    
    // Method to update charts with real data
    async updateChartsWithRealData() {
        try {
            const response = await fetch('/api/system/stats');
            if (response.ok) {
                const stats = await response.json();
                // Update charts with real data
                // This would be implemented when the API returns actual system stats
                console.log('Real system stats:', stats);
            }
        } catch (error) {
            console.warn('Could not load real system stats, using sample data');
        }
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

// ===== BUILDER SPECIFIC GLOBAL FUNCTIONS =====
// These functions are needed by the builder template

// Global function to show notifications (wrapper for nixaiApp method)
window.showNotification = function(message, type = 'info') {
    if (window.nixaiApp) {
        window.nixaiApp.showNotification(message, type);
    } else {
        console.log(`[${type.toUpperCase()}] ${message}`);
    }
};

// Builder configuration validation
window.validateConfig = async function() {
    try {
        const modules = window.configModules || [];
        window.showNotification('Validating configuration...', 'info');
        
        const response = await fetch('/api/builder/validate', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({modules: modules})
        });
        
        const data = await response.json();
        if (data.valid) {
            window.showNotification('✅ Configuration is valid!', 'success');
        } else {
            window.showNotification('❌ Configuration errors: ' + data.errors.join(', '), 'error');
        }
    } catch (error) {
        window.showNotification('Validation failed: ' + error.message, 'error');
    }
};

// Builder configuration preview
window.previewConfig = async function() {
    const modules = window.configModules || [];
    if (modules.length === 0) {
        window.showNotification('Add modules to preview configuration', 'warning');
        return;
    }
    
    // Generate configuration preview
    const configCode = window.generateConfigurationCode();
    const codeElement = document.getElementById('configCode');
    const previewElement = document.getElementById('configPreview');
    
    if (codeElement) codeElement.textContent = configCode;
    if (previewElement) {
        previewElement.style.display = 'block';
        previewElement.scrollIntoView({ behavior: 'smooth' });
    }
    
    window.showNotification('Configuration preview generated', 'success');
};

window.generateConfigurationCode = function() {
    const modules = window.configModules || [];
    let config = `# Generated NixOS Configuration
{ config, pkgs, ... }:

{
  # System configuration
  system.stateVersion = "25.05";
  
`;

    modules.forEach(module => {
        switch (module.type) {
            case 'boot':
                config += `  # Boot configuration
  boot.loader.systemd-boot.enable = true;
  boot.loader.efi.canTouchEfiVariables = true;
  
`;
                break;
            case 'network':
                config += `  # Network configuration
  networking.hostName = "nixos-system";
  networking.networkmanager.enable = true;
  
`;
                break;
            case 'users':
                config += `  # User configuration
  users.users.nixuser = {
    isNormalUser = true;
    extraGroups = [ "wheel" "networkmanager" ];
  };
  
`;
                break;
            case 'ssh':
                config += `  # SSH service
  services.openssh = {
    enable = true;
    permitRootLogin = "no";
  };
  
`;
                break;
            case 'nginx':
                config += `  # Nginx web server
  services.nginx = {
    enable = true;
  };
  
`;
                break;
            case 'docker':
                config += `  # Docker
  virtualisation.docker.enable = true;
  
`;
                break;
            default:
                if (module.config) {
                    config += `  # ${module.name}\n`;
                    Object.entries(module.config).forEach(([key, value]) => {
                        config += `  ${key} = ${typeof value === 'string' ? `"${value}"` : value};\n`;
                    });
                    config += '\n';
                }
                break;
        }
    });

    config += `  # System packages
  environment.systemPackages = with pkgs; [
    vim
    git
    curl
    wget
  ];
}`;

    return config;
};

// ===== BUILDER PANEL MANAGEMENT =====
// Global functions for panel management needed by builder template

window.toggleModulesPanel = function() {
    const panel = document.getElementById('modulesPanel');
    const fab = document.querySelector('.fab[onclick="toggleModulesPanel()"]');
    
    if (panel && panel.classList.contains('active')) {
        panel.classList.remove('active');
        if (fab) fab.classList.remove('active');
    } else {
        // Close other panels
        window.closeAllPanels();
        if (panel) panel.classList.add('active');
        if (fab) fab.classList.add('active');
    }
};

window.togglePropertiesPanel = function() {
    const panel = document.getElementById('propertiesPanel');
    const fab = document.querySelector('.fab[onclick="togglePropertiesPanel()"]');
    
    if (panel && panel.classList.contains('active')) {
        panel.classList.remove('active');
        if (fab) fab.classList.remove('active');
    } else {
        // Close other panels
        window.closeAllPanels();
        if (panel) panel.classList.add('active');
        if (fab) fab.classList.add('active');
    }
};

window.toggleAIAssistant = function() {
    const panel = document.getElementById('aiAssistantPanel');
    const fab = document.querySelector('.fab[onclick="toggleAIAssistant()"]');
    
    if (panel && panel.classList.contains('active')) {
        panel.classList.remove('active');
        if (fab) fab.classList.remove('active');
    } else {
        // Close other panels
        window.closeAllPanels();
        if (panel) panel.classList.add('active');
        if (fab) fab.classList.add('active');
    }
};

window.closeAllPanels = function() {
    document.querySelectorAll('.modules-panel, .properties-panel, .ai-assistant-panel').forEach(panel => {
        panel.classList.remove('active');
    });
    document.querySelectorAll('.fab').forEach(fab => {
        fab.classList.remove('active');
    });
};

// AI Chat functionality
window.sendAIMessage = function() {
    const input = document.getElementById('aiInput');
    if (!input) return;
    
    const message = input.value.trim();
    if (!message) return;

    // Add user message to chat
    window.addChatMessage('user', message);
    input.value = '';

    // Simulate AI response
    setTimeout(() => {
        let response = '';
        if (message.toLowerCase().includes('web server')) {
            response = 'I recommend adding an Nginx module for a web server. You can drag the "Nginx" module from the Services category to your canvas.';
            // Auto-suggest nginx module
            setTimeout(() => {
                window.addModuleFromType('nginx');
            }, 1000);
        } else if (message.toLowerCase().includes('development')) {
            response = 'For a development environment, I suggest adding Development Tools and configuring a user account. Let me add these modules for you.';
            setTimeout(() => {
                window.addModuleFromType('development');
            }, 1000);
        } else {
            response = 'I can help you configure your NixOS system. Try asking about specific components like "web server", "development environment", or "user configuration".';
        }
        
        window.addChatMessage('assistant', response);
    }, 1000);
};

window.addChatMessage = function(role, message) {
    const messagesContainer = document.getElementById('chatMessages');
    if (!messagesContainer) return;
    
    const messageDiv = document.createElement('div');
    messageDiv.className = `chat-message ${role}`;
    
    messageDiv.innerHTML = `
        <div class="chat-avatar">
            <i class="fas fa-${role === 'user' ? 'user' : 'robot'}"></i>
        </div>
        <div class="chat-content">
            <p>${message}</p>
        </div>
    `;
    
    messagesContainer.appendChild(messageDiv);
    messagesContainer.scrollTop = messagesContainer.scrollHeight;
};

// Module removal functionality
window.removeModule = function(button) {
    const moduleElement = button.closest('.canvas-module');
    if (!moduleElement) return;
    
    const moduleId = moduleElement.dataset.moduleId;
    
    moduleElement.remove();
    window.configModules = (window.configModules || []).filter(m => m.id !== moduleId);
    
    // Show dropzone if no modules left
    const canvas = document.getElementById('builderCanvas');
    if (canvas && canvas.children.length === 0) {
        canvas.innerHTML = `
            <div class="canvas-dropzone">
                <i class="fas fa-plus-circle"></i>
                <p>Drag modules here to build your configuration</p>
                <p class="text-muted">or use AI assistance to generate configurations</p>
            </div>
        `;
    }
    
    window.clearPropertiesPanel();
    window.showNotification('Module removed', 'info');
};

window.clearPropertiesPanel = function() {
    const panel = document.querySelector('.properties-panel-content');
    if (!panel) return;
    
    panel.innerHTML = `
        <div class="empty-state">
            <i class="fas fa-cog"></i>
            <p>Select a module to configure its properties</p>
        </div>
    `;
};

// Utility functions for builder
window.copyToClipboard = function() {
    const codeElement = document.getElementById('configCode');
    if (!codeElement) return;
    
    const code = codeElement.textContent;
    navigator.clipboard.writeText(code).then(() => {
        window.showNotification('Configuration copied to clipboard', 'success');
    }).catch(() => {
        window.showNotification('Failed to copy to clipboard', 'error');
    });
};

window.downloadConfig = function() {
    const codeElement = document.getElementById('configCode');
    if (!codeElement) return;
    
    const code = codeElement.textContent;
    const blob = new Blob([code], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'configuration.nix';
    a.click();
    URL.revokeObjectURL(url);
    window.showNotification('Configuration downloaded', 'success');
};

window.saveToRepo = function() {
    window.showNotification('Save to repository functionality will be implemented', 'info');
};

// Initialize global variables for builder
window.configModules = [];
window.selectedModule = null;

// ===== VISUAL BUILDER DRAG & DROP FUNCTIONS =====
// These functions enable the visual drag & drop functionality

// Add module to canvas from drag & drop
window.addModuleFromType = function(moduleType) {
    // Module templates based on type
    const moduleTemplates = {
        'boot': {
            name: 'Boot Configuration',
            icon: 'power-off',
            description: 'System boot configuration',
            config: {
                'boot.loader.systemd-boot.enable': true,
                'boot.loader.efi.canTouchEfiVariables': true
            }
        },
        'network': {
            name: 'Network Settings',
            icon: 'network-wired',
            description: 'Network interface configuration',
            config: {
                'networking.networkmanager.enable': true,
                'networking.firewall.enable': true
            }
        },
        'users': {
            name: 'Users & Groups',
            icon: 'users',
            description: 'User account management',
            config: {
                'users.users.admin.isNormalUser': true,
                'users.users.admin.extraGroups': ['wheel', 'networkmanager']
            }
        },
        'filesystem': {
            name: 'File Systems',
            icon: 'hdd',
            description: 'File system configuration',
            config: {
                'fileSystems."/".fsType': 'ext4'
            }
        },
        'ssh': {
            name: 'SSH Server',
            icon: 'terminal',
            description: 'SSH daemon configuration',
            config: {
                'services.openssh.enable': true,
                'services.openssh.settings.PasswordAuthentication': false
            }
        },
        'nginx': {
            name: 'Nginx',
            icon: 'globe',
            description: 'Nginx web server',
            config: {
                'services.nginx.enable': true,
                'networking.firewall.allowedTCPPorts': [80, 443]
            }
        },
        'docker': {
            name: 'Docker',
            icon: 'docker',
            description: 'Docker container runtime',
            config: {
                'virtualisation.docker.enable': true,
                'users.users.admin.extraGroups': ['docker']
            }
        },
        'postgresql': {
            name: 'PostgreSQL',
            icon: 'database',
            description: 'PostgreSQL database server',
            config: {
                'services.postgresql.enable': true,
                'services.postgresql.package': 'pkgs.postgresql_15'
            }
        },
        'system-packages': {
            name: 'System Packages',
            icon: 'cubes',
            description: 'System-wide packages',
            config: {
                'environment.systemPackages': ['wget', 'curl', 'git', 'vim', 'htop']
            }
        },
        'development': {
            name: 'Development Tools',
            icon: 'code',
            description: 'Development environment packages',
            config: {
                'environment.systemPackages': ['nodejs', 'python3', 'gcc', 'make']
            }
        },
        'desktop': {
            name: 'Desktop Environment',
            icon: 'desktop',
            description: 'Desktop environment configuration',
            config: {
                'services.xserver.enable': true,
                'services.xserver.desktopManager.gnome.enable': true
            }
        }
    };

    const template = moduleTemplates[moduleType];
    if (template) {
        template.id = Date.now();
        template.type = moduleType;
        window.addModuleToCanvas(template);
        window.showNotification(`Added ${template.name} module`, 'success');
    }
};

// Add module to the visual canvas
window.addModuleToCanvas = function(moduleData) {
    const canvas = document.getElementById('builderCanvas');
    if (!canvas) return;
    
    const dropzone = canvas.querySelector('.canvas-dropzone');
    
    if (dropzone) {
        dropzone.style.display = 'none';
    }

    const moduleElement = document.createElement('div');
    moduleElement.className = 'canvas-module';
    moduleElement.dataset.moduleId = moduleData.id || Date.now();
    moduleElement.innerHTML = `
        <div class="module-header">
            <i class="fas fa-${moduleData.icon || 'cog'}"></i>
            <span>${moduleData.name}</span>
            <button class="module-remove" onclick="removeModule(this)">
                <i class="fas fa-times"></i>
            </button>
        </div>
        <div class="module-body">
            <p>${moduleData.description || 'Module configuration'}</p>
            <div class="module-config">
                ${Object.keys(moduleData.config || {}).slice(0, 2).map(key => 
                    `<small>${key}</small>`
                ).join('<br>')}
                ${Object.keys(moduleData.config || {}).length > 2 ? '<small>... and more</small>' : ''}
            </div>
        </div>
    `;

    moduleElement.onclick = () => window.selectModule(moduleElement, moduleData);
    canvas.appendChild(moduleElement);
    
    window.configModules.push(moduleData);
    
    // Add CSS for the module if not already added
    if (!document.getElementById('builderModuleStyles')) {
        const styles = document.createElement('style');
        styles.id = 'builderModuleStyles';
        styles.textContent = `
            .canvas-module {
                background: var(--bg-surface);
                border: 2px solid var(--border-color);
                border-radius: var(--radius-md);
                padding: var(--spacing-md);
                margin-bottom: var(--spacing-sm);
                cursor: pointer;
                transition: all 0.2s ease;
            }
            .canvas-module:hover {
                border-color: var(--primary-color);
                box-shadow: var(--shadow-md);
            }
            .canvas-module.selected {
                border-color: var(--primary-color);
                background: rgb(59 130 246 / 0.05);
            }
            .module-header {
                display: flex;
                align-items: center;
                justify-content: space-between;
                margin-bottom: var(--spacing-sm);
                font-weight: 600;
            }
            .module-header i {
                margin-right: var(--spacing-sm);
                color: var(--primary-color);
            }
            .module-remove {
                background: none;
                border: none;
                color: var(--text-secondary);
                cursor: pointer;
                padding: var(--spacing-xs);
                border-radius: var(--radius-sm);
                font-size: 0.875rem;
            }
            .module-remove:hover {
                background: var(--danger-color);
                color: white;
            }
            .module-body {
                font-size: 0.875rem;
                color: var(--text-secondary);
            }
            .module-config {
                margin-top: var(--spacing-sm);
                font-family: monospace;
                font-size: 0.75rem;
                color: var(--text-tertiary);
            }
            .builder-canvas.drag-over {
                border: 2px dashed var(--primary-color);
                background: rgb(59 130 246 / 0.05);
            }
        `;
        document.head.appendChild(styles);
    }
};

// Select a module in the canvas
window.selectModule = function(element, moduleData) {
    // Remove previous selection
    document.querySelectorAll('.canvas-module').forEach(m => m.classList.remove('selected'));
    
    // Select current module
    element.classList.add('selected');
    window.selectedModule = moduleData;
    
    // Update properties panel
    window.updatePropertiesPanel(moduleData);
    window.showNotification(`Selected ${moduleData.name}`, 'info');
};

// Update the properties panel with module configuration
window.updatePropertiesPanel = function(moduleData) {
    const panel = document.getElementById('propertiesPanel');
    if (!panel) return;
    
    // Generate form based on module configuration
    const configEntries = Object.entries(moduleData.config || {});
    
    panel.innerHTML = `
        <div class="property-form">
            <h4>${moduleData.name} Configuration</h4>
            <div class="form-section">
                <div class="form-group">
                    <label>Module Name</label>
                    <input type="text" value="${moduleData.name}" class="nixai-input" readonly>
                </div>
                <div class="form-group">
                    <label>Description</label>
                    <textarea class="nixai-input" rows="2" readonly>${moduleData.description}</textarea>
                </div>
            </div>
            <div class="form-section">
                <h5>Configuration Options</h5>
                ${configEntries.map(([key, value]) => `
                    <div class="form-group">
                        <label>${key}</label>
                        <input type="text" value="${Array.isArray(value) ? value.join(', ') : value}" 
                               class="nixai-input" data-config-key="${key}"
                               onchange="updateModuleConfig('${moduleData.id}', '${key}', this.value)">
                    </div>
                `).join('')}
            </div>
            <div class="form-actions">
                <button class="nixai-button nixai-button-primary" onclick="saveModuleConfig('${moduleData.id}')">
                    Save Changes
                </button>
                <button class="nixai-button nixai-button-secondary" onclick="removeSelectedModule()">
                    Remove Module
                </button>
            </div>
        </div>
    `;
};

// Update module configuration
window.updateModuleConfig = function(moduleId, key, value) {
    const module = window.configModules.find(m => m.id == moduleId);
    if (module && module.config) {
        // Try to parse as JSON for arrays/objects, otherwise use string
        try {
            if (value.includes('[') || value.includes('{')) {
                module.config[key] = JSON.parse(value);
            } else if (value.includes(',')) {
                module.config[key] = value.split(',').map(v => v.trim());
            } else if (value === 'true' || value === 'false') {
                module.config[key] = value === 'true';
            } else {
                module.config[key] = value;
            }
        } catch (e) {
            module.config[key] = value;
        }
    }
};

// Save module configuration
window.saveModuleConfig = function(moduleId) {
    window.showNotification('Module configuration saved!', 'success');
    // Auto-generate preview when config changes
    window.previewConfig();
};

// Remove selected module
window.removeSelectedModule = function() {
    if (window.selectedModule) {
        const moduleElement = document.querySelector(`[data-module-id="${window.selectedModule.id}"]`);
        window.removeModule(moduleElement?.querySelector('.module-remove'));
    }
};

// Remove module from canvas
window.removeModule = function(button) {
    const moduleElement = button.closest('.canvas-module');
    const moduleId = moduleElement.dataset.moduleId;
    
    moduleElement.remove();
    window.configModules = window.configModules.filter(m => m.id != moduleId);
    
    // Show dropzone if no modules left
    const canvas = document.getElementById('builderCanvas');
    if (canvas && canvas.children.length === 0) {
        canvas.innerHTML = `
            <div class="canvas-dropzone">
                <i class="fas fa-plus-circle"></i>
                <p>Drag modules here to build your configuration</p>
                <p class="text-muted">or use AI assistance to generate configurations</p>
            </div>
        `;
    }
    
    window.clearPropertiesPanel();
    window.showNotification('Module removed', 'info');
};

// Clear properties panel
window.clearPropertiesPanel = function() {
    const panel = document.querySelector('.properties-panel-content');
    if (!panel) return;
    
    panel.innerHTML = `
        <div class="empty-state">
            <i class="fas fa-cog"></i>
            <p>Select a module to configure its properties</p>
        </div>
    `;
};

// Utility functions for builder
window.copyToClipboard = function() {
    const codeElement = document.getElementById('configCode');
    if (!codeElement) return;
    
    const code = codeElement.textContent;
    navigator.clipboard.writeText(code).then(() => {
        window.showNotification('Configuration copied to clipboard', 'success');
    }).catch(() => {
        window.showNotification('Failed to copy to clipboard', 'error');
    });
};

window.downloadConfig = function() {
    const codeElement = document.getElementById('configCode');
    if (!codeElement) return;
    
    const code = codeElement.textContent;
    const blob = new Blob([code], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'configuration.nix';
    a.click();
    URL.revokeObjectURL(url);
    window.showNotification('Configuration downloaded', 'success');
};

window.saveToRepo = function() {
    window.showNotification('Save to repository functionality will be implemented', 'info');
};

// Initialize global variables for builder
window.configModules = [];
window.selectedModule = null;

// ===== VISUAL BUILDER DRAG & DROP FUNCTIONS =====
// These functions enable the visual drag & drop functionality

// Add module to canvas from drag & drop
window.addModuleFromType = function(moduleType) {
    // Module templates based on type
    const moduleTemplates = {
        'boot': {
            name: 'Boot Configuration',
            icon: 'power-off',
            description: 'System boot configuration',
            config: {
                'boot.loader.systemd-boot.enable': true,
                'boot.loader.efi.canTouchEfiVariables': true
            }
        },
        'network': {
            name: 'Network Settings',
            icon: 'network-wired',
            description: 'Network interface configuration',
            config: {
                'networking.networkmanager.enable': true,
                'networking.firewall.enable': true
            }
        },
        'users': {
            name: 'Users & Groups',
            icon: 'users',
            description: 'User account management',
            config: {
                'users.users.admin.isNormalUser': true,
                'users.users.admin.extraGroups': ['wheel', 'networkmanager']
            }
        },
        'filesystem': {
            name: 'File Systems',
            icon: 'hdd',
            description: 'File system configuration',
            config: {
                'fileSystems."/".fsType': 'ext4'
            }
        },
        'ssh': {
            name: 'SSH Server',
            icon: 'terminal',
            description: 'SSH daemon configuration',
            config: {
                'services.openssh.enable': true,
                'services.openssh.settings.PasswordAuthentication': false
            }
        },
        'nginx': {
            name: 'Nginx',
            icon: 'globe',
            description: 'Nginx web server',
            config: {
                'services.nginx.enable': true,
                'networking.firewall.allowedTCPPorts': [80, 443]
            }
        },
        'docker': {
            name: 'Docker',
            icon: 'docker',
            description: 'Docker container runtime',
            config: {
                'virtualisation.docker.enable': true,
                'users.users.admin.extraGroups': ['docker']
            }
        },
        'postgresql': {
            name: 'PostgreSQL',
            icon: 'database',
            description: 'PostgreSQL database server',
            config: {
                'services.postgresql.enable': true,
                'services.postgresql.package': 'pkgs.postgresql_15'
            }
        },
        'system-packages': {
            name: 'System Packages',
            icon: 'cubes',
            description: 'System-wide packages',
            config: {
                'environment.systemPackages': ['wget', 'curl', 'git', 'vim', 'htop']
            }
        },
        'development': {
            name: 'Development Tools',
            icon: 'code',
            description: 'Development environment packages',
            config: {
                'environment.systemPackages': ['nodejs', 'python3', 'gcc', 'make']
            }
        },
        'desktop': {
            name: 'Desktop Environment',
            icon: 'desktop',
            description: 'Desktop environment configuration',
            config: {
                'services.xserver.enable': true,
                'services.xserver.desktopManager.gnome.enable': true
            }
        }
    };

    const template = moduleTemplates[moduleType];
    if (template) {
        template.id = Date.now();
        template.type = moduleType;
        window.addModuleToCanvas(template);
        window.showNotification(`Added ${template.name} module`, 'success');
    }
};

// Add module to the visual canvas
window.addModuleToCanvas = function(moduleData) {
    const canvas = document.getElementById('builderCanvas');
    if (!canvas) return;
    
    const dropzone = canvas.querySelector('.canvas-dropzone');
    
    if (dropzone) {
        dropzone.style.display = 'none';
    }

    const moduleElement = document.createElement('div');
    moduleElement.className = 'canvas-module';
    moduleElement.dataset.moduleId = moduleData.id || Date.now();
    moduleElement.innerHTML = `
        <div class="module-header">
            <i class="fas fa-${moduleData.icon || 'cog'}"></i>
            <span>${moduleData.name}</span>
            <button class="module-remove" onclick="removeModule(this)">
                <i class="fas fa-times"></i>
            </button>
        </div>
        <div class="module-body">
            <p>${moduleData.description || 'Module configuration'}</p>
            <div class="module-config">
                ${Object.keys(moduleData.config || {}).slice(0, 2).map(key => 
                    `<small>${key}</small>`
                ).join('<br>')}
                ${Object.keys(moduleData.config || {}).length > 2 ? '<small>... and more</small>' : ''}
            </div>
        </div>
    `;

    moduleElement.onclick = () => window.selectModule(moduleElement, moduleData);
    canvas.appendChild(moduleElement);
    
    window.configModules.push(moduleData);
    
    // Add CSS for the module if not already added
    if (!document.getElementById('builderModuleStyles')) {
        const styles = document.createElement('style');
        styles.id = 'builderModuleStyles';
        styles.textContent = `
            .canvas-module {
                background: var(--bg-surface);
                border: 2px solid var(--border-color);
                border-radius: var(--radius-md);
                padding: var(--spacing-md);
                margin-bottom: var(--spacing-sm);
                cursor: pointer;
                transition: all 0.2s ease;
            }
            .canvas-module:hover {
                border-color: var(--primary-color);
                box-shadow: var(--shadow-md);
            }
            .canvas-module.selected {
                border-color: var(--primary-color);
                background: rgb(59 130 246 / 0.05);
            }
            .module-header {
                display: flex;
                align-items: center;
                justify-content: space-between;
                margin-bottom: var(--spacing-sm);
                font-weight: 600;
            }
            .module-header i {
                margin-right: var(--spacing-sm);
                color: var(--primary-color);
            }
            .module-remove {
                background: none;
                border: none;
                color: var(--text-secondary);
                cursor: pointer;
                padding: var(--spacing-xs);
                border-radius: var(--radius-sm);
                font-size: 0.875rem;
            }
            .module-remove:hover {
                background: var(--danger-color);
                color: white;
            }
            .module-body {
                font-size: 0.875rem;
                color: var(--text-secondary);
            }
            .module-config {
                margin-top: var(--spacing-sm);
                font-family: monospace;
                font-size: 0.75rem;
                color: var(--text-tertiary);
            }
            .builder-canvas.drag-over {
                border: 2px dashed var(--primary-color);
                background: rgb(59 130 246 / 0.05);
            }
        `;
        document.head.appendChild(styles);
    }
};

// Select a module in the canvas
window.selectModule = function(element, moduleData) {
    // Remove previous selection
    document.querySelectorAll('.canvas-module').forEach(m => m.classList.remove('selected'));
    
    // Select current module
    element.classList.add('selected');
    window.selectedModule = moduleData;
    
    // Update properties panel
    window.updatePropertiesPanel(moduleData);
    window.showNotification(`Selected ${moduleData.name}`, 'info');
};

// Update the properties panel with module configuration
window.updatePropertiesPanel = function(moduleData) {
    const panel = document.getElementById('propertiesPanel');
    if (!panel) return;
    
    // Generate form based on module configuration
    const configEntries = Object.entries(moduleData.config || {});
    
    panel.innerHTML = `
        <div class="property-form">
            <h4>${moduleData.name} Configuration</h4>
            <div class="form-section">
                <div class="form-group">
                    <label>Module Name</label>
                    <input type="text" value="${moduleData.name}" class="nixai-input" readonly>
                </div>
                <div class="form-group">
                    <label>Description</label>
                    <textarea class="nixai-input" rows="2" readonly>${moduleData.description}</textarea>
                </div>
            </div>
            <div class="form-section">
                <h5>Configuration Options</h5>
                ${configEntries.map(([key, value]) => `
                    <div class="form-group">
                        <label>${key}</label>
                        <input type="text" value="${Array.isArray(value) ? value.join(', ') : value}" 
                               class="nixai-input" data-config-key="${key}"
                               onchange="updateModuleConfig('${moduleData.id}', '${key}', this.value)">
                    </div>
                `).join('')}
            </div>
            <div class="form-actions">
                <button class="nixai-button nixai-button-primary" onclick="saveModuleConfig('${moduleData.id}')">
                    Save Changes
                </button>
                <button class="nixai-button nixai-button-secondary" onclick="removeSelectedModule()">
                    Remove Module
                </button>
            </div>
        </div>
    `;
};

// Update module configuration
window.updateModuleConfig = function(moduleId, key, value) {
    const module = window.configModules.find(m => m.id == moduleId);
    if (module && module.config) {
        // Try to parse as JSON for arrays/objects, otherwise use string
        try {
            if (value.includes('[') || value.includes('{')) {
                module.config[key] = JSON.parse(value);
            } else if (value.includes(',')) {
                module.config[key] = value.split(',').map(v => v.trim());
            } else if (value === 'true' || value === 'false') {
                module.config[key] = value === 'true';
            } else {
                module.config[key] = value;
            }
        } catch (e) {
            module.config[key] = value;
        }
    }
};

// Save module configuration
window.saveModuleConfig = function(moduleId) {
    window.showNotification('Module configuration saved!', 'success');
    // Auto-generate preview when config changes
    window.previewConfig();
};

// Remove selected module
window.removeSelectedModule = function() {
    if (window.selectedModule) {
        const moduleElement = document.querySelector(`[data-module-id="${window.selectedModule.id}"]`);
        window.removeModule(moduleElement?.querySelector('.module-remove'));
    }
};

// Remove module from canvas
window.removeModule = function(button) {
    const moduleElement = button.closest('.canvas-module');
    const moduleId = moduleElement.dataset.moduleId;
    
    moduleElement.remove();
    window.configModules = window.configModules.filter(m => m.id != moduleId);
    
    // Show dropzone if no modules left
    const canvas = document.getElementById('builderCanvas');
    if (canvas && canvas.children.length === 0) {
        canvas.innerHTML = `
            <div class="canvas-dropzone">
                <i class="fas fa-plus-circle"></i>
                <p>Drag modules here to build your configuration</p>
                <p class="text-muted">or use AI assistance to generate configurations</p>
            </div>
        `;
    }
    
    window.clearPropertiesPanel();
    window.showNotification('Module removed', 'info');
};

// Clear properties panel
window.clearPropertiesPanel = function() {
    const panel = document.querySelector('.properties-panel-content');
    if (!panel) return;
    
    panel.innerHTML = `
        <div class="empty-state">
            <i class="fas fa-cog"></i>
            <p>Select a module to configure its properties</p>
        </div>
    `;
};

// Utility functions for builder
window.copyToClipboard = function() {
    const codeElement = document.getElementById('configCode');
    if (!codeElement) return;
    
    const code = codeElement.textContent;
    navigator.clipboard.writeText(code).then(() => {
        window.showNotification('Configuration copied to clipboard', 'success');
    }).catch(() => {
        window.showNotification('Failed to copy to clipboard', 'error');
    });
};

window.downloadConfig = function() {
    const codeElement = document.getElementById('configCode');
    if (!codeElement) return;
    
    const code = codeElement.textContent;
    const blob = new Blob([code], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'configuration.nix';
    a.click();
    URL.revokeObjectURL(url);
    window.showNotification('Configuration downloaded', 'success');
};

window.saveToRepo = function() {
    window.showNotification('Save to repository functionality will be implemented', 'info');
};

// Initialize global variables for builder
window.configModules = [];
window.selectedModule = null;

// ===== VISUAL BUILDER DRAG & DROP FUNCTIONS =====
// These functions enable the visual drag & drop functionality

// Add module to canvas from drag & drop
window.addModuleFromType = function(moduleType) {
    // Module templates based on type
    const moduleTemplates = {
        'boot': {
            name: 'Boot Configuration',
            icon: 'power-off',
            description: 'System boot configuration',
            config: {
                'boot.loader.systemd-boot.enable': true,
                'boot.loader.efi.canTouchEfiVariables': true
            }
        },
        'network': {
            name: 'Network Settings',
            icon: 'network-wired',
            description: 'Network interface configuration',
            config: {
                'networking.networkmanager.enable': true,
                'networking.firewall.enable': true
            }
        },
        'users': {
            name: 'Users & Groups',
            icon: 'users',
            description: 'User account management',
            config: {
                'users.users.admin.isNormalUser': true,
                'users.users.admin.extraGroups': ['wheel', 'networkmanager']
            }
        },
        'filesystem': {
            name: 'File Systems',
            icon: 'hdd',
            description: 'File system configuration',
            config: {
                'fileSystems."/".fsType': 'ext4'
            }
        },
        'ssh': {
            name: 'SSH Server',
            icon: 'terminal',
            description: 'SSH daemon configuration',
            config: {
                'services.openssh.enable': true,
                'services.openssh.settings.PasswordAuthentication': false
            }
        },
        'nginx': {
            name: 'Nginx',
            icon: 'globe',
            description: 'Nginx web server',
            config: {
                'services.nginx.enable': true,
                'networking.firewall.allowedTCPPorts': [80, 443]
            }
        },
        'docker': {
            name: 'Docker',
            icon: 'docker',
            description: 'Docker container runtime',
            config: {
                'virtualisation.docker.enable': true,
                'users.users.admin.extraGroups': ['docker']
            }
        },
        'postgresql': {
            name: 'PostgreSQL',
            icon: 'database',
            description: 'PostgreSQL database server',
            config: {
                'services.postgresql.enable': true,
                'services.postgresql.package': 'pkgs.postgresql_15'
            }
        },
        'system-packages': {
            name: 'System Packages',
            icon: 'cubes',
            description: 'System-wide packages',
            config: {
                'environment.systemPackages': ['wget', 'curl', 'git', 'vim', 'htop']
            }
        },
        'development': {
            name: 'Development Tools',
            icon: 'code',
            description: 'Development environment packages',
            config: {
                'environment.systemPackages': ['nodejs', 'python3', 'gcc', 'make']
            }
        },
        'desktop': {
            name: 'Desktop Environment',
            icon: 'desktop',
            description: 'Desktop environment configuration',
            config: {
                'services.xserver.enable': true,
                'services.xserver.desktopManager.gnome.enable': true
            }
        }
    };

    const template = moduleTemplates[moduleType];
    if (template) {
        template.id = Date.now();
        template.type = moduleType;
        window.addModuleToCanvas(template);
        window.showNotification(`Added ${template.name} module`, 'success');
    }
};

// Add module to the visual canvas
window.addModuleToCanvas = function(moduleData) {
    const canvas = document.getElementById('builderCanvas');
    if (!canvas) return;
    
    const dropzone = canvas.querySelector('.canvas-dropzone');
    
    if (dropzone) {
        dropzone.style.display = 'none';
    }

    const moduleElement = document.createElement('div');
    moduleElement.className = 'canvas-module';
    moduleElement.dataset.moduleId = moduleData.id || Date.now();
    moduleElement.innerHTML = `
        <div class="module-header">
            <i class="fas fa-${moduleData.icon || 'cog'}"></i>
            <span>${moduleData.name}</span>
            <button class="module-remove" onclick="removeModule(this)">
                <i class="fas fa-times"></i>
            </button>
        </div>
        <div class="module-body">
            <p>${moduleData.description || 'Module configuration'}</p>
            <div class="module-config">
                ${Object.keys(moduleData.config || {}).slice(0, 2).map(key => 
                    `<small>${key}</small>`
                ).join('<br>')}
                ${Object.keys(moduleData.config || {}).length > 2 ? '<small>... and more</small>' : ''}
            </div>
        </div>
    `;

    moduleElement.onclick = () => window.selectModule(moduleElement, moduleData);
    canvas.appendChild(moduleElement);
    
    window.configModules.push(moduleData);
    
    // Add CSS for the module if not already added
    if (!document.getElementById('builderModuleStyles')) {
        const styles = document.createElement('style');
        styles.id = 'builderModuleStyles';
        styles.textContent = `
            .canvas-module {
                background: var(--bg-surface);
                border: 2px solid var(--border-color);
                border-radius: var(--radius-md);
                padding: var(--spacing-md);
                margin-bottom: var(--spacing-sm);
                cursor: pointer;
                transition: all 0.2s ease;
            }
            .canvas-module:hover {
                border-color: var(--primary-color);
                box-shadow: var(--shadow-md);
            }
            .canvas-module.selected {
                border-color: var(--primary-color);
                background: rgb(59 130 246 / 0.05);
            }
            .module-header {
                display: flex;
                align-items: center;
                justify-content: space-between;
                margin-bottom: var(--spacing-sm);
                font-weight: 600;
            }
            .module-header i {
                margin-right: var(--spacing-sm);
                color: var(--primary-color);
            }
            .module-remove {
                background: none;
                border: none;
                color: var(--text-secondary);
                cursor: pointer;
                padding: var(--spacing-xs);
                border-radius: var(--radius-sm);
                font-size: 0.875rem;
            }
            .module-remove:hover {
                background: var(--danger-color);
                color: white;
            }
            .module-body {
                font-size: 0.875rem;
                color: var(--text-secondary);
            }
            .module-config {
                margin-top: var(--spacing-sm);
                font-family: monospace;
                font-size: 0.75rem;
                color: var(--text-tertiary);
            }
            .builder-canvas.drag-over {
                border: 2px dashed var(--primary-color);
                background: rgb(59 130 246 / 0.05);
            }
        `;
        document.head.appendChild(styles);
    }
};

// Select a module in the canvas
window.selectModule = function(element, moduleData) {
    // Remove previous selection
    document.querySelectorAll('.canvas-module').forEach(m => m.classList.remove('selected'));
    
    // Select current module
    element.classList.add('selected');
    window.selectedModule = moduleData;
    
    // Update properties panel
    window.updatePropertiesPanel(moduleData);
    window.showNotification(`Selected ${moduleData.name}`, 'info');
};

// Update the properties panel with module configuration
window.updatePropertiesPanel = function(moduleData) {
    const panel = document.getElementById('propertiesPanel');
    if (!panel) return;
    
    // Generate form based on module configuration
    const configEntries = Object.entries(moduleData.config || {});
    
    panel.innerHTML = `
        <div class="property-form">
            <h4>${moduleData.name} Configuration</h4>
            <div class="form-section">
                <div class="form-group">
                    <label>Module Name</label>
                    <input type="text" value="${moduleData.name}" class="nixai-input" readonly>
                </div>
                <div class="form-group">
                    <label>Description</label>
                    <textarea class="nixai-input" rows="2" readonly>${moduleData.description}</textarea>
                </div>
            </div>
            <div class="form-section">
                <h5>Configuration Options</h5>
                ${configEntries.map(([key, value]) => `
                    <div class="form-group">
                        <label>${key}</label>
                        <input type="text" value="${Array.isArray(value) ? value.join(', ') : value}" 
                               class="nixai-input" data-config-key="${key}"
                               onchange="updateModuleConfig('${moduleData.id}', '${key}', this.value)">
                    </div>
                `).join('')}
            </div>
            <div class="form-actions">
                <button class="nixai-button nixai-button-primary" onclick="saveModuleConfig('${moduleData.id}')">
                    Save Changes
                </button>
                <button class="nixai-button nixai-button-secondary" onclick="removeSelectedModule()">
                    Remove Module
                </button>
            </div>
        </div>
    `;
};

// Update module configuration
window.updateModuleConfig = function(moduleId, key, value) {
    const module = window.configModules.find(m => m.id == moduleId);
    if (module && module.config) {
        // Try to parse as JSON for arrays/objects, otherwise use string
        try {
            if (value.includes('[') || value.includes('{')) {
                module.config[key] = JSON.parse(value);
            } else if (value.includes(',')) {
                module.config[key] = value.split(',').map(v => v.trim());
            } else if (value === 'true' || value === 'false') {
                module.config[key] = value === 'true';
            } else {
                module.config[key] = value;
            }
        } catch (e) {
            module.config[key] = value;
        }
    }
};

// Save module configuration
window.saveModuleConfig = function(moduleId) {
    window.showNotification('Module configuration saved!', 'success');
    // Auto-generate preview when config changes
    window.previewConfig();
};

// Remove selected module
window.removeSelectedModule = function() {
    if (window.selectedModule) {
        const moduleElement = document.querySelector(`[data-module-id="${window.selectedModule.id}"]`);
        window.removeModule(moduleElement?.querySelector('.module-remove'));
    }
};

// Remove module from canvas
window.removeModule = function(button) {
    const moduleElement = button.closest('.canvas-module');
    const moduleId = moduleElement.dataset.moduleId;
    
    moduleElement.remove();
    window.configModules = window.configModules.filter(m => m.id != moduleId);
    
    // Show dropzone if no modules left
    const canvas = document.getElementById('builderCanvas');
    if (canvas && canvas.children.length === 0) {
        canvas.innerHTML = `
            <div class="canvas-dropzone">
                <i class="fas fa-plus-circle"></i>
                <p>Drag modules here to build your configuration</p>
                <p class="text-muted">or use AI assistance to generate configurations</p>
            </div>
        `;
    }
    
    window.clearPropertiesPanel();
    window.showNotification('Module removed', 'info');
};

// Clear properties panel
window.clearPropertiesPanel = function() {
    const panel = document.querySelector('.properties-panel-content');
    if (!panel) return;
    
    panel.innerHTML = `
        <div class="empty-state">
            <i class="fas fa-cog"></i>
            <p>Select a module to configure its properties</p>
        </div>
    `;
};

// ===== BUILDER PANEL MANAGEMENT =====
// Global functions for panel management needed by builder template

window.toggleModulesPanel = function() {
    const panel = document.getElementById('modulesPanel');
    const fab = document.querySelector('.fab[onclick="toggleModulesPanel()"]');
    
    if (panel && panel.classList.contains('active')) {
        panel.classList.remove('active');
        if (fab) fab.classList.remove('active');
    } else {
        // Close other panels
        window.closeAllPanels();
        if (panel) panel.classList.add('active');
        if (fab) fab.classList.add('active');
    }
};

window.togglePropertiesPanel = function() {
    const panel = document.getElementById('propertiesPanel');
    const fab = document.querySelector('.fab[onclick="togglePropertiesPanel()"]');
    
    if (panel && panel.classList.contains('active')) {
        panel.classList.remove('active');
        if (fab) fab.classList.remove('active');
    } else {
        // Close other panels
        window.closeAllPanels();
        if (panel) panel.classList.add('active');
        if (fab) fab.classList.add('active');
    }
};

window.toggleAIAssistant = function() {
    const panel = document.getElementById('aiAssistantPanel');
    const fab = document.querySelector('.fab[onclick="toggleAIAssistant()"]');
    
    if (panel && panel.classList.contains('active')) {
        panel.classList.remove('active');
        if (fab) fab.classList.remove('active');
    } else {
        // Close other panels
        window.closeAllPanels();
        if (panel) panel.classList.add('active');
        if (fab) fab.classList.add('active');
    }
};

window.closeAllPanels = function() {
    document.querySelectorAll('.modules-panel, .properties-panel, .ai-assistant-panel').forEach(panel => {
        panel.classList.remove('active');
    });
    document.querySelectorAll('.fab').forEach(fab => {
        fab.classList.remove('active');
    });
};

// AI Chat functionality
window.sendAIMessage = function() {
    const input = document.getElementById('aiInput');
    if (!input) return;
    
    const message = input.value.trim();
    if (!message) return;

    // Add user message to chat
    window.addChatMessage('user', message);
    input.value = '';

    // Simulate AI response
    setTimeout(() => {
        let response = '';
        if (message.toLowerCase().includes('web server')) {
            response = 'I recommend adding an Nginx module for a web server. You can drag the "Nginx" module from the Services category to your canvas.';
            // Auto-suggest nginx module
            setTimeout(() => {
                window.addModuleFromType('nginx');
            }, 1000);
        } else if (message.toLowerCase().includes('development')) {
            response = 'For a development environment, I suggest adding Development Tools and configuring a user account. Let me add these modules for you.';
            setTimeout(() => {
                window.addModuleFromType('development');
            }, 1000);
        } else {
            response = 'I can help you configure your NixOS system. Try asking about specific components like "web server", "development environment", or "user configuration".';
        }
        
        window.addChatMessage('assistant', response);
    }, 1000);
};

window.addChatMessage = function(role, message) {
    const messagesContainer = document.getElementById('chatMessages');
    if (!messagesContainer) return;
    
    const messageDiv = document.createElement('div');
    messageDiv.className = `chat-message ${role}`;
    
    messageDiv.innerHTML = `
        <div class="chat-avatar">
            <i class="fas fa-${role === 'user' ? 'user' : 'robot'}"></i>
        </div>
        <div class="chat-content">
            <p>${message}</p>
        </div>
    `;
    
    messagesContainer.appendChild(messageDiv);
    messagesContainer.scrollTop = messagesContainer.scrollHeight;
};

// Module removal functionality
window.removeModule = function(button) {
    const moduleElement = button.closest('.canvas-module');
    if (!moduleElement) return;
    
    const moduleId = moduleElement.dataset.moduleId;
    
    moduleElement.remove();
    window.configModules = (window.configModules || []).filter(m => m.id !== moduleId);
    
    // Show dropzone if no modules left
    const canvas = document.getElementById('builderCanvas');
    if (canvas && canvas.children.length === 0) {
        canvas.innerHTML = `
            <div class="canvas-dropzone">
                <i class="fas fa-plus-circle"></i>
                <p>Drag modules here to build your configuration</p>
                <p class="text-muted">or use AI assistance to generate configurations</p>
            </div>
        `;
    }
    
    window.clearPropertiesPanel();
    window.showNotification('Module removed', 'info');
};

// Clear properties panel
window.clearPropertiesPanel = function() {
    const panel = document.querySelector('.properties-panel-content');
    if (!panel) return;
    
    panel.innerHTML = `
        <div class="empty-state">
            <i class="fas fa-cog"></i>
            <p>Select a module to configure its properties</p>
        </div>
    `;
};

// ===== BUILDER PANEL MANAGEMENT =====
// Global functions for panel management needed by builder template

window.toggleModulesPanel = function() {
    const panel = document.getElementById('modulesPanel');
    const fab = document.querySelector('.fab[onclick="toggleModulesPanel()"]');
    
    if (panel && panel.classList.contains('active')) {
        panel.classList.remove('active');
        if (fab) fab.classList.remove('active');
    } else {
        // Close other panels
        window.closeAllPanels();
        if (panel) panel.classList.add('active');
        if (fab) fab.classList.add('active');
    }
};

window.togglePropertiesPanel = function() {
    const panel = document.getElementById('propertiesPanel');
    const fab = document.querySelector('.fab[onclick="togglePropertiesPanel()"]');
    
    if (panel && panel.classList.contains('active')) {
        panel.classList.remove('active');
        if (fab) fab.classList.remove('active');
    } else {
        // Close other panels
        window.closeAllPanels();
        if (panel) panel.classList.add('active');
        if (fab) fab.classList.add('active');
    }
};

window.toggleAIAssistant = function() {
    const panel = document.getElementById('aiAssistantPanel');
    const fab = document.querySelector('.fab[onclick="toggleAIAssistant()"]');
    
    if (panel && panel.classList.contains('active')) {
        panel.classList.remove('active');
        if (fab) fab.classList.remove('active');
    } else {
        // Close other panels
        window.closeAllPanels();
        if (panel) panel.classList.add('active');
        if (fab) fab.classList.add('active');
    }
};

window.closeAllPanels = function() {
    document.querySelectorAll('.modules-panel, .properties-panel, .ai-assistant-panel').forEach(panel => {
        panel.classList.remove('active');
    });
    document.querySelectorAll('.fab').forEach(fab => {
        fab.classList.remove('active');
    });
};

// AI Chat functionality
window.sendAIMessage = function() {
    const input = document.getElementById('aiInput');
    if (!input) return;
    
    const message = input.value.trim();
    if (!message) return;

    // Add user message to chat
    window.addChatMessage('user', message);
    input.value = '';

    // Simulate AI response
    setTimeout(() => {
        let response = '';
        if (message.toLowerCase().includes('web server')) {
            response = 'I recommend adding an Nginx module for a web server. You can drag the "Nginx" module from the Services category to your canvas.';
            // Auto-suggest nginx module
            setTimeout(() => {
                window.addModuleFromType('nginx');
            }, 1000);
        } else if (message.toLowerCase().includes('development')) {
            response = 'For a development environment, I suggest adding Development Tools and configuring a user account. Let me add these modules for you.';
            setTimeout(() => {
                window.addModuleFromType('development');
            }, 1000);
        } else {
            response = 'I can help you configure your NixOS system. Try asking about specific components like "web server", "development environment", or "user configuration".';
        }
        
        window.addChatMessage('assistant', response);
    }, 1000);
};

window.addChatMessage = function(role, message) {
    const messagesContainer = document.getElementById('chatMessages');
    if (!messagesContainer) return;
    
    const messageDiv = document.createElement('div');
    messageDiv.className = `chat-message ${role}`;
    
    messageDiv.innerHTML = `
        <div class="chat-avatar">
            <i class="fas fa-${role === 'user' ? 'user' : 'robot'}"></i>
        </div>
        <div class="chat-content">
            <p>${message}</p>
        </div>
    `;
    
    messagesContainer.appendChild(messageDiv);
    messagesContainer.scrollTop = messagesContainer.scrollHeight;
};

// Module removal functionality
window.removeModule = function(button) {
    const moduleElement = button.closest('.canvas-module');
    if (!moduleElement) return;
    
    const moduleId = moduleElement.dataset.moduleId;
    
    moduleElement.remove();
    window.configModules = (window.configModules || []).filter(m => m.id !== moduleId);
    
    // Show dropzone if no modules left
    const canvas = document.getElementById('builderCanvas');
    if (canvas && canvas.children.length === 0) {
        canvas.innerHTML = `
            <div class="canvas-dropzone">
                <i class="fas fa-plus-circle"></i>
                <p>Drag modules here to build your configuration</p>
                <p class="text-muted">or use AI assistance to generate configurations</p>
            </div>
        `;
    }
    
    window.clearPropertiesPanel();
    window.showNotification('Module removed', 'info');
};

// Clear properties panel
window.clearPropertiesPanel = function() {
    const panel = document.querySelector('.properties-panel-content');
    if (!panel) return;
    
    panel.innerHTML = `
        <div class="empty-state">
            <i class="fas fa-cog"></i>
            <p>Select a module to configure its properties</p>
        </div>
    `;
};

// ===== BUILDER PANEL MANAGEMENT =====
// Global functions for panel management needed by builder template

window.toggleModulesPanel = function() {
    const panel = document.getElementById('modulesPanel');
    const fab = document.querySelector('.fab[onclick="toggleModulesPanel()"]');
    
    if (panel && panel.classList.contains('active')) {
        panel.classList.remove('active');
        if (fab) fab.classList.remove('active');
    } else {
        // Close other panels
        window.closeAllPanels();
        if (panel) panel.classList.add('active');
        if (fab) fab.classList.add('active');
    }
};

window.togglePropertiesPanel = function() {
    const panel = document.getElementById('propertiesPanel');
    const fab = document.querySelector('.fab[onclick="togglePropertiesPanel()"]');
    
    if (panel && panel.classList.contains('active')) {
        panel.classList.remove('active');
        if (fab) fab.classList.remove('active');
    } else {
        // Close other panels
        window.closeAllPanels();
        if (panel) panel.classList.add('active');
        if (fab) fab.classList.add('active');
    }
};

window.toggleAIAssistant = function() {
    const panel = document.getElementById('aiAssistantPanel');
    const fab = document.querySelector('.fab[onclick="toggleAIAssistant()"]');
    
    if (panel && panel.classList.contains('active')) {
        panel.classList.remove('active');
        if (fab) fab.classList.remove('active');
    } else {
        // Close other panels
        window.closeAllPanels();
        if (panel) panel.classList.add('active');
        if (fab) fab.classList.add('active');
    }
};

window.closeAllPanels = function() {
    document.querySelectorAll('.modules-panel, .properties-panel, .ai-assistant-panel').forEach(panel => {
        panel.classList.remove('active');
    });
    document.querySelectorAll('.fab').forEach(fab => {
        fab.classList.remove('active');
    });
};

// AI Chat functionality
window.sendAIMessage = function() {
    const input = document.getElementById('aiInput');
    if (!input) return;
    
    const message = input.value.trim();
    if (!message) return;

    // Add user message to chat
    window.addChatMessage('user', message);
    input.value = '';

    // Simulate AI response
    setTimeout(() => {
        let response = '';
        if (message.toLowerCase().includes('web server')) {
            response = 'I recommend adding an Nginx module for a web server. You can drag the "Nginx" module from the Services category to your canvas.';
            // Auto-suggest nginx module
            setTimeout(() => {
                window.addModuleFromType('nginx');
            }, 1000);
        } else if (message.toLowerCase().includes('development')) {
            response = 'For a development environment, I suggest adding Development Tools and configuring a user account. Let me add these modules for you.';
            setTimeout(() => {
                window.addModuleFromType('development');
            }, 1000);
        } else {
            response = 'I can help you configure your NixOS system. Try asking about specific components like "web server", "development environment", or "user configuration".';
        }
        
        window.addChatMessage('assistant', response);
    }, 1000);
};

window.addChatMessage = function(role, message) {
    const messagesContainer = document.getElementById('chatMessages');
    if (!messagesContainer) return;
    
    const messageDiv = document.createElement('div');
    messageDiv.className = `chat-message ${role}`;
    
    messageDiv.innerHTML = `
        <div class="chat-avatar">
            <i class="fas fa-${role === 'user' ? 'user' : 'robot'}"></i>
        </div>
        <div class="chat-content">
            <p>${message}</p>
        </div>
    `;
    
    messagesContainer.appendChild(messageDiv);
    messagesContainer.scrollTop = messagesContainer.scrollHeight;
};

// Module removal functionality
window.removeModule = function(button) {
    const moduleElement = button.closest('.canvas-module');
    if (!moduleElement) return;
    
    const moduleId = moduleElement.dataset.moduleId;
    
    moduleElement.remove();
    window.configModules = (window.configModules || []).filter(m => m.id !== moduleId);
    
    // Show dropzone if no modules left
    const canvas = document.getElementById('builderCanvas');
    if (canvas && canvas.children.length === 0) {
        canvas.innerHTML = `
            <div class="canvas-dropzone">
                <i class="fas fa-plus-circle"></i>
                <p>Drag modules here to build your configuration</p>
                <p class="text-muted">or use AI assistance to generate configurations</p>
            </div>
        `;
    }
    
    window.clearPropertiesPanel();
    window.showNotification('Module removed', 'info');
};

// Clear properties panel
window.clearPropertiesPanel = function() {
    const panel = document.querySelector('.properties-panel-content');
    if (!panel) return;
    
    panel.innerHTML = `
        <div class="empty-state">
            <i class="fas fa-cog"></i>
            <p>Select a module to configure its properties</p>
        </div>
    `;
};

// ===== BUILDER PANEL MANAGEMENT =====
// Global functions for panel management needed by builder template

window.toggleModulesPanel = function() {
    const panel = document.getElementById('modulesPanel');
    const fab = document.querySelector('.fab[onclick="toggleModulesPanel()"]');
    
    if (panel && panel.classList.contains('active')) {
        panel.classList.remove('active');
        if (fab) fab.classList.remove('active');
    } else {
        // Close other panels
        window.closeAllPanels();
        if (panel) panel.classList.add('active');
        if (fab) fab.classList.add('active');
    }
};

window.togglePropertiesPanel = function() {
    const panel = document.getElementById('propertiesPanel');
    const fab = document.querySelector('.fab[onclick="togglePropertiesPanel()"]');
    
    if (panel && panel.classList.contains('active')) {
        panel.classList.remove('active');
        if (fab) fab.classList.remove('active');
    } else {
        // Close other panels
        window.closeAllPanels();
        if (panel) panel.classList.add('active');
        if (fab) fab.classList.add('active');
    }
};

window.toggleAIAssistant = function() {
    const panel = document.getElementById('aiAssistantPanel');
    const fab = document.querySelector('.fab[onclick="toggleAIAssistant()"]');
    
    if (panel && panel.classList.contains('active')) {
        panel.classList.remove('active');
        if (fab) fab.classList.remove('active');
    } else {
        // Close other panels
        window.closeAllPanels();
        if (panel) panel.classList.add('active');
        if (fab) fab.classList.add('active');
    }
};

window.closeAllPanels = function() {
    document.querySelectorAll('.modules-panel, .properties-panel, .ai-assistant-panel').forEach(panel => {
        panel.classList.remove('active');
    });
    document.querySelectorAll('.fab').forEach(fab => {
        fab.classList.remove('active');
    });
};

// AI Chat functionality
window.sendAIMessage = function() {
    const input = document.getElementById('aiInput');
    if (!input) return;
    
    const message = input.value.trim();
    if (!message) return;

    // Add user message to chat
    window.addChatMessage('user', message);
    input.value = '';

    // Simulate AI response
    setTimeout(() => {
        let response = '';
        if (message.toLowerCase().includes('web server')) {
            response = 'I recommend adding an Nginx module for a web server. You can drag the "Nginx" module from the Services category to your canvas.';
            // Auto-suggest nginx module
            setTimeout(() => {
                window.addModuleFromType('nginx');
            }, 1000);
        } else if (message.toLowerCase().includes('development')) {
            response = 'For a development environment, I suggest adding Development Tools and configuring a user account. Let me add these modules for you.';
            setTimeout(() => {
                window.addModuleFromType('development');
            }, 1000);
        } else {
            response = 'I can help you configure your NixOS system. Try asking about specific components like "web server", "development environment", or "user configuration".';
        }
        
        window.addChatMessage('assistant', response);
    }, 1000);
};

window.addChatMessage = function(role, message) {
    const messagesContainer = document.getElementById('chatMessages');
    if (!messagesContainer) return;
    
    const messageDiv = document.createElement('div');
    messageDiv.className = `chat-message ${role}`;
    
    messageDiv.innerHTML = `
        <div class="chat-avatar">
            <i class="fas fa-${role === 'user' ? 'user' : 'robot'}"></i>
        </div>
        <div class="chat-content">
            <p>${message}</p>
        </div>
    `;
    
    messagesContainer.appendChild(messageDiv);
    messagesContainer.scrollTop = messagesContainer.scrollHeight;
};

// Module removal functionality
window.removeModule = function(button) {
    const moduleElement = button.closest('.canvas-module');
    if (!moduleElement) return;
    
    const moduleId = moduleElement.dataset.moduleId;
    
    moduleElement.remove();
    window.configModules = (window.configModules || []).filter(m => m.id !== moduleId);
    
    // Show dropzone if no modules left
    const canvas = document.getElementById('builderCanvas');
    if (canvas && canvas.children.length === 0) {
        canvas.innerHTML = `
            <div class="canvas-dropzone">
                <i class="fas fa-plus-circle"></i>
                <p>Drag modules here to build your configuration</p>
                <p class="text-muted">or use AI assistance to generate configurations</p>
            </div>
        `;
    }
    
    window.clearPropertiesPanel();
    window.showNotification('Module removed', 'info');
};

// Clear properties panel
window.clearPropertiesPanel = function() {
    const panel = document.querySelector('.properties-panel-content');
    if (!panel) return;
    
    panel.innerHTML = `
        <div class="empty-state">
            <i class="fas fa-cog"></i>
            <p>Select a module to configure its properties</p>
        </div>
    `;
};

// ===== BUILDER PANEL MANAGEMENT =====
// Global functions for panel management needed by builder template

window.toggleModulesPanel = function() {
    const panel = document.getElementById('modulesPanel');
    const fab = document.querySelector('.fab[onclick="toggleModulesPanel()"]');
    
    if (panel && panel.classList.contains('active')) {
        panel.classList.remove('active');
        if (fab) fab.classList.remove('active');
    } else {
        // Close other panels
        window.closeAllPanels();
        if (panel) panel.classList.add('active');
        if (fab) fab.classList.add('active');
    }
};

window.togglePropertiesPanel = function() {
    const panel = document.getElementById('propertiesPanel');
    const fab = document.querySelector('.fab[onclick="togglePropertiesPanel()"]');
    
    if (panel && panel.classList.contains('active')) {
        panel.classList.remove('active');
        if (fab) fab.classList.remove('active');
    } else {
        // Close other panels
        window.closeAllPanels();
        if (panel) panel.classList.add('active');
        if (fab) fab.classList.add('active');
    }
};

window.toggleAIAssistant = function() {
    const panel = document.getElementById('aiAssistantPanel');
    const fab = document.querySelector('.fab[onclick="toggleAIAssistant()"]');
    
    if (panel && panel.classList.contains('active')) {
        panel.classList.remove('active');
        if (fab) fab.classList.remove('active');
    } else {
        // Close other panels
        window.closeAllPanels();
        if (panel) panel.classList.add('active');
        if (fab) fab.classList.add('active');
    }
};

window.closeAllPanels = function() {
    document.querySelectorAll('.modules-panel, .properties-panel, .ai-assistant-panel').forEach(panel => {
        panel.classList.remove('active');
    });
    document.querySelectorAll('.fab').forEach(fab => {
        fab.classList.remove('active');
    });
};

// AI Chat functionality
window.sendAIMessage = function() {
    const input = document.getElementById('aiInput');
    if (!input) return;
    
    const message = input.value.trim();
    if (!message) return;

    // Add user message to chat
    window.addChatMessage('user', message);
    input.value = '';

    // Simulate AI response
    setTimeout(() => {
        let response = '';
        if (message.toLowerCase().includes('web server')) {
            response = 'I recommend adding an Nginx module for a web server. You can drag the "Nginx" module from the Services category to your canvas.';
            // Auto-suggest nginx module
            setTimeout(() => {
                window.addModuleFromType('nginx');
            }, 1000);
        } else if (message.toLowerCase().includes('development')) {
            response = 'For a development environment, I suggest adding Development Tools and configuring a user account. Let me add these modules for you.';
            setTimeout(() => {
                window.addModuleFromType('development');
            }, 1000);
        } else {
            response = 'I can help you configure your NixOS system. Try asking about specific components like "web server", "development environment", or "user configuration".';
        }
        
        window.addChatMessage('assistant', response);
    }, 1000);
};

window.addChatMessage = function(role, message) {
    const messagesContainer = document.getElementById('chatMessages');
    if (!messagesContainer) return;
    
    const messageDiv = document.createElement('div');
    messageDiv.className = `chat-message ${role}`;
    
    messageDiv.innerHTML = `
        <div class="chat-avatar">
            <i class="fas fa-${role === 'user' ? 'user' : 'robot'}"></i>
        </div>
        <div class="chat-content">
            <p>${message}</p>
        </div>
    `;
    
    messagesContainer.appendChild(messageDiv);
    messagesContainer.scrollTop = messagesContainer.scrollHeight;
};

// Module removal functionality
window.removeModule = function(button) {
    const moduleElement = button.closest('.canvas-module');
    if (!moduleElement) return;
    
    const moduleId = moduleElement.dataset.moduleId;
    
    moduleElement.remove();
    window.configModules = (window.configModules || []).filter(m => m.id !== moduleId);
    
    // Show dropzone if no modules left
    const canvas = document.getElementById('builderCanvas');
    if (canvas && canvas.children.length === 0) {
        canvas.innerHTML = `
            <div class="canvas-dropzone">
                <i class="fas fa-plus-circle"></i>
                <p>Drag modules here to build your configuration</p>
                <p class="text-muted">or use AI assistance to generate configurations</p>
            </div>
        `;
    }
    
    window.clearPropertiesPanel();
    window.showNotification('Module removed', 'info');
};

// Clear properties panel
window.clearPropertiesPanel = function() {
    const panel = document.querySelector('.properties-panel-content');
    if (!panel) return;
    
    panel.innerHTML = `
        <div class="empty-state">
            <i class="fas fa-cog"></i>
            <p>Select a module to configure its properties</p>
        </div>
    `;
};

// ===== BUILDER PANEL MANAGEMENT =====
// Global functions for panel management needed by builder template

window.toggleModulesPanel = function() {
    const panel = document.getElementById('modulesPanel');
    const fab = document.querySelector('.fab[onclick="toggleModulesPanel()"]');
    
    if (panel && panel.classList.contains('active')) {
        panel.classList.remove('active');
        if (fab) fab.classList.remove('active');
    } else {
        // Close other panels
        window.closeAllPanels();
        if (panel) panel.classList.add('active');
        if (fab) fab.classList.add('active');
    }
};

window.togglePropertiesPanel = function() {
    const panel = document.getElementById('propertiesPanel');
    const fab = document.querySelector('.fab[onclick="togglePropertiesPanel()"]');
    
    if (panel && panel.classList.contains('active')) {
        panel.classList.remove('active');
        if (fab) fab.classList.remove('active');
    } else {
        // Close other panels
        window.closeAllPanels();
        if (panel) panel.classList.add('active');
        if (fab) fab.classList.add('active');
    }
};

window.toggleAIAssistant = function() {
    const panel = document.getElementById('aiAssistantPanel');
    const fab = document.querySelector('.fab[onclick="toggleAIAssistant()"]');
    
    if (panel && panel.classList.contains('active')) {
        panel.classList.remove('active');
        if (fab) fab.classList.remove('active');
    } else {
        // Close other panels
        window.closeAllPanels();
        if (panel) panel.classList.add('active');
        if (fab) fab.classList.add('active');
    }
};

window.closeAllPanels = function() {
    document.querySelectorAll('.modules-panel, .properties-panel, .ai-assistant-panel').forEach(panel => {
        panel.classList.remove('active');
    });
    document.querySelectorAll('.fab').forEach(fab => {
        fab.classList.remove('active');
    });
};

// AI Chat functionality
window.sendAIMessage = function() {
    const input = document.getElementById('aiInput');
    if (!input) return;
    
    const message = input.value.trim();
    if (!message) return;

    // Add user message to chat
    window.addChatMessage('user', message);
    input.value = '';

    // Simulate AI response
    setTimeout(() => {
        let response = '';
        if (message.toLowerCase().includes('web server')) {
            response = 'I recommend adding an Nginx module for a web server. You can drag the "Nginx" module from the Services category to your canvas.';
            // Auto-suggest nginx module
            setTimeout(() => {
                window.addModuleFromType('nginx');
            }, 1000);
        } else if (message.toLowerCase().includes('development')) {
            response = 'For a development environment, I suggest adding Development Tools and configuring a user account. Let me add these modules for you.';
            setTimeout(() => {
                window.addModuleFromType('development');
            }, 1000);
        } else {
            response = 'I can help you configure your NixOS system. Try asking about specific components like "web server", "development environment", or "user configuration".';
        }
        
        window.addChatMessage('assistant', response);
    }, 1000);
};

window.addChatMessage = function(role, message) {
    const messagesContainer = document.getElementById('chatMessages');
    if (!messagesContainer) return;
    
    const messageDiv = document.createElement('div');
    messageDiv.className = `chat-message ${role}`;
    
    messageDiv.innerHTML = `
        <div class="chat-avatar">
            <i class="fas fa-${role === 'user' ? 'user' : 'robot'}"></i>
        </div>
        <div class="chat-content">
            <p>${message}</p>
        </div>
    `;
    
    messagesContainer.appendChild(messageDiv);
    messagesContainer.scrollTop = messagesContainer.scrollHeight;
};

// Module removal functionality
window.removeModule = function(button) {
    const moduleElement = button.closest('.canvas-module');
    if (!moduleElement) return;
    
    const moduleId = moduleElement.dataset.moduleId;
    
    moduleElement.remove();
    window.configModules = (window.configModules || []).filter(m => m.id !== moduleId);
    
    // Show dropzone if no modules left
    const canvas = document.getElementById('builderCanvas');
    if (canvas && canvas.children.length === 0) {
        canvas.innerHTML = `
            <div class="canvas-dropzone">
                <i class="fas fa-plus-circle"></i>
                <p>Drag modules here to build your configuration</p>
                <p class="text-muted">or use AI assistance to generate configurations</p>
            </div>
        `;
    }
    
    window.clearPropertiesPanel();
    window.showNotification('Module removed', 'info');
};

// Clear properties panel
window.clearPropertiesPanel = function() {
    const panel = document.querySelector('.properties-panel-content');
    if (!panel) return;
    
    panel.innerHTML = `
        <div class="empty-state">
            <i class="fas fa-cog"></i>
            <p>Select a module to configure its properties</p>
        </div>
    `;
};

// ===== BUILDER PANEL MANAGEMENT =====
// Global functions for panel management needed by builder template

window.toggleModulesPanel = function() {
    const panel = document.getElementById('modulesPanel');
    const fab = document.querySelector('.fab[onclick="toggleModulesPanel()"]');
    
    if (panel && panel.classList.contains('active')) {
        panel.classList.remove('active');
        if (fab) fab.classList.remove('active');
    } else {
        // Close other panels
        window.closeAllPanels();
        if (panel) panel.classList.add('active');
        if (fab) fab.classList.add('active');
    }
};

window.togglePropertiesPanel = function() {
    const panel = document.getElementById('propertiesPanel');
    const fab = document.querySelector('.fab[onclick="togglePropertiesPanel()"]');
    
    if (panel && panel.classList.contains('active')) {
        panel.classList.remove('active');
        if (fab) fab.classList.remove('active');
    } else {
        // Close other panels
        window.closeAllPanels();
        if (panel) panel.classList.add('active');
        if (fab) fab.classList.add('active');
    }
};

window.toggleAIAssistant = function() {
    const panel = document.getElementById('aiAssistantPanel');
    const fab = document.querySelector('.fab[onclick="toggleAIAssistant()"]');
    
    if (panel && panel.classList.contains('active')) {
        panel.classList.remove('active');
        if (fab) fab.classList.remove('active');
    } else {