---
name: tf-resource
description: Create or update terraform resource
---

# Terraform Resource Management

Use this skill to create or update terraform resource in the terraform-provider-cato provider

## Prerequisites

1. Graphql schema describing CRUD operations on the resource.
   By default, it can be fetched from 'https://system.cc.catonetworks.com/api/schema?with_undocumented=true', unless
   the user specifies otherwise. If it cannot be fetch, resolve the problem the user.
2. Name of the new resource. It shall be lower case, with underscores, starting with "cato_"
3. Directory in the `terraform-provider-cato` workspace where the code shall be placed. `./internal/provider/{domain}/{resource_package}`. For policy rules, use `./internal/provider/policy/{domain}/{resource_package}`.
   Suggest the resource package name based on the resource name, should contain lower case ASCII letters and numbers only.
4. Coding rules defined in `.agents/references/coding-rules.yaml`


## 1 Schema Analysis
The goal of the analysis is to prepare a description of a terraform schema of the resource.

1. Identify the resource APIs, ask the user unless it was provided in the prompt: 
   - Create mutation
   - Update mutation
   - Delete mutation
   - Read query

2. Get the input of the Create operation recursively, creating a tree of input attributes,
   and get the response/payload of the Create operation recursively.
3. Find a sub-tree in the response that is similar to the input possibly with additional computed values like `id`.
   Mark these values as 'computed'.
4. Find relevant create method on Go struct cato_go_sdk.Client, which name of is made by concatenating
   graphql mutation (e.g. for Mutation policy.ztnaAlwaysOn.addSection the Go method will be  PolicyZtnaAlwaysOnAddSection)
   and analyze its input parameter. It should match the graphql schema definition, report any discrepancies to the user.
5. If the attribute type is enum, find the related Go struct and add a Hint:  enum={enum-name}
   e.g.: enum=PolicyRulePositionEnum
6. Handle special cases (see "Special Cases" below)
7. Determine the properties of the attributes (see "Attribute Properties" below)
8. Create `schema_description.yaml` file based on the input attribute tree plus the computed values from the response (see `assets/schema_description_*.yaml` for the format.).
   Store the file in the selected {resource_package} directory
   Keep the same tree structure as is in the input.
   Sort the attributes (at all levels) by name.
   Present it to the user and have it approved.
9. Create `ai_context.md` in the {resource_package} directory, keep the relevant information there,
   including
   - resource name
   - references to the API CRUD mutation


### 1.1 Special Cases
### 1.1.1 References
If the input object contains 2 attributes with types 'ObjectRefBy' and 'String', replace it with the following schema_description subtree:
```yaml
- Name: <attribute name>
  Description: <attribute description>
  Type: nested
  Attributes:
    - Name: id
      Description: <attribute> ID
      Type: string
      Mode: computed_optional
    - Name: name
      Description: <attribute> name
      Type: string
      Mode: computed_optional
  Mode: <required or optional>
  Hints:
    - idName
```


### 1.2 Attribute Properties
#### 1.2.1 Attribute Name
Name of the attribute, snail_case format

#### 1.2.2 Attribute Description
Description of the attribute in plain English.

#### 1.2.3 Attribute Type
Data type of the attribute, one of string,bool,float32,float64,int32,int64,list,map,set,nested.

**Graphql simple type mapping:**
| Graphql Schema | TF Schema Description |
| --- | --- | 
| String | string |
| Int | int64 |
| Float | float64 |
| ID | string |
| Boolean | bool |
| enum types | string |
| scalar types | string |

**Graphql complex type mapping:**
- if the Graphql attribute is not a simple type (String, Int, Float, Boolean, enum, scalar) and is not a list
  (list is expressed with square brackets), then its type shall be 'nested'
- if it is a Graphql list then:
  Suggest if terraform should use a list, map or set and confirm with the user.
  If the order if the items is important, use 'list', otherwise ask the user, suggest the best approach.
    
#### 1.2.4 Attribute ItemType
The same rules as 'Type', applied to a nested attribute or a collection item 

#### 1.2.5 Attribute Mode
one of computed,required,optional,optional_computed.

- if the attribute is not in the input, and is only returned, set the mode to 'computed'
- if the Graphql attribute is required in the input (marked with '!') set the mode to 'required'
- if the attribute is not required in the input, and input is graphql enum type, set the mode to 'required'
- if the attribute is not required in the input, and its type is 'list', 'map', or 'set',  set the mode to 'optional_computed'
- if the attribute is not required in the input, and type is string, and input is not enum, set the mode to 'optional_computed'
- if the attribute is not required in the input, and type is bool, set the mode to 'optional_computed'
- if the attribute is not required in the input, and its type is not 'list', 'map', 'set' or 'string',  set the mode to 'optional'

#### 1.2.6 Attribute Hint
- if the attribute is of type 'nested' and the sub-attributes are called 'id' and 'name' (both string), add a 'idName' to the set of hints
- if the attribute type is 'list', 'map', 'set', 'string', or 'bool' and its mode is 'optional_computed', add 'modifyPlan' to the set of hints


## 2 Provider Implementation
### 2.1 Schema file
Create or update the schema file named 'schema_{resource-name}.go' where the resource name is the snail_case resource name without the "cato_" prefix.
Use the `assets/schema_template.go`, replace the placeholders `<...>` with the appropriate value.
Do not delete or replace body of the Schema function if it already exists.
Update the Schema method and its helper methods, so that it matches the approved `schema_description.yaml`,
Use the coding-rules `TFSCHEMA-*`.


### 2.2 Model file
Create or update the model file named 'model_{resource-name}.go' where the resource name is the snail_case resource name without the "cato_" prefix.
The Go types and variables must match the schema_description and the Schema from the Schema method.
Use the coding-rules `MODEL-*`.


### 2.3 Client file
Create or update the client file named 'client_{resource-name}.go' where the resource name is the snail_case resource name without the "cato_" prefix.
The file shall contain interface definition of the Cato Client 
```go
type Client interface {
   // methods of cato_go_sdk.Client used in the current package
 }
```

### 2.4 Resource file
Create or update the resource file named 'resource_{resource-name}.go' where the resource name is the snail_case resource name without the "cato_" prefix.
Use the `assets/resource_template.go` for the scaffolding, replace the placeholders `<...>` with the appropriate value.
The resource file shall contain the NewXXX function and the following methods:
  - getClient
  - Metadata
  - Configure
  - ImportState
  - Create
  - Read
  - Update
  - Delete
Other helper functions, converters must be in a converters file.

#### 2.4.1 Method getClient
The `getClient` method must be made based on resource_template.go template.

#### 2.4.2 Method Metadata
The `Metadata` method must be made based on resource_template.go template.

#### 2.4.3 Method Configure
The `Configure` method must be made based on resource_template.go template.

#### 2.4.4 Method ImportState
If the resource-related API supports fetching by ID and the resource has the `id` attribute,
use the `ImportStatePassthroughID` helper function to provide information to the framework,
no other code is needed in the method body.
Otherwise ask the user if and how it is to be implemented.
The `ImportStatePassthroughID` requires the ID path to be specified, for example
```go
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resource.ImportStatePassthroughID(ctx, path.Root("rule").AtName("id"), req, resp)
```

#### 2.4.5 Method Create
The `Create` method should have the following structure:
- get a plan
- call the API
   - prepare input for the create API call (converter helper function)
   - call the API
   - check for errors
   - convert the response into terraform type system (converter helper function)
- update the state

If there is one API call per resource, the 'call API' block will be part of the Create method directly,
else there should be calls to helper methods, each helper implementing its 'call API' block.

Getting a plan example:
```go
var plan ExampleResource
diags := req.Plan.Get(ctx, &plan)
resp.Diagnostics.Append(diags...)
if resp.Diagnostics.HasError() {
   return
}
resourceName := plan.xxxxxxx.ValueString() // for debugging purposes
```
#### 2.4.5.1 Calling the API
To call API, use the following pattern 
```go
errMsg := fmt.Sprintf("failed to create ... %s", resourceName)
result, err := r.getClient().ExampleMethod(ctx, ....)
if err != nil {
   resp.Diagnostics.AddError(errMsg, err.Error())
   return
}
```
If the Create API returns the resource ID, validate it has been returned, e.g.
```go
newID := result.GetXXX().GetXXXX().GetID()
if newID == "" {
   resp.Diagnostics.AddError(errMsg, "API did not return resource ID")
   return
}
```

#### 2.4.6 Method Read
The `Read` method should have the following structure:
- get the state, extract resource ID to read
- call the API
   - prepare input for the READ API call if needed (converter helper function)
   - call the API; use r.getClient() to get the client object
   - check for errors
   - convert the response into terraform type system (converter helper function)
- update the state

If there is one API call per resource, the 'call API' block will be part of the Read method directly,
else there should be calls to helper methods, each helper implementing its 'call API' block.

#### 2.4.7 Method Update
The `Update` method should follow the same pattern as the `Create` method.

#### 2.4.8 Method Delete
The `Delete` method should have the following structure:
- get the state, extract resource ID to delete
- call the API
   - prepare input for the Delete API call if needed (converter helper function)
   - call the API; use r.getClient() to get the client object
   - check for errors



### 2.5 Converter file
Create or update the converter file named 'converters_{resource-name}.go' where the resource name is the snail_case resource name without the "cato_" prefix.
Use the `CONV-*` rules for the converters.
There are two major groups of converters:
- input builders: convert TF type system (typically plan) into a plain Go structures (typically SDK types defined in cato_models package)
- parsers: convert API responses into TF type system


#### 2.5.1 Input builders
The core of the input builder function should look like this:
```go
input := cato_models.ExampleInput{
  ExampleBoolField:  plan.ExampleBool.ValueBool(),
  ExampleStrings:    convert.PrepareStringList[string](ctx, plan.ExampleStrings, diags),
}
```
Reuse helper builder functions if possible.
- builders in this package that build inputs for Create or Update API calls
- builders in shared sub-component package (e.g. `./internal/provider/shared/idname/convert*.go)
- builders in shared converter package (e.g. `./internal/provider/shared/convert/*.go)


#### 2.5.2 Parsers
The core of the parser function should look like this:
```go
state = &ExampleResource{
  ExampleStringField:   types.StringValue(string(valueFromAPI)),
  ExampleComplexField:  r.parseExampleMethod(ctx, valueFromAPI, &diags),
  ExampleRefField:      idname.ParseIDNameSet(ctx, valueFromAPI, &diags),
```

Reuse helper parser functions if possible.
- parsers in this package that transform responses from Create, Read or Update API calls
- parsers in shared sub-component package (e.g. `./internal/provider/shared/idname/convert*.go)
- parsers in shared converter package (e.g. `./internal/provider/shared/convert/*.go)
