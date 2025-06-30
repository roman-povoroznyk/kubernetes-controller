// pkg/config/validation.go
package config

import (
	"fmt"
	"net"
	"regexp"
	"time"

	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/errors"
)

// ConfigValidator validates configuration
type ConfigValidator struct {
	config *Config
}

// NewConfigValidator creates a new config validator
func NewConfigValidator(config *Config) *ConfigValidator {
	return &ConfigValidator{config: config}
}

// ValidateAll validates all configuration sections
func (v *ConfigValidator) ValidateAll() error {
	if err := v.ValidateGeneral(); err != nil {
		return err
	}
	
	if err := v.ValidateController(); err != nil {
		return err
	}
	
	if err := v.ValidateMultiCluster(); err != nil {
		return err
	}
	
	if err := v.ValidateNetwork(); err != nil {
		return err
	}
	
	return nil
}

// ValidateGeneral validates general configuration
func (v *ConfigValidator) ValidateGeneral() error {
	// Validate log level
	validLogLevels := []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}
	if !v.isValidLogLevel(v.config.LogLevel) {
		return errors.NewValidationError(fmt.Sprintf("invalid log level '%s', must be one of: %v", v.config.LogLevel, validLogLevels))
	}
	
	return nil
}

// ValidateController validates controller configuration
func (v *ConfigValidator) ValidateController() error {
	// Validate controller mode
	if v.config.Controller.Mode != "single" && v.config.Controller.Mode != "multi" {
		return errors.NewValidationError(fmt.Sprintf("invalid controller mode '%s', must be 'single' or 'multi'", v.config.Controller.Mode))
	}
	
	// Validate single cluster configuration
	if v.config.Controller.Mode == "single" || v.config.Controller.Mode == "" {
		if err := v.validateSingleCluster(); err != nil {
			return err
		}
	}
	
	// Validate resync period
	if v.config.Controller.ResyncPeriod < time.Second {
		return errors.NewValidationError(fmt.Sprintf("resync period must be at least 1 second, got %v", v.config.Controller.ResyncPeriod))
	}
	
	return nil
}

// ValidateMultiCluster validates multi-cluster configuration
func (v *ConfigValidator) ValidateMultiCluster() error {
	// Validate connection timeout
	if v.config.MultiCluster.ConnectionTimeout < time.Second {
		return errors.NewValidationError(fmt.Sprintf("connection timeout must be at least 1 second, got %v", v.config.MultiCluster.ConnectionTimeout))
	}
	
	// Validate max concurrent connections
	if v.config.MultiCluster.MaxConcurrentConns < 1 || v.config.MultiCluster.MaxConcurrentConns > 1000 {
		return errors.NewValidationError(fmt.Sprintf("max concurrent connections must be between 1 and 1000, got %d", v.config.MultiCluster.MaxConcurrentConns))
	}
	
	// Validate clusters
	if len(v.config.MultiCluster.Clusters) == 0 && v.config.Controller.Mode == "multi" {
		return errors.NewValidationError("multi-cluster mode requires at least one cluster configuration")
	}
	
	primaryCount := 0
	for i, cluster := range v.config.MultiCluster.Clusters {
		if err := v.validateCluster(i, cluster); err != nil {
			return err
		}
		if cluster.Primary {
			primaryCount++
		}
	}
	
	if len(v.config.MultiCluster.Clusters) > 0 && primaryCount == 0 {
		return errors.NewValidationError("at least one cluster must be marked as primary")
	}
	
	if primaryCount > 1 {
		return errors.NewValidationError(fmt.Sprintf("only one cluster can be marked as primary, found %d", primaryCount))
	}
	
	return nil
}

// ValidateNetwork validates network-related configuration
func (v *ConfigValidator) ValidateNetwork() error {
	// Validate ports
	if err := v.validatePort("metrics port", v.config.Controller.Single.MetricsPort); err != nil {
		return err
	}
	
	if err := v.validatePort("health port", v.config.Controller.Single.HealthPort); err != nil {
		return err
	}
	
	// Check for port conflicts
	if v.config.Controller.Single.MetricsPort == v.config.Controller.Single.HealthPort {
		return errors.NewValidationError("metrics port and health port cannot be the same")
	}
	
	return nil
}

// validateSingleCluster validates single cluster configuration
func (v *ConfigValidator) validateSingleCluster() error {
	// Validate namespace (if specified)
	if v.config.Controller.Single.Namespace != "" {
		if !v.isValidKubernetesName(v.config.Controller.Single.Namespace) {
			return errors.NewValidationError(fmt.Sprintf("invalid namespace name '%s'", v.config.Controller.Single.Namespace))
		}
	}
	
	// Validate leader election
	if v.config.Controller.Single.LeaderElection.Enabled {
		if v.config.Controller.Single.LeaderElection.ID == "" {
			return errors.NewValidationError("leader election ID is required when leader election is enabled")
		}
		
		if !v.isValidKubernetesName(v.config.Controller.Single.LeaderElection.ID) {
			return errors.NewValidationError(fmt.Sprintf("invalid leader election ID '%s'", v.config.Controller.Single.LeaderElection.ID))
		}
		
		if v.config.Controller.Single.LeaderElection.Namespace != "" {
			if !v.isValidKubernetesName(v.config.Controller.Single.LeaderElection.Namespace) {
				return errors.NewValidationError(fmt.Sprintf("invalid leader election namespace '%s'", v.config.Controller.Single.LeaderElection.Namespace))
			}
		}
	}
	
	return nil
}

// validateCluster validates a single cluster configuration
func (v *ConfigValidator) validateCluster(index int, cluster ClusterConfig) error {
	if cluster.Name == "" {
		return errors.NewValidationError(fmt.Sprintf("cluster at index %d is missing name", index))
	}
	
	if !v.isValidKubernetesName(cluster.Name) {
		return errors.NewValidationError(fmt.Sprintf("invalid cluster name '%s' at index %d", cluster.Name, index))
	}
	
	if cluster.Namespace != "" && !v.isValidKubernetesName(cluster.Namespace) {
		return errors.NewValidationError(fmt.Sprintf("invalid namespace '%s' for cluster '%s'", cluster.Namespace, cluster.Name))
	}
	
	return nil
}

// validatePort validates a port number
func (v *ConfigValidator) validatePort(name string, port int) error {
	if port < 1 || port > 65535 {
		return errors.NewValidationError(fmt.Sprintf("%s must be between 1 and 65535, got %d", name, port))
	}
	
	// Check if port is privileged (< 1024) and warn
	if port < 1024 {
		// This could be a warning instead of an error
		// For now, we'll allow it but could add logging
	}
	
	return nil
}

// isValidLogLevel checks if log level is valid
func (v *ConfigValidator) isValidLogLevel(level string) bool {
	validLevels := []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}
	for _, valid := range validLevels {
		if level == valid {
			return true
		}
	}
	return false
}

// isValidKubernetesName checks if a name is valid for Kubernetes resources
func (v *ConfigValidator) isValidKubernetesName(name string) bool {
	// Kubernetes resource names must be lowercase alphanumeric characters or '-'
	// and must start and end with an alphanumeric character
	pattern := `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	matched, _ := regexp.MatchString(pattern, name)
	return matched && len(name) <= 253
}

// isValidIP checks if a string is a valid IP address
func (v *ConfigValidator) isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// isValidPort checks if a port number is valid
func (v *ConfigValidator) isValidPort(port int) bool {
	return port >= 1 && port <= 65535
}

// ValidateAndReport validates configuration and returns detailed error report
func (v *ConfigValidator) ValidateAndReport() *ValidationReport {
	report := &ValidationReport{
		Valid:    true,
		Warnings: []string{},
		Errors:   []string{},
	}
	
	if err := v.ValidateAll(); err != nil {
		report.Valid = false
		report.Errors = append(report.Errors, err.Error())
	}
	
	// Add warnings for potential issues
	v.addWarnings(report)
	
	return report
}

// addWarnings adds warnings for potential configuration issues
func (v *ConfigValidator) addWarnings(report *ValidationReport) {
	// Warn about privileged ports
	if v.config.Controller.Single.MetricsPort < 1024 {
		report.Warnings = append(report.Warnings, fmt.Sprintf("metrics port %d is privileged and may require root access", v.config.Controller.Single.MetricsPort))
	}
	
	if v.config.Controller.Single.HealthPort < 1024 {
		report.Warnings = append(report.Warnings, fmt.Sprintf("health port %d is privileged and may require root access", v.config.Controller.Single.HealthPort))
	}
	
	// Warn about disabled leader election in production
	if !v.config.Controller.Single.LeaderElection.Enabled {
		report.Warnings = append(report.Warnings, "leader election is disabled, this may cause issues in high-availability setups")
	}
	
	// Warn about too many concurrent connections
	if v.config.MultiCluster.MaxConcurrentConns > 100 {
		report.Warnings = append(report.Warnings, fmt.Sprintf("max concurrent connections (%d) is very high and may cause resource exhaustion", v.config.MultiCluster.MaxConcurrentConns))
	}
	
	// Warn about short connection timeout
	if v.config.MultiCluster.ConnectionTimeout < 10*time.Second {
		report.Warnings = append(report.Warnings, fmt.Sprintf("connection timeout (%v) is very short and may cause connection issues", v.config.MultiCluster.ConnectionTimeout))
	}
}

// ValidationReport contains validation results
type ValidationReport struct {
	Valid    bool     `json:"valid"`
	Warnings []string `json:"warnings,omitempty"`
	Errors   []string `json:"errors,omitempty"`
}
