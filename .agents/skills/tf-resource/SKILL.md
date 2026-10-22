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
- if the attribute is not required in the input, and its type is 'list', 'map', or 'set',  set the mode to 'optional_computed'
- if the attribute is not required in the input, and its type is not 'list', 'map', or 'set',  set the mode to 'optional'

#### 1.2.6 Attribute Hint
- if the attribute is of type 'nested' and the sub-attributes are called 'id' and 'name' (both string), add a 'idName' to the set of hints
- if the attribute type is 'list', 'map', or 'set', and its mode is 'optional_computed', add 'modifyPlan' to the set of hints



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

### 2.3 Resource file
Create or update the resource file named 'resource_{resource-name}.go' where the resource name is the snail_case resource name without the "cato_" prefix.
Use the `assets/resource_template.go`, replace the placeholders `<...>` with the appropriate value.
Do not delete or replace bodies of the functions ImportState, Create, Read, Update, Delete if they already exist
