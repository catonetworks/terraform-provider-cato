# Terraform Provider Resource Implementation Guidelines

## Purpose and scope

This document describes the logical structure of the resource implementation go code

## Logical Structure
### 1. Scaffolding
| TF Framework Callback Methods |
| :---                          |
| - Metadata                    |
| - Configure                   |

### 2. Schema Declaration
- defines TF resource schema and its internal representation
  - consistent structure / types / tfsdk tags

| Schema              | TF Model      |
| :---                | :---          |
| - Schema definition | - Go structs  |
| - Validators        | - AttrTypes   |
| - Plan Modifiers    | - ObjectTypes | 

### 3. CRUD Operations
| TF Framework Callback Methods |
| :---                          |
| - Create                      |
| - Read                        |
| - Update                      |
| - Delete                      |
| - ImportState                 |


#### Simple Create Method:
1. Build API Request from the Plan
2. Call API
3. Parse API response, update State
- error handling

#### Complex Create Method:
- Call multiple helpers that implement "Simple Create Method"
