// NixAI Visual Configuration Builder - Complete Implementation

class ConfigBuilder {
    constructor() {
        this.components = [];
        this.templates = [];
        this.selectedComponent = null;
        this.canvas = document.getElementById('builder-canvas');
        this.preview = document.getElementById('nix-config-output');
        this.library = document.getElementById('component-library');
        this.nextId = 1;
        
        this.init();
    }
    
    async init() {
        await this.loadTemplates();
        this.setupEventListeners();
        this.renderComponentLibrary();
        this.updatePreview();
    }
    
    async loadTemplates() {
        try {
            const response = await fetch('/api/templates');
            this.templates = await response.json();
            
            // Add built-in NixOS components if no templates loaded
            if (this.templates.length === 0) {
                this.templates = this.getBuiltinComponents();
            }
        } catch (error) {
            console.error('Failed to load templates:', error);
            this.templates = this.getBuiltinComponents();
        }
    }
    
    getBuiltinComponents() {
        return [
            {
                name: 'SSH Service',
                description: 'Enable SSH daemon for remote access',
                file: 'services.openssh = {\n  enable = true;\n  settings.PasswordAuthentication = false;\n};'
            },
            {
                name: 'Nginx Web Server',
                description: 'HTTP/HTTPS web server',
                file: 'services.nginx = {\n  enable = true;\n  virtualHosts."example.com" = {\n    enableACME = true;\n    forceSSL = true;\n  };\n};'
            },
            {
                name: 'PostgreSQL Database',
                description: 'PostgreSQL database server',
                file: 'services.postgresql = {\n  enable = true;\n  package = pkgs.postgresql_15;\n  initialDatabases = [{ name = "mydb"; }];\n};'
            },
            {
                name: 'Docker Container Runtime',
                description: 'Docker containerization platform',
                file: 'virtualisation.docker = {\n  enable = true;\n  enableOnBoot = true;\n};'
            },
            {
                name: 'Firewall Configuration',
                description: 'Network firewall settings',
                file: 'networking.firewall = {\n  enable = true;\n  allowedTCPPorts = [ 22 80 443 ];\n};'
            }
        ];
    }
    
    renderComponentLibrary() {
        this.library.innerHTML = `
            <h2>🔧 Component Library</h2>
            <div class="components-list">
                ${this.templates.map(template => `
                    <div class="component-item" draggable="true" data-template="${encodeURIComponent(JSON.stringify(template))}">
                        <div class="component-header">${template.name}</div>
                        <div class="component-description">${template.description}</div>
                    </div>
                `).join('')}
            </div>
        `;
    }
    
    setupEventListeners() {
        // Drag and drop from library to canvas
        this.library.addEventListener('dragstart', (e) => {
            if (e.target.classList.contains('component-item')) {
                e.dataTransfer.setData('text/plain', e.target.dataset.template);
            }
        });
        
        // Canvas drop zone
        this.canvas.addEventListener('dragover', (e) => {
            e.preventDefault();
            this.canvas.classList.add('drag-over');
        });
        
        this.canvas.addEventListener('dragleave', (e) => {
            if (!this.canvas.contains(e.relatedTarget)) {
                this.canvas.classList.remove('drag-over');
            }
        });
        
        this.canvas.addEventListener('drop', (e) => {
            e.preventDefault();
            this.canvas.classList.remove('drag-over');
            
            const templateData = e.dataTransfer.getData('text/plain');
            if (templateData) {
                const template = JSON.parse(decodeURIComponent(templateData));
                const rect = this.canvas.getBoundingClientRect();
                const x = e.clientX - rect.left;
                const y = e.clientY - rect.top;
                
                this.addComponent(template, x, y);
            }
        });
        
        // Component selection and movement
        this.canvas.addEventListener('mousedown', (e) => {
            if (e.target.classList.contains('canvas-component')) {
                this.selectComponent(e.target);
                this.startDragging(e);
            } else {
                this.deselectAll();
            }
        });
        
        // Export button
        document.getElementById('export-btn').addEventListener('click', () => {
            this.exportConfiguration();
        });
        
        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Delete' && this.selectedComponent) {
                this.removeComponent(this.selectedComponent);
            }
        });
    }
    
    addComponent(template, x, y) {
        const component = {
            id: this.nextId++,
            template: template,
            x: Math.max(0, x - 75),
            y: Math.max(0, y - 25),
            width: 150,
            height: 80
        };
        
        this.components.push(component);
        this.renderComponent(component);
        this.updatePreview();
    }
    
    renderComponent(component) {
        const element = document.createElement('div');
        element.className = 'canvas-component';
        element.dataset.componentId = component.id;
        element.style.left = component.x + 'px';
        element.style.top = component.y + 'px';
        element.style.width = component.width + 'px';
        element.style.minHeight = component.height + 'px';
        
        element.innerHTML = `
            <div class="component-header">${component.template.name}</div>
            <div class="component-description">${component.template.description}</div>
        `;
        
        this.canvas.appendChild(element);
    }
    
    selectComponent(element) {
        this.deselectAll();
        element.classList.add('selected');
        this.selectedComponent = element;
    }
    
    deselectAll() {
        this.canvas.querySelectorAll('.canvas-component').forEach(el => {
            el.classList.remove('selected');
        });
        this.selectedComponent = null;
    }
    
    removeComponent(element) {
        const componentId = parseInt(element.dataset.componentId);
        this.components = this.components.filter(c => c.id !== componentId);
        element.remove();
        this.selectedComponent = null;
        this.updatePreview();
    }
    
    startDragging(e) {
        const component = e.target;
        const rect = component.getBoundingClientRect();
        const offsetX = e.clientX - rect.left;
        const offsetY = e.clientY - rect.top;
        
        const onMouseMove = (e) => {
            const canvasRect = this.canvas.getBoundingClientRect();
            const x = e.clientX - canvasRect.left - offsetX;
            const y = e.clientY - canvasRect.top - offsetY;
            
            component.style.left = Math.max(0, x) + 'px';
            component.style.top = Math.max(0, y) + 'px';
            
            // Update component data
            const componentId = parseInt(component.dataset.componentId);
            const componentData = this.components.find(c => c.id === componentId);
            if (componentData) {
                componentData.x = Math.max(0, x);
                componentData.y = Math.max(0, y);
            }
        };
        
        const onMouseUp = () => {
            document.removeEventListener('mousemove', onMouseMove);
            document.removeEventListener('mouseup', onMouseUp);
        };
        
        document.addEventListener('mousemove', onMouseMove);
        document.addEventListener('mouseup', onMouseUp);
    }
    
    updatePreview() {
        if (this.components.length === 0) {
            this.preview.textContent = `# NixOS Configuration
# Drag components from the library to build your configuration

{ config, pkgs, ... }:

{
  # Add components here by dragging from the library
  
  # System configuration
  system.stateVersion = "24.05";
}`;
            return;
        }
        
        const configParts = this.components.map(component => {
            return `  # ${component.template.name}\n  ${component.template.file.split('\n').join('\n  ')}`;
        });
        
        const fullConfig = `# NixOS Configuration
# Generated by NixAI Visual Configuration Builder

{ config, pkgs, ... }:

{
${configParts.join('\n\n')}

  # System configuration
  system.stateVersion = "24.05";
}`;
        
        this.preview.textContent = fullConfig;
        this.highlightSyntax();
    }
    
    highlightSyntax() {
        // Basic Nix syntax highlighting
        let content = this.preview.innerHTML;
        
        // Keywords
        content = content.replace(/\b(let|in|with|import|inherit|if|then|else|assert|rec)\b/g, '<span class="nix-keyword">$1</span>');
        
        // Strings
        content = content.replace(/"([^"\\]|\\.)*"/g, '<span class="nix-string">$&</span>');
        
        // Comments
        content = content.replace(/#[^\n]*/g, '<span class="nix-comment">$&</span>');
        
        // Attributes
        content = content.replace(/(\w+)\s*=/g, '<span class="nix-attribute">$1</span>=');
        
        this.preview.innerHTML = content;
    }
    
    exportConfiguration() {
        const config = this.preview.textContent;
        const blob = new Blob([config], { type: 'text/plain' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = 'configuration.nix';
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
        
        // Show success message
        this.showNotification('Configuration exported successfully!', 'success');
    }
    
    showNotification(message, type = 'info') {
        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        notification.textContent = message;
        notification.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            background: ${type === 'success' ? '#10b981' : '#3b82f6'};
            color: white;
            padding: 12px 20px;
            border-radius: 8px;
            box-shadow: 0 4px 12px rgba(0,0,0,0.2);
            z-index: 1000;
            animation: slideIn 0.3s ease-out;
        `;
        
        document.body.appendChild(notification);
        
        setTimeout(() => {
            notification.style.animation = 'slideOut 0.3s ease-in';
            setTimeout(() => {
                document.body.removeChild(notification);
            }, 300);
        }, 3000);
    }
}

// Add CSS animations
const style = document.createElement('style');
style.textContent = `
    @keyframes slideIn {
        from { transform: translateX(100%); opacity: 0; }
        to { transform: translateX(0); opacity: 1; }
    }
    
    @keyframes slideOut {
        from { transform: translateX(0); opacity: 1; }
        to { transform: translateX(100%); opacity: 0; }
    }
`;
document.head.appendChild(style);

// Initialize the application
document.addEventListener('DOMContentLoaded', () => {
    new ConfigBuilder();
});
