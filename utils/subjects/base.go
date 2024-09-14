package subjects

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
)

// SubjectBuilder is a struct that holds the parameters for building a NATS subject
type SubjectBuilder struct {
	GlobalUUID  string
	AppID       string
	subjectType string
}

const IsApp = "app"
const IsProxy = "proxy"
const IsBios = "bios"

// NewSubjectBuilder creates a new SubjectBuilder
func NewSubjectBuilder(globalUUID, appID string, subjectType string) *SubjectBuilder {
	if subjectType != IsApp && subjectType != IsProxy && subjectType != IsBios {
		log.Fatal(fmt.Sprintf("subject type %s is not supported try: %s or %s or %s", subjectType, IsApp, IsProxy, IsBios))
	}
	return &SubjectBuilder{
		GlobalUUID:  globalUUID,
		AppID:       appID,
		subjectType: subjectType,
	}
}

// BuildSubject builds a NATS subject based on the proxy usage
func (sb *SubjectBuilder) BuildSubject(action, resource, scope string) string {
	if sb.subjectType == IsBios {
		return fmt.Sprintf("%s.%s.%s.%s", sb.GlobalUUID, action, resource, scope)
	} else if sb.subjectType == IsApp {
		return fmt.Sprintf("%s.%s.%s.%s", sb.AppID, action, resource, scope)
	} else if sb.subjectType == IsProxy {
		return fmt.Sprintf("%s.proxy.%s.%s.%s.%s", sb.GlobalUUID, sb.AppID, action, resource, scope)
	}
	return ""
}

// BuildMessage builds a JSON message from a map
func BuildMessage(payload map[string]interface{}) (string, error) {
	msgBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(msgBytes), nil
}

// GetSubjectParts to split and retrieve everything after "proxy."
func GetSubjectParts(subject string) string {
	parts := strings.SplitN(subject, "proxy.", 2)
	if len(parts) < 2 {
		fmt.Println("No part found after 'proxy.'")
		return ""
	}
	return parts[1]
}

// GetAppID to extract appID after "proxy." or return an error
func GetAppID(subject string) (string, error) {
	// Split by "proxy."
	parts := strings.SplitN(subject, "proxy.", 2)
	if len(parts) < 2 {
		return "", errors.New("no part found after 'proxy.'")
	}

	// Split the remaining part by "."
	subParts := strings.SplitN(parts[1], ".", 2)
	if len(subParts) < 2 {
		return "", errors.New("appID not found after 'proxy.'")
	}

	// The first part should be the appID
	appID := subParts[0]
	return appID, nil
}
