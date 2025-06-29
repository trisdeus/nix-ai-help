// nixai Visual Configuration Builder
class ConfigurationBuilder {
    constructor() {
        this.canvas = null;
        this.components = new Map();
        this.connections = [];
        this.selectedComponents = new Set();
        this.draggedComponent = null;
        this.isDragging = false;
        this.isConnecting = false;
        this.connectionStart = null;
        this.zoom = 1.0;
        this.panX = 0;
        this.panY = 0;
        
        this.init();
    }
    
    init() {
        this.setupEventListeners();
        this.loadComponentLibrary();
        this.initializeCanvas();
        this.setupWebSocket();
    }
    
    setupEventListeners() {
        // Search functionality
        document.getElementById('component-search').addEventListener('input', (e) => {
            this.filterComponents(e.target.value);
        });
        
        // Category filter
        document.getElementById('category-filter').addEventListener('change', (e) => {
            this.filterByCategory(e.target.value);
        });
        
        // Difficulty filter
        document.getElementById('difficulty-filter').addEventListener('change', (e) => {
            this.filterByDifficulty(e.target.value);
        });
        
        // Toolbar buttons
        document.getElementById('save-btn').addEventListener('click', () => this.saveConfiguration());
        document.getElementById('load-btn').addEventListener('click', () => this.loadConfiguration());
        document.getElementById('auto-layout-btn').addEventListener('click', () => this.showAutoLayoutModal());
        document.getElementById('validate-btn').addEventListener('click', () => this.validateConfiguration());
        document.getElementById('dependency-graph-btn').addEventListener('click', () => this.showDependencyGraph());
        
        // Zoom controls
        document.getElementById('zoom-in-btn').addEventListener('click', () => this.zoomIn());
        document.getElementById('zoom-out-btn').addEventListener('click', () => this.zoomOut());
        document.getElementById('grid-toggle-btn').addEventListener('click', () => this.toggleGrid());
        document.getElementById('snap-toggle-btn').addEventListener('click', () => this.toggleSnap());
        
        // Preview controls
        document.getElementById('refresh-preview-btn').addEventListener('click', () => this.refreshPreview());
        document.getElementById('copy-config-btn').addEventListener('click', () => this.copyConfiguration());
        document.getElementById('download-config-btn').addEventListener('click', () => this.downloadConfiguration());
        document.getElementById('preview-mode').addEventListener('change', (e) => {
            this.updatePreviewMode(e.target.value);
        });
        
        // Canvas events
        const canvasContainer = document.getElementById('canvas-container');
        canvasContainer.addEventListener('dragover', (e) => this.handleDragOver(e));
        canvasContainer.addEventListener('drop', (e) => this.handleDrop(e));
        canvasContainer.addEventListener('click', (e) => this.handleCanvasClick(e));
        canvasContainer.addEventListener('contextmenu', (e) => this.handleContextMenu(e));
        
        // Modal events
        this.setupModalEvents();
        
        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => this.handleKeyboard(e));
        
        // Auto-save
        setInterval(() => this.autoSave(), 30000); // Auto-save every 30 seconds
    }
    
    setupModalEvents() {
        // Component config modal
        document.getElementById('close-config-modal').addEventListener('click', () => {
            this.hideModal('component-config-modal');
        });
        document.getElementById('cancel-config-btn').addEventListener('click', () => {
            this.hideModal('component-config-modal');
        });
        document.getElementById('apply-config-btn').addEventListener('click', () => {
            this.applyComponentConfiguration();
        });
        
        // Dependency graph modal
        document.getElementById('close-graph-modal').addEventListener('click', () => {
            this.hideModal('dependency-graph-modal');
        });
        
        // Auto layout modal
        document.getElementById('cancel-layout-btn').addEventListener('click', () => {
            this.hideModal('auto-layout-modal');
        });
        document.getElementById('apply-layout-btn').addEventListener('click', () => {
            this.applyAutoLayout();
        });
    }
    
    async loadComponentLibrary() {
        try {
            this.showLoading(true);
            const response = await axios.get('/api/components');
            const components = response.data;
            
            this.renderComponentList(components);
            this.showLoading(false);
        } catch (error) {
            console.error('Failed to load component library:', error);
            this.showError('Failed to load component library');
            this.showLoading(false);
        }
    }
    
    renderComponentList(components) {
        const componentList = document.getElementById('component-list');
        componentList.innerHTML = '';
        
        // Group components by category
        const grouped = this.groupComponentsByCategory(components);
        
        Object.entries(grouped).forEach(([category, categoryComponents]) => {
            const categoryHeader = document.createElement('div');
            categoryHeader.className = 'mb-2 pb-2 border-b border-gray-200';
            categoryHeader.innerHTML = `
                <h3 class="text-sm font-semibold text-gray-700 uppercase tracking-wide">
                    ${category.replace('_', ' ')}
                </h3>
            `;
            componentList.appendChild(categoryHeader);
            
            categoryComponents.forEach(component => {
                const componentItem = this.createComponentItem(component);
                componentList.appendChild(componentItem);
            });
        });
    }
    
    createComponentItem(component) {
        const item = document.createElement('div');
        item.className = 'component-item bg-white p-3 rounded-lg border border-gray-200 cursor-grab';
        item.draggable = true;
        item.dataset.componentId = component.id;
        
        const difficultyColor = this.getDifficultyColor(component.difficulty);
        
        item.innerHTML = `
            <div class="flex items-start space-x-3">
                <div class="flex-shrink-0">
                    <div class="w-8 h-8 rounded-lg flex items-center justify-center text-lg" 
                         style="background-color: ${component.color}20; color: ${component.color}">
                        ${component.icon}
                    </div>
                </div>
                <div class="flex-1 min-w-0">
                    <div class="flex items-center space-x-2">
                        <h4 class="text-sm font-medium text-gray-900 truncate">${component.name}</h4>
                        <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${difficultyColor}">
                            ${component.difficulty}
                        </span>
                    </div>
                    <p class="text-xs text-gray-500 mt-1 line-clamp-2">${component.description}</p>
                    <div class="flex items-center mt-2 space-x-2">
                        <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-800">
                            ${component.type}
                        </span>
                        ${component.tags.slice(0, 2).map(tag => 
                            `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800">${tag}</span>`
                        ).join('')}
                    </div>
                </div>
            </div>
        `;
        
        // Drag events
        item.addEventListener('dragstart', (e) => this.handleDragStart(e, component));
        item.addEventListener('dragend', (e) => this.handleDragEnd(e));
        
        // Click for details
        item.addEventListener('click', (e) => {
            e.stopPropagation();
            this.showComponentDetails(component);
        });
        
        return item;
    }
    
    groupComponentsByCategory(components) {
        return components.reduce((groups, component) => {
            const category = component.category || 'Other';
            if (!groups[category]) {
                groups[category] = [];
            }
            groups[category].push(component);
            return groups;
        }, {});
    }
    
    getDifficultyColor(difficulty) {
        const colors = {
            'beginner': 'bg-green-100 text-green-800',
            'intermediate': 'bg-yellow-100 text-yellow-800',
            'advanced': 'bg-orange-100 text-orange-800',
            'expert': 'bg-red-100 text-red-800'
        };
        return colors[difficulty] || 'bg-gray-100 text-gray-800';
    }
    
    handleDragStart(e, component) {
        this.draggedComponent = component;
        e.dataTransfer.setData('text/plain', component.id);
        e.dataTransfer.effectAllowed = 'copy';
        
        // Create drag image
        const dragImage = e.target.cloneNode(true);
        dragImage.style.opacity = '0.7';
        dragImage.style.transform = 'rotate(5deg)';
        document.body.appendChild(dragImage);
        e.dataTransfer.setDragImage(dragImage, 50, 25);
        
        setTimeout(() => document.body.removeChild(dragImage), 0);
    }
    
    handleDragEnd(e) {
        this.draggedComponent = null;
    }
    
    handleDragOver(e) {
        e.preventDefault();
        e.dataTransfer.dropEffect = 'copy';
    }
    
    handleDrop(e) {
        e.preventDefault();
        
        if (!this.draggedComponent) return;
        
        const rect = e.currentTarget.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;
        
        this.addComponentToCanvas(this.draggedComponent, { x, y });
        this.draggedComponent = null;
    }
    
    async addComponentToCanvas(component, position) {
        try {
            const response = await axios.post('/api/canvas/components', {
                component_id: component.id,
                position: position
            });
            
            const placedComponent = response.data;
            this.renderCanvasComponent(placedComponent);
            this.updateStatusBar();
            this.refreshPreview();
            
        } catch (error) {
            console.error('Failed to add component:', error);
            this.showError('Failed to add component to canvas');
        }
    }
    
    renderCanvasComponent(placedComponent) {
        const canvasContent = document.getElementById('canvas-content');
        
        const componentElement = document.createElement('div');
        componentElement.className = 'canvas-component bg-white rounded-lg shadow-md p-3 border-2 border-transparent';
        componentElement.style.left = `${placedComponent.position.x}px`;
        componentElement.style.top = `${placedComponent.position.y}px`;
        componentElement.style.width = `${placedComponent.size.width}px`;
        componentElement.style.height = `${placedComponent.size.height}px`;
        componentElement.dataset.instanceId = placedComponent.instance_id;
        
        componentElement.innerHTML = `
            <div class="flex items-center space-x-2 mb-2">
                <div class="w-6 h-6 rounded flex items-center justify-center text-sm" 
                     style="background-color: ${placedComponent.component.color}20; color: ${placedComponent.component.color}">
                    ${placedComponent.component.icon}
                </div>
                <h4 class="text-sm font-medium text-gray-900 truncate flex-1">${placedComponent.component.name}</h4>
                <div class="flex space-x-1">
                    <button class="config-btn text-gray-400 hover:text-blue-600 text-xs" title="Configure">
                        <i class="fas fa-cog"></i>
                    </button>
                    <button class="remove-btn text-gray-400 hover:text-red-600 text-xs" title="Remove">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
            </div>
            <div class="text-xs text-gray-500">
                <div class="truncate">${placedComponent.component.type}</div>
                <div class="truncate">${placedComponent.component.description}</div>
            </div>
            
            <!-- Connection points -->
            <div class="connection-point connection-point-top" style="top: -4px; left: 50%; transform: translateX(-50%);"></div>
            <div class="connection-point connection-point-right" style="top: 50%; right: -4px; transform: translateY(-50%);"></div>
            <div class="connection-point connection-point-bottom" style="bottom: -4px; left: 50%; transform: translateX(-50%);"></div>
            <div class="connection-point connection-point-left" style="top: 50%; left: -4px; transform: translateY(-50%);"></div>
        `;
        
        // Make draggable
        this.makeComponentDraggable(componentElement);
        
        // Add event listeners
        componentElement.querySelector('.config-btn').addEventListener('click', (e) => {
            e.stopPropagation();
            this.configureComponent(placedComponent);
        });
        
        componentElement.querySelector('.remove-btn').addEventListener('click', (e) => {
            e.stopPropagation();
            this.removeComponent(placedComponent.instance_id);
        });
        
        componentElement.addEventListener('click', (e) => {
            e.stopPropagation();
            this.selectComponent(placedComponent.instance_id, !e.ctrlKey);
        });
        
        // Add connection points
        this.setupConnectionPoints(componentElement);
        
        canvasContent.appendChild(componentElement);
        this.components.set(placedComponent.instance_id, placedComponent);
    }
    
    makeComponentDraggable(element) {
        let isDragging = false;
        let startX, startY, startLeft, startTop;
        
        element.addEventListener('mousedown', (e) => {
            if (e.target.closest('.config-btn, .remove-btn, .connection-point')) return;
            
            isDragging = true;
            startX = e.clientX;
            startY = e.clientY;
            startLeft = parseInt(element.style.left) || 0;
            startTop = parseInt(element.style.top) || 0;
            
            element.style.cursor = 'grabbing';
            element.style.zIndex = '1000';
            
            e.preventDefault();
        });
        
        document.addEventListener('mousemove', (e) => {
            if (!isDragging) return;
            
            const deltaX = e.clientX - startX;
            const deltaY = e.clientY - startY;
            
            element.style.left = `${startLeft + deltaX}px`;
            element.style.top = `${startTop + deltaY}px`;
            
            this.updateConnections();
        });
        
        document.addEventListener('mouseup', () => {
            if (isDragging) {
                isDragging = false;
                element.style.cursor = 'move';
                element.style.zIndex = '';
                
                // Save new position
                const instanceId = element.dataset.instanceId;
                const newPosition = {
                    x: parseInt(element.style.left),
                    y: parseInt(element.style.top)
                };
                
                this.saveComponentPosition(instanceId, newPosition);
            }
        });
    }
    
    setupConnectionPoints(componentElement) {
        const connectionPoints = componentElement.querySelectorAll('.connection-point');
        
        connectionPoints.forEach(point => {
            point.className += ' absolute w-2 h-2 bg-blue-500 rounded-full opacity-0 hover:opacity-100 cursor-pointer transition-opacity';
            
            point.addEventListener('mousedown', (e) => {
                e.stopPropagation();
                this.startConnection(componentElement.dataset.instanceId, point);
            });
        });
    }
    
    startConnection(fromId, fromPoint) {
        this.isConnecting = true;
        this.connectionStart = { componentId: fromId, point: fromPoint };
        
        document.body.style.cursor = 'crosshair';
        
        // Show all connection points
        document.querySelectorAll('.connection-point').forEach(point => {
            point.style.opacity = '1';
        });
        
        // Add temporary connection line
        this.addTemporaryConnectionLine();
    }
    
    addTemporaryConnectionLine() {
        const svg = document.getElementById('canvas-svg');
        const tempLine = document.createElementNS('http://www.w3.org/2000/svg', 'line');
        tempLine.id = 'temp-connection';
        tempLine.className = 'connection-line';
        tempLine.style.stroke = '#3b82f6';
        tempLine.style.strokeWidth = '2';
        tempLine.style.strokeDasharray = '5,5';
        
        svg.appendChild(tempLine);
        
        document.addEventListener('mousemove', this.updateTemporaryLine.bind(this));
    }
    
    updateTemporaryLine(e) {
        const tempLine = document.getElementById('temp-connection');
        if (!tempLine || !this.connectionStart) return;
        
        const canvasRect = document.getElementById('canvas-container').getBoundingClientRect();
        const startPoint = this.getConnectionPointPosition(this.connectionStart.componentId, this.connectionStart.point);
        
        tempLine.setAttribute('x1', startPoint.x);
        tempLine.setAttribute('y1', startPoint.y);
        tempLine.setAttribute('x2', e.clientX - canvasRect.left);
        tempLine.setAttribute('y2', e.clientY - canvasRect.top);
    }
    
    getConnectionPointPosition(componentId, point) {
        const element = document.querySelector(`[data-instance-id="${componentId}"]`);
        if (!element) return { x: 0, y: 0 };
        
        const rect = element.getBoundingClientRect();
        const canvasRect = document.getElementById('canvas-container').getBoundingClientRect();
        
        const centerX = rect.left - canvasRect.left + rect.width / 2;
        const centerY = rect.top - canvasRect.top + rect.height / 2;
        
        if (point.classList.contains('connection-point-top')) {
            return { x: centerX, y: rect.top - canvasRect.top };
        } else if (point.classList.contains('connection-point-right')) {
            return { x: rect.right - canvasRect.left, y: centerY };
        } else if (point.classList.contains('connection-point-bottom')) {
            return { x: centerX, y: rect.bottom - canvasRect.top };
        } else if (point.classList.contains('connection-point-left')) {
            return { x: rect.left - canvasRect.left, y: centerY };
        }
        
        return { x: centerX, y: centerY };
    }
    
    finishConnection(toId, toPoint) {
        if (!this.isConnecting || !this.connectionStart) return;
        
        const fromId = this.connectionStart.componentId;
        
        if (fromId === toId) {
            this.cancelConnection();
            return;
        }
        
        this.createConnection(fromId, toId, 'dependency');
        this.cancelConnection();
    }
    
    cancelConnection() {
        this.isConnecting = false;
        this.connectionStart = null;
        
        document.body.style.cursor = '';
        document.removeEventListener('mousemove', this.updateTemporaryLine);
        
        // Hide connection points
        document.querySelectorAll('.connection-point').forEach(point => {
            point.style.opacity = '';
        });
        
        // Remove temporary line
        const tempLine = document.getElementById('temp-connection');
        if (tempLine) {
            tempLine.remove();
        }
    }
    
    async createConnection(fromId, toId, type) {
        try {
            const response = await axios.post('/api/canvas/connections', {
                from_id: fromId,
                to_id: toId,
                type: type
            });
            
            const connection = response.data;
            this.connections.push(connection);
            this.renderConnection(connection);
            this.updateStatusBar();
            
        } catch (error) {
            console.error('Failed to create connection:', error);
            this.showError('Failed to create connection');
        }
    }
    
    renderConnection(connection) {
        const svg = document.getElementById('connections');
        
        const line = document.createElementNS('http://www.w3.org/2000/svg', 'line');
        line.id = `connection-${connection.id}`;
        line.className = `connection-line ${connection.type}-line`;
        line.dataset.connectionId = connection.id;
        
        const fromPos = this.getComponentPosition(connection.from_id);
        const toPos = this.getComponentPosition(connection.to_id);
        
        line.setAttribute('x1', fromPos.x);
        line.setAttribute('y1', fromPos.y);
        line.setAttribute('x2', toPos.x);
        line.setAttribute('y2', toPos.y);
        line.setAttribute('marker-end', 'url(#arrowhead)');
        
        line.addEventListener('click', (e) => {
            e.stopPropagation();
            this.selectConnection(connection.id);
        });
        
        line.addEventListener('contextmenu', (e) => {
            e.preventDefault();
            this.showConnectionContextMenu(connection.id, e.clientX, e.clientY);
        });
        
        svg.appendChild(line);
    }
    
    getComponentPosition(instanceId) {
        const element = document.querySelector(`[data-instance-id="${instanceId}"]`);
        if (!element) return { x: 0, y: 0 };
        
        const rect = element.getBoundingClientRect();
        const canvasRect = document.getElementById('canvas-container').getBoundingClientRect();
        
        return {
            x: rect.left - canvasRect.left + rect.width / 2,
            y: rect.top - canvasRect.top + rect.height / 2
        };
    }
    
    updateConnections() {
        this.connections.forEach(connection => {
            const line = document.getElementById(`connection-${connection.id}`);
            if (!line) return;
            
            const fromPos = this.getComponentPosition(connection.from_id);
            const toPos = this.getComponentPosition(connection.to_id);
            
            line.setAttribute('x1', fromPos.x);
            line.setAttribute('y1', fromPos.y);
            line.setAttribute('x2', toPos.x);
            line.setAttribute('y2', toPos.y);
        });
    }
    
    async saveComponentPosition(instanceId, position) {
        try {
            await axios.put(`/api/canvas/components/${instanceId}/position`, {
                position: position
            });
        } catch (error) {
            console.error('Failed to save component position:', error);
        }
    }
    
    selectComponent(instanceId, clearSelection = true) {
        if (clearSelection) {
            this.selectedComponents.clear();
            document.querySelectorAll('.canvas-component.selected').forEach(el => {
                el.classList.remove('selected');
            });
        }
        
        this.selectedComponents.add(instanceId);
        const element = document.querySelector(`[data-instance-id="${instanceId}"]`);
        if (element) {
            element.classList.add('selected');
        }
    }
    
    async removeComponent(instanceId) {
        if (!confirm('Are you sure you want to remove this component?')) return;
        
        try {
            await axios.delete(`/api/canvas/components/${instanceId}`);
            
            // Remove from DOM
            const element = document.querySelector(`[data-instance-id="${instanceId}"]`);
            if (element) {
                element.remove();
            }
            
            // Remove connections
            this.connections = this.connections.filter(conn => {
                if (conn.from_id === instanceId || conn.to_id === instanceId) {
                    const line = document.getElementById(`connection-${conn.id}`);
                    if (line) line.remove();
                    return false;
                }
                return true;
            });
            
            this.components.delete(instanceId);
            this.selectedComponents.delete(instanceId);
            this.updateStatusBar();
            this.refreshPreview();
            
        } catch (error) {
            console.error('Failed to remove component:', error);
            this.showError('Failed to remove component');
        }
    }
    
    configureComponent(placedComponent) {
        const modal = document.getElementById('component-config-modal');
        const title = document.getElementById('config-modal-title');
        const content = document.getElementById('config-modal-content');
        
        title.textContent = `Configure ${placedComponent.component.name}`;
        content.innerHTML = this.generateConfigurationForm(placedComponent);
        
        this.showModal('component-config-modal');
        this.currentConfigComponent = placedComponent;
    }
    
    generateConfigurationForm(placedComponent) {
        let html = '';
        
        placedComponent.component.options.forEach(option => {
            const currentValue = placedComponent.config[option.name] || option.default_value;
            
            html += `
                <div class="mb-4">
                    <label class="block text-sm font-medium text-gray-700 mb-1">
                        ${option.name}
                        ${option.required ? '<span class="text-red-500">*</span>' : ''}
                    </label>
                    <p class="text-xs text-gray-500 mb-2">${option.description}</p>
            `;
            
            if (option.type === 'bool') {
                html += `
                    <input type="checkbox" id="config-${option.name}" name="${option.name}" 
                           class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                           ${currentValue ? 'checked' : ''}>
                `;
            } else if (option.type === 'string' && option.options && option.options.length > 0) {
                html += `<select id="config-${option.name}" name="${option.name}" class="w-full border border-gray-300 rounded-lg px-3 py-2">`;
                option.options.forEach(opt => {
                    html += `<option value="${opt}" ${currentValue === opt ? 'selected' : ''}>${opt}</option>`;
                });
                html += `</select>`;
            } else {
                html += `
                    <input type="${option.type === 'int' ? 'number' : 'text'}" 
                           id="config-${option.name}" name="${option.name}"
                           value="${currentValue || ''}"
                           class="w-full border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                           ${option.required ? 'required' : ''}>
                `;
            }
            
            html += '</div>';
        });
        
        return html;
    }
    
    async applyComponentConfiguration() {
        const form = document.getElementById('config-modal-content');
        const config = {};
        
        // Collect form data
        const inputs = form.querySelectorAll('input, select');
        inputs.forEach(input => {
            if (input.type === 'checkbox') {
                config[input.name] = input.checked;
            } else if (input.type === 'number') {
                config[input.name] = parseInt(input.value) || 0;
            } else {
                config[input.name] = input.value;
            }
        });
        
        try {
            await axios.put(`/api/canvas/components/${this.currentConfigComponent.instance_id}/config`, {
                config: config
            });
            
            this.hideModal('component-config-modal');
            this.refreshPreview();
            this.showSuccess('Component configuration updated');
            
        } catch (error) {
            console.error('Failed to update component configuration:', error);
            this.showError('Failed to update component configuration');
        }
    }
    
    async refreshPreview() {
        try {
            const mode = document.getElementById('preview-mode').value;
            const response = await axios.post('/api/preview/generate', {
                mode: mode
            });
            
            const result = response.data;
            this.displayPreview(result);
            
        } catch (error) {
            console.error('Failed to generate preview:', error);
            this.showError('Failed to generate preview');
        }
    }
    
    displayPreview(result) {
        const content = document.getElementById('preview-content');
        const statusBar = document.getElementById('validation-status');
        const issuesContainer = document.getElementById('issues-summary');
        
        if (result.success) {
            content.innerHTML = `<pre><code class="language-nix">${this.escapeHtml(result.configuration)}</code></pre>`;
            statusBar.innerHTML = '<i class="fas fa-circle text-green-500 mr-1"></i><span class="text-green-600">Valid</span>';
        } else {
            content.innerHTML = `<pre><code class="language-nix">${this.escapeHtml(result.configuration)}</code></pre>`;
            statusBar.innerHTML = '<i class="fas fa-circle text-red-500 mr-1"></i><span class="text-red-600">Errors</span>';
        }
        
        // Display issues
        issuesContainer.innerHTML = '';
        
        result.errors.forEach(error => {
            const errorDiv = document.createElement('div');
            errorDiv.className = 'text-xs text-red-600 bg-red-50 p-2 rounded';
            errorDiv.textContent = error.message;
            issuesContainer.appendChild(errorDiv);
        });
        
        result.warnings.forEach(warning => {
            const warningDiv = document.createElement('div');
            warningDiv.className = 'text-xs text-yellow-600 bg-yellow-50 p-2 rounded';
            warningDiv.textContent = warning.message;
            issuesContainer.appendChild(warningDiv);
        });
        
        // Update metadata
        if (result.metadata) {
            document.getElementById('component-count').textContent = `${result.metadata.component_count} components`;
        }
    }
    
    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
    
    updateStatusBar() {
        const componentCount = this.components.size;
        const connectionCount = this.connections.length;
        
        document.getElementById('component-count').textContent = `${componentCount} components`;
        document.getElementById('connection-count').textContent = `${connectionCount} connections`;
    }
    
    showModal(modalId) {
        document.getElementById(modalId).classList.remove('hidden');
    }
    
    hideModal(modalId) {
        document.getElementById(modalId).classList.add('hidden');
    }
    
    showLoading(show) {
        const overlay = document.getElementById('loading-overlay');
        if (show) {
            overlay.classList.remove('hidden');
        } else {
            overlay.classList.add('hidden');
        }
    }
    
    showError(message) {
        this.showNotification(message, 'error');
    }
    
    showSuccess(message) {
        this.showNotification(message, 'success');
    }
    
    showNotification(message, type) {
        // Simple notification system
        const notification = document.createElement('div');
        notification.className = `fixed top-4 right-4 z-50 p-4 rounded-lg shadow-lg ${
            type === 'error' ? 'bg-red-500 text-white' : 'bg-green-500 text-white'
        }`;
        notification.textContent = message;
        
        document.body.appendChild(notification);
        
        setTimeout(() => {
            notification.remove();
        }, 5000);
    }
    
    handleKeyboard(e) {
        if (e.ctrlKey || e.metaKey) {
            switch (e.key) {
                case 's':
                    e.preventDefault();
                    this.saveConfiguration();
                    break;
                case 'z':
                    e.preventDefault();
                    if (e.shiftKey) {
                        this.redo();
                    } else {
                        this.undo();
                    }
                    break;
                case 'a':
                    e.preventDefault();
                    this.selectAll();
                    break;
                case 'c':
                    e.preventDefault();
                    this.copySelected();
                    break;
                case 'v':
                    e.preventDefault();
                    this.pasteComponents();
                    break;
            }
        } else if (e.key === 'Delete' || e.key === 'Backspace') {
            this.deleteSelected();
        } else if (e.key === 'Escape') {
            this.clearSelection();
            this.cancelConnection();
        }
    }
    
    async saveConfiguration() {
        try {
            this.showLoading(true);
            const response = await axios.post('/api/canvas/save');
            this.showSuccess('Configuration saved successfully');
            this.showLoading(false);
        } catch (error) {
            console.error('Failed to save configuration:', error);
            this.showError('Failed to save configuration');
            this.showLoading(false);
        }
    }
    
    filterComponents(query) {
        const items = document.querySelectorAll('.component-item');
        items.forEach(item => {
            const text = item.textContent.toLowerCase();
            if (text.includes(query.toLowerCase())) {
                item.style.display = '';
            } else {
                item.style.display = 'none';
            }
        });
    }
    
    initializeCanvas() {
        // Initialize canvas with default settings
        this.refreshPreview();
        this.updateStatusBar();
    }
    
    setupWebSocket() {
        // Setup WebSocket for real-time collaboration (if needed)
        // This would connect to the collaboration features
    }
}

// Initialize the configuration builder when the page loads
document.addEventListener('DOMContentLoaded', () => {
    new ConfigurationBuilder();
});
