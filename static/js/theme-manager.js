// Theme Management for CodeValdCortex
class ThemeManager {
    constructor() {
        this.themes = {
            'light': {
                name: 'Light',
                preview: '#ffffff',
                description: 'Clean and bright'
            },
            'midnight-coral': {
                name: 'Midnight Coral',
                preview: '#FF6B6B',
                description: 'Professional with vibrant accents'
            },
            'slate-purple': {
                name: 'Slate Purple',
                preview: '#8b5cf6',
                description: 'Bold and tech-forward'
            },
            'charcoal-emerald': {
                name: 'Charcoal Emerald',
                preview: '#10b981',
                description: 'Fresh and trustworthy'
            },
            'navy-orange': {
                name: 'Navy Orange',
                preview: '#f97316',
                description: 'Energetic and confident'
            },
            'obsidian-cyan': {
                name: 'Obsidian Cyan',
                preview: '#06b6d4',
                description: 'Sleek and minimalist'
            },
            'dark': {
                name: 'Dark Mode',
                preview: '#121212',
                description: 'Easy on the eyes'
            }
        };

        this.currentTheme = this.getStoredTheme() || 'light';
        this.init();
    }

    init() {
        this.applyTheme(this.currentTheme);
        this.setupEventListeners();
    }

    getStoredTheme() {
        return localStorage.getItem('cvx-theme');
    }

    setStoredTheme(theme) {
        localStorage.setItem('cvx-theme', theme);
    }

    applyTheme(themeName) {
        // Remove existing theme classes/attributes
        const themes = Object.keys(this.themes);
        themes.forEach(theme => {
            document.documentElement.classList.remove(`theme-${theme}`);
        });

        // Apply new theme
        if (themeName === 'light') {
            document.documentElement.removeAttribute('data-theme');
        } else {
            document.documentElement.setAttribute('data-theme', themeName);
        }

        this.currentTheme = themeName;
        this.setStoredTheme(themeName);

        // Trigger custom event for other components
        document.dispatchEvent(new CustomEvent('theme-changed', {
            detail: { theme: themeName }
        }));
    }

    setupEventListeners() {
        // Listen for theme change events from Alpine components
        document.addEventListener('change-theme', (event) => {
            this.applyTheme(event.detail.theme);
        });
    }

    getThemeList() {
        return this.themes;
    }

    getCurrentTheme() {
        return this.currentTheme;
    }
}

// Initialize theme manager
window.themeManager = new ThemeManager();

// Alpine.js component for theme switcher
document.addEventListener('alpine:init', () => {
    Alpine.data('themeSwitcher', () => ({
        isOpen: false,
        currentTheme: window.themeManager.getCurrentTheme(),

        init() {
            // Listen for theme changes
            document.addEventListener('theme-changed', (event) => {
                this.currentTheme = event.detail.theme;
            });

            // Close dropdown when clicking outside
            document.addEventListener('click', (event) => {
                if (!this.$el.contains(event.target)) {
                    this.isOpen = false;
                }
            });
        },

        toggleDropdown() {
            this.isOpen = !this.isOpen;
        },

        selectTheme(themeName) {
            window.themeManager.applyTheme(themeName);
            this.currentTheme = themeName;
            this.isOpen = false;
        },

        getCurrentThemeName() {
            const themes = window.themeManager.getThemeList();
            return themes[this.currentTheme]?.name || 'Light';
        }
    }));
});