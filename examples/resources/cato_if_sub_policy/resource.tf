// Internet Firewall sub-policy (a nested policy scoped by a SUB_POLICY_SCOPE rule).
//
// NOTE: The Cato API has no updateSubPolicy mutation, so changing `name` or
// `description` forces replacement, which also removes any rules contained in
// the sub-policy. The scope must set at least one non-ANY match field.
// Changing `at.position` or `at.ref` moves the sub-policy scope in place and
// preserves its child rules. Refresh retains the last applied placement;
// external reordering is not detected. If cato_bulk_if_move_rule manages
// ordering, keep `at` unchanged and change positions in that resource instead.

// minimal sub-policy
resource "cato_if_sub_policy" "minimal" {
  name        = "Remote Users Sub-Policy"
  description = "Rules that only apply to remote users"

  at = {
    position = "LAST_IN_POLICY"
  }

  scope = {
    enabled           = true
    connection_origin = "REMOTE"
    source            = {}
    destination       = {}
  }
}

// Scope tracking is API-managed; omit it from scope.

// full sub-policy with a scoped source and a nested rule owned by it
resource "cato_if_sub_policy" "full" {
  name        = "Marketing Sub-Policy"
  description = "Rules scoped to the Marketing user group"

  at = {
    position = "LAST_IN_POLICY"
  }

  scope = {
    enabled = true
    source = {
      users_group = [
        {
          name = "Marketing-Teams"
        }
      ]
    }
    destination = {}
  }

  // Optional: manage the API-created cleanup rule with this sub-policy.
  cleanup_rule = {
    action  = "BLOCK"
    enabled = true
    tracking = {
      event = {
        enabled = true
      }
    }
  }
}

// a rule placed inside the "full" sub-policy via sub_policy_id
resource "cato_if_rule" "in_sub_policy" {
  sub_policy_id = cato_if_sub_policy.full.id

  // Rules owned by a sub-policy are always positioned before the sub-policy
  // cleanup rule; the `at` position below is required by the schema but the
  // provider anchors the rule inside the sub-policy automatically.
  at = {
    position = "LAST_IN_POLICY"
  }

  rule = {
    name    = "Block test.com for Marketing"
    enabled = true
    action  = "BLOCK"
    source  = {}
    destination = {
      domain = ["test.com"]
    }
    tracking = {
      event = {
        enabled = true
      }
    }
  }

  depends_on = [cato_if_sub_policy.full]
}
