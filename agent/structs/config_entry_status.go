package structs

import (
	"time"

	"github.com/hashicorp/consul/acl"
)

// ResourceReference is a reference to a ConfigEntry
// with an optional reference to a subsection of that ConfigEntry
// that can be specified as SectionName
type ResourceReference struct {
	// Kind is the kind of ConfigEntry that this resource refers to.
	Kind string
	// Name is the identifier for the ConfigEntry this resource refers to.
	Name string
	// SectionName is a generic subresource identifier that specifies
	// a subset of the ConfigEntry to which this reference applies. Usage
	// of this field should be up to the controller that leverages it. If
	// unused, this should be blank.
	SectionName string

	acl.EnterpriseMeta
}

// Status is used for propagating back asynchronously calculated
// messages from control loops to a user
type Status struct {
	// Conditions is the set of condition objects associated with
	// a ConfigEntry status.
	Conditions []Condition
}

// Condition is used for a single message and state associated
// with an object. For example, a ConfigEntry that references
// multiple other resources may have different statuses with
// respect to each of those resources.
type Condition struct {
	// Type is a value from a bounded set of condition types for a given controlled object
	Type string
	// Status is a value from a bounded set of statuses that an object might have
	Status ConditionStatus
	// Reason is a value from a bounded set of reasons for a given type
	Reason string
	// Message is a message that gives more detailed information about
	// why a Condition has a given status and reason
	Message string
	// Resource is an optional reference to a resource for which this
	// condition applies
	Resource *ResourceReference
	// LastTransitionTime is the time at which this Condition was created
	LastTransitionTime *time.Time
}

// ConditionStatus is a bounded set of statuses that an object might have
type ConditionStatus string

const (
	ConditionStatusTrue    ConditionStatus = "True"
	ConditionStatusFalse   ConditionStatus = "False"
	ConditionStatusUnknown ConditionStatus = "Unknown"
)

func checkConditionStatus(status ConditionStatus) error {
	switch status {
	case ConditionStatusTrue:
	case ConditionStatusFalse:
	case ConditionStatusUnknown:
		return nil
	default:
		return fmt.Errorf("unrecognized ConditionStatus %s", status)
	}
}

type conditionReasons struct {
    ConditionStatusTrue: []string,
    ConditionStatusFalse: []string,
    ConditionStatusUnknown: []string,
}

// RouteConditionType is a type of condition for a route.
type RouteConditionType string

// RouteConditionReason is a reason for a route condition.
type RouteConditionReason string

const (
	// This condition indicates whether the route has been accepted or rejected
	// by a Gateway, and why.
	//
	// Possible reasons for this condition to be true are:
	//
	// * "Accepted"
	//
	// Possible reasons for this condition to be False are:
	//
	// * "NotAllowedByListeners"
	// * "NoMatchingListenerHostname"
	// * "NoMatchingParent"
	// * "UnsupportedValue"
	// * "ParentRefNotPermitted"
	//
	// Possible reasons for this condition to be Unknown are:
	//
	// * "Pending"
	//
	// Controllers may raise this condition with other reasons,
	// but should prefer to use the reasons listed above to improve
	// interoperability.
	RouteConditionAccepted RouteConditionType = "Accepted"

	// This reason is used with the "Accepted" condition when the Route has been
	// accepted by the Gateway.
	RouteReasonAccepted RouteConditionReason = "Accepted"

	// This reason is used with the "Accepted" condition when the route has not
	// been accepted by a Gateway because the Gateway has no Listener whose
	// allowedRoutes criteria permit the route
	RouteReasonNotAllowedByListeners RouteConditionReason = "NotAllowedByListeners"

	// This reason is used with the "Accepted" condition when the Gateway has no
	// compatible Listeners whose Hostname matches the route
	RouteReasonNoMatchingListenerHostname RouteConditionReason = "NoMatchingListenerHostname"

	// This reason is used with the "Accepted" condition when there are
	// no matching Parents. In the case of Gateways, this can occur when
	// a Route ParentRef specifies a Port and/or SectionName that does not
	// match any Listeners in the Gateway.
	RouteReasonNoMatchingParent RouteConditionReason = "NoMatchingParent"

	// This reason is used with the "Accepted" condition when a value for an Enum
	// is not recognized.
	RouteReasonUnsupportedValue RouteConditionReason = "UnsupportedValue"

	// This reason is used with the "Accepted" condition when the route has not
	// been accepted by a Gateway because it has a cross-namespace parentRef,
	// but no ReferenceGrant in the other namespace allows such a reference.
	RouteReasonParentRefNotPermitted RouteConditionReason = "ParentRefNotPermitted"

	// This reason is used with the "Accepted" when a controller has not yet
	// reconciled the route.
	RouteReasonPending RouteConditionReason = "Pending"

	// This condition indicates whether the controller was able to resolve all
	// the object references for the Route.
	//
	// Possible reasons for this condition to be true are:
	//
	// * "ResolvedRefs"
	//
	// Possible reasons for this condition to be false are:
	//
	// * "RefNotPermitted"
	// * "InvalidKind"
	// * "BackendNotFound"
	//
	// Controllers may raise this condition with other reasons,
	// but should prefer to use the reasons listed above to improve
	// interoperability.
	RouteConditionResolvedRefs RouteConditionType = "ResolvedRefs"

	// This reason is used with the "ResolvedRefs" condition when the condition
	// is true.
	RouteReasonResolvedRefs RouteConditionReason = "ResolvedRefs"

	// This reason is used with the "ResolvedRefs" condition when
	// one of the Listener's Routes has a BackendRef to an object in
	// another namespace, where the object in the other namespace does
	// not have a ReferenceGrant explicitly allowing the reference.
	RouteReasonRefNotPermitted RouteConditionReason = "RefNotPermitted"

	// This reason is used with the "ResolvedRefs" condition when
	// one of the Route's rules has a reference to an unknown or unsupported
	// Group and/or Kind.
	RouteReasonInvalidKind RouteConditionReason = "InvalidKind"

	// This reason is used with the "ResolvedRefs" condition when one of the
	// Route's rules has a reference to a resource that does not exist.
	RouteReasonBackendNotFound RouteConditionReason = "BackendNotFound"
)

// NewRouteCondition is a helper to build allowable Conditions for a Route config entry
func NewRouteCondition(name RouteConditionType, status ConditionStatus, reason RouteConditionReason, message string) Condition {
	if err = checkRouteConditionReason(name, status, reason); err != nil {
		panic(err)
	}

	return Condition{
        Type:               name,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: time.Now(),
	}
}

func checkRouteConditionReason(name RouteConditionType, status ConditionStatus, reason RouteConditionReason) error {
	if err := checkConditionStatus(status); err != nil {
		return err
	}

    reasons, ok := routeConditionReasons[name]; if !ok {
		return fmt.Errorf("unrecognized RouteConditionType %s", name)
    }

    if !slices.Contains(reasons[status], reason) {
        return fmt.Errorf("route condition reason %s not allowed for route condition type %s with status %s", reason, name, status)
	}

	return nil
}

var routeConditionReasons {
    RouteConditionAccepted: conditionReasons{
        ConditionStatusTrue: [
            RouteConditionReasonAccepted
        ],
        ConditionStatusFalse: [
            RouteReasonNotAllowedByListeners,
            RouteReasonNoMatchingListenerHostname,
            RouteReasonNoMatchingParent,
            RouteReasonUnsupportedValue,
            RouteReasonParentRefNotPermitted,
        ],
        ConditionStatusUnknown: [
            RouteReasonPending,
        ],
    },
    RouteConditionResolvedRefs: conditionReasons{
        ConditionStatusTrue: [
            RouteReasonResolvedRefs,
        ],
        ConditionStatusFalse: [
            RouteReasonRefNotPermitted,
            RouteReasonInvalidKind,
            RouteReasonBackendNotFound,
        ],
        ConditionStatusUnknown: [
        ],
    }
}
