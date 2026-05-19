package provider

// DEPRECATED: This file previously contained semantic equality functions for WAN firewall rules.
// These functions have been replaced by the ExceptionsSetModifier plan modifier
// which provides better handling of exceptions state management.
//
// The ExceptionsSetModifier approach is preferred because:
// 1. It works with the supported Terraform plugin framework version (1.14.1)
// 2. It preserves nested object IDs correctly
// 3. It avoids "Provider produced inconsistent result after apply" errors
// 4. It follows the existing codebase patterns for set handling
//
// The semantic equality functions used unsupported types like tftypes.SemanticEqualityRequest
// that are not available in framework version 1.14.1. The plan modifier approach
// achieves the same goals while being compatible with the current framework version.
