// AI Policy Wizard - Alpine.js Component

window.policyWizard = function () {
    return {
        // Wizard state
        currentStep: 1,
        saving: false,
        saveSuccess: false,
        saveError: false,
        errorMessage: '',
        agencyID: '',

        // UI helpers
        frameworksSelected: {
            soc2: false,
            gdpr: false,
            hipaa: false,
            iso27001: false
        },
        prohibitedActionsText: '',

        // Policy data structure
        policy: {
            agency_id: '',
            version: '1.0',
            owner: '',
            stance: {
                adoption_level: 'controlled',
                risk_tolerance: 'medium',
                compliance_frameworks: []
            },
            models: {
                allowed_providers: [],
                denied_providers: [],
                fallback_behavior: 'fail_safe'
            },
            autonomy: {
                default_level: 'L2',
                role_overrides: [],
                escalation_rules: []
            },
            data_access: {
                classification_required: false,
                rules: [
                    {
                        classification: 'public',
                        allowed_operations: ['read', 'write'],
                        retention_days: 30,
                        requires_justification: false,
                        requires_approval: false,
                        audit_all_access: false,
                        explicit_grant_required: false,
                        dual_approval: false
                    },
                    {
                        classification: 'internal',
                        allowed_operations: ['read', 'write'],
                        retention_days: 90,
                        requires_justification: false,
                        requires_approval: false,
                        audit_all_access: false,
                        explicit_grant_required: false,
                        dual_approval: false
                    },
                    {
                        classification: 'confidential',
                        allowed_operations: ['read'],
                        retention_days: 365,
                        requires_justification: true,
                        requires_approval: true,
                        audit_all_access: true,
                        explicit_grant_required: true,
                        dual_approval: false
                    },
                    {
                        classification: 'restricted',
                        allowed_operations: ['read'],
                        retention_days: 730,
                        requires_justification: true,
                        requires_approval: true,
                        audit_all_access: true,
                        explicit_grant_required: true,
                        dual_approval: true
                    }
                ],
                pii_handling: {
                    detect_pii: false,
                    anonymization_required: false,
                    cross_border_transfer: 'require_approval',
                    deletion_on_request: false
                }
            },
            actions: {
                approval_workflows: [],
                prohibited_actions: [],
                rollback_requirements: []
            },
            risk: {
                scoring_enabled: false,
                thresholds: {
                    low: 25,
                    medium: 50,
                    high: 75,
                    critical: 90
                },
                mitigation: {}
            },
            compliance: {
                frameworks: [],
                audit_requirements: {
                    log_all_actions: false,
                    immutable_audit_log: false,
                    retention_years: 7
                },
                reporting: {
                    daily_summary: false,
                    weekly_compliance_report: false,
                    quarterly_risk_assessment: false
                }
            },
            monitoring: {
                real_time_policy_violations: false,
                alerts: []
            }
        },

        init() {
            const container = document.querySelector('.policy-wizard-container');
            this.agencyID = container.dataset.agencyId;
            this.policy.agency_id = this.agencyID;

            // Get current user from session (if available)
            // For MVP, use a placeholder
            this.policy.owner = 'admin'; // TODO: Get from session

            // Load existing policy if present
            const existingPolicyData = container.dataset.existingPolicy;
            if (existingPolicyData) {
                try {
                    const existingPolicy = JSON.parse(existingPolicyData);
                    this.loadExistingPolicy(existingPolicy);
                } catch (e) {
                    console.error('Failed to parse existing policy:', e);
                }
            } else {
                // Initialize with one default provider
                this.addProvider();
            }

            // Sync frameworks
            this.syncFrameworks();
        },

        loadExistingPolicy(existingPolicy) {
            this.policy = existingPolicy;

            // Sync UI helpers
            if (existingPolicy.stance && existingPolicy.stance.compliance_frameworks) {
                existingPolicy.stance.compliance_frameworks.forEach(fw => {
                    const key = fw.toLowerCase().replace(/\s/g, '');
                    if (this.frameworksSelected.hasOwnProperty(key)) {
                        this.frameworksSelected[key] = true;
                    }
                });
            }

            // Convert models array to text for each provider
            if (existingPolicy.models && existingPolicy.models.allowed_providers) {
                existingPolicy.models.allowed_providers.forEach(provider => {
                    provider.modelsText = provider.models ? provider.models.join(', ') : '';
                });
            }

            // Convert prohibited actions array to text
            if (existingPolicy.actions && existingPolicy.actions.prohibited_actions) {
                this.prohibitedActionsText = existingPolicy.actions.prohibited_actions.join(', ');
            }
        },

        // Navigation
        nextStep() {
            if (this.currentStep < 6) {
                this.currentStep++;
            }
        },

        previousStep() {
            if (this.currentStep > 1) {
                this.currentStep--;
            }
        },

        goToStep(step) {
            if (step >= 1 && step <= 6) {
                this.currentStep = step;
            }
        },

        // Validation
        isValid() {
            // Basic validation
            if (!this.policy.agency_id) return false;
            if (!this.policy.owner) return false;
            if (this.policy.models.allowed_providers.length === 0) return false;
            return true;
        },

        // Provider Management
        addProvider() {
            this.policy.models.allowed_providers.push({
                provider: '',
                models: [],
                modelsText: '',
                data_residency: '',
                max_tokens_per_request: 4000,
                monthly_budget_usd: 0,
                current_spend_usd: 0
            });
        },

        removeProvider(index) {
            this.policy.models.allowed_providers.splice(index, 1);
        },

        updateProviderModels(index) {
            const provider = this.policy.models.allowed_providers[index];
            if (provider.modelsText) {
                provider.models = provider.modelsText.split(',').map(m => m.trim()).filter(m => m);
            } else {
                provider.models = [];
            }
        },

        // Role Override Management
        addRoleOverride() {
            this.policy.autonomy.role_overrides.push({
                role_id: '',
                role_name: '',
                level: 'L2',
                justification: '',
                requires_approval_from: []
            });
        },

        removeRoleOverride(index) {
            this.policy.autonomy.role_overrides.splice(index, 1);
        },

        // Compliance Frameworks
        syncFrameworks() {
            this.policy.stance.compliance_frameworks = [];
            Object.keys(this.frameworksSelected).forEach(key => {
                if (this.frameworksSelected[key]) {
                    let name = key.toUpperCase();
                    if (key === 'iso27001') name = 'ISO 27001';
                    this.policy.stance.compliance_frameworks.push(name);
                }
            });

            // Update compliance.frameworks
            this.policy.compliance.frameworks = this.policy.stance.compliance_frameworks.map(name => ({
                name: name,
                controls: [],
                enabled: true
            }));
        },

        // Prohibited Actions
        updateProhibitedActions() {
            if (this.prohibitedActionsText) {
                this.policy.actions.prohibited_actions = this.prohibitedActionsText
                    .split(',')
                    .map(a => a.trim())
                    .filter(a => a);
            } else {
                this.policy.actions.prohibited_actions = [];
            }
        },

        // Save Policy
        async savePolicy() {
            if (!this.isValid()) {
                this.errorMessage = 'Please fill in all required fields';
                this.saveError = true;
                setTimeout(() => this.saveError = false, 5000);
                return;
            }

            this.saving = true;
            this.saveSuccess = false;
            this.saveError = false;

            // Sync frameworks before saving
            this.syncFrameworks();

            try {
                const response = await fetch(`/api/agencies/${this.agencyID}/policy`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(this.policy)
                });

                if (!response.ok) {
                    const error = await response.json();
                    const errorMsg = error.error || 'Failed to save policy';

                    // Navigate to relevant step based on error message
                    this.handleValidationError(errorMsg);

                    throw new Error(errorMsg);
                }

                const result = await response.json();
                console.log('Policy saved:', result);

                this.saveSuccess = true;
                setTimeout(() => this.saveSuccess = false, 5000);

                // Update policy with server response
                if (result.policy) {
                    this.policy = result.policy;
                }
            } catch (error) {
                console.error('Save error:', error);
                this.errorMessage = error.message;
                this.saveError = true;
                setTimeout(() => this.saveError = false, 5000);
            } finally {
                this.saving = false;
            }
        },

        // Handle validation errors by navigating to relevant step
        handleValidationError(errorMsg) {
            const lowerMsg = errorMsg.toLowerCase();

            // Map error messages to steps
            if (lowerMsg.includes('adoption') || lowerMsg.includes('risk tolerance') || lowerMsg.includes('stance')) {
                this.currentStep = 1;
            } else if (lowerMsg.includes('provider') || lowerMsg.includes('model')) {
                this.currentStep = 2;
            } else if (lowerMsg.includes('autonomy') || lowerMsg.includes('role override')) {
                this.currentStep = 3;
            } else if (lowerMsg.includes('data') || lowerMsg.includes('classification') || lowerMsg.includes('pii')) {
                this.currentStep = 4;
            } else if (lowerMsg.includes('action') || lowerMsg.includes('risk') || lowerMsg.includes('threshold')) {
                this.currentStep = 5;
            } else if (lowerMsg.includes('compliance') || lowerMsg.includes('audit') || lowerMsg.includes('monitoring')) {
                this.currentStep = 6;
            }
        }
    };
};
