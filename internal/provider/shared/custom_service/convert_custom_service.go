package customservice

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	portrange "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/port_range"
)

//nolint:gocyclo,funlen
func ParseCustomService(ctx context.Context, item interface{}, attrName string) types.Object {
	tflog.Debug(ctx, "customservice.ParseCustomService() "+attrName+" - "+fmt.Sprintf("%v", item))

	// Get the reflect.Value of the input
	itemValue := reflect.ValueOf(item)

	// Handle nil or invalid input
	if item == nil || itemValue.Kind() != reflect.Struct {
		if itemValue.Kind() == reflect.Pointer && !itemValue.IsNil() {
			itemValue = itemValue.Elem()
			// Keep dereferencing until we get to a struct or can't anymore
			for itemValue.Kind() == reflect.Pointer && !itemValue.IsNil() {
				itemValue = itemValue.Elem()
			}
			if itemValue.Kind() != reflect.Struct {
				return types.ObjectNull(CustomServiceAttrTypes)
			}
		} else {
			return types.ObjectNull(CustomServiceAttrTypes)
		}
	}

	// Handle pointer to struct
	if itemValue.Kind() == reflect.Pointer {
		tflog.Debug(ctx, "customservice.ParseCustomService() itemValue.Kind()- "+fmt.Sprintf("%v", itemValue))
		if itemValue.IsNil() {
			return types.ObjectNull(CustomServiceAttrTypes)
		}
		itemValue = itemValue.Elem()
		tflog.Debug(ctx, "customservice.ParseCustomService() itemValue.Elem()- "+fmt.Sprintf("%v", itemValue))
	}

	// Get fields
	portField := itemValue.FieldByName("Port")
	protocolField := itemValue.FieldByName("Protocol")
	portRangeField := itemValue.FieldByName("PortRange")

	tflog.Debug(ctx, "customservice.ParseCustomService() portField - "+fmt.Sprintf("%v", portField))
	tflog.Debug(ctx, "customservice.ParseCustomService() protocolField - "+fmt.Sprintf("%v", protocolField))
	tflog.Debug(ctx, "customservice.ParseCustomService() portRangeField - "+fmt.Sprintf("%v", portRangeField))
	// Handle port_range first to check if it's set
	var portRangeVal types.Object
	var hasPortRange bool
	if portRangeField.Kind() == reflect.Pointer {
		if portRangeField.IsNil() {
			portRangeVal = types.ObjectNull(portrange.PortRangeAttrTypes)
			hasPortRange = false
		} else {
			portRangeField = portRangeField.Elem()
			hasPortRange = true
		}
	}
	if portRangeField.IsValid() && hasPortRange {
		from := portRangeField.FieldByName("From")
		to := portRangeField.FieldByName("To")
		var diagsTmp diag.Diagnostics
		portRangeVal, diagsTmp = types.ObjectValue(
			portrange.PortRangeAttrTypes,
			map[string]attr.Value{
				"from": types.StringValue(from.String()),
				"to":   types.StringValue(to.String()),
			},
		)
		_ = diagsTmp
	} else if !hasPortRange {
		portRangeVal = types.ObjectNull(portrange.PortRangeAttrTypes)
	}

	// Handle port field - match config logic behavior
	// When port_range is set, port should always be null even if API returns empty array
	var portList types.List
	switch {
	case hasPortRange:
		// When port_range is present, port must be null to match schema validation
		portList = types.ListNull(types.StringType)
	case portField.IsValid() && portField.Kind() == reflect.Slice:
		// Only process port field if port_range is not set
		switch {
		case portField.IsNil():
			// Source data was null - don't include port field in state to match config
			portList = types.ListNull(types.StringType)
		case portField.Len() == 0:
			// Source data was empty array [] - treat as null to avoid inconsistency
			// Note: empty array from API likely means port_range is set
			portList = types.ListNull(types.StringType)
		default:
			// Source data has values - include them
			ports := make([]attr.Value, portField.Len())
			for i := range portField.Len() {
				portValue := portField.Index(i)
				var portStr string
				switch portValue.Kind() {
				case reflect.String:
					portStr = portValue.String()
				case reflect.Int, reflect.Int64:
					portStr = fmt.Sprintf("%d", portValue.Int())
				default:
					tflog.Info(ctx, "customservice.ParseCustomService() unsupported port type - "+portValue.Kind().String())
					portStr = fmt.Sprintf("%v", portValue.Interface())
				}
				ports[i] = types.StringValue(portStr)
			}
			var diagsTmp diag.Diagnostics
			portList, diagsTmp = types.ListValue(types.StringType, ports)
			_ = diagsTmp
		}
	default:
		// Invalid or non-slice field
		portList = types.ListNull(types.StringType)
	}

	// Handle protocol
	var protocolVal types.String
	if protocolField.IsValid() {
		protocolVal = types.StringValue(protocolField.String())
	} else {
		protocolVal = types.StringNull()
	}

	// Create final custom service object
	obj, diagstmp := types.ObjectValue(
		CustomServiceAttrTypes,
		map[string]attr.Value{
			"port":       portList,
			"port_range": portRangeVal,
			"protocol":   protocolVal,
		},
	)
	tflog.Debug(ctx, "customservice.ParseCustomService() obj - "+obj.String())
	_ = diagstmp
	return obj
}

// ParseExceptionCustomService handles custom services in exceptions which use "portRangeCustomService" field instead of "portRange"
//
//nolint:gocyclo,funlen
func ParseExceptionCustomService(ctx context.Context, item interface{}, attrName string) types.Object {
	tflog.Debug(ctx, "parseExceptionCustomService() "+attrName+" - "+fmt.Sprintf("%v", item))

	// Get the reflect.Value of the input
	itemValue := reflect.ValueOf(item)

	// Handle nil or invalid input
	if item == nil || itemValue.Kind() != reflect.Struct {
		if itemValue.Kind() == reflect.Pointer && !itemValue.IsNil() {
			itemValue = itemValue.Elem()
			// Keep dereferencing until we get to a struct or can't anymore
			for itemValue.Kind() == reflect.Pointer && !itemValue.IsNil() {
				itemValue = itemValue.Elem()
			}
			if itemValue.Kind() != reflect.Struct {
				return types.ObjectNull(CustomServiceAttrTypes)
			}
		} else {
			return types.ObjectNull(CustomServiceAttrTypes)
		}
	}

	// Handle pointer to struct
	if itemValue.Kind() == reflect.Pointer {
		tflog.Debug(ctx, "parseExceptionCustomService() itemValue.Kind()- "+fmt.Sprintf("%v", itemValue))
		if itemValue.IsNil() {
			return types.ObjectNull(CustomServiceAttrTypes)
		}
		itemValue = itemValue.Elem()
		tflog.Debug(ctx, "parseExceptionCustomService() itemValue.Elem()- "+fmt.Sprintf("%v", itemValue))
	}

	// Get fields - note that exceptions use "PortRangeCustomService" instead of "PortRange"
	portField := itemValue.FieldByName("Port")
	protocolField := itemValue.FieldByName("Protocol")
	portRangeField := itemValue.FieldByName("PortRangeCustomService") // Different field name for exceptions

	tflog.Debug(ctx, "parseExceptionCustomService() portField - "+fmt.Sprintf("%v", portField))
	tflog.Debug(ctx, "parseExceptionCustomService() protocolField - "+fmt.Sprintf("%v", protocolField))
	tflog.Debug(ctx, "parseExceptionCustomService() portRangeField - "+fmt.Sprintf("%v", portRangeField))

	// Handle port_range first (from PortRangeCustomService field) to check if it's set
	var portRangeVal types.Object
	var hasPortRange bool
	if portRangeField.Kind() == reflect.Pointer {
		if portRangeField.IsNil() {
			portRangeVal = types.ObjectNull(portrange.PortRangeAttrTypes)
			hasPortRange = false
		} else {
			portRangeField = portRangeField.Elem()
			hasPortRange = true
		}
	}
	if portRangeField.IsValid() && !portRangeField.IsZero() && hasPortRange {
		from := portRangeField.FieldByName("From")
		to := portRangeField.FieldByName("To")
		if from.IsValid() && to.IsValid() {
			var diagsTmp diag.Diagnostics
			portRangeVal, diagsTmp = types.ObjectValue(
				portrange.PortRangeAttrTypes,
				map[string]attr.Value{
					"from": types.StringValue(from.String()),
					"to":   types.StringValue(to.String()),
				},
			)
			_ = diagsTmp
		} else {
			portRangeVal = types.ObjectNull(portrange.PortRangeAttrTypes)
			hasPortRange = false
		}
	} else if !hasPortRange {
		portRangeVal = types.ObjectNull(portrange.PortRangeAttrTypes)
	}

	// Handle port field - match config logic behavior
	// When port_range is set, port should always be null even if API returns empty array
	var portList types.List
	switch {
	case hasPortRange:
		// When port_range is present, port must be null to match schema validation
		portList = types.ListNull(types.StringType)
	case portField.IsValid() && portField.Kind() == reflect.Slice:
		// Only process port field if port_range is not set
		switch {
		case portField.IsNil():
			// Source data was null - don't include port field in state to match config
			portList = types.ListNull(types.StringType)
		case portField.Len() == 0:
			// Source data was empty array [] - treat as null to avoid inconsistency
			// Note: empty array from API likely means port_range is set
			portList = types.ListNull(types.StringType)
		default:
			// Source data has values - include them
			ports := make([]attr.Value, portField.Len())
			for i := range portField.Len() {
				portValue := portField.Index(i)
				var portStr string
				switch portValue.Kind() {
				case reflect.String:
					portStr = portValue.String()
				case reflect.Int, reflect.Int64:
					portStr = fmt.Sprintf("%d", portValue.Int())
				default:
					tflog.Info(ctx, "parseExceptionCustomService() unsupported port type - "+portValue.Kind().String())
					portStr = fmt.Sprintf("%v", portValue.Interface())
				}
				ports[i] = types.StringValue(portStr)
			}
			var diagsTmp diag.Diagnostics
			portList, diagsTmp = types.ListValue(types.StringType, ports)
			_ = diagsTmp
		}
	default:
		// Invalid or non-slice field
		portList = types.ListNull(types.StringType)
	}

	// Handle protocol
	var protocolVal types.String
	if protocolField.IsValid() {
		protocolVal = types.StringValue(protocolField.String())
	} else {
		protocolVal = types.StringNull()
	}

	// Create final custom service object
	obj, diagstmp := types.ObjectValue(
		CustomServiceAttrTypes,
		map[string]attr.Value{
			"port":       portList,
			"port_range": portRangeVal,
			"protocol":   protocolVal,
		},
	)
	tflog.Debug(ctx, "parseExceptionCustomService() obj - "+obj.String())
	_ = diagstmp
	return obj
}
