package communication

import (
	"path/filepath"
	"strings"
)

// SubscriptionMatcher handles pattern matching for pub/sub subscriptions
type SubscriptionMatcher struct {
}

// NewSubscriptionMatcher creates a new subscription matcher
func NewSubscriptionMatcher() *SubscriptionMatcher {
	return &SubscriptionMatcher{}
}

// MatchesPattern checks if an event name matches a glob pattern
// Supports patterns like:
// - "state.*" matches "state.changed", "state.updated"
// - "task.completed" matches exactly "task.completed"
// - "*" matches everything
// - "*.error" matches "validation.error", "processing.error"
func (sm *SubscriptionMatcher) MatchesPattern(eventName, pattern string) bool {
	// Use filepath.Match for glob-style pattern matching
	matched, err := filepath.Match(pattern, eventName)
	if err != nil {
		// If pattern is invalid, return false
		return false
	}
	return matched
}

// MatchesSubscription checks if a publication matches a subscription
func (sm *SubscriptionMatcher) MatchesSubscription(pub *Publication, sub *Subscription) bool {
	// Check if subscription is active
	if !sub.Active {
		return false
	}

	// Check publisher agent ID filter (if specified)
	if sub.PublisherAgentID != nil && *sub.PublisherAgentID != "" {
		if pub.PublisherAgentID != *sub.PublisherAgentID {
			return false
		}
	}

	// Check publisher agent type filter (if specified)
	if sub.PublisherAgentType != nil && *sub.PublisherAgentType != "" {
		if pub.PublisherAgentType != *sub.PublisherAgentType {
			return false
		}
	}

	// Check publication type filter (if specified)
	if len(sub.PublicationTypes) > 0 {
		matched := false
		for _, pubType := range sub.PublicationTypes {
			if pub.PublicationType == pubType {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check event pattern
	if !sm.MatchesPattern(pub.EventName, sub.EventPattern) {
		return false
	}

	// Check filter conditions (if any)
	if len(sub.FilterConditions) > 0 {
		if !sm.matchesFilterConditions(pub, sub.FilterConditions) {
			return false
		}
	}

	return true
}

// matchesFilterConditions checks if publication payload matches filter conditions
func (sm *SubscriptionMatcher) matchesFilterConditions(pub *Publication, conditions map[string]interface{}) bool {
	// For each condition, check if the publication payload contains matching value
	for key, expectedValue := range conditions {
		// Check if key exists in payload
		actualValue, exists := pub.Payload[key]
		if !exists {
			return false
		}

		// Check if values match (simple equality check for now)
		if actualValue != expectedValue {
			// For string values, try case-insensitive comparison
			actualStr, actualIsString := actualValue.(string)
			expectedStr, expectedIsString := expectedValue.(string)
			if actualIsString && expectedIsString {
				if !strings.EqualFold(actualStr, expectedStr) {
					return false
				}
			} else {
				return false
			}
		}
	}

	return true
}

// FilterMatchingPublications filters publications based on subscriptions
func (sm *SubscriptionMatcher) FilterMatchingPublications(publications []*Publication, subscriptions []*Subscription) []*Publication {
	var matched []*Publication

	for _, pub := range publications {
		for _, sub := range subscriptions {
			if sm.MatchesSubscription(pub, sub) {
				matched = append(matched, pub)
				break // Only add publication once even if it matches multiple subscriptions
			}
		}
	}

	return matched
}

// GetMatchingSubscriptions returns all subscriptions that match a publication
func (sm *SubscriptionMatcher) GetMatchingSubscriptions(pub *Publication, subscriptions []*Subscription) []*Subscription {
	var matched []*Subscription

	for _, sub := range subscriptions {
		if sm.MatchesSubscription(pub, sub) {
			matched = append(matched, sub)
		}
	}

	return matched
}
