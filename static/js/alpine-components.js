// Alpine.js components for CodeValdCortex Dashboard

// Dashboard component
function dashboard() {
    return {
        search: '',
        filter: 'all',
        healthFilter: 'all',

        init() {
            this.setupEventListeners();
        },

        setupEventListeners() {
            // Listen for HTMX events
            document.body.addEventListener('htmx:afterSwap', (e) => {
                // Content updated
            });

            document.body.addEventListener('htmx:responseError', (e) => {
                console.error('HTMX error', e.detail);
                this.showError('Failed to update. Please refresh the page.');
            });
        },

        showError(message) {
            // Simple error notification
            const notification = document.createElement('div');
            notification.className = 'fixed bottom-4 right-4 bg-red-500 text-white px-6 py-3 rounded-lg shadow-lg z-50';
            notification.textContent = message;
            document.body.appendChild(notification);

            setTimeout(() => {
                notification.remove();
            }, 5000);
        }
    }
}

// Metrics chart component
function metricsChart() {
    return {
        chart: null,
        agentId: null,

        init(agentId) {
            this.agentId = agentId;
            const canvas = this.$el.querySelector('canvas');

            if (!canvas) {
                console.error('Canvas not found');
                return;
            }

            this.chart = new Chart(canvas, {
                type: 'line',
                data: {
                    labels: [],
                    datasets: [{
                        label: 'CPU Usage (%)',
                        data: [],
                        borderColor: 'rgb(75, 192, 192)',
                        backgroundColor: 'rgba(75, 192, 192, 0.1)',
                        tension: 0.1,
                        fill: true
                    }, {
                        label: 'Memory (MB)',
                        data: [],
                        borderColor: 'rgb(255, 99, 132)',
                        backgroundColor: 'rgba(255, 99, 132, 0.1)',
                        tension: 0.1,
                        fill: true,
                        yAxisID: 'y1'
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    interaction: {
                        mode: 'index',
                        intersect: false,
                    },
                    scales: {
                        y: {
                            type: 'linear',
                            display: true,
                            position: 'left',
                            beginAtZero: true,
                            max: 100,
                            title: {
                                display: true,
                                text: 'CPU (%)'
                            }
                        },
                        y1: {
                            type: 'linear',
                            display: true,
                            position: 'right',
                            beginAtZero: true,
                            title: {
                                display: true,
                                text: 'Memory (MB)'
                            },
                            grid: {
                                drawOnChartArea: false,
                            },
                        },
                    }
                }
            });

            // Store globally for HTMX to access
            window.updateChart = (data) => this.updateData(data);
        },

        updateData(data) {
            const metrics = typeof data === 'string' ? JSON.parse(data) : data;
            const now = new Date().toLocaleTimeString();

            this.chart.data.labels.push(now);
            this.chart.data.datasets[0].data.push(metrics.cpu || 0);
            this.chart.data.datasets[1].data.push(metrics.memory || 0);

            // Keep only last 20 data points
            if (this.chart.data.labels.length > 20) {
                this.chart.data.labels.shift();
                this.chart.data.datasets[0].data.shift();
                this.chart.data.datasets[1].data.shift();
            }

            this.chart.update('none'); // Update without animation for better performance
        }
    }
}

// Log viewer component
function logViewer() {
    return {
        level: 'all',
        search: '',
        autoScroll: true,

        filterLogs() {
            // Filter logic handled by HTMX requests
            const params = new URLSearchParams();
            if (this.level !== 'all') params.append('level', this.level);
            if (this.search) params.append('search', this.search);
        },

        scrollToBottom() {
            if (this.autoScroll) {
                const container = this.$el.querySelector('#logs-container');
                if (container) {
                    container.scrollTop = container.scrollHeight;
                }
            }
        }
    }
}

// Make functions globally available
window.dashboard = dashboard;
window.metricsChart = metricsChart;
window.logViewer = logViewer;
